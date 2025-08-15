package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"Pulse/internal/models"
)

type ruleRepository struct {
	db *sqlx.DB
	tx *sqlx.Tx
}

// NewRuleRepository 创建规则仓储实例
func NewRuleRepository(db *sqlx.DB) RuleRepository {
	return &ruleRepository{
		db: db,
	}
}

// NewRuleRepositoryWithTx 创建带事务的规则仓储实例
func NewRuleRepositoryWithTx(tx *sqlx.Tx) RuleRepository {
	return &ruleRepository{
		tx: tx,
	}
}

// getExecutor 获取数据库执行器（事务或普通连接）
func (r *ruleRepository) getExecutor() sqlx.ExtContext {
	if r.tx != nil {
		return r.tx
	}
	return r.db
}

// Create 创建规则
func (r *ruleRepository) Create(ctx context.Context, rule *models.Rule) error {
	// 生成ID
	rule.ID = uuid.New().String()
	rule.CreatedAt = time.Now()
	rule.UpdatedAt = rule.CreatedAt
	
	// 序列化标签和注解
	labelsJSON, err := json.Marshal(rule.Labels)
	if err != nil {
		return fmt.Errorf("序列化标签失败: %w", err)
	}
	
	annotationsJSON, err := json.Marshal(rule.Annotations)
	if err != nil {
		return fmt.Errorf("序列化注解失败: %w", err)
	}
	
	conditionsJSON, err := json.Marshal(rule.Conditions)
	if err != nil {
		return fmt.Errorf("序列化条件失败: %w", err)
	}
	
	actionsJSON, err := json.Marshal(rule.Actions)
	if err != nil {
		return fmt.Errorf("序列化动作失败: %w", err)
	}
	
	query := `
		INSERT INTO rules (
			id, name, description, type, severity, status, enabled, expression,
			conditions, actions, labels, annotations, data_source_id,
			evaluation_interval, for_duration, keep_firing_for, threshold,
			recovery_threshold, no_data_state, exec_err_state,
			created_by, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23
		)
	`
	
	_, err = r.getExecutor().ExecContext(ctx, query,
		rule.ID, rule.Name, rule.Description, rule.Type, rule.Severity,
		rule.Status, rule.Enabled, rule.Expression, string(conditionsJSON),
		string(actionsJSON), string(labelsJSON), string(annotationsJSON),
		rule.DataSourceID, rule.EvaluationInterval, rule.ForDuration,
		rule.KeepFiringFor, rule.Threshold, rule.RecoveryThreshold,
		rule.NoDataState, rule.ExecErrState, rule.CreatedBy,
		rule.CreatedAt, rule.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("创建规则失败: %w", err)
	}
	
	return nil
}

// GetByID 根据ID获取规则
func (r *ruleRepository) GetByID(ctx context.Context, id string) (*models.Rule, error) {
	var rule models.Rule
	var labelsJSON, annotationsJSON, conditionsJSON, actionsJSON string

	query := `
		SELECT id, name, description, type, severity, status, enabled, expression,
		       conditions, actions, labels, annotations, data_source_id,
		       evaluation_interval, for_duration, keep_firing_for, threshold,
		       recovery_threshold, no_data_state, exec_err_state,
		       last_eval_at, last_eval_result, eval_count, alert_count,
		       created_by, updated_by, created_at, updated_at
		FROM rules 
		WHERE id = $1 AND deleted_at IS NULL`

	err := r.getExecutor().QueryRowxContext(ctx, query, id).Scan(
		&rule.ID, &rule.Name, &rule.Description, &rule.Type, &rule.Severity,
		&rule.Status, &rule.Enabled, &rule.Expression, &conditionsJSON,
		&actionsJSON, &labelsJSON, &annotationsJSON, &rule.DataSourceID,
		&rule.EvaluationInterval, &rule.ForDuration, &rule.KeepFiringFor,
		&rule.Threshold, &rule.RecoveryThreshold, &rule.NoDataState,
		&rule.ExecErrState, &rule.LastEvalAt, &rule.LastEvalResult,
		&rule.EvalCount, &rule.AlertCount, &rule.CreatedBy, &rule.UpdatedBy,
		&rule.CreatedAt, &rule.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("规则不存在")
		}
		return nil, fmt.Errorf("获取规则失败: %w", err)
	}

	// 反序列化条件
	if conditionsJSON != "" {
		err = json.Unmarshal([]byte(conditionsJSON), &rule.Conditions)
		if err != nil {
			return nil, fmt.Errorf("反序列化条件失败: %w", err)
		}
	}

	// 反序列化动作
	if actionsJSON != "" {
		err = json.Unmarshal([]byte(actionsJSON), &rule.Actions)
		if err != nil {
			return nil, fmt.Errorf("反序列化动作失败: %w", err)
		}
	}

	// 反序列化标签
	if labelsJSON != "" {
		err = json.Unmarshal([]byte(labelsJSON), &rule.Labels)
		if err != nil {
			return nil, fmt.Errorf("反序列化标签失败: %w", err)
		}
	}

	// 反序列化注解
	if annotationsJSON != "" {
		err = json.Unmarshal([]byte(annotationsJSON), &rule.Annotations)
		if err != nil {
			return nil, fmt.Errorf("反序列化注解失败: %w", err)
		}
	}

	return &rule, nil
}

// Update 更新规则
func (r *ruleRepository) Update(ctx context.Context, rule *models.Rule) error {
	rule.UpdatedAt = time.Now()

	// 序列化条件
	conditionsJSON, err := json.Marshal(rule.Conditions)
	if err != nil {
		return fmt.Errorf("序列化条件失败: %w", err)
	}

	// 序列化动作
	actionsJSON, err := json.Marshal(rule.Actions)
	if err != nil {
		return fmt.Errorf("序列化动作失败: %w", err)
	}

	// 序列化标签
	labelsJSON, err := json.Marshal(rule.Labels)
	if err != nil {
		return fmt.Errorf("序列化标签失败: %w", err)
	}

	// 序列化注解
	annotationsJSON, err := json.Marshal(rule.Annotations)
	if err != nil {
		return fmt.Errorf("序列化注解失败: %w", err)
	}

	query := `
		UPDATE rules SET
			name = $2,
			description = $3,
			type = $4,
			severity = $5,
			status = $6,
			enabled = $7,
			expression = $8,
			conditions = $9,
			actions = $10,
			labels = $11,
			annotations = $12,
			data_source_id = $13,
			evaluation_interval = $14,
			for_duration = $15,
			keep_firing_for = $16,
			threshold = $17,
			recovery_threshold = $18,
			no_data_state = $19,
			exec_err_state = $20,
			updated_by = $21,
			updated_at = $22
		WHERE id = $1 AND deleted_at IS NULL`

	_, err = r.getExecutor().ExecContext(ctx, query,
		rule.ID, rule.Name, rule.Description, rule.Type, rule.Severity,
		rule.Status, rule.Enabled, rule.Expression, string(conditionsJSON),
		string(actionsJSON), string(labelsJSON), string(annotationsJSON),
		rule.DataSourceID, rule.EvaluationInterval, rule.ForDuration,
		rule.KeepFiringFor, rule.Threshold, rule.RecoveryThreshold,
		rule.NoDataState, rule.ExecErrState, rule.UpdatedBy, rule.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("更新规则失败: %w", err)
	}

	return nil
}

// Delete 删除规则
func (r *ruleRepository) Delete(ctx context.Context, id string) error {
	query := `
		UPDATE rules 
		SET deleted_at = NOW() 
		WHERE id = $1 AND deleted_at IS NULL
	`
	
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("删除规则失败: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("获取删除结果失败: %w", err)
	}
	
	if rowsAffected == 0 {
		return models.ErrRuleNotFound
	}
	
	return nil
}

// SoftDelete 软删除规则
func (r *ruleRepository) SoftDelete(ctx context.Context, id string) error {
	now := time.Now()
	query := `
		UPDATE rules SET 
			deleted_at = $1,
			updated_at = $1
		WHERE id = $2 AND deleted_at IS NULL`

	_, err := r.db.ExecContext(ctx, query, now, id)
	if err != nil {
		return fmt.Errorf("软删除规则失败: %w", err)
	}
	return nil
}

// List 获取规则列表
func (r *ruleRepository) List(ctx context.Context, filter *models.RuleFilter) (*models.RuleList, error) {
	var conditions []string
	var args []interface{}
	argIndex := 1

	conditions = append(conditions, "deleted_at IS NULL")

	if filter != nil {
		if filter.DataSourceID != nil {
			conditions = append(conditions, fmt.Sprintf("data_source_id = $%d", argIndex))
			args = append(args, *filter.DataSourceID)
			argIndex++
		}

		if filter.Status != nil {
			conditions = append(conditions, fmt.Sprintf("status = $%d", argIndex))
			args = append(args, *filter.Status)
			argIndex++
		}

		if filter.Severity != nil {
			conditions = append(conditions, fmt.Sprintf("severity = $%d", argIndex))
			args = append(args, *filter.Severity)
			argIndex++
		}

		if filter.Keyword != nil && *filter.Keyword != "" {
			conditions = append(conditions, fmt.Sprintf("(name ILIKE $%d OR description ILIKE $%d)", argIndex, argIndex))
			args = append(args, "%"+*filter.Keyword+"%")
			argIndex++
		}
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// 获取总数
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM rules %s", whereClause)
	var total int64
	err := sqlx.GetContext(ctx, r.db, &total, countQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("获取规则总数失败: %w", err)
	}

	// 构建查询
	query := fmt.Sprintf(`
		SELECT id, name, description, type, severity, status, enabled, expression,
		       conditions, actions, labels, annotations, data_source_id,
		       evaluation_interval, for_duration, keep_firing_for, threshold,
		       recovery_threshold, no_data_state, exec_err_state,
		       last_eval_at, last_eval_result, eval_count, alert_count,
		       created_by, updated_by, created_at, updated_at
		FROM rules %s
		ORDER BY created_at DESC`, whereClause)

	// 添加分页
	if filter != nil && filter.Page > 0 && filter.PageSize > 0 {
		offset := (filter.Page - 1) * filter.PageSize
		query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
		args = append(args, filter.PageSize, offset)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("查询规则列表失败: %w", err)
	}
	defer rows.Close()

	var rules []*models.Rule
	for rows.Next() {
		var rule models.Rule
		var labelsJSON, annotationsJSON, conditionsJSON, actionsJSON string

		err := rows.Scan(
			&rule.ID, &rule.Name, &rule.Description, &rule.Type, &rule.Severity,
			&rule.Status, &rule.Enabled, &rule.Expression, &conditionsJSON,
			&actionsJSON, &labelsJSON, &annotationsJSON, &rule.DataSourceID,
			&rule.EvaluationInterval, &rule.ForDuration, &rule.KeepFiringFor,
			&rule.Threshold, &rule.RecoveryThreshold, &rule.NoDataState,
			&rule.ExecErrState, &rule.LastEvalAt, &rule.LastEvalResult,
			&rule.EvalCount, &rule.AlertCount, &rule.CreatedBy, &rule.UpdatedBy,
			&rule.CreatedAt, &rule.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("扫描规则数据失败: %w", err)
		}

		// 反序列化条件
		if conditionsJSON != "" {
			err = json.Unmarshal([]byte(conditionsJSON), &rule.Conditions)
			if err != nil {
				return nil, fmt.Errorf("反序列化条件失败: %w", err)
			}
		}

		// 反序列化动作
		if actionsJSON != "" {
			err = json.Unmarshal([]byte(actionsJSON), &rule.Actions)
			if err != nil {
				return nil, fmt.Errorf("反序列化动作失败: %w", err)
			}
		}

		// 反序列化标签
		if labelsJSON != "" {
			err = json.Unmarshal([]byte(labelsJSON), &rule.Labels)
			if err != nil {
				return nil, fmt.Errorf("反序列化标签失败: %w", err)
			}
		}

		// 反序列化注解
		if annotationsJSON != "" {
			err = json.Unmarshal([]byte(annotationsJSON), &rule.Annotations)
			if err != nil {
				return nil, fmt.Errorf("反序列化注解失败: %w", err)
			}
		}

		rules = append(rules, &rule)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("遍历规则数据失败: %w", err)
	}

	// 计算分页信息
	var totalPages int64 = 1
	if filter != nil && filter.PageSize > 0 {
		totalPages = (total + int64(filter.PageSize) - 1) / int64(filter.PageSize)
	}

	return &models.RuleList{
		Rules:      rules,
		Total:      total,
		TotalPages: totalPages,
	}, nil
}

// Count 获取规则总数
func (r *ruleRepository) Count(ctx context.Context, filter *models.RuleFilter) (int64, error) {
	var conditions []string
	var args []interface{}
	argIndex := 1

	conditions = append(conditions, "deleted_at IS NULL")

	if filter != nil {
		if filter.DataSourceID != nil {
			conditions = append(conditions, fmt.Sprintf("data_source_id = $%d", argIndex))
			args = append(args, *filter.DataSourceID)
			argIndex++
		}

		if filter.Status != nil {
			conditions = append(conditions, fmt.Sprintf("status = $%d", argIndex))
			args = append(args, *filter.Status)
			argIndex++
		}

		if filter.Enabled != nil {
			conditions = append(conditions, fmt.Sprintf("enabled = $%d", argIndex))
			args = append(args, *filter.Enabled)
			argIndex++
		}
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	query := fmt.Sprintf("SELECT COUNT(*) FROM rules %s", whereClause)
	var count int64
	err := r.db.GetContext(ctx, &count, query, args...)
	if err != nil {
		return 0, fmt.Errorf("获取规则总数失败: %w", err)
	}

	return count, nil
}

// Exists 检查规则是否存在
func (r *ruleRepository) Exists(ctx context.Context, id string) (bool, error) {
	var count int
	query := `SELECT COUNT(*) FROM rules WHERE id = $1 AND deleted_at IS NULL`
	err := r.db.GetContext(ctx, &count, query, id)
	if err != nil {
		return false, fmt.Errorf("检查规则是否存在失败: %w", err)
	}
	return count > 0, nil
}

// Enable 启用规则
func (r *ruleRepository) Enable(ctx context.Context, id string) error {
	now := time.Now()
	query := `
		UPDATE rules SET 
			enabled = true,
			status = $1,
			updated_at = $2
		WHERE id = $3 AND deleted_at IS NULL`

	_, err := r.db.ExecContext(ctx, query, models.RuleStatusActive, now, id)
	if err != nil {
		return fmt.Errorf("启用规则失败: %w", err)
	}
	return nil
}

// Disable 禁用规则
func (r *ruleRepository) Disable(ctx context.Context, id string) error {
	now := time.Now()
	query := `
		UPDATE rules SET 
			enabled = false,
			status = $1,
			updated_at = $2
		WHERE id = $3 AND deleted_at IS NULL`

	_, err := r.db.ExecContext(ctx, query, models.RuleStatusInactive, now, id)
	if err != nil {
		return fmt.Errorf("禁用规则失败: %w", err)
	}
	return nil
}

// UpdateEvaluation 更新规则评估信息
func (r *ruleRepository) UpdateEvaluation(ctx context.Context, id string, lastEval, nextEval time.Time, count int64) error {
	now := time.Now()
	query := `
		UPDATE rules SET 
			last_eval_at = $1,
			last_eval_result = $2,
			eval_count = $3,
			updated_at = $4
		WHERE id = $5 AND deleted_at IS NULL`

	_, err := r.db.ExecContext(ctx, query, lastEval, "success", count, now, id)
	if err != nil {
		return fmt.Errorf("更新规则评估信息失败: %w", err)
	}
	return nil
}

// GetByName 根据名称获取规则
func (r *ruleRepository) GetByName(ctx context.Context, name string) (*models.Rule, error) {
	var rule models.Rule
	var conditionsJSON, actionsJSON, labelsJSON, annotationsJSON string

	query := `
		SELECT id, name, description, type, severity, status, enabled, expression,
		       conditions, actions, labels, annotations, data_source_id,
		       evaluation_interval, for_duration, keep_firing_for, threshold,
		       recovery_threshold, no_data_state, exec_err_state,
		       last_eval_at, last_eval_result, eval_count, alert_count,
		       created_by, updated_by, created_at, updated_at
		FROM rules 
		WHERE name = $1 AND deleted_at IS NULL`

	err := r.getExecutor().QueryRowxContext(ctx, query, name).Scan(
		&rule.ID, &rule.Name, &rule.Description, &rule.Type, &rule.Severity,
		&rule.Status, &rule.Enabled, &rule.Expression, &conditionsJSON, &actionsJSON,
		&labelsJSON, &annotationsJSON, &rule.DataSourceID, &rule.EvaluationInterval,
		&rule.ForDuration, &rule.KeepFiringFor, &rule.Threshold, &rule.RecoveryThreshold,
		&rule.NoDataState, &rule.ExecErrState, &rule.LastEvalAt, &rule.LastEvalResult,
		&rule.EvalCount, &rule.AlertCount, &rule.CreatedBy, &rule.UpdatedBy,
		&rule.CreatedAt, &rule.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, models.ErrRuleNotFound
		}
		return nil, fmt.Errorf("根据名称获取规则失败: %w", err)
	}

	// 反序列化条件
	if conditionsJSON != "" {
		err = json.Unmarshal([]byte(conditionsJSON), &rule.Conditions)
		if err != nil {
			return nil, fmt.Errorf("反序列化条件失败: %w", err)
		}
	}

	// 反序列化动作
	if actionsJSON != "" {
		err = json.Unmarshal([]byte(actionsJSON), &rule.Actions)
		if err != nil {
			return nil, fmt.Errorf("反序列化动作失败: %w", err)
		}
	}

	// 反序列化标签
	if labelsJSON != "" {
		err = json.Unmarshal([]byte(labelsJSON), &rule.Labels)
		if err != nil {
			return nil, fmt.Errorf("反序列化标签失败: %w", err)
		}
	}

	// 反序列化注解
	if annotationsJSON != "" {
		err = json.Unmarshal([]byte(annotationsJSON), &rule.Annotations)
		if err != nil {
			return nil, fmt.Errorf("反序列化注解失败: %w", err)
		}
	}

	return &rule, nil
}

// GetStats 获取规则统计信息
func (r *ruleRepository) GetStats(ctx context.Context, filter *models.RuleFilter) (*models.RuleStats, error) {
	query := `
		SELECT 
			COUNT(*) as total,
			COUNT(CASE WHEN status = 'active' THEN 1 END) as active,
			COUNT(CASE WHEN status = 'inactive' THEN 1 END) as inactive,
			COUNT(CASE WHEN status = 'disabled' THEN 1 END) as disabled
		FROM rules 
		WHERE deleted_at IS NULL
	`
	
	stats := &models.RuleStats{
		ByType:     make(map[models.RuleType]int64),
		ByStatus:   make(map[models.RuleStatus]int64),
		BySeverity: make(map[models.AlertSeverity]int64),
	}
	
	var total, active, inactive, disabled int64
	err := r.getExecutor().QueryRowxContext(ctx, query).Scan(&total, &active, &inactive, &disabled)
	if err != nil {
		return nil, err
	}
	
	stats.Total = total
	stats.ActiveRules = active
	stats.Disabled = disabled
	
	return stats, nil
}

// IncrementAlertCount 增加规则的告警计数
func (r *ruleRepository) IncrementAlertCount(ctx context.Context, id string) error {
	query := `
		UPDATE rules 
		SET alert_count = alert_count + 1,
			last_alert_at = NOW(),
			updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`
	
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	
	if rowsAffected == 0 {
		return errors.New("规则不存在")
	}
	
	return nil
}

// IncrementEvaluationCount 增加规则的评估计数
func (r *ruleRepository) IncrementEvaluationCount(ctx context.Context, id string) error {
	query := `
		UPDATE rules 
		SET evaluation_count = evaluation_count + 1,
			last_eval_at = NOW(),
			updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`
	
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	
	if rowsAffected == 0 {
		return errors.New("规则不存在")
	}
	
	return nil
}

// SetTesting 设置规则为测试状态
func (r *ruleRepository) SetTesting(ctx context.Context, id string) error {
	query := `
		UPDATE rules 
		SET status = 'testing',
			updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`
	
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	
	if rowsAffected == 0 {
		return errors.New("规则不存在")
	}
	
	return nil
}

// TestRule 测试规则
func (r *ruleRepository) TestRule(ctx context.Context, rule *models.Rule) (*models.RuleTestResult, error) {
	start := time.Now()
	
	// 验证规则
	if err := rule.Validate(); err != nil {
		errorMsg := err.Error()
		return &models.RuleTestResult{
			Success:  false,
			Error:    &errorMsg,
			EvalTime: time.Since(start),
		}, nil
	}
	
	// 这里应该实际执行规则测试逻辑
	// 目前返回一个模拟的成功结果
	return &models.RuleTestResult{
		Success:  true,
		Result:   "规则测试通过",
		EvalTime: time.Since(start),
		DataPoints: []map[string]interface{}{
			{"timestamp": time.Now(), "value": 100},
		},
	}, nil
}

// UpdateLastEvaluation 更新最后评估信息
func (r *ruleRepository) UpdateLastEvaluation(ctx context.Context, id string, evalTime time.Time, result bool, error string) error {
	query := `
		UPDATE rules 
		SET last_eval_at = $1,
			last_eval_result = $2,
			eval_count = eval_count + 1,
			updated_at = NOW()
		WHERE id = $3 AND deleted_at IS NULL
	`
	
	// 将结果转换为字符串存储
	var resultStr string
	if result {
		if error == "" {
			resultStr = "success"
		} else {
			resultStr = "success_with_warning"
		}
	} else {
		if error == "" {
			resultStr = "failed"
		} else {
			resultStr = error
		}
	}
	
	result_exec, err := r.getExecutor().ExecContext(ctx, query, evalTime, resultStr, id)
	if err != nil {
		return fmt.Errorf("更新规则评估信息失败: %w", err)
	}
	
	rowsAffected, err := result_exec.RowsAffected()
	if err != nil {
		return fmt.Errorf("获取影响行数失败: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("规则不存在或已删除: %s", id)
	}
	
	return nil
}

// GetActiveCount 获取活跃规则数量
func (r *ruleRepository) GetActiveCount(ctx context.Context) (int64, error) {
	query := `SELECT COUNT(*) FROM rules WHERE enabled = true AND deleted_at IS NULL`
	var count int64
	err := r.db.GetContext(ctx, &count, query)
	if err != nil {
		return 0, fmt.Errorf("获取活跃规则数量失败: %w", err)
	}
	return count, nil
}

// GetByDataSourceID 根据数据源ID获取规则列表
func (r *ruleRepository) GetByDataSourceID(ctx context.Context, dataSourceID string) ([]*models.Rule, error) {
	query := `
		SELECT id, name, description, type, severity, status, enabled, expression,
		       conditions, actions, labels, annotations, data_source_id,
		       evaluation_interval, for_duration, keep_firing_for, threshold,
		       recovery_threshold, no_data_state, exec_err_state,
		       last_eval_at, last_eval_result, eval_count, alert_count,
		       created_by, updated_by, created_at, updated_at
		FROM rules 
		WHERE data_source_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
	`
	
	rows, err := r.getExecutor().QueryContext(ctx, query, dataSourceID)
	if err != nil {
		return nil, fmt.Errorf("查询数据源规则失败: %w", err)
	}
	defer rows.Close()
	
	var rules []*models.Rule
	for rows.Next() {
		var rule models.Rule
		var labelsJSON, annotationsJSON, conditionsJSON, actionsJSON string
		
		err := rows.Scan(
			&rule.ID, &rule.Name, &rule.Description, &rule.Type, &rule.Severity,
			&rule.Status, &rule.Enabled, &rule.Expression, &conditionsJSON,
			&actionsJSON, &labelsJSON, &annotationsJSON, &rule.DataSourceID,
			&rule.EvaluationInterval, &rule.ForDuration, &rule.KeepFiringFor,
			&rule.Threshold, &rule.RecoveryThreshold, &rule.NoDataState,
			&rule.ExecErrState, &rule.LastEvalAt, &rule.LastEvalResult,
			&rule.EvalCount, &rule.AlertCount, &rule.CreatedBy, &rule.UpdatedBy,
			&rule.CreatedAt, &rule.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("扫描规则数据失败: %w", err)
		}
		
		// 反序列化条件
		if conditionsJSON != "" {
			err = json.Unmarshal([]byte(conditionsJSON), &rule.Conditions)
			if err != nil {
				return nil, fmt.Errorf("反序列化条件失败: %w", err)
			}
		}
		
		// 反序列化动作
		if actionsJSON != "" {
			err = json.Unmarshal([]byte(actionsJSON), &rule.Actions)
			if err != nil {
				return nil, fmt.Errorf("反序列化动作失败: %w", err)
			}
		}
		
		// 反序列化标签
		if labelsJSON != "" {
			err = json.Unmarshal([]byte(labelsJSON), &rule.Labels)
			if err != nil {
				return nil, fmt.Errorf("反序列化标签失败: %w", err)
			}
		}
		
		// 反序列化注解
		if annotationsJSON != "" {
			err = json.Unmarshal([]byte(annotationsJSON), &rule.Annotations)
			if err != nil {
				return nil, fmt.Errorf("反序列化注解失败: %w", err)
			}
		}
		
		rules = append(rules, &rule)
	}
	
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("遍历规则数据失败: %w", err)
	}
	
	return rules, nil
}

// GetErrorCount 获取错误规则数量
func (r *ruleRepository) GetErrorCount(ctx context.Context) (int64, error) {
	query := `
		SELECT COUNT(*) 
		FROM rules 
		WHERE last_eval_result = false AND deleted_at IS NULL
	`
	
	var count int64
	err := r.getExecutor().QueryRowxContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get error count: %w", err)
	}
	
	return count, nil
}

// GetActiveRules 获取活跃规则列表
func (r *ruleRepository) GetActiveRules(ctx context.Context) ([]*models.Rule, error) {
	query := `
		SELECT id, name, description, type, severity, status, enabled, expression,
		       conditions, actions, labels, annotations, data_source_id,
		       evaluation_interval, for_duration, keep_firing_for, threshold,
		       recovery_threshold, no_data_state, exec_err_state,
		       last_eval_at, last_eval_result, eval_count, alert_count,
		       created_by, updated_by, created_at, updated_at
		FROM rules 
		WHERE enabled = true AND status = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC`

	rows, err := r.getExecutor().QueryContext(ctx, query, models.RuleStatusActive)
	if err != nil {
		return nil, fmt.Errorf("获取活跃规则失败: %w", err)
	}
	defer rows.Close()

	var rules []*models.Rule
	for rows.Next() {
		var rule models.Rule
		var labelsJSON, annotationsJSON, conditionsJSON, actionsJSON string

		err := rows.Scan(
			&rule.ID, &rule.Name, &rule.Description, &rule.Type, &rule.Severity,
			&rule.Status, &rule.Enabled, &rule.Expression, &conditionsJSON,
			&actionsJSON, &labelsJSON, &annotationsJSON, &rule.DataSourceID,
			&rule.EvaluationInterval, &rule.ForDuration, &rule.KeepFiringFor,
			&rule.Threshold, &rule.RecoveryThreshold, &rule.NoDataState,
			&rule.ExecErrState, &rule.LastEvalAt, &rule.LastEvalResult,
			&rule.EvalCount, &rule.AlertCount, &rule.CreatedBy, &rule.UpdatedBy,
			&rule.CreatedAt, &rule.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("扫描规则数据失败: %w", err)
		}

		// 反序列化条件
		if conditionsJSON != "" {
			err = json.Unmarshal([]byte(conditionsJSON), &rule.Conditions)
			if err != nil {
				return nil, fmt.Errorf("反序列化条件失败: %w", err)
			}
		}

		// 反序列化动作
		if actionsJSON != "" {
			err = json.Unmarshal([]byte(actionsJSON), &rule.Actions)
			if err != nil {
				return nil, fmt.Errorf("反序列化动作失败: %w", err)
			}
		}

		// 反序列化标签
		if labelsJSON != "" {
			err = json.Unmarshal([]byte(labelsJSON), &rule.Labels)
			if err != nil {
				return nil, fmt.Errorf("反序列化标签失败: %w", err)
			}
		}

		// 反序列化注解
		if annotationsJSON != "" {
			err = json.Unmarshal([]byte(annotationsJSON), &rule.Annotations)
			if err != nil {
				return nil, fmt.Errorf("反序列化注解失败: %w", err)
			}
		}

		rules = append(rules, &rule)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("遍历规则数据失败: %w", err)
	}

	return rules, nil
}

// GetRulesForEvaluation 获取需要评估的规则列表
func (r *ruleRepository) GetRulesForEvaluation(ctx context.Context) ([]*models.Rule, error) {
	query := `
		SELECT id, name, description, type, severity, status, enabled, expression,
		       conditions, actions, labels, annotations, data_source_id,
		       evaluation_interval, for_duration, keep_firing_for, threshold,
		       recovery_threshold, no_data_state, exec_err_state,
		       last_eval_at, last_eval_result, eval_count, alert_count,
		       created_by, updated_by, created_at, updated_at
		FROM rules 
		WHERE enabled = true AND status = $1 AND deleted_at IS NULL
		  AND (last_eval_at IS NULL OR 
		       last_eval_at + evaluation_interval <= CURRENT_TIMESTAMP)
		ORDER BY last_eval_at ASC NULLS FIRST`

	rows, err := r.db.QueryContext(ctx, query, models.RuleStatusActive)
	if err != nil {
		return nil, fmt.Errorf("获取待评估规则列表失败: %w", err)
	}
	defer rows.Close()

	var rules []*models.Rule
	for rows.Next() {
		var rule models.Rule
		var labelsJSON, annotationsJSON, conditionsJSON, actionsJSON string

		err := rows.Scan(
			&rule.ID, &rule.Name, &rule.Description, &rule.Type, &rule.Severity,
			&rule.Status, &rule.Enabled, &rule.Expression, &conditionsJSON,
			&actionsJSON, &labelsJSON, &annotationsJSON, &rule.DataSourceID,
			&rule.EvaluationInterval, &rule.ForDuration, &rule.KeepFiringFor,
			&rule.Threshold, &rule.RecoveryThreshold, &rule.NoDataState,
			&rule.ExecErrState, &rule.LastEvalAt, &rule.LastEvalResult,
			&rule.EvalCount, &rule.AlertCount, &rule.CreatedBy, &rule.UpdatedBy,
			&rule.CreatedAt, &rule.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("扫描待评估规则失败: %w", err)
		}

		// 反序列化条件
		if conditionsJSON != "" {
			err = json.Unmarshal([]byte(conditionsJSON), &rule.Conditions)
			if err != nil {
				return nil, fmt.Errorf("反序列化条件失败: %w", err)
			}
		}

		// 反序列化动作
		if actionsJSON != "" {
			err = json.Unmarshal([]byte(actionsJSON), &rule.Actions)
			if err != nil {
				return nil, fmt.Errorf("反序列化动作失败: %w", err)
			}
		}

		// 反序列化标签
		if labelsJSON != "" {
			err = json.Unmarshal([]byte(labelsJSON), &rule.Labels)
			if err != nil {
				return nil, fmt.Errorf("反序列化标签失败: %w", err)
			}
		}

		// 反序列化注解
		if annotationsJSON != "" {
			err = json.Unmarshal([]byte(annotationsJSON), &rule.Annotations)
			if err != nil {
				return nil, fmt.Errorf("反序列化注解失败: %w", err)
			}
		}

		rules = append(rules, &rule)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("遍历待评估规则失败: %w", err)
	}

	return rules, nil
}

// BatchCreate 批量创建规则
func (r *ruleRepository) BatchCreate(ctx context.Context, rules []*models.Rule) error {
	if len(rules) == 0 {
		return nil
	}

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("开始事务失败: %w", err)
	}
	defer tx.Rollback()

	for _, rule := range rules {
		if rule.ID == "" {
			rule.ID = uuid.New().String()
		}

		now := time.Now()
		rule.CreatedAt = now
		rule.UpdatedAt = now

		if rule.Status == "" {
			rule.Status = models.RuleStatusActive
		}

		// 序列化条件、动作、标签和注解
		conditionsJSON, err := json.Marshal(rule.Conditions)
		if err != nil {
			return fmt.Errorf("序列化条件失败: %w", err)
		}

		actionsJSON, err := json.Marshal(rule.Actions)
		if err != nil {
			return fmt.Errorf("序列化动作失败: %w", err)
		}

		labelsJSON, err := json.Marshal(rule.Labels)
		if err != nil {
			return fmt.Errorf("序列化标签失败: %w", err)
		}

		annotationsJSON, err := json.Marshal(rule.Annotations)
		if err != nil {
			return fmt.Errorf("序列化注解失败: %w", err)
		}

		query := `
			INSERT INTO rules (
				id, name, description, type, severity, status, enabled, expression,
				conditions, actions, labels, annotations, data_source_id,
				evaluation_interval, for_duration, keep_firing_for, threshold,
				recovery_threshold, no_data_state, exec_err_state,
				created_by, created_at, updated_at
			) VALUES (
				$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23
			)`

		_, err = tx.ExecContext(ctx, query,
			rule.ID, rule.Name, rule.Description, rule.Type, rule.Severity,
			rule.Status, rule.Enabled, rule.Expression, string(conditionsJSON),
			string(actionsJSON), string(labelsJSON), string(annotationsJSON), rule.DataSourceID,
			rule.EvaluationInterval, rule.ForDuration, rule.KeepFiringFor,
			rule.Threshold, rule.RecoveryThreshold, rule.NoDataState, rule.ExecErrState,
			rule.CreatedBy, rule.CreatedAt, rule.UpdatedAt,
		)
		if err != nil {
			return fmt.Errorf("批量创建规则失败: %w", err)
		}
	}

	return tx.Commit()
}

// Activate 激活规则
func (r *ruleRepository) Activate(ctx context.Context, id string) error {
	now := time.Now()
	query := `
		UPDATE rules SET 
			status = $1,
			enabled = true,
			updated_at = $2
		WHERE id = $3 AND deleted_at IS NULL`

	result, err := r.db.ExecContext(ctx, query, models.RuleStatusActive, now, id)
	if err != nil {
		return fmt.Errorf("激活规则失败: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("获取激活结果失败: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("规则不存在或已被删除")
	}

	return nil
}

// Deactivate 停用规则
func (r *ruleRepository) Deactivate(ctx context.Context, id string) error {
	now := time.Now()
	query := `
		UPDATE rules SET 
			status = $1,
			enabled = false,
			updated_at = $2
		WHERE id = $3 AND deleted_at IS NULL`

	result, err := r.db.ExecContext(ctx, query, models.RuleStatusInactive, now, id)
	if err != nil {
		return fmt.Errorf("停用规则失败: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("获取停用结果失败: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("规则不存在或已被删除")
	}

	return nil
}

// BatchUpdate 批量更新规则
func (r *ruleRepository) BatchUpdate(ctx context.Context, rules []*models.Rule) error {
	if len(rules) == 0 {
		return nil
	}

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("开始事务失败: %w", err)
	}
	defer tx.Rollback()

	for _, rule := range rules {
		rule.UpdatedAt = time.Now()

		// 序列化条件、动作、标签和注解
		conditionsJSON, err := json.Marshal(rule.Conditions)
		if err != nil {
			return fmt.Errorf("序列化条件失败: %w", err)
		}

		actionsJSON, err := json.Marshal(rule.Actions)
		if err != nil {
			return fmt.Errorf("序列化动作失败: %w", err)
		}

		labelsJSON, err := json.Marshal(rule.Labels)
		if err != nil {
			return fmt.Errorf("序列化标签失败: %w", err)
		}

		annotationsJSON, err := json.Marshal(rule.Annotations)
		if err != nil {
			return fmt.Errorf("序列化注解失败: %w", err)
		}

		query := `
			UPDATE rules SET 
				name = $1,
				description = $2,
				type = $3,
				severity = $4,
				status = $5,
				enabled = $6,
				expression = $7,
				conditions = $8,
				actions = $9,
				labels = $10,
				annotations = $11,
				data_source_id = $12,
				evaluation_interval = $13,
				for_duration = $14,
				keep_firing_for = $15,
				threshold = $16,
				recovery_threshold = $17,
				no_data_state = $18,
				exec_err_state = $19,
				updated_by = $20,
				updated_at = $21
			WHERE id = $22 AND deleted_at IS NULL`

		_, err = tx.ExecContext(ctx, query,
			rule.Name, rule.Description, rule.Type, rule.Severity,
			rule.Status, rule.Enabled, rule.Expression, string(conditionsJSON),
			string(actionsJSON), string(labelsJSON), string(annotationsJSON), rule.DataSourceID,
			rule.EvaluationInterval, rule.ForDuration, rule.KeepFiringFor,
			rule.Threshold, rule.RecoveryThreshold, rule.NoDataState, rule.ExecErrState,
			rule.UpdatedBy, rule.UpdatedAt, rule.ID,
		)
		if err != nil {
			return fmt.Errorf("批量更新规则失败: %w", err)
		}
	}

	return tx.Commit()
}

// BatchDelete 批量删除规则
func (r *ruleRepository) BatchDelete(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return nil
	}

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("开始事务失败: %w", err)
	}
	defer tx.Rollback()

	now := time.Now()
	for _, id := range ids {
		query := `
			UPDATE rules SET 
				deleted_at = $1,
				updated_at = $1
			WHERE id = $2 AND deleted_at IS NULL`

		_, err := tx.ExecContext(ctx, query, now, id)
		if err != nil {
			return fmt.Errorf("删除规则 %s 失败: %w", id, err)
		}
	}

	return tx.Commit()
}

// BatchActivate 批量激活规则
func (r *ruleRepository) BatchActivate(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return nil
	}

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("开始事务失败: %w", err)
	}
	defer tx.Rollback()

	now := time.Now()
	for _, id := range ids {
		query := `
			UPDATE rules SET 
				status = $1,
				enabled = true,
				updated_at = $2
			WHERE id = $3 AND deleted_at IS NULL`

		_, err := tx.ExecContext(ctx, query, models.RuleStatusActive, now, id)
		if err != nil {
			return fmt.Errorf("激活规则 %s 失败: %w", id, err)
		}
	}

	return tx.Commit()
}

// BatchDeactivate 批量停用规则
func (r *ruleRepository) BatchDeactivate(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return nil
	}

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("开始事务失败: %w", err)
	}
	defer tx.Rollback()

	now := time.Now()
	for _, id := range ids {
		query := `
			UPDATE rules SET 
				status = $1,
				enabled = false,
				updated_at = $2
			WHERE id = $3 AND deleted_at IS NULL`

		_, err := tx.ExecContext(ctx, query, models.RuleStatusInactive, now, id)
		if err != nil {
			return fmt.Errorf("停用规则 %s 失败: %w", id, err)
		}
	}

	return tx.Commit()
}