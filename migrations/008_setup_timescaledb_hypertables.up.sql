-- 设置TimescaleDB超表配置
-- 创建时间: 2024-01-01
-- 描述: 为时序数据表创建超表，设置数据保留策略和压缩策略

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

-- 创建rule_evaluations超表（规则执行历史时序数据）
SELECT create_hypertable(
    'rule_evaluations',
    'evaluation_time',
    chunk_time_interval => INTERVAL '1 day',
    if_not_exists => TRUE
);

-- 创建data_source_health_checks超表（数据源健康检查时序数据）
SELECT create_hypertable(
    'data_source_health_checks',
    'check_time',
    chunk_time_interval => INTERVAL '1 hour',
    if_not_exists => TRUE
);

-- 创建data_source_queries超表（数据源查询历史时序数据）
SELECT create_hypertable(
    'data_source_queries',
    'query_time',
    chunk_time_interval => INTERVAL '1 day',
    if_not_exists => TRUE
);

-- ticket_status_history 表将在后续迁移中创建和配置

-- 创建knowledge_document_access_logs超表（文档访问日志时序数据）
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

-- 设置数据保留策略

-- alert_metrics: 保留90天的详细数据
SELECT add_retention_policy(
    'alert_metrics',
    INTERVAL '90 days',
    if_not_exists => TRUE
);

-- alert_history: 保留1年的历史数据
SELECT add_retention_policy(
    'alert_history',
    INTERVAL '1 year',
    if_not_exists => TRUE
);

-- rule_evaluations: 保留30天的执行历史
SELECT add_retention_policy(
    'rule_evaluations',
    INTERVAL '30 days',
    if_not_exists => TRUE
);

-- data_source_health_checks: 保留30天的健康检查数据
SELECT add_retention_policy(
    'data_source_health_checks',
    INTERVAL '30 days',
    if_not_exists => TRUE
);

-- data_source_queries: 保留7天的查询历史
SELECT add_retention_policy(
    'data_source_queries',
    INTERVAL '7 days',
    if_not_exists => TRUE
);

-- knowledge_document_access_logs: 保留6个月的访问日志
SELECT add_retention_policy(
    'knowledge_document_access_logs',
    INTERVAL '6 months',
    if_not_exists => TRUE
);

-- user_audit_logs: 保留1年的审计日志
SELECT add_retention_policy(
    'user_audit_logs',
    INTERVAL '1 year',
    if_not_exists => TRUE
);

-- 设置压缩策略（对于较旧的数据进行压缩以节省存储空间）

-- alert_metrics: 7天后压缩
SELECT add_compression_policy(
    'alert_metrics',
    INTERVAL '7 days',
    if_not_exists => TRUE
);

-- alert_history: 30天后压缩
SELECT add_compression_policy(
    'alert_history',
    INTERVAL '30 days',
    if_not_exists => TRUE
);

-- rule_evaluations: 3天后压缩
SELECT add_compression_policy(
    'rule_evaluations',
    INTERVAL '3 days',
    if_not_exists => TRUE
);

-- data_source_health_checks: 3天后压缩
SELECT add_compression_policy(
    'data_source_health_checks',
    INTERVAL '3 days',
    if_not_exists => TRUE
);

-- data_source_queries: 1天后压缩
SELECT add_compression_policy(
    'data_source_queries',
    INTERVAL '1 day',
    if_not_exists => TRUE
);



-- knowledge_document_access_logs: 7天后压缩
SELECT add_compression_policy(
    'knowledge_document_access_logs',
    INTERVAL '7 days',
    if_not_exists => TRUE
);

-- user_audit_logs: 30天后压缩
SELECT add_compression_policy(
    'user_audit_logs',
    INTERVAL '30 days',
    if_not_exists => TRUE
);

-- 创建连续聚合视图（Continuous Aggregates）用于性能优化

-- 每小时告警指标聚合
CREATE MATERIALIZED VIEW alert_metrics_hourly
WITH (timescaledb.continuous) AS
SELECT
    time_bucket('1 hour', timestamp) AS hour,
    alert_id,
    metric_name,
    AVG(value) AS avg_value,
    MAX(value) AS max_value,
    MIN(value) AS min_value,
    COUNT(*) AS count
FROM alert_metrics
GROUP BY hour, alert_id, metric_name
WITH NO DATA;

-- 每日告警指标聚合
CREATE MATERIALIZED VIEW alert_metrics_daily
WITH (timescaledb.continuous) AS
SELECT
    time_bucket('1 day', timestamp) AS day,
    alert_id,
    metric_name,
    AVG(value) AS avg_value,
    MAX(value) AS max_value,
    MIN(value) AS min_value,
    COUNT(*) AS count
FROM alert_metrics
GROUP BY day, alert_id, metric_name
WITH NO DATA;

-- 每小时数据源健康检查聚合
CREATE MATERIALIZED VIEW data_source_health_hourly
WITH (timescaledb.continuous) AS
SELECT
    time_bucket('1 hour', check_time) AS hour,
    data_source_id,
    AVG(duration_ms) AS avg_response_time,
    MAX(duration_ms) AS max_response_time,
    COUNT(*) AS total_checks,
    COUNT(*) FILTER (WHERE success = true) AS healthy_checks,
    COUNT(*) FILTER (WHERE success = false) AS unhealthy_checks
FROM data_source_health_checks
GROUP BY hour, data_source_id
WITH NO DATA;

-- 每日文档访问统计
CREATE MATERIALIZED VIEW knowledge_document_access_daily
WITH (timescaledb.continuous) AS
SELECT
    time_bucket('1 day', accessed_at) AS day,
    document_id,
    access_type,
    COUNT(*) AS access_count,
    COUNT(DISTINCT user_id) AS unique_users,
    COUNT(DISTINCT session_id) AS unique_sessions
FROM knowledge_document_access_logs
GROUP BY day, document_id, access_type
WITH NO DATA;

-- 设置连续聚合的刷新策略

-- 每小时告警指标聚合：每15分钟刷新
SELECT add_continuous_aggregate_policy(
    'alert_metrics_hourly',
    start_offset => INTERVAL '2 hours',
    end_offset => INTERVAL '15 minutes',
    schedule_interval => INTERVAL '15 minutes',
    if_not_exists => TRUE
);

-- 每日告警指标聚合：每小时刷新
SELECT add_continuous_aggregate_policy(
    'alert_metrics_daily',
    start_offset => INTERVAL '2 days',
    end_offset => INTERVAL '1 hour',
    schedule_interval => INTERVAL '1 hour',
    if_not_exists => TRUE
);

-- 每小时数据源健康检查聚合：每15分钟刷新
SELECT add_continuous_aggregate_policy(
    'data_source_health_hourly',
    start_offset => INTERVAL '2 hours',
    end_offset => INTERVAL '15 minutes',
    schedule_interval => INTERVAL '15 minutes',
    if_not_exists => TRUE
);

-- 每日文档访问统计：每小时刷新
SELECT add_continuous_aggregate_policy(
    'knowledge_document_access_daily',
    start_offset => INTERVAL '2 days',
    end_offset => INTERVAL '1 hour',
    schedule_interval => INTERVAL '1 hour',
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

-- 添加注释
COMMENT ON MATERIALIZED VIEW alert_metrics_hourly IS '告警指标每小时聚合视图，用于性能优化';
COMMENT ON MATERIALIZED VIEW alert_metrics_daily IS '告警指标每日聚合视图，用于趋势分析';
COMMENT ON MATERIALIZED VIEW data_source_health_hourly IS '数据源健康检查每小时聚合视图';
COMMENT ON MATERIALIZED VIEW knowledge_document_access_daily IS '知识库文档访问每日统计视图';

-- 创建时序数据查询的辅助函数

-- 获取指定时间范围内的告警指标统计
CREATE OR REPLACE FUNCTION get_alert_metrics_stats(
    p_alert_id UUID,
    p_metric_name TEXT,
    p_start_time TIMESTAMP WITH TIME ZONE,
    p_end_time TIMESTAMP WITH TIME ZONE
)
RETURNS TABLE (
    avg_value DOUBLE PRECISION,
    max_value DOUBLE PRECISION,
    min_value DOUBLE PRECISION,
    count BIGINT
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        AVG(am.value) as avg_value,
        MAX(am.value) as max_value,
        MIN(am.value) as min_value,
        COUNT(*) as count
    FROM alert_metrics am
    WHERE am.alert_id = p_alert_id
      AND am.metric_name = p_metric_name
      AND am.timestamp >= p_start_time
      AND am.timestamp <= p_end_time;
END;
$$ LANGUAGE plpgsql;

-- 获取数据源健康状态统计
CREATE OR REPLACE FUNCTION get_data_source_health_stats(
    p_data_source_id UUID,
    p_start_time TIMESTAMP WITH TIME ZONE,
    p_end_time TIMESTAMP WITH TIME ZONE
)
RETURNS TABLE (
    total_checks BIGINT,
    healthy_checks BIGINT,
    unhealthy_checks BIGINT,
    avg_response_time DOUBLE PRECISION,
    uptime_percentage DOUBLE PRECISION
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        COUNT(*) as total_checks,
        COUNT(*) FILTER (WHERE dsh.success = true) as healthy_checks,
        COUNT(*) FILTER (WHERE dsh.success = false) as unhealthy_checks,
        AVG(dsh.duration_ms) as avg_response_time,
        (COUNT(*) FILTER (WHERE dsh.success = true) * 100.0 / COUNT(*)) as uptime_percentage
    FROM data_source_health_checks dsh
    WHERE dsh.data_source_id = p_data_source_id
      AND dsh.check_time >= p_start_time
      AND dsh.check_time <= p_end_time;
END;
$$ LANGUAGE plpgsql;

-- 输出配置信息
DO $$
BEGIN
    RAISE NOTICE 'TimescaleDB超表配置完成:';
    RAISE NOTICE '- 已创建8个超表用于时序数据存储';
    RAISE NOTICE '- 已设置数据保留策略（7天-2年不等）';
    RAISE NOTICE '- 已设置压缩策略（1天-30天后压缩）';
    RAISE NOTICE '- 已创建4个连续聚合视图用于性能优化';
    RAISE NOTICE '- 已创建时序查询辅助函数';
END $$;