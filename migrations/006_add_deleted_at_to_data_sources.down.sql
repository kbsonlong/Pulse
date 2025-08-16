-- 回滚data_sources表的软删除字段
-- 创建时间: 2024-01-01
-- 描述: 移除数据源表的软删除支持

-- 删除相关索引
DROP INDEX IF EXISTS idx_data_sources_deleted_at;
DROP INDEX IF EXISTS idx_data_sources_name;
DROP INDEX IF EXISTS idx_data_sources_status;
DROP INDEX IF EXISTS idx_data_sources_type;
DROP INDEX IF EXISTS idx_data_sources_enabled;
DROP INDEX IF EXISTS idx_data_sources_health_status;
DROP INDEX IF EXISTS idx_data_sources_type_status;
DROP INDEX IF EXISTS idx_data_sources_status_enabled;

-- 恢复原始name唯一约束
CREATE UNIQUE INDEX idx_data_sources_name ON data_sources(name);

-- 删除deleted_at字段
ALTER TABLE data_sources DROP COLUMN deleted_at;