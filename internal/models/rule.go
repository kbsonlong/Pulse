package models

import (
	"time"
	"errors"
	"strings"
	"encoding/json"
)

// RuleType 规则类型
type RuleType string

const (
	RuleTypeMetric    RuleType = "metric"    // 指标规则
	RuleTypeLog       RuleType = "log"       // 日志规则
	RuleTypeComposite RuleType = "composite" // 复合规则
	RuleTypeAnomaly   RuleType = "anomaly"   // 异常检测规则
)

// RuleStatus 规则状态
type RuleStatus string

const (
	RuleStatusActive   RuleStatus = "active"   // 激活
	RuleStatusInactive RuleStatus = "inactive" // 未激活
	RuleStatusDisabled RuleStatus = "disabled" // 禁用
	RuleStatusTesting  RuleStatus = "testing"  // 测试中
)

// RuleOperator 规则操作符
type RuleOperator string

const (
	RuleOperatorGT  RuleOperator = "gt"  // 大于
	RuleOperatorGTE RuleOperator = "gte" // 大于等于
	RuleOperatorLT  RuleOperator = "lt"  // 小于
	RuleOperatorLTE RuleOperator = "lte" // 小于等于
	RuleOperatorEQ  RuleOperator = "eq"  // 等于
	RuleOperatorNE  RuleOperator = "ne"  // 不等于
	RuleOperatorIN  RuleOperator = "in"  // 包含
	RuleOperatorNIN RuleOperator = "nin" // 不包含
)

// RuleCondition 规则条件
type RuleCondition struct {
	Field    string        `json:"field"`    // 字段名
	Operator RuleOperator `json:"operator"` // 操作符
	Value    interface{}   `json:"value"`    // 值
	Logic    *string       `json:"logic,omitempty"` // 逻辑操作符 (AND, OR)
}

// RuleAction 规则动作
type RuleAction struct {
	Type       string                 `json:"type"`       // 动作类型: alert, webhook, email, sms
	Target     string                 `json:"target"`     // 目标地址
	Template   *string                `json:"template,omitempty"`   // 模板
	Parameters map[string]interface{} `json:"parameters,omitempty"` // 参数
}

// Rule 规则模型
type Rule struct {
	ID              string           `json:"id" db:"id"`
	DataSourceID    string           `json:"data_source_id" db:"data_source_id"`
	Name            string           `json:"name" db:"name"`
	Description     string           `json:"description" db:"description"`
	Type            RuleType         `json:"type" db:"type"`
	Status          RuleStatus       `json:"status" db:"status"`
	Enabled         bool             `json:"enabled" db:"enabled"`
	Severity        AlertSeverity    `json:"severity" db:"severity"`
	Expression      string           `json:"expression" db:"expression"`
	Conditions      []RuleCondition  `json:"conditions" db:"conditions"`
	Actions         []RuleAction     `json:"actions" db:"actions"`
	Labels          map[string]string `json:"labels" db:"labels"`
	Annotations     map[string]string `json:"annotations" db:"annotations"`
	EvaluationInterval time.Duration `json:"evaluation_interval" db:"evaluation_interval"`
	ForDuration     time.Duration    `json:"for_duration" db:"for_duration"`
	KeepFiringFor   time.Duration    `json:"keep_firing_for" db:"keep_firing_for"`
	Threshold       *float64         `json:"threshold,omitempty" db:"threshold"`
	RecoveryThreshold *float64       `json:"recovery_threshold,omitempty" db:"recovery_threshold"`
	NoDataState     *string          `json:"no_data_state,omitempty" db:"no_data_state"`
	ExecErrState    *string          `json:"exec_err_state,omitempty" db:"exec_err_state"`
	LastEvalAt      *time.Time       `json:"last_eval_at,omitempty" db:"last_eval_at"`
	LastEvalResult  *string          `json:"last_eval_result,omitempty" db:"last_eval_result"`
	EvalCount       int64            `json:"eval_count" db:"eval_count"`
	AlertCount      int64            `json:"alert_count" db:"alert_count"`
	CreatedBy       string           `json:"created_by" db:"created_by"`
	UpdatedBy       *string          `json:"updated_by,omitempty" db:"updated_by"`
	CreatedAt       time.Time        `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time        `json:"updated_at" db:"updated_at"`
	DeletedAt       *time.Time       `json:"deleted_at,omitempty" db:"deleted_at"`
}

// RuleCreateRequest 创建规则请求
type RuleCreateRequest struct {
	DataSourceID       string            `json:"data_source_id" binding:"required"`
	Name               string            `json:"name" binding:"required,min=1,max=200"`
	Description        string            `json:"description" binding:"required,min=1,max=1000"`
	Type               RuleType          `json:"type" binding:"required"`
	Severity           AlertSeverity     `json:"severity" binding:"required"`
	Expression         string            `json:"expression" binding:"required"`
	Conditions         []RuleCondition   `json:"conditions,omitempty"`
	Actions            []RuleAction      `json:"actions,omitempty"`
	Labels             map[string]string `json:"labels,omitempty"`
	Annotations        map[string]string `json:"annotations,omitempty"`
	EvaluationInterval time.Duration     `json:"evaluation_interval" binding:"required"`
	ForDuration        time.Duration     `json:"for_duration"`
	Threshold          *float64          `json:"threshold,omitempty"`
	RecoveryThreshold  *float64          `json:"recovery_threshold,omitempty"`
	NoDataState        *string           `json:"no_data_state,omitempty"`
	ExecErrState       *string           `json:"exec_err_state,omitempty"`
}

// RuleUpdateRequest 更新规则请求
type RuleUpdateRequest struct {
	Name               *string            `json:"name,omitempty" binding:"omitempty,min=1,max=200"`
	Description        *string            `json:"description,omitempty" binding:"omitempty,min=1,max=1000"`
	Type               *RuleType          `json:"type,omitempty"`
	Status             *RuleStatus        `json:"status,omitempty"`
	Severity           *AlertSeverity     `json:"severity,omitempty"`
	Expression         *string            `json:"expression,omitempty"`
	Conditions         *[]RuleCondition   `json:"conditions,omitempty"`
	Actions            *[]RuleAction      `json:"actions,omitempty"`
	Labels             *map[string]string `json:"labels,omitempty"`
	Annotations        *map[string]string `json:"annotations,omitempty"`
	EvaluationInterval *time.Duration     `json:"evaluation_interval,omitempty"`
	ForDuration        *time.Duration     `json:"for_duration,omitempty"`
	Threshold          *float64           `json:"threshold,omitempty"`
	RecoveryThreshold  *float64           `json:"recovery_threshold,omitempty"`
	NoDataState        *string            `json:"no_data_state,omitempty"`
	ExecErrState       *string            `json:"exec_err_state,omitempty"`
}

// RuleTestRequest 测试规则请求
type RuleTestRequest struct {
	Expression string          `json:"expression" binding:"required"`
	Conditions []RuleCondition `json:"conditions,omitempty"`
	TimeRange  *TimeRange      `json:"time_range,omitempty"`
}

// RuleTestResult 测试规则结果
type RuleTestResult struct {
	Success    bool                   `json:"success"`
	Result     interface{}            `json:"result,omitempty"`
	Error      *string                `json:"error,omitempty"`
	EvalTime   time.Duration          `json:"eval_time"`
	DataPoints []map[string]interface{} `json:"data_points,omitempty"`
}

// RuleFilter 规则查询过滤器
type RuleFilter struct {
	DataSourceID *string       `json:"data_source_id,omitempty"`
	Type         *RuleType     `json:"type,omitempty"`
	Status       *RuleStatus   `json:"status,omitempty"`
	Severity     *AlertSeverity `json:"severity,omitempty"`
	Enabled      *bool         `json:"enabled,omitempty"`
	Keyword      *string       `json:"keyword,omitempty"` // 搜索名称、描述
	Labels       map[string]string `json:"labels,omitempty"`
	CreatedBy    *string       `json:"created_by,omitempty"`
	StartTime    *time.Time    `json:"start_time,omitempty"`
	EndTime      *time.Time    `json:"end_time,omitempty"`
	Page         int           `json:"page" binding:"min=1"`
	PageSize     int           `json:"page_size" binding:"min=1,max=100"`
	SortBy       *string       `json:"sort_by,omitempty"`
	SortOrder    *string       `json:"sort_order,omitempty"` // asc, desc
}

// RuleList 规则列表响应
type RuleList struct {
	Rules      []*Rule `json:"rules"`
	Total      int64   `json:"total"`
	Page       int     `json:"page"`
	PageSize   int     `json:"page_size"`
	TotalPages int64   `json:"total_pages"`
}

// RuleStats 规则统计
type RuleStats struct {
	Total    int64                 `json:"total"`
	ByType   map[RuleType]int64    `json:"by_type"`
	ByStatus map[RuleStatus]int64  `json:"by_status"`
	BySeverity map[AlertSeverity]int64 `json:"by_severity"`
	ActiveRules int64              `json:"active_rules"`
	FiringRules int64              `json:"firing_rules"`
	Enabled     int64              `json:"enabled"`
	Disabled    int64              `json:"disabled"`
	Active      int64              `json:"active"`
	Triggered   int64              `json:"triggered"`
	Pending     int64              `json:"pending"`
}

// TimeRange 时间范围
type TimeRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// 验证方法

// Validate 验证规则数据
func (r *Rule) Validate() error {
	if strings.TrimSpace(r.Name) == "" {
		return errors.New("规则名称不能为空")
	}
	
	if len(r.Name) > 200 {
		return errors.New("规则名称长度不能超过200个字符")
	}
	
	if strings.TrimSpace(r.Description) == "" {
		return errors.New("规则描述不能为空")
	}
	
	if len(r.Description) > 1000 {
		return errors.New("规则描述长度不能超过1000个字符")
	}
	
	if !r.Type.IsValid() {
		return errors.New("无效的规则类型")
	}
	
	if !r.Status.IsValid() {
		return errors.New("无效的规则状态")
	}
	
	if !r.Severity.IsValid() {
		return errors.New("无效的告警严重级别")
	}
	
	if strings.TrimSpace(r.DataSourceID) == "" {
		return errors.New("数据源ID不能为空")
	}
	
	if strings.TrimSpace(r.Expression) == "" {
		return errors.New("规则表达式不能为空")
	}
	
	if r.EvaluationInterval <= 0 {
		return errors.New("评估间隔必须大于0")
	}
	
	if r.ForDuration < 0 {
		return errors.New("持续时间不能为负数")
	}
	
	if strings.TrimSpace(r.CreatedBy) == "" {
		return errors.New("创建者不能为空")
	}
	
	// 验证条件
	for i, condition := range r.Conditions {
		if err := condition.Validate(); err != nil {
			return errors.New("条件 " + string(rune(i+1)) + " 验证失败: " + err.Error())
		}
	}
	
	// 验证动作
	for i, action := range r.Actions {
		if err := action.Validate(); err != nil {
			return errors.New("动作 " + string(rune(i+1)) + " 验证失败: " + err.Error())
		}
	}
	
	return nil
}

// IsValid 检查规则类型是否有效
func (t RuleType) IsValid() bool {
	switch t {
	case RuleTypeMetric, RuleTypeLog, RuleTypeComposite, RuleTypeAnomaly:
		return true
	default:
		return false
	}
}

// IsValid 检查规则状态是否有效
func (s RuleStatus) IsValid() bool {
	switch s {
	case RuleStatusActive, RuleStatusInactive, RuleStatusDisabled, RuleStatusTesting:
		return true
	default:
		return false
	}
}

// IsValid 检查规则操作符是否有效
func (o RuleOperator) IsValid() bool {
	switch o {
	case RuleOperatorGT, RuleOperatorGTE, RuleOperatorLT, RuleOperatorLTE,
		 RuleOperatorEQ, RuleOperatorNE, RuleOperatorIN, RuleOperatorNIN:
		return true
	default:
		return false
	}
}

// Validate 验证规则条件
func (c *RuleCondition) Validate() error {
	if strings.TrimSpace(c.Field) == "" {
		return errors.New("字段名不能为空")
	}
	
	if !c.Operator.IsValid() {
		return errors.New("无效的操作符")
	}
	
	if c.Value == nil {
		return errors.New("值不能为空")
	}
	
	if c.Logic != nil && *c.Logic != "AND" && *c.Logic != "OR" {
		return errors.New("逻辑操作符只能是 AND 或 OR")
	}
	
	return nil
}

// Validate 验证规则动作
func (a *RuleAction) Validate() error {
	if strings.TrimSpace(a.Type) == "" {
		return errors.New("动作类型不能为空")
	}
	
	validTypes := []string{"alert", "webhook", "email", "sms", "dingtalk", "wechat", "slack"}
	valid := false
	for _, validType := range validTypes {
		if a.Type == validType {
			valid = true
			break
		}
	}
	if !valid {
		return errors.New("无效的动作类型")
	}
	
	if strings.TrimSpace(a.Target) == "" {
		return errors.New("目标地址不能为空")
	}
	
	return nil
}

// IsActive 检查规则是否激活
func (r *Rule) IsActive() bool {
	return r.Status == RuleStatusActive
}

// IsDisabled 检查规则是否禁用
func (r *Rule) IsDisabled() bool {
	return r.Status == RuleStatusDisabled
}

// IsTesting 检查规则是否在测试中
func (r *Rule) IsTesting() bool {
	return r.Status == RuleStatusTesting
}

// ShouldEvaluate 检查规则是否应该被评估
func (r *Rule) ShouldEvaluate() bool {
	if !r.IsActive() {
		return false
	}
	
	if r.LastEvalAt == nil {
		return true
	}
	
	return time.Since(*r.LastEvalAt) >= r.EvaluationInterval
}

// Validate 验证创建规则请求
func (req *RuleCreateRequest) Validate() error {
	if strings.TrimSpace(req.Name) == "" {
		return errors.New("规则名称不能为空")
	}
	
	if len(req.Name) > 200 {
		return errors.New("规则名称长度不能超过200个字符")
	}
	
	if strings.TrimSpace(req.Description) == "" {
		return errors.New("规则描述不能为空")
	}
	
	if len(req.Description) > 1000 {
		return errors.New("规则描述长度不能超过1000个字符")
	}
	
	if !req.Type.IsValid() {
		return errors.New("无效的规则类型")
	}
	
	if !req.Severity.IsValid() {
		return errors.New("无效的告警严重级别")
	}
	
	if strings.TrimSpace(req.DataSourceID) == "" {
		return errors.New("数据源ID不能为空")
	}
	
	if strings.TrimSpace(req.Expression) == "" {
		return errors.New("规则表达式不能为空")
	}
	
	if req.EvaluationInterval <= 0 {
		return errors.New("评估间隔必须大于0")
	}
	
	if req.ForDuration < 0 {
		return errors.New("持续时间不能为负数")
	}
	
	// 验证条件
	for i, condition := range req.Conditions {
		if err := condition.Validate(); err != nil {
			return errors.New("条件 " + string(rune(i+1)) + " 验证失败: " + err.Error())
		}
	}
	
	// 验证动作
	for i, action := range req.Actions {
		if err := action.Validate(); err != nil {
			return errors.New("动作 " + string(rune(i+1)) + " 验证失败: " + err.Error())
		}
	}
	
	return nil
}

// MarshalConditions 序列化条件为JSON
func (r *Rule) MarshalConditions() ([]byte, error) {
	if r.Conditions == nil {
		return json.Marshal([]RuleCondition{})
	}
	return json.Marshal(r.Conditions)
}

// UnmarshalConditions 反序列化条件从JSON
func (r *Rule) UnmarshalConditions(data []byte) error {
	return json.Unmarshal(data, &r.Conditions)
}

// MarshalActions 序列化动作为JSON
func (r *Rule) MarshalActions() ([]byte, error) {
	if r.Actions == nil {
		return json.Marshal([]RuleAction{})
	}
	return json.Marshal(r.Actions)
}

// UnmarshalActions 反序列化动作从JSON
func (r *Rule) UnmarshalActions(data []byte) error {
	return json.Unmarshal(data, &r.Actions)
}

// MarshalLabels 序列化标签为JSON
func (r *Rule) MarshalLabels() ([]byte, error) {
	if r.Labels == nil {
		return json.Marshal(map[string]string{})
	}
	return json.Marshal(r.Labels)
}

// UnmarshalLabels 反序列化标签从JSON
func (r *Rule) UnmarshalLabels(data []byte) error {
	if r.Labels == nil {
		r.Labels = make(map[string]string)
	}
	return json.Unmarshal(data, &r.Labels)
}

// MarshalAnnotations 序列化注解为JSON
func (r *Rule) MarshalAnnotations() ([]byte, error) {
	if r.Annotations == nil {
		return json.Marshal(map[string]string{})
	}
	return json.Marshal(r.Annotations)
}

// UnmarshalAnnotations 反序列化注解从JSON
func (r *Rule) UnmarshalAnnotations(data []byte) error {
	if r.Annotations == nil {
		r.Annotations = make(map[string]string)
	}
	return json.Unmarshal(data, &r.Annotations)
}