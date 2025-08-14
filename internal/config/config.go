package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
)

// Config 应用配置结构
type Config struct {
	// 应用基本配置
	App AppConfig `mapstructure:",squash"`

	// 数据库配置
	Database DatabaseConfig `mapstructure:",squash"`

	// Redis 配置
	Redis RedisConfig `mapstructure:",squash"`

	// JWT 配置
	JWT JWTConfig `mapstructure:",squash"`

	// 告警配置
	Alert AlertConfig `mapstructure:",squash"`

	// 通知配置
	Notification NotificationConfig `mapstructure:",squash"`

	// 监控数据源配置
	DataSources DataSourcesConfig `mapstructure:",squash"`

	// 文件存储配置
	FileStorage FileStorageConfig `mapstructure:",squash"`

	// 安全配置
	Security SecurityConfig `mapstructure:",squash"`

	// 性能配置
	Performance PerformanceConfig `mapstructure:",squash"`

	// 健康检查配置
	HealthCheck HealthCheckConfig `mapstructure:",squash"`
}

// AppConfig 应用基本配置
type AppConfig struct {
	Env         string `mapstructure:"APP_ENV" validate:"oneof=development staging production"`
	Name        string `mapstructure:"APP_NAME" validate:"required"`
	Version     string `mapstructure:"APP_VERSION" validate:"required"`
	Port        int    `mapstructure:"PORT" validate:"required,min=1,max=65535"`
	Host        string `mapstructure:"APP_HOST"`
	Debug       bool   `mapstructure:"DEBUG"`
	Environment string `mapstructure:"APP_ENVIRONMENT"`

	// 日志配置
	LogLevel  string `mapstructure:"LOG_LEVEL" validate:"oneof=debug info warn error"`
	LogFormat string `mapstructure:"LOG_FORMAT" validate:"oneof=json text"`

	// 性能分析配置
	PProfEnabled bool `mapstructure:"PPROF_ENABLED"`
	PProfPort    int  `mapstructure:"PPROF_PORT" validate:"min=1,max=65535"`

	// API 文档配置
	APIDocsEnabled bool   `mapstructure:"API_DOCS_ENABLED"`
	APIDocsPath    string `mapstructure:"API_DOCS_PATH"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Host            string        `mapstructure:"DB_HOST"`
	Port            int           `mapstructure:"DB_PORT"`
	User            string        `mapstructure:"DB_USER"`
	Password        string        `mapstructure:"DB_PASSWORD"`
	Name            string        `mapstructure:"DB_NAME"`
	SSLMode         string        `mapstructure:"DB_SSL_MODE"`
	MaxOpenConns    int           `mapstructure:"DB_MAX_OPEN_CONNS"`
	MaxIdleConns    int           `mapstructure:"DB_MAX_IDLE_CONNS"`
	ConnMaxLifetime time.Duration `mapstructure:"DB_CONN_MAX_LIFETIME"`
	ConnMaxIdleTime time.Duration `mapstructure:"DB_CONN_MAX_IDLE_TIME"`
	MigrationPath   string        `mapstructure:"DB_MIGRATION_PATH"`
	MigrationTable  string        `mapstructure:"DB_MIGRATION_TABLE"`
	AutoMigrate     bool          `mapstructure:"DB_AUTO_MIGRATE"`
}

// RedisConfig Redis 配置
type RedisConfig struct {
	Host         string `mapstructure:"REDIS_HOST" validate:"required"`
	Port         int    `mapstructure:"REDIS_PORT" validate:"required,min=1,max=65535"`
	Password     string `mapstructure:"REDIS_PASSWORD"`
	DB           int    `mapstructure:"REDIS_DB" validate:"min=0,max=15"`
	PoolSize     int    `mapstructure:"REDIS_POOL_SIZE" validate:"min=1"`
	MinIdleConns int    `mapstructure:"REDIS_MIN_IDLE_CONNS" validate:"min=0"`
}

// JWTConfig JWT 配置
type JWTConfig struct {
	Secret             string        `mapstructure:"JWT_SECRET" validate:"required,min=32"`
	AccessTokenExpire  time.Duration `mapstructure:"JWT_ACCESS_TOKEN_EXPIRE"`
	RefreshTokenExpire time.Duration `mapstructure:"JWT_REFRESH_TOKEN_EXPIRE"`
}

// AlertConfig 告警配置
type AlertConfig struct {
	EvaluationInterval       time.Duration `mapstructure:"ALERT_EVALUATION_INTERVAL"`
	HistoryRetentionDays     int           `mapstructure:"ALERT_HISTORY_RETENTION_DAYS" validate:"min=1"`
	MaxConcurrentEvaluations int           `mapstructure:"ALERT_MAX_CONCURRENT_EVALUATIONS" validate:"min=1"`
}

// NotificationConfig 通知配置
type NotificationConfig struct {
	// 邮件配置
	SMTP SMTPConfig `mapstructure:",squash"`

	// 钉钉配置
	DingTalk DingTalkConfig `mapstructure:",squash"`

	// 企业微信配置
	WeCom WeComConfig `mapstructure:",squash"`

	// Slack 配置
	Slack SlackConfig `mapstructure:",squash"`
}

// SMTPConfig 邮件配置
type SMTPConfig struct {
	Host     string `mapstructure:"SMTP_HOST"`
	Port     int    `mapstructure:"SMTP_PORT" validate:"min=1,max=65535"`
	Username string `mapstructure:"SMTP_USERNAME"`
	Password string `mapstructure:"SMTP_PASSWORD"`
	From     string `mapstructure:"SMTP_FROM" validate:"email"`
	TLS      bool   `mapstructure:"SMTP_TLS"`
}

// DingTalkConfig 钉钉配置
type DingTalkConfig struct {
	WebhookURL string `mapstructure:"DINGTALK_WEBHOOK_URL" validate:"url"`
	Secret     string `mapstructure:"DINGTALK_SECRET"`
}

// WeComConfig 企业微信配置
type WeComConfig struct {
	WebhookURL string `mapstructure:"WECOM_WEBHOOK_URL" validate:"url"`
}

// SlackConfig Slack 配置
type SlackConfig struct {
	WebhookURL string `mapstructure:"SLACK_WEBHOOK_URL" validate:"url"`
}

// DataSourcesConfig 监控数据源配置
type DataSourcesConfig struct {
	Prometheus PrometheusConfig `mapstructure:",squash"`
	Grafana    GrafanaConfig    `mapstructure:",squash"`
	InfluxDB   InfluxDBConfig   `mapstructure:",squash"`
}

// PrometheusConfig Prometheus 配置
type PrometheusConfig struct {
	URL        string        `mapstructure:"PROMETHEUS_URL" validate:"url"`
	Timeout    time.Duration `mapstructure:"PROMETHEUS_TIMEOUT"`
	MaxSamples int           `mapstructure:"PROMETHEUS_MAX_SAMPLES" validate:"min=1"`
}

// GrafanaConfig Grafana 配置
type GrafanaConfig struct {
	URL      string `mapstructure:"GRAFANA_URL" validate:"url"`
	APIKey   string `mapstructure:"GRAFANA_API_KEY"`
	Username string `mapstructure:"GRAFANA_USERNAME"`
	Password string `mapstructure:"GRAFANA_PASSWORD"`
}

// InfluxDBConfig InfluxDB 配置
type InfluxDBConfig struct {
	URL      string `mapstructure:"INFLUXDB_URL" validate:"url"`
	Token    string `mapstructure:"INFLUXDB_TOKEN"`
	Org      string `mapstructure:"INFLUXDB_ORG"`
	Bucket   string `mapstructure:"INFLUXDB_BUCKET"`
	Username string `mapstructure:"INFLUXDB_USERNAME"`
	Password string `mapstructure:"INFLUXDB_PASSWORD"`
}

// FileStorageConfig 文件存储配置
type FileStorageConfig struct {
	Type      string `mapstructure:"FILE_STORAGE_TYPE" validate:"oneof=local s3 oss"`
	LocalPath string `mapstructure:"FILE_STORAGE_LOCAL_PATH"`
	S3        S3Config `mapstructure:",squash"`
	OSS       OSSConfig `mapstructure:",squash"`
}

// S3Config S3 配置
type S3Config struct {
	Region          string `mapstructure:"S3_REGION"`
	Bucket          string `mapstructure:"S3_BUCKET"`
	AccessKeyID     string `mapstructure:"S3_ACCESS_KEY_ID"`
	SecretAccessKey string `mapstructure:"S3_SECRET_ACCESS_KEY"`
	Endpoint        string `mapstructure:"S3_ENDPOINT"`
}

// OSSConfig 阿里云 OSS 配置
type OSSConfig struct {
	Region          string `mapstructure:"OSS_REGION"`
	Bucket          string `mapstructure:"OSS_BUCKET"`
	AccessKeyID     string `mapstructure:"OSS_ACCESS_KEY_ID"`
	SecretAccessKey string `mapstructure:"OSS_SECRET_ACCESS_KEY"`
	Endpoint        string `mapstructure:"OSS_ENDPOINT"`
}

// SecurityConfig 安全配置
type SecurityConfig struct {
	// CORS 配置
	CORSAllowedOrigins []string `mapstructure:"CORS_ALLOWED_ORIGINS"`
	CORSAllowedMethods []string `mapstructure:"CORS_ALLOWED_METHODS"`
	CORSAllowedHeaders []string `mapstructure:"CORS_ALLOWED_HEADERS"`

	// 频率限制配置
	RateLimitEnabled bool `mapstructure:"RATE_LIMIT_ENABLED"`
	RateLimitRPS     int  `mapstructure:"RATE_LIMIT_RPS" validate:"min=1"`
	RateLimitBurst   int  `mapstructure:"RATE_LIMIT_BURST" validate:"min=1"`

	// API Key 配置
	APIKeyEnabled bool   `mapstructure:"API_KEY_ENABLED"`
	APIKeyHeader  string `mapstructure:"API_KEY_HEADER"`
}

// PerformanceConfig 性能配置
type PerformanceConfig struct {
	MaxRequestSize int           `mapstructure:"PERF_MAX_REQUEST_SIZE"`
	MaxConcurrency int           `mapstructure:"PERF_MAX_CONCURRENCY"`
	ReadTimeout    time.Duration `mapstructure:"PERF_READ_TIMEOUT"`
	WriteTimeout   time.Duration `mapstructure:"PERF_WRITE_TIMEOUT"`
	IdleTimeout    time.Duration `mapstructure:"PERF_IDLE_TIMEOUT"`

	// 工作池配置
	WorkerPoolSize  int `mapstructure:"WORKER_POOL_SIZE" validate:"min=1"`
	QueueBufferSize int `mapstructure:"QUEUE_BUFFER_SIZE" validate:"min=1"`
}

// HealthCheckConfig 健康检查配置
type HealthCheckConfig struct {
	Enabled  bool          `mapstructure:"HEALTH_CHECK_ENABLED"`
	Interval time.Duration `mapstructure:"HEALTH_CHECK_INTERVAL"`
	Timeout  time.Duration `mapstructure:"HEALTH_CHECK_TIMEOUT"`
}

// Load 加载配置
func Load(envFile ...string) (*Config, error) {
	// 设置默认环境文件
	envFileName := ".env"
	if len(envFile) > 0 && envFile[0] != "" {
		envFileName = envFile[0]
	}

	// 设置配置文件名和路径
	viper.SetConfigFile(envFileName)
	viper.SetConfigType("env")
	viper.AddConfigPath(".")

	// 自动读取环境变量
	viper.AutomaticEnv()

	// 读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		// 如果配置文件不存在，只使用环境变量
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	// 创建配置实例
	cfg := &Config{}

	// 解析配置
	if err := viper.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// 设置默认值
	cfg.setDefaults()

	// 处理字符串数组环境变量
	cfg.processStringSliceEnvVars()

	return cfg, nil
}

// Validate 验证配置
func (c *Config) Validate() error {
	validate := validator.New()
	return validate.Struct(c)
}

// setDefaults 设置默认值
func (c *Config) setDefaults() {
	// 应用默认值
	if c.App.Env == "" {
		c.App.Env = "development"
	}
	if c.App.Name == "" {
		c.App.Name = "Alert Management Platform"
	}
	if c.App.Version == "" {
		c.App.Version = "1.0.0"
	}
	if c.App.Port == 0 {
		c.App.Port = 8080
	}
	if c.App.Host == "" {
		c.App.Host = "0.0.0.0"
	}
	if c.App.LogLevel == "" {
		c.App.LogLevel = "info"
	}
	if c.App.LogFormat == "" {
		c.App.LogFormat = "json"
	}
	if c.App.APIDocsPath == "" {
		c.App.APIDocsPath = "/docs"
	}

	// 数据库默认值
	if c.Database.Host == "" {
		c.Database.Host = "localhost"
	}
	if c.Database.Port == 0 {
		c.Database.Port = 5432
	}
	if c.Database.SSLMode == "" {
		c.Database.SSLMode = "disable"
	}
	if c.Database.MaxOpenConns == 0 {
		c.Database.MaxOpenConns = 25
	}
	if c.Database.MaxIdleConns == 0 {
		c.Database.MaxIdleConns = 5
	}
	if c.Database.ConnMaxLifetime == 0 {
		c.Database.ConnMaxLifetime = 5 * time.Minute
	}
	if c.Database.ConnMaxIdleTime == 0 {
		c.Database.ConnMaxIdleTime = 5 * time.Minute
	}
	if c.Database.MigrationPath == "" {
		c.Database.MigrationPath = "file://./migrations"
	}
	if c.Database.MigrationTable == "" {
		c.Database.MigrationTable = "schema_migrations"
	}

	// Redis 默认值
	if c.Redis.Host == "" {
		c.Redis.Host = "localhost"
	}
	if c.Redis.Port == 0 {
		c.Redis.Port = 6379
	}
	if c.Redis.PoolSize == 0 {
		c.Redis.PoolSize = 10
	}
	if c.Redis.MinIdleConns == 0 {
		c.Redis.MinIdleConns = 2
	}

	// JWT 默认值
	if c.JWT.AccessTokenExpire == 0 {
		c.JWT.AccessTokenExpire = 24 * time.Hour
	}
	if c.JWT.RefreshTokenExpire == 0 {
		c.JWT.RefreshTokenExpire = 7 * 24 * time.Hour
	}

	// 告警默认值
	if c.Alert.EvaluationInterval == 0 {
		c.Alert.EvaluationInterval = 30 * time.Second
	}
	if c.Alert.HistoryRetentionDays == 0 {
		c.Alert.HistoryRetentionDays = 30
	}
	if c.Alert.MaxConcurrentEvaluations == 0 {
		c.Alert.MaxConcurrentEvaluations = 10
	}

	// 性能默认值
	if c.Performance.MaxRequestSize == 0 {
		c.Performance.MaxRequestSize = 32 << 20 // 32MB
	}
	if c.Performance.MaxConcurrency == 0 {
		c.Performance.MaxConcurrency = 1000
	}
	if c.Performance.ReadTimeout == 0 {
		c.Performance.ReadTimeout = 30 * time.Second
	}
	if c.Performance.WriteTimeout == 0 {
		c.Performance.WriteTimeout = 30 * time.Second
	}
	if c.Performance.IdleTimeout == 0 {
		c.Performance.IdleTimeout = 120 * time.Second
	}
	if c.Performance.WorkerPoolSize == 0 {
		c.Performance.WorkerPoolSize = 10
	}
	if c.Performance.QueueBufferSize == 0 {
		c.Performance.QueueBufferSize = 1000
	}

	// 健康检查默认值
	if c.HealthCheck.Interval == 0 {
		c.HealthCheck.Interval = 30 * time.Second
	}
	if c.HealthCheck.Timeout == 0 {
		c.HealthCheck.Timeout = 5 * time.Second
	}

	// 安全默认值
	if len(c.Security.CORSAllowedOrigins) == 0 {
		c.Security.CORSAllowedOrigins = []string{"*"}
	}
	if len(c.Security.CORSAllowedMethods) == 0 {
		c.Security.CORSAllowedMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	}
	if len(c.Security.CORSAllowedHeaders) == 0 {
		c.Security.CORSAllowedHeaders = []string{"*"}
	}
	if c.Security.RateLimitRPS == 0 {
		c.Security.RateLimitRPS = 100
	}
	if c.Security.RateLimitBurst == 0 {
		c.Security.RateLimitBurst = 200
	}
	if c.Security.APIKeyHeader == "" {
		c.Security.APIKeyHeader = "X-API-Key"
	}

	// 文件存储默认值
	if c.FileStorage.Type == "" {
		c.FileStorage.Type = "local"
	}
	if c.FileStorage.LocalPath == "" {
		c.FileStorage.LocalPath = "./uploads"
	}
}

// processStringSliceEnvVars 处理字符串数组环境变量
func (c *Config) processStringSliceEnvVars() {
	// 处理 CORS 相关的环境变量
	if corsOrigins := os.Getenv("CORS_ALLOWED_ORIGINS"); corsOrigins != "" {
		c.Security.CORSAllowedOrigins = strings.Split(corsOrigins, ",")
	}
	if corsMethods := os.Getenv("CORS_ALLOWED_METHODS"); corsMethods != "" {
		c.Security.CORSAllowedMethods = strings.Split(corsMethods, ",")
	}
	if corsHeaders := os.Getenv("CORS_ALLOWED_HEADERS"); corsHeaders != "" {
		c.Security.CORSAllowedHeaders = strings.Split(corsHeaders, ",")
	}
}

// IsProduction 判断是否为生产环境
func (c *Config) IsProduction() bool {
	return c.App.Env == "production"
}

// IsDevelopment 判断是否为开发环境
func (c *Config) IsDevelopment() bool {
	return c.App.Env == "development"
}

// GetServerAddress 获取服务器地址
func (c *Config) GetServerAddress() string {
	return fmt.Sprintf("%s:%d", c.App.Host, c.App.Port)
}

// GetDSN 获取数据库连接字符串
func (d *DatabaseConfig) GetDSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		d.Host, d.Port, d.User, d.Password, d.Name, d.SSLMode)
}

// GetDSNWithoutPassword 获取不包含密码的数据库连接字符串（用于日志）
func (d *DatabaseConfig) GetDSNWithoutPassword() string {
	return fmt.Sprintf("host=%s port=%d user=%s dbname=%s sslmode=%s",
		d.Host, d.Port, d.User, d.Name, d.SSLMode)
}

// PostgresConfig 别名，用于兼容性
type PostgresConfig = DatabaseConfig

// MigrationConfig 迁移配置
type MigrationConfig struct {
	Path        string        `mapstructure:"MIGRATION_PATH"`
	Table       string        `mapstructure:"MIGRATION_TABLE"`
	LockTimeout time.Duration `mapstructure:"MIGRATION_LOCK_TIMEOUT"`
}