-- 回滚告警表
-- 创建时间: 2024-01-01
-- 描述: 删除告警相关的所有表和索引

-- 删除触发器
DROP TRIGGER IF EXISTS update_alerts_updated_at ON alerts;
DROP TRIGGER IF EXISTS update_alert_groups_updated_at ON alert_groups;
DROP TRIGGER IF EXISTS update_alert_silences_updated_at ON alert_silences;

-- 删除表（按依赖关系逆序删除）
DROP TABLE IF EXISTS alert_silences CASCADE;
DROP TABLE IF EXISTS alert_group_members CASCADE;
DROP TABLE IF EXISTS alert_groups CASCADE;
DROP TABLE IF EXISTS alert_metrics CASCADE;
DROP TABLE IF EXISTS alert_history CASCADE;
DROP TABLE IF EXISTS alerts CASCADE;