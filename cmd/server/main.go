package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	"Pulse/internal/config"
	"Pulse/internal/database"
)

func main() {
	// 初始化日志
	logger, err := initLogger()
	if err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	logger.Info("Starting Alert Management Platform")

	// 加载配置
	cfg, err := config.Load(".env")
	if err != nil {
		logger.Fatal("Failed to load config", zap.Error(err))
	}

	// 验证配置
	if err := cfg.Validate(); err != nil {
		logger.Fatal("Invalid config", zap.Error(err))
	}

	logger.Info("Configuration loaded",
		zap.String("environment", cfg.App.Environment),
		zap.String("version", cfg.App.Version),
		zap.String("address", cfg.GetServerAddress()),
	)

	// 连接数据库
	db, err := database.New(&cfg.Database, logger)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	// 运行数据库迁移
	if cfg.Database.AutoMigrate {
		logger.Info("Running database migrations")
		if err := db.RunMigrations(); err != nil {
			logger.Fatal("Failed to run migrations", zap.Error(err))
		}
		logger.Info("Database migrations completed")
	} else {
		logger.Info("Auto migration disabled, skipping")
	}

	// 检查数据库健康状态
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := db.Health(ctx); err != nil {
		logger.Fatal("Database health check failed", zap.Error(err))
	}
	logger.Info("Database health check passed")

	// 创建 HTTP 服务器
	server := &http.Server{
		Addr:         cfg.GetServerAddress(),
		Handler:      setupRoutes(db, logger),
		ReadTimeout:  cfg.Performance.ReadTimeout,
		WriteTimeout: cfg.Performance.WriteTimeout,
		IdleTimeout:  cfg.Performance.IdleTimeout,
	}

	// 启动服务器
	go func() {
		logger.Info("Starting HTTP server", zap.String("address", server.Addr))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// 优雅关闭
	ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", zap.Error(err))
	} else {
		logger.Info("Server exited gracefully")
	}
}

// initLogger 初始化日志器
func initLogger() (*zap.Logger, error) {
	env := os.Getenv("APP_ENVIRONMENT")
	if env == "" {
		env = "development"
	}

	if env == "production" {
		return zap.NewProduction()
	}
	return zap.NewDevelopment()
}

// setupRoutes 设置路由
func setupRoutes(db *database.DB, logger *zap.Logger) http.Handler {
	mux := http.NewServeMux()

	// 健康检查端点
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		// 检查数据库健康状态
		if err := db.Health(ctx); err != nil {
			logger.Error("Health check failed", zap.Error(err))
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte(`{"status":"unhealthy","error":"` + err.Error() + `"}`))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy"}`))
	})

	// 数据库状态端点
	mux.HandleFunc("/db/status", func(w http.ResponseWriter, r *http.Request) {
		stats := db.GetStats()
		version, dirty, err := db.MigrationStatus()

		response := fmt.Sprintf(`{
			"migration_version": %d,
			"migration_dirty": %t,
			"migration_error": %q,
			"open_connections": %d,
			"in_use": %d,
			"idle": %d,
			"wait_count": %d,
			"wait_duration": "%s",
			"max_idle_closed": %d,
			"max_idle_time_closed": %d,
			"max_lifetime_closed": %d
		}`,
			version, dirty, getErrorString(err),
			stats.OpenConnections, stats.InUse, stats.Idle,
			stats.WaitCount, stats.WaitDuration.String(),
			stats.MaxIdleClosed, stats.MaxIdleTimeClosed, stats.MaxLifetimeClosed,
		)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(response))
	})

	// 根路径
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message":"Alert Management Platform API","version":"v1.0.0"}`))
	})

	return mux
}

// getErrorString 获取错误字符串
func getErrorString(err error) string {
	if err != nil {
		return err.Error()
	}
	return ""
}
