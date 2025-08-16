package middleware

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// RequestIDMiddleware 请求ID中间件
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头获取请求ID，如果没有则生成新的
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}

		// 设置到上下文和响应头
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)

		c.Next()
	}
}

// LoggerConfig 日志中间件配置
type LoggerConfig struct {
	Logger        *logrus.Logger
	SkipPaths     []string
	EnableDetails bool
}

// LoggerMiddleware 日志中间件
func LoggerMiddleware(config LoggerConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查是否跳过日志记录
		path := c.Request.URL.Path
		for _, skipPath := range config.SkipPaths {
			if path == skipPath {
				c.Next()
				return
			}
		}

		// 记录开始时间
		start := time.Now()
		requestID := c.GetString("request_id")

		// 记录请求体（如果启用详细日志）
		var requestBody []byte
		if config.EnableDetails && c.Request.Body != nil {
			requestBody, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		}

		// 处理请求
		c.Next()

		// 计算处理时间
		latency := time.Since(start)
		status := c.Writer.Status()
		clientIP := c.ClientIP()
		method := c.Request.Method
		userAgent := c.Request.UserAgent()

		// 构建日志字段
		logFields := logrus.Fields{
			"request_id": requestID,
			"method":     method,
			"path":       path,
			"status":     status,
			"latency":    latency,
			"client_ip":  clientIP,
			"user_agent": userAgent,
		}

		// 添加详细信息（如果启用）
		if config.EnableDetails {
			if len(requestBody) > 0 {
				logFields["request_body"] = string(requestBody)
			}
			logFields["query_params"] = c.Request.URL.RawQuery
		}

		// 根据状态码选择日志级别
		logMessage := fmt.Sprintf("%s %s %d", method, path, status)
		switch {
		case status >= 500:
			config.Logger.WithFields(logFields).Error(logMessage)
		case status >= 400:
			config.Logger.WithFields(logFields).Warn(logMessage)
		default:
			config.Logger.WithFields(logFields).Info(logMessage)
		}
	}
}

// CORSConfig CORS中间件配置
type CORSConfig struct {
	AllowOrigins     []string
	AllowMethods     []string
	AllowHeaders     []string
	ExposeHeaders    []string
	AllowCredentials bool
	MaxAge           time.Duration
}

// DefaultCORSConfig 默认CORS配置
func DefaultCORSConfig() CORSConfig {
	return CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
		AllowHeaders: []string{
			"Origin",
			"Content-Length",
			"Content-Type",
			"Authorization",
			"X-Requested-With",
			"X-Request-ID",
			"X-API-Key",
		},
		ExposeHeaders: []string{
			"X-Request-ID",
			"X-Total-Count",
			"X-Page-Count",
		},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}
}

// CORSMiddleware CORS中间件
func CORSMiddleware(config CORSConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// 检查是否允许该来源
		allowedOrigin := ""
		for _, allowOrigin := range config.AllowOrigins {
			if allowOrigin == "*" || allowOrigin == origin {
				allowedOrigin = allowOrigin
				break
			}
		}

		if allowedOrigin != "" {
			if allowedOrigin == "*" {
				c.Header("Access-Control-Allow-Origin", "*")
			} else {
				c.Header("Access-Control-Allow-Origin", origin)
			}

			c.Header("Access-Control-Allow-Methods", strings.Join(config.AllowMethods, ", "))
			c.Header("Access-Control-Allow-Headers", strings.Join(config.AllowHeaders, ", "))

			if len(config.ExposeHeaders) > 0 {
				c.Header("Access-Control-Expose-Headers", strings.Join(config.ExposeHeaders, ", "))
			}

			if config.AllowCredentials {
				c.Header("Access-Control-Allow-Credentials", "true")
			}

			if config.MaxAge > 0 {
				c.Header("Access-Control-Max-Age", fmt.Sprintf("%.0f", config.MaxAge.Seconds()))
			}
		}

		// 处理预检请求
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// SecurityConfig 安全中间件配置
type SecurityConfig struct {
	EnableXSSProtection      bool
	EnableContentTypeNoSniff bool
	EnableFrameDeny          bool
	EnableHSTS               bool
	HSTSMaxAge               time.Duration
	ContentSecurityPolicy    string
	ReferrerPolicy           string
}

// DefaultSecurityConfig 默认安全配置
func DefaultSecurityConfig() SecurityConfig {
	return SecurityConfig{
		EnableXSSProtection:      true,
		EnableContentTypeNoSniff: true,
		EnableFrameDeny:          true,
		EnableHSTS:               true,
		HSTSMaxAge:               365 * 24 * time.Hour, // 1年
		ContentSecurityPolicy:    "default-src 'self'",
		ReferrerPolicy:           "strict-origin-when-cross-origin",
	}
}

// SecurityMiddleware 安全头中间件
func SecurityMiddleware(config SecurityConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// X-XSS-Protection
		if config.EnableXSSProtection {
			c.Header("X-XSS-Protection", "1; mode=block")
		}

		// X-Content-Type-Options
		if config.EnableContentTypeNoSniff {
			c.Header("X-Content-Type-Options", "nosniff")
		}

		// X-Frame-Options
		if config.EnableFrameDeny {
			c.Header("X-Frame-Options", "DENY")
		}

		// Strict-Transport-Security
		if config.EnableHSTS && c.Request.TLS != nil {
			c.Header("Strict-Transport-Security", fmt.Sprintf("max-age=%.0f; includeSubDomains", config.HSTSMaxAge.Seconds()))
		}

		// Content-Security-Policy
		if config.ContentSecurityPolicy != "" {
			c.Header("Content-Security-Policy", config.ContentSecurityPolicy)
		}

		// Referrer-Policy
		if config.ReferrerPolicy != "" {
			c.Header("Referrer-Policy", config.ReferrerPolicy)
		}

		c.Next()
	}
}

// RecoveryConfig 恢复中间件配置
type RecoveryConfig struct {
	Logger        *logrus.Logger
	EnableDetails bool
}

// RecoveryMiddleware 恢复中间件（处理panic）
func RecoveryMiddleware(config RecoveryConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				requestID := c.GetString("request_id")

				// 记录panic日志
				logFields := logrus.Fields{
					"request_id": requestID,
					"method":     c.Request.Method,
					"path":       c.Request.URL.Path,
					"client_ip":  c.ClientIP(),
					"panic":      err,
				}

				if config.EnableDetails {
					logFields["user_agent"] = c.Request.UserAgent()
					logFields["query_params"] = c.Request.URL.RawQuery
				}

				config.Logger.WithFields(logFields).Error("Panic recovered")

				// 返回错误响应
				c.JSON(http.StatusInternalServerError, gin.H{
					"error":      "internal_server_error",
					"message":    "An internal server error occurred",
					"request_id": requestID,
				})

				c.Abort()
			}
		}()

		c.Next()
	}
}

// TimeoutConfig 超时中间件配置
type TimeoutConfig struct {
	Timeout time.Duration
	Message string
}

// TimeoutMiddleware 超时中间件
func TimeoutMiddleware(config TimeoutConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 创建带超时的context
		ctx, cancel := context.WithTimeout(c.Request.Context(), config.Timeout)
		defer cancel()

		// 替换请求的context
		c.Request = c.Request.WithContext(ctx)

		// 创建一个channel来接收处理完成的信号
		done := make(chan struct{})
		go func() {
			c.Next()
			close(done)
		}()

		// 等待处理完成或超时
		select {
		case <-done:
			// 处理完成
			return
		case <-ctx.Done():
			// 超时
			c.JSON(http.StatusRequestTimeout, gin.H{
				"error":   "request_timeout",
				"message": config.Message,
			})
			c.Abort()
			return
		}
	}
}

// HealthCheckMiddleware 健康检查中间件（跳过某些中间件）
func HealthCheckMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 为健康检查端点设置特殊标记
		if c.Request.URL.Path == "/health" || c.Request.URL.Path == "/status" {
			c.Set("skip_auth", true)
			c.Set("skip_rate_limit", true)
		}
		c.Next()
	}
}

// MetricsMiddleware 指标收集中间件
func MetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		c.Next()

		status := c.Writer.Status()
		latency := time.Since(start)

		// 这里可以集成Prometheus或其他指标系统
		// 目前只是记录到上下文中
		c.Set("metrics", gin.H{
			"method":   method,
			"path":     path,
			"status":   status,
			"latency":  latency,
			"timestamp": start,
		})
	}
}