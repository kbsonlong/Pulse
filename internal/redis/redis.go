package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"

	"Pulse/internal/config"
)

// Client Redis客户端包装器
type Client struct {
	rdb    *redis.Client
	logger *zap.Logger
	config *config.RedisConfig
}

// New 创建新的Redis客户端
func New(cfg *config.RedisConfig, logger *zap.Logger) (*Client, error) {
	if cfg == nil {
		return nil, fmt.Errorf("redis config is required")
	}

	if logger == nil {
		return nil, fmt.Errorf("logger is required")
	}

	// 创建Redis客户端
	rdb := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password:     cfg.Password,
		DB:           cfg.DB,
		PoolSize:     cfg.PoolSize,
		MinIdleConns: cfg.MinIdleConns,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolTimeout:  4 * time.Second,
		IdleTimeout:  5 * time.Minute,
	})

	client := &Client{
		rdb:    rdb,
		logger: logger,
		config: cfg,
	}

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	logger.Info("Redis client connected successfully",
		zap.String("host", cfg.Host),
		zap.Int("port", cfg.Port),
		zap.Int("db", cfg.DB),
		zap.Int("pool_size", cfg.PoolSize),
	)

	return client, nil
}

// Close 关闭Redis连接
func (c *Client) Close() error {
	if c.rdb != nil {
		err := c.rdb.Close()
		if err != nil {
			c.logger.Error("Failed to close redis connection", zap.Error(err))
			return err
		}
		c.logger.Info("Redis connection closed")
	}
	return nil
}

// Ping 测试Redis连接
func (c *Client) Ping(ctx context.Context) error {
	result := c.rdb.Ping(ctx)
	if result.Err() != nil {
		return fmt.Errorf("redis ping failed: %w", result.Err())
	}
	return nil
}

// Health 获取Redis健康状态
func (c *Client) Health(ctx context.Context) map[string]interface{} {
	stats := c.rdb.PoolStats()
	status := "healthy"

	// 检查连接状态
	if err := c.Ping(ctx); err != nil {
		status = "unhealthy"
	}

	return map[string]interface{}{
		"status":         status,
		"host":           c.config.Host,
		"port":           c.config.Port,
		"db":             c.config.DB,
		"pool_size":      c.config.PoolSize,
		"min_idle_conns": c.config.MinIdleConns,
		"hits":           stats.Hits,
		"misses":         stats.Misses,
		"timeouts":       stats.Timeouts,
		"total_conns":    stats.TotalConns,
		"idle_conns":     stats.IdleConns,
		"stale_conns":    stats.StaleConns,
	}
}

// GetClient 获取原始Redis客户端
func (c *Client) GetClient() *redis.Client {
	return c.rdb
}

// Set 设置键值对
func (c *Client) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return c.rdb.Set(ctx, key, value, expiration).Err()
}

// Get 获取值
func (c *Client) Get(ctx context.Context, key string) (string, error) {
	return c.rdb.Get(ctx, key).Result()
}

// Del 删除键
func (c *Client) Del(ctx context.Context, keys ...string) error {
	return c.rdb.Del(ctx, keys...).Err()
}

// Exists 检查键是否存在
func (c *Client) Exists(ctx context.Context, keys ...string) (int64, error) {
	return c.rdb.Exists(ctx, keys...).Result()
}

// Expire 设置键的过期时间
func (c *Client) Expire(ctx context.Context, key string, expiration time.Duration) error {
	return c.rdb.Expire(ctx, key, expiration).Err()
}

// TTL 获取键的剩余生存时间
func (c *Client) TTL(ctx context.Context, key string) (time.Duration, error) {
	return c.rdb.TTL(ctx, key).Result()
}

// HSet 设置哈希字段
func (c *Client) HSet(ctx context.Context, key string, values ...interface{}) error {
	return c.rdb.HSet(ctx, key, values...).Err()
}

// HGet 获取哈希字段值
func (c *Client) HGet(ctx context.Context, key, field string) (string, error) {
	return c.rdb.HGet(ctx, key, field).Result()
}

// HGetAll 获取哈希的所有字段和值
func (c *Client) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	return c.rdb.HGetAll(ctx, key).Result()
}

// HDel 删除哈希字段
func (c *Client) HDel(ctx context.Context, key string, fields ...string) error {
	return c.rdb.HDel(ctx, key, fields...).Err()
}

// LPush 从列表左侧推入元素
func (c *Client) LPush(ctx context.Context, key string, values ...interface{}) error {
	return c.rdb.LPush(ctx, key, values...).Err()
}

// RPush 从列表右侧推入元素
func (c *Client) RPush(ctx context.Context, key string, values ...interface{}) error {
	return c.rdb.RPush(ctx, key, values...).Err()
}

// LPop 从列表左侧弹出元素
func (c *Client) LPop(ctx context.Context, key string) (string, error) {
	return c.rdb.LPop(ctx, key).Result()
}

// RPop 从列表右侧弹出元素
func (c *Client) RPop(ctx context.Context, key string) (string, error) {
	return c.rdb.RPop(ctx, key).Result()
}

// LLen 获取列表长度
func (c *Client) LLen(ctx context.Context, key string) (int64, error) {
	return c.rdb.LLen(ctx, key).Result()
}

// SAdd 向集合添加成员
func (c *Client) SAdd(ctx context.Context, key string, members ...interface{}) error {
	return c.rdb.SAdd(ctx, key, members...).Err()
}

// SMembers 获取集合的所有成员
func (c *Client) SMembers(ctx context.Context, key string) ([]string, error) {
	return c.rdb.SMembers(ctx, key).Result()
}

// SRem 从集合移除成员
func (c *Client) SRem(ctx context.Context, key string, members ...interface{}) error {
	return c.rdb.SRem(ctx, key, members...).Err()
}

// ZAdd 向有序集合添加成员
func (c *Client) ZAdd(ctx context.Context, key string, members ...*redis.Z) error {
	return c.rdb.ZAdd(ctx, key, members...).Err()
}

// ZRange 获取有序集合指定范围的成员
func (c *Client) ZRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	return c.rdb.ZRange(ctx, key, start, stop).Result()
}

// ZRem 从有序集合移除成员
func (c *Client) ZRem(ctx context.Context, key string, members ...interface{}) error {
	return c.rdb.ZRem(ctx, key, members...).Err()
}

// LRem 从列表中移除元素
func (c *Client) LRem(ctx context.Context, key string, count int64, value interface{}) error {
	return c.rdb.LRem(ctx, key, count, value).Err()
}
