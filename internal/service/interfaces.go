package service

import (
	"context"

	"pulse/internal/models"
)

// AlertService 告警服务接口
type AlertService interface {
	Create(ctx context.Context, alert *models.Alert) error
	GetByID(ctx context.Context, id string) (*models.Alert, error)
	List(ctx context.Context, filter *models.AlertFilter) ([]*models.Alert, int64, error)
	Update(ctx context.Context, alert *models.Alert) error
	Delete(ctx context.Context, id string) error
	Acknowledge(ctx context.Context, id string, userID string) error
	Resolve(ctx context.Context, id string, userID string) error
}

// RuleService 规则服务接口
type RuleService interface {
	Create(ctx context.Context, rule *models.Rule) error
	GetByID(ctx context.Context, id string) (*models.Rule, error)
	List(ctx context.Context, filter *models.RuleFilter) ([]*models.Rule, int64, error)
	Update(ctx context.Context, rule *models.Rule) error
	Delete(ctx context.Context, id string) error
	Enable(ctx context.Context, id string) error
	Disable(ctx context.Context, id string) error
}

// DataSourceService 数据源服务接口
type DataSourceService interface {
	Create(ctx context.Context, dataSource *models.DataSource) error
	GetByID(ctx context.Context, id string) (*models.DataSource, error)
	List(ctx context.Context, filter *models.DataSourceFilter) ([]*models.DataSource, int64, error)
	Update(ctx context.Context, dataSource *models.DataSource) error
	Delete(ctx context.Context, id string) error
	TestConnection(ctx context.Context, id string) error
}

// TicketService 工单服务接口
type TicketService interface {
	Create(ctx context.Context, ticket *models.Ticket) error
	GetByID(ctx context.Context, id string) (*models.Ticket, error)
	List(ctx context.Context, filter *models.TicketFilter) ([]*models.Ticket, int64, error)
	Update(ctx context.Context, ticket *models.Ticket) error
	Delete(ctx context.Context, id string) error
	Assign(ctx context.Context, id string, assigneeID string) error
	UpdateStatus(ctx context.Context, id string, status models.TicketStatus) error
}

// KnowledgeService 知识库服务接口
type KnowledgeService interface {
	Create(ctx context.Context, knowledge *models.Knowledge) error
	GetByID(ctx context.Context, id string) (*models.Knowledge, error)
	List(ctx context.Context, filter *models.KnowledgeFilter) ([]*models.Knowledge, int64, error)
	Update(ctx context.Context, knowledge *models.Knowledge) error
	Delete(ctx context.Context, id string) error
	Search(ctx context.Context, query string) ([]*models.Knowledge, error)
}

// UserService 用户服务接口
type UserService interface {
	Create(ctx context.Context, user *models.User) error
	GetByID(ctx context.Context, id string) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	List(ctx context.Context, filter *models.UserFilter) ([]*models.User, int64, error)
	Update(ctx context.Context, user *models.User) error
	Delete(ctx context.Context, id string) error
	UpdatePassword(ctx context.Context, id string, oldPassword, newPassword string) error
}

// AuthService 认证服务接口
type AuthService interface {
	Login(ctx context.Context, email, password string) (*models.AuthToken, error)
	RefreshToken(ctx context.Context, refreshToken string) (*models.AuthToken, error)
	Logout(ctx context.Context, token string) error
	ValidateToken(ctx context.Context, token string) (*models.User, error)
	ResetPassword(ctx context.Context, email string) error
}

// NotificationService 通知服务接口
type NotificationService interface {
	Send(ctx context.Context, notification *models.Notification) error
	SendBatch(ctx context.Context, notifications []*models.Notification) error
	GetTemplates(ctx context.Context) ([]*models.NotificationTemplate, error)
	CreateTemplate(ctx context.Context, template *models.NotificationTemplate) error
}

// WebhookService Webhook服务接口
type WebhookService interface {
	Create(ctx context.Context, webhook *models.Webhook) error
	GetByID(ctx context.Context, id string) (*models.Webhook, error)
	List(ctx context.Context, filter *models.WebhookFilter) ([]*models.Webhook, int64, error)
	Update(ctx context.Context, webhook *models.Webhook) error
	Delete(ctx context.Context, id string) error
	Trigger(ctx context.Context, id string, payload interface{}) error
}

// ConfigService 配置服务接口
type ConfigService interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key, value string) error
	Delete(ctx context.Context, key string) error
	List(ctx context.Context, prefix string) (map[string]string, error)
}