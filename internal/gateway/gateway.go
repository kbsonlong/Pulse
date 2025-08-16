package gateway

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"Pulse/internal/middleware"
	"Pulse/internal/service"
)

// Gateway API网关接口
type Gateway interface {
	SetupRoutes() http.Handler
	RegisterMiddleware(middleware gin.HandlerFunc)
	GetRouter() *gin.Engine
}

// gateway API网关实现
type gateway struct {
	serviceManager service.ServiceManager
	logger         *zap.Logger
	router         *gin.Engine
	middlewares    []gin.HandlerFunc
	redisClient    *redis.Client
	rateLimiter    *middleware.RedisRateLimiter
	jwtSecret      string
	apiKeys        map[string]string
}

// GatewayConfig 网关配置
type GatewayConfig struct {
	JWTSecret   string
	RedisClient *redis.Client
	APIKeys     map[string]string
}

// NewGateway 创建新的API网关
func NewGateway(serviceManager service.ServiceManager, logger *zap.Logger, config GatewayConfig) Gateway {
	// 设置Gin模式
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()

	// 创建Redis限流器
	var rateLimiter *middleware.RedisRateLimiter
	if config.RedisClient != nil {
		rateLimiter = middleware.NewRedisRateLimiter(config.RedisClient, logger)
	}

	g := &gateway{
		serviceManager: serviceManager,
		logger:         logger,
		router:         router,
		middlewares:    make([]gin.HandlerFunc, 0),
		redisClient:    config.RedisClient,
		rateLimiter:    rateLimiter,
		jwtSecret:      config.JWTSecret,
		apiKeys:        config.APIKeys,
	}

	// 注册默认中间件
	g.registerDefaultMiddlewares()

	return g
}

// SetupRoutes 设置路由
func (g *gateway) SetupRoutes() http.Handler {
	// 应用中间件
	for _, middleware := range g.middlewares {
		g.router.Use(middleware)
	}

	// 健康检查路由
	g.router.GET("/health", g.healthCheck)
	g.router.GET("/status", g.statusCheck)

	// API版本路由组
	v1 := g.router.Group("/api/v1")
	{
		// 认证相关路由
		auth := v1.Group("/auth")
		{
			auth.POST("/login", g.login)
			auth.POST("/logout", g.logout)
			auth.POST("/refresh", g.refreshToken)
			auth.POST("/reset-password", g.resetPassword)
		}

		// 需要认证的路由
		protected := v1.Group("/")
		protected.Use(g.authMiddleware())
		{
			// 告警相关路由
			alerts := protected.Group("/alerts")
			{
				alerts.GET("", g.listAlerts)
				alerts.POST("", g.createAlert)
				alerts.GET("/:id", g.getAlert)
				alerts.PUT("/:id", g.updateAlert)
				alerts.DELETE("/:id", g.deleteAlert)
				alerts.POST("/:id/acknowledge", g.acknowledgeAlert)
				alerts.POST("/:id/resolve", g.resolveAlert)
			}

			// 规则相关路由
			rules := protected.Group("/rules")
			{
				rules.GET("", g.listRules)
				rules.POST("", g.createRule)
				rules.GET("/:id", g.getRule)
				rules.PUT("/:id", g.updateRule)
				rules.DELETE("/:id", g.deleteRule)
				rules.POST("/:id/enable", g.enableRule)
				rules.POST("/:id/disable", g.disableRule)
			}

			// 数据源相关路由
			datasources := protected.Group("/datasources")
			{
				datasources.GET("", g.listDataSources)
				datasources.POST("", g.createDataSource)
				datasources.GET("/:id", g.getDataSource)
				datasources.PUT("/:id", g.updateDataSource)
				datasources.DELETE("/:id", g.deleteDataSource)
				datasources.POST("/:id/test", g.testDataSource)
			}

			// 工单相关路由
			tickets := protected.Group("/tickets")
			{
				tickets.GET("", g.listTickets)
				tickets.POST("", g.createTicket)
				tickets.GET("/:id", g.getTicket)
				tickets.PUT("/:id", g.updateTicket)
				tickets.DELETE("/:id", g.deleteTicket)
				tickets.POST("/:id/assign", g.assignTicket)
			}

			// 知识库相关路由
			knowledge := protected.Group("/knowledge")
			{
				knowledge.GET("", g.listKnowledge)
				knowledge.POST("", g.createKnowledge)
				knowledge.GET("/:id", g.getKnowledge)
				knowledge.PUT("/:id", g.updateKnowledge)
				knowledge.DELETE("/:id", g.deleteKnowledge)
				knowledge.GET("/search", g.searchKnowledge)
			}

			// 用户相关路由
			users := protected.Group("/users")
			{
				users.GET("", g.listUsers)
				users.POST("", g.createUser)
				users.GET("/:id", g.getUser)
				users.PUT("/:id", g.updateUser)
				users.DELETE("/:id", g.deleteUser)
			}

			// Webhook相关路由
			webhooks := protected.Group("/webhooks")
			{
				webhooks.GET("", g.listWebhooks)
				webhooks.POST("", g.createWebhook)
				webhooks.GET("/:id", g.getWebhook)
				webhooks.PUT("/:id", g.updateWebhook)
				webhooks.DELETE("/:id", g.deleteWebhook)
			}

			// 系统配置相关路由
			config := protected.Group("/config")
			{
				config.GET("", g.listConfig)
				config.POST("", g.setConfig)
				config.DELETE("/:key", g.deleteConfig)
			}

			// Worker状态路由
			workers := protected.Group("/workers")
			{
				workers.GET("/status", g.getWorkerStatus)
			}
		}
	}

	return g.router
}

// RegisterMiddleware 注册中间件
func (g *gateway) RegisterMiddleware(middleware gin.HandlerFunc) {
	g.middlewares = append(g.middlewares, middleware)
}

// registerDefaultMiddlewares 注册默认中间件
func (g *gateway) registerDefaultMiddlewares() {
	// 请求ID中间件
	g.RegisterMiddleware(middleware.RequestIDMiddleware())

	// 日志中间件
	g.RegisterMiddleware(middleware.LoggingMiddleware(g.logger))

	// 恢复中间件
	g.RegisterMiddleware(middleware.RecoveryMiddleware(g.logger))

	// CORS中间件
	g.RegisterMiddleware(middleware.CORSMiddleware())

	// 指标收集中间件
	g.RegisterMiddleware(middleware.MetricsMiddleware())

	// 基础限流中间件（如果没有Redis，使用内存限流）
	if g.rateLimiter == nil {
		g.RegisterMiddleware(middleware.RateLimitMiddleware(middleware.RateLimitConfig{
			RequestsPerMinute: 60,
			BurstSize:         10,
		}))
	}
}

// loggerMiddleware 日志中间件
func (g *gateway) loggerMiddleware() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		g.logger.Info("HTTP Request",
			zap.String("method", param.Method),
			zap.String("path", param.Path),
			zap.Int("status", param.StatusCode),
			zap.Duration("latency", param.Latency),
			zap.String("ip", param.ClientIP),
			zap.String("user_agent", param.Request.UserAgent()),
		)
		return ""
	})
}

// corsMiddleware CORS中间件
func (g *gateway) corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// rateLimitMiddleware 限流中间件
func (g *gateway) rateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: 实现限流逻辑
		c.Next()
	}
}

// authMiddleware 认证中间件（JWT）
func (g *gateway) authMiddleware() gin.HandlerFunc {
	return middleware.AuthMiddleware(g.jwtSecret)
}

// apiKeyAuthMiddleware API Key认证中间件
func (g *gateway) apiKeyAuthMiddleware() gin.HandlerFunc {
	return middleware.APIKeyMiddleware(g.apiKeys)
}

// rbacMiddleware RBAC权限控制中间件
func (g *gateway) rbacMiddleware(roles ...string) gin.HandlerFunc {
	return middleware.RoleBasedAccessControl(roles...)
}

// redisRateLimitMiddleware Redis分布式限流中间件
func (g *gateway) redisRateLimitMiddleware(config middleware.RedisRateLimitConfig) gin.HandlerFunc {
	if g.rateLimiter != nil {
		return g.rateLimiter.RedisRateLimitMiddleware(config)
	}
	// 如果没有Redis，使用内存限流
	return middleware.RateLimitMiddleware(middleware.RateLimitConfig{
		RequestsPerMinute: config.RequestsPerMinute,
		BurstSize:         config.BurstSize,
	})
}

// circuitBreakerMiddleware 熔断器中间件
func (g *gateway) circuitBreakerMiddleware(serviceName string, config middleware.CircuitBreakerConfig) gin.HandlerFunc {
	if g.rateLimiter != nil {
		return g.rateLimiter.CircuitBreakerMiddleware(serviceName, config)
	}
	// 如果没有Redis，跳过熔断器
	return func(c *gin.Context) {
		c.Next()
	}
}

// GetRouter 获取Gin路由器
func (g *gateway) GetRouter() *gin.Engine {
	return g.router
}