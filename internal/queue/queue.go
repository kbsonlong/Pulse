package queue

import (
	"context"
	"time"
)

// Message 消息结构
type Message struct {
	ID       string                 `json:"id"`
	Topic    string                 `json:"topic"`
	Payload  []byte                 `json:"payload"`
	Headers  map[string]string      `json:"headers,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
	Retry    int                    `json:"retry"`
	MaxRetry int                    `json:"max_retry"`
	Delay    time.Duration          `json:"delay"`
	CreatedAt time.Time             `json:"created_at"`
	ScheduledAt *time.Time          `json:"scheduled_at,omitempty"`
}

// Handler 消息处理器函数类型
type Handler func(ctx context.Context, msg *Message) error

// Producer 消息生产者接口
type Producer interface {
	// Publish 发布消息
	Publish(ctx context.Context, topic string, payload []byte, opts ...PublishOption) error
	
	// PublishWithDelay 延迟发布消息
	PublishWithDelay(ctx context.Context, topic string, payload []byte, delay time.Duration, opts ...PublishOption) error
	
	// PublishBatch 批量发布消息
	PublishBatch(ctx context.Context, messages []*Message) error
	
	// Close 关闭生产者
	Close() error
}

// Consumer 消息消费者接口
type Consumer interface {
	// Subscribe 订阅主题
	Subscribe(ctx context.Context, topic string, handler Handler, opts ...SubscribeOption) error
	
	// Unsubscribe 取消订阅
	Unsubscribe(topic string) error
	
	// Start 启动消费者
	Start(ctx context.Context) error
	
	// Stop 停止消费者
	Stop() error
}

// Queue 消息队列接口
type Queue interface {
	Producer
	Consumer
	
	// Health 获取队列健康状态
	Health(ctx context.Context) map[string]interface{}
}

// PublishOption 发布选项
type PublishOption func(*PublishOptions)

// PublishOptions 发布选项配置
type PublishOptions struct {
	Headers    map[string]string
	Metadata   map[string]interface{}
	MaxRetry   int
	Priority   int
	Expiration time.Duration
}

// WithHeaders 设置消息头
func WithHeaders(headers map[string]string) PublishOption {
	return func(opts *PublishOptions) {
		opts.Headers = headers
	}
}

// WithMetadata 设置元数据
func WithMetadata(metadata map[string]interface{}) PublishOption {
	return func(opts *PublishOptions) {
		opts.Metadata = metadata
	}
}

// WithMaxRetry 设置最大重试次数
func WithMaxRetry(maxRetry int) PublishOption {
	return func(opts *PublishOptions) {
		opts.MaxRetry = maxRetry
	}
}

// WithPriority 设置消息优先级
func WithPriority(priority int) PublishOption {
	return func(opts *PublishOptions) {
		opts.Priority = priority
	}
}

// WithExpiration 设置消息过期时间
func WithExpiration(expiration time.Duration) PublishOption {
	return func(opts *PublishOptions) {
		opts.Expiration = expiration
	}
}

// SubscribeOption 订阅选项
type SubscribeOption func(*SubscribeOptions)

// SubscribeOptions 订阅选项配置
type SubscribeOptions struct {
	Concurrency   int
	MaxRetry      int
	RetryDelay    time.Duration
	AckTimeout    time.Duration
	PrefetchCount int
	AutoAck       bool
}

// WithConcurrency 设置并发数
func WithConcurrency(concurrency int) SubscribeOption {
	return func(opts *SubscribeOptions) {
		opts.Concurrency = concurrency
	}
}

// WithSubscribeMaxRetry 设置订阅最大重试次数
func WithSubscribeMaxRetry(maxRetry int) SubscribeOption {
	return func(opts *SubscribeOptions) {
		opts.MaxRetry = maxRetry
	}
}

// WithRetryDelay 设置重试延迟
func WithRetryDelay(delay time.Duration) SubscribeOption {
	return func(opts *SubscribeOptions) {
		opts.RetryDelay = delay
	}
}

// WithAckTimeout 设置确认超时时间
func WithAckTimeout(timeout time.Duration) SubscribeOption {
	return func(opts *SubscribeOptions) {
		opts.AckTimeout = timeout
	}
}

// WithPrefetchCount 设置预取数量
func WithPrefetchCount(count int) SubscribeOption {
	return func(opts *SubscribeOptions) {
		opts.PrefetchCount = count
	}
}

// WithAutoAck 设置自动确认
func WithAutoAck(autoAck bool) SubscribeOption {
	return func(opts *SubscribeOptions) {
		opts.AutoAck = autoAck
	}
}

// applyPublishOptions 应用发布选项
func applyPublishOptions(opts ...PublishOption) *PublishOptions {
	options := &PublishOptions{
		Headers:  make(map[string]string),
		Metadata: make(map[string]interface{}),
		MaxRetry: 3,
		Priority: 0,
	}
	
	for _, opt := range opts {
		opt(options)
	}
	
	return options
}

// applySubscribeOptions 应用订阅选项
func applySubscribeOptions(opts ...SubscribeOption) *SubscribeOptions {
	options := &SubscribeOptions{
		Concurrency:   1,
		MaxRetry:      3,
		RetryDelay:    time.Second,
		AckTimeout:    30 * time.Second,
		PrefetchCount: 1,
		AutoAck:       false,
	}
	
	for _, opt := range opts {
		opt(options)
	}
	
	return options
}