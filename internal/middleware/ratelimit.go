package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// RedisRateLimiter Redis分布式限流器
type RedisRateLimiter struct {
	client *redis.Client
	logger *zap.Logger
}

// NewRedisRateLimiter 创建Redis限流器
func NewRedisRateLimiter(client *redis.Client, logger *zap.Logger) *RedisRateLimiter {
	return &RedisRateLimiter{
		client: client,
		logger: logger,
	}
}

// RedisRateLimitConfig Redis限流配置
type RedisRateLimitConfig struct {
	RequestsPerMinute int           // 每分钟请求数
	BurstSize         int           // 突发大小
	Window            time.Duration // 时间窗口
	KeyPrefix         string        // Redis键前缀
}

// DefaultRedisRateLimitConfig 默认Redis限流配置
func DefaultRedisRateLimitConfig() RedisRateLimitConfig {
	return RedisRateLimitConfig{
		RequestsPerMinute: 60,
		BurstSize:         10,
		Window:            time.Minute,
		KeyPrefix:         "rate_limit:",
	}
}

// RedisRateLimitMiddleware Redis分布式限流中间件
func (r *RedisRateLimiter) RedisRateLimitMiddleware(config RedisRateLimitConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		userID := c.GetString("user_id")
		
		// 构建限流键
		key := config.KeyPrefix + clientIP
		if userID != "" {
			key = config.KeyPrefix + "user:" + userID
		}

		ctx := context.Background()
		now := time.Now()

		// 使用滑动窗口算法
		allowed, remaining, resetTime, err := r.slidingWindowRateLimit(ctx, key, config, now)
		if err != nil {
			r.logger.Error("Rate limit check failed",
				zap.String("key", key),
				zap.Error(err),
				zap.String("request_id", c.GetString("request_id")),
			)
			// 限流检查失败时，允许请求通过（fail-open策略）
			c.Next()
			return
		}

		// 设置限流响应头
		c.Header("X-RateLimit-Limit", strconv.Itoa(config.RequestsPerMinute))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))
		c.Header("X-RateLimit-Reset", strconv.FormatInt(resetTime.Unix(), 10))
		c.Header("X-RateLimit-Window", config.Window.String())

		if !allowed {
			r.logger.Warn("Rate limit exceeded",
				zap.String("key", key),
				zap.String("client_ip", clientIP),
				zap.String("user_id", userID),
				zap.Int("limit", config.RequestsPerMinute),
				zap.String("request_id", c.GetString("request_id")),
			)

			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":      "Rate limit exceeded",
				"limit":      config.RequestsPerMinute,
				"remaining":  remaining,
				"reset_time": resetTime.Unix(),
				"window":     config.Window.String(),
				"request_id": c.GetString("request_id"),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// slidingWindowRateLimit 滑动窗口限流算法
func (r *RedisRateLimiter) slidingWindowRateLimit(ctx context.Context, key string, config RedisRateLimitConfig, now time.Time) (bool, int, time.Time, error) {
	// 使用Redis的ZSET实现滑动窗口
	pipe := r.client.Pipeline()

	// 移除过期的请求记录
	expiredBefore := now.Add(-config.Window).UnixNano()
	pipe.ZRemRangeByScore(ctx, key, "0", strconv.FormatInt(expiredBefore, 10))

	// 获取当前窗口内的请求数
	countCmd := pipe.ZCard(ctx, key)

	// 添加当前请求
	currentTime := now.UnixNano()
	pipe.ZAdd(ctx, key, redis.Z{
		Score:  float64(currentTime),
		Member: fmt.Sprintf("%d", currentTime),
	})

	// 设置键的过期时间
	pipe.Expire(ctx, key, config.Window+time.Minute)

	// 执行管道
	_, err := pipe.Exec(ctx)
	if err != nil {
		return false, 0, now, err
	}

	// 获取当前请求数
	currentCount, err := countCmd.Result()
	if err != nil {
		return false, 0, now, err
	}

	// 检查是否超过限制
	allowed := int(currentCount) <= config.RequestsPerMinute
	remaining := config.RequestsPerMinute - int(currentCount)
	if remaining < 0 {
		remaining = 0
	}

	// 计算重置时间（下一个窗口开始时间）
	resetTime := now.Add(config.Window)

	return allowed, remaining, resetTime, nil
}

// CircuitBreakerState 熔断器状态
type CircuitBreakerState int

const (
	StateClosed CircuitBreakerState = iota
	StateHalfOpen
	StateOpen
)

// CircuitBreakerConfig 熔断器配置
type CircuitBreakerConfig struct {
	FailureThreshold   int           // 失败阈值
	SuccessThreshold   int           // 成功阈值（半开状态）
	Timeout            time.Duration // 熔断超时时间
	MonitoringPeriod   time.Duration // 监控周期
	KeyPrefix          string        // Redis键前缀
}

// DefaultCircuitBreakerConfig 默认熔断器配置
func DefaultCircuitBreakerConfig() CircuitBreakerConfig {
	return CircuitBreakerConfig{
		FailureThreshold:  5,
		SuccessThreshold:  3,
		Timeout:           time.Minute * 5,
		MonitoringPeriod:  time.Minute,
		KeyPrefix:         "circuit_breaker:",
	}
}

// CircuitBreakerMiddleware 熔断器中间件
func (r *RedisRateLimiter) CircuitBreakerMiddleware(serviceName string, config CircuitBreakerConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := context.Background()
		key := config.KeyPrefix + serviceName

		// 检查熔断器状态
		state, err := r.getCircuitBreakerState(ctx, key, config)
		if err != nil {
			r.logger.Error("Circuit breaker state check failed",
				zap.String("service", serviceName),
				zap.Error(err),
				zap.String("request_id", c.GetString("request_id")),
			)
			// 检查失败时，允许请求通过
			c.Next()
			return
		}

		// 如果熔断器开启，拒绝请求
		if state == StateOpen {
			r.logger.Warn("Circuit breaker is open",
				zap.String("service", serviceName),
				zap.String("request_id", c.GetString("request_id")),
			)

			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error":      "Service temporarily unavailable",
				"service":    serviceName,
				"state":      "open",
				"request_id": c.GetString("request_id"),
			})
			c.Abort()
			return
		}

		// 记录请求开始时间
		start := time.Now()

		// 处理请求
		c.Next()

		// 记录请求结果
		duration := time.Since(start)
		success := c.Writer.Status() < 500

		// 更新熔断器统计
		err = r.updateCircuitBreakerStats(ctx, key, config, success, duration)
		if err != nil {
			r.logger.Error("Failed to update circuit breaker stats",
				zap.String("service", serviceName),
				zap.Error(err),
				zap.String("request_id", c.GetString("request_id")),
			)
		}
	}
}

// getCircuitBreakerState 获取熔断器状态
func (r *RedisRateLimiter) getCircuitBreakerState(ctx context.Context, key string, config CircuitBreakerConfig) (CircuitBreakerState, error) {
	now := time.Now()

	// 获取熔断器数据
	pipe := r.client.Pipeline()
	stateCmd := pipe.HGet(ctx, key, "state")
	lastFailureCmd := pipe.HGet(ctx, key, "last_failure")
	failureCountCmd := pipe.HGet(ctx, key, "failure_count")
	successCountCmd := pipe.HGet(ctx, key, "success_count")

	_, err := pipe.Exec(ctx)
	if err != nil && err != redis.Nil {
		return StateClosed, err
	}

	// 解析状态
	stateStr, _ := stateCmd.Result()
	lastFailureStr, _ := lastFailureCmd.Result()
	failureCountStr, _ := failureCountCmd.Result()
	successCountStr, _ := successCountCmd.Result()

	// 默认状态为关闭
	state := StateClosed
	if stateStr != "" {
		switch stateStr {
		case "open":
			state = StateOpen
		case "half_open":
			state = StateHalfOpen
		}
	}

	// 如果是开启状态，检查是否应该转为半开状态
	if state == StateOpen && lastFailureStr != "" {
		lastFailure, err := strconv.ParseInt(lastFailureStr, 10, 64)
		if err == nil {
			lastFailureTime := time.Unix(lastFailure, 0)
			if now.Sub(lastFailureTime) >= config.Timeout {
				// 转为半开状态
				state = StateHalfOpen
				r.client.HSet(ctx, key, "state", "half_open")
				r.client.HSet(ctx, key, "success_count", "0")
			}
		}
	}

	// 如果是半开状态，检查成功次数
	if state == StateHalfOpen && successCountStr != "" {
		successCount, err := strconv.Atoi(successCountStr)
		if err == nil && successCount >= config.SuccessThreshold {
			// 转为关闭状态
			state = StateClosed
			r.client.HSet(ctx, key, "state", "closed")
			r.client.HSet(ctx, key, "failure_count", "0")
			r.client.HSet(ctx, key, "success_count", "0")
		}
	}

	// 检查失败次数是否达到阈值
	if state == StateClosed && failureCountStr != "" {
		failureCount, err := strconv.Atoi(failureCountStr)
		if err == nil && failureCount >= config.FailureThreshold {
			// 转为开启状态
			state = StateOpen
			r.client.HSet(ctx, key, "state", "open")
			r.client.HSet(ctx, key, "last_failure", strconv.FormatInt(now.Unix(), 10))
		}
	}

	return state, nil
}

// updateCircuitBreakerStats 更新熔断器统计
func (r *RedisRateLimiter) updateCircuitBreakerStats(ctx context.Context, key string, config CircuitBreakerConfig, success bool, duration time.Duration) error {
	now := time.Now()

	if success {
		// 成功请求
		pipe := r.client.Pipeline()
		pipe.HIncrBy(ctx, key, "success_count", 1)
		pipe.HSet(ctx, key, "last_success", strconv.FormatInt(now.Unix(), 10))
		pipe.Expire(ctx, key, config.MonitoringPeriod*10) // 保留更长时间的统计数据
		_, err := pipe.Exec(ctx)
		return err
	} else {
		// 失败请求
		pipe := r.client.Pipeline()
		pipe.HIncrBy(ctx, key, "failure_count", 1)
		pipe.HSet(ctx, key, "last_failure", strconv.FormatInt(now.Unix(), 10))
		pipe.Expire(ctx, key, config.MonitoringPeriod*10)
		_, err := pipe.Exec(ctx)
		return err
	}
}