package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"pulse/internal/models"
)

// webhookRepository Webhook仓储实现
type webhookRepository struct {
	db *sqlx.DB
	tx *sqlx.Tx
}

// NewWebhookRepository 创建新的Webhook仓储
func NewWebhookRepository(db *sqlx.DB) WebhookRepository {
	return &webhookRepository{db: db}
}

// NewWebhookRepositoryWithTx 创建带事务的Webhook仓储
func NewWebhookRepositoryWithTx(tx *sqlx.Tx) WebhookRepository {
	return &webhookRepository{tx: tx}
}

// getDB 获取数据库连接
func (r *webhookRepository) getDB() interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
} {
	if r.tx != nil {
		return r.tx
	}
	return r.db
}

// Create 创建Webhook
func (r *webhookRepository) Create(ctx context.Context, webhook *models.Webhook) error {
	webhook.ID = uuid.New()
	webhook.CreatedAt = time.Now()
	webhook.UpdatedAt = time.Now()
	
	// 设置默认值
	if webhook.Timeout == 0 {
		webhook.Timeout = 30
	}
	if webhook.RetryCount == 0 {
		webhook.RetryCount = 3
	}
	if webhook.Status == "" {
		webhook.Status = models.WebhookStatusActive
	}
	
	// 序列化JSON字段
	eventsJSON, _ := json.Marshal(webhook.Events)
	headersJSON, _ := json.Marshal(webhook.Headers)
	
	query := `
		INSERT INTO webhooks (id, name, url, secret, events, headers, timeout, retry_count, status, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`
	
	_, err := r.getDB().ExecContext(ctx, query,
		webhook.ID, webhook.Name, webhook.URL, webhook.Secret,
		string(eventsJSON), string(headersJSON),
		webhook.Timeout, webhook.RetryCount, webhook.Status,
		webhook.CreatedBy, webhook.CreatedAt, webhook.UpdatedAt,
	)
	
	return err
}

// GetByID 根据ID获取Webhook
func (r *webhookRepository) GetByID(ctx context.Context, id string) (*models.Webhook, error) {
	// Convert string to UUID
	webhookID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid webhook ID format: %w", err)
	}
	query := `
		SELECT id, name, url, secret, events, headers, timeout, retry_count,
		       status, last_triggered, created_by, created_at, updated_at
		FROM webhooks
		WHERE id = $1
	`
	
	var webhook models.Webhook
	var eventsJSON, headersJSON string
	
	err = r.getDB().QueryRowContext(ctx, query, webhookID).Scan(
		&webhook.ID, &webhook.Name, &webhook.URL, &webhook.Secret,
		&eventsJSON, &headersJSON, &webhook.Timeout, &webhook.RetryCount,
		&webhook.Status, &webhook.LastTriggered, &webhook.CreatedBy,
		&webhook.CreatedAt, &webhook.UpdatedAt,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	
	// 反序列化JSON字段
	json.Unmarshal([]byte(eventsJSON), &webhook.Events)
	json.Unmarshal([]byte(headersJSON), &webhook.Headers)
	
	return &webhook, nil
}

// Update 更新Webhook
func (r *webhookRepository) Update(ctx context.Context, webhook *models.Webhook) error {
	webhook.UpdatedAt = time.Now()
	
	// 序列化Events
	eventsJSON, err := json.Marshal(webhook.Events)
	if err != nil {
		return fmt.Errorf("序列化事件列表失败: %w", err)
	}
	
	// 序列化Headers
	headersJSON, err := json.Marshal(webhook.Headers)
	if err != nil {
		return fmt.Errorf("序列化请求头失败: %w", err)
	}
	
	query := `
		UPDATE webhooks SET
			name = $2,
			url = $3,
			secret = $4,
			events = $5,
			headers = $6,
			timeout = $7,
			retry_count = $8,
			status = $9,
			updated_at = $10
		WHERE id = $1 AND deleted_at IS NULL
	`
	
	_, err = r.getDB().ExecContext(ctx, query,
		webhook.ID, webhook.Name, webhook.URL, webhook.Secret,
		string(eventsJSON), string(headersJSON), webhook.Timeout, webhook.RetryCount,
		webhook.Status, webhook.UpdatedAt,
	)
	
	return err
}

// Delete 删除Webhook
func (r *webhookRepository) Delete(ctx context.Context, id string) error {
	webhookID, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("无效的Webhook ID: %w", err)
	}
	
	query := `DELETE FROM webhooks WHERE id = $1`
	_, err = r.getDB().ExecContext(ctx, query, webhookID)
	return err
}

// SoftDelete 软删除Webhook
func (r *webhookRepository) SoftDelete(ctx context.Context, id string) error {
	webhookID, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("无效的Webhook ID: %w", err)
	}
	
	query := `UPDATE webhooks SET deleted_at = $2 WHERE id = $1`
	_, err = r.getDB().ExecContext(ctx, query, webhookID, time.Now())
	return err
}

// List 获取Webhook列表
func (r *webhookRepository) List(ctx context.Context, filter *models.WebhookFilter) (*models.WebhookList, error) {
	query := `
		SELECT id, name, url, secret, events, headers, timeout, retry_count,
		       status, last_triggered, created_by, created_at, updated_at
		FROM webhooks
		WHERE deleted_at IS NULL
	`
	args := []interface{}{}
	argIndex := 0
	
	if filter.Name != nil {
		argIndex++
		query += fmt.Sprintf(" AND name ILIKE $%d", argIndex)
		args = append(args, "%"+*filter.Name+"%")
	}
	
	if filter.Status != nil {
		argIndex++
		query += fmt.Sprintf(" AND status = $%d", argIndex)
		args = append(args, *filter.Status)
	}
	
	if filter.CreatedBy != nil {
		argIndex++
		query += fmt.Sprintf(" AND created_by = $%d", argIndex)
		args = append(args, *filter.CreatedBy)
	}
	
	// 计算总数
	countQuery := "SELECT COUNT(*) FROM (" + query + ") as count_query"
	var total int64
	err := r.getDB().QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("获取Webhook总数失败: %w", err)
	}
	
	// 添加排序和分页
	query += " ORDER BY created_at DESC"
	if filter.PageSize > 0 {
		argIndex++
		query += fmt.Sprintf(" LIMIT $%d", argIndex)
		args = append(args, filter.PageSize)
		
		if filter.Page > 0 {
			argIndex++
			query += fmt.Sprintf(" OFFSET $%d", argIndex)
			args = append(args, (filter.Page-1)*filter.PageSize)
		}
	}
	
	rows, err := r.getDB().QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("查询Webhook列表失败: %w", err)
	}
	defer rows.Close()
	
	var webhooks []*models.Webhook
	for rows.Next() {
		var webhook models.Webhook
		var eventsJSON, headersJSON string
		
		err := rows.Scan(
			&webhook.ID, &webhook.Name, &webhook.URL, &webhook.Secret,
			&eventsJSON, &headersJSON, &webhook.Timeout, &webhook.RetryCount,
			&webhook.Status, &webhook.LastTriggered, &webhook.CreatedBy,
			&webhook.CreatedAt, &webhook.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("扫描Webhook数据失败: %w", err)
		}
		
		// 反序列化Events
		if err := json.Unmarshal([]byte(eventsJSON), &webhook.Events); err != nil {
			return nil, fmt.Errorf("反序列化事件列表失败: %w", err)
		}
		
		// 反序列化Headers
		if err := json.Unmarshal([]byte(headersJSON), &webhook.Headers); err != nil {
			return nil, fmt.Errorf("反序列化请求头失败: %w", err)
		}
		
		webhooks = append(webhooks, &webhook)
	}
	
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("遍历Webhook数据失败: %w", err)
	}
	
	return &models.WebhookList{
		Webhooks: webhooks,
		Total:    total,
		Page:     filter.Page,
		PageSize: filter.PageSize,
	}, nil
}

// Count 获取Webhook总数
func (r *webhookRepository) Count(ctx context.Context, filter *models.WebhookFilter) (int64, error) {
	query := "SELECT COUNT(*) FROM webhooks WHERE deleted_at IS NULL"
	args := []interface{}{}
	argIndex := 0
	
	if filter.Name != nil {
		argIndex++
		query += fmt.Sprintf(" AND name ILIKE $%d", argIndex)
		args = append(args, "%"+*filter.Name+"%")
	}
	
	if filter.Status != nil {
		argIndex++
		query += fmt.Sprintf(" AND status = $%d", argIndex)
		args = append(args, *filter.Status)
	}
	
	if filter.CreatedBy != nil {
		argIndex++
		query += fmt.Sprintf(" AND created_by = $%d", argIndex)
		args = append(args, *filter.CreatedBy)
	}
	
	var count int64
	err := r.getDB().QueryRowContext(ctx, query, args...).Scan(&count)
	return count, err
}

// Exists 检查Webhook是否存在
func (r *webhookRepository) Exists(ctx context.Context, id string) (bool, error) {
	webhookID, err := uuid.Parse(id)
	if err != nil {
		return false, fmt.Errorf("无效的Webhook ID: %w", err)
	}
	
	query := "SELECT COUNT(*) FROM webhooks WHERE id = $1 AND deleted_at IS NULL"
	var count int
	err = r.getDB().QueryRowContext(ctx, query, webhookID).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// GetByURL 根据URL获取Webhook
func (r *webhookRepository) GetByURL(ctx context.Context, url string) (*models.Webhook, error) {
	query := `
		SELECT id, name, url, secret, events, headers, timeout, retry_count,
		       status, last_triggered, created_by, created_at, updated_at
		FROM webhooks
		WHERE url = $1 AND deleted_at IS NULL
	`
	
	var webhook models.Webhook
	var eventsJSON, headersJSON string
	
	err := r.getDB().QueryRowContext(ctx, query, url).Scan(
		&webhook.ID, &webhook.Name, &webhook.URL, &webhook.Secret,
		&eventsJSON, &headersJSON, &webhook.Timeout, &webhook.RetryCount,
		&webhook.Status, &webhook.LastTriggered, &webhook.CreatedBy,
		&webhook.CreatedAt, &webhook.UpdatedAt,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	
	// 反序列化Events
	if err := json.Unmarshal([]byte(eventsJSON), &webhook.Events); err != nil {
		return nil, fmt.Errorf("反序列化事件列表失败: %w", err)
	}
	
	// 反序列化Headers
	if err := json.Unmarshal([]byte(headersJSON), &webhook.Headers); err != nil {
		return nil, fmt.Errorf("反序列化请求头失败: %w", err)
	}
	
	return &webhook, nil
}

// UpdateStatus 更新Webhook状态
func (r *webhookRepository) UpdateStatus(ctx context.Context, id string, status models.WebhookStatus) error {
	webhookID, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("无效的Webhook ID: %w", err)
	}
	
	query := `UPDATE webhooks SET status = $2, updated_at = $3 WHERE id = $1 AND deleted_at IS NULL`
	_, err = r.getDB().ExecContext(ctx, query, webhookID, status, time.Now())
	return err
}

// Enable 启用Webhook
func (r *webhookRepository) Enable(ctx context.Context, id string) error {
	return r.UpdateStatus(ctx, id, models.WebhookStatusActive)
}

// Disable 禁用Webhook
func (r *webhookRepository) Disable(ctx context.Context, id string) error {
	return r.UpdateStatus(ctx, id, models.WebhookStatusInactive)
}

// CreateLog 创建Webhook日志
func (r *webhookRepository) CreateLog(ctx context.Context, log *models.WebhookLog) error {
	if log.ID == uuid.Nil {
		log.ID = uuid.New()
	}
	
	log.CreatedAt = time.Now()
	
	query := `
		INSERT INTO webhook_logs (
			id, webhook_id, event, payload, status_code, response, error, duration, created_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9
		)
	`
	
	_, err := r.getDB().ExecContext(ctx, query,
		log.ID, log.WebhookID, log.Event, log.Payload,
		log.StatusCode, log.Response, log.Error, log.Duration, log.CreatedAt,
	)
	
	return err
}

// GetLogs 获取Webhook日志列表
func (r *webhookRepository) GetLogs(ctx context.Context, webhookID string, filter *models.WebhookLogFilter) (*models.WebhookLogList, error) {
	query := `
		SELECT id, webhook_id, event, payload, status_code, response, error, duration, created_at
		FROM webhook_logs
		WHERE webhook_id = $1
	`
	args := []interface{}{webhookID}
	argIndex := 1

	if filter.WebhookID != nil && filter.WebhookID.String() != webhookID {
		// 如果filter中的WebhookID与参数不同，使用filter中的
		args[0] = filter.WebhookID.String()
	}
	
	if filter.Event != nil {
		argIndex++
		query += fmt.Sprintf(" AND event = $%d", argIndex)
		args = append(args, *filter.Event)
	}
	
	if filter.StatusCode != nil {
		argIndex++
		query += fmt.Sprintf(" AND status_code = $%d", argIndex)
		args = append(args, *filter.StatusCode)
	}
	
	if filter.StartTime != nil {
		argIndex++
		query += fmt.Sprintf(" AND created_at >= $%d", argIndex)
		args = append(args, *filter.StartTime)
	}
	
	if filter.EndTime != nil {
		argIndex++
		query += fmt.Sprintf(" AND created_at <= $%d", argIndex)
		args = append(args, *filter.EndTime)
	}
	
	// 计算总数
	countQuery := "SELECT COUNT(*) FROM (" + query + ") as count_query"
	var total int64
	err := r.getDB().QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("获取Webhook日志总数失败: %w", err)
	}
	
	// 添加排序和分页
	query += " ORDER BY created_at DESC"
	if filter.PageSize > 0 {
		argIndex++
		query += fmt.Sprintf(" LIMIT $%d", argIndex)
		args = append(args, filter.PageSize)
		
		if filter.Page > 0 {
			argIndex++
			query += fmt.Sprintf(" OFFSET $%d", argIndex)
			args = append(args, (filter.Page-1)*filter.PageSize)
		}
	}
	
	rows, err := r.getDB().QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("查询Webhook日志失败: %w", err)
	}
	defer rows.Close()
	
	var logs []*models.WebhookLog
	for rows.Next() {
		var log models.WebhookLog
		
		err := rows.Scan(
			&log.ID, &log.WebhookID, &log.Event, &log.Payload,
			&log.StatusCode, &log.Response, &log.Error, &log.Duration, &log.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("扫描Webhook日志数据失败: %w", err)
		}
		
		logs = append(logs, &log)
	}
	
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("遍历Webhook日志数据失败: %w", err)
	}
	
	return &models.WebhookLogList{
		Logs:     logs,
		Total:    total,
		Page:     filter.Page,
		PageSize: filter.PageSize,
	}, nil
}

// GetLogByID 根据ID获取Webhook日志
func (r *webhookRepository) GetLogByID(ctx context.Context, id string) (*models.WebhookLog, error) {
	logID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("无效的日志ID: %w", err)
	}
	
	query := `
		SELECT id, webhook_id, event, payload, status_code, response, error, duration, created_at
		FROM webhook_logs
		WHERE id = $1
	`
	
	var log models.WebhookLog
	err = r.getDB().QueryRowContext(ctx, query, logID).Scan(
		&log.ID, &log.WebhookID, &log.Event, &log.Payload,
		&log.StatusCode, &log.Response, &log.Error, &log.Duration, &log.CreatedAt,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	
	return &log, nil
}

// DeleteLogs 删除Webhook日志
func (r *webhookRepository) DeleteLogs(ctx context.Context, webhookID string, before time.Time) (int64, error) {
	webhookUUID, err := uuid.Parse(webhookID)
	if err != nil {
		return 0, fmt.Errorf("无效的Webhook ID: %w", err)
	}
	
	query := `DELETE FROM webhook_logs WHERE webhook_id = $1 AND created_at < $2`
	result, err := r.getDB().ExecContext(ctx, query, webhookUUID, before)
	if err != nil {
		return 0, err
	}
	
	rowsAffected, err := result.RowsAffected()
	return rowsAffected, err
}

// GetStats 获取Webhook统计信息
func (r *webhookRepository) GetStats(ctx context.Context, webhookID string, startTime, endTime time.Time) (*models.WebhookStats, error) {
	webhookUUID, err := uuid.Parse(webhookID)
	if err != nil {
		return nil, fmt.Errorf("无效的Webhook ID: %w", err)
	}
	
	query := `
		SELECT 
			COUNT(*) as total_requests,
			COUNT(CASE WHEN status_code >= 200 AND status_code < 300 THEN 1 END) as success_count,
			COUNT(CASE WHEN status_code < 200 OR status_code >= 300 THEN 1 END) as failure_count,
			AVG(duration) as avg_duration
		FROM webhook_logs 
		WHERE webhook_id = $1 AND created_at BETWEEN $2 AND $3
	`
	
	var stats models.WebhookStats
	err = r.getDB().QueryRowContext(ctx, query, webhookUUID, startTime, endTime).Scan(
		&stats.TotalRequests, &stats.SuccessCount, &stats.FailureCount, &stats.AvgDuration,
	)
	
	if err != nil {
		return nil, err
	}
	
	// 计算成功率
	if stats.TotalRequests > 0 {
		stats.SuccessRate = float64(stats.SuccessCount) / float64(stats.TotalRequests) * 100
	}
	
	return &stats, nil
}

// IncrementSuccessCount 增加成功计数
func (r *webhookRepository) IncrementSuccessCount(ctx context.Context, webhookID string) error {
	webhookUUID, err := uuid.Parse(webhookID)
	if err != nil {
		return fmt.Errorf("无效的Webhook ID: %w", err)
	}
	
	query := `UPDATE webhooks SET success_count = success_count + 1, updated_at = $2 WHERE id = $1`
	_, err = r.getDB().ExecContext(ctx, query, webhookUUID, time.Now())
	return err
}

// IncrementFailureCount 增加失败计数
func (r *webhookRepository) IncrementFailureCount(ctx context.Context, webhookID string) error {
	webhookUUID, err := uuid.Parse(webhookID)
	if err != nil {
		return fmt.Errorf("无效的Webhook ID: %w", err)
	}
	
	query := `UPDATE webhooks SET failure_count = failure_count + 1, updated_at = $2 WHERE id = $1`
	_, err = r.getDB().ExecContext(ctx, query, webhookUUID, time.Now())
	return err
}

// UpdateLastTriggered 更新最后触发时间
func (r *webhookRepository) UpdateLastTriggered(ctx context.Context, webhookID string) error {
	webhookUUID, err := uuid.Parse(webhookID)
	if err != nil {
		return fmt.Errorf("无效的Webhook ID: %w", err)
	}
	
	now := time.Now()
	query := `UPDATE webhooks SET last_triggered = $2, updated_at = $3 WHERE id = $1`
	_, err = r.getDB().ExecContext(ctx, query, webhookUUID, now, now)
	return err
}

// BatchCreate 批量创建Webhook
func (r *webhookRepository) BatchCreate(ctx context.Context, webhooks []*models.Webhook) error {
	if len(webhooks) == 0 {
		return nil
	}
	
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("开始事务失败: %w", err)
	}
	defer tx.Rollback()
	
	repoWithTx := NewWebhookRepositoryWithTx(tx)
	
	for _, webhook := range webhooks {
		if err := repoWithTx.Create(ctx, webhook); err != nil {
			return fmt.Errorf("批量创建Webhook失败: %w", err)
		}
	}
	
	return tx.Commit()
}

// BatchUpdate 批量更新Webhook
func (r *webhookRepository) BatchUpdate(ctx context.Context, webhooks []*models.Webhook) error {
	if len(webhooks) == 0 {
		return nil
	}
	
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("开始事务失败: %w", err)
	}
	defer tx.Rollback()
	
	repoWithTx := NewWebhookRepositoryWithTx(tx)
	
	for _, webhook := range webhooks {
		if err := repoWithTx.Update(ctx, webhook); err != nil {
			return fmt.Errorf("批量更新Webhook失败: %w", err)
		}
	}
	
	return tx.Commit()
}

// BatchEnable 批量启用Webhook
func (r *webhookRepository) BatchEnable(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("开始事务失败: %w", err)
	}
	defer tx.Rollback()
	
	repoWithTx := NewWebhookRepositoryWithTx(tx)
	
	for _, id := range ids {
		if err := repoWithTx.Enable(ctx, id); err != nil {
			return fmt.Errorf("批量启用Webhook失败: %w", err)
		}
	}
	
	return tx.Commit()
}

// BatchDisable 批量禁用Webhook
func (r *webhookRepository) BatchDisable(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("开始事务失败: %w", err)
	}
	defer tx.Rollback()
	
	repoWithTx := NewWebhookRepositoryWithTx(tx)
	
	for _, id := range ids {
		if err := repoWithTx.Disable(ctx, id); err != nil {
			return fmt.Errorf("批量禁用Webhook失败: %w", err)
		}
	}
	
	return tx.Commit()
}

// BatchDelete 批量删除Webhook
func (r *webhookRepository) BatchDelete(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("开始事务失败: %w", err)
	}
	defer tx.Rollback()
	
	repoWithTx := NewWebhookRepositoryWithTx(tx)
	
	for _, id := range ids {
		if err := repoWithTx.Delete(ctx, id); err != nil {
			return fmt.Errorf("批量删除Webhook失败: %w", err)
		}
	}
	
	return tx.Commit()
}

// CleanupLogs 清理旧的Webhook日志
func (r *webhookRepository) CleanupLogs(ctx context.Context, before time.Time) (int64, error) {
	query := `
		DELETE FROM webhook_logs
		WHERE created_at < $1
	`
	
	result, err := r.getDB().ExecContext(ctx, query, before)
	if err != nil {
		return 0, err
	}
	
	rowsAffected, err := result.RowsAffected()
	return rowsAffected, err
}

// CleanupInactive 清理非活跃的Webhook
func (r *webhookRepository) CleanupInactive(ctx context.Context, before time.Time) (int64, error) {
	query := `
		DELETE FROM webhooks
		WHERE status = $1 AND last_triggered < $2
	`
	
	result, err := r.getDB().ExecContext(ctx, query, models.WebhookStatusInactive, before)
	if err != nil {
		return 0, err
	}
	
	rowsAffected, err := result.RowsAffected()
	return rowsAffected, err
}