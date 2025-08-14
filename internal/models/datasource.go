package models

import (
	"time"
	"errors"
	"strings"
	"encoding/json"
	"net/url"
)

// DataSourceType 数据源类型
type DataSourceType string

const (
	DataSourceTypePrometheus DataSourceType = "prometheus" // Prometheus
	DataSourceTypeInfluxDB   DataSourceType = "influxdb"   // InfluxDB
	DataSourceTypeElastic    DataSourceType = "elastic"    // Elasticsearch
	DataSourceTypeMySQL      DataSourceType = "mysql"      // MySQL
	DataSourceTypePostgreSQL DataSourceType = "postgresql" // PostgreSQL
	DataSourceTypeRedis      DataSourceType = "redis"      // Redis
	DataSourceTypeKafka      DataSourceType = "kafka"      // Kafka
	DataSourceTypeGrafana    DataSourceType = "grafana"    // Grafana
	DataSourceTypeZabbix     DataSourceType = "zabbix"     // Zabbix
	DataSourceTypeCustom     DataSourceType = "custom"     // 自定义
)

// DataSourceStatus 数据源状态
type DataSourceStatus string

const (
	DataSourceStatusActive      DataSourceStatus = "active"      // 激活
	DataSourceStatusInactive    DataSourceStatus = "inactive"    // 未激活
	DataSourceStatusDisabled    DataSourceStatus = "disabled"    // 禁用
	DataSourceStatusError       DataSourceStatus = "error"       // 错误
	DataSourceStatusMaintenance DataSourceStatus = "maintenance" // 维护中
)

// DataSourceConfig 数据源配置
type DataSourceConfig struct {
	URL              string            `json:"url"`
	Username         *string           `json:"username,omitempty"`
	Password         *string           `json:"password,omitempty"`
	Token            *string           `json:"token,omitempty"`
	Database         *string           `json:"database,omitempty"`
	Timeout          *time.Duration    `json:"timeout,omitempty"`
	MaxConnections   *int              `json:"max_connections,omitempty"`
	SSLMode          *string           `json:"ssl_mode,omitempty"`
	Headers          map[string]string `json:"headers,omitempty"`
	Parameters       map[string]interface{} `json:"parameters,omitempty"`
	RetentionPolicy  *string           `json:"retention_policy,omitempty"`
	Measurement      *string           `json:"measurement,omitempty"`
	Index            *string           `json:"index,omitempty"`
	Topic            *string           `json:"topic,omitempty"`
}

// DataSource 数据源模型
type DataSource struct {
	ID              string            `json:"id" db:"id"`
	Name            string            `json:"name" db:"name"`
	Description     string            `json:"description" db:"description"`
	Type            DataSourceType    `json:"type" db:"type"`
	Status          DataSourceStatus  `json:"status" db:"status"`
	Config          DataSourceConfig  `json:"config" db:"config"`
	Tags            []string          `json:"tags" db:"tags"`
	Version         *string           `json:"version,omitempty" db:"version"`
	HealthCheckURL  *string           `json:"health_check_url,omitempty" db:"health_check_url"`
	HealthStatus    *string           `json:"health_status,omitempty" db:"health_status"`
	LastHealthCheck *time.Time        `json:"last_health_check,omitempty" db:"last_health_check"`
	ErrorMessage    *string           `json:"error_message,omitempty" db:"error_message"`
	Metrics         *DataSourceMetrics `json:"metrics,omitempty" db:"metrics"`
	CreatedBy       string            `json:"created_by" db:"created_by"`
	UpdatedBy       *string           `json:"updated_by,omitempty" db:"updated_by"`
	CreatedAt       time.Time         `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time         `json:"updated_at" db:"updated_at"`
	DeletedAt       *time.Time        `json:"deleted_at,omitempty" db:"deleted_at"`
}

// DataSourceMetrics 数据源指标
type DataSourceMetrics struct {
	ConnectionCount   int64     `json:"connection_count"`
	QueryCount        int64     `json:"query_count"`
	ErrorCount        int64     `json:"error_count"`
	AvgResponseTime   float64   `json:"avg_response_time"`
	LastQueryAt       *time.Time `json:"last_query_at,omitempty"`
	DataSize          *int64    `json:"data_size,omitempty"`
	Uptime            *float64  `json:"uptime,omitempty"`
}

// DataSourceCreateRequest 创建数据源请求
type DataSourceCreateRequest struct {
	Name            string            `json:"name" binding:"required,min=1,max=100"`
	Description     string            `json:"description" binding:"required,min=1,max=500"`
	Type            DataSourceType    `json:"type" binding:"required"`
	Config          DataSourceConfig  `json:"config" binding:"required"`
	Tags            []string          `json:"tags,omitempty"`
	Version         *string           `json:"version,omitempty"`
	HealthCheckURL  *string           `json:"health_check_url,omitempty"`
}

// DataSourceUpdateRequest 更新数据源请求
type DataSourceUpdateRequest struct {
	Name            *string           `json:"name,omitempty" binding:"omitempty,min=1,max=100"`
	Description     *string           `json:"description,omitempty" binding:"omitempty,min=1,max=500"`
	Status          *DataSourceStatus `json:"status,omitempty"`
	Config          *DataSourceConfig `json:"config,omitempty"`
	Tags            *[]string         `json:"tags,omitempty"`
	Version         *string           `json:"version,omitempty"`
	HealthCheckURL  *string           `json:"health_check_url,omitempty"`
}

// DataSourceTestRequest 测试数据源请求
type DataSourceTestRequest struct {
	Type   DataSourceType   `json:"type" binding:"required"`
	Config DataSourceConfig `json:"config" binding:"required"`
}

// DataSourceTestResult 测试数据源结果
type DataSourceTestResult struct {
	Success      bool          `json:"success"`
	Message      string        `json:"message"`
	ResponseTime time.Duration `json:"response_time"`
	Version      *string       `json:"version,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	Error        *string       `json:"error,omitempty"`
}

// DataSourceFilter 数据源查询过滤器
type DataSourceFilter struct {
	Type         *DataSourceType   `json:"type,omitempty"`
	Status       *DataSourceStatus `json:"status,omitempty"`
	Keyword      *string           `json:"keyword,omitempty"` // 搜索名称、描述
	Tags         []string          `json:"tags,omitempty"`
	CreatedBy    *string           `json:"created_by,omitempty"`
	HealthStatus *string           `json:"health_status,omitempty"`
	StartTime    *time.Time        `json:"start_time,omitempty"`
	EndTime      *time.Time        `json:"end_time,omitempty"`
	Page         int               `json:"page" binding:"min=1"`
	PageSize     int               `json:"page_size" binding:"min=1,max=100"`
	SortBy       *string           `json:"sort_by,omitempty"`
	SortOrder    *string           `json:"sort_order,omitempty"` // asc, desc
}

// DataSourceList 数据源列表响应
type DataSourceList struct {
	DataSources []*DataSource `json:"data_sources"`
	Total       int64         `json:"total"`
	Page        int           `json:"page"`
	PageSize    int           `json:"page_size"`
	TotalPages  int           `json:"total_pages"`
}

// DataSourceStats 数据源统计
type DataSourceStats struct {
	Total        int64                        `json:"total"`
	ByType       map[DataSourceType]int64     `json:"by_type"`
	ByStatus     map[DataSourceStatus]int64   `json:"by_status"`
	HealthyCount int64                        `json:"healthy_count"`
	ErrorCount   int64                        `json:"error_count"`
	TotalQueries int64                        `json:"total_queries"`
	AvgUptime    float64                      `json:"avg_uptime"`
}

// DataSourceQuery 数据源查询请求
type DataSourceQuery struct {
	DataSourceID string                 `json:"data_source_id" binding:"required"`
	Query        string                 `json:"query" binding:"required"`
	TimeRange    *TimeRange             `json:"time_range,omitempty"`
	Parameters   map[string]interface{} `json:"parameters,omitempty"`
	Limit        *int                   `json:"limit,omitempty"`
	Offset       *int                   `json:"offset,omitempty"`
}

// DataSourceQueryResult 数据源查询结果
type DataSourceQueryResult struct {
	Success      bool                     `json:"success"`
	Data         []map[string]interface{} `json:"data,omitempty"`
	Columns      []string                 `json:"columns,omitempty"`
	RowCount     int64                    `json:"row_count"`
	QueryTime    time.Duration            `json:"query_time"`
	Error        *string                  `json:"error,omitempty"`
	Metadata     map[string]interface{}   `json:"metadata,omitempty"`
}

// 验证方法

// Validate 验证数据源数据
func (ds *DataSource) Validate() error {
	if strings.TrimSpace(ds.Name) == "" {
		return errors.New("数据源名称不能为空")
	}
	
	if len(ds.Name) > 100 {
		return errors.New("数据源名称长度不能超过100个字符")
	}
	
	if strings.TrimSpace(ds.Description) == "" {
		return errors.New("数据源描述不能为空")
	}
	
	if len(ds.Description) > 500 {
		return errors.New("数据源描述长度不能超过500个字符")
	}
	
	if !ds.Type.IsValid() {
		return errors.New("无效的数据源类型")
	}
	
	if !ds.Status.IsValid() {
		return errors.New("无效的数据源状态")
	}
	
	if err := ds.Config.Validate(); err != nil {
		return errors.New("数据源配置验证失败: " + err.Error())
	}
	
	if strings.TrimSpace(ds.CreatedBy) == "" {
		return errors.New("创建者不能为空")
	}
	
	return nil
}

// IsValid 检查数据源类型是否有效
func (t DataSourceType) IsValid() bool {
	switch t {
	case DataSourceTypePrometheus, DataSourceTypeInfluxDB, DataSourceTypeElastic,
		 DataSourceTypeMySQL, DataSourceTypePostgreSQL, DataSourceTypeRedis,
		 DataSourceTypeKafka, DataSourceTypeGrafana, DataSourceTypeZabbix,
		 DataSourceTypeCustom:
		return true
	default:
		return false
	}
}

// IsValid 检查数据源状态是否有效
func (s DataSourceStatus) IsValid() bool {
	switch s {
	case DataSourceStatusActive, DataSourceStatusInactive, DataSourceStatusDisabled,
		 DataSourceStatusError, DataSourceStatusMaintenance:
		return true
	default:
		return false
	}
}

// Validate 验证数据源配置
func (c *DataSourceConfig) Validate() error {
	if strings.TrimSpace(c.URL) == "" {
		return errors.New("数据源URL不能为空")
	}
	
	// 验证URL格式
	if _, err := url.Parse(c.URL); err != nil {
		return errors.New("无效的URL格式: " + err.Error())
	}
	
	// 验证超时时间
	if c.Timeout != nil && *c.Timeout <= 0 {
		return errors.New("超时时间必须大于0")
	}
	
	// 验证最大连接数
	if c.MaxConnections != nil && *c.MaxConnections <= 0 {
		return errors.New("最大连接数必须大于0")
	}
	
	// 验证SSL模式
	if c.SSLMode != nil {
		validSSLModes := []string{"disable", "require", "verify-ca", "verify-full"}
		valid := false
		for _, mode := range validSSLModes {
			if *c.SSLMode == mode {
				valid = true
				break
			}
		}
		if !valid {
			return errors.New("无效的SSL模式")
		}
	}
	
	return nil
}

// IsActive 检查数据源是否激活
func (ds *DataSource) IsActive() bool {
	return ds.Status == DataSourceStatusActive
}

// IsHealthy 检查数据源是否健康
func (ds *DataSource) IsHealthy() bool {
	return ds.HealthStatus != nil && *ds.HealthStatus == "healthy"
}

// IsError 检查数据源是否有错误
func (ds *DataSource) IsError() bool {
	return ds.Status == DataSourceStatusError
}

// GetConnectionString 获取连接字符串（隐藏敏感信息）
func (ds *DataSource) GetConnectionString() string {
	if ds.Config.Password != nil {
		// 隐藏密码
		return strings.Replace(ds.Config.URL, *ds.Config.Password, "***", -1)
	}
	return ds.Config.URL
}

// Validate 验证创建数据源请求
func (req *DataSourceCreateRequest) Validate() error {
	if strings.TrimSpace(req.Name) == "" {
		return errors.New("数据源名称不能为空")
	}
	
	if len(req.Name) > 100 {
		return errors.New("数据源名称长度不能超过100个字符")
	}
	
	if strings.TrimSpace(req.Description) == "" {
		return errors.New("数据源描述不能为空")
	}
	
	if len(req.Description) > 500 {
		return errors.New("数据源描述长度不能超过500个字符")
	}
	
	if !req.Type.IsValid() {
		return errors.New("无效的数据源类型")
	}
	
	if err := req.Config.Validate(); err != nil {
		return errors.New("数据源配置验证失败: " + err.Error())
	}
	
	return nil
}

// Validate 验证数据源查询请求
func (req *DataSourceQuery) Validate() error {
	if strings.TrimSpace(req.DataSourceID) == "" {
		return errors.New("数据源ID不能为空")
	}
	
	if strings.TrimSpace(req.Query) == "" {
		return errors.New("查询语句不能为空")
	}
	
	if req.Limit != nil && *req.Limit <= 0 {
		return errors.New("限制数量必须大于0")
	}
	
	if req.Offset != nil && *req.Offset < 0 {
		return errors.New("偏移量不能为负数")
	}
	
	return nil
}

// MarshalConfig 序列化配置为JSON
func (ds *DataSource) MarshalConfig() ([]byte, error) {
	return json.Marshal(ds.Config)
}

// UnmarshalConfig 反序列化配置从JSON
func (ds *DataSource) UnmarshalConfig(data []byte) error {
	return json.Unmarshal(data, &ds.Config)
}

// MarshalTags 序列化标签为JSON
func (ds *DataSource) MarshalTags() ([]byte, error) {
	if ds.Tags == nil {
		return json.Marshal([]string{})
	}
	return json.Marshal(ds.Tags)
}

// UnmarshalTags 反序列化标签从JSON
func (ds *DataSource) UnmarshalTags(data []byte) error {
	return json.Unmarshal(data, &ds.Tags)
}

// MarshalMetrics 序列化指标为JSON
func (ds *DataSource) MarshalMetrics() ([]byte, error) {
	if ds.Metrics == nil {
		return json.Marshal(nil)
	}
	return json.Marshal(ds.Metrics)
}

// UnmarshalMetrics 反序列化指标从JSON
func (ds *DataSource) UnmarshalMetrics(data []byte) error {
	if len(data) == 0 || string(data) == "null" {
		ds.Metrics = nil
		return nil
	}
	if ds.Metrics == nil {
		ds.Metrics = &DataSourceMetrics{}
	}
	return json.Unmarshal(data, ds.Metrics)
}

// GetDisplayName 获取显示名称
func (t DataSourceType) GetDisplayName() string {
	switch t {
	case DataSourceTypePrometheus:
		return "Prometheus"
	case DataSourceTypeInfluxDB:
		return "InfluxDB"
	case DataSourceTypeElastic:
		return "Elasticsearch"
	case DataSourceTypeMySQL:
		return "MySQL"
	case DataSourceTypePostgreSQL:
		return "PostgreSQL"
	case DataSourceTypeRedis:
		return "Redis"
	case DataSourceTypeKafka:
		return "Kafka"
	case DataSourceTypeGrafana:
		return "Grafana"
	case DataSourceTypeZabbix:
		return "Zabbix"
	case DataSourceTypeCustom:
		return "自定义"
	default:
		return string(t)
	}
}

// GetDefaultPort 获取默认端口
func (t DataSourceType) GetDefaultPort() int {
	switch t {
	case DataSourceTypePrometheus:
		return 9090
	case DataSourceTypeInfluxDB:
		return 8086
	case DataSourceTypeElastic:
		return 9200
	case DataSourceTypeMySQL:
		return 3306
	case DataSourceTypePostgreSQL:
		return 5432
	case DataSourceTypeRedis:
		return 6379
	case DataSourceTypeKafka:
		return 9092
	case DataSourceTypeGrafana:
		return 3000
	case DataSourceTypeZabbix:
		return 10051
	default:
		return 80
	}
}