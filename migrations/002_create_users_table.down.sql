-- 回滚用户表创建
-- 创建时间: 2024-01-01
-- 描述: 删除用户相关的所有表、索引、触发器和函数

-- 删除表（按依赖关系逆序删除）
DROP TABLE IF EXISTS user_audit_logs CASCADE;
DROP TABLE IF EXISTS user_sessions CASCADE;
DROP TABLE IF EXISTS users CASCADE;

-- 删除触发器函数
DROP FUNCTION IF EXISTS update_updated_at_column() CASCADE;