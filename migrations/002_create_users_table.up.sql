-- 创建用户表
-- 创建时间: 2024-01-01
-- 描述: 创建用户表，包含用户认证、权限管理和审计信息

-- 创建用户表
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    full_name VARCHAR(100) NOT NULL,
    phone VARCHAR(20),
    avatar_url TEXT,
    
    -- 用户状态和角色
    status user_status NOT NULL DEFAULT 'active',
    role user_role NOT NULL DEFAULT 'viewer',
    
    -- 权限相关
    permissions JSONB DEFAULT '{}',
    department VARCHAR(100),
    team VARCHAR(100),
    
    -- 认证相关
    email_verified BOOLEAN DEFAULT FALSE,
    email_verified_at TIMESTAMP WITH TIME ZONE,
    phone_verified BOOLEAN DEFAULT FALSE,
    phone_verified_at TIMESTAMP WITH TIME ZONE,
    
    -- 登录相关
    last_login_at TIMESTAMP WITH TIME ZONE,
    last_login_ip INET,
    login_count INTEGER DEFAULT 0,
    failed_login_attempts INTEGER DEFAULT 0,
    locked_until TIMESTAMP WITH TIME ZONE,
    
    -- 密码相关
    password_changed_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    password_reset_token VARCHAR(255),
    password_reset_expires_at TIMESTAMP WITH TIME ZONE,
    
    -- 审计字段
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_by UUID,
    updated_by UUID,
    
    -- 软删除
    deleted_at TIMESTAMP WITH TIME ZONE,
    deleted_by UUID,
    
    -- 元数据
    metadata JSONB DEFAULT '{}',
    
    -- 约束
    CONSTRAINT users_username_length CHECK (char_length(username) >= 3),
    CONSTRAINT users_email_format CHECK (email ~* '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$'),
    CONSTRAINT users_phone_format CHECK (phone IS NULL OR phone ~* '^\+?[1-9]\d{1,14}$'),
    CONSTRAINT users_failed_attempts_check CHECK (failed_login_attempts >= 0),
    CONSTRAINT users_login_count_check CHECK (login_count >= 0)
);

-- 创建索引
CREATE INDEX idx_users_username ON users(username) WHERE deleted_at IS NULL;
CREATE INDEX idx_users_email ON users(email) WHERE deleted_at IS NULL;
CREATE INDEX idx_users_status ON users(status) WHERE deleted_at IS NULL;
CREATE INDEX idx_users_role ON users(role) WHERE deleted_at IS NULL;
CREATE INDEX idx_users_department ON users(department) WHERE deleted_at IS NULL;
CREATE INDEX idx_users_team ON users(team) WHERE deleted_at IS NULL;
CREATE INDEX idx_users_created_at ON users(created_at);
CREATE INDEX idx_users_updated_at ON users(updated_at);
CREATE INDEX idx_users_deleted_at ON users(deleted_at) WHERE deleted_at IS NOT NULL;
CREATE INDEX idx_users_last_login_at ON users(last_login_at);
CREATE INDEX idx_users_email_verified ON users(email_verified);

-- 创建 GIN 索引用于 JSONB 字段
CREATE INDEX idx_users_permissions_gin ON users USING GIN(permissions);
CREATE INDEX idx_users_metadata_gin ON users USING GIN(metadata);

-- 创建复合索引
CREATE INDEX idx_users_status_role ON users(status, role) WHERE deleted_at IS NULL;
CREATE INDEX idx_users_department_team ON users(department, team) WHERE deleted_at IS NULL;

-- 创建更新时间触发器函数
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- 创建更新时间触发器
CREATE TRIGGER update_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- 创建用户会话表
CREATE TABLE user_sessions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    session_token VARCHAR(255) UNIQUE NOT NULL,
    refresh_token VARCHAR(255) UNIQUE,
    
    -- 会话信息
    ip_address INET,
    user_agent TEXT,
    device_info JSONB DEFAULT '{}',
    
    -- 时间相关
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    last_accessed_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    -- 状态
    is_active BOOLEAN DEFAULT TRUE,
    revoked_at TIMESTAMP WITH TIME ZONE,
    revoked_by UUID,
    revoke_reason VARCHAR(255),
    
    -- 元数据
    metadata JSONB DEFAULT '{}'
);

-- 创建用户会话表索引
CREATE INDEX idx_user_sessions_user_id ON user_sessions(user_id);
CREATE INDEX idx_user_sessions_session_token ON user_sessions(session_token);
CREATE INDEX idx_user_sessions_refresh_token ON user_sessions(refresh_token);
CREATE INDEX idx_user_sessions_expires_at ON user_sessions(expires_at);
CREATE INDEX idx_user_sessions_is_active ON user_sessions(is_active);
CREATE INDEX idx_user_sessions_created_at ON user_sessions(created_at);
CREATE INDEX idx_user_sessions_last_accessed_at ON user_sessions(last_accessed_at);

-- 创建用户操作日志表
CREATE TABLE user_audit_logs (
    id UUID DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    
    -- 操作信息
    action VARCHAR(100) NOT NULL,
    resource_type VARCHAR(100),
    resource_id VARCHAR(255),
    
    -- 请求信息
    ip_address INET,
    user_agent TEXT,
    request_id VARCHAR(255),
    
    -- 操作详情
    old_values JSONB,
    new_values JSONB,
    changes JSONB,
    
    -- 结果
    success BOOLEAN NOT NULL,
    error_message TEXT,
    
    -- 时间
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    -- 元数据
    metadata JSONB DEFAULT '{}',
    
    -- 复合主键包含分区列
    PRIMARY KEY (id, created_at)
);

-- 创建用户操作日志表索引
CREATE INDEX idx_user_audit_logs_user_id ON user_audit_logs(user_id);
CREATE INDEX idx_user_audit_logs_action ON user_audit_logs(action);
CREATE INDEX idx_user_audit_logs_resource_type ON user_audit_logs(resource_type);
CREATE INDEX idx_user_audit_logs_resource_id ON user_audit_logs(resource_id);
CREATE INDEX idx_user_audit_logs_created_at ON user_audit_logs(created_at);
CREATE INDEX idx_user_audit_logs_success ON user_audit_logs(success);
CREATE INDEX idx_user_audit_logs_ip_address ON user_audit_logs(ip_address);

-- 创建 GIN 索引用于 JSONB 字段
CREATE INDEX idx_user_audit_logs_old_values_gin ON user_audit_logs USING GIN(old_values);
CREATE INDEX idx_user_audit_logs_new_values_gin ON user_audit_logs USING GIN(new_values);
CREATE INDEX idx_user_audit_logs_changes_gin ON user_audit_logs USING GIN(changes);
CREATE INDEX idx_user_audit_logs_metadata_gin ON user_audit_logs USING GIN(metadata);

-- 插入默认管理员用户
INSERT INTO users (
    username,
    email,
    password_hash,
    full_name,
    status,
    role,
    email_verified,
    email_verified_at
) VALUES (
    'admin',
    'admin@example.com',
    crypt('admin123', gen_salt('bf')), -- 使用 bcrypt 加密
    'System Administrator',
    'active',
    'admin',
    TRUE,
    CURRENT_TIMESTAMP
);

-- 添加表注释
COMMENT ON TABLE users IS '用户表，存储系统用户的基本信息、认证信息和权限信息';
COMMENT ON TABLE user_sessions IS '用户会话表，存储用户登录会话信息';
COMMENT ON TABLE user_audit_logs IS '用户操作审计日志表，记录用户的所有操作行为';

-- 添加列注释
COMMENT ON COLUMN users.id IS '用户唯一标识符';
COMMENT ON COLUMN users.username IS '用户名，系统内唯一';
COMMENT ON COLUMN users.email IS '邮箱地址，系统内唯一';
COMMENT ON COLUMN users.password_hash IS '密码哈希值';
COMMENT ON COLUMN users.permissions IS '用户权限配置，JSON格式';
COMMENT ON COLUMN users.metadata IS '用户元数据，JSON格式';