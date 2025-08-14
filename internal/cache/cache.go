package cache

import (
	"context"
	"time"
)

// Cache 缓存接口
type Cache interface {
	// Get 获取缓存值
	Get(ctx context.Context, key string) (string, error)
	
	// Set 设置缓存值
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	
	// Del 删除缓存
	Del(ctx context.Context, keys ...string) error
	
	// Exists 检查键是否存在
	Exists(ctx context.Context, keys ...string) (int64, error)
	
	// Expire 设置过期时间
	Expire(ctx context.Context, key string, ttl time.Duration) error
	
	// TTL 获取剩余过期时间
	TTL(ctx context.Context, key string) (time.Duration, error)
	
	// Incr 递增
	Incr(ctx context.Context, key string) (int64, error)
	
	// IncrBy 按指定值递增
	IncrBy(ctx context.Context, key string, value int64) (int64, error)
	
	// Decr 递减
	Decr(ctx context.Context, key string) (int64, error)
	
	// DecrBy 按指定值递减
	DecrBy(ctx context.Context, key string, value int64) (int64, error)
	
	// MGet 批量获取
	MGet(ctx context.Context, keys ...string) ([]interface{}, error)
	
	// MSet 批量设置
	MSet(ctx context.Context, pairs ...interface{}) error
	
	// FlushAll 清空所有缓存
	FlushAll(ctx context.Context) error
	
	// Close 关闭连接
	Close() error
	
	// Ping 健康检查
	Ping(ctx context.Context) error
}

// HashCache 哈希缓存接口
type HashCache interface {
	// HGet 获取哈希字段值
	HGet(ctx context.Context, key, field string) (string, error)
	
	// HSet 设置哈希字段值
	HSet(ctx context.Context, key string, values ...interface{}) error
	
	// HDel 删除哈希字段
	HDel(ctx context.Context, key string, fields ...string) error
	
	// HExists 检查哈希字段是否存在
	HExists(ctx context.Context, key, field string) (bool, error)
	
	// HGetAll 获取所有哈希字段
	HGetAll(ctx context.Context, key string) (map[string]string, error)
	
	// HKeys 获取所有哈希字段名
	HKeys(ctx context.Context, key string) ([]string, error)
	
	// HVals 获取所有哈希字段值
	HVals(ctx context.Context, key string) ([]string, error)
	
	// HLen 获取哈希字段数量
	HLen(ctx context.Context, key string) (int64, error)
	
	// HMGet 批量获取哈希字段值
	HMGet(ctx context.Context, key string, fields ...string) ([]interface{}, error)
	
	// HMSet 批量设置哈希字段值
	HMSet(ctx context.Context, key string, values ...interface{}) error
	
	// HIncrBy 哈希字段递增
	HIncrBy(ctx context.Context, key, field string, incr int64) (int64, error)
}

// ListCache 列表缓存接口
type ListCache interface {
	// LPush 从左侧推入
	LPush(ctx context.Context, key string, values ...interface{}) error
	
	// RPush 从右侧推入
	RPush(ctx context.Context, key string, values ...interface{}) error
	
	// LPop 从左侧弹出
	LPop(ctx context.Context, key string) (string, error)
	
	// RPop 从右侧弹出
	RPop(ctx context.Context, key string) (string, error)
	
	// LLen 获取列表长度
	LLen(ctx context.Context, key string) (int64, error)
	
	// LRange 获取列表范围
	LRange(ctx context.Context, key string, start, stop int64) ([]string, error)
	
	// LIndex 获取列表指定位置元素
	LIndex(ctx context.Context, key string, index int64) (string, error)
	
	// LSet 设置列表指定位置元素
	LSet(ctx context.Context, key string, index int64, value interface{}) error
	
	// LTrim 修剪列表
	LTrim(ctx context.Context, key string, start, stop int64) error
}

// SetCache 集合缓存接口
type SetCache interface {
	// SAdd 添加集合成员
	SAdd(ctx context.Context, key string, members ...interface{}) error
	
	// SRem 删除集合成员
	SRem(ctx context.Context, key string, members ...interface{}) error
	
	// SMembers 获取所有集合成员
	SMembers(ctx context.Context, key string) ([]string, error)
	
	// SIsMember 检查是否为集合成员
	SIsMember(ctx context.Context, key string, member interface{}) (bool, error)
	
	// SCard 获取集合成员数量
	SCard(ctx context.Context, key string) (int64, error)
	
	// SPop 随机弹出集合成员
	SPop(ctx context.Context, key string) (string, error)
	
	// SRandMember 随机获取集合成员
	SRandMember(ctx context.Context, key string) (string, error)
}

// ZSetCache 有序集合缓存接口
type ZSetCache interface {
	// ZAdd 添加有序集合成员
	ZAdd(ctx context.Context, key string, members ...interface{}) error
	
	// ZRem 删除有序集合成员
	ZRem(ctx context.Context, key string, members ...interface{}) error
	
	// ZRange 获取有序集合范围
	ZRange(ctx context.Context, key string, start, stop int64) ([]string, error)
	
	// ZRangeWithScores 获取有序集合范围（带分数）
	ZRangeWithScores(ctx context.Context, key string, start, stop int64) ([]interface{}, error)
	
	// ZRangeByScore 按分数范围获取有序集合
	ZRangeByScore(ctx context.Context, key string, min, max string) ([]string, error)
	
	// ZCard 获取有序集合成员数量
	ZCard(ctx context.Context, key string) (int64, error)
	
	// ZScore 获取成员分数
	ZScore(ctx context.Context, key, member string) (float64, error)
	
	// ZRank 获取成员排名
	ZRank(ctx context.Context, key, member string) (int64, error)
}

// CacheOptions 缓存选项
type CacheOptions struct {
	// Prefix 键前缀
	Prefix string
	
	// KeyPrefix 键前缀（别名）
	KeyPrefix string
	
	// DefaultTTL 默认过期时间
	DefaultTTL time.Duration
	
	// Serializer 序列化器
	Serializer Serializer
	
	// Namespace 命名空间
	Namespace string
}

// CacheOption 缓存选项函数
type CacheOption func(*CacheOptions)

// WithPrefix 设置键前缀
func WithPrefix(prefix string) CacheOption {
	return func(opts *CacheOptions) {
		opts.Prefix = prefix
	}
}

// WithDefaultTTL 设置默认过期时间
func WithDefaultTTL(ttl time.Duration) CacheOption {
	return func(opts *CacheOptions) {
		opts.DefaultTTL = ttl
	}
}

// WithSerializer 设置序列化器
func WithSerializer(serializer Serializer) CacheOption {
	return func(opts *CacheOptions) {
		opts.Serializer = serializer
	}
}

// WithNamespace 设置命名空间
func WithNamespace(namespace string) CacheOption {
	return func(opts *CacheOptions) {
		opts.Namespace = namespace
	}
}

// Serializer 序列化器接口
type Serializer interface {
	// Serialize 序列化
	Serialize(v interface{}) ([]byte, error)
	
	// Deserialize 反序列化
	Deserialize(data []byte, v interface{}) error
}

// applyCacheOptions 应用缓存选项
func applyCacheOptions(opts ...CacheOption) *CacheOptions {
	options := &CacheOptions{
		Prefix:     "cache:",
		DefaultTTL: 1 * time.Hour,
		Serializer: &JSONSerializer{},
		Namespace:  "default",
	}
	
	for _, opt := range opts {
		opt(options)
	}
	
	return options
}