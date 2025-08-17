package models

import (
	"time"

	"github.com/google/uuid"
)

// NotificationStatus 通知状态
type NotificationStatus string

const (
	NotificationStatusPending NotificationStatus = "pending"
	NotificationStatusSent    NotificationStatus = "sent"
	NotificationStatusFailed  NotificationStatus = "failed"
	NotificationStatusRetry   NotificationStatus = "retry"
)

// NotificationType 通知类型
type NotificationType string

const (
	NotificationTypeEmail    NotificationType = "email"
	NotificationTypeSMS      NotificationType = "sms"
	NotificationTypeDingTalk NotificationType = "dingtalk"
	NotificationTypeWeChat   NotificationType = "wechat"
	NotificationTypeSlack    NotificationType = "slack"
	NotificationTypeWebhook  NotificationType = "webhook"
)

// Notification 通知记录
type Notification struct {
	ID          uuid.UUID          `json:"id" db:"id"`
	AlertID     uuid.UUID          `json:"alert_id" db:"alert_id"`
	Type        NotificationType   `json:"type" db:"type"`
	Recipient   string             `json:"recipient" db:"recipient"`
	Subject     string             `json:"subject" db:"subject"`
	Content     string             `json:"content" db:"content"`
	Status      NotificationStatus `json:"status" db:"status"`
	RetryCount  int                `json:"retry_count" db:"retry_count"`
	MaxRetries  int                `json:"max_retries" db:"max_retries"`
	LastError   *string            `json:"last_error,omitempty" db:"last_error"`
	SentAt      *time.Time         `json:"sent_at,omitempty" db:"sent_at"`
	CreatedAt   time.Time          `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at" db:"updated_at"`
}

// NotificationTemplate 通知模板
type NotificationTemplate struct {
	ID          uuid.UUID        `json:"id" db:"id"`
	Name        string           `json:"name" db:"name"`
	Type        NotificationType `json:"type" db:"type"`
	Subject     string           `json:"subject" db:"subject"`
	Content     string           `json:"content" db:"content"`
	Variables   []string         `json:"variables" db:"variables"`
	IsDefault   bool             `json:"is_default" db:"is_default"`
	CreatedBy   uuid.UUID        `json:"created_by" db:"created_by"`
	CreatedAt   time.Time        `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at" db:"updated_at"`
}

// NotificationFilter 通知查询过滤器
type NotificationFilter struct {
	AlertID    *uuid.UUID          `json:"alert_id,omitempty"`
	Type       *NotificationType   `json:"type,omitempty"`
	Status     *NotificationStatus `json:"status,omitempty"`
	Recipient  *string             `json:"recipient,omitempty"`
	StartTime  *time.Time          `json:"start_time,omitempty"`
	EndTime    *time.Time          `json:"end_time,omitempty"`
	Page       int                 `json:"page"`
	PageSize   int                 `json:"page_size"`
	SortBy     string              `json:"sort_by"`
	SortOrder  string              `json:"sort_order"`
}

// NotificationList 通知列表
type NotificationList struct {
	Items      []*Notification `json:"items"`
	Total      int64           `json:"total"`
	Page       int             `json:"page"`
	PageSize   int             `json:"page_size"`
	TotalPages int             `json:"total_pages"`
}

// NotificationStats 通知统计
type NotificationStats struct {
	Total      int64 `json:"total"`
	Pending    int64 `json:"pending"`
	Sent       int64 `json:"sent"`
	Failed     int64 `json:"failed"`
	Retry      int64 `json:"retry"`
	SuccessRate float64 `json:"success_rate"`
}