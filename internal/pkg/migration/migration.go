package migration

import (
	"context"
	"database/sql"
	"fmt"
	"path/filepath"
	"time"

	"pulse/internal/config"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

// Migrator 迁移管理器
type Migrator struct {
	migrate *migrate.Migrate
	config  *config.MigrationConfig
	db      *sql.DB
}

// New 创建新的迁移管理器
func New(db *sql.DB, cfg *config.MigrationConfig) (*Migrator, error) {
	// 创建 postgres 驱动实例
	driver, err := postgres.WithInstance(db, &postgres.Config{
		MigrationsTable: cfg.Table,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create postgres driver: %w", err)
	}

	// 获取迁移文件路径
	migrationPath, err := filepath.Abs(cfg.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute migration path: %w", err)
	}

	// 创建 migrate 实例
	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", migrationPath),
		"postgres",
		driver,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create migrate instance: %w", err)
	}

	return &Migrator{
		migrate: m,
		config:  cfg,
		db:      db,
	}, nil
}

// Up 执行所有待执行的迁移
func (m *Migrator) Up(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, m.config.LockTimeout)
	defer cancel()

	err := m.migrate.Up()
	if err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}

// Down 回滚所有迁移
func (m *Migrator) Down(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, m.config.LockTimeout)
	defer cancel()

	err := m.migrate.Down()
	if err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to rollback migrations: %w", err)
	}

	return nil
}

// Steps 执行指定步数的迁移
func (m *Migrator) Steps(ctx context.Context, n int) error {
	ctx, cancel := context.WithTimeout(ctx, m.config.LockTimeout)
	defer cancel()

	err := m.migrate.Steps(n)
	if err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run %d migration steps: %w", n, err)
	}

	return nil
}

// Migrate 迁移到指定版本
func (m *Migrator) Migrate(ctx context.Context, version uint) error {
	ctx, cancel := context.WithTimeout(ctx, m.config.LockTimeout)
	defer cancel()

	err := m.migrate.Migrate(version)
	if err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to migrate to version %d: %w", version, err)
	}

	return nil
}

// Version 获取当前迁移版本
func (m *Migrator) Version() (uint, bool, error) {
	return m.migrate.Version()
}

// Force 强制设置迁移版本（用于修复损坏的迁移状态）
func (m *Migrator) Force(version int) error {
	err := m.migrate.Force(version)
	if err != nil {
		return fmt.Errorf("failed to force version %d: %w", version, err)
	}

	return nil
}

// Drop 删除所有表和迁移历史
func (m *Migrator) Drop(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, m.config.LockTimeout)
	defer cancel()

	err := m.migrate.Drop()
	if err != nil {
		return fmt.Errorf("failed to drop database: %w", err)
	}

	return nil
}

// Close 关闭迁移管理器
func (m *Migrator) Close() error {
	sourceErr, dbErr := m.migrate.Close()
	if sourceErr != nil {
		return fmt.Errorf("failed to close migration source: %w", sourceErr)
	}
	if dbErr != nil {
		return fmt.Errorf("failed to close migration database: %w", dbErr)
	}

	return nil
}

// Status 获取迁移状态信息
func (m *Migrator) Status() (*MigrationStatus, error) {
	version, dirty, err := m.Version()
	if err != nil {
		return nil, fmt.Errorf("failed to get migration version: %w", err)
	}

	return &MigrationStatus{
		Version:   version,
		Dirty:     dirty,
		Timestamp: time.Now(),
	}, nil
}

// MigrationStatus 迁移状态
type MigrationStatus struct {
	Version   uint      `json:"version"`
	Dirty     bool      `json:"dirty"`
	Timestamp time.Time `json:"timestamp"`
}

// IsUpToDate 检查是否为最新版本
func (s *MigrationStatus) IsUpToDate() bool {
	return !s.Dirty
}

// NeedsMigration 检查是否需要迁移
func (m *Migrator) NeedsMigration() (bool, error) {
	_, dirty, err := m.Version()
	if err != nil {
		if err == migrate.ErrNilVersion {
			// 没有迁移历史，需要执行迁移
			return true, nil
		}
		return false, fmt.Errorf("failed to check migration status: %w", err)
	}

	// 如果状态是 dirty，说明上次迁移失败，需要修复
	return dirty, nil
}

// CreateMigrationFile 创建新的迁移文件
func CreateMigrationFile(migrationPath, name string) (string, string, error) {
	timestamp := time.Now().Unix()
	upFile := filepath.Join(migrationPath, fmt.Sprintf("%d_%s.up.sql", timestamp, name))
	downFile := filepath.Join(migrationPath, fmt.Sprintf("%d_%s.down.sql", timestamp, name))

	return upFile, downFile, nil
}
