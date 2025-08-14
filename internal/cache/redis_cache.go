package cache

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
)

// RedisCache Redis缓存实现
type RedisCache struct {
	client *redis.Client
	opts   *CacheOptions
}

// NewRedisCache 创建Redis缓存
func NewRedisCache(client *redis.Client, opts ...CacheOption) *RedisCache {
	options := applyCacheOptions(opts...)
	return &RedisCache{
		client: client,
		opts:   options,
	}
}

// buildKey 构建缓存键
func (r *RedisCache) buildKey(key string) string {
	if r.opts.Namespace != "" && r.opts.Prefix != "" {
		return fmt.Sprintf("%s:%s:%s", r.opts.Namespace, r.opts.Prefix, key)
	} else if r.opts.Namespace != "" {
		return fmt.Sprintf("%s:%s", r.opts.Namespace, key)
	} else if r.opts.Prefix != "" {
		return fmt.Sprintf("%s%s", r.opts.Prefix, key)
	}
	return key
}

// Get 获取缓存值
func (r *RedisCache) Get(ctx context.Context, key string) (string, error) {
	cacheKey := r.buildKey(key)
	result, err := r.client.Get(ctx, cacheKey).Result()
	if err != nil {
		if err == redis.Nil {
			return "", nil
		}
		return "", fmt.Errorf("failed to get cache: %w", err)
	}
	return result, nil
}

// Set 设置缓存值
func (r *RedisCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	cacheKey := r.buildKey(key)
	
	// 序列化值
	data, err := r.opts.Serializer.Serialize(value)
	if err != nil {
		return fmt.Errorf("failed to serialize value: %w", err)
	}
	
	// 使用默认TTL如果未指定
	if ttl <= 0 {
		ttl = r.opts.DefaultTTL
	}
	
	err = r.client.Set(ctx, cacheKey, data, ttl).Err()
	if err != nil {
		return fmt.Errorf("failed to set cache: %w", err)
	}
	
	return nil
}

// Del 删除缓存
func (r *RedisCache) Del(ctx context.Context, keys ...string) error {
	cacheKeys := make([]string, len(keys))
	for i, key := range keys {
		cacheKeys[i] = r.buildKey(key)
	}
	
	err := r.client.Del(ctx, cacheKeys...).Err()
	if err != nil {
		return fmt.Errorf("failed to delete cache: %w", err)
	}
	
	return nil
}

// Exists 检查键是否存在
func (r *RedisCache) Exists(ctx context.Context, keys ...string) (int64, error) {
	cacheKeys := make([]string, len(keys))
	for i, key := range keys {
		cacheKeys[i] = r.buildKey(key)
	}
	
	result, err := r.client.Exists(ctx, cacheKeys...).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to check existence: %w", err)
	}
	
	return result, nil
}

// Expire 设置过期时间
func (r *RedisCache) Expire(ctx context.Context, key string, ttl time.Duration) error {
	cacheKey := r.buildKey(key)
	err := r.client.Expire(ctx, cacheKey, ttl).Err()
	if err != nil {
		return fmt.Errorf("failed to set expiration: %w", err)
	}
	return nil
}

// TTL 获取剩余过期时间
func (r *RedisCache) TTL(ctx context.Context, key string) (time.Duration, error) {
	cacheKey := r.buildKey(key)
	result, err := r.client.TTL(ctx, cacheKey).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get TTL: %w", err)
	}
	return result, nil
}

// Incr 递增
func (r *RedisCache) Incr(ctx context.Context, key string) (int64, error) {
	cacheKey := r.buildKey(key)
	result, err := r.client.Incr(ctx, cacheKey).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to increment: %w", err)
	}
	return result, nil
}

// IncrBy 按指定值递增
func (r *RedisCache) IncrBy(ctx context.Context, key string, value int64) (int64, error) {
	cacheKey := r.buildKey(key)
	result, err := r.client.IncrBy(ctx, cacheKey, value).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to increment by value: %w", err)
	}
	return result, nil
}

// Decr 递减
func (r *RedisCache) Decr(ctx context.Context, key string) (int64, error) {
	cacheKey := r.buildKey(key)
	result, err := r.client.Decr(ctx, cacheKey).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to decrement: %w", err)
	}
	return result, nil
}

// DecrBy 按指定值递减
func (r *RedisCache) DecrBy(ctx context.Context, key string, value int64) (int64, error) {
	cacheKey := r.buildKey(key)
	result, err := r.client.DecrBy(ctx, cacheKey, value).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to decrement by value: %w", err)
	}
	return result, nil
}

// MGet 批量获取
func (r *RedisCache) MGet(ctx context.Context, keys ...string) ([]interface{}, error) {
	cacheKeys := make([]string, len(keys))
	for i, key := range keys {
		cacheKeys[i] = r.buildKey(key)
	}
	
	result, err := r.client.MGet(ctx, cacheKeys...).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get multiple values: %w", err)
	}
	
	return result, nil
}

// MSet 批量设置
func (r *RedisCache) MSet(ctx context.Context, pairs ...interface{}) error {
	if len(pairs)%2 != 0 {
		return fmt.Errorf("pairs must be even number")
	}
	
	// 构建键值对
	cachePairs := make([]interface{}, len(pairs))
	for i := 0; i < len(pairs); i += 2 {
		key, ok := pairs[i].(string)
		if !ok {
			return fmt.Errorf("key must be string")
		}
		cachePairs[i] = r.buildKey(key)
		
		// 序列化值
		data, err := r.opts.Serializer.Serialize(pairs[i+1])
		if err != nil {
			return fmt.Errorf("failed to serialize value: %w", err)
		}
		cachePairs[i+1] = data
	}
	
	err := r.client.MSet(ctx, cachePairs...).Err()
	if err != nil {
		return fmt.Errorf("failed to set multiple values: %w", err)
	}
	
	return nil
}

// FlushAll 清空所有缓存
func (r *RedisCache) FlushAll(ctx context.Context) error {
	err := r.client.FlushAll(ctx).Err()
	if err != nil {
		return fmt.Errorf("failed to flush all: %w", err)
	}
	return nil
}

// Close 关闭连接
func (r *RedisCache) Close() error {
	return r.client.Close()
}

// Ping 健康检查
func (r *RedisCache) Ping(ctx context.Context) error {
	err := r.client.Ping(ctx).Err()
	if err != nil {
		return fmt.Errorf("redis ping failed: %w", err)
	}
	return nil
}

// Hash操作实现

// HGet 获取哈希字段值
func (r *RedisCache) HGet(ctx context.Context, key, field string) (string, error) {
	cacheKey := r.buildKey(key)
	result, err := r.client.HGet(ctx, cacheKey, field).Result()
	if err != nil {
		if err == redis.Nil {
			return "", nil
		}
		return "", fmt.Errorf("failed to get hash field: %w", err)
	}
	return result, nil
}

// HSet 设置哈希字段值
func (r *RedisCache) HSet(ctx context.Context, key string, values ...interface{}) error {
	cacheKey := r.buildKey(key)
	err := r.client.HSet(ctx, cacheKey, values...).Err()
	if err != nil {
		return fmt.Errorf("failed to set hash field: %w", err)
	}
	return nil
}

// HDel 删除哈希字段
func (r *RedisCache) HDel(ctx context.Context, key string, fields ...string) error {
	cacheKey := r.buildKey(key)
	err := r.client.HDel(ctx, cacheKey, fields...).Err()
	if err != nil {
		return fmt.Errorf("failed to delete hash fields: %w", err)
	}
	return nil
}

// HExists 检查哈希字段是否存在
func (r *RedisCache) HExists(ctx context.Context, key, field string) (bool, error) {
	cacheKey := r.buildKey(key)
	result, err := r.client.HExists(ctx, cacheKey, field).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check hash field existence: %w", err)
	}
	return result, nil
}

// HGetAll 获取所有哈希字段
func (r *RedisCache) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	cacheKey := r.buildKey(key)
	result, err := r.client.HGetAll(ctx, cacheKey).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get all hash fields: %w", err)
	}
	return result, nil
}

// HKeys 获取所有哈希字段名
func (r *RedisCache) HKeys(ctx context.Context, key string) ([]string, error) {
	cacheKey := r.buildKey(key)
	result, err := r.client.HKeys(ctx, cacheKey).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get hash keys: %w", err)
	}
	return result, nil
}

// HVals 获取所有哈希字段值
func (r *RedisCache) HVals(ctx context.Context, key string) ([]string, error) {
	cacheKey := r.buildKey(key)
	result, err := r.client.HVals(ctx, cacheKey).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get hash values: %w", err)
	}
	return result, nil
}

// HLen 获取哈希字段数量
func (r *RedisCache) HLen(ctx context.Context, key string) (int64, error) {
	cacheKey := r.buildKey(key)
	result, err := r.client.HLen(ctx, cacheKey).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get hash length: %w", err)
	}
	return result, nil
}

// HMGet 批量获取哈希字段值
func (r *RedisCache) HMGet(ctx context.Context, key string, fields ...string) ([]interface{}, error) {
	cacheKey := r.buildKey(key)
	result, err := r.client.HMGet(ctx, cacheKey, fields...).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get multiple hash fields: %w", err)
	}
	return result, nil
}

// HMSet 批量设置哈希字段值
func (r *RedisCache) HMSet(ctx context.Context, key string, values ...interface{}) error {
	cacheKey := r.buildKey(key)
	err := r.client.HMSet(ctx, cacheKey, values...).Err()
	if err != nil {
		return fmt.Errorf("failed to set multiple hash fields: %w", err)
	}
	return nil
}

// HIncrBy 哈希字段递增
func (r *RedisCache) HIncrBy(ctx context.Context, key, field string, incr int64) (int64, error) {
	cacheKey := r.buildKey(key)
	result, err := r.client.HIncrBy(ctx, cacheKey, field, incr).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to increment hash field: %w", err)
	}
	return result, nil
}

// List操作实现

// LPush 从左侧推入
func (r *RedisCache) LPush(ctx context.Context, key string, values ...interface{}) error {
	cacheKey := r.buildKey(key)
	err := r.client.LPush(ctx, cacheKey, values...).Err()
	if err != nil {
		return fmt.Errorf("failed to left push: %w", err)
	}
	return nil
}

// RPush 从右侧推入
func (r *RedisCache) RPush(ctx context.Context, key string, values ...interface{}) error {
	cacheKey := r.buildKey(key)
	err := r.client.RPush(ctx, cacheKey, values...).Err()
	if err != nil {
		return fmt.Errorf("failed to right push: %w", err)
	}
	return nil
}

// LPop 从左侧弹出
func (r *RedisCache) LPop(ctx context.Context, key string) (string, error) {
	cacheKey := r.buildKey(key)
	result, err := r.client.LPop(ctx, cacheKey).Result()
	if err != nil {
		if err == redis.Nil {
			return "", nil
		}
		return "", fmt.Errorf("failed to left pop: %w", err)
	}
	return result, nil
}

// RPop 从右侧弹出
func (r *RedisCache) RPop(ctx context.Context, key string) (string, error) {
	cacheKey := r.buildKey(key)
	result, err := r.client.RPop(ctx, cacheKey).Result()
	if err != nil {
		if err == redis.Nil {
			return "", nil
		}
		return "", fmt.Errorf("failed to right pop: %w", err)
	}
	return result, nil
}

// LLen 获取列表长度
func (r *RedisCache) LLen(ctx context.Context, key string) (int64, error) {
	cacheKey := r.buildKey(key)
	result, err := r.client.LLen(ctx, cacheKey).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get list length: %w", err)
	}
	return result, nil
}

// LRange 获取列表范围
func (r *RedisCache) LRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	cacheKey := r.buildKey(key)
	result, err := r.client.LRange(ctx, cacheKey, start, stop).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get list range: %w", err)
	}
	return result, nil
}

// LIndex 获取列表指定位置元素
func (r *RedisCache) LIndex(ctx context.Context, key string, index int64) (string, error) {
	cacheKey := r.buildKey(key)
	result, err := r.client.LIndex(ctx, cacheKey, index).Result()
	if err != nil {
		if err == redis.Nil {
			return "", nil
		}
		return "", fmt.Errorf("failed to get list index: %w", err)
	}
	return result, nil
}

// LSet 设置列表指定位置元素
func (r *RedisCache) LSet(ctx context.Context, key string, index int64, value interface{}) error {
	cacheKey := r.buildKey(key)
	err := r.client.LSet(ctx, cacheKey, index, value).Err()
	if err != nil {
		return fmt.Errorf("failed to set list index: %w", err)
	}
	return nil
}

// LTrim 修剪列表
func (r *RedisCache) LTrim(ctx context.Context, key string, start, stop int64) error {
	cacheKey := r.buildKey(key)
	err := r.client.LTrim(ctx, cacheKey, start, stop).Err()
	if err != nil {
		return fmt.Errorf("failed to trim list: %w", err)
	}
	return nil
}

// Set操作实现

// SAdd 添加集合成员
func (r *RedisCache) SAdd(ctx context.Context, key string, members ...interface{}) error {
	cacheKey := r.buildKey(key)
	err := r.client.SAdd(ctx, cacheKey, members...).Err()
	if err != nil {
		return fmt.Errorf("failed to add set members: %w", err)
	}
	return nil
}

// SRem 删除集合成员
func (r *RedisCache) SRem(ctx context.Context, key string, members ...interface{}) error {
	cacheKey := r.buildKey(key)
	err := r.client.SRem(ctx, cacheKey, members...).Err()
	if err != nil {
		return fmt.Errorf("failed to remove set members: %w", err)
	}
	return nil
}

// SMembers 获取所有集合成员
func (r *RedisCache) SMembers(ctx context.Context, key string) ([]string, error) {
	cacheKey := r.buildKey(key)
	result, err := r.client.SMembers(ctx, cacheKey).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get set members: %w", err)
	}
	return result, nil
}

// SIsMember 检查是否为集合成员
func (r *RedisCache) SIsMember(ctx context.Context, key string, member interface{}) (bool, error) {
	cacheKey := r.buildKey(key)
	result, err := r.client.SIsMember(ctx, cacheKey, member).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check set membership: %w", err)
	}
	return result, nil
}

// SCard 获取集合成员数量
func (r *RedisCache) SCard(ctx context.Context, key string) (int64, error) {
	cacheKey := r.buildKey(key)
	result, err := r.client.SCard(ctx, cacheKey).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get set cardinality: %w", err)
	}
	return result, nil
}

// SPop 随机弹出集合成员
func (r *RedisCache) SPop(ctx context.Context, key string) (string, error) {
	cacheKey := r.buildKey(key)
	result, err := r.client.SPop(ctx, cacheKey).Result()
	if err != nil {
		if err == redis.Nil {
			return "", nil
		}
		return "", fmt.Errorf("failed to pop set member: %w", err)
	}
	return result, nil
}

// SRandMember 随机获取集合成员
func (r *RedisCache) SRandMember(ctx context.Context, key string) (string, error) {
	cacheKey := r.buildKey(key)
	result, err := r.client.SRandMember(ctx, cacheKey).Result()
	if err != nil {
		if err == redis.Nil {
			return "", nil
		}
		return "", fmt.Errorf("failed to get random set member: %w", err)
	}
	return result, nil
}

// ZSet操作实现

// ZAdd 添加有序集合成员
func (r *RedisCache) ZAdd(ctx context.Context, key string, members ...interface{}) error {
	cacheKey := r.buildKey(key)
	
	// 转换为Redis Z结构
	var zMembers []*redis.Z
	for i := 0; i < len(members); i += 2 {
		if i+1 >= len(members) {
			return fmt.Errorf("invalid members format")
		}
		
		score, err := strconv.ParseFloat(fmt.Sprintf("%v", members[i]), 64)
		if err != nil {
			return fmt.Errorf("invalid score: %w", err)
		}
		
		zMembers = append(zMembers, &redis.Z{
			Score:  score,
			Member: members[i+1],
		})
	}
	
	err := r.client.ZAdd(ctx, cacheKey, zMembers...).Err()
	if err != nil {
		return fmt.Errorf("failed to add sorted set members: %w", err)
	}
	return nil
}

// ZRem 删除有序集合成员
func (r *RedisCache) ZRem(ctx context.Context, key string, members ...interface{}) error {
	cacheKey := r.buildKey(key)
	err := r.client.ZRem(ctx, cacheKey, members...).Err()
	if err != nil {
		return fmt.Errorf("failed to remove sorted set members: %w", err)
	}
	return nil
}

// ZRange 获取有序集合范围
func (r *RedisCache) ZRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	cacheKey := r.buildKey(key)
	result, err := r.client.ZRange(ctx, cacheKey, start, stop).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get sorted set range: %w", err)
	}
	return result, nil
}

// ZRangeWithScores 获取有序集合范围（带分数）
func (r *RedisCache) ZRangeWithScores(ctx context.Context, key string, start, stop int64) ([]interface{}, error) {
	cacheKey := r.buildKey(key)
	result, err := r.client.ZRangeWithScores(ctx, cacheKey, start, stop).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get sorted set range with scores: %w", err)
	}
	
	// 转换为interface{}切片
	var output []interface{}
	for _, z := range result {
		output = append(output, z.Member, z.Score)
	}
	
	return output, nil
}

// ZRangeByScore 按分数范围获取有序集合
func (r *RedisCache) ZRangeByScore(ctx context.Context, key string, min, max string) ([]string, error) {
	cacheKey := r.buildKey(key)
	result, err := r.client.ZRangeByScore(ctx, cacheKey, &redis.ZRangeBy{
		Min: min,
		Max: max,
	}).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get sorted set range by score: %w", err)
	}
	return result, nil
}

// ZCard 获取有序集合成员数量
func (r *RedisCache) ZCard(ctx context.Context, key string) (int64, error) {
	cacheKey := r.buildKey(key)
	result, err := r.client.ZCard(ctx, cacheKey).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get sorted set cardinality: %w", err)
	}
	return result, nil
}

// ZScore 获取成员分数
func (r *RedisCache) ZScore(ctx context.Context, key, member string) (float64, error) {
	cacheKey := r.buildKey(key)
	result, err := r.client.ZScore(ctx, cacheKey, member).Result()
	if err != nil {
		if err == redis.Nil {
			return 0, nil
		}
		return 0, fmt.Errorf("failed to get sorted set member score: %w", err)
	}
	return result, nil
}

// ZRank 获取成员排名
func (r *RedisCache) ZRank(ctx context.Context, key, member string) (int64, error) {
	cacheKey := r.buildKey(key)
	result, err := r.client.ZRank(ctx, cacheKey, member).Result()
	if err != nil {
		if err == redis.Nil {
			return -1, nil
		}
		return 0, fmt.Errorf("failed to get sorted set member rank: %w", err)
	}
	return result, nil
}