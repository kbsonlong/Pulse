package repository

import (
	"context"
	"time"

	"pulse/internal/models"
)

// NotificationRepository 通知仓储接口
type NotificationRepository interface {
	// 基础CRUD操作
	Create(ctx context.Context, notification *models.Notification) error
	GetByID(ctx context.Context, id string) (*models.Notification, error)
	Update(ctx context.Context, notification *models.Notification) error
	Delete(ctx context.Context, id string) error
	SoftDelete(ctx context.Context, id string) error

	// 查询操作
	List(ctx context.Context, filter *models.NotificationFilter) (*models.NotificationList, error)
	Count(ctx context.Context, filter *models.NotificationFilter) (int64, error)
	Exists(ctx context.Context, id string) (bool, error)
	GetByAlertID(ctx context.Context, alertID string) ([]*models.Notification, error)
	GetByRecipient(ctx context.Context, recipient string) ([]*models.Notification, error)

	// 通知状态管理
	UpdateStatus(ctx context.Context, id string, status models.NotificationStatus) error
	MarkAsSent(ctx context.Context, id string, sentAt time.Time) error
	MarkAsFailed(ctx context.Context, id string, errorMsg string) error
	IncrementRetryCount(ctx context.Context, id string) error

	// 通知统计
	GetStats(ctx context.Context, filter *models.NotificationFilter) (*models.NotificationStats, error)
	GetSentCount(ctx context.Context, start, end time.Time) (int64, error)
	GetFailedCount(ctx context.Context, start, end time.Time) (int64, error)
	GetPendingCount(ctx context.Context) (int64, error)

	// 模板管理
	CreateTemplate(ctx context.Context, template *models.NotificationTemplate) error
	GetTemplate(ctx context.Context, id string) (*models.NotificationTemplate, error)
	GetTemplates(ctx context.Context) ([]*models.NotificationTemplate, error)
	UpdateTemplate(ctx context.Context, template *models.NotificationTemplate) error
	DeleteTemplate(ctx context.Context, id string) error
	GetTemplateByName(ctx context.Context, name string) (*models.NotificationTemplate, error)
	GetTemplatesByType(ctx context.Context, notificationType models.NotificationType) ([]*models.NotificationTemplate, error)

	// 批量操作
	BatchCreate(ctx context.Context, notifications []*models.Notification) error
	BatchUpdate(ctx context.Context, notifications []*models.Notification) error
	BatchUpdateStatus(ctx context.Context, ids []string, status models.NotificationStatus) error

	// 清理操作
	CleanupSent(ctx context.Context, before time.Time) (int64, error)
	CleanupFailed(ctx context.Context, before time.Time) (int64, error)
}