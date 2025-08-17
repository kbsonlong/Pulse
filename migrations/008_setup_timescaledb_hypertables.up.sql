-- 设置TimescaleDB超表配置（简化版本，移除压缩策略）
-- 创建时间: 2024-01-01
-- 描述: 为时序数据表创建超表，不使用压缩策略以避免columnstore依赖

-- 确保TimescaleDB扩展已启用
-- 这个检查在001_init_timescaledb.up.sql中已经完成

-- 创建alert_metrics超表（告警指标时序数据）
SELECT create_hypertable(
    'alert_metrics',
    'timestamp',
    chunk_time_interval => INTERVAL '1 day',
    if_not_exists => TRUE
);

-- 创建alert_history超表（告警历史时序数据）
SELECT create_hypertable(
    'alert_history',
    'timestamp',
    chunk_time_interval => INTERVAL '1 day',
    if_not_exists => TRUE
);

-- 创建rule_evaluations超表（规则评估时序数据）
SELECT create_hypertable(
    'rule_evaluations',
    'evaluated_at',
    chunk_time_interval => INTERVAL '1 day',
    if_not_exists => TRUE
);

-- 创建data_source_health_checks超表（数据源健康检查时序数据）
SELECT create_hypertable(
    'data_source_health_checks',
    'check_time',
    chunk_time_interval => INTERVAL '1 day',
    if_not_exists => TRUE
);

-- 创建data_source_queries超表（数据源查询时序数据）
SELECT create_hypertable(
    'data_source_queries',
    'executed_at',
    chunk_time_interval => INTERVAL '1 day',
    if_not_exists => TRUE
);

-- 创建knowledge_document_access_logs超表（知识库访问日志时序数据）
SELECT create_hypertable(
    'knowledge_document_access_logs',
    'accessed_at',
    chunk_time_interval => INTERVAL '1 day',
    if_not_exists => TRUE
);

-- 创建user_audit_logs超表（用户审计日志时序数据）
SELECT create_hypertable(
    'user_audit_logs',
    'created_at',
    chunk_time_interval => INTERVAL '1 day',
    if_not_exists => TRUE
);

-- 设置数据保留策略（删除旧数据以节省存储空间）

-- alert_metrics: 保留90天
SELECT add_retention_policy(
    'alert_metrics',
    INTERVAL '90 days',
    if_not_exists => TRUE
);

-- alert_history: 保留2年
SELECT add_retention_policy(
    'alert_history',
    INTERVAL '2 years',
    if_not_exists => TRUE
);

-- rule_evaluations: 保留30天
SELECT add_retention_policy(
    'rule_evaluations',
    INTERVAL '30 days',
    if_not_exists => TRUE
);

-- data_source_health_checks: 保留30天
SELECT add_retention_policy(
    'data_source_health_checks',
    INTERVAL '30 days',
    if_not_exists => TRUE
);

-- data_source_queries: 保留7天
SELECT add_retention_policy(
    'data_source_queries',
    INTERVAL '7 days',
    if_not_exists => TRUE
);

-- knowledge_document_access_logs: 保留1年
SELECT add_retention_policy(
    'knowledge_document_access_logs',
    INTERVAL '1 year',
    if_not_exists => TRUE
);

-- user_audit_logs: 保留2年
SELECT add_retention_policy(
    'user_audit_logs',
    INTERVAL '2 years',
    if_not_exists => TRUE
);

-- 创建索引优化查询性能

-- alert_metrics 时序优化索引
CREATE INDEX IF NOT EXISTS idx_alert_metrics_alert_metric_time 
ON alert_metrics (alert_id, metric_name, timestamp DESC);

-- data_source_health_checks 时序优化索引
CREATE INDEX IF NOT EXISTS idx_data_source_health_source_time 
ON data_source_health_checks (data_source_id, check_time DESC);

-- knowledge_document_access_logs 时序优化索引
CREATE INDEX IF NOT EXISTS idx_knowledge_access_doc_time 
ON knowledge_document_access_logs (document_id, accessed_at DESC);

-- user_audit_logs 时序优化索引
CREATE INDEX IF NOT EXISTS idx_user_audit_user_time 
ON user_audit_logs (user_id, created_at DESC);

-- 输出配置信息
DO $$
BEGIN
    RAISE NOTICE 'TimescaleDB超表配置完成（简化版本）:';
    RAISE NOTICE '- 已创建7个超表用于时序数据存储';
    RAISE NOTICE '- 已设置数据保留策略（7天-2年不等）';
    RAISE NOTICE '- 已创建时序优化索引';
    RAISE NOTICE '- 跳过压缩策略以避免columnstore依赖';
END $$;