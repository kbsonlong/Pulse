-- 回滚 TimescaleDB 扩展初始化
-- 创建时间: 2024-01-01
-- 描述: 删除所有创建的扩展、枚举类型和配置

-- 删除枚举类型（按依赖关系逆序删除）
DROP TYPE IF EXISTS knowledge_doc_status CASCADE;
DROP TYPE IF EXISTS knowledge_doc_type CASCADE;
DROP TYPE IF EXISTS notification_status CASCADE;
DROP TYPE IF EXISTS notification_channel_type CASCADE;
DROP TYPE IF EXISTS ticket_priority CASCADE;
DROP TYPE IF EXISTS ticket_status CASCADE;
DROP TYPE IF EXISTS datasource_status CASCADE;
DROP TYPE IF EXISTS datasource_type CASCADE;
DROP TYPE IF EXISTS rule_status CASCADE;
DROP TYPE IF EXISTS alert_status CASCADE;
DROP TYPE IF EXISTS alert_severity CASCADE;
DROP TYPE IF EXISTS user_role CASCADE;
DROP TYPE IF EXISTS user_status CASCADE;

-- 删除扩展（按依赖关系逆序删除）
DROP EXTENSION IF EXISTS btree_gin CASCADE;
DROP EXTENSION IF EXISTS pgcrypto CASCADE;
DROP EXTENSION IF EXISTS "uuid-ossp" CASCADE;
DROP EXTENSION IF EXISTS timescaledb CASCADE;