package monitor

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// HealthStatus 健康状态
type HealthStatus string

const (
	HealthStatusHealthy   HealthStatus = "healthy"
	HealthStatusUnhealthy HealthStatus = "unhealthy"
	HealthStatusDegraded  HealthStatus = "degraded"
	HealthStatusUnknown   HealthStatus = "unknown"
)

// HealthCheck 健康检查接口
type HealthCheck interface {
	// Name 返回健康检查名称
	Name() string
	
	// Check 执行健康检查
	Check(ctx context.Context) HealthResult
}

// HealthResult 健康检查结果
type HealthResult struct {
	// Status 健康状态
	Status HealthStatus `json:"status"`
	
	// Message 状态消息
	Message string `json:"message"`
	
	// Details 详细信息
	Details map[string]interface{} `json:"details,omitempty"`
	
	// Timestamp 检查时间
	Timestamp time.Time `json:"timestamp"`
	
	// Duration 检查耗时
	Duration time.Duration `json:"duration"`
	
	// Error 错误信息
	Error error `json:"error,omitempty"`
}

// HealthMonitor 健康监控器
type HealthMonitor struct {
	checks   map[string]HealthCheck
	results  map[string]HealthResult
	mu       sync.RWMutex
	interval time.Duration
	timeout  time.Duration
	running  bool
	cancel   context.CancelFunc
	callbacks []HealthCallback
}

// HealthCallback 健康状态变化回调
type HealthCallback func(name string, oldResult, newResult HealthResult)

// NewHealthMonitor 创建健康监控器
func NewHealthMonitor(interval, timeout time.Duration) *HealthMonitor {
	return &HealthMonitor{
		checks:   make(map[string]HealthCheck),
		results:  make(map[string]HealthResult),
		interval: interval,
		timeout:  timeout,
	}
}

// AddCheck 添加健康检查
func (h *HealthMonitor) AddCheck(check HealthCheck) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.checks[check.Name()] = check
}

// RemoveCheck 移除健康检查
func (h *HealthMonitor) RemoveCheck(name string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.checks, name)
	delete(h.results, name)
}

// AddCallback 添加状态变化回调
func (h *HealthMonitor) AddCallback(callback HealthCallback) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.callbacks = append(h.callbacks, callback)
}

// Start 启动健康监控
func (h *HealthMonitor) Start(ctx context.Context) {
	h.mu.Lock()
	if h.running {
		h.mu.Unlock()
		return
	}
	h.running = true
	ctx, cancel := context.WithCancel(ctx)
	h.cancel = cancel
	h.mu.Unlock()
	
	// 立即执行一次检查
	h.runChecks(ctx)
	
	// 定期执行检查
	ticker := time.NewTicker(h.interval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			h.mu.Lock()
			h.running = false
			h.mu.Unlock()
			return
		case <-ticker.C:
			h.runChecks(ctx)
		}
	}
}

// Stop 停止健康监控
func (h *HealthMonitor) Stop() {
	h.mu.Lock()
	defer h.mu.Unlock()
	
	if h.cancel != nil {
		h.cancel()
		h.cancel = nil
	}
	h.running = false
}

// runChecks 执行所有健康检查
func (h *HealthMonitor) runChecks(ctx context.Context) {
	h.mu.RLock()
	checks := make(map[string]HealthCheck)
	for name, check := range h.checks {
		checks[name] = check
	}
	h.mu.RUnlock()
	
	// 并发执行所有检查
	var wg sync.WaitGroup
	resultsChan := make(chan struct {
		name   string
		result HealthResult
	}, len(checks))
	
	for name, check := range checks {
		wg.Add(1)
		go func(name string, check HealthCheck) {
			defer wg.Done()
			
			// 设置超时
			checkCtx, cancel := context.WithTimeout(ctx, h.timeout)
			defer cancel()
			
			start := time.Now()
			result := check.Check(checkCtx)
			result.Duration = time.Since(start)
			result.Timestamp = time.Now()
			
			resultsChan <- struct {
				name   string
				result HealthResult
			}{name: name, result: result}
		}(name, check)
	}
	
	go func() {
		wg.Wait()
		close(resultsChan)
	}()
	
	// 收集结果并触发回调
	for item := range resultsChan {
		h.mu.Lock()
		oldResult, exists := h.results[item.name]
		h.results[item.name] = item.result
		callbacks := h.callbacks
		h.mu.Unlock()
		
		// 触发状态变化回调
		if exists && oldResult.Status != item.result.Status {
			for _, callback := range callbacks {
				go callback(item.name, oldResult, item.result)
			}
		}
	}
}

// GetResults 获取所有健康检查结果
func (h *HealthMonitor) GetResults() map[string]HealthResult {
	h.mu.RLock()
	defer h.mu.RUnlock()
	
	results := make(map[string]HealthResult)
	for name, result := range h.results {
		results[name] = result
	}
	return results
}

// GetResult 获取指定健康检查结果
func (h *HealthMonitor) GetResult(name string) (HealthResult, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	
	result, exists := h.results[name]
	return result, exists
}

// GetOverallStatus 获取整体健康状态
func (h *HealthMonitor) GetOverallStatus() HealthStatus {
	h.mu.RLock()
	defer h.mu.RUnlock()
	
	if len(h.results) == 0 {
		return HealthStatusUnknown
	}
	
	healthyCount := 0
	unhealthyCount := 0
	degradedCount := 0
	
	for _, result := range h.results {
		switch result.Status {
		case HealthStatusHealthy:
			healthyCount++
		case HealthStatusUnhealthy:
			unhealthyCount++
		case HealthStatusDegraded:
			degradedCount++
		}
	}
	
	// 如果有任何不健康的服务，整体状态为不健康
	if unhealthyCount > 0 {
		return HealthStatusUnhealthy
	}
	
	// 如果有降级的服务，整体状态为降级
	if degradedCount > 0 {
		return HealthStatusDegraded
	}
	
	// 所有服务都健康
	return HealthStatusHealthy
}

// IsRunning 检查监控器是否正在运行
func (h *HealthMonitor) IsRunning() bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.running
}

// GetSummary 获取健康检查摘要
func (h *HealthMonitor) GetSummary() map[string]interface{} {
	h.mu.RLock()
	defer h.mu.RUnlock()
	
	summary := map[string]interface{}{
		"overall_status": h.GetOverallStatus(),
		"total_checks":   len(h.checks),
		"timestamp":      time.Now(),
		"running":        h.running,
	}
	
	statusCount := map[HealthStatus]int{
		HealthStatusHealthy:   0,
		HealthStatusUnhealthy: 0,
		HealthStatusDegraded:  0,
		HealthStatusUnknown:   0,
	}
	
	for _, result := range h.results {
		statusCount[result.Status]++
	}
	
	summary["status_count"] = statusCount
	return summary
}

// CheckOnce 执行一次性健康检查
func (h *HealthMonitor) CheckOnce(ctx context.Context, name string) (HealthResult, error) {
	h.mu.RLock()
	check, exists := h.checks[name]
	h.mu.RUnlock()
	
	if !exists {
		return HealthResult{}, fmt.Errorf("health check '%s' not found", name)
	}
	
	checkCtx, cancel := context.WithTimeout(ctx, h.timeout)
	defer cancel()
	
	start := time.Now()
	result := check.Check(checkCtx)
	result.Duration = time.Since(start)
	result.Timestamp = time.Now()
	
	return result, nil
}