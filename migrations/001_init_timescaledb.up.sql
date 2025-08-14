-- 初始化 TimescaleDB 扩展
-- 创建时间: 2024-01-01
-- 描述: 启用 TimescaleDB 扩展，为时序数据存储做准备

-- 创建 TimescaleDB 扩展
CREATE EXTENSION IF NOT EXISTS timescaledb CASCADE;

-- 创建 UUID 扩展（用于生成唯一标识符）
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- 创建 pgcrypto 扩展（用于密码加密）
CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- 创建 btree_gin 扩展（用于优化 GIN 索引）
CREATE EXTENSION IF NOT EXISTS btree_gin;

-- 设置时区为 UTC
SET timezone = 'UTC';

-- 创建枚举类型

-- 用户状态枚举
CREATE TYPE user_status AS ENUM ('active', 'inactive', 'suspended');

-- 用户角色枚举
CREATE TYPE user_role AS ENUM ('admin', 'operator', 'viewer');

-- 告警级别枚举
CREATE TYPE alert_severity AS ENUM ('critical', 'high', 'medium', 'low', 'info');

-- 告警状态枚举
CREATE TYPE alert_status AS ENUM ('firing', 'resolved', 'suppressed', 'acknowledged');

-- 规则状态枚举
CREATE TYPE rule_status AS ENUM ('active', 'inactive', 'draft');

-- 数据源类型枚举
CREATE TYPE datasource_type AS ENUM ('prometheus', 'grafana', 'elasticsearch', 'influxdb', 'custom');

-- 数据源状态枚举
CREATE TYPE datasource_status AS ENUM ('active', 'inactive', 'error');

-- 工单状态枚举
CREATE TYPE ticket_status AS ENUM ('open', 'in_progress', 'resolved', 'closed', 'cancelled');

-- 工单优先级枚举
CREATE TYPE ticket_priority AS ENUM ('urgent', 'high', 'medium', 'low');

-- 通知渠道类型枚举
CREATE TYPE notification_channel_type AS ENUM ('email', 'sms', 'webhook', 'dingtalk', 'wechat', 'slack');

-- 通知状态枚举
CREATE TYPE notification_status AS ENUM ('pending', 'sent', 'failed', 'delivered');

-- 知识库文档类型枚举
CREATE TYPE knowledge_doc_type AS ENUM ('runbook', 'troubleshooting', 'faq', 'guide', 'reference');

-- 知识库文档状态枚举
CREATE TYPE knowledge_doc_status AS ENUM ('draft', 'published', 'archived');