package monitor

import (
	"encoding/json"
	"net/http"
	"time"
)

// HealthHandler 健康检查HTTP处理器
type HealthHandler struct {
	monitor *HealthMonitor
}

// NewHealthHandler 创建健康检查HTTP处理器
func NewHealthHandler(monitor *HealthMonitor) *HealthHandler {
	return &HealthHandler{
		monitor: monitor,
	}
}

// HealthResponse 健康检查响应
type HealthResponse struct {
	Status    HealthStatus               `json:"status"`
	Timestamp time.Time                 `json:"timestamp"`
	Checks    map[string]HealthResult    `json:"checks"`
	Summary   map[string]interface{}     `json:"summary"`
}

// HandleHealth 处理健康检查请求
func (h *HealthHandler) HandleHealth(w http.ResponseWriter, r *http.Request) {
	
	// 设置响应头
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")
	
	// 获取健康检查结果
	results := h.monitor.GetResults()
	overallStatus := h.monitor.GetOverallStatus()
	summary := h.monitor.GetSummary()
	
	// 构建响应
	response := HealthResponse{
		Status:    overallStatus,
		Timestamp: time.Now(),
		Checks:    results,
		Summary:   summary,
	}
	
	// 根据整体状态设置HTTP状态码
	var statusCode int
	switch overallStatus {
	case HealthStatusHealthy:
		statusCode = http.StatusOK
	case HealthStatusDegraded:
		statusCode = http.StatusOK // 降级状态仍返回200，但在响应体中标明
	case HealthStatusUnhealthy:
		statusCode = http.StatusServiceUnavailable
	default:
		statusCode = http.StatusServiceUnavailable
	}
	
	w.WriteHeader(statusCode)
	
	// 编码并发送响应
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(response); err != nil {
		http.Error(w, "Failed to encode health response", http.StatusInternalServerError)
		return
	}
}

// HandleHealthCheck 处理单个健康检查请求
func (h *HealthHandler) HandleHealthCheck(w http.ResponseWriter, r *http.Request) {
	
	// 获取检查名称
	checkName := r.URL.Query().Get("name")
	if checkName == "" {
		http.Error(w, "Missing 'name' parameter", http.StatusBadRequest)
		return
	}
	
	// 检查是否强制执行
	forceCheck := r.URL.Query().Get("force") == "true"
	
	w.Header().Set("Content-Type", "application/json")
	
	var result HealthResult
	var err error
	
	if forceCheck {
		// 强制执行检查
		result, err = h.monitor.CheckOnce(r.Context(), checkName)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
	} else {
		// 获取缓存的结果
		var exists bool
		result, exists = h.monitor.GetResult(checkName)
		if !exists {
			http.Error(w, "Health check not found", http.StatusNotFound)
			return
		}
	}
	
	// 根据检查状态设置HTTP状态码
	var statusCode int
	switch result.Status {
	case HealthStatusHealthy:
		statusCode = http.StatusOK
	case HealthStatusDegraded:
		statusCode = http.StatusOK
	case HealthStatusUnhealthy:
		statusCode = http.StatusServiceUnavailable
	default:
		statusCode = http.StatusServiceUnavailable
	}
	
	w.WriteHeader(statusCode)
	
	// 编码并发送响应
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(result); err != nil {
		http.Error(w, "Failed to encode health check result", http.StatusInternalServerError)
		return
	}
}

// HandleLiveness 处理存活性检查请求（简单的ping检查）
func (h *HealthHandler) HandleLiveness(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	
	response := map[string]interface{}{
		"status":    "alive",
		"timestamp": time.Now(),
		"service":   "alert-management-platform",
	}
	
	json.NewEncoder(w).Encode(response)
}

// HandleReadiness 处理就绪性检查请求
func (h *HealthHandler) HandleReadiness(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	// 检查监控器是否正在运行
	if !h.monitor.IsRunning() {
		w.WriteHeader(http.StatusServiceUnavailable)
		response := map[string]interface{}{
			"status":    "not_ready",
			"message":   "Health monitor is not running",
			"timestamp": time.Now(),
		}
		json.NewEncoder(w).Encode(response)
		return
	}
	
	// 获取整体状态
	overallStatus := h.monitor.GetOverallStatus()
	
	var statusCode int
	var message string
	
	switch overallStatus {
	case HealthStatusHealthy:
		statusCode = http.StatusOK
		message = "Service is ready"
	case HealthStatusDegraded:
		statusCode = http.StatusOK
		message = "Service is ready but degraded"
	case HealthStatusUnhealthy:
		statusCode = http.StatusServiceUnavailable
		message = "Service is not ready"
	default:
		statusCode = http.StatusServiceUnavailable
		message = "Service readiness unknown"
	}
	
	w.WriteHeader(statusCode)
	
	response := map[string]interface{}{
		"status":         overallStatus,
		"message":        message,
		"timestamp":      time.Now(),
		"monitor_running": h.monitor.IsRunning(),
	}
	
	json.NewEncoder(w).Encode(response)
}

// HandleMetrics 处理指标请求
func (h *HealthHandler) HandleMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	results := h.monitor.GetResults()
	summary := h.monitor.GetSummary()
	
	// 构建指标数据
	metrics := map[string]interface{}{
		"summary": summary,
		"checks":  make(map[string]interface{}),
	}
	
	// 为每个检查添加指标
	for name, result := range results {
		checkMetrics := map[string]interface{}{
			"status":           result.Status,
			"duration_seconds": result.Duration.Seconds(),
			"timestamp":        result.Timestamp.Unix(),
			"healthy":          result.Status == HealthStatusHealthy,
		}
		
		if result.Error != nil {
			checkMetrics["error"] = result.Error.Error()
		}
		
		metrics["checks"].(map[string]interface{})[name] = checkMetrics
	}
	
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(metrics)
}

// RegisterRoutes 注册健康检查路由
func (h *HealthHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/health", h.HandleHealth)
	mux.HandleFunc("/health/check", h.HandleHealthCheck)
	mux.HandleFunc("/health/live", h.HandleLiveness)
	mux.HandleFunc("/health/ready", h.HandleReadiness)
	mux.HandleFunc("/health/metrics", h.HandleMetrics)
}

// Middleware 健康检查中间件
func (h *HealthHandler) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 在请求处理前检查服务健康状态
		if h.monitor.GetOverallStatus() == HealthStatusUnhealthy {
			// 如果服务不健康，可以选择拒绝请求或记录警告
			w.Header().Set("X-Service-Health", "unhealthy")
		} else {
			w.Header().Set("X-Service-Health", "healthy")
		}
		
		next.ServeHTTP(w, r)
	})
}