-- 创建规则表
-- 创建时间: 2024-01-01
-- 描述: 创建告警规则相关表，包含规则定义、规则组和执行历史

-- 创建规则组表
CREATE TABLE rule_groups (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    
    -- 规则组基本信息
    name VARCHAR(255) NOT NULL,
    display_name VARCHAR(255),
    description TEXT,
    
    -- 规则组配置
    interval_seconds INTEGER NOT NULL DEFAULT 60, -- 评估间隔（秒）
    evaluation_timeout_seconds INTEGER DEFAULT 30, -- 评估超时（秒）
    
    -- 数据源信息
    data_source_id UUID, -- 关联的数据源ID
    
    -- 状态信息
    status rule_status NOT NULL DEFAULT 'active',
    enabled BOOLEAN DEFAULT TRUE,
    
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
    CONSTRAINT rule_groups_name_unique UNIQUE (name),
    CONSTRAINT rule_groups_interval_check CHECK (interval_seconds > 0),
    CONSTRAINT rule_groups_timeout_check CHECK (evaluation_timeout_seconds > 0)
);

-- 创建规则表
CREATE TABLE rules (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    
    -- 规则基本信息
    name VARCHAR(255) NOT NULL,
    display_name VARCHAR(255),
    description TEXT,
    
    -- 规则分组
    group_id UUID NOT NULL REFERENCES rule_groups(id) ON DELETE CASCADE,
    
    -- 规则类型和配置
    rule_type VARCHAR(50) NOT NULL DEFAULT 'prometheus', -- prometheus, custom, threshold等
    query TEXT NOT NULL, -- 查询表达式
    condition TEXT, -- 条件表达式
    
    -- 告警配置
    severity alert_severity NOT NULL DEFAULT 'warning',
    for_duration INTEGER DEFAULT 0, -- 持续时间（秒）
    
    -- 标签和注解
    labels JSONB DEFAULT '{}',
    annotations JSONB DEFAULT '{}',
    
    -- 状态信息
    status rule_status NOT NULL DEFAULT 'active',
    enabled BOOLEAN DEFAULT TRUE,
    
    -- 执行统计
    last_evaluation_at TIMESTAMP WITH TIME ZONE,
    last_evaluation_duration_ms INTEGER,
    evaluation_count BIGINT DEFAULT 0,
    error_count BIGINT DEFAULT 0,
    last_error TEXT,
    last_error_at TIMESTAMP WITH TIME ZONE,
    
    -- 告警统计
    alert_count BIGINT DEFAULT 0,
    last_alert_at TIMESTAMP WITH TIME ZONE,
    
    -- 数据源信息
    data_source_id UUID, -- 关联的数据源ID
    
    -- 审计字段
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_by UUID REFERENCES users(id),
    updated_by UUID REFERENCES users(id),
    
    -- 元数据
    metadata JSONB DEFAULT '{}',
    
    -- 约束
    CONSTRAINT rules_name_group_unique UNIQUE (name, group_id),
    CONSTRAINT rules_for_duration_check CHECK (for_duration >= 0),
    CONSTRAINT rules_evaluation_count_check CHECK (evaluation_count >= 0),
    CONSTRAINT rules_error_count_check CHECK (error_count >= 0),
    CONSTRAINT rules_alert_count_check CHECK (alert_count >= 0)
);

-- 创建规则执行历史表（时序数据）
CREATE TABLE rule_evaluations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    rule_id UUID NOT NULL REFERENCES rules(id) ON DELETE CASCADE,
    
    -- 执行信息
    evaluation_time TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    duration_ms INTEGER NOT NULL,
    
    -- 执行结果
    success BOOLEAN NOT NULL,
    error_message TEXT,
    
    -- 查询结果
    query_result JSONB,
    samples_count INTEGER DEFAULT 0,
    
    -- 告警结果
    alerts_fired INTEGER DEFAULT 0,
    alerts_resolved INTEGER DEFAULT 0,
    
    -- 元数据
    metadata JSONB DEFAULT '{}',
    
    -- 约束
    CONSTRAINT rule_evaluations_duration_check CHECK (duration_ms >= 0),
    CONSTRAINT rule_evaluations_samples_check CHECK (samples_count >= 0),
    CONSTRAINT rule_evaluations_alerts_fired_check CHECK (alerts_fired >= 0),
    CONSTRAINT rule_evaluations_alerts_resolved_check CHECK (alerts_resolved >= 0)
);

-- 创建规则模板表
CREATE TABLE rule_templates (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    
    -- 模板基本信息
    name VARCHAR(255) NOT NULL,
    display_name VARCHAR(255),
    description TEXT,
    category VARCHAR(100), -- 模板分类
    
    -- 模板内容
    template_content JSONB NOT NULL, -- 模板定义
    variables JSONB DEFAULT '{}', -- 模板变量
    
    -- 模板配置
    rule_type VARCHAR(50) NOT NULL,
    default_severity alert_severity DEFAULT 'warning',
    
    -- 标签和注解
    labels JSONB DEFAULT '{}',
    annotations JSONB DEFAULT '{}',
    
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
    CONSTRAINT rule_templates_name_unique UNIQUE (name),
    CONSTRAINT rule_templates_usage_count_check CHECK (usage_count >= 0)
);

-- 创建规则依赖表
CREATE TABLE rule_dependencies (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    rule_id UUID NOT NULL REFERENCES rules(id) ON DELETE CASCADE,
    depends_on_rule_id UUID NOT NULL REFERENCES rules(id) ON DELETE CASCADE,
    
    -- 依赖类型
    dependency_type VARCHAR(50) NOT NULL DEFAULT 'blocking', -- blocking, informational
    
    -- 依赖条件
    condition TEXT,
    
    -- 审计字段
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_by UUID REFERENCES users(id),
    
    -- 约束
    CONSTRAINT rule_dependencies_unique UNIQUE (rule_id, depends_on_rule_id),
    CONSTRAINT rule_dependencies_no_self_ref CHECK (rule_id != depends_on_rule_id)
);

-- 创建规则标签表（用于快速标签查询）
CREATE TABLE rule_labels (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    rule_id UUID NOT NULL REFERENCES rules(id) ON DELETE CASCADE,
    
    -- 标签信息
    label_key VARCHAR(255) NOT NULL,
    label_value VARCHAR(255) NOT NULL,
    
    -- 审计字段
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    -- 约束
    CONSTRAINT rule_labels_unique UNIQUE (rule_id, label_key, label_value)
);

-- 创建索引

-- rule_groups 表索引
CREATE INDEX idx_rule_groups_name ON rule_groups(name);
CREATE INDEX idx_rule_groups_status ON rule_groups(status);
CREATE INDEX idx_rule_groups_enabled ON rule_groups(enabled);
CREATE INDEX idx_rule_groups_data_source_id ON rule_groups(data_source_id);
CREATE INDEX idx_rule_groups_created_at ON rule_groups(created_at);
CREATE INDEX idx_rule_groups_updated_at ON rule_groups(updated_at);

-- GIN 索引
CREATE INDEX idx_rule_groups_labels_gin ON rule_groups USING GIN(labels);
CREATE INDEX idx_rule_groups_annotations_gin ON rule_groups USING GIN(annotations);

-- rules 表索引
CREATE INDEX idx_rules_name ON rules(name);
CREATE INDEX idx_rules_group_id ON rules(group_id);
CREATE INDEX idx_rules_rule_type ON rules(rule_type);
CREATE INDEX idx_rules_severity ON rules(severity);
CREATE INDEX idx_rules_status ON rules(status);
CREATE INDEX idx_rules_enabled ON rules(enabled);
CREATE INDEX idx_rules_data_source_id ON rules(data_source_id);
CREATE INDEX idx_rules_last_evaluation_at ON rules(last_evaluation_at);
CREATE INDEX idx_rules_last_alert_at ON rules(last_alert_at);
CREATE INDEX idx_rules_created_at ON rules(created_at);
CREATE INDEX idx_rules_updated_at ON rules(updated_at);

-- GIN 索引
CREATE INDEX idx_rules_labels_gin ON rules USING GIN(labels);
CREATE INDEX idx_rules_annotations_gin ON rules USING GIN(annotations);

-- 复合索引
CREATE INDEX idx_rules_group_status ON rules(group_id, status);
CREATE INDEX idx_rules_enabled_status ON rules(enabled, status);
CREATE INDEX idx_rules_severity_status ON rules(severity, status);

-- rule_evaluations 表索引
CREATE INDEX idx_rule_evaluations_rule_id ON rule_evaluations(rule_id);
CREATE INDEX idx_rule_evaluations_evaluation_time ON rule_evaluations(evaluation_time);
CREATE INDEX idx_rule_evaluations_success ON rule_evaluations(success);
CREATE INDEX idx_rule_evaluations_duration_ms ON rule_evaluations(duration_ms);

-- 复合索引
CREATE INDEX idx_rule_evaluations_rule_time ON rule_evaluations(rule_id, evaluation_time);
CREATE INDEX idx_rule_evaluations_rule_success ON rule_evaluations(rule_id, success);

-- rule_templates 表索引
CREATE INDEX idx_rule_templates_name ON rule_templates(name);
CREATE INDEX idx_rule_templates_category ON rule_templates(category);
CREATE INDEX idx_rule_templates_rule_type ON rule_templates(rule_type);
CREATE INDEX idx_rule_templates_enabled ON rule_templates(enabled);
CREATE INDEX idx_rule_templates_public ON rule_templates(public);
CREATE INDEX idx_rule_templates_usage_count ON rule_templates(usage_count);
CREATE INDEX idx_rule_templates_created_at ON rule_templates(created_at);

-- rule_dependencies 表索引
CREATE INDEX idx_rule_dependencies_rule_id ON rule_dependencies(rule_id);
CREATE INDEX idx_rule_dependencies_depends_on ON rule_dependencies(depends_on_rule_id);
CREATE INDEX idx_rule_dependencies_type ON rule_dependencies(dependency_type);

-- rule_labels 表索引
CREATE INDEX idx_rule_labels_rule_id ON rule_labels(rule_id);
CREATE INDEX idx_rule_labels_key ON rule_labels(label_key);
CREATE INDEX idx_rule_labels_value ON rule_labels(label_value);
CREATE INDEX idx_rule_labels_key_value ON rule_labels(label_key, label_value);

-- 创建更新时间触发器
CREATE TRIGGER update_rule_groups_updated_at
    BEFORE UPDATE ON rule_groups
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_rules_updated_at
    BEFORE UPDATE ON rules
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_rule_templates_updated_at
    BEFORE UPDATE ON rule_templates
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- 添加表注释
COMMENT ON TABLE rule_groups IS '规则组表，用于组织和管理告警规则';
COMMENT ON TABLE rules IS '告警规则表，定义告警规则的查询条件和配置';
COMMENT ON TABLE rule_evaluations IS '规则执行历史表，记录规则评估的执行结果';
COMMENT ON TABLE rule_templates IS '规则模板表，提供预定义的规则模板';
COMMENT ON TABLE rule_dependencies IS '规则依赖表，定义规则之间的依赖关系';
COMMENT ON TABLE rule_labels IS '规则标签表，用于快速标签查询和过滤';

-- 添加列注释
COMMENT ON COLUMN rules.query IS '规则查询表达式，如PromQL查询';
COMMENT ON COLUMN rules.condition IS '告警触发条件表达式';
COMMENT ON COLUMN rules.for_duration IS '告警持续时间阈值（秒）';
COMMENT ON COLUMN rule_templates.template_content IS '规则模板内容，JSON格式定义';
COMMENT ON COLUMN rule_templates.variables IS '模板变量定义，JSON格式';
COMMENT ON COLUMN rule_dependencies.dependency_type IS '依赖类型：blocking（阻塞）、informational（信息性）';