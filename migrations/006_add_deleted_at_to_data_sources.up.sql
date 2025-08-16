-- 为data_sources表添加软删除字段
-- 创建时间: 2024-01-01
-- 描述: 为数据源表添加软删除支持

-- 添加deleted_at字段
ALTER TABLE data_sources ADD COLUMN deleted_at TIMESTAMP WITH TIME ZONE;

-- 创建deleted_at索引
CREATE INDEX idx_data_sources_deleted_at ON data_sources(deleted_at) WHERE deleted_at IS NOT NULL;

-- 更新现有索引以支持软删除
DROP INDEX IF EXISTS idx_data_sources_name;
CREATE UNIQUE INDEX idx_data_sources_name ON data_sources(name) WHERE deleted_at IS NULL;

-- 为状态字段创建索引（排除已删除记录）
CREATE INDEX idx_data_sources_status ON data_sources(status) WHERE deleted_at IS NULL;
CREATE INDEX idx_data_sources_type ON data_sources(type) WHERE deleted_at IS NULL;
CREATE INDEX idx_data_sources_enabled ON data_sources(enabled) WHERE deleted_at IS NULL;
CREATE INDEX idx_data_sources_health_status ON data_sources(last_health_check_status) WHERE deleted_at IS NULL;

-- 创建复合索引
CREATE INDEX idx_data_sources_type_status ON data_sources(type, status) WHERE deleted_at IS NULL;
CREATE INDEX idx_data_sources_status_enabled ON data_sources(status, enabled) WHERE deleted_at IS NULL;