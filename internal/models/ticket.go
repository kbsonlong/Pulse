package models

import (
	"time"
	"errors"
	"strings"
	"encoding/json"
)

// TicketType 工单类型
type TicketType string

const (
	TicketTypeIncident    TicketType = "incident"    // 事件
	TicketTypeProblem     TicketType = "problem"     // 问题
	TicketTypeChange      TicketType = "change"      // 变更
	TicketTypeRequest     TicketType = "request"     // 请求
	TicketTypeMaintenance TicketType = "maintenance" // 维护
	TicketTypeAlert       TicketType = "alert"       // 告警
)

// TicketStatus 工单状态
type TicketStatus string

const (
	TicketStatusOpen       TicketStatus = "open"       // 打开
	TicketStatusAssigned   TicketStatus = "assigned"   // 已分配
	TicketStatusInProgress TicketStatus = "in_progress" // 处理中
	TicketStatusPending    TicketStatus = "pending"    // 等待中
	TicketStatusResolved   TicketStatus = "resolved"   // 已解决
	TicketStatusClosed     TicketStatus = "closed"     // 已关闭
	TicketStatusCancelled  TicketStatus = "cancelled"  // 已取消
)

// TicketPriority 工单优先级
type TicketPriority string

const (
	TicketPriorityLow      TicketPriority = "low"      // 低
	TicketPriorityMedium   TicketPriority = "medium"   // 中
	TicketPriorityHigh     TicketPriority = "high"     // 高
	TicketPriorityCritical TicketPriority = "critical" // 紧急
	TicketPriorityUrgent   TicketPriority = "urgent"   // 非常紧急
)

// TicketSeverity 工单严重程度
type TicketSeverity string

const (
	TicketSeverityInfo     TicketSeverity = "info"     // 信息
	TicketSeverityWarning  TicketSeverity = "warning"  // 警告
	TicketSeverityMinor    TicketSeverity = "minor"    // 轻微
	TicketSeverityMajor    TicketSeverity = "major"    // 重要
	TicketSeverityCritical TicketSeverity = "critical" // 严重
)

// TicketSource 工单来源
type TicketSource string

const (
	TicketSourceManual    TicketSource = "manual"    // 手动创建
	TicketSourceAlert     TicketSource = "alert"     // 告警触发
	TicketSourceAPI       TicketSource = "api"       // API创建
	TicketSourceEmail     TicketSource = "email"     // 邮件
	TicketSourceWebhook   TicketSource = "webhook"   // Webhook
	TicketSourceScheduled TicketSource = "scheduled" // 定时任务
)

// TicketComment 工单评论
type TicketComment struct {
	ID        string    `json:"id" db:"id"`
	TicketID  string    `json:"ticket_id" db:"ticket_id"`
	UserID    string    `json:"user_id" db:"user_id"`
	UserName  string    `json:"user_name" db:"user_name"`
	Content   string    `json:"content" db:"content"`
	IsPrivate bool      `json:"is_private" db:"is_private"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// TicketAttachment 工单附件
type TicketAttachment struct {
	ID        string    `json:"id" db:"id"`
	TicketID  string    `json:"ticket_id" db:"ticket_id"`
	FileName  string    `json:"file_name" db:"file_name"`
	FileSize  int64     `json:"file_size" db:"file_size"`
	FileType  string    `json:"file_type" db:"file_type"`
	FilePath  string    `json:"file_path" db:"file_path"`
	UploadBy  string    `json:"upload_by" db:"upload_by"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// TicketHistory 工单历史记录
type TicketHistory struct {
	ID        string                 `json:"id" db:"id"`
	TicketID  string                 `json:"ticket_id" db:"ticket_id"`
	UserID    string                 `json:"user_id" db:"user_id"`
	UserName  string                 `json:"user_name" db:"user_name"`
	Action    string                 `json:"action" db:"action"`
	OldValue  *string                `json:"old_value,omitempty" db:"old_value"`
	NewValue  *string                `json:"new_value,omitempty" db:"new_value"`
	Changes   map[string]interface{} `json:"changes,omitempty" db:"changes"`
	Comment   *string                `json:"comment,omitempty" db:"comment"`
	CreatedAt time.Time              `json:"created_at" db:"created_at"`
}

// TicketSLA SLA配置
type TicketSLA struct {
	ResponseTime   *time.Duration `json:"response_time,omitempty"`   // 响应时间
	ResolutionTime *time.Duration `json:"resolution_time,omitempty"` // 解决时间
	EscalationTime *time.Duration `json:"escalation_time,omitempty"` // 升级时间
}

// Ticket 工单模型
type Ticket struct {
	ID              string            `json:"id" db:"id"`
	Number          string            `json:"number" db:"number"`
	Title           string            `json:"title" db:"title"`
	Description     string            `json:"description" db:"description"`
	Type            TicketType        `json:"type" db:"type"`
	Status          TicketStatus      `json:"status" db:"status"`
	Priority        TicketPriority    `json:"priority" db:"priority"`
	Severity        TicketSeverity    `json:"severity" db:"severity"`
	Source          TicketSource      `json:"source" db:"source"`
	Category        *string           `json:"category,omitempty" db:"category"`
	Subcategory     *string           `json:"subcategory,omitempty" db:"subcategory"`
	Tags            []string          `json:"tags" db:"tags"`
	Labels          map[string]string `json:"labels" db:"labels"`
	AlertID         *string           `json:"alert_id,omitempty" db:"alert_id"`
	RuleID          *string           `json:"rule_id,omitempty" db:"rule_id"`
	DataSourceID    *string           `json:"data_source_id,omitempty" db:"data_source_id"`
	ReporterID      string            `json:"reporter_id" db:"reporter_id"`
	ReporterName    string            `json:"reporter_name" db:"reporter_name"`
	AssigneeID      *string           `json:"assignee_id,omitempty" db:"assignee_id"`
	AssigneeName    *string           `json:"assignee_name,omitempty" db:"assignee_name"`
	TeamID          *string           `json:"team_id,omitempty" db:"team_id"`
	TeamName        *string           `json:"team_name,omitempty" db:"team_name"`
	SLA             *TicketSLA        `json:"sla,omitempty" db:"sla"`
	DueDate         *time.Time        `json:"due_date,omitempty" db:"due_date"`
	ResponseTime    *time.Time        `json:"response_time,omitempty" db:"response_time"`
	ResolutionTime  *time.Time        `json:"resolution_time,omitempty" db:"resolution_time"`
	FirstResponseAt *time.Time        `json:"first_response_at,omitempty" db:"first_response_at"`
	ResolvedAt      *time.Time        `json:"resolved_at,omitempty" db:"resolved_at"`
	ClosedAt        *time.Time        `json:"closed_at,omitempty" db:"closed_at"`
	ReopenedAt      *time.Time        `json:"reopened_at,omitempty" db:"reopened_at"`
	ReopenCount     int               `json:"reopen_count" db:"reopen_count"`
	CommentCount    int               `json:"comment_count" db:"comment_count"`
	AttachmentCount int               `json:"attachment_count" db:"attachment_count"`
	WorkTime        *time.Duration    `json:"work_time,omitempty" db:"work_time"`
	EstimatedTime   *time.Duration    `json:"estimated_time,omitempty" db:"estimated_time"`
	ActualTime      *time.Duration    `json:"actual_time,omitempty" db:"actual_time"`
	Resolution      *string           `json:"resolution,omitempty" db:"resolution"`
	RootCause       *string           `json:"root_cause,omitempty" db:"root_cause"`
	Workaround      *string           `json:"workaround,omitempty" db:"workaround"`
	Impact          *string           `json:"impact,omitempty" db:"impact"`
	Urgency         *string           `json:"urgency,omitempty" db:"urgency"`
	BusinessImpact  *string           `json:"business_impact,omitempty" db:"business_impact"`
	CustomFields    map[string]interface{} `json:"custom_fields,omitempty" db:"custom_fields"`
	CreatedAt       time.Time         `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time         `json:"updated_at" db:"updated_at"`
	DeletedAt       *time.Time        `json:"deleted_at,omitempty" db:"deleted_at"`
}

// TicketCreateRequest 创建工单请求
type TicketCreateRequest struct {
	Title          string            `json:"title" binding:"required,min=1,max=200"`
	Description    string            `json:"description" binding:"required,min=1,max=5000"`
	Type           TicketType        `json:"type" binding:"required"`
	Priority       TicketPriority    `json:"priority" binding:"required"`
	Severity       TicketSeverity    `json:"severity" binding:"required"`
	Category       *string           `json:"category,omitempty"`
	Subcategory    *string           `json:"subcategory,omitempty"`
	Tags           []string          `json:"tags,omitempty"`
	Labels         map[string]string `json:"labels,omitempty"`
	AlertID        *string           `json:"alert_id,omitempty"`
	RuleID         *string           `json:"rule_id,omitempty"`
	DataSourceID   *string           `json:"data_source_id,omitempty"`
	AssigneeID     *string           `json:"assignee_id,omitempty"`
	TeamID         *string           `json:"team_id,omitempty"`
	DueDate        *time.Time        `json:"due_date,omitempty"`
	EstimatedTime  *time.Duration    `json:"estimated_time,omitempty"`
	Impact         *string           `json:"impact,omitempty"`
	Urgency        *string           `json:"urgency,omitempty"`
	BusinessImpact *string           `json:"business_impact,omitempty"`
	CustomFields   map[string]interface{} `json:"custom_fields,omitempty"`
}

// TicketUpdateRequest 更新工单请求
type TicketUpdateRequest struct {
	Title          *string            `json:"title,omitempty" binding:"omitempty,min=1,max=200"`
	Description    *string            `json:"description,omitempty" binding:"omitempty,min=1,max=5000"`
	Status         *TicketStatus      `json:"status,omitempty"`
	Priority       *TicketPriority    `json:"priority,omitempty"`
	Severity       *TicketSeverity    `json:"severity,omitempty"`
	Category       *string            `json:"category,omitempty"`
	Subcategory    *string            `json:"subcategory,omitempty"`
	Tags           *[]string          `json:"tags,omitempty"`
	Labels         *map[string]string `json:"labels,omitempty"`
	AssigneeID     *string            `json:"assignee_id,omitempty"`
	TeamID         *string            `json:"team_id,omitempty"`
	DueDate        *time.Time         `json:"due_date,omitempty"`
	EstimatedTime  *time.Duration     `json:"estimated_time,omitempty"`
	Resolution     *string            `json:"resolution,omitempty"`
	RootCause      *string            `json:"root_cause,omitempty"`
	Workaround     *string            `json:"workaround,omitempty"`
	Impact         *string            `json:"impact,omitempty"`
	Urgency        *string            `json:"urgency,omitempty"`
	BusinessImpact *string            `json:"business_impact,omitempty"`
	CustomFields   *map[string]interface{} `json:"custom_fields,omitempty"`
}

// TicketAssignRequest 分配工单请求
type TicketAssignRequest struct {
	AssigneeID *string `json:"assignee_id,omitempty"`
	TeamID     *string `json:"team_id,omitempty"`
	Comment    *string `json:"comment,omitempty"`
}

// TicketCommentRequest 添加评论请求
type TicketCommentRequest struct {
	Content   string `json:"content" binding:"required,min=1,max=2000"`
	IsPrivate bool   `json:"is_private"`
}

// TicketFilter 工单查询过滤器
type TicketFilter struct {
	Type           *TicketType     `json:"type,omitempty"`
	Status         *TicketStatus   `json:"status,omitempty"`
	Priority       *TicketPriority `json:"priority,omitempty"`
	Severity       *TicketSeverity `json:"severity,omitempty"`
	Source         *TicketSource   `json:"source,omitempty"`
	Category       *string         `json:"category,omitempty"`
	Subcategory    *string         `json:"subcategory,omitempty"`
	Keyword        *string         `json:"keyword,omitempty"` // 搜索标题、描述
	Tags           []string        `json:"tags,omitempty"`
	ReporterID     *string         `json:"reporter_id,omitempty"`
	AssigneeID     *string         `json:"assignee_id,omitempty"`
	TeamID         *string         `json:"team_id,omitempty"`
	AlertID        *string         `json:"alert_id,omitempty"`
	RuleID         *string         `json:"rule_id,omitempty"`
	DataSourceID   *string         `json:"data_source_id,omitempty"`
	CreatedStart   *time.Time      `json:"created_start,omitempty"`
	CreatedEnd     *time.Time      `json:"created_end,omitempty"`
	DueDateStart   *time.Time      `json:"due_date_start,omitempty"`
	DueDateEnd     *time.Time      `json:"due_date_end,omitempty"`
	Overdue        *bool           `json:"overdue,omitempty"`
	Page           int             `json:"page" binding:"min=1"`
	PageSize       int             `json:"page_size" binding:"min=1,max=100"`
	SortBy         *string         `json:"sort_by,omitempty"`
	SortOrder      *string         `json:"sort_order,omitempty"` // asc, desc
}

// TicketList 工单列表响应
type TicketList struct {
	Tickets    []*Ticket `json:"tickets"`
	Total      int64     `json:"total"`
	Page       int       `json:"page"`
	PageSize   int       `json:"page_size"`
	TotalPages int       `json:"total_pages"`
}

// TicketStats 工单统计
type TicketStats struct {
	Total         int64                     `json:"total"`
	ByType        map[TicketType]int64      `json:"by_type"`
	ByStatus      map[TicketStatus]int64    `json:"by_status"`
	ByPriority    map[TicketPriority]int64  `json:"by_priority"`
	BySeverity    map[TicketSeverity]int64  `json:"by_severity"`
	BySource      map[TicketSource]int64    `json:"by_source"`
	OpenCount     int64                     `json:"open_count"`
	ResolvedCount int64                     `json:"resolved_count"`
	OverdueCount  int64                     `json:"overdue_count"`
	AvgResolutionTime time.Duration         `json:"avg_resolution_time"`
	AvgResponseTime   time.Duration         `json:"avg_response_time"`
	SLACompliance     float64               `json:"sla_compliance"`
}

// TicketTrendPoint 工单趋势数据点
type TicketTrendPoint struct {
	Time     time.Time `json:"time"`
	Created  int64     `json:"created"`
	Resolved int64     `json:"resolved"`
	Closed   int64     `json:"closed"`
}

// 验证方法

// Validate 验证工单数据
func (t *Ticket) Validate() error {
	if strings.TrimSpace(t.Title) == "" {
		return errors.New("工单标题不能为空")
	}
	
	if len(t.Title) > 200 {
		return errors.New("工单标题长度不能超过200个字符")
	}
	
	if strings.TrimSpace(t.Description) == "" {
		return errors.New("工单描述不能为空")
	}
	
	if len(t.Description) > 5000 {
		return errors.New("工单描述长度不能超过5000个字符")
	}
	
	if !t.Type.IsValid() {
		return errors.New("无效的工单类型")
	}
	
	if !t.Status.IsValid() {
		return errors.New("无效的工单状态")
	}
	
	if !t.Priority.IsValid() {
		return errors.New("无效的工单优先级")
	}
	
	if !t.Severity.IsValid() {
		return errors.New("无效的工单严重程度")
	}
	
	if !t.Source.IsValid() {
		return errors.New("无效的工单来源")
	}
	
	if strings.TrimSpace(t.ReporterID) == "" {
		return errors.New("报告人不能为空")
	}
	
	return nil
}

// IsValid 检查工单类型是否有效
func (t TicketType) IsValid() bool {
	switch t {
	case TicketTypeIncident, TicketTypeProblem, TicketTypeChange,
		 TicketTypeRequest, TicketTypeMaintenance, TicketTypeAlert:
		return true
	default:
		return false
	}
}

// IsValid 检查工单状态是否有效
func (s TicketStatus) IsValid() bool {
	switch s {
	case TicketStatusOpen, TicketStatusAssigned, TicketStatusInProgress,
		 TicketStatusPending, TicketStatusResolved, TicketStatusClosed,
		 TicketStatusCancelled:
		return true
	default:
		return false
	}
}

// IsValid 检查工单优先级是否有效
func (p TicketPriority) IsValid() bool {
	switch p {
	case TicketPriorityLow, TicketPriorityMedium, TicketPriorityHigh,
		 TicketPriorityCritical, TicketPriorityUrgent:
		return true
	default:
		return false
	}
}

// IsValid 检查工单严重程度是否有效
func (s TicketSeverity) IsValid() bool {
	switch s {
	case TicketSeverityInfo, TicketSeverityWarning, TicketSeverityMinor,
		 TicketSeverityMajor, TicketSeverityCritical:
		return true
	default:
		return false
	}
}

// IsValid 检查工单来源是否有效
func (s TicketSource) IsValid() bool {
	switch s {
	case TicketSourceManual, TicketSourceAlert, TicketSourceAPI,
		 TicketSourceEmail, TicketSourceWebhook, TicketSourceScheduled:
		return true
	default:
		return false
	}
}

// Validate 验证创建工单请求
func (req *TicketCreateRequest) Validate() error {
	if strings.TrimSpace(req.Title) == "" {
		return errors.New("工单标题不能为空")
	}
	
	if len(req.Title) > 200 {
		return errors.New("工单标题长度不能超过200个字符")
	}
	
	if strings.TrimSpace(req.Description) == "" {
		return errors.New("工单描述不能为空")
	}
	
	if len(req.Description) > 5000 {
		return errors.New("工单描述长度不能超过5000个字符")
	}
	
	if !req.Type.IsValid() {
		return errors.New("无效的工单类型")
	}
	
	if !req.Priority.IsValid() {
		return errors.New("无效的工单优先级")
	}
	
	if !req.Severity.IsValid() {
		return errors.New("无效的工单严重程度")
	}
	
	return nil
}

// Validate 验证工单评论请求
func (req *TicketCommentRequest) Validate() error {
	if strings.TrimSpace(req.Content) == "" {
		return errors.New("评论内容不能为空")
	}
	
	if len(req.Content) > 2000 {
		return errors.New("评论内容长度不能超过2000个字符")
	}
	
	return nil
}

// 辅助方法

// IsOpen 检查工单是否打开
func (t *Ticket) IsOpen() bool {
	return t.Status == TicketStatusOpen || t.Status == TicketStatusAssigned || t.Status == TicketStatusInProgress
}

// IsResolved 检查工单是否已解决
func (t *Ticket) IsResolved() bool {
	return t.Status == TicketStatusResolved
}

// IsClosed 检查工单是否已关闭
func (t *Ticket) IsClosed() bool {
	return t.Status == TicketStatusClosed
}

// IsOverdue 检查工单是否逾期
func (t *Ticket) IsOverdue() bool {
	if t.DueDate == nil {
		return false
	}
	return time.Now().After(*t.DueDate) && !t.IsResolved() && !t.IsClosed()
}

// GetAge 获取工单年龄（创建至今的时间）
func (t *Ticket) GetAge() time.Duration {
	return time.Since(t.CreatedAt)
}

// GetResolutionTime 获取解决时间
func (t *Ticket) GetResolutionTime() *time.Duration {
	if t.ResolvedAt == nil {
		return nil
	}
	duration := t.ResolvedAt.Sub(t.CreatedAt)
	return &duration
}

// GetResponseTime 获取响应时间
func (t *Ticket) GetResponseTime() *time.Duration {
	if t.FirstResponseAt == nil {
		return nil
	}
	duration := t.FirstResponseAt.Sub(t.CreatedAt)
	return &duration
}

// GetPriorityLevel 获取优先级数值
func (p TicketPriority) GetPriorityLevel() int {
	switch p {
	case TicketPriorityLow:
		return 1
	case TicketPriorityMedium:
		return 2
	case TicketPriorityHigh:
		return 3
	case TicketPriorityCritical:
		return 4
	case TicketPriorityUrgent:
		return 5
	default:
		return 0
	}
}

// GetSeverityLevel 获取严重程度数值
func (s TicketSeverity) GetSeverityLevel() int {
	switch s {
	case TicketSeverityInfo:
		return 1
	case TicketSeverityWarning:
		return 2
	case TicketSeverityMinor:
		return 3
	case TicketSeverityMajor:
		return 4
	case TicketSeverityCritical:
		return 5
	default:
		return 0
	}
}

// MarshalTags 序列化标签为JSON
func (t *Ticket) MarshalTags() ([]byte, error) {
	if t.Tags == nil {
		return json.Marshal([]string{})
	}
	return json.Marshal(t.Tags)
}

// UnmarshalTags 反序列化标签从JSON
func (t *Ticket) UnmarshalTags(data []byte) error {
	return json.Unmarshal(data, &t.Tags)
}

// MarshalLabels 序列化标签为JSON
func (t *Ticket) MarshalLabels() ([]byte, error) {
	if t.Labels == nil {
		return json.Marshal(map[string]string{})
	}
	return json.Marshal(t.Labels)
}

// UnmarshalLabels 反序列化标签从JSON
func (t *Ticket) UnmarshalLabels(data []byte) error {
	return json.Unmarshal(data, &t.Labels)
}

// MarshalSLA 序列化SLA为JSON
func (t *Ticket) MarshalSLA() ([]byte, error) {
	if t.SLA == nil {
		return json.Marshal(nil)
	}
	return json.Marshal(t.SLA)
}

// UnmarshalSLA 反序列化SLA从JSON
func (t *Ticket) UnmarshalSLA(data []byte) error {
	if len(data) == 0 || string(data) == "null" {
		t.SLA = nil
		return nil
	}
	if t.SLA == nil {
		t.SLA = &TicketSLA{}
	}
	return json.Unmarshal(data, t.SLA)
}

// MarshalCustomFields 序列化自定义字段为JSON
func (t *Ticket) MarshalCustomFields() ([]byte, error) {
	if t.CustomFields == nil {
		return json.Marshal(map[string]interface{}{})
	}
	return json.Marshal(t.CustomFields)
}

// UnmarshalCustomFields 反序列化自定义字段从JSON
func (t *Ticket) UnmarshalCustomFields(data []byte) error {
	return json.Unmarshal(data, &t.CustomFields)
}

// MarshalChanges 序列化变更为JSON
func (h *TicketHistory) MarshalChanges() ([]byte, error) {
	if h.Changes == nil {
		return json.Marshal(map[string]interface{}{})
	}
	return json.Marshal(h.Changes)
}

// UnmarshalChanges 反序列化变更从JSON
func (h *TicketHistory) UnmarshalChanges(data []byte) error {
	return json.Unmarshal(data, &h.Changes)
}

// GetDisplayName 获取显示名称
func (t TicketType) GetDisplayName() string {
	switch t {
	case TicketTypeIncident:
		return "事件"
	case TicketTypeProblem:
		return "问题"
	case TicketTypeChange:
		return "变更"
	case TicketTypeRequest:
		return "请求"
	case TicketTypeMaintenance:
		return "维护"
	case TicketTypeAlert:
		return "告警"
	default:
		return string(t)
	}
}

// GetDisplayName 获取显示名称
func (s TicketStatus) GetDisplayName() string {
	switch s {
	case TicketStatusOpen:
		return "打开"
	case TicketStatusAssigned:
		return "已分配"
	case TicketStatusInProgress:
		return "处理中"
	case TicketStatusPending:
		return "等待中"
	case TicketStatusResolved:
		return "已解决"
	case TicketStatusClosed:
		return "已关闭"
	case TicketStatusCancelled:
		return "已取消"
	default:
		return string(s)
	}
}

// GetDisplayName 获取显示名称
func (p TicketPriority) GetDisplayName() string {
	switch p {
	case TicketPriorityLow:
		return "低"
	case TicketPriorityMedium:
		return "中"
	case TicketPriorityHigh:
		return "高"
	case TicketPriorityCritical:
		return "紧急"
	case TicketPriorityUrgent:
		return "非常紧急"
	default:
		return string(p)
	}
}

// GetDisplayName 获取显示名称
func (s TicketSeverity) GetDisplayName() string {
	switch s {
	case TicketSeverityInfo:
		return "信息"
	case TicketSeverityWarning:
		return "警告"
	case TicketSeverityMinor:
		return "轻微"
	case TicketSeverityMajor:
		return "重要"
	case TicketSeverityCritical:
		return "严重"
	default:
		return string(s)
	}
}