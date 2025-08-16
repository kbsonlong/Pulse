-- 创建工单表
-- 创建时间: 2024-01-01
-- 描述: 创建工单相关表，包含工单管理、状态跟踪和处理流程

-- 创建工单表
CREATE TABLE tickets (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    
    -- 工单基本信息
    ticket_number VARCHAR(50) NOT NULL, -- 工单编号，如 TK-2024-001
    title VARCHAR(500) NOT NULL,
    description TEXT,
    
    -- 工单分类
    category VARCHAR(100), -- incident, request, problem, change等
    subcategory VARCHAR(100),
    
    -- 优先级和状态
    priority ticket_priority NOT NULL DEFAULT 'medium',
    status ticket_status NOT NULL DEFAULT 'open',
    
    -- 关联信息
    alert_id UUID REFERENCES alerts(id), -- 关联的告警
    parent_ticket_id UUID REFERENCES tickets(id), -- 父工单
    
    -- 处理人员
    reporter_id UUID NOT NULL REFERENCES users(id), -- 报告人
    assignee_id UUID REFERENCES users(id), -- 处理人
    team_id UUID, -- 处理团队ID
    
    -- 时间信息
    reported_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    assigned_at TIMESTAMP WITH TIME ZONE,
    started_at TIMESTAMP WITH TIME ZONE,
    resolved_at TIMESTAMP WITH TIME ZONE,
    closed_at TIMESTAMP WITH TIME ZONE,
    
    -- SLA信息
    sla_level VARCHAR(50), -- P1, P2, P3, P4等
    response_sla_minutes INTEGER, -- 响应SLA（分钟）
    resolution_sla_minutes INTEGER, -- 解决SLA（分钟）
    response_deadline TIMESTAMP WITH TIME ZONE,
    resolution_deadline TIMESTAMP WITH TIME ZONE,
    sla_breached BOOLEAN DEFAULT FALSE,
    
    -- 工作量统计
    estimated_hours DOUBLE PRECISION,
    actual_hours DOUBLE PRECISION DEFAULT 0,
    
    -- 解决信息
    resolution TEXT,
    resolution_category VARCHAR(100),
    root_cause TEXT,
    
    -- 标签和注解
    labels JSONB DEFAULT '{}',
    tags VARCHAR(255)[], -- 标签数组
    
    -- 自定义字段
    custom_fields JSONB DEFAULT '{}',
    
    -- 审计字段
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_by UUID REFERENCES users(id),
    updated_by UUID REFERENCES users(id),
    
    -- 元数据
    metadata JSONB DEFAULT '{}',
    
    -- 约束
    CONSTRAINT tickets_ticket_number_unique UNIQUE (ticket_number),
    CONSTRAINT tickets_estimated_hours_check CHECK (estimated_hours IS NULL OR estimated_hours >= 0),
    CONSTRAINT tickets_actual_hours_check CHECK (actual_hours >= 0),
    CONSTRAINT tickets_response_sla_check CHECK (response_sla_minutes IS NULL OR response_sla_minutes > 0),
    CONSTRAINT tickets_resolution_sla_check CHECK (resolution_sla_minutes IS NULL OR resolution_sla_minutes > 0),
    CONSTRAINT tickets_time_check CHECK (
        (assigned_at IS NULL OR assigned_at >= reported_at) AND
        (started_at IS NULL OR started_at >= reported_at) AND
        (resolved_at IS NULL OR resolved_at >= reported_at) AND
        (closed_at IS NULL OR closed_at >= reported_at)
    )
);

-- 创建工单状态历史表（时序数据）
CREATE TABLE ticket_status_history (
    id UUID DEFAULT uuid_generate_v4(),
    ticket_id UUID NOT NULL REFERENCES tickets(id) ON DELETE CASCADE,
    
    -- 状态变更信息
    old_status ticket_status,
    new_status ticket_status NOT NULL,
    old_assignee_id UUID REFERENCES users(id),
    new_assignee_id UUID REFERENCES users(id),
    
    -- 变更原因
    change_reason TEXT,
    comment TEXT,
    
    -- 操作者信息
    changed_by UUID NOT NULL REFERENCES users(id),
    
    -- 时间戳
    changed_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    -- 元数据
    metadata JSONB DEFAULT '{}',
    
    -- 复合主键包含分区列
    PRIMARY KEY (id, changed_at)
);

-- 创建工单评论表
CREATE TABLE ticket_comments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    ticket_id UUID NOT NULL REFERENCES tickets(id) ON DELETE CASCADE,
    
    -- 评论内容
    content TEXT NOT NULL,
    content_type VARCHAR(20) DEFAULT 'text', -- text, markdown, html
    
    -- 评论类型
    comment_type VARCHAR(50) DEFAULT 'comment', -- comment, internal_note, system_update
    
    -- 可见性
    is_internal BOOLEAN DEFAULT FALSE, -- 是否内部评论
    is_system BOOLEAN DEFAULT FALSE, -- 是否系统评论
    
    -- 作者信息
    author_id UUID NOT NULL REFERENCES users(id),
    
    -- 回复信息
    parent_comment_id UUID REFERENCES ticket_comments(id),
    
    -- 附件信息
    attachments JSONB DEFAULT '[]', -- 附件列表
    
    -- 审计字段
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    -- 元数据
    metadata JSONB DEFAULT '{}'
);

-- 创建工单附件表
CREATE TABLE ticket_attachments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    ticket_id UUID NOT NULL REFERENCES tickets(id) ON DELETE CASCADE,
    comment_id UUID REFERENCES ticket_comments(id) ON DELETE CASCADE,
    
    -- 文件信息
    filename VARCHAR(255) NOT NULL,
    original_filename VARCHAR(255) NOT NULL,
    file_path VARCHAR(1000) NOT NULL,
    file_size BIGINT NOT NULL,
    mime_type VARCHAR(100),
    
    -- 文件哈希
    file_hash VARCHAR(64), -- SHA256哈希
    
    -- 上传信息
    uploaded_by UUID NOT NULL REFERENCES users(id),
    uploaded_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    -- 元数据
    metadata JSONB DEFAULT '{}',
    
    -- 约束
    CONSTRAINT ticket_attachments_file_size_check CHECK (file_size > 0)
);

-- 创建工单工作日志表
CREATE TABLE ticket_work_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    ticket_id UUID NOT NULL REFERENCES tickets(id) ON DELETE CASCADE,
    
    -- 工作信息
    description TEXT NOT NULL,
    work_type VARCHAR(50) DEFAULT 'investigation', -- investigation, development, testing, documentation等
    
    -- 时间信息
    start_time TIMESTAMP WITH TIME ZONE NOT NULL,
    end_time TIMESTAMP WITH TIME ZONE,
    duration_minutes INTEGER, -- 工作时长（分钟）
    
    -- 工作者信息
    worker_id UUID NOT NULL REFERENCES users(id),
    
    -- 是否计费
    billable BOOLEAN DEFAULT TRUE,
    
    -- 审计字段
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    -- 元数据
    metadata JSONB DEFAULT '{}',
    
    -- 约束
    CONSTRAINT ticket_work_logs_time_check CHECK (end_time IS NULL OR end_time >= start_time),
    CONSTRAINT ticket_work_logs_duration_check CHECK (duration_minutes IS NULL OR duration_minutes > 0)
);

-- 创建工单关联表
CREATE TABLE ticket_relationships (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    source_ticket_id UUID NOT NULL REFERENCES tickets(id) ON DELETE CASCADE,
    target_ticket_id UUID NOT NULL REFERENCES tickets(id) ON DELETE CASCADE,
    
    -- 关联类型
    relationship_type VARCHAR(50) NOT NULL, -- blocks, blocked_by, duplicates, related_to, caused_by等
    
    -- 关联描述
    description TEXT,
    
    -- 创建者信息
    created_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    -- 约束
    CONSTRAINT ticket_relationships_unique UNIQUE (source_ticket_id, target_ticket_id, relationship_type),
    CONSTRAINT ticket_relationships_no_self_ref CHECK (source_ticket_id != target_ticket_id)
);

-- 创建工单SLA跟踪表
CREATE TABLE ticket_sla_tracking (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    ticket_id UUID NOT NULL REFERENCES tickets(id) ON DELETE CASCADE,
    
    -- SLA类型
    sla_type VARCHAR(50) NOT NULL, -- response, resolution, escalation
    
    -- SLA配置
    sla_minutes INTEGER NOT NULL,
    start_time TIMESTAMP WITH TIME ZONE NOT NULL,
    deadline TIMESTAMP WITH TIME ZONE NOT NULL,
    
    -- SLA状态
    status VARCHAR(20) NOT NULL DEFAULT 'active', -- active, met, breached, paused
    met_at TIMESTAMP WITH TIME ZONE,
    breached_at TIMESTAMP WITH TIME ZONE,
    
    -- 暂停信息
    paused_duration_minutes INTEGER DEFAULT 0,
    
    -- 审计字段
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    -- 元数据
    metadata JSONB DEFAULT '{}',
    
    -- 约束
    CONSTRAINT ticket_sla_tracking_sla_minutes_check CHECK (sla_minutes > 0),
    CONSTRAINT ticket_sla_tracking_paused_duration_check CHECK (paused_duration_minutes >= 0),
    CONSTRAINT ticket_sla_tracking_deadline_check CHECK (deadline > start_time)
);

-- 创建索引

-- tickets 表索引
CREATE INDEX idx_tickets_ticket_number ON tickets(ticket_number);
CREATE INDEX idx_tickets_title ON tickets(title);
CREATE INDEX idx_tickets_category ON tickets(category);
CREATE INDEX idx_tickets_priority ON tickets(priority);
CREATE INDEX idx_tickets_status ON tickets(status);
CREATE INDEX idx_tickets_alert_id ON tickets(alert_id);
CREATE INDEX idx_tickets_parent_ticket_id ON tickets(parent_ticket_id);
CREATE INDEX idx_tickets_reporter_id ON tickets(reporter_id);
CREATE INDEX idx_tickets_assignee_id ON tickets(assignee_id);
CREATE INDEX idx_tickets_team_id ON tickets(team_id);
CREATE INDEX idx_tickets_reported_at ON tickets(reported_at);
CREATE INDEX idx_tickets_assigned_at ON tickets(assigned_at);
CREATE INDEX idx_tickets_resolved_at ON tickets(resolved_at);
CREATE INDEX idx_tickets_closed_at ON tickets(closed_at);
CREATE INDEX idx_tickets_sla_level ON tickets(sla_level);
CREATE INDEX idx_tickets_response_deadline ON tickets(response_deadline);
CREATE INDEX idx_tickets_resolution_deadline ON tickets(resolution_deadline);
CREATE INDEX idx_tickets_sla_breached ON tickets(sla_breached);
CREATE INDEX idx_tickets_created_at ON tickets(created_at);
CREATE INDEX idx_tickets_updated_at ON tickets(updated_at);

-- GIN 索引
CREATE INDEX idx_tickets_labels_gin ON tickets USING GIN(labels);
CREATE INDEX idx_tickets_tags_gin ON tickets USING GIN(tags);
CREATE INDEX idx_tickets_custom_fields_gin ON tickets USING GIN(custom_fields);

-- 复合索引
CREATE INDEX idx_tickets_status_priority ON tickets(status, priority);
CREATE INDEX idx_tickets_assignee_status ON tickets(assignee_id, status);
CREATE INDEX idx_tickets_category_status ON tickets(category, status);
CREATE INDEX idx_tickets_reported_status ON tickets(reported_at, status);

-- 创建工单状态历史索引
CREATE INDEX idx_ticket_status_history_ticket_id ON ticket_status_history(ticket_id);
CREATE INDEX idx_ticket_status_history_changed_at ON ticket_status_history(changed_at);
CREATE INDEX idx_ticket_status_history_changed_by ON ticket_status_history(changed_by);
CREATE INDEX idx_ticket_status_history_new_status ON ticket_status_history(new_status);

-- 配置 TimescaleDB 超表
SELECT create_hypertable('ticket_status_history', 'changed_at', 
    chunk_time_interval => INTERVAL '1 day',
    if_not_exists => TRUE
);

-- 设置数据保留策略：保留2年的工单状态历史
SELECT add_retention_policy('ticket_status_history', INTERVAL '2 years', if_not_exists => TRUE);

-- 注意：压缩策略需要在TimescaleDB 2.0+版本中启用columnstore
-- SELECT add_compression_policy('ticket_status_history', INTERVAL '30 days', if_not_exists => TRUE);

-- ticket_comments 表索引
CREATE INDEX idx_ticket_comments_ticket_id ON ticket_comments(ticket_id);
CREATE INDEX idx_ticket_comments_author_id ON ticket_comments(author_id);
CREATE INDEX idx_ticket_comments_parent_comment_id ON ticket_comments(parent_comment_id);
CREATE INDEX idx_ticket_comments_comment_type ON ticket_comments(comment_type);
CREATE INDEX idx_ticket_comments_is_internal ON ticket_comments(is_internal);
CREATE INDEX idx_ticket_comments_created_at ON ticket_comments(created_at);

-- ticket_attachments 表索引
CREATE INDEX idx_ticket_attachments_ticket_id ON ticket_attachments(ticket_id);
CREATE INDEX idx_ticket_attachments_comment_id ON ticket_attachments(comment_id);
CREATE INDEX idx_ticket_attachments_filename ON ticket_attachments(filename);
CREATE INDEX idx_ticket_attachments_file_hash ON ticket_attachments(file_hash);
CREATE INDEX idx_ticket_attachments_uploaded_by ON ticket_attachments(uploaded_by);
CREATE INDEX idx_ticket_attachments_uploaded_at ON ticket_attachments(uploaded_at);

-- ticket_work_logs 表索引
CREATE INDEX idx_ticket_work_logs_ticket_id ON ticket_work_logs(ticket_id);
CREATE INDEX idx_ticket_work_logs_worker_id ON ticket_work_logs(worker_id);
CREATE INDEX idx_ticket_work_logs_work_type ON ticket_work_logs(work_type);
CREATE INDEX idx_ticket_work_logs_start_time ON ticket_work_logs(start_time);
CREATE INDEX idx_ticket_work_logs_billable ON ticket_work_logs(billable);

-- ticket_relationships 表索引
CREATE INDEX idx_ticket_relationships_source_ticket_id ON ticket_relationships(source_ticket_id);
CREATE INDEX idx_ticket_relationships_target_ticket_id ON ticket_relationships(target_ticket_id);
CREATE INDEX idx_ticket_relationships_type ON ticket_relationships(relationship_type);

-- ticket_sla_tracking 表索引
CREATE INDEX idx_ticket_sla_tracking_ticket_id ON ticket_sla_tracking(ticket_id);
CREATE INDEX idx_ticket_sla_tracking_sla_type ON ticket_sla_tracking(sla_type);
CREATE INDEX idx_ticket_sla_tracking_status ON ticket_sla_tracking(status);
CREATE INDEX idx_ticket_sla_tracking_deadline ON ticket_sla_tracking(deadline);
CREATE INDEX idx_ticket_sla_tracking_start_time ON ticket_sla_tracking(start_time);

-- 创建更新时间触发器
CREATE TRIGGER update_tickets_updated_at
    BEFORE UPDATE ON tickets
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_ticket_comments_updated_at
    BEFORE UPDATE ON ticket_comments
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_ticket_work_logs_updated_at
    BEFORE UPDATE ON ticket_work_logs
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_ticket_sla_tracking_updated_at
    BEFORE UPDATE ON ticket_sla_tracking
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- 创建工单编号生成函数
CREATE OR REPLACE FUNCTION generate_ticket_number()
RETURNS TEXT AS $$
DECLARE
    year_part TEXT;
    sequence_num INTEGER;
    ticket_num TEXT;
BEGIN
    -- 获取当前年份
    year_part := EXTRACT(YEAR FROM CURRENT_DATE)::TEXT;
    
    -- 获取当年的工单序号
    SELECT COALESCE(MAX(CAST(SUBSTRING(ticket_number FROM 'TK-' || year_part || '-(\d+)') AS INTEGER)), 0) + 1
    INTO sequence_num
    FROM tickets
    WHERE ticket_number LIKE 'TK-' || year_part || '-%';
    
    -- 生成工单编号
    ticket_num := 'TK-' || year_part || '-' || LPAD(sequence_num::TEXT, 6, '0');
    
    RETURN ticket_num;
END;
$$ LANGUAGE plpgsql;

-- 添加表注释
COMMENT ON TABLE tickets IS '工单表，存储工单的基本信息和处理状态';
COMMENT ON TABLE ticket_status_history IS '工单状态历史表，记录工单状态变更历史';
COMMENT ON TABLE ticket_comments IS '工单评论表，存储工单的评论和讨论';
COMMENT ON TABLE ticket_attachments IS '工单附件表，存储工单相关的文件附件';
COMMENT ON TABLE ticket_work_logs IS '工单工作日志表，记录工单的工作时间和内容';
COMMENT ON TABLE ticket_relationships IS '工单关联表，定义工单之间的关联关系';
COMMENT ON TABLE ticket_sla_tracking IS '工单SLA跟踪表，监控工单的SLA执行情况';

-- 添加列注释
COMMENT ON COLUMN tickets.ticket_number IS '工单编号，格式：TK-YYYY-XXXXXX';
COMMENT ON COLUMN tickets.sla_level IS 'SLA等级，如P1（1小时）、P2（4小时）、P3（1天）、P4（3天）';
COMMENT ON COLUMN tickets.tags IS '工单标签数组，用于分类和搜索';
COMMENT ON COLUMN ticket_comments.is_internal IS '是否内部评论，内部评论对客户不可见';
COMMENT ON COLUMN ticket_work_logs.billable IS '是否计费工时，用于成本核算';
COMMENT ON COLUMN ticket_relationships.relationship_type IS '关联类型：blocks（阻塞）、duplicates（重复）、related_to（相关）等';