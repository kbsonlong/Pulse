package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

// RateLimitConfig 限流配置
type RateLimitConfig struct {
	RedisClient    *redis.Client
	Logger         *logrus.Logger
	KeyPrefix      string        // Redis键前缀
	DefaultLimit   int           // 默认限制次数
	DefaultWindow  time.Duration // 默认时间窗口
	SkipSuccessful bool          // 是否跳过成功请求的计数
	KeyGenerator   func(*gin.Context) string // 自定义键生成器
}

// DefaultRateLimitConfig 默认限流配置
func DefaultRateLimitConfig(redisClient *redis.Client) RateLimitConfig {
	return RateLimitConfig{
		RedisClient:    redisClient,
		KeyPrefix:      "rate_limit:",
		DefaultLimit:   100,
		DefaultWindow:  time.Minute,
		SkipSuccessful: false,
		KeyGenerator: func(c *gin.Context) string {
			return c.ClientIP()
		},
	}
}

// RateLimitMiddleware 限流中间件
func RateLimitMiddleware(config RateLimitConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查是否跳过限流
		if c.GetBool("skip_rate_limit") {
			c.Next()
			return
		}

		// 生成限流键
		key := config.KeyPrefix + config.KeyGenerator(c)
		limit := config.DefaultLimit
		window := config.DefaultWindow

		// 检查限流
		allowed, remaining, resetTime, err := checkRateLimit(config.RedisClient, key, limit, window)
		if err != nil {
			if config.Logger != nil {
				config.Logger.WithFields(logrus.Fields{
					"error":      err,
					"key":        key,
					"request_id": c.GetString("request_id"),
				}).Error("Rate limit check failed")
			}
			// 限流检查失败时，允许请求通过（fail-open策略）
			c.Next()
			return
		}

		// 设置限流相关的响应头
		c.Header("X-RateLimit-Limit", strconv.Itoa(limit))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))
		c.Header("X-RateLimit-Reset", strconv.FormatInt(resetTime, 10))

		if !allowed {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "rate_limit_exceeded",
				"message": "Rate limit exceeded. Please try again later.",
				"retry_after": resetTime - time.Now().Unix(),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// checkRateLimit 检查限流状态
func checkRateLimit(redisClient *redis.Client, key string, limit int, window time.Duration) (allowed bool, remaining int, resetTime int64, err error) {
	ctx := context.Background()
	now := time.Now()
	windowStart := now.Truncate(window)
	resetTime = windowStart.Add(window).Unix()

	// 使用Lua脚本确保原子性
	luaScript := `
		local key = KEYS[1]
		local window_start = ARGV[1]
		local limit = tonumber(ARGV[2])
		local ttl = tonumber(ARGV[3])
		
		-- 清理过期的计数
		redis.call('ZREMRANGEBYSCORE', key, 0, window_start - 1)
		
		-- 获取当前计数
		local current = redis.call('ZCARD', key)
		
		if current < limit then
			-- 添加当前请求
			redis.call('ZADD', key, window_start, window_start)
			redis.call('EXPIRE', key, ttl)
			return {1, limit - current - 1}
		else
			return {0, 0}
		end
	`

	result, err := redisClient.Eval(ctx, luaScript, []string{key}, windowStart.Unix(), limit, int(window.Seconds())).Result()
	if err != nil {
		return false, 0, resetTime, err
	}

	results := result.([]interface{})
	allowed = results[0].(int64) == 1
	remaining = int(results[1].(int64))

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
	Name               string        // 熔断器名称
	MaxRequests        uint32        // 半开状态下的最大请求数
	Interval           time.Duration // 统计时间窗口
	Timeout            time.Duration // 熔断器打开后的超时时间
	ReadyToTrip        func(counts Counts) bool // 判断是否应该打开熔断器
	OnStateChange      func(name string, from CircuitBreakerState, to CircuitBreakerState) // 状态变化回调
	IsSuccessful       func(*gin.Context) bool // 判断请求是否成功
	ShouldTrip         func(counts Counts) bool // 自定义熔断条件
	FallbackResponse   func(*gin.Context)       // 熔断时的降级响应
}

// Counts 统计计数
type Counts struct {
	Requests         uint32
	TotalSuccesses   uint32
	TotalFailures    uint32
	ConsecutiveSuccesses uint32
	ConsecutiveFailures  uint32
}

// CircuitBreaker 熔断器
type CircuitBreaker struct {
	name         string
	maxRequests  uint32
	interval     time.Duration
	timeout      time.Duration
	readyToTrip  func(counts Counts) bool
	onStateChange func(name string, from CircuitBreakerState, to CircuitBreakerState)
	isSuccessful func(*gin.Context) bool
	fallbackResponse func(*gin.Context)

	mutex      sync.Mutex
	state      CircuitBreakerState
	generation uint64
	counts     Counts
	expiry     time.Time
}

// NewCircuitBreaker 创建新的熔断器
func NewCircuitBreaker(config CircuitBreakerConfig) *CircuitBreaker {
	cb := &CircuitBreaker{
		name:         config.Name,
		maxRequests:  config.MaxRequests,
		interval:     config.Interval,
		timeout:      config.Timeout,
		readyToTrip:  config.ReadyToTrip,
		onStateChange: config.OnStateChange,
		isSuccessful: config.IsSuccessful,
		fallbackResponse: config.FallbackResponse,
	}

	// 设置默认值
	if cb.maxRequests == 0 {
		cb.maxRequests = 1
	}
	if cb.interval == 0 {
		cb.interval = time.Minute
	}
	if cb.timeout == 0 {
		cb.timeout = time.Minute
	}
	if cb.readyToTrip == nil {
		cb.readyToTrip = func(counts Counts) bool {
			return counts.ConsecutiveFailures > 5
		}
	}
	if cb.isSuccessful == nil {
		cb.isSuccessful = func(c *gin.Context) bool {
			return c.Writer.Status() < 500
		}
	}
	if cb.fallbackResponse == nil {
		cb.fallbackResponse = func(c *gin.Context) {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error":   "service_unavailable",
				"message": "Service is temporarily unavailable. Please try again later.",
			})
		}
	}

	cb.toNewGeneration(time.Now())
	return cb
}

// Execute 执行请求
func (cb *CircuitBreaker) Execute(c *gin.Context) {
	generation, err := cb.beforeRequest()
	if err != nil {
		cb.fallbackResponse(c)
		c.Abort()
		return
	}

	c.Next()

	cb.afterRequest(generation, cb.isSuccessful(c))
}

// beforeRequest 请求前检查
func (cb *CircuitBreaker) beforeRequest() (uint64, error) {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	now := time.Now()
	state, generation := cb.currentState(now)

	if state == StateOpen {
		return generation, fmt.Errorf("circuit breaker is open")
	} else if state == StateHalfOpen && cb.counts.Requests >= cb.maxRequests {
		return generation, fmt.Errorf("too many requests")
	}

	cb.counts.onRequest()
	return generation, nil
}

// afterRequest 请求后处理
func (cb *CircuitBreaker) afterRequest(before uint64, success bool) {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	now := time.Now()
	state, generation := cb.currentState(now)
	if generation != before {
		return
	}

	if success {
		cb.onSuccess(state, now)
	} else {
		cb.onFailure(state, now)
	}
}

// currentState 获取当前状态
func (cb *CircuitBreaker) currentState(now time.Time) (CircuitBreakerState, uint64) {
	switch cb.state {
	case StateClosed:
		if !cb.expiry.IsZero() && cb.expiry.Before(now) {
			cb.toNewGeneration(now)
		}
	case StateOpen:
		if cb.expiry.Before(now) {
			cb.setState(StateHalfOpen, now)
		}
	}
	return cb.state, cb.generation
}

// onSuccess 成功处理
func (cb *CircuitBreaker) onSuccess(state CircuitBreakerState, now time.Time) {
	cb.counts.onSuccess()

	if state == StateHalfOpen {
		cb.setState(StateClosed, now)
	}
}

// onFailure 失败处理
func (cb *CircuitBreaker) onFailure(state CircuitBreakerState, now time.Time) {
	cb.counts.onFailure()

	if cb.readyToTrip(cb.counts) {
		cb.setState(StateOpen, now)
	}
}

// setState 设置状态
func (cb *CircuitBreaker) setState(state CircuitBreakerState, now time.Time) {
	if cb.state == state {
		return
	}

	prev := cb.state
	cb.state = state

	cb.toNewGeneration(now)

	if cb.onStateChange != nil {
		cb.onStateChange(cb.name, prev, state)
	}
}

// toNewGeneration 切换到新的统计周期
func (cb *CircuitBreaker) toNewGeneration(now time.Time) {
	cb.generation++
	cb.counts = Counts{}

	var zero time.Time
	switch cb.state {
	case StateClosed:
		if cb.interval == 0 {
			cb.expiry = zero
		} else {
			cb.expiry = now.Add(cb.interval)
		}
	case StateOpen:
		cb.expiry = now.Add(cb.timeout)
	default: // StateHalfOpen
		cb.expiry = zero
	}
}

// onRequest 请求计数
func (c *Counts) onRequest() {
	c.Requests++
}

// onSuccess 成功计数
func (c *Counts) onSuccess() {
	c.TotalSuccesses++
	c.ConsecutiveSuccesses++
	c.ConsecutiveFailures = 0
}

// onFailure 失败计数
func (c *Counts) onFailure() {
	c.TotalFailures++
	c.ConsecutiveFailures++
	c.ConsecutiveSuccesses = 0
}

// CircuitBreakerMiddleware 熔断器中间件
func CircuitBreakerMiddleware(cb *CircuitBreaker) gin.HandlerFunc {
	return func(c *gin.Context) {
		cb.Execute(c)
	}
}

// DefaultCircuitBreakerConfig 默认熔断器配置
func DefaultCircuitBreakerConfig(name string) CircuitBreakerConfig {
	return CircuitBreakerConfig{
		Name:        name,
		MaxRequests: 3,
		Interval:    time.Minute,
		Timeout:     time.Minute,
		ReadyToTrip: func(counts Counts) bool {
			failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
			return counts.Requests >= 3 && failureRatio >= 0.6
		},
		OnStateChange: func(name string, from CircuitBreakerState, to CircuitBreakerState) {
			// 可以在这里记录状态变化日志或发送告警
		},
	}
}