package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"pulse/internal/models"
)

// 错误定义
var (
	ErrNotificationTemplateNotFound = errors.New("notification template not found")
)

// notificationRepository 通知仓储实现
type notificationRepository struct {
	db *sqlx.DB
	tx *sqlx.Tx
}

// NewNotificationRepository 创建新的通知仓储
func NewNotificationRepository(db *sqlx.DB) NotificationRepository {
	return &notificationRepository{db: db}
}

// NewNotificationRepositoryWithTx 创建带事务的通知仓储
func NewNotificationRepositoryWithTx(tx *sqlx.Tx) NotificationRepository {
	return &notificationRepository{tx: tx}
}

// getDB 获取数据库连接或事务
func (r *notificationRepository) getDB() sqlx.ExtContext {
	if r.tx != nil {
		return r.tx
	}
	return r.db
}

// Create 创建通知
func (r *notificationRepository) Create(ctx context.Context, notification *models.Notification) error {
	query := `
		INSERT INTO notifications (
			id, alert_id, type, recipient, subject, content, 
			status, retry_count, max_retries, last_error, 
			sent_at, created_at, updated_at
		) VALUES (
			:id, :alert_id, :type, :recipient, :subject, :content,
			:status, :retry_count, :max_retries, :last_error,
			:sent_at, :created_at, :updated_at
		)`

	_, err := sqlx.NamedExecContext(ctx, r.db, query, notification)
	return err
}

// GetByID 根据ID获取通知
func (r *notificationRepository) GetByID(ctx context.Context, id string) (*models.Notification, error) {
	uuid, err := uuid.Parse(id)
	if err != nil {
		return nil, err
	}
	var notification models.Notification
	query := `SELECT * FROM notifications WHERE id = $1`

	err = sqlx.GetContext(ctx, r.db, &notification, query, uuid)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &notification, nil
}

// Update 更新通知
func (r *notificationRepository) Update(ctx context.Context, notification *models.Notification) error {
	notification.UpdatedAt = time.Now()
	query := `
		UPDATE notifications SET
			status = :status,
			retry_count = :retry_count,
			last_error = :last_error,
			sent_at = :sent_at,
			updated_at = :updated_at
		WHERE id = :id`

	_, err := sqlx.NamedExecContext(ctx, r.db, query, notification)
	return err
}

// Delete 删除通知
func (r *notificationRepository) Delete(ctx context.Context, id string) error {
	uuid, err := uuid.Parse(id)
	if err != nil {
		return err
	}
	query := `DELETE FROM notifications WHERE id = $1`
	_, err = r.db.ExecContext(ctx, query, uuid)
	return err
}

// SoftDelete 软删除通知
func (r *notificationRepository) SoftDelete(ctx context.Context, id string) error {
	uuid, err := uuid.Parse(id)
	if err != nil {
		return err
	}
	query := `UPDATE notifications SET deleted_at = $1 WHERE id = $2`
	_, err = r.db.ExecContext(ctx, query, time.Now(), uuid)
	return err
}

// List 获取通知列表
func (r *notificationRepository) List(ctx context.Context, filter *models.NotificationFilter) (*models.NotificationList, error) {
	where, args := r.buildWhereClause(filter)
	
	// 计算偏移量
	offset := (filter.Page - 1) * filter.PageSize
	limit := filter.PageSize
	
	query := fmt.Sprintf(`
		SELECT * FROM notifications
		%s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, where, len(args)+1, len(args)+2)

	args = append(args, limit, offset)

	var notifications []*models.Notification
	err := sqlx.SelectContext(ctx, r.db, &notifications, query, args...)
	if err != nil {
		return nil, err
	}

	// 获取总数
	totalQuery := fmt.Sprintf("SELECT COUNT(*) FROM notifications %s", where)
	var total int64
	err = sqlx.GetContext(ctx, r.db, &total, totalQuery, args[:len(args)-2]...)
	if err != nil {
		return nil, err
	}

	// 计算总页数
	totalPages := int((total + int64(filter.PageSize) - 1) / int64(filter.PageSize))

	return &models.NotificationList{
		Items:      notifications,
		Total:      total,
		Page:       filter.Page,
		PageSize:   filter.PageSize,
		TotalPages: totalPages,
	}, nil
}

// CleanupFailed 清理失败的通知
func (r *notificationRepository) CleanupFailed(ctx context.Context, before time.Time) (int64, error) {
	query := `DELETE FROM notifications WHERE status = $1 AND created_at < $2`
	result, err := r.db.ExecContext(ctx, query, models.NotificationStatusFailed, before)
	if err != nil {
		return 0, err
	}
	count, err := result.RowsAffected()
	return count, err
}

// CleanupSent 清理已发送的通知
func (r *notificationRepository) CleanupSent(ctx context.Context, before time.Time) (int64, error) {
	query := `DELETE FROM notifications WHERE status = 'sent' AND created_at < $1`
	result, err := r.db.ExecContext(ctx, query, before)
	if err != nil {
		return 0, err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}
	return rowsAffected, nil
}

// buildWhereClause 构建WHERE子句
func (r *notificationRepository) buildWhereClause(filter *models.NotificationFilter) (string, []interface{}) {
	if filter == nil {
		return "", nil
	}

	var conditions []string
	var args []interface{}
	argIndex := 1

	if filter.AlertID != nil {
		conditions = append(conditions, fmt.Sprintf("alert_id = $%d", argIndex))
		args = append(args, *filter.AlertID)
		argIndex++
	}

	if filter.Type != nil {
		conditions = append(conditions, fmt.Sprintf("type = $%d", argIndex))
		args = append(args, *filter.Type)
		argIndex++
	}

	if filter.Status != nil {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argIndex))
		args = append(args, *filter.Status)
		argIndex++
	}

	if filter.Recipient != nil {
		conditions = append(conditions, fmt.Sprintf("recipient ILIKE $%d", argIndex))
		args = append(args, "%"+*filter.Recipient+"%")
		argIndex++
	}

	if filter.StartTime != nil {
		conditions = append(conditions, fmt.Sprintf("created_at >= $%d", argIndex))
		args = append(args, *filter.StartTime)
		argIndex++
	}

	if filter.EndTime != nil {
		conditions = append(conditions, fmt.Sprintf("created_at <= $%d", argIndex))
		args = append(args, *filter.EndTime)
		argIndex++
	}

	if len(conditions) == 0 {
		return "", args
	}

	return "WHERE " + strings.Join(conditions, " AND "), args
}

// GetByAlertID 根据告警ID获取通知列表
func (r *notificationRepository) GetByAlertID(ctx context.Context, alertID string) ([]*models.Notification, error) {
	uuid, err := uuid.Parse(alertID)
	if err != nil {
		return nil, err
	}
	var notifications []*models.Notification
	query := `SELECT * FROM notifications WHERE alert_id = $1 ORDER BY created_at DESC`

	err = sqlx.SelectContext(ctx, r.db, &notifications, query, uuid)
	return notifications, err
}

// GetByRecipient 根据接收者获取通知列表
func (r *notificationRepository) GetByRecipient(ctx context.Context, recipient string) ([]*models.Notification, error) {
	var notifications []*models.Notification
	query := `SELECT * FROM notifications WHERE recipient = $1 ORDER BY created_at DESC`

	err := sqlx.SelectContext(ctx, r.db, &notifications, query, recipient)
	return notifications, err
}

// Count 统计通知数量
func (r *notificationRepository) Count(ctx context.Context, filter *models.NotificationFilter) (int64, error) {
	where, args := r.buildWhereClause(filter)
	query := fmt.Sprintf("SELECT COUNT(*) FROM notifications %s", where)
	var count int64
	err := sqlx.GetContext(ctx, r.db, &count, query, args...)
	return count, err
}

// Exists 检查通知是否存在
func (r *notificationRepository) Exists(ctx context.Context, id string) (bool, error) {
	uuid, err := uuid.Parse(id)
	if err != nil {
		return false, err
	}
	var count int
	query := `SELECT COUNT(*) FROM notifications WHERE id = $1`
	err = sqlx.GetContext(ctx, r.db, &count, query, uuid)
	return count > 0, err
}

// UpdateStatus 更新通知状态
func (r *notificationRepository) UpdateStatus(ctx context.Context, id string, status models.NotificationStatus) error {
	uuid, err := uuid.Parse(id)
	if err != nil {
		return err
	}
	now := time.Now()
	var sentAt *time.Time
	if status == models.NotificationStatusSent {
		sentAt = &now
	}

	query := `
		UPDATE notifications SET
			status = $1,
			sent_at = $2,
			updated_at = $3
		WHERE id = $4`

	_, err = r.db.ExecContext(ctx, query, status, sentAt, now, uuid)
	return err
}

// MarkAsSent 标记为已发送
func (r *notificationRepository) MarkAsSent(ctx context.Context, id string, sentAt time.Time) error {
	uuid, err := uuid.Parse(id)
	if err != nil {
		return err
	}
	query := `
		UPDATE notifications SET
			status = $1,
			sent_at = $2,
			updated_at = $3
		WHERE id = $4`

	_, err = r.db.ExecContext(ctx, query, models.NotificationStatusSent, sentAt, time.Now(), uuid)
	return err
}

// MarkAsFailed 标记为失败
func (r *notificationRepository) MarkAsFailed(ctx context.Context, id string, errorMsg string) error {
	uuid, err := uuid.Parse(id)
	if err != nil {
		return err
	}
	query := `
		UPDATE notifications SET
			status = $1,
			last_error = $2,
			updated_at = $3
		WHERE id = $4`

	_, err = r.db.ExecContext(ctx, query, models.NotificationStatusFailed, errorMsg, time.Now(), uuid)
	return err
}

// IncrementRetryCount 增加重试次数
func (r *notificationRepository) IncrementRetryCount(ctx context.Context, id string) error {
	uuid, err := uuid.Parse(id)
	if err != nil {
		return err
	}
	query := `
		UPDATE notifications SET
			retry_count = retry_count + 1,
			updated_at = $1
		WHERE id = $2`

	_, err = r.db.ExecContext(ctx, query, time.Now(), uuid)
	return err
}

// GetSentCount 获取已发送数量
func (r *notificationRepository) GetSentCount(ctx context.Context, start, end time.Time) (int64, error) {
	var count int64
	query := `SELECT COUNT(*) FROM notifications WHERE status = $1 AND created_at BETWEEN $2 AND $3`
	err := sqlx.GetContext(ctx, r.db, &count, query, models.NotificationStatusSent, start, end)
	return count, err
}

// GetFailedCount 获取失败数量
func (r *notificationRepository) GetFailedCount(ctx context.Context, start, end time.Time) (int64, error) {
	var count int64
	query := `SELECT COUNT(*) FROM notifications WHERE status = $1 AND created_at BETWEEN $2 AND $3`
	err := sqlx.GetContext(ctx, r.db, &count, query, models.NotificationStatusFailed, start, end)
	return count, err
}

// GetPendingCount 获取待处理数量
func (r *notificationRepository) GetPendingCount(ctx context.Context) (int64, error) {
	var count int64
	query := `SELECT COUNT(*) FROM notifications WHERE status = $1`
	err := sqlx.GetContext(ctx, r.db, &count, query, models.NotificationStatusPending)
	return count, err
}

// BatchUpdate 批量更新通知
func (r *notificationRepository) BatchUpdate(ctx context.Context, notifications []*models.Notification) error {
	if len(notifications) == 0 {
		return nil
	}

	// 使用数据库连接开始事务
	db := r.db
	tx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `
		UPDATE notifications SET
			status = $1,
			error_message = $2,
			sent_at = $3,
			retry_count = $4,
			updated_at = $5
		WHERE id = $6`

	for _, notification := range notifications {
		_, err = tx.ExecContext(ctx, query,
			notification.Status,
			notification.LastError, // 使用LastError字段
			notification.SentAt,
			notification.RetryCount,
			time.Now(),
			notification.ID,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// BatchUpdateStatus 批量更新通知状态
func (r *notificationRepository) BatchUpdateStatus(ctx context.Context, ids []string, status models.NotificationStatus) error {
	if len(ids) == 0 {
		return nil
	}

	// 使用数据库连接开始事务
	db := r.db
	tx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `
		UPDATE notifications SET
			status = $1,
			updated_at = $2
		WHERE id = $3`

	for _, id := range ids {
		uuid, err := uuid.Parse(id)
		if err != nil {
			return err
		}
		_, err = tx.ExecContext(ctx, query, status, time.Now(), uuid)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// GetStats 获取通知统计
func (r *notificationRepository) GetStats(ctx context.Context, filter *models.NotificationFilter) (*models.NotificationStats, error) {
	where, args := r.buildWhereClause(filter)

	query := fmt.Sprintf(`
		SELECT 
			COUNT(*) as total,
			COUNT(CASE WHEN status = 'pending' THEN 1 END) as pending,
			COUNT(CASE WHEN status = 'sent' THEN 1 END) as sent,
			COUNT(CASE WHEN status = 'failed' THEN 1 END) as failed,
			COUNT(CASE WHEN status = 'retry' THEN 1 END) as retry
		FROM notifications %s
	`, where)

	var stats models.NotificationStats
	err := sqlx.GetContext(ctx, r.db, &stats, query, args...)
	if err != nil {
		return nil, err
	}

	// 计算成功率
	if stats.Total > 0 {
		stats.SuccessRate = float64(stats.Sent) / float64(stats.Total) * 100
	}

	return &stats, nil
}

// BatchCreate 批量创建通知
func (r *notificationRepository) BatchCreate(ctx context.Context, notifications []*models.Notification) error {
	if len(notifications) == 0 {
		return nil
	}

	query := `
		INSERT INTO notifications (
			id, alert_id, type, recipient, subject, content, 
			status, retry_count, max_retries, last_error, 
			sent_at, created_at, updated_at
		) VALUES (
			:id, :alert_id, :type, :recipient, :subject, :content,
			:status, :retry_count, :max_retries, :last_error,
			:sent_at, :created_at, :updated_at
		)`

	_, err := sqlx.NamedExecContext(ctx, r.db, query, notifications)
	return err
}

// DeleteOldNotifications 删除旧通知
func (r *notificationRepository) DeleteOldNotifications(ctx context.Context, before time.Time) (int64, error) {
	query := `DELETE FROM notifications WHERE created_at < $1`
	result, err := r.db.ExecContext(ctx, query, before)
	if err != nil {
		return 0, err
	}

	rowsAffected, err := result.RowsAffected()
	return rowsAffected, err
}

// CreateTemplate 创建通知模板
func (r *notificationRepository) CreateTemplate(ctx context.Context, template *models.NotificationTemplate) error {
	query := `
		INSERT INTO notification_templates (
			id, name, type, subject, content, variables, 
			is_default, created_by, created_at, updated_at
		) VALUES (
			:id, :name, :type, :subject, :content, :variables,
			:is_default, :created_by, :created_at, :updated_at
		)`

	_, err := sqlx.NamedExecContext(ctx, r.db, query, template)
	return err
}

// GetTemplate 根据ID获取模板
func (r *notificationRepository) GetTemplate(ctx context.Context, id string) (*models.NotificationTemplate, error) {
	uuid, err := uuid.Parse(id)
	if err != nil {
		return nil, err
	}

	var template models.NotificationTemplate
	query := `SELECT * FROM notification_templates WHERE id = $1`
	err = sqlx.GetContext(ctx, r.db, &template, query, uuid)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotificationTemplateNotFound
		}
		return nil, err
	}

	return &template, nil
}

// GetTemplateByID 根据ID获取通知模板
func (r *notificationRepository) GetTemplateByID(ctx context.Context, id string) (*models.NotificationTemplate, error) {
	uuid, err := uuid.Parse(id)
	if err != nil {
		return nil, err
	}
	var template models.NotificationTemplate
	query := `SELECT * FROM notification_templates WHERE id = $1`
	err = sqlx.GetContext(ctx, r.db, &template, query, uuid)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotificationTemplateNotFound
		}
		return nil, err
	}
	return &template, nil
}

// GetTemplateByName 根据名称获取模板
func (r *notificationRepository) GetTemplateByName(ctx context.Context, name string) (*models.NotificationTemplate, error) {
	var template models.NotificationTemplate
	query := `SELECT * FROM notification_templates WHERE name = $1`
	err := sqlx.GetContext(ctx, r.db, &template, query, name)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotificationTemplateNotFound
		}
		return nil, err
	}

	return &template, nil
}

// GetTemplates 获取模板列表
func (r *notificationRepository) GetTemplates(ctx context.Context) ([]*models.NotificationTemplate, error) {
	var templates []*models.NotificationTemplate
	query := `SELECT * FROM notification_templates ORDER BY created_at DESC`

	err := sqlx.SelectContext(ctx, r.db, &templates, query)
	if err != nil {
		return nil, err
	}

	return templates, nil
}

// GetTemplatesByType 根据类型获取模板列表
func (r *notificationRepository) GetTemplatesByType(ctx context.Context, notificationType models.NotificationType) ([]*models.NotificationTemplate, error) {
	var templates []*models.NotificationTemplate
	query := `SELECT * FROM notification_templates WHERE type = $1 ORDER BY created_at DESC`

	err := sqlx.SelectContext(ctx, r.db, &templates, query, notificationType)
	if err != nil {
		return nil, err
	}

	return templates, nil
}

// UpdateTemplate 更新模板
func (r *notificationRepository) UpdateTemplate(ctx context.Context, template *models.NotificationTemplate) error {
	template.UpdatedAt = time.Now()
	query := `
		UPDATE notification_templates SET
			name = :name,
			subject = :subject,
			content = :content,
			variables = :variables,
			is_default = :is_default,
			updated_at = :updated_at
		WHERE id = :id`

	_, err := sqlx.NamedExecContext(ctx, r.db, query, template)
	return err
}

// DeleteTemplate 删除模板
func (r *notificationRepository) DeleteTemplate(ctx context.Context, id string) error {
	uuid, err := uuid.Parse(id)
	if err != nil {
		return err
	}
	query := `DELETE FROM notification_templates WHERE id = $1`
	_, err = r.db.ExecContext(ctx, query, uuid)
	return err
}