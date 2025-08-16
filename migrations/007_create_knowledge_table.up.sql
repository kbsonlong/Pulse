-- 创建知识库表
-- 创建时间: 2024-01-01
-- 描述: 创建知识库相关表，包含文档管理、分类、版本控制和搜索功能

-- 创建知识库分类表
CREATE TABLE knowledge_categories (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    
    -- 分类信息
    name VARCHAR(200) NOT NULL,
    slug VARCHAR(200) NOT NULL, -- URL友好的标识符
    description TEXT,
    
    -- 层级结构
    parent_id UUID REFERENCES knowledge_categories(id),
    level INTEGER NOT NULL DEFAULT 0,
    path VARCHAR(1000), -- 分类路径，如 /root/category1/subcategory
    
    -- 显示设置
    display_order INTEGER DEFAULT 0,
    icon VARCHAR(100), -- 图标名称
    color VARCHAR(20), -- 颜色代码
    
    -- 状态
    is_active BOOLEAN DEFAULT TRUE,
    is_public BOOLEAN DEFAULT TRUE,
    
    -- 权限
    required_role VARCHAR(50), -- 访问所需角色
    
    -- 统计信息
    document_count INTEGER DEFAULT 0,
    
    -- 审计字段
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_by UUID REFERENCES users(id),
    updated_by UUID REFERENCES users(id),
    
    -- 元数据
    metadata JSONB DEFAULT '{}',
    
    -- 约束
    CONSTRAINT knowledge_categories_slug_unique UNIQUE (slug),
    CONSTRAINT knowledge_categories_level_check CHECK (level >= 0),
    CONSTRAINT knowledge_categories_no_self_parent CHECK (id != parent_id)
);

-- 创建知识库文档表
CREATE TABLE knowledge_documents (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    
    -- 文档基本信息
    title VARCHAR(500) NOT NULL,
    slug VARCHAR(500) NOT NULL, -- URL友好的标识符
    summary TEXT, -- 文档摘要
    content TEXT NOT NULL, -- 文档内容
    content_type VARCHAR(50) DEFAULT 'markdown', -- markdown, html, text
    
    -- 分类和标签
    category_id UUID REFERENCES knowledge_categories(id),
    tags VARCHAR(255)[], -- 标签数组
    
    -- 文档类型和状态
    document_type knowledge_doc_type DEFAULT 'guide',
    status knowledge_doc_status DEFAULT 'draft',
    
    -- 优先级和重要性
    priority INTEGER DEFAULT 0, -- 数值越大优先级越高
    is_featured BOOLEAN DEFAULT FALSE, -- 是否精选
    is_pinned BOOLEAN DEFAULT FALSE, -- 是否置顶
    
    -- 访问控制
    is_public BOOLEAN DEFAULT TRUE,
    required_role VARCHAR(50), -- 访问所需角色
    allowed_users UUID[], -- 允许访问的用户ID数组
    allowed_teams UUID[], -- 允许访问的团队ID数组
    
    -- 版本信息
    version INTEGER DEFAULT 1,
    is_latest_version BOOLEAN DEFAULT TRUE,
    parent_document_id UUID REFERENCES knowledge_documents(id), -- 原始文档ID（用于版本控制）
    
    -- 发布信息
    published_at TIMESTAMP WITH TIME ZONE,
    expires_at TIMESTAMP WITH TIME ZONE, -- 过期时间
    
    -- 作者信息
    author_id UUID NOT NULL REFERENCES users(id),
    reviewer_id UUID REFERENCES users(id), -- 审核者
    reviewed_at TIMESTAMP WITH TIME ZONE,
    
    -- 统计信息
    view_count INTEGER DEFAULT 0,
    like_count INTEGER DEFAULT 0,
    comment_count INTEGER DEFAULT 0,
    download_count INTEGER DEFAULT 0,
    
    -- 搜索权重
    search_weight DOUBLE PRECISION DEFAULT 1.0,
    
    -- SEO信息
    meta_title VARCHAR(200),
    meta_description VARCHAR(500),
    meta_keywords VARCHAR(500),
    
    -- 附件信息
    attachments JSONB DEFAULT '[]', -- 附件列表
    
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
    CONSTRAINT knowledge_documents_slug_unique UNIQUE (slug),
    CONSTRAINT knowledge_documents_version_check CHECK (version > 0),
    CONSTRAINT knowledge_documents_priority_check CHECK (priority >= 0),
    CONSTRAINT knowledge_documents_search_weight_check CHECK (search_weight > 0),
    CONSTRAINT knowledge_documents_view_count_check CHECK (view_count >= 0),
    CONSTRAINT knowledge_documents_like_count_check CHECK (like_count >= 0),
    CONSTRAINT knowledge_documents_comment_count_check CHECK (comment_count >= 0),
    CONSTRAINT knowledge_documents_download_count_check CHECK (download_count >= 0)
);

-- 创建文档版本历史表
CREATE TABLE knowledge_document_versions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    document_id UUID NOT NULL REFERENCES knowledge_documents(id) ON DELETE CASCADE,
    
    -- 版本信息
    version_number INTEGER NOT NULL,
    title VARCHAR(500) NOT NULL,
    content TEXT NOT NULL,
    content_type VARCHAR(50) DEFAULT 'markdown',
    
    -- 变更信息
    change_summary TEXT, -- 变更摘要
    change_type VARCHAR(50) DEFAULT 'update', -- create, update, minor_edit, major_revision
    
    -- 版本作者
    author_id UUID NOT NULL REFERENCES users(id),
    
    -- 时间戳
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    -- 元数据
    metadata JSONB DEFAULT '{}',
    
    -- 约束
    CONSTRAINT knowledge_document_versions_unique UNIQUE (document_id, version_number),
    CONSTRAINT knowledge_document_versions_version_check CHECK (version_number > 0)
);

-- 创建文档评论表
CREATE TABLE knowledge_document_comments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    document_id UUID NOT NULL REFERENCES knowledge_documents(id) ON DELETE CASCADE,
    
    -- 评论内容
    content TEXT NOT NULL,
    content_type VARCHAR(20) DEFAULT 'text', -- text, markdown
    
    -- 评论类型
    comment_type VARCHAR(50) DEFAULT 'comment', -- comment, suggestion, question
    
    -- 回复信息
    parent_comment_id UUID REFERENCES knowledge_document_comments(id),
    
    -- 评论者信息
    author_id UUID NOT NULL REFERENCES users(id),
    
    -- 状态
    is_approved BOOLEAN DEFAULT TRUE,
    is_hidden BOOLEAN DEFAULT FALSE,
    
    -- 有用性评分
    helpful_count INTEGER DEFAULT 0,
    unhelpful_count INTEGER DEFAULT 0,
    
    -- 审计字段
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    -- 元数据
    metadata JSONB DEFAULT '{}',
    
    -- 约束
    CONSTRAINT knowledge_document_comments_helpful_check CHECK (helpful_count >= 0),
    CONSTRAINT knowledge_document_comments_unhelpful_check CHECK (unhelpful_count >= 0)
);

-- 创建文档附件表
CREATE TABLE knowledge_document_attachments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    document_id UUID NOT NULL REFERENCES knowledge_documents(id) ON DELETE CASCADE,
    
    -- 文件信息
    filename VARCHAR(255) NOT NULL,
    original_filename VARCHAR(255) NOT NULL,
    file_path VARCHAR(1000) NOT NULL,
    file_size BIGINT NOT NULL,
    mime_type VARCHAR(100),
    
    -- 文件类型
    file_type VARCHAR(50), -- image, document, video, audio, archive等
    
    -- 文件哈希
    file_hash VARCHAR(64), -- SHA256哈希
    
    -- 显示设置
    display_name VARCHAR(255),
    description TEXT,
    display_order INTEGER DEFAULT 0,
    
    -- 访问控制
    is_public BOOLEAN DEFAULT TRUE,
    download_count INTEGER DEFAULT 0,
    
    -- 上传信息
    uploaded_by UUID NOT NULL REFERENCES users(id),
    uploaded_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    -- 元数据
    metadata JSONB DEFAULT '{}',
    
    -- 约束
    CONSTRAINT knowledge_document_attachments_file_size_check CHECK (file_size > 0),
    CONSTRAINT knowledge_document_attachments_download_count_check CHECK (download_count >= 0)
);

-- 创建文档标签表
CREATE TABLE knowledge_tags (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    
    -- 标签信息
    name VARCHAR(100) NOT NULL,
    slug VARCHAR(100) NOT NULL,
    description TEXT,
    
    -- 显示设置
    color VARCHAR(20), -- 标签颜色
    icon VARCHAR(100), -- 图标名称
    
    -- 统计信息
    usage_count INTEGER DEFAULT 0,
    
    -- 状态
    is_active BOOLEAN DEFAULT TRUE,
    
    -- 审计字段
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_by UUID REFERENCES users(id),
    
    -- 元数据
    metadata JSONB DEFAULT '{}',
    
    -- 约束
    CONSTRAINT knowledge_tags_name_unique UNIQUE (name),
    CONSTRAINT knowledge_tags_slug_unique UNIQUE (slug),
    CONSTRAINT knowledge_tags_usage_count_check CHECK (usage_count >= 0)
);

-- 创建文档标签关联表
CREATE TABLE knowledge_document_tags (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    document_id UUID NOT NULL REFERENCES knowledge_documents(id) ON DELETE CASCADE,
    tag_id UUID NOT NULL REFERENCES knowledge_tags(id) ON DELETE CASCADE,
    
    -- 关联信息
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_by UUID REFERENCES users(id),
    
    -- 约束
    CONSTRAINT knowledge_document_tags_unique UNIQUE (document_id, tag_id)
);

-- 创建文档访问日志表（时序数据）
CREATE TABLE knowledge_document_access_logs (
    id UUID DEFAULT uuid_generate_v4(),
    document_id UUID NOT NULL REFERENCES knowledge_documents(id) ON DELETE CASCADE,
    
    -- 访问信息
    user_id UUID REFERENCES users(id), -- 可能为空（匿名访问）
    session_id VARCHAR(100), -- 会话ID
    
    -- 访问详情
    access_type VARCHAR(50) DEFAULT 'view', -- view, download, print, share
    user_agent TEXT,
    ip_address INET,
    referer TEXT,
    
    -- 地理位置
    country VARCHAR(100),
    region VARCHAR(100),
    city VARCHAR(100),
    
    -- 时间戳
    accessed_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    -- 元数据
    metadata JSONB DEFAULT '{}',
    
    -- 复合主键包含分区列
    PRIMARY KEY (id, accessed_at)
);

-- 创建文档评分表
CREATE TABLE knowledge_document_ratings (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    document_id UUID NOT NULL REFERENCES knowledge_documents(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    
    -- 评分信息
    rating INTEGER NOT NULL, -- 1-5星评分
    review TEXT, -- 评价内容
    
    -- 评分维度
    accuracy_rating INTEGER, -- 准确性评分
    usefulness_rating INTEGER, -- 有用性评分
    clarity_rating INTEGER, -- 清晰度评分
    
    -- 审计字段
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    -- 元数据
    metadata JSONB DEFAULT '{}',
    
    -- 约束
    CONSTRAINT knowledge_document_ratings_unique UNIQUE (document_id, user_id),
    CONSTRAINT knowledge_document_ratings_rating_check CHECK (rating >= 1 AND rating <= 5),
    CONSTRAINT knowledge_document_ratings_accuracy_check CHECK (accuracy_rating IS NULL OR (accuracy_rating >= 1 AND accuracy_rating <= 5)),
    CONSTRAINT knowledge_document_ratings_usefulness_check CHECK (usefulness_rating IS NULL OR (usefulness_rating >= 1 AND usefulness_rating <= 5)),
    CONSTRAINT knowledge_document_ratings_clarity_check CHECK (clarity_rating IS NULL OR (clarity_rating >= 1 AND clarity_rating <= 5))
);

-- 创建索引

-- knowledge_categories 表索引
CREATE INDEX idx_knowledge_categories_name ON knowledge_categories(name);
CREATE INDEX idx_knowledge_categories_slug ON knowledge_categories(slug);
CREATE INDEX idx_knowledge_categories_parent_id ON knowledge_categories(parent_id);
CREATE INDEX idx_knowledge_categories_level ON knowledge_categories(level);
CREATE INDEX idx_knowledge_categories_is_active ON knowledge_categories(is_active);
CREATE INDEX idx_knowledge_categories_is_public ON knowledge_categories(is_public);
CREATE INDEX idx_knowledge_categories_display_order ON knowledge_categories(display_order);

-- knowledge_documents 表索引
CREATE INDEX idx_knowledge_documents_title ON knowledge_documents(title);
CREATE INDEX idx_knowledge_documents_slug ON knowledge_documents(slug);
CREATE INDEX idx_knowledge_documents_category_id ON knowledge_documents(category_id);
CREATE INDEX idx_knowledge_documents_document_type ON knowledge_documents(document_type);
CREATE INDEX idx_knowledge_documents_status ON knowledge_documents(status);
CREATE INDEX idx_knowledge_documents_author_id ON knowledge_documents(author_id);
CREATE INDEX idx_knowledge_documents_is_public ON knowledge_documents(is_public);
CREATE INDEX idx_knowledge_documents_is_featured ON knowledge_documents(is_featured);
CREATE INDEX idx_knowledge_documents_is_pinned ON knowledge_documents(is_pinned);
CREATE INDEX idx_knowledge_documents_published_at ON knowledge_documents(published_at);
CREATE INDEX idx_knowledge_documents_created_at ON knowledge_documents(created_at);
CREATE INDEX idx_knowledge_documents_updated_at ON knowledge_documents(updated_at);
CREATE INDEX idx_knowledge_documents_view_count ON knowledge_documents(view_count);
CREATE INDEX idx_knowledge_documents_priority ON knowledge_documents(priority);
CREATE INDEX idx_knowledge_documents_is_latest_version ON knowledge_documents(is_latest_version);
CREATE INDEX idx_knowledge_documents_parent_document_id ON knowledge_documents(parent_document_id);

-- GIN 索引
CREATE INDEX idx_knowledge_documents_tags_gin ON knowledge_documents USING GIN(tags);
CREATE INDEX idx_knowledge_documents_content_gin ON knowledge_documents USING GIN(to_tsvector('english', title || ' ' || COALESCE(summary, '') || ' ' || content));
CREATE INDEX idx_knowledge_documents_allowed_users_gin ON knowledge_documents USING GIN(allowed_users);
CREATE INDEX idx_knowledge_documents_allowed_teams_gin ON knowledge_documents USING GIN(allowed_teams);

-- 复合索引
CREATE INDEX idx_knowledge_documents_status_public ON knowledge_documents(status, is_public);
CREATE INDEX idx_knowledge_documents_category_status ON knowledge_documents(category_id, status);
CREATE INDEX idx_knowledge_documents_author_status ON knowledge_documents(author_id, status);
CREATE INDEX idx_knowledge_documents_featured_priority ON knowledge_documents(is_featured, priority DESC);

-- knowledge_document_versions 表索引
CREATE INDEX idx_knowledge_document_versions_document_id ON knowledge_document_versions(document_id);
CREATE INDEX idx_knowledge_document_versions_version_number ON knowledge_document_versions(version_number);
CREATE INDEX idx_knowledge_document_versions_author_id ON knowledge_document_versions(author_id);
CREATE INDEX idx_knowledge_document_versions_created_at ON knowledge_document_versions(created_at);

-- knowledge_document_comments 表索引
CREATE INDEX idx_knowledge_document_comments_document_id ON knowledge_document_comments(document_id);
CREATE INDEX idx_knowledge_document_comments_author_id ON knowledge_document_comments(author_id);
CREATE INDEX idx_knowledge_document_comments_parent_comment_id ON knowledge_document_comments(parent_comment_id);
CREATE INDEX idx_knowledge_document_comments_is_approved ON knowledge_document_comments(is_approved);
CREATE INDEX idx_knowledge_document_comments_created_at ON knowledge_document_comments(created_at);

-- knowledge_document_attachments 表索引
CREATE INDEX idx_knowledge_document_attachments_document_id ON knowledge_document_attachments(document_id);
CREATE INDEX idx_knowledge_document_attachments_filename ON knowledge_document_attachments(filename);
CREATE INDEX idx_knowledge_document_attachments_file_type ON knowledge_document_attachments(file_type);
CREATE INDEX idx_knowledge_document_attachments_file_hash ON knowledge_document_attachments(file_hash);
CREATE INDEX idx_knowledge_document_attachments_uploaded_by ON knowledge_document_attachments(uploaded_by);

-- knowledge_tags 表索引
CREATE INDEX idx_knowledge_tags_name ON knowledge_tags(name);
CREATE INDEX idx_knowledge_tags_slug ON knowledge_tags(slug);
CREATE INDEX idx_knowledge_tags_usage_count ON knowledge_tags(usage_count DESC);
CREATE INDEX idx_knowledge_tags_is_active ON knowledge_tags(is_active);

-- knowledge_document_tags 表索引
CREATE INDEX idx_knowledge_document_tags_document_id ON knowledge_document_tags(document_id);
CREATE INDEX idx_knowledge_document_tags_tag_id ON knowledge_document_tags(tag_id);

-- knowledge_document_access_logs 表索引
CREATE INDEX idx_knowledge_document_access_logs_document_id ON knowledge_document_access_logs(document_id);
CREATE INDEX idx_knowledge_document_access_logs_user_id ON knowledge_document_access_logs(user_id);
CREATE INDEX idx_knowledge_document_access_logs_accessed_at ON knowledge_document_access_logs(accessed_at);
CREATE INDEX idx_knowledge_document_access_logs_access_type ON knowledge_document_access_logs(access_type);
CREATE INDEX idx_knowledge_document_access_logs_ip_address ON knowledge_document_access_logs(ip_address);

-- knowledge_document_ratings 表索引
CREATE INDEX idx_knowledge_document_ratings_document_id ON knowledge_document_ratings(document_id);
CREATE INDEX idx_knowledge_document_ratings_user_id ON knowledge_document_ratings(user_id);
CREATE INDEX idx_knowledge_document_ratings_rating ON knowledge_document_ratings(rating);
CREATE INDEX idx_knowledge_document_ratings_created_at ON knowledge_document_ratings(created_at);

-- 创建更新时间触发器
CREATE TRIGGER update_knowledge_categories_updated_at
    BEFORE UPDATE ON knowledge_categories
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_knowledge_documents_updated_at
    BEFORE UPDATE ON knowledge_documents
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_knowledge_document_comments_updated_at
    BEFORE UPDATE ON knowledge_document_comments
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_knowledge_tags_updated_at
    BEFORE UPDATE ON knowledge_tags
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_knowledge_document_ratings_updated_at
    BEFORE UPDATE ON knowledge_document_ratings
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- 创建分类路径更新函数
CREATE OR REPLACE FUNCTION update_category_path()
RETURNS TRIGGER AS $$
BEGIN
    -- 更新分类路径
    IF NEW.parent_id IS NULL THEN
        NEW.path := '/' || NEW.slug;
        NEW.level := 0;
    ELSE
        SELECT path || '/' || NEW.slug, level + 1
        INTO NEW.path, NEW.level
        FROM knowledge_categories
        WHERE id = NEW.parent_id;
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- 创建分类路径更新触发器
CREATE TRIGGER update_knowledge_categories_path
    BEFORE INSERT OR UPDATE OF parent_id, slug ON knowledge_categories
    FOR EACH ROW
    EXECUTE FUNCTION update_category_path();

-- 创建文档统计更新函数
CREATE OR REPLACE FUNCTION update_document_stats()
RETURNS TRIGGER AS $$
BEGIN
    -- 更新分类的文档数量
    IF TG_OP = 'INSERT' THEN
        UPDATE knowledge_categories 
        SET document_count = document_count + 1 
        WHERE id = NEW.category_id;
        RETURN NEW;
    ELSIF TG_OP = 'DELETE' THEN
        UPDATE knowledge_categories 
        SET document_count = document_count - 1 
        WHERE id = OLD.category_id;
        RETURN OLD;
    ELSIF TG_OP = 'UPDATE' AND OLD.category_id != NEW.category_id THEN
        UPDATE knowledge_categories 
        SET document_count = document_count - 1 
        WHERE id = OLD.category_id;
        UPDATE knowledge_categories 
        SET document_count = document_count + 1 
        WHERE id = NEW.category_id;
        RETURN NEW;
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- 创建文档统计更新触发器
CREATE TRIGGER update_knowledge_document_stats
    AFTER INSERT OR UPDATE OR DELETE ON knowledge_documents
    FOR EACH ROW
    EXECUTE FUNCTION update_document_stats();

-- 添加表注释
COMMENT ON TABLE knowledge_categories IS '知识库分类表，支持层级结构';
COMMENT ON TABLE knowledge_documents IS '知识库文档表，存储文档内容和元数据';
COMMENT ON TABLE knowledge_document_versions IS '文档版本历史表，记录文档的版本变更';
COMMENT ON TABLE knowledge_document_comments IS '文档评论表，存储用户对文档的评论和反馈';
COMMENT ON TABLE knowledge_document_attachments IS '文档附件表，存储文档相关的文件附件';
COMMENT ON TABLE knowledge_tags IS '知识库标签表，用于文档分类和搜索';
COMMENT ON TABLE knowledge_document_tags IS '文档标签关联表，多对多关系';
COMMENT ON TABLE knowledge_document_access_logs IS '文档访问日志表，记录文档访问统计';
COMMENT ON TABLE knowledge_document_ratings IS '文档评分表，存储用户对文档的评分';

-- 添加列注释
COMMENT ON COLUMN knowledge_categories.path IS '分类路径，如 /root/category1/subcategory';
COMMENT ON COLUMN knowledge_documents.slug IS 'URL友好的标识符，用于SEO优化';
COMMENT ON COLUMN knowledge_documents.search_weight IS '搜索权重，影响搜索结果排序';
COMMENT ON COLUMN knowledge_documents.is_latest_version IS '是否为最新版本，用于版本控制';
COMMENT ON COLUMN knowledge_document_access_logs.access_type IS '访问类型：view（查看）、download（下载）、print（打印）、share（分享）';
COMMENT ON COLUMN knowledge_document_ratings.rating IS '1-5星评分，5星为最高';