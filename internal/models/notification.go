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