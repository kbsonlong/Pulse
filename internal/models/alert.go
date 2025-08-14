package models

import (
	"time"
	"errors"
	"strings"
	"encoding/json"
)

// AlertSeverity 告警严重级别
type AlertSeverity string

const (
	AlertSeverityCritical AlertSeverity = "critical" // 严重
	AlertSeverityHigh     AlertSeverity = "high"     // 高
	AlertSeverityMedium   AlertSeverity = "medium"   // 中
	AlertSeverityLow      AlertSeverity = "low"      // 低
	AlertSeverityInfo     AlertSeverity = "info"     // 信息
)

// AlertStatus 告警状态
type AlertStatus string

const (
	AlertStatusFiring    AlertStatus = "firing"    // 触发中
	AlertStatusResolved  AlertStatus = "resolved"  // 已解决
	AlertStatusSilenced  AlertStatus = "silenced"  // 已静默
	AlertStatusAcked     AlertStatus = "acked"     // 已确认
	AlertStatusSuppressed AlertStatus = "suppressed" // 已抑制
)

// AlertSource 告警来源
type AlertSource string

const (
	AlertSourcePrometheus AlertSource = "prometheus" // Prometheus
	AlertSourceGrafana    AlertSource = "grafana"    // Grafana
	AlertSourceZabbix     AlertSource = "zabbix"     // Zabbix
	AlertSourceCustom     AlertSource = "custom"     // 自定义
	AlertSourceSystem     AlertSource = "system"     // 系统
)

// Alert 告警模型
type Alert struct {
	ID              string                 `json:"id" db:"id"`
	RuleID          *string                `json:"rule_id,omitempty" db:"rule_id"`
	DataSourceID    string                 `json:"data_source_id" db:"data_source_id"`
	Name            string                 `json:"name" db:"name"`
	Description     string                 `json:"description" db:"description"`
	Severity        AlertSeverity          `json:"severity" db:"severity"`
	Status          AlertStatus            `json:"status" db:"status"`
	Source          AlertSource            `json:"source" db:"source"`
	Labels          map[string]string      `json:"labels" db:"labels"`
	Annotations     map[string]string      `json:"annotations" db:"annotations"`
	Value           *float64               `json:"value,omitempty" db:"value"`
	Threshold       *float64               `json:"threshold,omitempty" db:"threshold"`
	Expression      string                 `json:"expression" db:"expression"`
	StartsAt        time.Time              `json:"starts_at" db:"starts_at"`
	EndsAt          *time.Time             `json:"ends_at,omitempty" db:"ends_at"`
	LastEvalAt      time.Time              `json:"last_eval_at" db:"last_eval_at"`
	EvalCount       int64                  `json:"eval_count" db:"eval_count"`
	Fingerprint     string                 `json:"fingerprint" db:"fingerprint"`
	GeneratorURL    *string                `json:"generator_url,omitempty" db:"generator_url"`
	SilenceID       *string                `json:"silence_id,omitempty" db:"silence_id"`
	AckedBy         *string                `json:"acked_by,omitempty" db:"acked_by"`
	AckedAt         *time.Time             `json:"acked_at,omitempty" db:"acked_at"`
	ResolvedBy      *string                `json:"resolved_by,omitempty" db:"resolved_by"`
	ResolvedAt      *time.Time             `json:"resolved_at,omitempty" db:"resolved_at"`
	CreatedAt       time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at" db:"updated_at"`
	DeletedAt       *time.Time             `json:"deleted_at,omitempty" db:"deleted_at"`
}

// AlertCreateRequest 创建告警请求
type AlertCreateRequest struct {
	RuleID       *string           `json:"rule_id,omitempty"`
	DataSourceID string            `json:"data_source_id" binding:"required"`
	Name         string            `json:"name" binding:"required,min=1,max=200"`
	Description  string            `json:"description" binding:"required,min=1,max=1000"`
	Severity     AlertSeverity     `json:"severity" binding:"required"`
	Source       AlertSource       `json:"source" binding:"required"`
	Labels       map[string]string `json:"labels,omitempty"`
	Annotations  map[string]string `json:"annotations,omitempty"`
	Value        *float64          `json:"value,omitempty"`
	Threshold    *float64          `json:"threshold,omitempty"`
	Expression   string            `json:"expression" binding:"required"`
	StartsAt     *time.Time        `json:"starts_at,omitempty"`
	GeneratorURL *string           `json:"generator_url,omitempty"`
}

// AlertUpdateRequest 更新告警请求
type AlertUpdateRequest struct {
	Name        *string            `json:"name,omitempty" binding:"omitempty,min=1,max=200"`
	Description *string            `json:"description,omitempty" binding:"omitempty,min=1,max=1000"`
	Severity    *AlertSeverity     `json:"severity,omitempty"`
	Status      *AlertStatus       `json:"status,omitempty"`
	Labels      *map[string]string `json:"labels,omitempty"`
	Annotations *map[string]string `json:"annotations,omitempty"`
	Value       *float64           `json:"value,omitempty"`
	Threshold   *float64           `json:"threshold,omitempty"`
}

// AlertAckRequest 确认告警请求
type AlertAckRequest struct {
	Comment *string `json:"comment,omitempty" binding:"omitempty,max=500"`
}

// AlertResolveRequest 解决告警请求
type AlertResolveRequest struct {
	Comment *string `json:"comment,omitempty" binding:"omitempty,max=500"`
}

// AlertSilenceRequest 静默告警请求
type AlertSilenceRequest struct {
	Duration time.Duration `json:"duration" binding:"required"`
	Comment  *string       `json:"comment,omitempty" binding:"omitempty,max=500"`
}

// AlertFilter 告警查询过滤器
type AlertFilter struct {
	RuleID       *string        `json:"rule_id,omitempty"`
	DataSourceID *string        `json:"data_source_id,omitempty"`
	Severity     *AlertSeverity `json:"severity,omitempty"`
	Status       *AlertStatus   `json:"status,omitempty"`
	Source       *AlertSource   `json:"source,omitempty"`
	Keyword      *string        `json:"keyword,omitempty"` // 搜索名称、描述
	Labels       map[string]string `json:"labels,omitempty"`
	StartTime    *time.Time     `json:"start_time,omitempty"`
	EndTime      *time.Time     `json:"end_time,omitempty"`
	Page         int            `json:"page" binding:"min=1"`
	PageSize     int            `json:"page_size" binding:"min=1,max=100"`
	SortBy       *string        `json:"sort_by,omitempty"`
	SortOrder    *string        `json:"sort_order,omitempty"` // asc, desc
}

// AlertList 告警列表响应
type AlertList struct {
	Alerts     []*Alert `json:"alerts"`
	Total      int64    `json:"total"`
	Page       int      `json:"page"`
	PageSize   int      `json:"page_size"`
	TotalPages int      `json:"total_pages"`
}

// AlertStats 告警统计
type AlertStats struct {
	Total      int64                    `json:"total"`
	BySeverity map[AlertSeverity]int64  `json:"by_severity"`
	ByStatus   map[AlertStatus]int64    `json:"by_status"`
	BySource   map[AlertSource]int64    `json:"by_source"`
	Trend      []*AlertTrendPoint      `json:"trend"`
}

// AlertTrendPoint 告警趋势点
type AlertTrendPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Count     int64     `json:"count"`
}

// AlertHistory 告警历史记录
type AlertHistory struct {
	ID        string                 `json:"id" db:"id"`
	AlertID   string                 `json:"alert_id" db:"alert_id"`
	Action    string                 `json:"action" db:"action"` // created, updated, acked, resolved, silenced
	OldValue  map[string]interface{} `json:"old_value,omitempty" db:"old_value"`
	NewValue  map[string]interface{} `json:"new_value,omitempty" db:"new_value"`
	UserID    *string                `json:"user_id,omitempty" db:"user_id"`
	Comment   *string                `json:"comment,omitempty" db:"comment"`
	CreatedAt time.Time              `json:"created_at" db:"created_at"`
}

// 验证方法

// Validate 验证告警数据
func (a *Alert) Validate() error {
	if strings.TrimSpace(a.Name) == "" {
		return errors.New("告警名称不能为空")
	}
	
	if len(a.Name) > 200 {
		return errors.New("告警名称长度不能超过200个字符")
	}
	
	if strings.TrimSpace(a.Description) == "" {
		return errors.New("告警描述不能为空")
	}
	
	if len(a.Description) > 1000 {
		return errors.New("告警描述长度不能超过1000个字符")
	}
	
	if !a.Severity.IsValid() {
		return errors.New("无效的告警严重级别")
	}
	
	if !a.Status.IsValid() {
		return errors.New("无效的告警状态")
	}
	
	if !a.Source.IsValid() {
		return errors.New("无效的告警来源")
	}
	
	if strings.TrimSpace(a.DataSourceID) == "" {
		return errors.New("数据源ID不能为空")
	}
	
	if strings.TrimSpace(a.Expression) == "" {
		return errors.New("告警表达式不能为空")
	}
	
	if strings.TrimSpace(a.Fingerprint) == "" {
		return errors.New("告警指纹不能为空")
	}
	
	return nil
}

// IsValid 检查告警严重级别是否有效
func (s AlertSeverity) IsValid() bool {
	switch s {
	case AlertSeverityCritical, AlertSeverityHigh, AlertSeverityMedium, AlertSeverityLow, AlertSeverityInfo:
		return true
	default:
		return false
	}
}

// IsValid 检查告警状态是否有效
func (s AlertStatus) IsValid() bool {
	switch s {
	case AlertStatusFiring, AlertStatusResolved, AlertStatusSilenced, AlertStatusAcked, AlertStatusSuppressed:
		return true
	default:
		return false
	}
}

// IsValid 检查告警来源是否有效
func (s AlertSource) IsValid() bool {
	switch s {
	case AlertSourcePrometheus, AlertSourceGrafana, AlertSourceZabbix, AlertSourceCustom, AlertSourceSystem:
		return true
	default:
		return false
	}
}

// GetSeverityLevel 获取严重级别的数值（用于排序）
func (s AlertSeverity) GetSeverityLevel() int {
	switch s {
	case AlertSeverityCritical:
		return 5
	case AlertSeverityHigh:
		return 4
	case AlertSeverityMedium:
		return 3
	case AlertSeverityLow:
		return 2
	case AlertSeverityInfo:
		return 1
	default:
		return 0
	}
}

// IsFiring 检查告警是否正在触发
func (a *Alert) IsFiring() bool {
	return a.Status == AlertStatusFiring
}

// IsResolved 检查告警是否已解决
func (a *Alert) IsResolved() bool {
	return a.Status == AlertStatusResolved
}

// IsAcked 检查告警是否已确认
func (a *Alert) IsAcked() bool {
	return a.Status == AlertStatusAcked
}

// IsSilenced 检查告警是否已静默
func (a *Alert) IsSilenced() bool {
	return a.Status == AlertStatusSilenced
}

// GetDuration 获取告警持续时间
func (a *Alert) GetDuration() time.Duration {
	if a.EndsAt != nil {
		return a.EndsAt.Sub(a.StartsAt)
	}
	return time.Since(a.StartsAt)
}

// Validate 验证创建告警请求
func (req *AlertCreateRequest) Validate() error {
	if strings.TrimSpace(req.Name) == "" {
		return errors.New("告警名称不能为空")
	}
	
	if len(req.Name) > 200 {
		return errors.New("告警名称长度不能超过200个字符")
	}
	
	if strings.TrimSpace(req.Description) == "" {
		return errors.New("告警描述不能为空")
	}
	
	if len(req.Description) > 1000 {
		return errors.New("告警描述长度不能超过1000个字符")
	}
	
	if !req.Severity.IsValid() {
		return errors.New("无效的告警严重级别")
	}
	
	if !req.Source.IsValid() {
		return errors.New("无效的告警来源")
	}
	
	if strings.TrimSpace(req.DataSourceID) == "" {
		return errors.New("数据源ID不能为空")
	}
	
	if strings.TrimSpace(req.Expression) == "" {
		return errors.New("告警表达式不能为空")
	}
	
	return nil
}

// MarshalLabels 序列化标签为JSON
func (a *Alert) MarshalLabels() ([]byte, error) {
	if a.Labels == nil {
		return json.Marshal(map[string]string{})
	}
	return json.Marshal(a.Labels)
}

// UnmarshalLabels 反序列化标签从JSON
func (a *Alert) UnmarshalLabels(data []byte) error {
	if a.Labels == nil {
		a.Labels = make(map[string]string)
	}
	return json.Unmarshal(data, &a.Labels)
}

// MarshalAnnotations 序列化注解为JSON
func (a *Alert) MarshalAnnotations() ([]byte, error) {
	if a.Annotations == nil {
		return json.Marshal(map[string]string{})
	}
	return json.Marshal(a.Annotations)
}

// UnmarshalAnnotations 反序列化注解从JSON
func (a *Alert) UnmarshalAnnotations(data []byte) error {
	if a.Annotations == nil {
		a.Annotations = make(map[string]string)
	}
	return json.Unmarshal(data, &a.Annotations)
}