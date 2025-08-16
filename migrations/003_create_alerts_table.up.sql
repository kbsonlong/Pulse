-- 创建告警表
-- 创建时间: 2024-01-01
-- 描述: 创建告警相关表，包含告警事件、历史记录和时序数据

-- 创建告警表（主表）
CREATE TABLE alerts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    
    -- 告警基本信息
    alert_name VARCHAR(255) NOT NULL,
    alert_id VARCHAR(255) NOT NULL, -- 外部系统的告警ID
    fingerprint VARCHAR(64) NOT NULL, -- 告警指纹，用于去重
    
    -- 告警分类
    severity alert_severity NOT NULL,
    status alert_status NOT NULL DEFAULT 'firing',
    
    -- 告警内容
    summary TEXT NOT NULL,
    description TEXT,
    message TEXT,
    
    -- 标签和注解
    labels JSONB NOT NULL DEFAULT '{}',
    annotations JSONB DEFAULT '{}',
    
    -- 来源信息
    source_type VARCHAR(50) NOT NULL, -- prometheus, grafana, custom等
    source_id VARCHAR(255), -- 数据源ID
    rule_id UUID, -- 关联的规则ID
    
    -- 时间信息
    starts_at TIMESTAMP WITH TIME ZONE NOT NULL,
    ends_at TIMESTAMP WITH TIME ZONE,
    resolved_at TIMESTAMP WITH TIME ZONE,
    
    -- 处理信息
    assigned_to UUID REFERENCES users(id),
    acknowledged_by UUID REFERENCES users(id),
    acknowledged_at TIMESTAMP WITH TIME ZONE,
    acknowledged_note TEXT,
    
    -- 抑制信息
    suppressed BOOLEAN DEFAULT FALSE,
    suppressed_by UUID REFERENCES users(id),
    suppressed_at TIMESTAMP WITH TIME ZONE,
    suppressed_until TIMESTAMP WITH TIME ZONE,
    suppressed_reason TEXT,
    
    -- 升级信息
    escalated BOOLEAN DEFAULT FALSE,
    escalated_at TIMESTAMP WITH TIME ZONE,
    escalated_to UUID REFERENCES users(id),
    escalation_level INTEGER DEFAULT 0,
    
    -- 通知信息
    notification_sent BOOLEAN DEFAULT FALSE,
    notification_count INTEGER DEFAULT 0,
    last_notification_at TIMESTAMP WITH TIME ZONE,
    
    -- 关联信息
    parent_alert_id UUID REFERENCES alerts(id),
    ticket_id UUID, -- 关联的工单ID
    
    -- 审计字段
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_by UUID REFERENCES users(id),
    updated_by UUID REFERENCES users(id),
    
    -- 元数据
    metadata JSONB DEFAULT '{}',
    
    -- 约束
    CONSTRAINT alerts_fingerprint_unique UNIQUE (fingerprint),
    CONSTRAINT alerts_escalation_level_check CHECK (escalation_level >= 0),
    CONSTRAINT alerts_notification_count_check CHECK (notification_count >= 0),
    CONSTRAINT alerts_time_check CHECK (ends_at IS NULL OR ends_at >= starts_at)
);

-- 创建告警历史表（时序数据）
CREATE TABLE alert_history (
    id UUID DEFAULT uuid_generate_v4(),
    alert_id UUID NOT NULL REFERENCES alerts(id) ON DELETE CASCADE,
    
    -- 状态变更信息
    old_status alert_status,
    new_status alert_status NOT NULL,
    old_severity alert_severity,
    new_severity alert_severity,
    
    -- 变更原因
    change_type VARCHAR(50) NOT NULL, -- created, updated, resolved, acknowledged, suppressed等
    change_reason TEXT,
    
    -- 操作者信息
    changed_by UUID REFERENCES users(id),
    
    -- 变更详情
    old_values JSONB,
    new_values JSONB,
    changes JSONB,
    
    -- 时间戳
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    -- 元数据
    metadata JSONB DEFAULT '{}',
    
    -- 复合主键包含分区列
    PRIMARY KEY (id, timestamp)
);

-- 创建告警指标数据表（时序数据）
CREATE TABLE alert_metrics (
    id UUID DEFAULT uuid_generate_v4(),
    alert_id UUID NOT NULL REFERENCES alerts(id) ON DELETE CASCADE,
    
    -- 指标信息
    metric_name VARCHAR(255) NOT NULL,
    metric_value DOUBLE PRECISION NOT NULL,
    metric_unit VARCHAR(50),
    
    -- 标签
    labels JSONB DEFAULT '{}',
    
    -- 时间戳
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    -- 元数据
    metadata JSONB DEFAULT '{}',
    
    -- 复合主键包含分区列
    PRIMARY KEY (id, timestamp)
);

-- 创建告警分组表
CREATE TABLE alert_groups (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    
    -- 分组信息
    group_key VARCHAR(255) NOT NULL,
    group_name VARCHAR(255) NOT NULL,
    
    -- 分组规则
    group_by JSONB NOT NULL, -- 分组字段
    group_labels JSONB DEFAULT '{}',
    
    -- 状态信息
    status alert_status NOT NULL DEFAULT 'firing',
    severity alert_severity NOT NULL,
    
    -- 告警数量
    alert_count INTEGER DEFAULT 0,
    firing_count INTEGER DEFAULT 0,
    resolved_count INTEGER DEFAULT 0,
    
    -- 时间信息
    first_alert_at TIMESTAMP WITH TIME ZONE,
    last_alert_at TIMESTAMP WITH TIME ZONE,
    resolved_at TIMESTAMP WITH TIME ZONE,
    
    -- 处理信息
    assigned_to UUID REFERENCES users(id),
    acknowledged_by UUID REFERENCES users(id),
    acknowledged_at TIMESTAMP WITH TIME ZONE,
    
    -- 审计字段
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    -- 元数据
    metadata JSONB DEFAULT '{}',
    
    -- 约束
    CONSTRAINT alert_groups_group_key_unique UNIQUE (group_key),
    CONSTRAINT alert_groups_alert_count_check CHECK (alert_count >= 0),
    CONSTRAINT alert_groups_firing_count_check CHECK (firing_count >= 0),
    CONSTRAINT alert_groups_resolved_count_check CHECK (resolved_count >= 0)
);

-- 创建告警分组关联表
CREATE TABLE alert_group_members (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    group_id UUID NOT NULL REFERENCES alert_groups(id) ON DELETE CASCADE,
    alert_id UUID NOT NULL REFERENCES alerts(id) ON DELETE CASCADE,
    
    -- 加入时间
    joined_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    -- 唯一约束
    CONSTRAINT alert_group_members_unique UNIQUE (group_id, alert_id)
);

-- 创建告警抑制规则表
CREATE TABLE alert_silences (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    
    -- 抑制规则信息
    name VARCHAR(255) NOT NULL,
    comment TEXT,
    
    -- 匹配规则
    matchers JSONB NOT NULL, -- 匹配条件
    
    -- 时间范围
    starts_at TIMESTAMP WITH TIME ZONE NOT NULL,
    ends_at TIMESTAMP WITH TIME ZONE NOT NULL,
    
    -- 状态
    active BOOLEAN DEFAULT TRUE,
    
    -- 创建者信息
    created_by UUID NOT NULL REFERENCES users(id),
    
    -- 审计字段
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    -- 元数据
    metadata JSONB DEFAULT '{}',
    
    -- 约束
    CONSTRAINT alert_silences_time_check CHECK (ends_at > starts_at)
);

-- 创建索引

-- alerts 表索引
CREATE INDEX idx_alerts_alert_name ON alerts(alert_name);
CREATE INDEX idx_alerts_alert_id ON alerts(alert_id);
CREATE INDEX idx_alerts_fingerprint ON alerts(fingerprint);
CREATE INDEX idx_alerts_severity ON alerts(severity);
CREATE INDEX idx_alerts_status ON alerts(status);
CREATE INDEX idx_alerts_source_type ON alerts(source_type);
CREATE INDEX idx_alerts_source_id ON alerts(source_id);
CREATE INDEX idx_alerts_rule_id ON alerts(rule_id);
CREATE INDEX idx_alerts_starts_at ON alerts(starts_at);
CREATE INDEX idx_alerts_ends_at ON alerts(ends_at);
CREATE INDEX idx_alerts_resolved_at ON alerts(resolved_at);
CREATE INDEX idx_alerts_assigned_to ON alerts(assigned_to);
CREATE INDEX idx_alerts_acknowledged_by ON alerts(acknowledged_by);
CREATE INDEX idx_alerts_suppressed ON alerts(suppressed);
CREATE INDEX idx_alerts_escalated ON alerts(escalated);
CREATE INDEX idx_alerts_created_at ON alerts(created_at);
CREATE INDEX idx_alerts_updated_at ON alerts(updated_at);

-- GIN 索引用于 JSONB 字段
CREATE INDEX idx_alerts_labels_gin ON alerts USING GIN(labels);
CREATE INDEX idx_alerts_annotations_gin ON alerts USING GIN(annotations);
CREATE INDEX idx_alerts_metadata_gin ON alerts USING GIN(metadata);

-- 复合索引
CREATE INDEX idx_alerts_status_severity ON alerts(status, severity);
CREATE INDEX idx_alerts_source_status ON alerts(source_type, status);
CREATE INDEX idx_alerts_starts_at_status ON alerts(starts_at, status);

-- alert_history 表索引
CREATE INDEX idx_alert_history_alert_id ON alert_history(alert_id);
CREATE INDEX idx_alert_history_timestamp ON alert_history(timestamp);
CREATE INDEX idx_alert_history_change_type ON alert_history(change_type);
CREATE INDEX idx_alert_history_changed_by ON alert_history(changed_by);
CREATE INDEX idx_alert_history_new_status ON alert_history(new_status);

-- alert_metrics 表索引
CREATE INDEX idx_alert_metrics_alert_id ON alert_metrics(alert_id);
CREATE INDEX idx_alert_metrics_timestamp ON alert_metrics(timestamp);
CREATE INDEX idx_alert_metrics_metric_name ON alert_metrics(metric_name);
CREATE INDEX idx_alert_metrics_metric_value ON alert_metrics(metric_value);

-- GIN 索引
CREATE INDEX idx_alert_metrics_labels_gin ON alert_metrics USING GIN(labels);

-- alert_groups 表索引
CREATE INDEX idx_alert_groups_group_key ON alert_groups(group_key);
CREATE INDEX idx_alert_groups_status ON alert_groups(status);
CREATE INDEX idx_alert_groups_severity ON alert_groups(severity);
CREATE INDEX idx_alert_groups_assigned_to ON alert_groups(assigned_to);
CREATE INDEX idx_alert_groups_created_at ON alert_groups(created_at);

-- alert_group_members 表索引
CREATE INDEX idx_alert_group_members_group_id ON alert_group_members(group_id);
CREATE INDEX idx_alert_group_members_alert_id ON alert_group_members(alert_id);
CREATE INDEX idx_alert_group_members_joined_at ON alert_group_members(joined_at);

-- alert_silences 表索引
CREATE INDEX idx_alert_silences_name ON alert_silences(name);
CREATE INDEX idx_alert_silences_active ON alert_silences(active);
CREATE INDEX idx_alert_silences_starts_at ON alert_silences(starts_at);
CREATE INDEX idx_alert_silences_ends_at ON alert_silences(ends_at);
CREATE INDEX idx_alert_silences_created_by ON alert_silences(created_by);

-- GIN 索引
CREATE INDEX idx_alert_silences_matchers_gin ON alert_silences USING GIN(matchers);

-- 创建更新时间触发器
CREATE TRIGGER update_alerts_updated_at
    BEFORE UPDATE ON alerts
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_alert_groups_updated_at
    BEFORE UPDATE ON alert_groups
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_alert_silences_updated_at
    BEFORE UPDATE ON alert_silences
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- 添加表注释
COMMENT ON TABLE alerts IS '告警表，存储告警事件的详细信息';
COMMENT ON TABLE alert_history IS '告警历史表，记录告警状态变更历史';
COMMENT ON TABLE alert_metrics IS '告警指标数据表，存储告警相关的时序指标数据';
COMMENT ON TABLE alert_groups IS '告警分组表，用于告警聚合和批量处理';
COMMENT ON TABLE alert_group_members IS '告警分组成员表，记录告警与分组的关联关系';
COMMENT ON TABLE alert_silences IS '告警抑制规则表，定义告警抑制条件和时间范围';

-- 添加列注释
COMMENT ON COLUMN alerts.fingerprint IS '告警指纹，用于告警去重和识别';
COMMENT ON COLUMN alerts.labels IS '告警标签，JSON格式存储键值对';
COMMENT ON COLUMN alerts.annotations IS '告警注解，JSON格式存储额外信息';
COMMENT ON COLUMN alert_history.change_type IS '变更类型：created, updated, resolved, acknowledged, suppressed等';
COMMENT ON COLUMN alert_silences.matchers IS '抑制规则匹配条件，JSON格式';