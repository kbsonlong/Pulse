package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"

	"go.uber.org/zap"

	"Pulse/internal/config"
	"Pulse/internal/database"
)

func main() {
	// 定义命令行参数
	var (
		action  = flag.String("action", "up", "Migration action: up, down, status, version")
		steps   = flag.String("steps", "1", "Number of steps for down migration")
		envFile = flag.String("env", ".env", "Environment file path")
	)
	flag.Parse()

	// 初始化日志
	logger, err := zap.NewDevelopment()
	if err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	// 加载配置
	cfg, err := config.Load(*envFile)
	if err != nil {
		logger.Fatal("Failed to load config", zap.Error(err))
	}

	// 验证配置
	if err := cfg.Validate(); err != nil {
		logger.Fatal("Invalid config", zap.Error(err))
	}

	// 连接数据库
	db, err := database.New(&cfg.Database, logger)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	// 执行迁移操作
	switch *action {
	case "up":
		if err := db.RunMigrations(); err != nil {
			logger.Fatal("Failed to run migrations", zap.Error(err))
		}
		logger.Info("Migrations completed successfully")

	case "down":
		stepsInt, err := strconv.Atoi(*steps)
		if err != nil {
			logger.Fatal("Invalid steps value", zap.Error(err))
		}
		if err := db.RollbackMigrations(stepsInt); err != nil {
			logger.Fatal("Failed to rollback migrations", zap.Error(err))
		}
		logger.Info("Migrations rolled back successfully", zap.Int("steps", stepsInt))

	case "status", "version":
		version, dirty, err := db.MigrationStatus()
		if err != nil {
			logger.Fatal("Failed to get migration status", zap.Error(err))
		}
		logger.Info("Migration status",
			zap.Uint("version", version),
			zap.Bool("dirty", dirty),
		)
		fmt.Printf("Current migration version: %d\n", version)
		fmt.Printf("Dirty state: %t\n", dirty)

	default:
		logger.Fatal("Unknown action", zap.String("action", *action))
	}
}
