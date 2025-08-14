package lock

import (
	"context"
	"time"
)

// Lock 分布式锁接口
type Lock interface {
	// Acquire 获取锁
	Acquire(ctx context.Context, key string, ttl time.Duration) (bool, error)
	
	// Release 释放锁
	Release(ctx context.Context, key string) error
	
	// Extend 延长锁的过期时间
	Extend(ctx context.Context, key string, ttl time.Duration) error
	
	// IsLocked 检查锁是否存在
	IsLocked(ctx context.Context, key string) (bool, error)
	
	// GetTTL 获取锁的剩余时间
	GetTTL(ctx context.Context, key string) (time.Duration, error)
}

// Mutex 分布式互斥锁接口
type Mutex interface {
	// Lock 加锁（阻塞直到获取锁）
	Lock(ctx context.Context) error
	
	// TryLock 尝试加锁（非阻塞）
	TryLock(ctx context.Context) (bool, error)
	
	// Unlock 解锁
	Unlock(ctx context.Context) error
	
	// IsLocked 检查是否已锁定
	IsLocked(ctx context.Context) (bool, error)
}

// LockOptions 锁选项
type LockOptions struct {
	// TTL 锁的生存时间
	TTL time.Duration
	
	// RetryInterval 重试间隔
	RetryInterval time.Duration
	
	// MaxRetries 最大重试次数
	MaxRetries int
	
	// AutoRenew 是否自动续期
	AutoRenew bool
	
	// RenewInterval 续期间隔
	RenewInterval time.Duration
	
	// Metadata 元数据
	Metadata map[string]string
}

// LockOption 锁选项函数
type LockOption func(*LockOptions)

// WithTTL 设置锁的生存时间
func WithTTL(ttl time.Duration) LockOption {
	return func(opts *LockOptions) {
		opts.TTL = ttl
	}
}

// WithRetryInterval 设置重试间隔
func WithRetryInterval(interval time.Duration) LockOption {
	return func(opts *LockOptions) {
		opts.RetryInterval = interval
	}
}

// WithMaxRetries 设置最大重试次数
func WithMaxRetries(maxRetries int) LockOption {
	return func(opts *LockOptions) {
		opts.MaxRetries = maxRetries
	}
}

// WithAutoRenew 设置自动续期
func WithAutoRenew(autoRenew bool) LockOption {
	return func(opts *LockOptions) {
		opts.AutoRenew = autoRenew
	}
}

// WithRenewInterval 设置续期间隔
func WithRenewInterval(interval time.Duration) LockOption {
	return func(opts *LockOptions) {
		opts.RenewInterval = interval
	}
}

// WithMetadata 设置元数据
func WithMetadata(metadata map[string]string) LockOption {
	return func(opts *LockOptions) {
		opts.Metadata = metadata
	}
}

// applyLockOptions 应用锁选项
func applyLockOptions(opts ...LockOption) *LockOptions {
	options := &LockOptions{
		TTL:           30 * time.Second,
		RetryInterval: 100 * time.Millisecond,
		MaxRetries:    10,
		AutoRenew:     false,
		RenewInterval: 10 * time.Second,
		Metadata:      make(map[string]string),
	}
	
	for _, opt := range opts {
		opt(options)
	}
	
	return options
}