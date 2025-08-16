-- 创建数据源表
-- 创建时间: 2024-01-01
-- 描述: 创建数据源相关表，包含数据源配置、连接信息和健康状态

-- 创建数据源表
CREATE TABLE data_sources (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    
    -- 数据源基本信息
    name VARCHAR(255) NOT NULL,
    display_name VARCHAR(255),
    description TEXT,
    
    -- 数据源类型和配置
    type datasource_type NOT NULL,
    version VARCHAR(50),
    
    -- 连接配置
    url VARCHAR(1000) NOT NULL,
    username VARCHAR(255),
    password_encrypted TEXT, -- 加密存储的密码
    
    -- 认证配置
    auth_type VARCHAR(50) DEFAULT 'basic', -- basic, token, oauth, certificate等
    auth_config JSONB DEFAULT '{}', -- 认证相关配置
    
    -- TLS配置
    tls_enabled BOOLEAN DEFAULT FALSE,
    tls_skip_verify BOOLEAN DEFAULT FALSE,
    tls_cert_file VARCHAR(500),
    tls_key_file VARCHAR(500),
    tls_ca_file VARCHAR(500),
    
    -- 连接配置
    timeout_seconds INTEGER DEFAULT 30,
    max_connections INTEGER DEFAULT 10,
    connection_pool_size INTEGER DEFAULT 5,
    
    -- 查询配置
    query_timeout_seconds INTEGER DEFAULT 60,
    max_query_length INTEGER DEFAULT 10000,
    
    -- 状态信息
    status datasource_status NOT NULL DEFAULT 'active',
    enabled BOOLEAN DEFAULT TRUE,
    
    -- 健康检查
    health_check_enabled BOOLEAN DEFAULT TRUE,
    health_check_interval_seconds INTEGER DEFAULT 60,
    health_check_timeout_seconds INTEGER DEFAULT 10,
    health_check_query TEXT,
    last_health_check_at TIMESTAMP WITH TIME ZONE,
    last_health_check_status VARCHAR(20), -- healthy, unhealthy, unknown
    last_health_check_error TEXT,
    consecutive_failures INTEGER DEFAULT 0,
    
    -- 性能统计
    total_queries BIGINT DEFAULT 0,
    successful_queries BIGINT DEFAULT 0,
    failed_queries BIGINT DEFAULT 0,
    avg_query_duration_ms DOUBLE PRECISION DEFAULT 0,
    last_query_at TIMESTAMP WITH TIME ZONE,
    
    -- 标签和注解
    labels JSONB DEFAULT '{}',
    annotations JSONB DEFAULT '{}',
    
    -- 审计字段
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_by UUID REFERENCES users(id),
    updated_by UUID REFERENCES users(id),
    
    -- 元数据
    metadata JSONB DEFAULT '{}',
    
    -- 约束
    CONSTRAINT data_sources_name_unique UNIQUE (name),
    CONSTRAINT data_sources_timeout_check CHECK (timeout_seconds > 0),
    CONSTRAINT data_sources_max_connections_check CHECK (max_connections > 0),
    CONSTRAINT data_sources_pool_size_check CHECK (connection_pool_size > 0),
    CONSTRAINT data_sources_query_timeout_check CHECK (query_timeout_seconds > 0),
    CONSTRAINT data_sources_health_interval_check CHECK (health_check_interval_seconds > 0),
    CONSTRAINT data_sources_health_timeout_check CHECK (health_check_timeout_seconds > 0),
    CONSTRAINT data_sources_consecutive_failures_check CHECK (consecutive_failures >= 0),
    CONSTRAINT data_sources_queries_check CHECK (total_queries >= 0 AND successful_queries >= 0 AND failed_queries >= 0),
    CONSTRAINT data_sources_avg_duration_check CHECK (avg_query_duration_ms >= 0)
);

-- 创建数据源健康检查历史表（时序数据）
CREATE TABLE data_source_health_checks (
    id UUID DEFAULT uuid_generate_v4(),
    data_source_id UUID NOT NULL REFERENCES data_sources(id) ON DELETE CASCADE,
    
    -- 检查信息
    check_time TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    duration_ms INTEGER NOT NULL,
    
    -- 检查结果
    status VARCHAR(20) NOT NULL, -- healthy, unhealthy, timeout, error
    success BOOLEAN NOT NULL,
    error_message TEXT,
    
    -- 响应信息
    response_code INTEGER,
    response_size INTEGER,
    response_time_ms INTEGER,
    
    -- 元数据
    metadata JSONB DEFAULT '{}',
    
    -- 复合主键包含分区列
    PRIMARY KEY (id, check_time),
    
    -- 约束
    CONSTRAINT data_source_health_checks_duration_check CHECK (duration_ms >= 0),
    CONSTRAINT data_source_health_checks_response_time_check CHECK (response_time_ms IS NULL OR response_time_ms >= 0),
    CONSTRAINT data_source_health_checks_response_size_check CHECK (response_size IS NULL OR response_size >= 0)
);

-- 创建数据源查询历史表（时序数据）
CREATE TABLE data_source_queries (
    id UUID DEFAULT uuid_generate_v4(),
    data_source_id UUID NOT NULL REFERENCES data_sources(id) ON DELETE CASCADE,
    
    -- 查询信息
    query_time TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    query_text TEXT NOT NULL,
    query_hash VARCHAR(64), -- 查询哈希，用于去重统计
    
    -- 执行信息
    duration_ms INTEGER NOT NULL,
    success BOOLEAN NOT NULL,
    error_message TEXT,
    
    -- 结果信息
    result_count INTEGER,
    result_size INTEGER,
    
    -- 请求者信息
    user_id UUID REFERENCES users(id),
    source_component VARCHAR(100), -- 发起查询的组件
    
    -- 元数据
    metadata JSONB DEFAULT '{}',
    
    -- 复合主键包含分区列
    PRIMARY KEY (id, query_time),
    
    -- 约束
    CONSTRAINT data_source_queries_duration_check CHECK (duration_ms >= 0),
    CONSTRAINT data_source_queries_result_count_check CHECK (result_count IS NULL OR result_count >= 0),
    CONSTRAINT data_source_queries_result_size_check CHECK (result_size IS NULL OR result_size >= 0)
);

-- 创建数据源配置模板表
CREATE TABLE data_source_templates (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    
    -- 模板基本信息
    name VARCHAR(255) NOT NULL,
    display_name VARCHAR(255),
    description TEXT,
    category VARCHAR(100), -- 模板分类
    
    -- 模板配置
    type datasource_type NOT NULL,
    template_config JSONB NOT NULL, -- 模板配置
    default_values JSONB DEFAULT '{}', -- 默认值
    required_fields JSONB DEFAULT '[]', -- 必填字段
    
    -- 状态信息
    enabled BOOLEAN DEFAULT TRUE,
    public BOOLEAN DEFAULT FALSE, -- 是否公开模板
    
    -- 使用统计
    usage_count INTEGER DEFAULT 0,
    
    -- 审计字段
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_by UUID REFERENCES users(id),
    updated_by UUID REFERENCES users(id),
    
    -- 元数据
    metadata JSONB DEFAULT '{}',
    
    -- 约束
    CONSTRAINT data_source_templates_name_unique UNIQUE (name),
    CONSTRAINT data_source_templates_usage_count_check CHECK (usage_count >= 0)
);

-- 创建数据源标签表（用于快速标签查询）
CREATE TABLE data_source_labels (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    data_source_id UUID NOT NULL REFERENCES data_sources(id) ON DELETE CASCADE,
    
    -- 标签信息
    label_key VARCHAR(255) NOT NULL,
    label_value VARCHAR(255) NOT NULL,
    
    -- 审计字段
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    -- 约束
    CONSTRAINT data_source_labels_unique UNIQUE (data_source_id, label_key, label_value)
);

-- 创建数据源访问控制表
CREATE TABLE data_source_permissions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    data_source_id UUID NOT NULL REFERENCES data_sources(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    
    -- 权限类型
    permission_type VARCHAR(50) NOT NULL, -- read, write, admin
    
    -- 权限范围
    scope JSONB DEFAULT '{}', -- 权限范围限制
    
    -- 审计字段
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_by UUID REFERENCES users(id),
    
    -- 约束
    CONSTRAINT data_source_permissions_unique UNIQUE (data_source_id, user_id, permission_type)
);

-- 创建索引

-- data_sources 表索引
CREATE INDEX idx_data_sources_name ON data_sources(name);
CREATE INDEX idx_data_sources_type ON data_sources(type);
CREATE INDEX idx_data_sources_status ON data_sources(status);
CREATE INDEX idx_data_sources_enabled ON data_sources(enabled);
CREATE INDEX idx_data_sources_last_health_check_at ON data_sources(last_health_check_at);
CREATE INDEX idx_data_sources_last_health_check_status ON data_sources(last_health_check_status);
CREATE INDEX idx_data_sources_consecutive_failures ON data_sources(consecutive_failures);
CREATE INDEX idx_data_sources_last_query_at ON data_sources(last_query_at);
CREATE INDEX idx_data_sources_created_at ON data_sources(created_at);
CREATE INDEX idx_data_sources_updated_at ON data_sources(updated_at);

-- GIN 索引
CREATE INDEX idx_data_sources_labels_gin ON data_sources USING GIN(labels);
CREATE INDEX idx_data_sources_annotations_gin ON data_sources USING GIN(annotations);
CREATE INDEX idx_data_sources_auth_config_gin ON data_sources USING GIN(auth_config);

-- 复合索引
CREATE INDEX idx_data_sources_type_status ON data_sources(type, status);
CREATE INDEX idx_data_sources_enabled_status ON data_sources(enabled, status);
CREATE INDEX idx_data_sources_health_status_time ON data_sources(last_health_check_status, last_health_check_at);

-- data_source_health_checks 表索引
CREATE INDEX idx_data_source_health_checks_data_source_id ON data_source_health_checks(data_source_id);
CREATE INDEX idx_data_source_health_checks_check_time ON data_source_health_checks(check_time);
CREATE INDEX idx_data_source_health_checks_status ON data_source_health_checks(status);
CREATE INDEX idx_data_source_health_checks_success ON data_source_health_checks(success);

-- 复合索引
CREATE INDEX idx_data_source_health_checks_ds_time ON data_source_health_checks(data_source_id, check_time);
CREATE INDEX idx_data_source_health_checks_ds_status ON data_source_health_checks(data_source_id, status);

-- data_source_queries 表索引
CREATE INDEX idx_data_source_queries_data_source_id ON data_source_queries(data_source_id);
CREATE INDEX idx_data_source_queries_query_time ON data_source_queries(query_time);
CREATE INDEX idx_data_source_queries_query_hash ON data_source_queries(query_hash);
CREATE INDEX idx_data_source_queries_success ON data_source_queries(success);
CREATE INDEX idx_data_source_queries_user_id ON data_source_queries(user_id);
CREATE INDEX idx_data_source_queries_source_component ON data_source_queries(source_component);
CREATE INDEX idx_data_source_queries_duration_ms ON data_source_queries(duration_ms);

-- 复合索引
CREATE INDEX idx_data_source_queries_ds_time ON data_source_queries(data_source_id, query_time);
CREATE INDEX idx_data_source_queries_ds_success ON data_source_queries(data_source_id, success);
CREATE INDEX idx_data_source_queries_user_time ON data_source_queries(user_id, query_time);

-- data_source_templates 表索引
CREATE INDEX idx_data_source_templates_name ON data_source_templates(name);
CREATE INDEX idx_data_source_templates_type ON data_source_templates(type);
CREATE INDEX idx_data_source_templates_category ON data_source_templates(category);
CREATE INDEX idx_data_source_templates_enabled ON data_source_templates(enabled);
CREATE INDEX idx_data_source_templates_public ON data_source_templates(public);
CREATE INDEX idx_data_source_templates_usage_count ON data_source_templates(usage_count);
CREATE INDEX idx_data_source_templates_created_at ON data_source_templates(created_at);

-- data_source_labels 表索引
CREATE INDEX idx_data_source_labels_data_source_id ON data_source_labels(data_source_id);
CREATE INDEX idx_data_source_labels_key ON data_source_labels(label_key);
CREATE INDEX idx_data_source_labels_value ON data_source_labels(label_value);
CREATE INDEX idx_data_source_labels_key_value ON data_source_labels(label_key, label_value);

-- data_source_permissions 表索引
CREATE INDEX idx_data_source_permissions_data_source_id ON data_source_permissions(data_source_id);
CREATE INDEX idx_data_source_permissions_user_id ON data_source_permissions(user_id);
CREATE INDEX idx_data_source_permissions_type ON data_source_permissions(permission_type);

-- 创建更新时间触发器
CREATE TRIGGER update_data_sources_updated_at
    BEFORE UPDATE ON data_sources
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_data_source_templates_updated_at
    BEFORE UPDATE ON data_source_templates
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- 添加表注释
COMMENT ON TABLE data_sources IS '数据源表，存储各种数据源的配置和连接信息';
COMMENT ON TABLE data_source_health_checks IS '数据源健康检查历史表，记录健康检查结果';
COMMENT ON TABLE data_source_queries IS '数据源查询历史表，记录查询执行情况';
COMMENT ON TABLE data_source_templates IS '数据源模板表，提供预定义的数据源配置模板';
COMMENT ON TABLE data_source_labels IS '数据源标签表，用于快速标签查询和过滤';
COMMENT ON TABLE data_source_permissions IS '数据源权限表，控制用户对数据源的访问权限';

-- 添加列注释
COMMENT ON COLUMN data_sources.password_encrypted IS '加密存储的密码，使用AES加密';
COMMENT ON COLUMN data_sources.auth_config IS '认证配置，JSON格式存储各种认证方式的配置';
COMMENT ON COLUMN data_sources.consecutive_failures IS '连续失败次数，用于健康检查和熔断';
COMMENT ON COLUMN data_source_templates.template_config IS '数据源模板配置，JSON格式定义';
COMMENT ON COLUMN data_source_templates.required_fields IS '必填字段列表，JSON数组格式';
COMMENT ON COLUMN data_source_permissions.scope IS '权限范围限制，JSON格式定义具体的权限范围';