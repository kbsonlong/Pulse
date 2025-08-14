-- 回滚规则表
-- 创建时间: 2024-01-01
-- 描述: 删除规则相关的所有表和索引

-- 删除触发器
DROP TRIGGER IF EXISTS update_rule_groups_updated_at ON rule_groups;
DROP TRIGGER IF EXISTS update_rules_updated_at ON rules;
DROP TRIGGER IF EXISTS update_rule_templates_updated_at ON rule_templates;

-- 删除表（按依赖关系逆序删除）
DROP TABLE IF EXISTS rule_labels CASCADE;
DROP TABLE IF EXISTS rule_dependencies CASCADE;
DROP TABLE IF EXISTS rule_templates CASCADE;
DROP TABLE IF EXISTS rule_evaluations CASCADE;
DROP TABLE IF EXISTS rules CASCADE;
DROP TABLE IF EXISTS rule_groups CASCADE;