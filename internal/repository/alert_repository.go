package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"Pulse/internal/models"
)

// alertRepository 告警仓储实现
type alertRepository struct {
	db *sqlx.DB
	tx *sqlx.Tx
}

// NewAlertRepository 创建告警仓储实例
func NewAlertRepository(db *sqlx.DB) AlertRepository {
	return &alertRepository{
		db: db,
	}
}

// NewAlertRepositoryWithTx 创建带事务的告警仓储实例
func NewAlertRepositoryWithTx(tx *sqlx.Tx) AlertRepository {
	return &alertRepository{
		tx: tx,
	}
}

// getExecutor 获取数据库执行器（事务或普通连接）
func (r *alertRepository) getExecutor() sqlx.ExtContext {
	if r.tx != nil {
		return r.tx
	}
	return r.db
}

// Create 创建告警
func (r *alertRepository) Create(ctx context.Context, alert *models.Alert) error {
	// 生成告警ID
	if alert.ID == "" {
		alert.ID = uuid.New().String()
	}

	// 设置创建时间
	now := time.Now()
	alert.CreatedAt = now
	alert.UpdatedAt = now

	// 设置默认状态
	if alert.Status == "" {
		alert.Status = models.AlertStatusFiring
	}

	// 序列化标签和注解
	labelsJSON, err := json.Marshal(alert.Labels)
	if err != nil {
		return fmt.Errorf("序列化标签失败: %w", err)
	}

	annotationsJSON, err := json.Marshal(alert.Annotations)
	if err != nil {
		return fmt.Errorf("序列化注解失败: %w", err)
	}

	query := `
		INSERT INTO alerts (
			id, rule_id, data_source_id, name, description, severity, status, source,
			labels, annotations, value, threshold, expression, starts_at, ends_at,
			last_evaluated_at, evaluation_count, fingerprint, generator_url,
			silence_id, acked_by, acked_at, resolved_by, resolved_at,
			created_at, updated_at
		) VALUES (
			:id, :rule_id, :data_source_id, :name, :description, :severity, :status, :source,
			:labels, :annotations, :value, :threshold, :expression, :starts_at, :ends_at,
			:last_evaluated_at, :evaluation_count, :fingerprint, :generator_url,
			:silence_id, :acked_by, :acked_at, :resolved_by, :resolved_at,
			:created_at, :updated_at
		)`

	// 创建用于数据库插入的结构体
	type alertDB struct {
		*models.Alert
		Labels      string `db:"labels"`
		Annotations string `db:"annotations"`
	}

	alertData := &alertDB{
		Alert:       alert,
		Labels:      string(labelsJSON),
		Annotations: string(annotationsJSON),
	}

	_, err = r.db.NamedExecContext(ctx, query, alertData)
	if err != nil {
		return fmt.Errorf("创建告警失败: %w", err)
	}

	return nil
}

// CleanupResolved 清理已解决的告警
func (r *alertRepository) CleanupResolved(ctx context.Context, before time.Time) (int64, error) {
	query := `
		DELETE FROM alerts 
		WHERE status = $1 AND resolved_at < $2`

	result, err := r.getExecutor().ExecContext(ctx, query, models.AlertStatusResolved, before)
	if err != nil {
		return 0, fmt.Errorf("清理已解决告警失败: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("获取清理行数失败: %w", err)
	}

	return rowsAffected, nil
}

// CleanupExpired 清理过期告警
func (r *alertRepository) CleanupExpired(ctx context.Context) (int64, error) {
	query := `
		DELETE FROM alerts 
		WHERE ends_at IS NOT NULL AND ends_at < NOW() - INTERVAL '7 days'`

	result, err := r.getExecutor().ExecContext(ctx, query)
	if err != nil {
		return 0, fmt.Errorf("清理过期告警失败: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("获取清理行数失败: %w", err)
	}

	return rowsAffected, nil
}

// GetActiveCount 获取活跃告警数量
func (r *alertRepository) GetActiveCount(ctx context.Context) (int64, error) {
	var count int64
	query := `SELECT COUNT(*) FROM alerts WHERE status IN ('firing', 'pending') AND deleted_at IS NULL`
	err := r.db.GetContext(ctx, &count, query)
	if err != nil {
		return 0, fmt.Errorf("获取活跃告警数量失败: %w", err)
	}
	return count, nil
}

// GetCriticalCount 获取严重告警数量
func (r *alertRepository) GetCriticalCount(ctx context.Context) (int64, error) {
	var count int64
	query := `SELECT COUNT(*) FROM alerts WHERE severity = 'critical' AND status IN ('firing', 'pending') AND deleted_at IS NULL`
	err := r.db.GetContext(ctx, &count, query)
	if err != nil {
		return 0, fmt.Errorf("获取严重告警数量失败: %w", err)
	}
	return count, nil
}

// GetByID 根据ID获取告警
func (r *alertRepository) GetByID(ctx context.Context, id string) (*models.Alert, error) {
	var alert models.Alert
	var labelsJSON, annotationsJSON string

	query := `
		SELECT id, rule_id, data_source_id, name, description, severity, status, source,
		       labels, annotations, value, threshold, expression, starts_at, ends_at,
		       last_eval_at, eval_count, fingerprint, generator_url,
		       silence_id, acked_by, acked_at, resolved_by, resolved_at,
		       created_at, updated_at, deleted_at
		FROM alerts 
		WHERE id = $1 AND deleted_at IS NULL`

	row := r.getExecutor().QueryRowxContext(ctx, query, id)
	err := row.Scan(
		&alert.ID, &alert.RuleID, &alert.DataSourceID, &alert.Name, &alert.Description,
		&alert.Severity, &alert.Status, &alert.Source, &labelsJSON, &annotationsJSON,
		&alert.Value, &alert.Threshold, &alert.Expression, &alert.StartsAt, &alert.EndsAt,
		&alert.LastEvalAt, &alert.EvalCount, &alert.Fingerprint, &alert.GeneratorURL,
		&alert.SilenceID, &alert.AckedBy, &alert.AckedAt, &alert.ResolvedBy, &alert.ResolvedAt,
		&alert.CreatedAt, &alert.UpdatedAt, &alert.DeletedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("告警不存在")
		}
		return nil, fmt.Errorf("获取告警失败: %w", err)
	}

	// 反序列化标签和注解
	if labelsJSON != "" {
		err = json.Unmarshal([]byte(labelsJSON), &alert.Labels)
		if err != nil {
			return nil, fmt.Errorf("反序列化标签失败: %w", err)
		}
	}

	if annotationsJSON != "" {
		err = json.Unmarshal([]byte(annotationsJSON), &alert.Annotations)
		if err != nil {
			return nil, fmt.Errorf("反序列化注解失败: %w", err)
		}
	}

	return &alert, nil
}

// Update 更新告警
func (r *alertRepository) Update(ctx context.Context, alert *models.Alert) error {
	alert.UpdatedAt = time.Now()

	// 序列化标签和注解
	labelsJSON, err := json.Marshal(alert.Labels)
	if err != nil {
		return fmt.Errorf("序列化标签失败: %w", err)
	}

	annotationsJSON, err := json.Marshal(alert.Annotations)
	if err != nil {
		return fmt.Errorf("序列化注解失败: %w", err)
	}

	query := `
		UPDATE alerts SET 
			rule_id = $1,
			data_source_id = $2,
			name = $3,
			description = $4,
			severity = $5,
			status = $6,
			source = $7,
			labels = $8,
			annotations = $9,
			value = $10,
			threshold = $11,
			expression = $12,
			starts_at = $13,
			ends_at = $14,
			last_eval_at = $15,
			eval_count = $16,
			fingerprint = $17,
			generator_url = $18,
			silence_id = $19,
			acked_by = $20,
			acked_at = $21,
			resolved_by = $22,
			resolved_at = $23,
			updated_at = $24
		WHERE id = $25 AND deleted_at IS NULL`

	result, err := r.db.ExecContext(ctx, query,
		alert.RuleID, alert.DataSourceID, alert.Name, alert.Description,
		alert.Severity, alert.Status, alert.Source, string(labelsJSON), string(annotationsJSON),
		alert.Value, alert.Threshold, alert.Expression, alert.StartsAt, alert.EndsAt,
		alert.LastEvalAt, alert.EvalCount, alert.Fingerprint, alert.GeneratorURL,
		alert.SilenceID, alert.AckedBy, alert.AckedAt, alert.ResolvedBy, alert.ResolvedAt,
		alert.UpdatedAt, alert.ID,
	)
	if err != nil {
		return fmt.Errorf("更新告警失败: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("获取更新结果失败: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("告警不存在或已被删除")
	}

	return nil
}

// Delete 硬删除告警
func (r *alertRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM alerts WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("删除告警失败: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("获取删除结果失败: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("告警不存在")
	}

	return nil
}

// SoftDelete 软删除告警
func (r *alertRepository) SoftDelete(ctx context.Context, id string) error {
	now := time.Now()
	query := `
		UPDATE alerts SET 
			deleted_at = $1,
			updated_at = $1
		WHERE id = $2 AND deleted_at IS NULL`

	result, err := r.db.ExecContext(ctx, query, now, id)
	if err != nil {
		return fmt.Errorf("软删除告警失败: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("获取删除结果失败: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("告警不存在或已被删除")
	}

	return nil
}

// List 获取告警列表
func (r *alertRepository) List(ctx context.Context, filter *models.AlertFilter) (*models.AlertList, error) {
	if filter == nil {
		filter = &models.AlertFilter{Page: 1, PageSize: 20}
	}

	// 设置默认值
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 {
		filter.PageSize = 20
	}
	if filter.PageSize > 100 {
		filter.PageSize = 100
	}

	// 构建查询条件
	var conditions []string
	var args []interface{}
	argIndex := 1

	conditions = append(conditions, "deleted_at IS NULL")

	if filter.RuleID != nil {
		conditions = append(conditions, fmt.Sprintf("rule_id = $%d", argIndex))
		args = append(args, *filter.RuleID)
		argIndex++
	}

	if filter.DataSourceID != nil {
		conditions = append(conditions, fmt.Sprintf("data_source_id = $%d", argIndex))
		args = append(args, *filter.DataSourceID)
		argIndex++
	}

	if filter.Severity != nil {
		conditions = append(conditions, fmt.Sprintf("severity = $%d", argIndex))
		args = append(args, *filter.Severity)
		argIndex++
	}

	if filter.Status != nil {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argIndex))
		args = append(args, *filter.Status)
		argIndex++
	}

	if filter.Source != nil {
		conditions = append(conditions, fmt.Sprintf("source = $%d", argIndex))
		args = append(args, *filter.Source)
		argIndex++
	}

	if filter.Keyword != nil && *filter.Keyword != "" {
		keyword := "%" + *filter.Keyword + "%"
		conditions = append(conditions, fmt.Sprintf("(name ILIKE $%d OR description ILIKE $%d)", argIndex, argIndex))
		args = append(args, keyword)
		argIndex++
	}

	if filter.StartTime != nil {
		conditions = append(conditions, fmt.Sprintf("starts_at >= $%d", argIndex))
		args = append(args, *filter.StartTime)
		argIndex++
	}

	if filter.EndTime != nil {
		conditions = append(conditions, fmt.Sprintf("starts_at <= $%d", argIndex))
		args = append(args, *filter.EndTime)
		argIndex++
	}

	// 处理标签过滤
	if len(filter.Labels) > 0 {
		for key, value := range filter.Labels {
			conditions = append(conditions, fmt.Sprintf("labels ->> $%d = $%d", argIndex, argIndex+1))
			args = append(args, key, value)
			argIndex += 2
		}
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// 获取总数
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM alerts %s", whereClause)
	var total int64
	err := r.db.GetContext(ctx, &total, countQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("获取告警总数失败: %w", err)
	}

	// 获取告警列表
	offset := (filter.Page - 1) * filter.PageSize
	listQuery := fmt.Sprintf(`
		SELECT id, rule_id, data_source_id, name, description, severity, status, source,
		       labels, annotations, value, threshold, expression, starts_at, ends_at,
		       last_eval_at, eval_count, fingerprint, generator_url,
		       silence_id, acked_by, acked_at, resolved_by, resolved_at,
		       created_at, updated_at
		FROM alerts %s
		ORDER BY starts_at DESC
		LIMIT $%d OFFSET $%d`, whereClause, argIndex, argIndex+1)

	args = append(args, filter.PageSize, offset)

	rows, err := r.db.QueryContext(ctx, listQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("获取告警列表失败: %w", err)
	}
	defer rows.Close()

	var alerts []*models.Alert
	for rows.Next() {
		var alert models.Alert
		var labelsJSON, annotationsJSON string

		err := rows.Scan(
			&alert.ID, &alert.RuleID, &alert.DataSourceID, &alert.Name, &alert.Description,
			&alert.Severity, &alert.Status, &alert.Source, &labelsJSON, &annotationsJSON,
			&alert.Value, &alert.Threshold, &alert.Expression, &alert.StartsAt, &alert.EndsAt,
			&alert.LastEvalAt, &alert.EvalCount, &alert.Fingerprint, &alert.GeneratorURL,
			&alert.SilenceID, &alert.AckedBy, &alert.AckedAt, &alert.ResolvedBy, &alert.ResolvedAt,
			&alert.CreatedAt, &alert.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("扫描告警数据失败: %w", err)
		}

		// 反序列化标签和注解
		if labelsJSON != "" {
			err = json.Unmarshal([]byte(labelsJSON), &alert.Labels)
			if err != nil {
				return nil, fmt.Errorf("反序列化标签失败: %w", err)
			}
		}

		if annotationsJSON != "" {
			err = json.Unmarshal([]byte(annotationsJSON), &alert.Annotations)
			if err != nil {
				return nil, fmt.Errorf("反序列化注解失败: %w", err)
			}
		}

		alerts = append(alerts, &alert)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("遍历告警数据失败: %w", err)
	}

	totalPages := int((total + int64(filter.PageSize) - 1) / int64(filter.PageSize))

	return &models.AlertList{
		Alerts:     alerts,
		Total:      total,
		Page:       filter.Page,
		PageSize:   filter.PageSize,
		TotalPages: totalPages,
	}, nil
}

// Count 获取告警总数
func (r *alertRepository) Count(ctx context.Context, filter *models.AlertFilter) (int64, error) {
	var conditions []string
	var args []interface{}
	argIndex := 1

	conditions = append(conditions, "deleted_at IS NULL")

	if filter != nil {
		if filter.RuleID != nil {
			conditions = append(conditions, fmt.Sprintf("rule_id = $%d", argIndex))
			args = append(args, *filter.RuleID)
			argIndex++
		}

		if filter.DataSourceID != nil {
			conditions = append(conditions, fmt.Sprintf("data_source_id = $%d", argIndex))
			args = append(args, *filter.DataSourceID)
			argIndex++
		}

		if filter.Severity != nil {
			conditions = append(conditions, fmt.Sprintf("severity = $%d", argIndex))
			args = append(args, *filter.Severity)
			argIndex++
		}

		if filter.Status != nil {
			conditions = append(conditions, fmt.Sprintf("status = $%d", argIndex))
			args = append(args, *filter.Status)
			argIndex++
		}

		if filter.Source != nil {
			conditions = append(conditions, fmt.Sprintf("source = $%d", argIndex))
			args = append(args, *filter.Source)
			argIndex++
		}

		if filter.StartTime != nil {
			conditions = append(conditions, fmt.Sprintf("starts_at >= $%d", argIndex))
			args = append(args, *filter.StartTime)
			argIndex++
		}

		if filter.EndTime != nil {
			conditions = append(conditions, fmt.Sprintf("starts_at <= $%d", argIndex))
			args = append(args, *filter.EndTime)
			argIndex++
		}

		// 处理标签过滤
		if len(filter.Labels) > 0 {
			for key, value := range filter.Labels {
				conditions = append(conditions, fmt.Sprintf("labels ->> $%d = $%d", argIndex, argIndex+1))
				args = append(args, key, value)
				argIndex += 2
			}
		}
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	query := fmt.Sprintf("SELECT COUNT(*) FROM alerts %s", whereClause)
	var count int64
	err := r.db.GetContext(ctx, &count, query, args...)
	if err != nil {
		return 0, fmt.Errorf("获取告警总数失败: %w", err)
	}

	return count, nil
}

// Exists 检查告警是否存在
func (r *alertRepository) Exists(ctx context.Context, id string) (bool, error) {
	var count int
	query := `SELECT COUNT(*) FROM alerts WHERE id = $1 AND deleted_at IS NULL`
	err := r.db.GetContext(ctx, &count, query, id)
	if err != nil {
		return false, fmt.Errorf("检查告警存在性失败: %w", err)
	}
	return count > 0, nil
}

// GetByFingerprint 根据指纹获取告警
func (r *alertRepository) GetByFingerprint(ctx context.Context, fingerprint string) (*models.Alert, error) {
	var alert models.Alert
	var labelsJSON, annotationsJSON string

	query := `
		SELECT id, rule_id, data_source_id, name, description, severity, status, source,
		       labels, annotations, value, threshold, expression, starts_at, ends_at,
		       last_eval_at, eval_count, fingerprint, generator_url,
		       silence_id, acked_by, acked_at, resolved_by, resolved_at,
		       created_at, updated_at, deleted_at
		FROM alerts 
		WHERE fingerprint = $1 AND deleted_at IS NULL`

	row := r.getExecutor().QueryRowxContext(ctx, query, fingerprint)
	err := row.Scan(
		&alert.ID, &alert.RuleID, &alert.DataSourceID, &alert.Name, &alert.Description,
		&alert.Severity, &alert.Status, &alert.Source, &labelsJSON, &annotationsJSON,
		&alert.Value, &alert.Threshold, &alert.Expression, &alert.StartsAt, &alert.EndsAt,
		&alert.LastEvalAt, &alert.EvalCount, &alert.Fingerprint, &alert.GeneratorURL,
		&alert.SilenceID, &alert.AckedBy, &alert.AckedAt, &alert.ResolvedBy, &alert.ResolvedAt,
		&alert.CreatedAt, &alert.UpdatedAt, &alert.DeletedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("告警不存在")
		}
		return nil, fmt.Errorf("获取告警失败: %w", err)
	}

	// 反序列化标签和注解
	if labelsJSON != "" {
		err = json.Unmarshal([]byte(labelsJSON), &alert.Labels)
		if err != nil {
			return nil, fmt.Errorf("反序列化标签失败: %w", err)
		}
	}

	if annotationsJSON != "" {
		err = json.Unmarshal([]byte(annotationsJSON), &alert.Annotations)
		if err != nil {
			return nil, fmt.Errorf("反序列化注解失败: %w", err)
		}
	}

	return &alert, nil
}

// Acknowledge 确认告警
func (r *alertRepository) Acknowledge(ctx context.Context, id, userID string, comment *string) error {
	now := time.Now()
	query := `
		UPDATE alerts SET 
			status = $1,
			acked_by = $2,
			acked_at = $3,
			updated_at = $3
		WHERE id = $4 AND deleted_at IS NULL`

	result, err := r.db.ExecContext(ctx, query, models.AlertStatusAcked, userID, now, id)
	if err != nil {
		return fmt.Errorf("确认告警失败: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("获取确认结果失败: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("告警不存在或已被删除")
	}

	// 如果有评论，记录到历史中
	if comment != nil && *comment != "" {
		// 这里可以添加历史记录逻辑
	}

	return nil
}

// Resolve 解决告警
func (r *alertRepository) Resolve(ctx context.Context, id, userID string, comment *string) error {
	now := time.Now()
	query := `
		UPDATE alerts SET 
			status = $1,
			resolved_by = $2,
			resolved_at = $3,
			ends_at = $3,
			updated_at = $3
		WHERE id = $4 AND deleted_at IS NULL`

	result, err := r.db.ExecContext(ctx, query, models.AlertStatusResolved, userID, now, id)
	if err != nil {
		return fmt.Errorf("解决告警失败: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("获取解决结果失败: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("告警不存在或已被删除")
	}

	// 如果有评论，记录到历史中
	if comment != nil && *comment != "" {
		// 这里可以添加历史记录逻辑
	}

	return nil
}

// Silence 静默告警
func (r *alertRepository) Silence(ctx context.Context, id, silenceID string, duration time.Duration) error {
	now := time.Now()
	query := `
		UPDATE alerts SET 
			status = $1,
			silence_id = $2,
			updated_at = $3
		WHERE id = $4 AND deleted_at IS NULL`

	result, err := r.db.ExecContext(ctx, query, models.AlertStatusSilenced, silenceID, now, id)
	if err != nil {
		return fmt.Errorf("静默告警失败: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("获取静默结果失败: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("告警不存在或已被删除")
	}

	return nil
}

// Unsilence 取消静默告警
func (r *alertRepository) Unsilence(ctx context.Context, id string) error {
	now := time.Now()
	query := `
		UPDATE alerts SET 
			status = $1,
			silence_id = NULL,
			updated_at = $2
		WHERE id = $3 AND deleted_at IS NULL`

	result, err := r.db.ExecContext(ctx, query, models.AlertStatusFiring, now, id)
	if err != nil {
		return fmt.Errorf("取消静默告警失败: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("获取取消静默结果失败: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("告警不存在或已被删除")
	}

	return nil
}

// GetStats 获取告警统计信息
func (r *alertRepository) GetStats(ctx context.Context, filter *models.AlertFilter) (*models.AlertStats, error) {
	var conditions []string
	var args []interface{}
	argIndex := 1

	conditions = append(conditions, "deleted_at IS NULL")

	if filter != nil {
		if filter.RuleID != nil {
			conditions = append(conditions, fmt.Sprintf("rule_id = $%d", argIndex))
			args = append(args, *filter.RuleID)
			argIndex++
		}

		if filter.DataSourceID != nil {
			conditions = append(conditions, fmt.Sprintf("data_source_id = $%d", argIndex))
			args = append(args, *filter.DataSourceID)
			argIndex++
		}

		if filter.StartTime != nil {
			conditions = append(conditions, fmt.Sprintf("starts_at >= $%d", argIndex))
			args = append(args, *filter.StartTime)
			argIndex++
		}

		if filter.EndTime != nil {
			conditions = append(conditions, fmt.Sprintf("starts_at <= $%d", argIndex))
			args = append(args, *filter.EndTime)
			argIndex++
		}
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// 获取总数
	totalQuery := fmt.Sprintf("SELECT COUNT(*) FROM alerts %s", whereClause)
	var total int64
	err := r.db.GetContext(ctx, &total, totalQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("获取告警总数失败: %w", err)
	}

	// 按严重级别统计
	bySeverityQuery := fmt.Sprintf(`
		SELECT severity, COUNT(*) 
		FROM alerts %s 
		GROUP BY severity`, whereClause)

	bySeverityRows, err := r.db.QueryContext(ctx, bySeverityQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("按严重级别统计失败: %w", err)
	}
	defer bySeverityRows.Close()

	bySeverity := make(map[models.AlertSeverity]int64)
	for bySeverityRows.Next() {
		var severity models.AlertSeverity
		var count int64
		err := bySeverityRows.Scan(&severity, &count)
		if err != nil {
			return nil, fmt.Errorf("扫描严重级别统计失败: %w", err)
		}
		bySeverity[severity] = count
	}

	// 按状态统计
	byStatusQuery := fmt.Sprintf(`
		SELECT status, COUNT(*) 
		FROM alerts %s 
		GROUP BY status`, whereClause)

	byStatusRows, err := r.db.QueryContext(ctx, byStatusQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("按状态统计失败: %w", err)
	}
	defer byStatusRows.Close()

	byStatus := make(map[models.AlertStatus]int64)
	for byStatusRows.Next() {
		var status models.AlertStatus
		var count int64
		err := byStatusRows.Scan(&status, &count)
		if err != nil {
			return nil, fmt.Errorf("扫描状态统计失败: %w", err)
		}
		byStatus[status] = count
	}

	// 按来源统计
	bySourceQuery := fmt.Sprintf(`
		SELECT source, COUNT(*) 
		FROM alerts %s 
		GROUP BY source`, whereClause)

	bySourceRows, err := r.db.QueryContext(ctx, bySourceQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("按来源统计失败: %w", err)
	}
	defer bySourceRows.Close()

	bySource := make(map[models.AlertSource]int64)
	for bySourceRows.Next() {
		var source models.AlertSource
		var count int64
		err := bySourceRows.Scan(&source, &count)
		if err != nil {
			return nil, fmt.Errorf("扫描来源统计失败: %w", err)
		}
		bySource[source] = count
	}

	return &models.AlertStats{
		Total:      total,
		BySeverity: bySeverity,
		ByStatus:   byStatus,
		BySource:   bySource,
		Trend:      []*models.AlertTrendPoint{}, // 趋势数据需要单独实现
	}, nil
}

// GetTrend 获取告警趋势数据
func (r *alertRepository) GetTrend(ctx context.Context, start, end time.Time, interval string) ([]*models.AlertTrendPoint, error) {
	conditions := []string{"deleted_at IS NULL", "starts_at >= $1", "starts_at <= $2"}
	args := []interface{}{start, end}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// 根据间隔类型构建时间分组
	var timeGroup string
	switch interval {
	case "hour":
		timeGroup = "date_trunc('hour', starts_at)"
	case "day":
		timeGroup = "date_trunc('day', starts_at)"
	case "week":
		timeGroup = "date_trunc('week', starts_at)"
	case "month":
		timeGroup = "date_trunc('month', starts_at)"
	default:
		timeGroup = "date_trunc('hour', starts_at)"
	}

	query := fmt.Sprintf(`
		SELECT %s as timestamp, COUNT(*) as count
		FROM alerts %s
		GROUP BY %s
		ORDER BY timestamp`, timeGroup, whereClause, timeGroup)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("获取告警趋势失败: %w", err)
	}
	defer rows.Close()

	var trend []*models.AlertTrendPoint
	for rows.Next() {
		var point models.AlertTrendPoint
		err := rows.Scan(&point.Timestamp, &point.Count)
		if err != nil {
			return nil, fmt.Errorf("扫描趋势数据失败: %w", err)
		}
		trend = append(trend, &point)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("遍历趋势数据失败: %w", err)
	}

	return trend, nil
}

// BatchCreate 批量创建告警
func (r *alertRepository) BatchCreate(ctx context.Context, alerts []*models.Alert) error {
	if len(alerts) == 0 {
		return nil
	}

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("开始事务失败: %w", err)
	}
	defer tx.Rollback()

	for _, alert := range alerts {
		if alert.ID == "" {
			alert.ID = uuid.New().String()
		}

		now := time.Now()
		alert.CreatedAt = now
		alert.UpdatedAt = now

		if alert.Status == "" {
			alert.Status = models.AlertStatusFiring
		}

		// 序列化标签和注解
		labelsJSON, err := json.Marshal(alert.Labels)
		if err != nil {
			return fmt.Errorf("序列化标签失败: %w", err)
		}

		annotationsJSON, err := json.Marshal(alert.Annotations)
		if err != nil {
			return fmt.Errorf("序列化注解失败: %w", err)
		}

		query := `
			INSERT INTO alerts (
				id, rule_id, data_source_id, name, description, severity, status, source,
				labels, annotations, value, threshold, expression, starts_at, ends_at,
				last_eval_at, eval_count, fingerprint, generator_url,
				silence_id, acked_by, acked_at, resolved_by, resolved_at,
				created_at, updated_at
			) VALUES (
				$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15,
				$16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26
			)`

		_, err = tx.ExecContext(ctx, query,
			alert.ID, alert.RuleID, alert.DataSourceID, alert.Name, alert.Description,
			alert.Severity, alert.Status, alert.Source, string(labelsJSON), string(annotationsJSON),
			alert.Value, alert.Threshold, alert.Expression, alert.StartsAt, alert.EndsAt,
			alert.LastEvalAt, alert.EvalCount, alert.Fingerprint, alert.GeneratorURL,
			alert.SilenceID, alert.AckedBy, alert.AckedAt, alert.ResolvedBy, alert.ResolvedAt,
			alert.CreatedAt, alert.UpdatedAt,
		)
		if err != nil {
			return fmt.Errorf("批量创建告警失败: %w", err)
		}
	}

	return tx.Commit()
}

// BatchUpdate 批量更新告警
func (r *alertRepository) BatchUpdate(ctx context.Context, alerts []*models.Alert) error {
	if len(alerts) == 0 {
		return nil
	}

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("开始事务失败: %w", err)
	}
	defer tx.Rollback()

	for _, alert := range alerts {
		alert.UpdatedAt = time.Now()

		// 序列化标签和注解
		labelsJSON, err := json.Marshal(alert.Labels)
		if err != nil {
			return fmt.Errorf("序列化标签失败: %w", err)
		}

		annotationsJSON, err := json.Marshal(alert.Annotations)
		if err != nil {
			return fmt.Errorf("序列化注解失败: %w", err)
		}

		query := `
			UPDATE alerts SET 
				rule_id = $1,
				data_source_id = $2,
				name = $3,
				description = $4,
				severity = $5,
				status = $6,
				source = $7,
				labels = $8,
				annotations = $9,
				value = $10,
				threshold = $11,
				expression = $12,
				starts_at = $13,
				ends_at = $14,
				last_eval_at = $15,
				eval_count = $16,
				fingerprint = $17,
				generator_url = $18,
				silence_id = $19,
				acked_by = $20,
				acked_at = $21,
				resolved_by = $22,
				resolved_at = $23,
				updated_at = $24
			WHERE id = $25 AND deleted_at IS NULL`

		_, err = tx.ExecContext(ctx, query,
			alert.RuleID, alert.DataSourceID, alert.Name, alert.Description,
			alert.Severity, alert.Status, alert.Source, string(labelsJSON), string(annotationsJSON),
			alert.Value, alert.Threshold, alert.Expression, alert.StartsAt, alert.EndsAt,
			alert.LastEvalAt, alert.EvalCount, alert.Fingerprint, alert.GeneratorURL,
			alert.SilenceID, alert.AckedBy, alert.AckedAt, alert.ResolvedBy, alert.ResolvedAt,
			alert.UpdatedAt, alert.ID,
		)
		if err != nil {
			return fmt.Errorf("批量更新告警失败: %w", err)
		}
	}

	return tx.Commit()
}

// BatchDelete 批量删除告警
func (r *alertRepository) BatchDelete(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return nil
	}

	// 构建占位符
	placeholders := make([]string, len(ids))
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = id
	}

	query := fmt.Sprintf(`
		UPDATE alerts SET 
			deleted_at = NOW(),
			updated_at = NOW()
		WHERE id IN (%s) AND deleted_at IS NULL`,
		strings.Join(placeholders, ", "))

	result, err := r.getExecutor().ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("批量删除告警失败: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("获取删除行数失败: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("没有找到要删除的告警")
	}

	return nil
}

// GetHistory 获取告警历史记录
func (r *alertRepository) GetHistory(ctx context.Context, alertID string) ([]*models.AlertHistory, error) {
	var histories []*models.AlertHistory
	query := `
		SELECT id, alert_id, action, old_value, new_value, user_id, comment, created_at
		FROM alert_histories 
		WHERE alert_id = $1 
		ORDER BY created_at DESC`

	err := r.db.SelectContext(ctx, &histories, query, alertID)
	if err != nil {
		return nil, fmt.Errorf("获取告警历史记录失败: %w", err)
	}

	return histories, nil
}

// AddHistory 添加告警历史记录
func (r *alertRepository) AddHistory(ctx context.Context, history *models.AlertHistory) error {
	// 生成历史记录ID
	if history.ID == "" {
		history.ID = uuid.New().String()
	}

	// 设置创建时间
	history.CreatedAt = time.Now()

	// 序列化JSON字段
	oldValueJSON, err := json.Marshal(history.OldValue)
	if err != nil {
		return fmt.Errorf("序列化旧值失败: %w", err)
	}

	newValueJSON, err := json.Marshal(history.NewValue)
	if err != nil {
		return fmt.Errorf("序列化新值失败: %w", err)
	}

	query := `
		INSERT INTO alert_histories (
			id, alert_id, action, old_value, new_value, user_id, comment, created_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8
		)`

	_, err = r.getExecutor().ExecContext(ctx, query,
		history.ID,
		history.AlertID,
		history.Action,
		string(oldValueJSON),
		string(newValueJSON),
		history.UserID,
		history.Comment,
		history.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("添加告警历史记录失败: %w", err)
	}

	return nil
}

// BatchAcknowledge 批量确认告警
func (r *alertRepository) BatchAcknowledge(ctx context.Context, ids []string, userID string, comment *string) error {
	if len(ids) == 0 {
		return nil
	}

	// 构建占位符
	placeholders := make([]string, len(ids))
	args := make([]interface{}, len(ids)+3)
	now := time.Now()

	args[0] = models.AlertStatusAcked
	args[1] = userID
	args[2] = now

	for i, id := range ids {
		placeholders[i] = fmt.Sprintf("$%d", i+4)
		args[i+3] = id
	}

	query := fmt.Sprintf(`
		UPDATE alerts SET 
			status = $1,
			acked_by = $2,
			acked_at = $3,
			updated_at = $3
		WHERE id IN (%s) AND deleted_at IS NULL`,
		strings.Join(placeholders, ", "))

	result, err := r.getExecutor().ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("批量确认告警失败: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("获取确认行数失败: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("没有找到要确认的告警")
	}

	// 为每个告警添加历史记录
	for _, id := range ids {
		history := &models.AlertHistory{
			AlertID: id,
			Action:  "acknowledged",
			UserID:  &userID,
			Comment: comment,
		}
		if err := r.AddHistory(ctx, history); err != nil {
			// 记录历史失败不影响主操作
			continue
		}
	}

	return nil
}

// BatchResolve 批量解决告警
func (r *alertRepository) BatchResolve(ctx context.Context, ids []string, userID string, comment *string) error {
	if len(ids) == 0 {
		return nil
	}

	// 构建占位符
	placeholders := make([]string, len(ids))
	args := make([]interface{}, len(ids)+3)
	now := time.Now()

	args[0] = models.AlertStatusResolved
	args[1] = userID
	args[2] = now

	for i, id := range ids {
		placeholders[i] = fmt.Sprintf("$%d", i+4)
		args[i+3] = id
	}

	query := fmt.Sprintf(`
		UPDATE alerts SET 
			status = $1,
			resolved_by = $2,
			resolved_at = $3,
			ends_at = $3,
			updated_at = $3
		WHERE id IN (%s) AND deleted_at IS NULL`,
		strings.Join(placeholders, ", "))

	result, err := r.getExecutor().ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("批量解决告警失败: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("获取解决行数失败: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("没有找到要解决的告警")
	}

	// 为每个告警添加历史记录
	for _, id := range ids {
		history := &models.AlertHistory{
			AlertID: id,
			Action:  "resolved",
			UserID:  &userID,
			Comment: comment,
		}
		if err := r.AddHistory(ctx, history); err != nil {
			// 记录历史失败不影响主操作
			continue
		}
	}

	return nil
}