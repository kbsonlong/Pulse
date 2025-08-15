package repository

import (
	"context"
	"time"

	"Pulse/internal/models"
)

// UserRepository 用户仓储接口
type UserRepository interface {
	// 基础CRUD操作
	Create(ctx context.Context, user *models.User) error
	GetByID(ctx context.Context, id string) (*models.User, error)
	GetByUsername(ctx context.Context, username string) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	Update(ctx context.Context, user *models.User) error
	Delete(ctx context.Context, id string) error
	SoftDelete(ctx context.Context, id string) error
	
	// 查询操作
	List(ctx context.Context, filter *models.UserFilter) (*models.UserList, error)
	Count(ctx context.Context, filter *models.UserFilter) (int64, error)
	Exists(ctx context.Context, id string) (bool, error)
	ExistsByUsername(ctx context.Context, username string) (bool, error)
	ExistsByEmail(ctx context.Context, email string) (bool, error)
	
	// 认证相关
	VerifyPassword(ctx context.Context, username, password string) (*models.User, error)
	UpdatePassword(ctx context.Context, id, hashedPassword string) error
	UpdateLastLogin(ctx context.Context, id string, loginTime time.Time) error
	
	// 状态管理
	UpdateStatus(ctx context.Context, id string, status models.UserStatus) error
	Activate(ctx context.Context, id string) error
	Deactivate(ctx context.Context, id string) error
	
	// 批量操作
	BatchCreate(ctx context.Context, users []*models.User) error
	BatchUpdate(ctx context.Context, users []*models.User) error
	BatchDelete(ctx context.Context, ids []string) error
}

// AlertRepository 告警仓储接口
type AlertRepository interface {
	// 基础CRUD操作
	Create(ctx context.Context, alert *models.Alert) error
	GetByID(ctx context.Context, id string) (*models.Alert, error)
	Update(ctx context.Context, alert *models.Alert) error
	Delete(ctx context.Context, id string) error
	SoftDelete(ctx context.Context, id string) error
	
	// 查询操作
	List(ctx context.Context, filter *models.AlertFilter) (*models.AlertList, error)
	Count(ctx context.Context, filter *models.AlertFilter) (int64, error)
	Exists(ctx context.Context, id string) (bool, error)
	GetByFingerprint(ctx context.Context, fingerprint string) (*models.Alert, error)
	
	// 告警状态管理
	Acknowledge(ctx context.Context, id, userID string, comment *string) error
	Resolve(ctx context.Context, id, userID string, comment *string) error
	Silence(ctx context.Context, id string, silenceID string, duration time.Duration) error
	Unsilence(ctx context.Context, id string) error
	
	// 告警统计
	GetStats(ctx context.Context, filter *models.AlertFilter) (*models.AlertStats, error)
	GetTrend(ctx context.Context, start, end time.Time, interval string) ([]*models.AlertTrendPoint, error)
	GetActiveCount(ctx context.Context) (int64, error)
	GetCriticalCount(ctx context.Context) (int64, error)
	
	// 告警历史
	GetHistory(ctx context.Context, alertID string) ([]*models.AlertHistory, error)
	AddHistory(ctx context.Context, history *models.AlertHistory) error
	
	// 批量操作
	BatchCreate(ctx context.Context, alerts []*models.Alert) error
	BatchUpdate(ctx context.Context, alerts []*models.Alert) error
	BatchAcknowledge(ctx context.Context, ids []string, userID string, comment *string) error
	BatchResolve(ctx context.Context, ids []string, userID string, comment *string) error
	
	// 清理操作
	CleanupResolved(ctx context.Context, before time.Time) (int64, error)
	CleanupExpired(ctx context.Context) (int64, error)
}

// RuleRepository 规则仓储接口
type RuleRepository interface {
	// 基础CRUD操作
	Create(ctx context.Context, rule *models.Rule) error
	GetByID(ctx context.Context, id string) (*models.Rule, error)
	Update(ctx context.Context, rule *models.Rule) error
	Delete(ctx context.Context, id string) error
	SoftDelete(ctx context.Context, id string) error
	
	// 查询操作
	List(ctx context.Context, filter *models.RuleFilter) (*models.RuleList, error)
	Count(ctx context.Context, filter *models.RuleFilter) (int64, error)
	Exists(ctx context.Context, id string) (bool, error)
	GetByName(ctx context.Context, name string) (*models.Rule, error)
	
	// 规则状态管理
	Activate(ctx context.Context, id string) error
	Deactivate(ctx context.Context, id string) error
	Enable(ctx context.Context, id string) error
	Disable(ctx context.Context, id string) error
	SetTesting(ctx context.Context, id string) error
	
	// 规则评估
	GetActiveRules(ctx context.Context) ([]*models.Rule, error)
	GetRulesForEvaluation(ctx context.Context) ([]*models.Rule, error)
	UpdateLastEvaluation(ctx context.Context, id string, evalTime time.Time, result bool, error string) error
	IncrementEvaluationCount(ctx context.Context, id string) error
	IncrementAlertCount(ctx context.Context, id string) error
	
	// 规则统计
	GetStats(ctx context.Context, filter *models.RuleFilter) (*models.RuleStats, error)
	GetActiveCount(ctx context.Context) (int64, error)
	GetErrorCount(ctx context.Context) (int64, error)
	
	// 规则测试
	TestRule(ctx context.Context, rule *models.Rule) (*models.RuleTestResult, error)
	
	// 批量操作
	BatchCreate(ctx context.Context, rules []*models.Rule) error
	BatchUpdate(ctx context.Context, rules []*models.Rule) error
	BatchActivate(ctx context.Context, ids []string) error
	BatchDeactivate(ctx context.Context, ids []string) error
}

// DataSourceRepository 数据源仓储接口
type DataSourceRepository interface {
	// 基础CRUD操作
	Create(ctx context.Context, dataSource *models.DataSource) error
	GetByID(ctx context.Context, id string) (*models.DataSource, error)
	Update(ctx context.Context, dataSource *models.DataSource) error
	Delete(ctx context.Context, id string) error
	SoftDelete(ctx context.Context, id string) error
	
	// 查询操作
	List(ctx context.Context, filter *models.DataSourceFilter) (*models.DataSourceList, error)
	Count(ctx context.Context, filter *models.DataSourceFilter) (int64, error)
	Exists(ctx context.Context, id string) (bool, error)
	GetByName(ctx context.Context, name string) (*models.DataSource, error)
	GetByType(ctx context.Context, dsType models.DataSourceType) ([]*models.DataSource, error)
	
	// 数据源状态管理
	Activate(ctx context.Context, id string) error
	Deactivate(ctx context.Context, id string) error
	UpdateHealthStatus(ctx context.Context, id string, isHealthy bool, error string) error
	UpdateLastHealthCheck(ctx context.Context, id string, checkTime time.Time) error
	
	// 数据源测试
	TestConnection(ctx context.Context, dataSource *models.DataSource) (*models.DataSourceTestResult, error)
	Query(ctx context.Context, id string, query *models.DataSourceQuery) (*models.DataSourceQueryResult, error)
	
	// 数据源统计
	GetStats(ctx context.Context, filter *models.DataSourceFilter) (*models.DataSourceStats, error)
	GetActiveCount(ctx context.Context) (int64, error)
	GetHealthyCount(ctx context.Context) (int64, error)
	GetUnhealthyCount(ctx context.Context) (int64, error)
	
	// 数据源指标
	UpdateMetrics(ctx context.Context, id string, metrics *models.DataSourceMetrics) error
	GetMetrics(ctx context.Context, id string) (*models.DataSourceMetrics, error)
	
	// 批量操作
	BatchCreate(ctx context.Context, dataSources []*models.DataSource) error
	BatchUpdate(ctx context.Context, dataSources []*models.DataSource) error
	BatchHealthCheck(ctx context.Context, ids []string) error
}

// TicketRepository 工单仓储接口
type TicketRepository interface {
	// 基础CRUD操作
	Create(ctx context.Context, ticket *models.Ticket) error
	GetByID(ctx context.Context, id string) (*models.Ticket, error)
	Update(ctx context.Context, ticket *models.Ticket) error
	Delete(ctx context.Context, id string) error
	SoftDelete(ctx context.Context, id string) error
	
	// 查询操作
	List(ctx context.Context, filter *models.TicketFilter) (*models.TicketList, error)
	Count(ctx context.Context, filter *models.TicketFilter) (int64, error)
	Exists(ctx context.Context, id string) (bool, error)
	GetByAlertID(ctx context.Context, alertID string) ([]*models.Ticket, error)
	
	// 工单状态管理
	Assign(ctx context.Context, id, assigneeID string) error
	Unassign(ctx context.Context, id string) error
	UpdateStatus(ctx context.Context, id string, status models.TicketStatus) error
	UpdatePriority(ctx context.Context, id string, priority models.TicketPriority) error
	Resolve(ctx context.Context, id, resolverID string, solution *string) error
	Close(ctx context.Context, id, closerID string) error
	Reopen(ctx context.Context, id, reopenerID string) error
	
	// 工单评论
	AddComment(ctx context.Context, comment *models.TicketComment) error
	GetComments(ctx context.Context, ticketID string) ([]*models.TicketComment, error)
	UpdateComment(ctx context.Context, comment *models.TicketComment) error
	DeleteComment(ctx context.Context, id string) error
	
	// 工单附件
	AddAttachment(ctx context.Context, attachment *models.TicketAttachment) error
	GetAttachments(ctx context.Context, ticketID string) ([]*models.TicketAttachment, error)
	DeleteAttachment(ctx context.Context, id string) error
	
	// 工单历史
	GetHistory(ctx context.Context, ticketID string) ([]*models.TicketHistory, error)
	AddHistory(ctx context.Context, history *models.TicketHistory) error
	
	// 工单统计
	GetStats(ctx context.Context, filter *models.TicketFilter) (*models.TicketStats, error)
	GetTrend(ctx context.Context, start, end time.Time, interval string) ([]*models.TicketTrendPoint, error)
	GetOpenCount(ctx context.Context) (int64, error)
	GetOverdueCount(ctx context.Context) (int64, error)
	GetMyTickets(ctx context.Context, userID string, filter *models.TicketFilter) (*models.TicketList, error)
	
	// SLA管理
	UpdateSLA(ctx context.Context, id string, sla *models.TicketSLA) error
	GetSLA(ctx context.Context, id string) (*models.TicketSLA, error)
	GetOverdueSLA(ctx context.Context) ([]*models.Ticket, error)
	
	// 批量操作
	BatchCreate(ctx context.Context, tickets []*models.Ticket) error
	BatchUpdate(ctx context.Context, tickets []*models.Ticket) error
	BatchAssign(ctx context.Context, ids []string, assigneeID string) error
	BatchUpdateStatus(ctx context.Context, ids []string, status models.TicketStatus) error
	
	// 清理操作
	CleanupClosed(ctx context.Context, before time.Time) (int64, error)
}



// KnowledgeRepository 知识库仓储接口
type KnowledgeRepository interface {
	// 基础CRUD操作
	Create(ctx context.Context, knowledge *models.Knowledge) error
	GetByID(ctx context.Context, id string) (*models.Knowledge, error)
	GetBySlug(ctx context.Context, slug string) (*models.Knowledge, error)
	Update(ctx context.Context, knowledge *models.Knowledge) error
	Delete(ctx context.Context, id string) error
	SoftDelete(ctx context.Context, id string) error
	
	// 查询操作
	List(ctx context.Context, filter *models.KnowledgeFilter) (*models.KnowledgeList, error)
	Count(ctx context.Context, filter *models.KnowledgeFilter) (int64, error)
	Exists(ctx context.Context, id string) (bool, error)
	ExistsBySlug(ctx context.Context, slug string) (bool, error)
	Search(ctx context.Context, query string, filter *models.KnowledgeFilter) (*models.KnowledgeSearchResult, error)
	
	// 知识状态管理
	UpdateStatus(ctx context.Context, id string, status models.KnowledgeStatus) error
	Publish(ctx context.Context, id, publisherID string) error
	Unpublish(ctx context.Context, id string) error
	Archive(ctx context.Context, id string) error
	Unarchive(ctx context.Context, id string) error
	SubmitForReview(ctx context.Context, id string) error
	Approve(ctx context.Context, id, reviewerID string, comment *string) error
	Reject(ctx context.Context, id, reviewerID string, comment *string) error
	
	// 知识版本管理
	CreateVersion(ctx context.Context, version *models.KnowledgeVersion) error
	GetVersions(ctx context.Context, knowledgeID string) ([]*models.KnowledgeVersion, error)
	GetVersion(ctx context.Context, knowledgeID, version string) (*models.KnowledgeVersion, error)
	RestoreVersion(ctx context.Context, knowledgeID, version string) error
	
	// 知识分类管理
	CreateCategory(ctx context.Context, category *models.KnowledgeCategory) error
	GetCategories(ctx context.Context) ([]*models.KnowledgeCategory, error)
	GetCategory(ctx context.Context, id string) (*models.KnowledgeCategory, error)
	UpdateCategory(ctx context.Context, category *models.KnowledgeCategory) error
	DeleteCategory(ctx context.Context, id string) error
	GetKnowledgeByCategory(ctx context.Context, categoryID string, filter *models.KnowledgeFilter) (*models.KnowledgeList, error)
	
	// 知识标签管理
	CreateTag(ctx context.Context, tag *models.KnowledgeTag) error
	GetTags(ctx context.Context) ([]*models.KnowledgeTag, error)
	GetTag(ctx context.Context, id string) (*models.KnowledgeTag, error)
	UpdateTag(ctx context.Context, tag *models.KnowledgeTag) error
	DeleteTag(ctx context.Context, id string) error
	GetKnowledgeByTag(ctx context.Context, tagName string, filter *models.KnowledgeFilter) (*models.KnowledgeList, error)
	UpdateTagUsage(ctx context.Context, tagName string, delta int64) error
	
	// 知识附件管理
	AddAttachment(ctx context.Context, attachment *models.KnowledgeAttachment) error
	GetAttachments(ctx context.Context, knowledgeID string) ([]*models.KnowledgeAttachment, error)
	DeleteAttachment(ctx context.Context, id string) error
	
	// 知识指标管理
	IncrementViewCount(ctx context.Context, id string) error
	IncrementLikeCount(ctx context.Context, id string) error
	IncrementDislikeCount(ctx context.Context, id string) error
	IncrementShareCount(ctx context.Context, id string) error
	IncrementDownloadCount(ctx context.Context, id string) error
	UpdateRating(ctx context.Context, id string, rating float64) error
	GetMetrics(ctx context.Context, id string) (*models.KnowledgeMetrics, error)
	
	// 知识统计
	GetStats(ctx context.Context, filter *models.KnowledgeFilter) (*models.KnowledgeStats, error)
	GetPopular(ctx context.Context, limit int) ([]*models.Knowledge, error)
	GetRecent(ctx context.Context, limit int) ([]*models.Knowledge, error)
	GetFeatured(ctx context.Context, limit int) ([]*models.Knowledge, error)
	GetRelated(ctx context.Context, knowledgeID string, limit int) ([]*models.Knowledge, error)
	
	// 批量操作
	BatchCreate(ctx context.Context, knowledge []*models.Knowledge) error
	BatchUpdate(ctx context.Context, knowledge []*models.Knowledge) error
	BatchPublish(ctx context.Context, ids []string, publisherID string) error
	BatchArchive(ctx context.Context, ids []string) error
	
	// 清理操作
	CleanupExpired(ctx context.Context) (int64, error)
	CleanupDrafts(ctx context.Context, before time.Time) (int64, error)
}

// PermissionRepository 权限仓储接口
type PermissionRepository interface {
	// 权限检查
	CheckPermission(ctx context.Context, userID string, permission models.Permission) (bool, error)
	CheckPermissions(ctx context.Context, userID string, permissions []models.Permission) (map[models.Permission]bool, error)
	GetUserPermissions(ctx context.Context, userID string) ([]models.Permission, error)
	
	// 权限组管理
	CreatePermissionGroup(ctx context.Context, group *models.PermissionGroup) error
	GetPermissionGroup(ctx context.Context, id string) (*models.PermissionGroup, error)
	UpdatePermissionGroup(ctx context.Context, group *models.PermissionGroup) error
	DeletePermissionGroup(ctx context.Context, id string) error
	ListPermissionGroups(ctx context.Context) ([]*models.PermissionGroup, error)
	
	// 用户权限覆盖管理
	CreatePermissionOverride(ctx context.Context, override *models.UserPermissionOverride) error
	GetPermissionOverride(ctx context.Context, id string) (*models.UserPermissionOverride, error)
	UpdatePermissionOverride(ctx context.Context, override *models.UserPermissionOverride) error
	DeletePermissionOverride(ctx context.Context, id string) error
	GetUserPermissionOverrides(ctx context.Context, userID string) ([]*models.UserPermissionOverride, error)
	
	// 权限覆盖操作
	GrantPermission(ctx context.Context, userID string, permission models.Permission, grantedBy, reason string, expiresAt *time.Time) error
	RevokePermission(ctx context.Context, userID string, permission models.Permission, revokedBy, reason string) error
	
	// 清理过期权限
	CleanupExpiredOverrides(ctx context.Context) (int64, error)
}

// Repository 仓储管理器接口
type Repository interface {
	// 获取各个仓储实例
	User() UserRepository
	Alert() AlertRepository
	Rule() RuleRepository
	DataSource() DataSourceRepository
	Ticket() TicketRepository
	Knowledge() KnowledgeRepository
	Permission() PermissionRepository
	
	// 事务管理
	BeginTx(ctx context.Context) (Repository, error)
	Commit() error
	Rollback() error
	
	// 健康检查
	HealthCheck(ctx context.Context) error
	
	// 关闭连接
	Close() error
}

// RepositoryManager 仓储管理器接口
type RepositoryManager interface {
	User() UserRepository
	Alert() AlertRepository
	Rule() RuleRepository
	DataSource() DataSourceRepository
	Ticket() TicketRepository
	Knowledge() KnowledgeRepository
	Permission() PermissionRepository
	Auth() AuthRepository

	// 事务管理
	BeginTx(ctx context.Context) (RepositoryManager, error)
	Commit() error
	Rollback() error
	Close() error
}