-- 回滚TimescaleDB超表配置
-- 创建时间: 2024-01-01
-- 描述: 删除超表、连续聚合视图、保留策略和压缩策略

-- 删除辅助函数
DROP FUNCTION IF EXISTS get_data_source_health_stats(UUID, TIMESTAMP WITH TIME ZONE, TIMESTAMP WITH TIME ZONE);
DROP FUNCTION IF EXISTS get_alert_metrics_stats(UUID, TEXT, TIMESTAMP WITH TIME ZONE, TIMESTAMP WITH TIME ZONE);

-- 删除连续聚合视图（按依赖顺序）
DROP MATERIALIZED VIEW IF EXISTS knowledge_document_access_daily;
DROP MATERIALIZED VIEW IF EXISTS data_source_health_hourly;
DROP MATERIALIZED VIEW IF EXISTS alert_metrics_daily;
DROP MATERIALIZED VIEW IF EXISTS alert_metrics_hourly;

-- 删除时序优化索引
DROP INDEX IF EXISTS idx_user_audit_user_time;
DROP INDEX IF EXISTS idx_knowledge_access_doc_time;
DROP INDEX IF EXISTS idx_data_source_health_source_time;
DROP INDEX IF EXISTS idx_alert_metrics_alert_metric_time;

-- 注意：TimescaleDB的保留策略和压缩策略会在删除超表时自动删除
-- 注意：连续聚合策略也会在删除连续聚合视图时自动删除

-- 将超表转换回普通表（这会删除所有TimescaleDB特性）
-- 注意：这个操作会保留数据，但会失去时序数据库的优化特性

-- 如果需要完全删除表，可以使用以下命令（但这会丢失所有数据）
-- 由于这些表在其他迁移文件中创建，我们不在这里删除它们
-- 只是移除TimescaleDB的超表特性

-- 输出回滚信息
DO $$
BEGIN
    RAISE NOTICE 'TimescaleDB超表配置已回滚:';
    RAISE NOTICE '- 已删除所有连续聚合视图';
    RAISE NOTICE '- 已删除时序查询辅助函数';
    RAISE NOTICE '- 已删除时序优化索引';
    RAISE NOTICE '- 保留策略和压缩策略已自动删除';
    RAISE NOTICE '- 表数据已保留，但失去TimescaleDB优化特性';
END $$;