package repository

import (
	"context"

	"github.com/jmoiron/sqlx"
	"Pulse/internal/crypto"
)

// repositoryManager 仓储管理器实现
type repositoryManager struct {
	db *sqlx.DB
	tx *sqlx.Tx
	encryptionService crypto.EncryptionService

	// 仓储实例
	userRepo       UserRepository
	alertRepo      AlertRepository
	ruleRepo       RuleRepository
	dataSourceRepo DataSourceRepository
	ticketRepo     TicketRepository
	knowledgeRepo  KnowledgeRepository
	permissionRepo PermissionRepository
	authRepo       AuthRepository
}

// NewRepositoryManager 创建新的仓储管理器
func NewRepositoryManager(db *sqlx.DB, encryptionService crypto.EncryptionService) RepositoryManager {
	return &repositoryManager{
		db: db,
		encryptionService: encryptionService,
		userRepo:       NewUserRepository(db),
		alertRepo:      NewAlertRepository(db),
		ruleRepo:       NewRuleRepository(db),
		dataSourceRepo: NewDataSourceRepository(db, encryptionService),
		ticketRepo:     NewTicketRepository(db),
		knowledgeRepo:  NewKnowledgeRepository(db),
		permissionRepo: NewPermissionRepository(db),
		authRepo:       NewAuthRepository(db),
	}
}

// User 获取用户仓储
func (r *repositoryManager) User() UserRepository {
	return r.userRepo
}

// Alert 获取告警仓储
func (r *repositoryManager) Alert() AlertRepository {
	return r.alertRepo
}

// Rule 获取规则仓储
func (r *repositoryManager) Rule() RuleRepository {
	return r.ruleRepo
}

// DataSource 获取数据源仓储
func (r *repositoryManager) DataSource() DataSourceRepository {
	return r.dataSourceRepo
}

// Ticket 获取工单仓储
func (r *repositoryManager) Ticket() TicketRepository {
	return r.ticketRepo
}

// Knowledge 获取知识库仓储
func (r *repositoryManager) Knowledge() KnowledgeRepository {
	return r.knowledgeRepo
}

// Permission 获取权限仓储
func (r *repositoryManager) Permission() PermissionRepository {
	return r.permissionRepo
}

// Auth 获取认证仓储
func (r *repositoryManager) Auth() AuthRepository {
	return r.authRepo
}

// BeginTx 开始事务
func (r *repositoryManager) BeginTx(ctx context.Context) (RepositoryManager, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, err
	}

	return &repositoryManager{
		db: r.db,
		tx: tx,
		encryptionService: r.encryptionService,
		userRepo:       NewUserRepositoryWithTx(tx),
		alertRepo:      NewAlertRepositoryWithTx(tx),
		ruleRepo:       NewRuleRepositoryWithTx(tx),
		dataSourceRepo: NewDataSourceRepositoryWithTx(tx, r.encryptionService),
		ticketRepo:     NewTicketRepositoryWithTx(tx),
		knowledgeRepo:  NewKnowledgeRepositoryWithTx(tx),
		permissionRepo: NewPermissionRepositoryWithTx(tx),
		authRepo:       NewAuthRepositoryWithTx(tx),
	}, nil
}

// Commit 提交事务
func (r *repositoryManager) Commit() error {
	if r.tx == nil {
		return nil
	}
	return r.tx.Commit()
}

// Rollback 回滚事务
func (r *repositoryManager) Rollback() error {
	if r.tx == nil {
		return nil
	}
	return r.tx.Rollback()
}

// Close 关闭连接
func (r *repositoryManager) Close() error {
	if r.tx != nil {
		_ = r.tx.Rollback()
	}
	return r.db.Close()
}