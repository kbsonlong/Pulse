package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// RequestIDMiddleware 请求ID中间件
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查是否已有请求ID
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			// 生成新的请求ID
			requestID = uuid.New().String()
		}

		// 设置请求ID到上下文和响应头
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)

		// 将请求ID添加到context中
		ctx := context.WithValue(c.Request.Context(), "request_id", requestID)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}

// LoggingMiddleware 日志中间件
func LoggingMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		// 获取请求ID
		requestID := param.Request.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = "unknown"
		}

		// 记录请求日志
		logger.Info("HTTP Request",
			zap.String("request_id", requestID),
			zap.String("method", param.Method),
			zap.String("path", param.Path),
			zap.String("query", param.Request.URL.RawQuery),
			zap.Int("status", param.StatusCode),
			zap.Duration("latency", param.Latency),
			zap.String("ip", param.ClientIP),
			zap.String("user_agent", param.Request.UserAgent()),
			zap.Int("body_size", param.BodySize),
		)

		return ""
	})
}

// CORSMiddleware CORS中间件
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		
		// 允许的源列表（生产环境应该配置具体的域名）
		allowedOrigins := []string{
			"http://localhost:3000",
			"http://localhost:8080",
			"http://127.0.0.1:3000",
			"http://127.0.0.1:8080",
		}

		// 检查是否为允许的源
		allowed := false
		for _, allowedOrigin := range allowedOrigins {
			if origin == allowedOrigin {
				allowed = true
				break
			}
		}

		if allowed {
			c.Header("Access-Control-Allow-Origin", origin)
		} else {
			c.Header("Access-Control-Allow-Origin", "*")
		}

		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Request-ID")
		c.Header("Access-Control-Expose-Headers", "X-Request-ID")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Max-Age", "86400")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// RecoveryMiddleware 恢复中间件
func RecoveryMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		requestID := c.GetString("request_id")
		if requestID == "" {
			requestID = "unknown"
		}

		logger.Error("Panic recovered",
			zap.String("request_id", requestID),
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.String("ip", c.ClientIP()),
			zap.Any("error", recovered),
		)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error":      "Internal server error",
			"request_id": requestID,
			"timestamp":  time.Now().Unix(),
		})
	})
}

// JWTClaims JWT声明结构
type JWTClaims struct {
	UserID   string   `json:"user_id"`
	Username string   `json:"username"`
	Roles    []string `json:"roles"`
	jwt.RegisteredClaims
}

// AuthMiddleware JWT认证中间件
func AuthMiddleware(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取Authorization头
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":      "Missing authorization header",
				"request_id": c.GetString("request_id"),
			})
			c.Abort()
			return
		}

		// 检查Bearer前缀
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":      "Invalid authorization header format",
				"request_id": c.GetString("request_id"),
			})
			c.Abort()
			return
		}

		tokenString := parts[1]

		// 解析JWT token
		token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
			// 验证签名方法
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(jwtSecret), nil
		})

		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":      "Invalid token: " + err.Error(),
				"request_id": c.GetString("request_id"),
			})
			c.Abort()
			return
		}

		// 验证token有效性
		if !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":      "Invalid token",
				"request_id": c.GetString("request_id"),
			})
			c.Abort()
			return
		}

		// 获取claims
		claims, ok := token.Claims.(*JWTClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":      "Invalid token claims",
				"request_id": c.GetString("request_id"),
			})
			c.Abort()
			return
		}

		// 将用户信息设置到上下文
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("roles", claims.Roles)

		c.Next()
	}
}

// APIKeyMiddleware API Key认证中间件
func APIKeyMiddleware(validAPIKeys map[string]string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从头部或查询参数获取API Key
		apiKey := c.GetHeader("X-API-Key")
		if apiKey == "" {
			apiKey = c.Query("api_key")
		}

		if apiKey == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":      "Missing API key",
				"request_id": c.GetString("request_id"),
			})
			c.Abort()
			return
		}

		// 验证API Key
		userID, exists := validAPIKeys[apiKey]
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":      "Invalid API key",
				"request_id": c.GetString("request_id"),
			})
			c.Abort()
			return
		}

		// 设置用户信息到上下文
		c.Set("user_id", userID)
		c.Set("auth_method", "api_key")

		c.Next()
	}
}

// RoleBasedAccessControl RBAC权限控制中间件
func RoleBasedAccessControl(requiredRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取用户角色
		userRoles, exists := c.Get("roles")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{
				"error":      "No roles found in token",
				"request_id": c.GetString("request_id"),
			})
			c.Abort()
			return
		}

		roles, ok := userRoles.([]string)
		if !ok {
			c.JSON(http.StatusForbidden, gin.H{
				"error":      "Invalid roles format",
				"request_id": c.GetString("request_id"),
			})
			c.Abort()
			return
		}

		// 检查是否有所需角色
		hasRequiredRole := false
		for _, userRole := range roles {
			for _, requiredRole := range requiredRoles {
				if userRole == requiredRole || userRole == "admin" {
					hasRequiredRole = true
					break
				}
			}
			if hasRequiredRole {
				break
			}
		}

		if !hasRequiredRole {
			c.JSON(http.StatusForbidden, gin.H{
				"error":      "Insufficient permissions",
				"required":   requiredRoles,
				"user_roles": roles,
				"request_id": c.GetString("request_id"),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RateLimitConfig 限流配置
type RateLimitConfig struct {
	RequestsPerMinute int
	BurstSize         int
}

// 简单的内存限流器（生产环境应使用Redis）
type rateLimiter struct {
	requests map[string][]time.Time
	config   RateLimitConfig
}

var globalRateLimiter = &rateLimiter{
	requests: make(map[string][]time.Time),
	config: RateLimitConfig{
		RequestsPerMinute: 60,
		BurstSize:         10,
	},
}

// RateLimitMiddleware 限流中间件
func RateLimitMiddleware(config RateLimitConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		now := time.Now()

		// 清理过期的请求记录
		if requests, exists := globalRateLimiter.requests[clientIP]; exists {
			var validRequests []time.Time
			for _, reqTime := range requests {
				if now.Sub(reqTime) < time.Minute {
					validRequests = append(validRequests, reqTime)
				}
			}
			globalRateLimiter.requests[clientIP] = validRequests
		}

		// 检查请求频率
		currentRequests := len(globalRateLimiter.requests[clientIP])
		if currentRequests >= config.RequestsPerMinute {
			c.Header("X-RateLimit-Limit", strconv.Itoa(config.RequestsPerMinute))
			c.Header("X-RateLimit-Remaining", "0")
			c.Header("X-RateLimit-Reset", strconv.FormatInt(now.Add(time.Minute).Unix(), 10))

			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":      "Rate limit exceeded",
				"limit":      config.RequestsPerMinute,
				"window":     "1 minute",
				"request_id": c.GetString("request_id"),
			})
			c.Abort()
			return
		}

		// 记录当前请求
		globalRateLimiter.requests[clientIP] = append(globalRateLimiter.requests[clientIP], now)

		// 设置限流头部
		remaining := config.RequestsPerMinute - currentRequests - 1
		c.Header("X-RateLimit-Limit", strconv.Itoa(config.RequestsPerMinute))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))
		c.Header("X-RateLimit-Reset", strconv.FormatInt(now.Add(time.Minute).Unix(), 10))

		c.Next()
	}
}

// MetricsMiddleware 指标收集中间件
func MetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		// 计算请求处理时间
		duration := time.Since(start)

		// 这里可以集成Prometheus或其他指标系统
		// 暂时只设置响应头
		c.Header("X-Response-Time", duration.String())
	}
}