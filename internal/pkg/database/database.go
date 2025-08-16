package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"pulse/internal/config"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // PostgreSQL 驱动
)

// DB 数据库连接管理器
type DB struct {
	*sqlx.DB
	config *config.PostgresConfig
}

// New 创建新的数据库连接
func New(cfg *config.PostgresConfig) (*DB, error) {
	dsn := cfg.GetDSN()

	db, err := sqlx.Connect("postgres", dsn)
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

	return &DB{
		DB:     db,
		config: cfg,
	}, nil
}

// Close 关闭数据库连接
func (db *DB) Close() error {
	return db.DB.Close()
}

// Ping 测试数据库连接
func (db *DB) Ping(ctx context.Context) error {
	return db.DB.PingContext(ctx)
}

// BeginTx 开始事务
func (db *DB) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sqlx.Tx, error) {
	return db.DB.BeginTxx(ctx, opts)
}

// WithTransaction 在事务中执行函数
func (db *DB) WithTransaction(ctx context.Context, fn func(*sqlx.Tx) error) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
	}()

	if err := fn(tx); err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}

// GetStats 获取数据库连接池统计信息
func (db *DB) GetStats() sql.DBStats {
	return db.DB.Stats()
}

// IsHealthy 检查数据库健康状态
func (db *DB) IsHealthy(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	// 执行简单查询测试连接
	var result int
	err := db.GetContext(ctx, &result, "SELECT 1")
	if err != nil {
		return fmt.Errorf("database health check failed: %w", err)
	}

	return nil
}

// EnableTimescaleDB 启用 TimescaleDB 扩展
func (db *DB) EnableTimescaleDB(ctx context.Context) error {
	// 创建 TimescaleDB 扩展
	_, err := db.ExecContext(ctx, "CREATE EXTENSION IF NOT EXISTS timescaledb CASCADE;")
	if err != nil {
		return fmt.Errorf("failed to create timescaledb extension: %w", err)
	}

	return nil
}

// CreateHypertable 创建超表
func (db *DB) CreateHypertable(ctx context.Context, tableName, timeColumn string, chunkTimeInterval string) error {
	query := fmt.Sprintf(
		"SELECT create_hypertable('%s', '%s', chunk_time_interval => INTERVAL '%s', if_not_exists => TRUE);",
		tableName, timeColumn, chunkTimeInterval,
	)

	_, err := db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to create hypertable %s: %w", tableName, err)
	}

	return nil
}

// SetRetentionPolicy 设置数据保留策略
func (db *DB) SetRetentionPolicy(ctx context.Context, tableName, retentionPeriod string) error {
	query := fmt.Sprintf(
		"SELECT add_retention_policy('%s', INTERVAL '%s', if_not_exists => TRUE);",
		tableName, retentionPeriod,
	)

	_, err := db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to set retention policy for %s: %w", tableName, err)
	}

	return nil
}
