package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"pulse/internal/config"
	redisClient "pulse/internal/redis"
)

// RedisQueue 基于Redis的消息队列实现
type RedisQueue struct {
	client      *redisClient.Client
	logger      *zap.Logger
	config      *config.Config
	subscribers map[string]*subscriber
	mu          sync.RWMutex
	ctx         context.Context
	cancel      context.CancelFunc
	wg          sync.WaitGroup
	running     bool
}

// subscriber 订阅者信息
type subscriber struct {
	topic   string
	handler Handler
	options *SubscribeOptions
	cancel  context.CancelFunc
}

// NewRedisQueue 创建新的Redis消息队列
func NewRedisQueue(client *redisClient.Client, cfg *config.Config, logger *zap.Logger) *RedisQueue {
	ctx, cancel := context.WithCancel(context.Background())

	return &RedisQueue{
		client:      client,
		logger:      logger,
		config:      cfg,
		subscribers: make(map[string]*subscriber),
		ctx:         ctx,
		cancel:      cancel,
	}
}

// Publish 发布消息
func (q *RedisQueue) Publish(ctx context.Context, topic string, payload []byte, opts ...PublishOption) error {
	options := applyPublishOptions(opts...)

	msg := &Message{
		ID:        uuid.New().String(),
		Topic:     topic,
		Payload:   payload,
		Headers:   options.Headers,
		Metadata:  options.Metadata,
		MaxRetry:  options.MaxRetry,
		CreatedAt: time.Now(),
	}

	return q.publishMessage(ctx, msg)
}

// PublishWithDelay 延迟发布消息
func (q *RedisQueue) PublishWithDelay(ctx context.Context, topic string, payload []byte, delay time.Duration, opts ...PublishOption) error {
	options := applyPublishOptions(opts...)

	scheduledAt := time.Now().Add(delay)
	msg := &Message{
		ID:          uuid.New().String(),
		Topic:       topic,
		Payload:     payload,
		Headers:     options.Headers,
		Metadata:    options.Metadata,
		MaxRetry:    options.MaxRetry,
		Delay:       delay,
		CreatedAt:   time.Now(),
		ScheduledAt: &scheduledAt,
	}

	// 将延迟消息添加到有序集合中
	return q.scheduleMessage(ctx, msg)
}

// PublishBatch 批量发布消息
func (q *RedisQueue) PublishBatch(ctx context.Context, messages []*Message) error {
	pipe := q.client.GetClient().Pipeline()

	for _, msg := range messages {
		if msg.ID == "" {
			msg.ID = uuid.New().String()
		}
		if msg.CreatedAt.IsZero() {
			msg.CreatedAt = time.Now()
		}

		msgData, err := json.Marshal(msg)
		if err != nil {
			return fmt.Errorf("failed to marshal message: %w", err)
		}

		queueKey := q.getQueueKey(msg.Topic)
		pipe.LPush(ctx, queueKey, msgData)
	}

	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to publish batch messages: %w", err)
	}

	q.logger.Info("Batch messages published", zap.Int("count", len(messages)))
	return nil
}

// Subscribe 订阅主题
func (q *RedisQueue) Subscribe(ctx context.Context, topic string, handler Handler, opts ...SubscribeOption) error {
	options := applySubscribeOptions(opts...)

	q.mu.Lock()
	defer q.mu.Unlock()

	if _, exists := q.subscribers[topic]; exists {
		return fmt.Errorf("topic %s already subscribed", topic)
	}

	subCtx, cancel := context.WithCancel(q.ctx)
	sub := &subscriber{
		topic:   topic,
		handler: handler,
		options: options,
		cancel:  cancel,
	}

	q.subscribers[topic] = sub

	// 启动消费者协程
	for i := 0; i < options.Concurrency; i++ {
		q.wg.Add(1)
		go q.consumeMessages(subCtx, sub, i)
	}

	q.logger.Info("Subscribed to topic",
		zap.String("topic", topic),
		zap.Int("concurrency", options.Concurrency),
	)

	return nil
}

// Unsubscribe 取消订阅
func (q *RedisQueue) Unsubscribe(topic string) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	sub, exists := q.subscribers[topic]
	if !exists {
		return fmt.Errorf("topic %s not subscribed", topic)
	}

	sub.cancel()
	delete(q.subscribers, topic)

	q.logger.Info("Unsubscribed from topic", zap.String("topic", topic))
	return nil
}

// Start 启动消费者
func (q *RedisQueue) Start(ctx context.Context) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.running {
		return fmt.Errorf("queue is already running")
	}

	q.running = true

	// 启动延迟消息处理器
	q.wg.Add(1)
	go q.processDelayedMessages()

	q.logger.Info("Redis queue started")
	return nil
}

// Stop 停止消费者
func (q *RedisQueue) Stop() error {
	q.mu.Lock()
	defer q.mu.Unlock()

	if !q.running {
		return nil
	}

	q.running = false
	q.cancel()

	// 等待所有协程结束
	q.wg.Wait()

	q.logger.Info("Redis queue stopped")
	return nil
}

// Close 关闭队列
func (q *RedisQueue) Close() error {
	return q.Stop()
}

// Health 获取队列健康状态
func (q *RedisQueue) Health(ctx context.Context) map[string]interface{} {
	q.mu.RLock()
	subscriberCount := len(q.subscribers)
	q.mu.RUnlock()

	redisHealth := q.client.Health(ctx)

	return map[string]interface{}{
		"status":           "healthy",
		"running":          q.running,
		"subscriber_count": subscriberCount,
		"redis":            redisHealth,
	}
}

// publishMessage 发布消息到队列
func (q *RedisQueue) publishMessage(ctx context.Context, msg *Message) error {
	msgData, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	queueKey := q.getQueueKey(msg.Topic)
	err = q.client.LPush(ctx, queueKey, msgData)
	if err != nil {
		return fmt.Errorf("failed to push message to queue: %w", err)
	}

	q.logger.Debug("Message published",
		zap.String("topic", msg.Topic),
		zap.String("message_id", msg.ID),
	)

	return nil
}

// scheduleMessage 调度延迟消息
func (q *RedisQueue) scheduleMessage(ctx context.Context, msg *Message) error {
	msgData, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal delayed message: %w", err)
	}

	delayedKey := q.getDelayedKey()
	score := float64(msg.ScheduledAt.Unix())

	err = q.client.ZAdd(ctx, delayedKey, &redis.Z{
		Score:  score,
		Member: msgData,
	})
	if err != nil {
		return fmt.Errorf("failed to schedule delayed message: %w", err)
	}

	q.logger.Debug("Message scheduled",
		zap.String("topic", msg.Topic),
		zap.String("message_id", msg.ID),
		zap.Time("scheduled_at", *msg.ScheduledAt),
	)

	return nil
}

// consumeMessages 消费消息
func (q *RedisQueue) consumeMessages(ctx context.Context, sub *subscriber, workerID int) {
	defer q.wg.Done()

	queueKey := q.getQueueKey(sub.topic)
	processingKey := q.getProcessingKey(sub.topic)

	q.logger.Info("Consumer worker started",
		zap.String("topic", sub.topic),
		zap.Int("worker_id", workerID),
	)

	for {
		select {
		case <-ctx.Done():
			q.logger.Info("Consumer worker stopped",
				zap.String("topic", sub.topic),
				zap.Int("worker_id", workerID),
			)
			return
		default:
			// 从队列中获取消息
			result, err := q.client.GetClient().BRPopLPush(ctx, queueKey, processingKey, time.Second).Result()
			if err != nil {
				if err == redis.Nil {
					// 没有消息，继续等待
					continue
				}
				q.logger.Error("Failed to pop message from queue",
					zap.String("topic", sub.topic),
					zap.Error(err),
				)
				continue
			}

			// 解析消息
			var msg Message
			if err := json.Unmarshal([]byte(result), &msg); err != nil {
				q.logger.Error("Failed to unmarshal message",
					zap.String("topic", sub.topic),
					zap.Error(err),
				)
				// 从处理队列中移除无效消息
				q.client.LRem(ctx, processingKey, 1, result)
				continue
			}

			// 处理消息
			q.handleMessage(ctx, sub, &msg, result, processingKey)
		}
	}
}

// handleMessage 处理单个消息
func (q *RedisQueue) handleMessage(ctx context.Context, sub *subscriber, msg *Message, msgData string, processingKey string) {
	start := time.Now()

	// 创建带超时的上下文
	msgCtx := ctx
	if sub.options.AckTimeout > 0 {
		var cancel context.CancelFunc
		msgCtx, cancel = context.WithTimeout(ctx, sub.options.AckTimeout)
		defer cancel()
	}

	// 调用处理器
	err := sub.handler(msgCtx, msg)

	duration := time.Since(start)

	if err != nil {
		q.logger.Error("Message handler failed",
			zap.String("topic", msg.Topic),
			zap.String("message_id", msg.ID),
			zap.Error(err),
			zap.Duration("duration", duration),
		)

		// 处理重试逻辑
		if msg.Retry < msg.MaxRetry {
			msg.Retry++
			q.retryMessage(ctx, msg)
		} else {
			q.logger.Error("Message exceeded max retry count",
				zap.String("topic", msg.Topic),
				zap.String("message_id", msg.ID),
				zap.Int("retry_count", msg.Retry),
			)
			// 可以将消息发送到死信队列
			q.sendToDeadLetterQueue(ctx, msg)
		}
	} else {
		q.logger.Debug("Message processed successfully",
			zap.String("topic", msg.Topic),
			zap.String("message_id", msg.ID),
			zap.Duration("duration", duration),
		)
	}

	// 从处理队列中移除消息
	q.client.LRem(ctx, processingKey, 1, msgData)
}

// retryMessage 重试消息
func (q *RedisQueue) retryMessage(ctx context.Context, msg *Message) {
	// 计算重试延迟
	retryDelay := time.Duration(msg.Retry) * time.Second
	if retryDelay > 60*time.Second {
		retryDelay = 60 * time.Second
	}

	// 延迟重新发布消息
	scheduledAt := time.Now().Add(retryDelay)
	msg.ScheduledAt = &scheduledAt

	if err := q.scheduleMessage(ctx, msg); err != nil {
		q.logger.Error("Failed to schedule retry message",
			zap.String("topic", msg.Topic),
			zap.String("message_id", msg.ID),
			zap.Error(err),
		)
	}
}

// sendToDeadLetterQueue 发送到死信队列
func (q *RedisQueue) sendToDeadLetterQueue(ctx context.Context, msg *Message) {
	deadLetterKey := q.getDeadLetterKey(msg.Topic)

	msgData, err := json.Marshal(msg)
	if err != nil {
		q.logger.Error("Failed to marshal dead letter message",
			zap.String("topic", msg.Topic),
			zap.String("message_id", msg.ID),
			zap.Error(err),
		)
		return
	}

	if err := q.client.LPush(ctx, deadLetterKey, msgData); err != nil {
		q.logger.Error("Failed to send message to dead letter queue",
			zap.String("topic", msg.Topic),
			zap.String("message_id", msg.ID),
			zap.Error(err),
		)
	}
}

// processDelayedMessages 处理延迟消息
func (q *RedisQueue) processDelayedMessages() {
	defer q.wg.Done()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-q.ctx.Done():
			return
		case <-ticker.C:
			q.processReadyDelayedMessages()
		}
	}
}

// processReadyDelayedMessages 处理准备好的延迟消息
func (q *RedisQueue) processReadyDelayedMessages() {
	ctx := context.Background()
	delayedKey := q.getDelayedKey()
	now := float64(time.Now().Unix())

	// 获取到期的延迟消息
	results, err := q.client.GetClient().ZRangeByScore(ctx, delayedKey, &redis.ZRangeBy{
		Min: "0",
		Max: fmt.Sprintf("%f", now),
	}).Result()

	if err != nil {
		q.logger.Error("Failed to get delayed messages", zap.Error(err))
		return
	}

	for _, result := range results {
		// 解析消息
		var msg Message
		if err := json.Unmarshal([]byte(result), &msg); err != nil {
			q.logger.Error("Failed to unmarshal delayed message", zap.Error(err))
			continue
		}

		// 发布消息到队列
		if err := q.publishMessage(ctx, &msg); err != nil {
			q.logger.Error("Failed to publish delayed message",
				zap.String("topic", msg.Topic),
				zap.String("message_id", msg.ID),
				zap.Error(err),
			)
			continue
		}

		// 从延迟队列中移除
		q.client.ZRem(ctx, delayedKey, result)
	}
}

// getQueueKey 获取队列键名
func (q *RedisQueue) getQueueKey(topic string) string {
	return fmt.Sprintf("queue:%s", topic)
}

// getProcessingKey 获取处理队列键名
func (q *RedisQueue) getProcessingKey(topic string) string {
	return fmt.Sprintf("queue:%s:processing", topic)
}

// getDelayedKey 获取延迟队列键名
func (q *RedisQueue) getDelayedKey() string {
	return "queue:delayed"
}

// getDeadLetterKey 获取死信队列键名
func (q *RedisQueue) getDeadLetterKey(topic string) string {
	return fmt.Sprintf("queue:%s:dead", topic)
}
