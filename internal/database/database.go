package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"go.uber.org/zap"

	"Pulse/internal/config"
)

// DB 数据库连接包装器
type DB struct {
	*sqlx.DB
	config *config.DatabaseConfig
	logger *zap.Logger
}

// NewConnection 创建新的数据库连接（兼容性函数）
func NewConnection(cfg *config.DatabaseConfig) (*DB, error) {
	// 创建一个默认的logger
	logger, _ := zap.NewDevelopment()
	return New(cfg, logger)
}

// New 创建新的数据库连接
func New(cfg *config.DatabaseConfig, logger *zap.Logger) (*DB, error) {
	if cfg == nil {
		return nil, fmt.Errorf("database config is required")
	}

	if logger == nil {
		return nil, fmt.Errorf("logger is required")
	}

	// 连接数据库
	db, err := sqlx.Connect("postgres", cfg.GetDSN())
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// 配置连接池
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	db.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logger.Info("Database connected successfully",
		zap.String("dsn", cfg.GetDSNWithoutPassword()),
		zap.Int("max_open_conns", cfg.MaxOpenConns),
		zap.Int("max_idle_conns", cfg.MaxIdleConns),
		zap.Duration("conn_max_lifetime", cfg.ConnMaxLifetime),
		zap.Duration("conn_max_idle_time", cfg.ConnMaxIdleTime),
	)

	return &DB{
		DB:     db,
		config: cfg,
		logger: logger,
	}, nil
}

// Close 关闭数据库连接
func (db *DB) Close() error {
	if db.DB != nil {
		db.logger.Info("Closing database connection")
		return db.DB.Close()
	}
	return nil
}

// Health 检查数据库健康状态
func (db *DB) Health(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// 检查连接
	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}

	// 检查连接池状态
	stats := db.Stats()
	db.logger.Debug("Database connection pool stats",
		zap.Int("open_connections", stats.OpenConnections),
		zap.Int("in_use", stats.InUse),
		zap.Int("idle", stats.Idle),
		zap.Int64("wait_count", stats.WaitCount),
		zap.Duration("wait_duration", stats.WaitDuration),
		zap.Int64("max_idle_closed", stats.MaxIdleClosed),
		zap.Int64("max_idle_time_closed", stats.MaxIdleTimeClosed),
		zap.Int64("max_lifetime_closed", stats.MaxLifetimeClosed),
	)

	// 检查是否有过多的等待
	if stats.WaitCount > 1000 {
		db.logger.Warn("High database connection wait count",
			zap.Int64("wait_count", stats.WaitCount),
			zap.Duration("wait_duration", stats.WaitDuration),
		)
	}

	return nil
}

// GetStats 获取数据库连接池统计信息
func (db *DB) GetStats() sql.DBStats {
	return db.Stats()
}

// RunMigrations 运行数据库迁移
func (db *DB) RunMigrations() error {
	db.logger.Info("Starting database migrations",
		zap.String("migration_path", db.config.MigrationPath),
		zap.String("migration_table", db.config.MigrationTable),
	)

	// 创建 postgres 驱动实例
	driver, err := postgres.WithInstance(db.DB.DB, &postgres.Config{
		MigrationsTable: db.config.MigrationTable,
	})
	if err != nil {
		return fmt.Errorf("failed to create postgres driver: %w", err)
	}

	// 创建 migrate 实例
	m, err := migrate.NewWithDatabaseInstance(
		db.config.MigrationPath,
		"postgres",
		driver,
	)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}
	defer m.Close()

	// 获取当前版本
	currentVersion, dirty, err := m.Version()
	if err != nil && err != migrate.ErrNilVersion {
		return fmt.Errorf("failed to get current migration version: %w", err)
	}

	db.logger.Info("Current migration status",
		zap.Uint("version", currentVersion),
		zap.Bool("dirty", dirty),
	)

	if dirty {
		return fmt.Errorf("database is in dirty state, manual intervention required")
	}

	// 运行迁移
	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	if err == migrate.ErrNoChange {
		db.logger.Info("No new migrations to apply")
	} else {
		// 获取新版本
		newVersion, _, err := m.Version()
		if err != nil {
			return fmt.Errorf("failed to get new migration version: %w", err)
		}

		db.logger.Info("Migrations applied successfully",
			zap.Uint("from_version", currentVersion),
			zap.Uint("to_version", newVersion),
		)
	}

	return nil
}

// RollbackMigrations 回滚数据库迁移
func (db *DB) RollbackMigrations(steps int) error {
	if steps <= 0 {
		return fmt.Errorf("steps must be positive")
	}

	db.logger.Info("Starting database migration rollback",
		zap.Int("steps", steps),
	)

	// 创建 postgres 驱动实例
	driver, err := postgres.WithInstance(db.DB.DB, &postgres.Config{
		MigrationsTable: db.config.MigrationTable,
	})
	if err != nil {
		return fmt.Errorf("failed to create postgres driver: %w", err)
	}

	// 创建 migrate 实例
	m, err := migrate.NewWithDatabaseInstance(
		db.config.MigrationPath,
		"postgres",
		driver,
	)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}
	defer m.Close()

	// 获取当前版本
	currentVersion, dirty, err := m.Version()
	if err != nil {
		return fmt.Errorf("failed to get current migration version: %w", err)
	}

	if dirty {
		return fmt.Errorf("database is in dirty state, manual intervention required")
	}

	// 回滚指定步数
	err = m.Steps(-steps)
	if err != nil {
		return fmt.Errorf("failed to rollback migrations: %w", err)
	}

	// 获取新版本
	newVersion, _, err := m.Version()
	if err != nil && err != migrate.ErrNilVersion {
		return fmt.Errorf("failed to get new migration version: %w", err)
	}

	db.logger.Info("Migrations rolled back successfully",
		zap.Uint("from_version", currentVersion),
		zap.Uint("to_version", newVersion),
		zap.Int("steps", steps),
	)

	return nil
}

// MigrationStatus 获取迁移状态
func (db *DB) MigrationStatus() (version uint, dirty bool, err error) {
	// 创建 postgres 驱动实例
	driver, err := postgres.WithInstance(db.DB.DB, &postgres.Config{
		MigrationsTable: db.config.MigrationTable,
	})
	if err != nil {
		return 0, false, fmt.Errorf("failed to create postgres driver: %w", err)
	}

	// 创建 migrate 实例
	m, err := migrate.NewWithDatabaseInstance(
		db.config.MigrationPath,
		"postgres",
		driver,
	)
	if err != nil {
		return 0, false, fmt.Errorf("failed to create migrate instance: %w", err)
	}
	defer m.Close()

	// 获取当前版本
	version, dirty, err = m.Version()
	if err != nil && err != migrate.ErrNilVersion {
		return 0, false, fmt.Errorf("failed to get migration version: %w", err)
	}

	return version, dirty, nil
}

// BeginTx 开始事务
func (db *DB) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sqlx.Tx, error) {
	return db.BeginTxx(ctx, opts)
}

// WithTransaction 在事务中执行函数
func (db *DB) WithTransaction(ctx context.Context, fn func(*sqlx.Tx) error) error {
	tx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				db.logger.Error("Failed to rollback transaction after panic",
					zap.Error(rollbackErr),
					zap.Any("panic", p),
				)
			}
			panic(p)
		}
	}()

	if err := fn(tx); err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			db.logger.Error("Failed to rollback transaction",
				zap.Error(rollbackErr),
				zap.Error(err),
			)
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// ExecContext 执行 SQL 语句
func (db *DB) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	start := time.Now()
	result, err := db.DB.ExecContext(ctx, query, args...)
	duration := time.Since(start)

	db.logQuery("EXEC", query, args, duration, err)
	return result, err
}

// QueryContext 查询数据
func (db *DB) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	start := time.Now()
	rows, err := db.DB.QueryContext(ctx, query, args...)
	duration := time.Since(start)

	db.logQuery("QUERY", query, args, duration, err)
	return rows, err
}

// QueryRowContext 查询单行数据
func (db *DB) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	start := time.Now()
	row := db.DB.QueryRowContext(ctx, query, args...)
	duration := time.Since(start)

	db.logQuery("QUERY_ROW", query, args, duration, nil)
	return row
}

// logQuery 记录查询日志
func (db *DB) logQuery(operation, query string, args []interface{}, duration time.Duration, err error) {
	fields := []zap.Field{
		zap.String("operation", operation),
		zap.String("query", query),
		zap.Duration("duration", duration),
	}

	if len(args) > 0 {
		fields = append(fields, zap.Any("args", args))
	}

	if err != nil {
		fields = append(fields, zap.Error(err))
		db.logger.Error("Database query failed", fields...)
	} else {
		// 只在调试模式下记录成功的查询
		if duration > 100*time.Millisecond {
			db.logger.Warn("Slow database query", fields...)
		} else {
			db.logger.Debug("Database query executed", fields...)
		}
	}
}
