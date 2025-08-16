package gateway

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"

	"pulse/internal/gateway/middleware"
)

// Gateway API网关
type Gateway struct {
	logger      *logrus.Logger
	router      *gin.Engine
	redisClient *redis.Client
	authService middleware.AuthService
	rbacService middleware.RBACService
}

// GatewayConfig 网关配置
type GatewayConfig struct {
	JWTSecret   string
	RedisClient *redis.Client
	APIKeys     map[string]string
}

// NewGateway 创建新的API网关
func NewGateway(logger *logrus.Logger, redisClient *redis.Client) *Gateway {
	// 设置Gin模式
	gin.SetMode(gin.ReleaseMode)

	// 创建路由器
	router := gin.New()

	// 创建认证服务
	authService := middleware.NewJWTAuthService("your-secret-key", 24*time.Hour)

	// 创建RBAC服务
	rbacService := middleware.NewDefaultRBACService()

	return &Gateway{
		logger:      logger,
		router:      router,
		redisClient: redisClient,
		authService: authService,
		rbacService: rbacService,
	}
}

// SetupRoutes 设置路由
func (g *Gateway) SetupRoutes() http.Handler {
	// 注册默认中间件
	g.registerDefaultMiddleware()

	// 注册路由
	g.registerRoutes()

	return g.router
}

// RegisterMiddleware 注册中间件
func (g *Gateway) RegisterMiddleware(middleware gin.HandlerFunc) {
	// 直接使用router的Use方法
	g.router.Use(middleware)
}

// registerDefaultMiddleware 注册默认中间件
func (g *Gateway) registerDefaultMiddleware() {
	// 请求ID中间件
	g.router.Use(middleware.RequestIDMiddleware())

	// 健康检查中间件（设置跳过标记）
	g.router.Use(middleware.HealthCheckMiddleware())

	// 安全头中间件
	g.router.Use(middleware.SecurityMiddleware(middleware.DefaultSecurityConfig()))

	// CORS中间件
	g.router.Use(middleware.CORSMiddleware(middleware.DefaultCORSConfig()))

	// 日志中间件
	loggerConfig := middleware.LoggerConfig{
		Logger:        g.logger,
		SkipPaths:     []string{"/health", "/status"},
		EnableDetails: false,
	}
	g.router.Use(middleware.LoggerMiddleware(loggerConfig))

	// 恢复中间件
	recoveryConfig := middleware.RecoveryConfig{
		Logger:        g.logger,
		EnableDetails: true,
	}
	g.router.Use(middleware.RecoveryMiddleware(recoveryConfig))

	// 限流中间件
	if g.redisClient != nil {
		rateLimitConfig := middleware.DefaultRateLimitConfig(g.redisClient)
		rateLimitConfig.Logger = g.logger
		g.router.Use(middleware.RateLimitMiddleware(rateLimitConfig))
	}

	// 指标收集中间件
	g.router.Use(middleware.MetricsMiddleware())

	// 超时中间件
	timeoutConfig := middleware.TimeoutConfig{
		Timeout: 30 * time.Second,
		Message: "Request timeout",
	}
	g.router.Use(middleware.TimeoutMiddleware(timeoutConfig))
}

















// registerRoutes 注册路由
func (g *Gateway) registerRoutes() {
	// 健康检查端点
	g.router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "healthy",
			"timestamp": time.Now().Unix(),
		})
	})

	// 状态检查端点
	g.router.GET("/status", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "running",
			"version": "1.0.0",
			"timestamp": time.Now().Unix(),
		})
	})

	// API路由组
	api := g.router.Group("/api/v1")
	{
		// 需要认证的路由
		api.Use(middleware.RequireAuthMiddleware(g.authService))
		
		// 告警相关路由
		alerts := api.Group("/alerts")
		{
			alerts.GET("", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "alerts endpoint"})
			})
			alerts.POST("", func(c *gin.Context) {
				c.JSON(http.StatusCreated, gin.H{"message": "alert created"})
			})
		}
	}
}

// GetRouter 获取Gin路由器
func (g *Gateway) GetRouter() *gin.Engine {
	return g.router
}