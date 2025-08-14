package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	"Pulse/internal/cache"
	"Pulse/internal/config"
	"Pulse/internal/lock"
	"Pulse/internal/monitor"
	"Pulse/internal/queue"
	"Pulse/internal/redis"
)

// Message 消息类型别名
type Message = queue.Message

func main() {
	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 创建logger
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	// 创建Redis客户端
	redisClient, err := redis.New(&cfg.Redis, logger)
	if err != nil {
		log.Fatalf("Failed to create Redis client: %v", err)
	}
	defer redisClient.Close()

	log.Println("Redis client created successfully")

	// 测试Redis连接
	ctx := context.Background()
	if err := redisClient.Ping(ctx); err != nil {
		log.Fatalf("Failed to ping Redis: %v", err)
	}
	log.Println("Redis ping successful")

	// 测试基本Redis操作
	testRedisOperations(ctx, redisClient)

	// 创建缓存客户端
	cacheClient := cache.NewRedisCache(redisClient.GetClient(), cache.WithPrefix("test:"))
	testCacheOperations(ctx, cacheClient)

	// 创建消息队列
	queueClient := queue.NewRedisQueue(redisClient, cfg, logger)
	testQueueOperations(ctx, queueClient)

	// 创建分布式锁
	lockClient := lock.NewRedisLock(redisClient.GetClient(), "test_lock")
	mutexClient := lock.NewRedisMutex(redisClient.GetClient(), "test_mutex")
	testLockOperations(ctx, lockClient, mutexClient)

	// 创建健康监控
	healthMonitor := monitor.NewHealthMonitor(30*time.Second, 5*time.Second)

	// 添加健康检查
	healthMonitor.AddCheck(monitor.NewRedisHealthCheck("redis", redisClient.GetClient()))
	healthMonitor.AddCheck(monitor.NewQueueHealthCheck("queue", redisClient.GetClient(), "test_queue"))
	healthMonitor.AddCheck(monitor.NewCacheHealthCheck("cache", redisClient.GetClient()))

	// 添加健康状态变化回调
	healthMonitor.AddCallback(func(name string, oldResult, newResult monitor.HealthResult) {
		log.Printf("Health status changed for %s: %s -> %s", name, oldResult.Status, newResult.Status)
	})

	// 启动健康监控
	monitorCtx, cancelMonitor := context.WithCancel(ctx)
	go healthMonitor.Start(monitorCtx)

	// 等待一段时间让健康检查运行
	time.Sleep(2 * time.Second)

	// 创建HTTP服务器用于健康检查API
	healthHandler := monitor.NewHealthHandler(healthMonitor)
	mux := http.NewServeMux()
	healthHandler.RegisterRoutes(mux)

	server := &http.Server{
		Addr:    ":8081",
		Handler: mux,
	}

	// 启动HTTP服务器
	go func() {
		log.Println("Starting health check server on :8081")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Health check server error: %v", err)
		}
	}()

	// 显示健康检查结果
	time.Sleep(2 * time.Second)
	showHealthResults(healthMonitor)

	log.Println("\nTest completed successfully!")
	log.Println("Health check endpoints available at:")
	log.Println("  - http://localhost:8081/health")
	log.Println("  - http://localhost:8081/health/live")
	log.Println("  - http://localhost:8081/health/ready")
	log.Println("  - http://localhost:8081/health/metrics")
	log.Println("\nPress Ctrl+C to exit...")

	// 等待中断信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("\nShutting down...")

	// 停止健康监控
	cancelMonitor()
	healthMonitor.Stop()

	// 停止HTTP服务器
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}

	log.Println("Shutdown complete")
}

func testRedisOperations(ctx context.Context, client *redis.Client) {
	log.Println("\n=== Testing Redis Operations ===")

	// 测试SET/GET
	key := "test:key"
	value := "test_value"

	if err := client.Set(ctx, key, value, time.Minute); err != nil {
		log.Printf("SET operation failed: %v", err)
		return
	}
	log.Printf("SET %s = %s", key, value)

	result, err := client.Get(ctx, key)
	if err != nil {
		log.Printf("GET operation failed: %v", err)
		return
	}
	log.Printf("GET %s = %s", key, result)

	// 清理
	client.Del(ctx, key)
	log.Println("Redis operations test completed")
}

func testCacheOperations(ctx context.Context, cache cache.Cache) {
	log.Println("\n=== Testing Cache Operations ===")

	// 测试基本缓存操作
	key := "cache_test"
	value := map[string]interface{}{
		"name": "test",
		"age":  25,
	}

	if err := cache.Set(ctx, key, value, time.Minute); err != nil {
		log.Printf("Cache SET failed: %v", err)
		return
	}
	log.Printf("Cache SET %s", key)

	result, err := cache.Get(ctx, key)
	if err != nil {
		log.Printf("Cache GET failed: %v", err)
		return
	}
	log.Printf("Cache GET %s = %+v", key, result)

	// 测试TTL
	ttl, err := cache.TTL(ctx, key)
	if err != nil {
		log.Printf("Cache TTL failed: %v", err)
	} else {
		log.Printf("Cache TTL %s = %v", key, ttl)
	}

	// 清理
	cache.Del(ctx, key)
	log.Println("Cache operations test completed")
}

func testQueueOperations(ctx context.Context, queue queue.Queue) {
	log.Println("\n=== Testing Queue Operations ===")

	// 测试发布消息
	topic := "test_topic"
	payload := []byte("test message")

	if err := queue.Publish(ctx, topic, payload); err != nil {
		log.Printf("Publish failed: %v", err)
		return
	}
	log.Printf("Published message to topic: %s", topic)

	// 测试订阅消息
	messageReceived := make(chan bool, 1)

	handler := func(ctx context.Context, msg *Message) error {
		log.Printf("Received message: %s", string(msg.Payload))
		messageReceived <- true
		return nil
	}

	if err := queue.Subscribe(ctx, topic, handler); err != nil {
		log.Printf("Subscribe failed: %v", err)
		return
	}

	// 等待消息接收
	select {
	case <-messageReceived:
		log.Println("Message received successfully")
	case <-time.After(5 * time.Second):
		log.Println("Timeout waiting for message")
	}

	log.Println("Queue operations test completed")
}

func testLockOperations(ctx context.Context, lockClient *lock.RedisLock, mutexClient *lock.RedisMutex) {
	log.Println("\n=== Testing Lock Operations ===")

	// 测试RedisLock
	lockKey := "test_lock"
	log.Printf("Testing RedisLock with key: %s", lockKey)

	// 获取锁
	acquired, err := lockClient.Acquire(ctx, lockKey, 10*time.Second)
	if err != nil {
		log.Printf("Failed to acquire lock: %v", err)
		return
	}
	if !acquired {
		log.Printf("Lock not acquired: %s", lockKey)
		return
	}
	log.Printf("Lock acquired: %s", lockKey)

	// 检查锁状态
	isLocked, err := lockClient.IsLocked(ctx, lockKey)
	if err != nil {
		log.Printf("Failed to check lock status: %v", err)
		return
	}
	log.Printf("Lock status: %v", isLocked)

	// 释放锁
	if err := lockClient.Release(ctx, lockKey); err != nil {
		log.Printf("Failed to release lock: %v", err)
		return
	}
	log.Printf("Lock released: %s", lockKey)

	// 测试RedisMutex
	log.Printf("Testing RedisMutex")

	// 加锁
	if err := mutexClient.Lock(ctx); err != nil {
		log.Printf("Failed to lock mutex: %v", err)
		return
	}
	log.Printf("Mutex locked")

	// 检查锁状态
	isLocked, err = mutexClient.IsLocked(ctx)
	if err != nil {
		log.Printf("Failed to check mutex status: %v", err)
		return
	}
	log.Printf("Mutex status: %v", isLocked)

	// 解锁
	if err := mutexClient.Unlock(ctx); err != nil {
		log.Printf("Failed to unlock mutex: %v", err)
		return
	}
	log.Printf("Mutex unlocked")

	log.Println("Lock operations test completed")
}

func showHealthResults(healthMonitor *monitor.HealthMonitor) {
	log.Println("\n=== Health Check Results ===")

	results := healthMonitor.GetResults()
	for name, result := range results {
		log.Printf("Check: %s, Status: %s, Duration: %v, Message: %s",
			name, result.Status, result.Duration, result.Message)
	}

	overallStatus := healthMonitor.GetOverallStatus()
	log.Printf("Overall Status: %s", overallStatus)

	summary := healthMonitor.GetSummary()
	log.Printf("Summary: %+v", summary)
}
