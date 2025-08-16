-- 回滚工单表
-- 创建时间: 2024-01-01
-- 描述: 删除工单相关表和函数

-- 删除触发器
DROP TRIGGER IF EXISTS update_ticket_sla_tracking_updated_at ON ticket_sla_tracking;
DROP TRIGGER IF EXISTS update_ticket_work_logs_updated_at ON ticket_work_logs;
DROP TRIGGER IF EXISTS update_ticket_comments_updated_at ON ticket_comments;
DROP TRIGGER IF EXISTS update_tickets_updated_at ON tickets;

-- 删除函数
DROP FUNCTION IF EXISTS generate_ticket_number();

-- 删除表（按依赖关系逆序删除）
DROP TABLE IF EXISTS ticket_sla_tracking;
DROP TABLE IF EXISTS ticket_relationships;
DROP TABLE IF EXISTS ticket_work_logs;
DROP TABLE IF EXISTS ticket_attachments;
DROP TABLE IF EXISTS ticket_comments;
DROP TABLE IF EXISTS ticket_status_history;
DROP TABLE IF EXISTS tickets;