-- 回滚知识库表
-- 创建时间: 2024-01-01
-- 描述: 删除知识库相关表和函数

-- 删除触发器
DROP TRIGGER IF EXISTS update_knowledge_document_stats ON knowledge_documents;
DROP TRIGGER IF EXISTS update_knowledge_categories_path ON knowledge_categories;
DROP TRIGGER IF EXISTS update_knowledge_document_ratings_updated_at ON knowledge_document_ratings;
DROP TRIGGER IF EXISTS update_knowledge_tags_updated_at ON knowledge_tags;
DROP TRIGGER IF EXISTS update_knowledge_document_comments_updated_at ON knowledge_document_comments;
DROP TRIGGER IF EXISTS update_knowledge_documents_updated_at ON knowledge_documents;
DROP TRIGGER IF EXISTS update_knowledge_categories_updated_at ON knowledge_categories;

-- 删除函数
DROP FUNCTION IF EXISTS update_document_stats();
DROP FUNCTION IF EXISTS update_category_path();

-- 删除表（按依赖关系逆序删除）
DROP TABLE IF EXISTS knowledge_document_ratings;
DROP TABLE IF EXISTS knowledge_document_access_logs;
DROP TABLE IF EXISTS knowledge_document_tags;
DROP TABLE IF EXISTS knowledge_tags;
DROP TABLE IF EXISTS knowledge_document_attachments;
DROP TABLE IF EXISTS knowledge_document_comments;
DROP TABLE IF EXISTS knowledge_document_versions;
DROP TABLE IF EXISTS knowledge_documents;
DROP TABLE IF EXISTS knowledge_categories;