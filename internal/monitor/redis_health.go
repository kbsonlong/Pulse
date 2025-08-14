package monitor

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

// RedisHealthCheck Redis健康检查
type RedisHealthCheck struct {
	name   string
	client *redis.Client
}

// NewRedisHealthCheck 创建Redis健康检查
func NewRedisHealthCheck(name string, client *redis.Client) *RedisHealthCheck {
	return &RedisHealthCheck{
		name:   name,
		client: client,
	}
}

// Name 返回健康检查名称
func (r *RedisHealthCheck) Name() string {
	return r.name
}

// Check 执行Redis健康检查
func (r *RedisHealthCheck) Check(ctx context.Context) HealthResult {
	start := time.Now()
	result := HealthResult{
		Timestamp: start,
		Details:   make(map[string]interface{}),
	}
	
	// 执行PING命令
	pingResult := r.client.Ping(ctx)
	if err := pingResult.Err(); err != nil {
		result.Status = HealthStatusUnhealthy
		result.Message = fmt.Sprintf("Redis ping failed: %v", err)
		result.Error = err
		result.Duration = time.Since(start)
		return result
	}
	
	// 检查连接池状态
	stats := r.client.PoolStats()
	result.Details["pool_stats"] = map[string]interface{}{
		"hits":         stats.Hits,
		"misses":       stats.Misses,
		"timeouts":     stats.Timeouts,
		"total_conns":  stats.TotalConns,
		"idle_conns":   stats.IdleConns,
		"stale_conns":  stats.StaleConns,
	}
	
	// 测试基本操作
	testKey := fmt.Sprintf("health_check_%d", time.Now().UnixNano())
	testValue := "test_value"
	
	// SET操作
	setResult := r.client.Set(ctx, testKey, testValue, time.Minute)
	if err := setResult.Err(); err != nil {
		result.Status = HealthStatusDegraded
		result.Message = fmt.Sprintf("Redis SET operation failed: %v", err)
		result.Error = err
		result.Duration = time.Since(start)
		return result
	}
	
	// GET操作
	getResult := r.client.Get(ctx, testKey)
	if err := getResult.Err(); err != nil {
		result.Status = HealthStatusDegraded
		result.Message = fmt.Sprintf("Redis GET operation failed: %v", err)
		result.Error = err
		result.Duration = time.Since(start)
		return result
	}
	
	if getResult.Val() != testValue {
		result.Status = HealthStatusDegraded
		result.Message = "Redis GET operation returned unexpected value"
		result.Duration = time.Since(start)
		return result
	}
	
	// 清理测试数据
	r.client.Del(ctx, testKey)
	
	// 获取Redis信息
	infoResult := r.client.Info(ctx, "server", "memory", "stats")
	if err := infoResult.Err(); err == nil {
		result.Details["redis_info"] = infoResult.Val()
	}
	
	result.Status = HealthStatusHealthy
	result.Message = "Redis is healthy"
	result.Duration = time.Since(start)
	
	return result
}

// QueueHealthCheck 消息队列健康检查
type QueueHealthCheck struct {
	name        string
	redisClient *redis.Client
	queueName   string
}

// NewQueueHealthCheck 创建消息队列健康检查
func NewQueueHealthCheck(name string, redisClient *redis.Client, queueName string) *QueueHealthCheck {
	return &QueueHealthCheck{
		name:        name,
		redisClient: redisClient,
		queueName:   queueName,
	}
}

// Name 返回健康检查名称
func (q *QueueHealthCheck) Name() string {
	return q.name
}

// Check 执行消息队列健康检查
func (q *QueueHealthCheck) Check(ctx context.Context) HealthResult {
	start := time.Now()
	result := HealthResult{
		Timestamp: start,
		Details:   make(map[string]interface{}),
	}
	
	// 检查队列长度
	lengthResult := q.redisClient.LLen(ctx, q.queueName)
	if err := lengthResult.Err(); err != nil {
		result.Status = HealthStatusUnhealthy
		result.Message = fmt.Sprintf("Failed to get queue length: %v", err)
		result.Error = err
		result.Duration = time.Since(start)
		return result
	}
	
	queueLength := lengthResult.Val()
	result.Details["queue_length"] = queueLength
	
	// 测试消息发送和接收
	testMessage := fmt.Sprintf("health_check_%d", time.Now().UnixNano())
	
	// 发送测试消息
	pushResult := q.redisClient.LPush(ctx, q.queueName, testMessage)
	if err := pushResult.Err(); err != nil {
		result.Status = HealthStatusDegraded
		result.Message = fmt.Sprintf("Failed to push test message: %v", err)
		result.Error = err
		result.Duration = time.Since(start)
		return result
	}
	
	// 接收测试消息
	popResult := q.redisClient.RPop(ctx, q.queueName)
	if err := popResult.Err(); err != nil {
		result.Status = HealthStatusDegraded
		result.Message = fmt.Sprintf("Failed to pop test message: %v", err)
		result.Error = err
		result.Duration = time.Since(start)
		return result
	}
	
	if popResult.Val() != testMessage {
		result.Status = HealthStatusDegraded
		result.Message = "Queue test message mismatch"
		result.Duration = time.Since(start)
		return result
	}
	
	// 检查队列状态
	if queueLength > 10000 {
		result.Status = HealthStatusDegraded
		result.Message = fmt.Sprintf("Queue length is high: %d", queueLength)
	} else {
		result.Status = HealthStatusHealthy
		result.Message = "Queue is healthy"
	}
	
	result.Duration = time.Since(start)
	return result
}

// CacheHealthCheck 缓存健康检查
type CacheHealthCheck struct {
	name        string
	redisClient *redis.Client
}

// NewCacheHealthCheck 创建缓存健康检查
func NewCacheHealthCheck(name string, redisClient *redis.Client) *CacheHealthCheck {
	return &CacheHealthCheck{
		name:        name,
		redisClient: redisClient,
	}
}

// Name 返回健康检查名称
func (c *CacheHealthCheck) Name() string {
	return c.name
}

// Check 执行缓存健康检查
func (c *CacheHealthCheck) Check(ctx context.Context) HealthResult {
	start := time.Now()
	result := HealthResult{
		Timestamp: start,
		Details:   make(map[string]interface{}),
	}
	
	// 测试缓存操作
	testKey := fmt.Sprintf("cache_health_check_%d", time.Now().UnixNano())
	testValue := "cache_test_value"
	expiration := time.Minute
	
	// SET操作（带过期时间）
	setResult := c.redisClient.Set(ctx, testKey, testValue, expiration)
	if err := setResult.Err(); err != nil {
		result.Status = HealthStatusUnhealthy
		result.Message = fmt.Sprintf("Cache SET operation failed: %v", err)
		result.Error = err
		result.Duration = time.Since(start)
		return result
	}
	
	// GET操作
	getResult := c.redisClient.Get(ctx, testKey)
	if err := getResult.Err(); err != nil {
		result.Status = HealthStatusDegraded
		result.Message = fmt.Sprintf("Cache GET operation failed: %v", err)
		result.Error = err
		result.Duration = time.Since(start)
		return result
	}
	
	if getResult.Val() != testValue {
		result.Status = HealthStatusDegraded
		result.Message = "Cache GET operation returned unexpected value"
		result.Duration = time.Since(start)
		return result
	}
	
	// 检查TTL
	ttlResult := c.redisClient.TTL(ctx, testKey)
	if err := ttlResult.Err(); err != nil {
		result.Status = HealthStatusDegraded
		result.Message = fmt.Sprintf("Cache TTL check failed: %v", err)
		result.Error = err
		result.Duration = time.Since(start)
		return result
	}
	
	ttl := ttlResult.Val()
	result.Details["test_key_ttl"] = ttl.Seconds()
	
	// 删除测试键
	delResult := c.redisClient.Del(ctx, testKey)
	if err := delResult.Err(); err != nil {
		result.Status = HealthStatusDegraded
		result.Message = fmt.Sprintf("Cache DEL operation failed: %v", err)
		result.Error = err
		result.Duration = time.Since(start)
		return result
	}
	
	// 获取内存使用情况
	memoryResult := c.redisClient.Info(ctx, "memory")
	if err := memoryResult.Err(); err == nil {
		result.Details["memory_info"] = memoryResult.Val()
	}
	
	result.Status = HealthStatusHealthy
	result.Message = "Cache is healthy"
	result.Duration = time.Since(start)
	
	return result
}