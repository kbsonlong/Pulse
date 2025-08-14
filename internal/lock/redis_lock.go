package lock

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
)

var (
	ErrLockNotAcquired = errors.New("lock not acquired")
	ErrLockNotHeld     = errors.New("lock not held")
	ErrLockExpired     = errors.New("lock expired")
)

// RedisLock Redis分布式锁实现
type RedisLock struct {
	client *redis.Client
	prefix string
}

// NewRedisLock 创建Redis分布式锁
func NewRedisLock(client *redis.Client, prefix string) *RedisLock {
	if prefix == "" {
		prefix = "lock:"
	}
	return &RedisLock{
		client: client,
		prefix: prefix,
	}
}

// Acquire 获取锁
func (r *RedisLock) Acquire(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	lockKey := r.prefix + key
	value := generateLockValue()
	
	result, err := r.client.SetNX(ctx, lockKey, value, ttl).Result()
	if err != nil {
		return false, fmt.Errorf("failed to acquire lock: %w", err)
	}
	
	return result, nil
}

// Release 释放锁
func (r *RedisLock) Release(ctx context.Context, key string) error {
	lockKey := r.prefix + key
	
	// 使用Lua脚本确保原子性删除
	luaScript := `
		if redis.call("exists", KEYS[1]) == 1 then
			return redis.call("del", KEYS[1])
		else
			return 0
		end
	`
	
	result, err := r.client.Eval(ctx, luaScript, []string{lockKey}).Result()
	if err != nil {
		return fmt.Errorf("failed to release lock: %w", err)
	}
	
	if result.(int64) == 0 {
		return ErrLockNotHeld
	}
	
	return nil
}

// Extend 延长锁的过期时间
func (r *RedisLock) Extend(ctx context.Context, key string, ttl time.Duration) error {
	lockKey := r.prefix + key
	
	// 使用Lua脚本确保原子性延期
	luaScript := `
		if redis.call("exists", KEYS[1]) == 1 then
			return redis.call("expire", KEYS[1], ARGV[1])
		else
			return 0
		end
	`
	
	result, err := r.client.Eval(ctx, luaScript, []string{lockKey}, int(ttl.Seconds())).Result()
	if err != nil {
		return fmt.Errorf("failed to extend lock: %w", err)
	}
	
	if result.(int64) == 0 {
		return ErrLockNotHeld
	}
	
	return nil
}

// IsLocked 检查锁是否存在
func (r *RedisLock) IsLocked(ctx context.Context, key string) (bool, error) {
	lockKey := r.prefix + key
	
	exists, err := r.client.Exists(ctx, lockKey).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check lock existence: %w", err)
	}
	
	return exists > 0, nil
}

// GetTTL 获取锁的剩余时间
func (r *RedisLock) GetTTL(ctx context.Context, key string) (time.Duration, error) {
	lockKey := r.prefix + key
	
	ttl, err := r.client.TTL(ctx, lockKey).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get lock TTL: %w", err)
	}
	
	if ttl == -2 {
		return 0, ErrLockNotHeld
	}
	
	return ttl, nil
}

// RedisMutex Redis分布式互斥锁
type RedisMutex struct {
	lock   *RedisLock
	key    string
	value  string
	opts   *LockOptions
	mu     sync.Mutex
	locked bool
	cancel context.CancelFunc
}

// NewRedisMutex 创建Redis分布式互斥锁
func NewRedisMutex(client *redis.Client, key string, opts ...LockOption) *RedisMutex {
	options := applyLockOptions(opts...)
	return &RedisMutex{
		lock:  NewRedisLock(client, "mutex:"),
		key:   key,
		value: generateLockValue(),
		opts:  options,
	}
}

// Lock 加锁（阻塞直到获取锁）
func (m *RedisMutex) Lock(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if m.locked {
		return errors.New("mutex already locked")
	}
	
	retries := 0
	for {
		acquired, err := m.lock.Acquire(ctx, m.key, m.opts.TTL)
		if err != nil {
			return err
		}
		
		if acquired {
			m.locked = true
			
			// 启动自动续期
			if m.opts.AutoRenew {
				m.startAutoRenew(ctx)
			}
			
			return nil
		}
		
		// 检查重试次数
		if retries >= m.opts.MaxRetries {
			return ErrLockNotAcquired
		}
		
		retries++
		
		// 等待重试间隔
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(m.opts.RetryInterval):
			continue
		}
	}
}

// TryLock 尝试加锁（非阻塞）
func (m *RedisMutex) TryLock(ctx context.Context) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if m.locked {
		return false, errors.New("mutex already locked")
	}
	
	acquired, err := m.lock.Acquire(ctx, m.key, m.opts.TTL)
	if err != nil {
		return false, err
	}
	
	if acquired {
		m.locked = true
		
		// 启动自动续期
		if m.opts.AutoRenew {
			m.startAutoRenew(ctx)
		}
	}
	
	return acquired, nil
}

// Unlock 解锁
func (m *RedisMutex) Unlock(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if !m.locked {
		return ErrLockNotHeld
	}
	
	// 停止自动续期
	if m.cancel != nil {
		m.cancel()
		m.cancel = nil
	}
	
	err := m.lock.Release(ctx, m.key)
	if err != nil {
		return err
	}
	
	m.locked = false
	return nil
}

// IsLocked 检查是否已锁定
func (m *RedisMutex) IsLocked(ctx context.Context) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if !m.locked {
		return false, nil
	}
	
	return m.lock.IsLocked(ctx, m.key)
}

// startAutoRenew 启动自动续期
func (m *RedisMutex) startAutoRenew(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	m.cancel = cancel
	
	go func() {
		ticker := time.NewTicker(m.opts.RenewInterval)
		defer ticker.Stop()
		
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := m.lock.Extend(ctx, m.key, m.opts.TTL); err != nil {
					// 续期失败，停止自动续期
					return
				}
			}
		}
	}()
}

// generateLockValue 生成锁值
func generateLockValue() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes) + ":" + strconv.FormatInt(time.Now().UnixNano(), 10)
}