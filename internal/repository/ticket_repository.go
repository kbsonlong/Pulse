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
	"github.com/lib/pq"

	"Pulse/internal/models"
)

type ticketRepository struct {
	db *sqlx.DB
	tx *sqlx.Tx
}

// NewTicketRepository 创建工单仓储实例
func NewTicketRepository(db *sqlx.DB) TicketRepository {
	return &ticketRepository{
		db: db,
	}
}

// NewTicketRepositoryWithTx 创建带事务的工单仓储实例
func NewTicketRepositoryWithTx(tx *sqlx.Tx) TicketRepository {
	return &ticketRepository{
		tx: tx,
	}
}

// getExecutor 获取数据库执行器（事务或普通连接）
func (r *ticketRepository) getExecutor() sqlx.ExtContext {
	if r.tx != nil {
		return r.tx
	}
	return r.db
}

// Create 创建工单
func (r *ticketRepository) Create(ctx context.Context, ticket *models.Ticket) error {
	if ticket.ID == "" {
		ticket.ID = uuid.New().String()
	}

	now := time.Now()
	ticket.CreatedAt = now
	ticket.UpdatedAt = now

	if ticket.Status == "" {
		ticket.Status = models.TicketStatusOpen
	}

	if ticket.Priority == "" {
		ticket.Priority = models.TicketPriorityMedium
	}

	// 序列化标签和自定义字段
	tagsJSON, err := json.Marshal(ticket.Tags)
	if err != nil {
		return fmt.Errorf("序列化标签失败: %w", err)
	}

	customFieldsJSON, err := json.Marshal(ticket.CustomFields)
	if err != nil {
		return fmt.Errorf("序列化自定义字段失败: %w", err)
	}

	query := `
		INSERT INTO tickets (
			id, number, title, description, status, priority, category, type, source,
			reporter_id, assignee_id, tags, custom_fields, due_date, sla_deadline,
			created_at, updated_at
		) VALUES (
			:id, :number, :title, :description, :status, :priority, :category, :type, :source,
			:reporter_id, :assignee_id, :tags, :custom_fields, :due_date, :sla_deadline,
			:created_at, :updated_at
		)`

	_, err = sqlx.NamedExecContext(ctx, r.getExecutor(), query, map[string]interface{}{
		"id":            ticket.ID,
		"number":        ticket.Number,
		"title":         ticket.Title,
		"description":   ticket.Description,
		"status":        ticket.Status,
		"priority":      ticket.Priority,
		"category":      ticket.Category,
		"type":          ticket.Type,
		"source":        ticket.Source,
		"reporter_id":   ticket.ReporterID,
		"assignee_id":   ticket.AssigneeID,
		"tags":          string(tagsJSON),
		"custom_fields": string(customFieldsJSON),
		"due_date":      ticket.DueDate,
		"sla_deadline":  ticket.SLADeadline,
		"created_at":    ticket.CreatedAt,
		"updated_at":    ticket.UpdatedAt,
	})

	if err != nil {
		return fmt.Errorf("创建工单失败: %w", err)
	}

	return nil
}

// GetByID 根据ID获取工单
func (r *ticketRepository) GetByID(ctx context.Context, id string) (*models.Ticket, error) {
	var ticket models.Ticket
	var tagsJSON, customFieldsJSON string

	query := `
		SELECT id, number, title, description, status, priority, category, type, source,
		       reporter_id, assignee_id, tags, custom_fields, due_date, sla_deadline,
		       resolved_at, closed_at, created_at, updated_at
		FROM tickets 
		WHERE id = $1 AND deleted_at IS NULL`

	err := r.getExecutor().QueryRowxContext(ctx, query, id).Scan(
		&ticket.ID, &ticket.Number, &ticket.Title, &ticket.Description, &ticket.Status, &ticket.Priority,
		&ticket.Category, &ticket.Type, &ticket.Source, &ticket.ReporterID, &ticket.AssigneeID,
		&tagsJSON, &customFieldsJSON, &ticket.DueDate, &ticket.SLADeadline,
		&ticket.ResolvedAt, &ticket.ClosedAt, &ticket.CreatedAt, &ticket.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("工单不存在")
		}
		return nil, fmt.Errorf("获取工单失败: %w", err)
	}

	// 反序列化标签
	if tagsJSON != "" {
		err = json.Unmarshal([]byte(tagsJSON), &ticket.Tags)
		if err != nil {
			return nil, fmt.Errorf("反序列化标签失败: %w", err)
		}
	}

	// 反序列化自定义字段
	if customFieldsJSON != "" {
		err = json.Unmarshal([]byte(customFieldsJSON), &ticket.CustomFields)
		if err != nil {
			return nil, fmt.Errorf("反序列化自定义字段失败: %w", err)
		}
	}

	return &ticket, nil
}

// Update 更新工单
func (r *ticketRepository) Update(ctx context.Context, ticket *models.Ticket) error {
	ticket.UpdatedAt = time.Now()

	// 序列化标签和自定义字段
	tagsJSON, err := json.Marshal(ticket.Tags)
	if err != nil {
		return fmt.Errorf("序列化标签失败: %w", err)
	}

	customFieldsJSON, err := json.Marshal(ticket.CustomFields)
	if err != nil {
		return fmt.Errorf("序列化自定义字段失败: %w", err)
	}

	query := `
		UPDATE tickets SET 
			title = :title,
			description = :description,
			status = :status,
			priority = :priority,
			category = :category,
			type = :type,
			source = :source,
			assignee_id = :assignee_id,
			tags = :tags,
			custom_fields = :custom_fields,
			due_date = :due_date,
			sla_deadline = :sla_deadline,
			resolved_at = :resolved_at,
			closed_at = :closed_at,
			updated_at = :updated_at
		WHERE id = :id AND deleted_at IS NULL`

	_, err = r.db.NamedExecContext(ctx, query, map[string]interface{}{
		"id":            ticket.ID,
		"title":         ticket.Title,
		"description":   ticket.Description,
		"status":        ticket.Status,
		"priority":      ticket.Priority,
		"category":      ticket.Category,
		"type":          ticket.Type,
		"source":        ticket.Source,
		"assignee_id":   ticket.AssigneeID,
		"tags":          string(tagsJSON),
		"custom_fields": string(customFieldsJSON),
		"due_date":      ticket.DueDate,
		"sla_deadline":  ticket.SLADeadline,
		"resolved_at":   ticket.ResolvedAt,
		"closed_at":     ticket.ClosedAt,
		"updated_at":    ticket.UpdatedAt,
	})

	if err != nil {
		return fmt.Errorf("更新工单失败: %w", err)
	}

	return nil
}

// Delete 硬删除工单
func (r *ticketRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM tickets WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("删除工单失败: %w", err)
	}
	return nil
}

// SoftDelete 软删除工单
func (r *ticketRepository) SoftDelete(ctx context.Context, id string) error {
	now := time.Now()
	query := `
		UPDATE tickets SET 
			deleted_at = $1,
			updated_at = $1
		WHERE id = $2 AND deleted_at IS NULL`

	_, err := r.db.ExecContext(ctx, query, now, id)
	if err != nil {
		return fmt.Errorf("软删除工单失败: %w", err)
	}
	return nil
}

// List 获取工单列表
func (r *ticketRepository) List(ctx context.Context, filter *models.TicketFilter) (*models.TicketList, error) {
	var conditions []string
	var args []interface{}
	argIndex := 1

	conditions = append(conditions, "deleted_at IS NULL")

	if filter != nil {
		if filter.Status != nil {
			conditions = append(conditions, fmt.Sprintf("status = $%d", argIndex))
			args = append(args, *filter.Status)
			argIndex++
		}

		if filter.Priority != nil {
			conditions = append(conditions, fmt.Sprintf("priority = $%d", argIndex))
			args = append(args, *filter.Priority)
			argIndex++
		}

		if filter.Category != nil {
			conditions = append(conditions, fmt.Sprintf("category = $%d", argIndex))
			args = append(args, *filter.Category)
			argIndex++
		}

		if filter.Type != nil {
			conditions = append(conditions, fmt.Sprintf("type = $%d", argIndex))
			args = append(args, *filter.Type)
			argIndex++
		}

		if filter.Source != nil {
			conditions = append(conditions, fmt.Sprintf("source = $%d", argIndex))
			args = append(args, *filter.Source)
			argIndex++
		}

		if filter.ReporterID != nil {
			conditions = append(conditions, fmt.Sprintf("reporter_id = $%d", argIndex))
			args = append(args, *filter.ReporterID)
			argIndex++
		}

		if filter.AssigneeID != nil {
			conditions = append(conditions, fmt.Sprintf("assignee_id = $%d", argIndex))
			args = append(args, *filter.AssigneeID)
			argIndex++
		}

		if filter.Keyword != nil && *filter.Keyword != "" {
			conditions = append(conditions, fmt.Sprintf("(title ILIKE $%d OR description ILIKE $%d)", argIndex, argIndex))
			args = append(args, "%"+*filter.Keyword+"%")
			argIndex++
		}

		if filter.CreatedStart != nil {
			conditions = append(conditions, fmt.Sprintf("created_at >= $%d", argIndex))
			args = append(args, *filter.CreatedStart)
			argIndex++
		}

		if filter.CreatedEnd != nil {
			conditions = append(conditions, fmt.Sprintf("created_at <= $%d", argIndex))
			args = append(args, *filter.CreatedEnd)
			argIndex++
		}

		if filter.DueDateStart != nil {
			conditions = append(conditions, fmt.Sprintf("due_date >= $%d", argIndex))
			args = append(args, *filter.DueDateStart)
			argIndex++
		}

		if filter.DueDateEnd != nil {
			conditions = append(conditions, fmt.Sprintf("due_date <= $%d", argIndex))
			args = append(args, *filter.DueDateEnd)
			argIndex++
		}



		if filter.Overdue != nil && *filter.Overdue {
			conditions = append(conditions, fmt.Sprintf("due_date < $%d AND status NOT IN ('resolved', 'closed')", argIndex))
			args = append(args, time.Now())
			argIndex++
		}
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// 获取总数
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM tickets %s", whereClause)
	var total int64
	err := r.db.GetContext(ctx, &total, countQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("获取工单总数失败: %w", err)
	}

	// 构建查询
	query := fmt.Sprintf(`
		SELECT id, number, title, description, status, priority, category, type, source,
		       reporter_id, assignee_id, tags, custom_fields, due_date, sla_deadline,
		       resolved_at, closed_at, created_at, updated_at
		FROM tickets %s
		ORDER BY created_at DESC`, whereClause)

	// 添加分页
	if filter != nil && filter.Page > 0 && filter.PageSize > 0 {
		offset := (filter.Page - 1) * filter.PageSize
		query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
		args = append(args, filter.PageSize, offset)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("查询工单列表失败: %w", err)
	}
	defer rows.Close()

	var tickets []*models.Ticket
	for rows.Next() {
		var ticket models.Ticket
		var tagsJSON, customFieldsJSON string

		err := rows.Scan(
			&ticket.ID, &ticket.Number, &ticket.Title, &ticket.Description, &ticket.Status, &ticket.Priority,
			&ticket.Category, &ticket.Type, &ticket.Source, &ticket.ReporterID, &ticket.AssigneeID,
			&tagsJSON, &customFieldsJSON, &ticket.DueDate, &ticket.SLADeadline,
			&ticket.ResolvedAt, &ticket.ClosedAt, &ticket.CreatedAt, &ticket.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("扫描工单数据失败: %w", err)
		}

		// 反序列化标签
		if tagsJSON != "" {
			err = json.Unmarshal([]byte(tagsJSON), &ticket.Tags)
			if err != nil {
				return nil, fmt.Errorf("反序列化标签失败: %w", err)
			}
		}

		// 反序列化自定义字段
		if customFieldsJSON != "" {
			err = json.Unmarshal([]byte(customFieldsJSON), &ticket.CustomFields)
			if err != nil {
				return nil, fmt.Errorf("反序列化自定义字段失败: %w", err)
			}
		}

		tickets = append(tickets, &ticket)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("遍历工单数据失败: %w", err)
	}

	// 计算分页信息
	var totalPages int64 = 1
	if filter != nil && filter.PageSize > 0 {
		totalPages = (total + int64(filter.PageSize) - 1) / int64(filter.PageSize)
	}

	return &models.TicketList{
		Tickets:    tickets,
		Total:      total,
		TotalPages: int(totalPages),
	}, nil
}

// Count 获取工单总数
func (r *ticketRepository) Count(ctx context.Context, filter *models.TicketFilter) (int64, error) {
	var conditions []string
	var args []interface{}
	argIndex := 1

	conditions = append(conditions, "deleted_at IS NULL")

	if filter != nil {
		if filter.Status != nil {
			conditions = append(conditions, fmt.Sprintf("status = $%d", argIndex))
			args = append(args, *filter.Status)
			argIndex++
		}

		if filter.Priority != nil {
			conditions = append(conditions, fmt.Sprintf("priority = $%d", argIndex))
			args = append(args, *filter.Priority)
			argIndex++
		}

		if filter.AssigneeID != nil {
			conditions = append(conditions, fmt.Sprintf("assignee_id = $%d", argIndex))
			args = append(args, *filter.AssigneeID)
			argIndex++
		}
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	query := fmt.Sprintf("SELECT COUNT(*) FROM tickets %s", whereClause)
	var count int64
	err := r.db.GetContext(ctx, &count, query, args...)
	if err != nil {
		return 0, fmt.Errorf("获取工单总数失败: %w", err)
	}

	return count, nil
}

// Exists 检查工单是否存在
func (r *ticketRepository) Exists(ctx context.Context, id string) (bool, error) {
	var count int
	query := `SELECT COUNT(*) FROM tickets WHERE id = $1 AND deleted_at IS NULL`
	err := r.db.GetContext(ctx, &count, query, id)
	if err != nil {
		return false, fmt.Errorf("检查工单是否存在失败: %w", err)
	}
	return count > 0, nil
}

// UpdateStatus 更新工单状态
func (r *ticketRepository) UpdateStatus(ctx context.Context, id string, status models.TicketStatus) error {
	now := time.Now()
	query := `
		UPDATE tickets SET 
			status = $1,
			updated_at = $2
		WHERE id = $3 AND deleted_at IS NULL`

	_, err := r.db.ExecContext(ctx, query, status, now, id)
	if err != nil {
		return fmt.Errorf("更新工单状态失败: %w", err)
	}
	return nil
}

// UpdatePriority 更新工单优先级
func (r *ticketRepository) UpdatePriority(ctx context.Context, id string, priority models.TicketPriority) error {
	now := time.Now()
	query := `
		UPDATE tickets SET 
			priority = $1,
			updated_at = $2
		WHERE id = $3 AND deleted_at IS NULL`

	_, err := r.db.ExecContext(ctx, query, priority, now, id)
	if err != nil {
		return fmt.Errorf("更新工单优先级失败: %w", err)
	}
	return nil
}

// Assign 分配工单
func (r *ticketRepository) Assign(ctx context.Context, id string, assigneeID string) error {
	now := time.Now()
	query := `
		UPDATE tickets SET 
			assignee_id = $1,
			updated_at = $2
		WHERE id = $3 AND deleted_at IS NULL`

	_, err := r.getExecutor().ExecContext(ctx, query, assigneeID, now, id)
	if err != nil {
		return fmt.Errorf("分配工单失败: %w", err)
	}

	return nil
}

// Unassign 取消分配工单
func (r *ticketRepository) Unassign(ctx context.Context, id string) error {
	now := time.Now()
	query := `
		UPDATE tickets SET 
			assignee_id = NULL,
			updated_at = $1
		WHERE id = $2 AND deleted_at IS NULL`

	_, err := r.getExecutor().ExecContext(ctx, query, now, id)
	if err != nil {
		return fmt.Errorf("取消分配工单失败: %w", err)
	}

	return nil
}

// Resolve 解决工单
func (r *ticketRepository) Resolve(ctx context.Context, id, resolverID string, solution *string) error {
	now := time.Now()
	query := `
		UPDATE tickets SET 
			status = $1,
			resolved_at = $2,
			resolution = $3,
			updated_at = $2
		WHERE id = $4 AND deleted_at IS NULL`

	var resolutionText *string
	if solution != nil {
		resolutionText = solution
	}

	_, err := r.getExecutor().ExecContext(ctx, query, models.TicketStatusResolved, now, resolutionText, id)
	if err != nil {
		return fmt.Errorf("解决工单失败: %w", err)
	}
	return nil
}

// Close 关闭工单
func (r *ticketRepository) Close(ctx context.Context, id string, closerID string) error {
	now := time.Now()
	query := `
		UPDATE tickets SET 
			status = $1,
			closed_at = $2,
			closed_by = $3,
			updated_at = $2
		WHERE id = $4 AND deleted_at IS NULL`

	_, err := r.db.ExecContext(ctx, query, models.TicketStatusClosed, now, closerID, id)
	if err != nil {
		return fmt.Errorf("关闭工单失败: %w", err)
	}
	return nil
}

// Reopen 重新打开工单
func (r *ticketRepository) Reopen(ctx context.Context, id, reopenerID string) error {
	now := time.Now()
	query := `
		UPDATE tickets SET 
			status = $1,
			reopened_at = $2,
			reopen_count = reopen_count + 1,
			resolved_at = NULL,
			closed_at = NULL,
			updated_at = $2
		WHERE id = $3 AND deleted_at IS NULL`

	_, err := r.getExecutor().ExecContext(ctx, query, models.TicketStatusOpen, now, id)
	if err != nil {
		return fmt.Errorf("重新打开工单失败: %w", err)
	}
	return nil
}

// AddComment 添加评论
func (r *ticketRepository) AddComment(ctx context.Context, comment *models.TicketComment) error {
	if comment.ID == "" {
		comment.ID = uuid.New().String()
	}

	now := time.Now()
	comment.CreatedAt = now
	comment.UpdatedAt = now

	query := `
		INSERT INTO ticket_comments (
			id, ticket_id, author_id, content, is_internal, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7
		)`

	_, err := r.getExecutor().ExecContext(ctx, query,
		comment.ID, comment.TicketID, comment.AuthorID, comment.Content,
		comment.IsInternal, comment.CreatedAt, comment.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("添加评论失败: %w", err)
	}

	// 更新工单的更新时间
	updateTicketQuery := `UPDATE tickets SET updated_at = $1 WHERE id = $2`
	_, err = r.db.ExecContext(ctx, updateTicketQuery, now, comment.TicketID)
	if err != nil {
		return fmt.Errorf("更新工单时间失败: %w", err)
	}

	return nil
}

// GetComments 获取工单评论
func (r *ticketRepository) GetComments(ctx context.Context, ticketID string) ([]*models.TicketComment, error) {
	query := `
		SELECT id, ticket_id, author_id, content, is_internal, created_at, updated_at
		FROM ticket_comments 
		WHERE ticket_id = $1 AND deleted_at IS NULL
		ORDER BY created_at ASC`

	rows, err := r.db.QueryContext(ctx, query, ticketID)
	if err != nil {
		return nil, fmt.Errorf("获取工单评论失败: %w", err)
	}
	defer rows.Close()

	var comments []*models.TicketComment
	for rows.Next() {
		var comment models.TicketComment
		err := rows.Scan(
			&comment.ID, &comment.TicketID, &comment.AuthorID, &comment.Content,
			&comment.IsInternal, &comment.CreatedAt, &comment.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("扫描评论数据失败: %w", err)
		}
		comments = append(comments, &comment)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("遍历评论数据失败: %w", err)
	}

	return comments, nil
}

// UpdateComment 更新工单评论
func (r *ticketRepository) UpdateComment(ctx context.Context, comment *models.TicketComment) error {
	comment.UpdatedAt = time.Now()

	query := `
		UPDATE ticket_comments SET 
			content = $1,
			is_internal = $2,
			updated_at = $3
		WHERE id = $4 AND deleted_at IS NULL`

	result, err := r.getExecutor().ExecContext(ctx, query, 
		comment.Content, comment.IsInternal, comment.UpdatedAt, comment.ID)
	if err != nil {
		return fmt.Errorf("更新评论失败: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("获取更新结果失败: %w", err)
	}

	if rowsAffected == 0 {
		return errors.New("评论不存在或已被删除")
	}

	return nil
}

// DeleteComment 删除工单评论
func (r *ticketRepository) DeleteComment(ctx context.Context, id string) error {
	query := `
		UPDATE ticket_comments 
		SET deleted_at = NOW() 
		WHERE id = $1 AND deleted_at IS NULL`

	result, err := r.getExecutor().ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("删除评论失败: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("获取删除结果失败: %w", err)
	}

	if rowsAffected == 0 {
		return errors.New("评论不存在或已被删除")
	}

	return nil
}

// AddAttachment 添加附件
func (r *ticketRepository) AddAttachment(ctx context.Context, attachment *models.TicketAttachment) error {
	if attachment.ID == "" {
		attachment.ID = uuid.New().String()
	}

	now := time.Now()
	attachment.CreatedAt = now

	query := `
		INSERT INTO ticket_attachments (
			id, ticket_id, filename, original_filename, file_path, file_size, mime_type, upload_by, created_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9
		)`

	_, err := r.getExecutor().ExecContext(ctx, query,
		attachment.ID, attachment.TicketID, attachment.Filename, attachment.OriginalFilename,
		attachment.FilePath, attachment.FileSize, attachment.MimeType, attachment.UploadBy, attachment.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("添加附件失败: %w", err)
	}

	return nil
}

// GetAttachments 获取工单附件
func (r *ticketRepository) GetAttachments(ctx context.Context, ticketID string) ([]*models.TicketAttachment, error) {
	query := `
		SELECT id, ticket_id, filename, original_filename, file_path, file_size, mime_type, upload_by, created_at
		FROM ticket_attachments 
		WHERE ticket_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, ticketID)
	if err != nil {
		return nil, fmt.Errorf("获取工单附件失败: %w", err)
	}
	defer rows.Close()

	var attachments []*models.TicketAttachment
	for rows.Next() {
		var attachment models.TicketAttachment
		err := rows.Scan(
			&attachment.ID, &attachment.TicketID, &attachment.Filename, &attachment.OriginalFilename,
			&attachment.FilePath, &attachment.FileSize, &attachment.MimeType, &attachment.UploadBy, &attachment.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("扫描附件数据失败: %w", err)
		}
		attachments = append(attachments, &attachment)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("遍历附件数据失败: %w", err)
	}

	return attachments, nil
}

// DeleteAttachment 删除工单附件
func (r *ticketRepository) DeleteAttachment(ctx context.Context, id string) error {
	query := `
		UPDATE ticket_attachments 
		SET deleted_at = NOW() 
		WHERE id = $1 AND deleted_at IS NULL`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("删除附件失败: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("获取删除结果失败: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("附件不存在或已被删除")
	}

	return nil
}

// GetHistory 获取工单历史
func (r *ticketRepository) GetHistory(ctx context.Context, ticketID string) ([]*models.TicketHistory, error) {
	query := `
		SELECT id, ticket_id, action, field, old_value, new_value, 
		       changes, user_id, user_name, comment, created_at
		FROM ticket_history 
		WHERE ticket_id = $1
		ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, ticketID)
	if err != nil {
		return nil, fmt.Errorf("获取工单历史失败: %w", err)
	}
	defer rows.Close()

	var history []*models.TicketHistory
	for rows.Next() {
		var h models.TicketHistory
		var changesJSON string
		err := rows.Scan(
			&h.ID, &h.TicketID, &h.Action, &h.Field,
			&h.OldValue, &h.NewValue, &changesJSON,
			&h.UserID, &h.UserName, &h.Comment, &h.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("扫描历史数据失败: %w", err)
		}

		// 反序列化变更详情
		if changesJSON != "" {
			err = json.Unmarshal([]byte(changesJSON), &h.Changes)
			if err != nil {
				return nil, fmt.Errorf("反序列化变更详情失败: %w", err)
			}
		}

		history = append(history, &h)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("遍历历史数据失败: %w", err)
	}

	return history, nil
}

// BatchAssign 批量分配工单
func (r *ticketRepository) BatchAssign(ctx context.Context, ids []string, assigneeID string) error {
	if len(ids) == 0 {
		return errors.New("工单ID列表不能为空")
	}
	
	if strings.TrimSpace(assigneeID) == "" {
		return errors.New("分配人ID不能为空")
	}
	
	query := `
		UPDATE tickets 
		SET assignee_id = $1,
			status = CASE 
				WHEN status = 'open' THEN 'assigned'
				ELSE status
			END,
			updated_at = NOW()
		WHERE id = ANY($2) AND deleted_at IS NULL
	`
	
	result, err := r.db.ExecContext(ctx, query, assigneeID, pq.Array(ids))
	if err != nil {
		return fmt.Errorf("批量分配工单失败: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	
	if rowsAffected == 0 {
		return errors.New("没有工单被分配")
	}
	
	return nil
}

// BatchUpdateStatus 批量更新工单状态
func (r *ticketRepository) BatchUpdateStatus(ctx context.Context, ids []string, status models.TicketStatus) error {
	if len(ids) == 0 {
		return errors.New("工单ID列表不能为空")
	}
	
	if !status.IsValid() {
		return errors.New("无效的工单状态")
	}
	
	query := `
		UPDATE tickets 
		SET status = $1,
			updated_at = NOW()
		WHERE id = ANY($2) AND deleted_at IS NULL
	`
	
	result, err := r.db.ExecContext(ctx, query, string(status), pq.Array(ids))
	if err != nil {
		return fmt.Errorf("批量更新工单状态失败: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	
	if rowsAffected == 0 {
		return errors.New("没有工单被更新")
	}
	
	return nil
}

// CleanupClosed 清理已关闭的工单
func (r *ticketRepository) CleanupClosed(ctx context.Context, before time.Time) (int64, error) {
	query := `
		DELETE FROM tickets 
		WHERE status = 'closed' 
			AND updated_at < $1
			AND deleted_at IS NULL
	`
	
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

// GetStats 获取工单统计信息
func (r *ticketRepository) GetStats(ctx context.Context, filter *models.TicketFilter) (*models.TicketStats, error) {
	stats := &models.TicketStats{
		ByStatus:   make(map[string]int64),
		ByPriority: make(map[string]int64),
		ByCategory: make(map[string]int64),
		ByType:     make(map[string]int64),
	}

	// 按状态统计
	statusQuery := `
		SELECT status, COUNT(*) 
		FROM tickets 
		WHERE deleted_at IS NULL 
		GROUP BY status`

	rows, err := r.db.QueryContext(ctx, statusQuery)
	if err != nil {
		return nil, fmt.Errorf("按状态统计失败: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var status string
		var count int64
		err := rows.Scan(&status, &count)
		if err != nil {
			return nil, fmt.Errorf("扫描状态统计失败: %w", err)
		}
		stats.ByStatus[status] = count
		stats.Total += count
	}

	// 按优先级统计
	priorityQuery := `
		SELECT priority, COUNT(*) 
		FROM tickets 
		WHERE deleted_at IS NULL 
		GROUP BY priority`

	rows, err = r.db.QueryContext(ctx, priorityQuery)
	if err != nil {
		return nil, fmt.Errorf("按优先级统计失败: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var priority string
		var count int64
		err := rows.Scan(&priority, &count)
		if err != nil {
			return nil, fmt.Errorf("扫描优先级统计失败: %w", err)
		}
		stats.ByPriority[priority] = count
	}

	// 计算其他统计指标
	// 获取未分配工单数
	unassignedQuery := `SELECT COUNT(*) FROM tickets WHERE assignee_id IS NULL AND deleted_at IS NULL`
	err = r.db.GetContext(ctx, &stats.Unassigned, unassignedQuery)
	if err != nil {
		return nil, fmt.Errorf("获取未分配工单数失败: %w", err)
	}

	// 获取逾期工单数
	overdueQuery := `SELECT COUNT(*) FROM tickets WHERE due_date < $1 AND status NOT IN ('resolved', 'closed') AND deleted_at IS NULL`
	err = r.db.GetContext(ctx, &stats.Overdue, overdueQuery, time.Now())
	if err != nil {
		return nil, fmt.Errorf("获取逾期工单数失败: %w", err)
	}

	// 获取即将到期工单数
	dueSoonQuery := `SELECT COUNT(*) FROM tickets WHERE due_date BETWEEN $1 AND $2 AND status NOT IN ('resolved', 'closed') AND deleted_at IS NULL`
	err = r.db.GetContext(ctx, &stats.DueSoon, dueSoonQuery, time.Now(), time.Now().Add(24*time.Hour))
	if err != nil {
		return nil, fmt.Errorf("获取即将到期工单数失败: %w", err)
	}

	return stats, nil
}

// GetSLAStatus 获取SLA状态
func (r *ticketRepository) GetSLAStatus(ctx context.Context, id string) (*models.TicketSLAStatusInfo, error) {
	ticket, err := r.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("获取工单失败: %w", err)
	}

	status := &models.TicketSLAStatusInfo{
		TicketID: id,
	}

	if ticket.SLADeadline != nil {
		status.Deadline = *ticket.SLADeadline
		now := time.Now()

		if now.After(status.Deadline) {
			status.Status = "breached"
			status.TimeRemaining = 0
			status.IsBreached = true
		} else {
			status.TimeRemaining = status.Deadline.Sub(now)
			if status.TimeRemaining <= 2*time.Hour {
				status.Status = "at_risk"
			} else {
				status.Status = "on_track"
			}
		}
	} else {
		status.Status = "no_sla"
	}

	return status, nil
}

// UpdateSLA 更新工单SLA配置
func (r *ticketRepository) UpdateSLA(ctx context.Context, id string, sla *models.TicketSLA) error {
	now := time.Now()
	
	// 序列化SLA配置
	slaJSON, err := json.Marshal(sla)
	if err != nil {
		return fmt.Errorf("序列化SLA配置失败: %w", err)
	}
	
	query := `
		UPDATE tickets SET 
			sla = $1,
			updated_at = $2
		WHERE id = $3 AND deleted_at IS NULL`

	result, err := r.getExecutor().ExecContext(ctx, query, slaJSON, now, id)
	if err != nil {
		return fmt.Errorf("更新SLA配置失败: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("获取更新结果失败: %w", err)
	}
	
	if rowsAffected == 0 {
		return errors.New("工单不存在或已被删除")
	}
	
	return nil
}

// BatchCreate 批量创建工单
func (r *ticketRepository) BatchCreate(ctx context.Context, tickets []*models.Ticket) error {
	if len(tickets) == 0 {
		return nil
	}

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("开始事务失败: %w", err)
	}
	defer tx.Rollback()

	for _, ticket := range tickets {
		if ticket.ID == "" {
			ticket.ID = uuid.New().String()
		}

		now := time.Now()
		ticket.CreatedAt = now
		ticket.UpdatedAt = now

		if ticket.Status == "" {
			ticket.Status = models.TicketStatusOpen
		}

		if ticket.Priority == "" {
			ticket.Priority = models.TicketPriorityMedium
		}

		// 序列化标签和自定义字段
		tagsJSON, err := json.Marshal(ticket.Tags)
		if err != nil {
			return fmt.Errorf("序列化标签失败: %w", err)
		}

		customFieldsJSON, err := json.Marshal(ticket.CustomFields)
		if err != nil {
			return fmt.Errorf("序列化自定义字段失败: %w", err)
		}

		query := `
			INSERT INTO tickets (
				id, title, description, status, priority, category, type, source,
				reporter_id, assignee_id, tags, custom_fields, due_date, sla_deadline,
				created_at, updated_at
			) VALUES (
				$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16
			)`

		_, err = tx.ExecContext(ctx, query,
			ticket.ID, ticket.Title, ticket.Description, ticket.Status, ticket.Priority,
			ticket.Category, ticket.Type, ticket.Source, ticket.ReporterID, ticket.AssigneeID,
			string(tagsJSON), string(customFieldsJSON), ticket.DueDate, ticket.SLADeadline,
			ticket.CreatedAt, ticket.UpdatedAt,
		)
		if err != nil {
			return fmt.Errorf("批量创建工单失败: %w", err)
		}
	}

	return tx.Commit()
}

// BatchUpdate 批量更新工单
func (r *ticketRepository) BatchUpdate(ctx context.Context, tickets []*models.Ticket) error {
	if len(tickets) == 0 {
		return nil
	}

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("开始事务失败: %w", err)
	}
	defer tx.Rollback()

	for _, ticket := range tickets {
		ticket.UpdatedAt = time.Now()

		// 序列化标签和自定义字段
		tagsJSON, err := json.Marshal(ticket.Tags)
		if err != nil {
			return fmt.Errorf("序列化标签失败: %w", err)
		}

		customFieldsJSON, err := json.Marshal(ticket.CustomFields)
		if err != nil {
			return fmt.Errorf("序列化自定义字段失败: %w", err)
		}

		query := `
			UPDATE tickets SET 
				title = $1,
				description = $2,
				status = $3,
				priority = $4,
				category = $5,
				type = $6,
				source = $7,
				assignee_id = $8,
				tags = $9,
				custom_fields = $10,
				due_date = $11,
				sla_deadline = $12,
				updated_at = $13
			WHERE id = $14 AND deleted_at IS NULL`

		_, err = tx.ExecContext(ctx, query,
			ticket.Title, ticket.Description, ticket.Status, ticket.Priority,
			ticket.Category, ticket.Type, ticket.Source, ticket.AssigneeID,
			string(tagsJSON), string(customFieldsJSON), ticket.DueDate, ticket.SLADeadline,
			ticket.UpdatedAt, ticket.ID,
		)
		if err != nil {
			return fmt.Errorf("批量更新工单失败: %w", err)
		}
	}

	return tx.Commit()
}

// BatchDelete 批量删除工单
func (r *ticketRepository) BatchDelete(ctx context.Context, ids []string) error {
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
			UPDATE tickets SET 
				deleted_at = $1,
				updated_at = $1
			WHERE id = $2 AND deleted_at IS NULL`

		_, err := tx.ExecContext(ctx, query, now, id)
		if err != nil {
			return fmt.Errorf("批量删除工单失败: %w", err)
		}
	}

	return tx.Commit()
}

// AddHistory 添加工单历史记录
func (r *ticketRepository) AddHistory(ctx context.Context, history *models.TicketHistory) error {
	if history.ID == "" {
		history.ID = uuid.New().String()
	}

	now := time.Now()
	history.CreatedAt = now

	// 序列化变更详情
	changesJSON, err := json.Marshal(history.Changes)
	if err != nil {
		return fmt.Errorf("序列化变更详情失败: %w", err)
	}

	query := `
		INSERT INTO ticket_history (
			id, ticket_id, action, field, old_value, new_value, 
			changes, user_id, user_name, comment, created_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
		)`

	_, err = r.getExecutor().ExecContext(ctx, query,
		history.ID, history.TicketID, history.Action, history.Field,
		history.OldValue, history.NewValue, string(changesJSON),
		history.UserID, history.UserName, history.Comment, history.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("添加工单历史记录失败: %w", err)
	}

	return nil
}

// GetMyTickets 获取我的工单
func (r *ticketRepository) GetMyTickets(ctx context.Context, userID string, filter *models.TicketFilter) (*models.TicketList, error) {
	if filter == nil {
		filter = &models.TicketFilter{}
	}

	// 设置默认分页
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 {
		filter.PageSize = 20
	}

	// 构建查询条件
	conditions := []string{"deleted_at IS NULL"}
	args := []interface{}{}
	argIndex := 1

	// 添加用户条件（分配给我的或我创建的）
	conditions = append(conditions, "(assignee_id = $"+fmt.Sprintf("%d", argIndex)+" OR reporter_id = $"+fmt.Sprintf("%d", argIndex)+")")
	args = append(args, userID)
	argIndex++

	// 添加其他过滤条件
	if filter.Status != nil {
		conditions = append(conditions, "status = $"+fmt.Sprintf("%d", argIndex))
		args = append(args, *filter.Status)
		argIndex++
	}

	if filter.Priority != nil {
		conditions = append(conditions, "priority = $"+fmt.Sprintf("%d", argIndex))
		args = append(args, *filter.Priority)
		argIndex++
	}

	if filter.Type != nil {
		conditions = append(conditions, "type = $"+fmt.Sprintf("%d", argIndex))
		args = append(args, *filter.Type)
		argIndex++
	}

	// 构建查询语句
	whereClause := "WHERE " + strings.Join(conditions, " AND ")

	// 计算总数
	countQuery := "SELECT COUNT(*) FROM tickets " + whereClause
	var total int64
	err := r.db.GetContext(ctx, &total, countQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("获取工单总数失败: %w", err)
	}

	// 构建排序
	orderBy := "ORDER BY created_at DESC"
	if filter.SortBy != nil {
		sortOrder := "DESC"
		if filter.SortOrder != nil && strings.ToUpper(*filter.SortOrder) == "ASC" {
			sortOrder = "ASC"
		}
		orderBy = fmt.Sprintf("ORDER BY %s %s", *filter.SortBy, sortOrder)
	}

	// 分页
	offset := (filter.Page - 1) * filter.PageSize
	limit := filter.PageSize

	// 查询数据
	query := `
		SELECT id, number, title, description, type, status, priority, severity, source,
		       category, subcategory, tags, labels, alert_id, rule_id, data_source_id,
		       reporter_id, reporter_name, assignee_id, assignee_name, team_id, team_name,
		       sla, sla_deadline, due_date, response_time, resolution_time,
		       first_response_at, resolved_at, closed_at, reopened_at, reopen_count,
		       comment_count, attachment_count, work_time, estimated_time, actual_time,
		       resolution, root_cause, workaround, impact, urgency, business_impact,
		       custom_fields, created_at, updated_at
		FROM tickets ` + whereClause + ` ` + orderBy + ` LIMIT $` + fmt.Sprintf("%d", argIndex) + ` OFFSET $` + fmt.Sprintf("%d", argIndex+1)

	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("查询我的工单失败: %w", err)
	}
	defer rows.Close()

	var tickets []*models.Ticket
	for rows.Next() {
		var ticket models.Ticket
		var tagsJSON, labelsJSON, customFieldsJSON string
		var slaJSON sql.NullString

		err := rows.Scan(
			&ticket.ID, &ticket.Number, &ticket.Title, &ticket.Description,
			&ticket.Type, &ticket.Status, &ticket.Priority, &ticket.Severity, &ticket.Source,
			&ticket.Category, &ticket.Subcategory, &tagsJSON, &labelsJSON,
			&ticket.AlertID, &ticket.RuleID, &ticket.DataSourceID,
			&ticket.ReporterID, &ticket.ReporterName, &ticket.AssigneeID, &ticket.AssigneeName,
			&ticket.TeamID, &ticket.TeamName, &slaJSON, &ticket.SLADeadline, &ticket.DueDate,
			&ticket.ResponseTime, &ticket.ResolutionTime, &ticket.FirstResponseAt,
			&ticket.ResolvedAt, &ticket.ClosedAt, &ticket.ReopenedAt, &ticket.ReopenCount,
			&ticket.CommentCount, &ticket.AttachmentCount, &ticket.WorkTime,
			&ticket.EstimatedTime, &ticket.ActualTime, &ticket.Resolution, &ticket.RootCause,
			&ticket.Workaround, &ticket.Impact, &ticket.Urgency, &ticket.BusinessImpact,
			&customFieldsJSON, &ticket.CreatedAt, &ticket.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("扫描工单数据失败: %w", err)
		}

		// 反序列化JSON字段
		if err := json.Unmarshal([]byte(tagsJSON), &ticket.Tags); err != nil {
			return nil, fmt.Errorf("反序列化标签失败: %w", err)
		}

		if err := json.Unmarshal([]byte(labelsJSON), &ticket.Labels); err != nil {
			return nil, fmt.Errorf("反序列化标签失败: %w", err)
		}

		if err := json.Unmarshal([]byte(customFieldsJSON), &ticket.CustomFields); err != nil {
			return nil, fmt.Errorf("反序列化自定义字段失败: %w", err)
		}

		if slaJSON.Valid {
			if err := json.Unmarshal([]byte(slaJSON.String), &ticket.SLA); err != nil {
				return nil, fmt.Errorf("反序列化SLA失败: %w", err)
			}
		}

		tickets = append(tickets, &ticket)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("遍历工单数据失败: %w", err)
	}

	// 计算分页信息
	totalPages := int((total + int64(filter.PageSize) - 1) / int64(filter.PageSize))

	return &models.TicketList{
		Tickets:     tickets,
		Total:       total,
		Page:        filter.Page,
		PageSize:    filter.PageSize,
		TotalPages:  totalPages,
		HasNext:     filter.Page < totalPages,
		HasPrevious: filter.Page > 1,
	}, nil
}

// GetByAlertID 根据告警ID获取工单
func (r *ticketRepository) GetByAlertID(ctx context.Context, alertID string) ([]*models.Ticket, error) {
	query := `
		SELECT id, number, title, description, type, status, priority, severity, source,
		       category, subcategory, tags, labels, alert_id, rule_id, data_source_id,
		       reporter_id, reporter_name, assignee_id, assignee_name, team_id, team_name,
		       sla, sla_deadline, due_date, response_time, resolution_time,
		       first_response_at, resolved_at, closed_at, reopened_at, reopen_count,
		       comment_count, attachment_count, work_time, estimated_time, actual_time,
		       resolution, root_cause, workaround, impact, urgency, business_impact,
		       custom_fields, created_at, updated_at
		FROM tickets 
		WHERE alert_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, alertID)
	if err != nil {
		return nil, fmt.Errorf("根据告警ID查询工单失败: %w", err)
	}
	defer rows.Close()

	var tickets []*models.Ticket
	for rows.Next() {
		var ticket models.Ticket
		var tagsJSON, labelsJSON, customFieldsJSON string
		var slaJSON sql.NullString

		err := rows.Scan(
			&ticket.ID, &ticket.Number, &ticket.Title, &ticket.Description,
			&ticket.Type, &ticket.Status, &ticket.Priority, &ticket.Severity, &ticket.Source,
			&ticket.Category, &ticket.Subcategory, &tagsJSON, &labelsJSON,
			&ticket.AlertID, &ticket.RuleID, &ticket.DataSourceID,
			&ticket.ReporterID, &ticket.ReporterName, &ticket.AssigneeID, &ticket.AssigneeName,
			&ticket.TeamID, &ticket.TeamName, &slaJSON, &ticket.SLADeadline, &ticket.DueDate,
			&ticket.ResponseTime, &ticket.ResolutionTime, &ticket.FirstResponseAt,
			&ticket.ResolvedAt, &ticket.ClosedAt, &ticket.ReopenedAt, &ticket.ReopenCount,
			&ticket.CommentCount, &ticket.AttachmentCount, &ticket.WorkTime,
			&ticket.EstimatedTime, &ticket.ActualTime, &ticket.Resolution, &ticket.RootCause,
			&ticket.Workaround, &ticket.Impact, &ticket.Urgency, &ticket.BusinessImpact,
			&customFieldsJSON, &ticket.CreatedAt, &ticket.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("扫描工单数据失败: %w", err)
		}

		// 反序列化JSON字段
		if err := json.Unmarshal([]byte(tagsJSON), &ticket.Tags); err != nil {
			return nil, fmt.Errorf("反序列化标签失败: %w", err)
		}

		if err := json.Unmarshal([]byte(labelsJSON), &ticket.Labels); err != nil {
			return nil, fmt.Errorf("反序列化标签失败: %w", err)
		}

		if err := json.Unmarshal([]byte(customFieldsJSON), &ticket.CustomFields); err != nil {
			return nil, fmt.Errorf("反序列化自定义字段失败: %w", err)
		}

		if slaJSON.Valid {
			if err := json.Unmarshal([]byte(slaJSON.String), &ticket.SLA); err != nil {
				return nil, fmt.Errorf("反序列化SLA失败: %w", err)
			}
		}

		tickets = append(tickets, &ticket)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("遍历工单数据失败: %w", err)
	}

	return tickets, nil
}

// GetOpenCount 获取开放状态工单数量
func (r *ticketRepository) GetOpenCount(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.GetContext(ctx, &count, `
		SELECT COUNT(*) 
		FROM tickets 
		WHERE status IN ('open', 'assigned', 'in_progress') 
		  AND deleted_at IS NULL
	`)
	return count, err
}

// GetOverdueCount 获取逾期工单数量
func (r *ticketRepository) GetOverdueCount(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.GetContext(ctx, &count, `
		SELECT COUNT(*) 
		FROM tickets 
		WHERE due_date < NOW() 
		  AND status NOT IN ('resolved', 'closed', 'cancelled') 
		  AND deleted_at IS NULL
	`)
	return count, err
}

// GetOverdueSLA 获取SLA逾期的工单
func (r *ticketRepository) GetOverdueSLA(ctx context.Context) ([]*models.Ticket, error) {
	query := `
		SELECT id, number, title, description, type, status, priority, severity, source,
		       category, subcategory, tags, labels, alert_id, rule_id, data_source_id,
		       reporter_id, reporter_name, assignee_id, assignee_name, team_id, team_name,
		       sla, sla_deadline, due_date, response_time, resolution_time,
		       first_response_at, resolved_at, closed_at, reopened_at, reopen_count,
		       comment_count, attachment_count, work_time, estimated_time, actual_time,
		       resolution, root_cause, workaround, impact, urgency, business_impact,
		       custom_fields, created_at, updated_at, deleted_at
		FROM tickets 
		WHERE deleted_at IS NULL 
		  AND sla_deadline IS NOT NULL 
		  AND sla_deadline < NOW() 
		  AND status NOT IN ('resolved', 'closed')
		ORDER BY sla_deadline ASC`

	rows, err := r.getExecutor().QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("获取SLA逾期工单失败: %w", err)
	}
	defer rows.Close()

	var tickets []*models.Ticket
	for rows.Next() {
		var ticket models.Ticket
		var tagsJSON, labelsJSON, customFieldsJSON string
		err := rows.Scan(
			&ticket.ID, &ticket.Number, &ticket.Title, &ticket.Description,
			&ticket.Type, &ticket.Status, &ticket.Priority, &ticket.Severity, &ticket.Source,
			&ticket.Category, &ticket.Subcategory, &tagsJSON, &labelsJSON,
			&ticket.AlertID, &ticket.RuleID, &ticket.DataSourceID,
			&ticket.ReporterID, &ticket.ReporterName, &ticket.AssigneeID, &ticket.AssigneeName,
			&ticket.TeamID, &ticket.TeamName, &ticket.SLA, &ticket.SLADeadline,
			&ticket.DueDate, &ticket.ResponseTime, &ticket.ResolutionTime,
			&ticket.FirstResponseAt, &ticket.ResolvedAt, &ticket.ClosedAt, &ticket.ReopenedAt,
			&ticket.ReopenCount, &ticket.CommentCount, &ticket.AttachmentCount,
			&ticket.WorkTime, &ticket.EstimatedTime, &ticket.ActualTime,
			&ticket.Resolution, &ticket.RootCause, &ticket.Workaround,
			&ticket.Impact, &ticket.Urgency, &ticket.BusinessImpact,
			&customFieldsJSON, &ticket.CreatedAt, &ticket.UpdatedAt, &ticket.DeletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("扫描工单数据失败: %w", err)
		}

		// 反序列化JSON字段
		if err := json.Unmarshal([]byte(tagsJSON), &ticket.Tags); err != nil {
			return nil, fmt.Errorf("反序列化tags失败: %w", err)
		}
		if err := json.Unmarshal([]byte(labelsJSON), &ticket.Labels); err != nil {
			return nil, fmt.Errorf("反序列化labels失败: %w", err)
		}
		if err := json.Unmarshal([]byte(customFieldsJSON), &ticket.CustomFields); err != nil {
			return nil, fmt.Errorf("反序列化custom_fields失败: %w", err)
		}

		tickets = append(tickets, &ticket)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("遍历工单数据失败: %w", err)
	}

	return tickets, nil
}

// GetSLA 根据工单ID获取SLA配置
func (r *ticketRepository) GetSLA(ctx context.Context, id string) (*models.TicketSLA, error) {
	query := `
		SELECT id, name, description, type, priority, severity, response_time,
		       resolution_time, escalation_rules, business_hours, holidays,
		       enabled, created_by, updated_by, created_at, updated_at
		FROM ticket_slas 
		WHERE id = (
			SELECT sla_id FROM tickets WHERE id = $1 AND deleted_at IS NULL
		) AND deleted_at IS NULL`

	var sla models.TicketSLA
	var escalationRulesJSON, businessHoursJSON, holidaysJSON string

	err := r.getExecutor().QueryRowxContext(ctx, query, id).Scan(
		&sla.ID, &sla.Name, &sla.Description, &sla.Type, &sla.Priority,
		&sla.Severity, &sla.ResponseTime, &sla.ResolutionTime,
		&escalationRulesJSON, &businessHoursJSON, &holidaysJSON,
		&sla.Enabled, &sla.CreatedBy, &sla.UpdatedBy,
		&sla.CreatedAt, &sla.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("工单SLA不存在")
		}
		return nil, fmt.Errorf("查询工单SLA失败: %w", err)
	}

	// 反序列化JSON字段
	if escalationRulesJSON != "" {
		err = json.Unmarshal([]byte(escalationRulesJSON), &sla.EscalationRules)
		if err != nil {
			return nil, fmt.Errorf("反序列化升级规则失败: %w", err)
		}
	}

	if businessHoursJSON != "" {
		err = json.Unmarshal([]byte(businessHoursJSON), &sla.BusinessHours)
		if err != nil {
			return nil, fmt.Errorf("反序列化工作时间失败: %w", err)
		}
	}

	if holidaysJSON != "" {
		err = json.Unmarshal([]byte(holidaysJSON), &sla.Holidays)
		if err != nil {
			return nil, fmt.Errorf("反序列化节假日失败: %w", err)
		}
	}

	return &sla, nil
}

// GetTrend 获取工单趋势数据
func (r *ticketRepository) GetTrend(ctx context.Context, start, end time.Time, interval string) ([]*models.TicketTrendPoint, error) {
	conditions := []string{"deleted_at IS NULL", "created_at >= $1", "created_at <= $2"}
	args := []interface{}{start, end}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// 根据间隔类型构建时间分组
	var timeGroup string
	switch interval {
	case "hour":
		timeGroup = "date_trunc('hour', created_at)"
	case "day":
		timeGroup = "date_trunc('day', created_at)"
	case "week":
		timeGroup = "date_trunc('week', created_at)"
	case "month":
		timeGroup = "date_trunc('month', created_at)"
	default:
		timeGroup = "date_trunc('day', created_at)"
	}

	query := fmt.Sprintf(`
		SELECT 
			%s as time_bucket,
			COUNT(CASE WHEN status = 'open' OR status = 'in_progress' THEN 1 END) as created,
			COUNT(CASE WHEN status = 'resolved' THEN 1 END) as resolved,
			COUNT(CASE WHEN status = 'closed' THEN 1 END) as closed
		FROM tickets 
		%s
		GROUP BY time_bucket
		ORDER BY time_bucket ASC
	`, timeGroup, whereClause)

	rows, err := r.getExecutor().QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("查询工单趋势失败: %w", err)
	}
	defer rows.Close()

	var points []*models.TicketTrendPoint
	for rows.Next() {
		var point models.TicketTrendPoint
		err := rows.Scan(&point.Time, &point.Created, &point.Resolved, &point.Closed)
		if err != nil {
			return nil, fmt.Errorf("扫描趋势数据失败: %w", err)
		}
		points = append(points, &point)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("遍历趋势数据失败: %w", err)
	}

	return points, nil
}