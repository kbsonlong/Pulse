-- 回滚数据源表
-- 创建时间: 2024-01-01
-- 描述: 删除数据源相关的所有表和索引

-- 删除触发器
DROP TRIGGER IF EXISTS update_data_sources_updated_at ON data_sources;
DROP TRIGGER IF EXISTS update_data_source_templates_updated_at ON data_source_templates;

-- 删除表（按依赖关系逆序删除）
DROP TABLE IF EXISTS data_source_permissions CASCADE;
DROP TABLE IF EXISTS data_source_labels CASCADE;
DROP TABLE IF EXISTS data_source_templates CASCADE;
DROP TABLE IF EXISTS data_source_queries CASCADE;
DROP TABLE IF EXISTS data_source_health_checks CASCADE;
DROP TABLE IF EXISTS data_sources CASCADE;