package models

import (
	"time"

	"github.com/google/uuid"
)

// WebhookStatus Webhook状态
type WebhookStatus string

const (
	WebhookStatusActive   WebhookStatus = "active"
	WebhookStatusInactive WebhookStatus = "inactive"
	WebhookStatusDisabled WebhookStatus = "disabled"
)

// WebhookEvent Webhook事件类型
type WebhookEvent string

const (
	WebhookEventAlertCreated   WebhookEvent = "alert.created"
	WebhookEventAlertUpdated   WebhookEvent = "alert.updated"
	WebhookEventAlertResolved  WebhookEvent = "alert.resolved"
	WebhookEventAlertAcknowledged WebhookEvent = "alert.acknowledged"
	WebhookEventRuleCreated    WebhookEvent = "rule.created"
	WebhookEventRuleUpdated    WebhookEvent = "rule.updated"
	WebhookEventRuleDeleted    WebhookEvent = "rule.deleted"
)

// Webhook Webhook配置
type Webhook struct {
	ID          uuid.UUID     `json:"id" db:"id"`
	Name        string        `json:"name" db:"name"`
	URL         string        `json:"url" db:"url"`
	Method      string        `json:"method" db:"method"`
	Secret      *string       `json:"secret,omitempty" db:"secret"`
	Events      []WebhookEvent `json:"events" db:"events"`
	Headers     map[string]string `json:"headers" db:"headers"`
	Timeout     int           `json:"timeout" db:"timeout"`
	RetryCount  int           `json:"retry_count" db:"retry_count"`
	RetryDelay  int           `json:"retry_delay" db:"retry_delay"`
	Status      WebhookStatus `json:"status" db:"status"`
	LastTriggered *time.Time  `json:"last_triggered,omitempty" db:"last_triggered"`
	CreatedBy   uuid.UUID     `json:"created_by" db:"created_by"`
	CreatedAt   time.Time     `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at" db:"updated_at"`
}

// WebhookFilter Webhook查询过滤器
type WebhookFilter struct {
	Name      *string        `json:"name,omitempty"`
	Status    *WebhookStatus `json:"status,omitempty"`
	CreatedBy *uuid.UUID     `json:"created_by,omitempty"`
	Page      int            `json:"page"`
	PageSize  int            `json:"page_size"`
}

// WebhookRuleFilter Webhook规则过滤器
type WebhookRuleFilter struct {
	ID        uuid.UUID `json:"id" db:"id"`
	WebhookID uuid.UUID `json:"webhook_id" db:"webhook_id"`
	Field     string    `json:"field" db:"field"`
	Operator  string    `json:"operator" db:"operator"`
	Value     string    `json:"value" db:"value"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// WebhookLog Webhook执行日志
type WebhookLog struct {
	ID         uuid.UUID `json:"id" db:"id"`
	WebhookID  uuid.UUID `json:"webhook_id" db:"webhook_id"`
	Event      WebhookEvent `json:"event" db:"event"`
	Payload    string    `json:"payload" db:"payload"`
	StatusCode int       `json:"status_code" db:"status_code"`
	Response   *string   `json:"response,omitempty" db:"response"`
	Error      *string   `json:"error,omitempty" db:"error"`
	Duration   int64     `json:"duration" db:"duration"` // 毫秒
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}

// WebhookList Webhook列表
type WebhookList struct {
	Webhooks []*Webhook `json:"webhooks"`
	Total    int64      `json:"total"`
	Page     int        `json:"page"`
	PageSize int        `json:"page_size"`
}

// WebhookLogFilter Webhook日志过滤器
type WebhookLogFilter struct {
	WebhookID  *uuid.UUID    `json:"webhook_id,omitempty"`
	Event      *WebhookEvent `json:"event,omitempty"`
	StatusCode *int          `json:"status_code,omitempty"`
	StartTime  *time.Time    `json:"start_time,omitempty"`
	EndTime    *time.Time    `json:"end_time,omitempty"`
	Page       int           `json:"page"`
	PageSize   int           `json:"page_size"`
}

// WebhookLogList Webhook日志列表
type WebhookLogList struct {
	Logs     []*WebhookLog `json:"logs"`
	Total    int64         `json:"total"`
	Page     int           `json:"page"`
	PageSize int           `json:"page_size"`
}

// WebhookStats Webhook统计信息
type WebhookStats struct {
	TotalRequests int64   `json:"total_requests"`
	SuccessCount  int64   `json:"success_count"`
	FailureCount  int64   `json:"failure_count"`
	SuccessRate   float64 `json:"success_rate"`
	AvgDuration   float64 `json:"avg_duration"`
}