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

type knowledgeRepository struct {
	db *sqlx.DB
	tx *sqlx.Tx
}

// NewKnowledgeRepository 创建知识库仓储实例
func NewKnowledgeRepository(db *sqlx.DB) KnowledgeRepository {
	return &knowledgeRepository{
		db: db,
	}
}

// NewKnowledgeRepositoryWithTx 创建带事务的知识库仓储实例
func NewKnowledgeRepositoryWithTx(tx *sqlx.Tx) KnowledgeRepository {
	return &knowledgeRepository{
		tx: tx,
	}
}

// getExecutor 获取数据库执行器（事务或普通连接）
func (r *knowledgeRepository) getExecutor() sqlx.ExtContext {
	if r.tx != nil {
		return r.tx
	}
	return r.db
}

// Create 创建知识库文章
func (r *knowledgeRepository) Create(ctx context.Context, article *models.KnowledgeArticle) error {
	if article.ID == "" {
		article.ID = uuid.New().String()
	}

	now := time.Now()
	article.CreatedAt = now
	article.UpdatedAt = now
	article.Version = "1"

	if article.Status == "" {
		article.Status = models.KnowledgeStatusDraft
	}

	// 序列化标签和元数据
	tagsJSON, err := json.Marshal(article.Tags)
	if err != nil {
		return fmt.Errorf("序列化标签失败: %w", err)
	}

	metadataJSON, err := json.Marshal(article.Metadata)
	if err != nil {
		return fmt.Errorf("序列化元数据失败: %w", err)
	}

	query := `
		INSERT INTO knowledge_articles (
			id, title, content, summary, category_id, status, type, language,
			author_id, reviewer_id, tags, metadata, version, view_count, like_count,
			is_featured, visibility, created_at, updated_at
		) VALUES (
			:id, :title, :content, :summary, :category_id, :status, :type, :language,
			:author_id, :reviewer_id, :tags, :metadata, :version, :view_count, :like_count,
			:is_featured, :visibility, :created_at, :updated_at
		)`

	_, err = sqlx.NamedExecContext(ctx, r.db, query, map[string]interface{}{
		"id":          article.ID,
		"title":       article.Title,
		"content":     article.Content,
		"summary":     article.Summary,
		"category_id": article.CategoryID,
		"status":      article.Status,
		"type":        article.Type,
		"language":    article.Language,
		"author_id":   article.AuthorID,
		"reviewer_id": article.ReviewerID,
		"tags":        string(tagsJSON),
		"metadata":    string(metadataJSON),
		"version":     article.Version,
		"view_count":  article.ViewCount,
		"like_count":  article.LikeCount,
		"is_featured": article.IsFeatured,
		"visibility":  article.Visibility,
		"created_at":  article.CreatedAt,
		"updated_at":  article.UpdatedAt,
	})

	if err != nil {
		return fmt.Errorf("创建知识库文章失败: %w", err)
	}

	return nil
}

// Unarchive 取消归档知识库
func (r *knowledgeRepository) Unarchive(ctx context.Context, id string) error {
	// 检查知识库是否存在且为归档状态
	var currentStatus models.KnowledgeStatus
	query := `SELECT status FROM knowledge WHERE id = $1 AND deleted_at IS NULL`
	err := r.getExecutor().QueryRowxContext(ctx, query, id).Scan(&currentStatus)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("知识库不存在")
		}
		return fmt.Errorf("获取知识库状态失败: %w", err)
	}
	
	// 只有归档状态的知识库才能取消归档
	if currentStatus != models.KnowledgeStatusArchived {
		return fmt.Errorf("只有归档状态的知识库才能取消归档")
	}
	
	// 更新状态为已发布
	updateQuery := `
		UPDATE knowledge 
		SET status = $1, archived_at = NULL, updated_at = CURRENT_TIMESTAMP
		WHERE id = $2 AND deleted_at IS NULL`
	
	_, err = r.db.ExecContext(ctx, updateQuery, models.KnowledgeStatusPublished, id)
	if err != nil {
		return fmt.Errorf("取消归档失败: %w", err)
	}
	
	return nil
}

// GetByID 根据ID获取知识库文章
func (r *knowledgeRepository) GetByID(ctx context.Context, id string) (*models.KnowledgeArticle, error) {
	var article models.KnowledgeArticle
	var tagsJSON, metadataJSON string

	query := `
		SELECT id, title, content, summary, category_id, status, type, language,
		       author_id, reviewer_id, tags, metadata, version, view_count, like_count,
		       is_featured, visibility, created_at, updated_at, published_at, reviewed_at
		FROM knowledge_articles
		WHERE slug = $1 AND deleted_at IS NULL`

	err := r.getExecutor().QueryRowxContext(ctx, query, id).Scan(
		&article.ID, &article.Title, &article.Content, &article.Summary, &article.CategoryID,
		&article.Status, &article.Type, &article.Language, &article.AuthorID, &article.ReviewerID,
		&tagsJSON, &metadataJSON, &article.Version, &article.ViewCount, &article.LikeCount,
		&article.IsFeatured, &article.Visibility, &article.PublishedAt, &article.CreatedAt, &article.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("知识库文章不存在")
		}
		return nil, fmt.Errorf("获取知识库文章失败: %w", err)
	}

	// 反序列化标签
	if tagsJSON != "" {
		err = json.Unmarshal([]byte(tagsJSON), &article.Tags)
		if err != nil {
			return nil, fmt.Errorf("反序列化标签失败: %w", err)
		}
	}

	// 反序列化元数据
	if metadataJSON != "" {
		err = json.Unmarshal([]byte(metadataJSON), &article.Metadata)
		if err != nil {
			return nil, fmt.Errorf("反序列化元数据失败: %w", err)
		}
	}

	return &article, nil
}

// GetBySlug 根据Slug获取知识库文章
func (r *knowledgeRepository) GetBySlug(ctx context.Context, slug string) (*models.KnowledgeArticle, error) {
	var article models.KnowledgeArticle
	var tagsJSON, metadataJSON string

	query := `
		SELECT id, title, content, summary, category_id, status, type, language,
		       author_id, reviewer_id, tags, metadata, version, view_count, like_count,
		       is_featured, visibility, published_at, created_at, updated_at
		FROM knowledge_articles 
		WHERE slug = $1 AND deleted_at IS NULL`

	err := r.getExecutor().QueryRowxContext(ctx, query, slug).Scan(
		&article.ID, &article.Title, &article.Content, &article.Summary, &article.CategoryID,
		&article.Status, &article.Type, &article.Language, &article.AuthorID, &article.ReviewerID,
		&tagsJSON, &metadataJSON, &article.Version, &article.ViewCount, &article.LikeCount,
		&article.IsFeatured, &article.Visibility, &article.PublishedAt, &article.CreatedAt, &article.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("知识库文章不存在")
		}
		return nil, fmt.Errorf("获取知识库文章失败: %w", err)
	}

	// 反序列化标签和元数据
	if tagsJSON != "" {
		err = json.Unmarshal([]byte(tagsJSON), &article.Tags)
		if err != nil {
			return nil, fmt.Errorf("反序列化标签失败: %w", err)
		}
	}

	if metadataJSON != "" {
		err = json.Unmarshal([]byte(metadataJSON), &article.Metadata)
		if err != nil {
			return nil, fmt.Errorf("反序列化元数据失败: %w", err)
		}
	}

	return &article, nil
}

// Update 更新知识库文章
func (r *knowledgeRepository) Update(ctx context.Context, article *models.KnowledgeArticle) error {
	article.UpdatedAt = time.Now()
	// 版本号递增（字符串类型）
	if article.Version == "" {
		article.Version = "1"
	} else {
		// 简单的版本号递增，假设版本号是数字字符串
		version := "1"
		if article.Version != "" {
			version = fmt.Sprintf("%s.1", article.Version)
		}
		article.Version = version
	}

	// 序列化标签和元数据
	tagsJSON, err := json.Marshal(article.Tags)
	if err != nil {
		return fmt.Errorf("序列化标签失败: %w", err)
	}

	metadataJSON, err := json.Marshal(article.Metadata)
	if err != nil {
		return fmt.Errorf("序列化元数据失败: %w", err)
	}

	query := `
		UPDATE knowledge_articles SET 
			title = :title,
			content = :content,
			summary = :summary,
			category_id = :category_id,
			status = :status,
			type = :type,
			language = :language,
			reviewer_id = :reviewer_id,
			tags = :tags,
			metadata = :metadata,
			version = :version,
			is_featured = :is_featured,
			visibility = :visibility,
			published_at = :published_at,
			updated_at = :updated_at
		WHERE id = :id AND deleted_at IS NULL`

	_, err = sqlx.NamedExecContext(ctx, r.getExecutor(), query, map[string]interface{}{
		"id":           article.ID,
		"title":        article.Title,
		"content":      article.Content,
		"summary":      article.Summary,
		"category_id":  article.CategoryID,
		"status":       article.Status,
		"type":         article.Type,
		"language":     article.Language,
		"reviewer_id":  article.ReviewerID,
		"tags":         string(tagsJSON),
		"metadata":     string(metadataJSON),
		"version":      article.Version,
		"is_featured":  article.IsFeatured,
		"visibility":   article.Visibility,
		"published_at": article.PublishedAt,
		"updated_at":   article.UpdatedAt,
	})

	if err != nil {
		return fmt.Errorf("更新知识库文章失败: %w", err)
	}

	return nil
}

// Delete 硬删除知识库文章
func (r *knowledgeRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM knowledge_articles WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("删除知识库文章失败: %w", err)
	}
	return nil
}

// SoftDelete 软删除知识库文章
func (r *knowledgeRepository) SoftDelete(ctx context.Context, id string) error {
	now := time.Now()
	query := `
		UPDATE knowledge_articles SET 
			deleted_at = $1,
			updated_at = $1
		WHERE id = $2 AND deleted_at IS NULL`

	_, err := r.db.ExecContext(ctx, query, now, id)
	if err != nil {
		return fmt.Errorf("软删除知识库文章失败: %w", err)
	}
	return nil
}

// List 获取知识库文章列表
func (r *knowledgeRepository) List(ctx context.Context, filter *models.KnowledgeFilter) (*models.KnowledgeList, error) {
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

		if filter.CategoryID != nil {
			conditions = append(conditions, fmt.Sprintf("category_id = $%d", argIndex))
			args = append(args, *filter.CategoryID)
			argIndex++
		}

		if filter.Type != nil {
			conditions = append(conditions, fmt.Sprintf("type = $%d", argIndex))
			args = append(args, *filter.Type)
			argIndex++
		}

		if filter.Language != nil {
			conditions = append(conditions, fmt.Sprintf("language = $%d", argIndex))
			args = append(args, *filter.Language)
			argIndex++
		}

		if filter.AuthorID != nil {
			conditions = append(conditions, fmt.Sprintf("author_id = $%d", argIndex))
			args = append(args, *filter.AuthorID)
			argIndex++
		}

		if filter.IsFeatured != nil {
			conditions = append(conditions, fmt.Sprintf("is_featured = $%d", argIndex))
			args = append(args, *filter.IsFeatured)
			argIndex++
		}

		if filter.Visibility != nil {
			conditions = append(conditions, fmt.Sprintf("visibility = $%d", argIndex))
			args = append(args, *filter.Visibility)
			argIndex++
		}

		if filter.Keyword != nil && *filter.Keyword != "" {
			conditions = append(conditions, fmt.Sprintf("(title ILIKE $%d OR content ILIKE $%d OR summary ILIKE $%d)", argIndex, argIndex, argIndex))
			args = append(args, "%"+*filter.Keyword+"%")
			argIndex++
		}

		if filter.Tags != nil && len(filter.Tags) > 0 {
			tagConditions := make([]string, len(filter.Tags))
			for i, tag := range filter.Tags {
				tagConditions[i] = fmt.Sprintf("tags::jsonb ? $%d", argIndex)
				args = append(args, tag)
				argIndex++
			}
			conditions = append(conditions, "("+strings.Join(tagConditions, " OR ")+")")
		}
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// 获取总数
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM knowledge_articles %s", whereClause)
	var total int64
	err := sqlx.GetContext(ctx, r.db, &total, countQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("获取知识库文章总数失败: %w", err)
	}

	// 构建查询
	orderBy := "ORDER BY created_at DESC"
	if filter != nil && filter.SortBy != nil {
		switch *filter.SortBy {
		case "title":
			orderBy = "ORDER BY title ASC"
		case "view_count":
			orderBy = "ORDER BY view_count DESC"
		case "like_count":
			orderBy = "ORDER BY like_count DESC"
		case "updated_at":
			orderBy = "ORDER BY updated_at DESC"
		}
	}

	query := fmt.Sprintf(`
		SELECT id, title, content, summary, category_id, status, type, language,
		       author_id, reviewer_id, tags, metadata, version, view_count, like_count,
		       is_featured, is_public, published_at, created_at, updated_at
		FROM knowledge_articles %s %s`, whereClause, orderBy)

	// 添加分页
	if filter != nil && filter.Page > 0 && filter.PageSize > 0 {
		offset := (filter.Page - 1) * filter.PageSize
		query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
		args = append(args, filter.PageSize, offset)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("查询知识库文章列表失败: %w", err)
	}
	defer rows.Close()

	var articles []*models.KnowledgeArticle
	for rows.Next() {
		var article models.KnowledgeArticle
		var tagsJSON, metadataJSON string

		err := rows.Scan(
			&article.ID, &article.Title, &article.Content, &article.Summary, &article.CategoryID,
			&article.Status, &article.Type, &article.Language, &article.AuthorID, &article.ReviewerID,
			&tagsJSON, &metadataJSON, &article.Version, &article.ViewCount, &article.LikeCount,
			&article.IsFeatured, &article.Visibility, &article.PublishedAt, &article.CreatedAt, &article.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("扫描知识库文章数据失败: %w", err)
		}

		// 反序列化标签和元数据
		if tagsJSON != "" {
			err = json.Unmarshal([]byte(tagsJSON), &article.Tags)
			if err != nil {
				return nil, fmt.Errorf("反序列化标签失败: %w", err)
			}
		}

		if metadataJSON != "" {
			err = json.Unmarshal([]byte(metadataJSON), &article.Metadata)
			if err != nil {
				return nil, fmt.Errorf("反序列化元数据失败: %w", err)
			}
		}

		articles = append(articles, &article)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("遍历知识库文章数据失败: %w", err)
	}

	// 计算分页信息
	var totalPages int64 = 1
	if filter != nil && filter.PageSize > 0 {
		totalPages = (total + int64(filter.PageSize) - 1) / int64(filter.PageSize)
	}

	return &models.KnowledgeList{
		Knowledge:  articles,
		Total:      total,
		TotalPages: int(totalPages),
	}, nil
}

// Count 获取知识库文章总数
func (r *knowledgeRepository) Count(ctx context.Context, filter *models.KnowledgeFilter) (int64, error) {
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

		if filter.CategoryID != nil {
			conditions = append(conditions, fmt.Sprintf("category_id = $%d", argIndex))
			args = append(args, *filter.CategoryID)
			argIndex++
		}

		if filter.Visibility != nil {
			conditions = append(conditions, fmt.Sprintf("visibility = $%d", argIndex))
			args = append(args, *filter.Visibility)
			argIndex++
		}
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	query := fmt.Sprintf("SELECT COUNT(*) FROM knowledge_articles %s", whereClause)
	var count int64
	err := sqlx.GetContext(ctx, r.db, &count, query, args...)
	if err != nil {
		return 0, fmt.Errorf("获取知识库文章总数失败: %w", err)
	}

	return count, nil
}

// Exists 检查知识库文章是否存在
func (r *knowledgeRepository) Exists(ctx context.Context, id string) (bool, error) {
	var count int
	query := `SELECT COUNT(*) FROM knowledge_articles WHERE id = $1 AND deleted_at IS NULL`
	err := sqlx.GetContext(ctx, r.db, &count, query, id)
	if err != nil {
		return false, fmt.Errorf("检查知识库文章是否存在失败: %w", err)
	}
	return count > 0, nil
}

// ExistsBySlug 检查Slug是否存在
func (r *knowledgeRepository) ExistsBySlug(ctx context.Context, slug string) (bool, error) {
	var count int
	query := `SELECT COUNT(*) FROM knowledge_articles WHERE slug = $1 AND deleted_at IS NULL`
	err := sqlx.GetContext(ctx, r.db, &count, query, slug)
	if err != nil {
		return false, fmt.Errorf("检查Slug是否存在失败: %w", err)
	}
	return count > 0, nil
}

// UpdateStatus 更新文章状态
func (r *knowledgeRepository) UpdateStatus(ctx context.Context, id string, status models.KnowledgeStatus) error {
	now := time.Now()
	query := `
		UPDATE knowledge_articles SET 
			status = $1,
			updated_at = $2
		WHERE id = $3 AND deleted_at IS NULL`

	if status == models.KnowledgeStatusPublished {
		query = `
			UPDATE knowledge_articles SET 
				status = $1,
				published_at = $2,
				updated_at = $2
			WHERE id = $3 AND deleted_at IS NULL`

		_, err := r.db.ExecContext(ctx, query, status, now, id)
		if err != nil {
			return fmt.Errorf("更新文章状态失败: %w", err)
		}
	} else {
		_, err := r.db.ExecContext(ctx, query, status, now, id)
		if err != nil {
			return fmt.Errorf("更新文章状态失败: %w", err)
		}
	}

	return nil
}

// Approve 审批知识库文章
func (r *knowledgeRepository) Approve(ctx context.Context, id, reviewerID string, comment *string) error {
	now := time.Now()
	
	// 更新文章状态为已审批
	query := `
		UPDATE knowledge_articles SET 
			status = $1,
			reviewer_id = $2,
			review_comment = $3,
			published_at = $4,
			updated_at = $4
		WHERE id = $5 AND deleted_at IS NULL AND status = $6`
	
	result, err := r.db.ExecContext(ctx, query, 
		models.KnowledgeStatusPublished, reviewerID, comment, now, id, models.KnowledgeStatusReview)
	if err != nil {
		return fmt.Errorf("审批知识库文章失败: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("获取影响行数失败: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("文章不存在或状态不正确")
	}
	
	return nil
}

// Reject 拒绝知识库文章
func (r *knowledgeRepository) Reject(ctx context.Context, id, reviewerID string, comment *string) error {
	now := time.Now()
	
	// 更新文章状态为已拒绝
	query := `
		UPDATE knowledge_articles SET 
			status = $1,
			reviewer_id = $2,
			review_comment = $3,
			updated_at = $4
		WHERE id = $5 AND deleted_at IS NULL AND status = $6`
	
	result, err := r.db.ExecContext(ctx, query, 
		models.KnowledgeStatusDraft, reviewerID, comment, now, id, models.KnowledgeStatusReview)
	if err != nil {
		return fmt.Errorf("拒绝知识库文章失败: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("获取影响行数失败: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("文章不存在或状态不正确")
	}
	
	return nil
}

// Publish 发布文章
func (r *knowledgeRepository) Publish(ctx context.Context, id, publisherID string) error {
	now := time.Now()
	query := `
		UPDATE knowledge_articles SET 
			status = $1,
			published_at = $2,
			updated_at = $2
		WHERE id = $3 AND deleted_at IS NULL`
	
	_, err := r.db.ExecContext(ctx, query, models.KnowledgeStatusPublished, now, id)
	if err != nil {
		return fmt.Errorf("发布文章失败: %w", err)
	}
	
	return nil
}

// Unpublish 取消发布文章
func (r *knowledgeRepository) Unpublish(ctx context.Context, id string) error {
	return r.UpdateStatus(ctx, id, models.KnowledgeStatusDraft)
}

// Archive 归档文章
func (r *knowledgeRepository) Archive(ctx context.Context, id string) error {
	return r.UpdateStatus(ctx, id, models.KnowledgeStatusArchived)
}

// IncrementViewCount 增加浏览次数
func (r *knowledgeRepository) IncrementViewCount(ctx context.Context, id string) error {
	query := `
		UPDATE knowledge_articles SET 
			view_count = view_count + 1,
			updated_at = $1
		WHERE id = $2 AND deleted_at IS NULL`

	_, err := r.db.ExecContext(ctx, query, time.Now(), id)
	if err != nil {
		return fmt.Errorf("增加浏览次数失败: %w", err)
	}
	return nil
}

// IncrementLikeCount 增加点赞次数
func (r *knowledgeRepository) IncrementLikeCount(ctx context.Context, id string) error {
	query := `
		UPDATE knowledge_articles SET 
			like_count = like_count + 1,
			updated_at = $1
		WHERE id = $2 AND deleted_at IS NULL`

	_, err := r.db.ExecContext(ctx, query, time.Now(), id)
	if err != nil {
		return fmt.Errorf("增加点赞次数失败: %w", err)
	}
	return nil
}

// DecrementLikeCount 减少点赞次数
func (r *knowledgeRepository) DecrementLikeCount(ctx context.Context, id string) error {
	query := `
		UPDATE knowledge_articles SET 
			like_count = GREATEST(like_count - 1, 0),
			updated_at = $1
		WHERE id = $2 AND deleted_at IS NULL`

	_, err := r.db.ExecContext(ctx, query, time.Now(), id)
	if err != nil {
		return fmt.Errorf("减少点赞次数失败: %w", err)
	}
	return nil
}

// IncrementDislikeCount 增加踩次数
func (r *knowledgeRepository) IncrementDislikeCount(ctx context.Context, id string) error {
	query := `
		UPDATE knowledge_articles SET 
			dislike_count = dislike_count + 1,
			updated_at = $1
		WHERE id = $2 AND deleted_at IS NULL`

	_, err := r.db.ExecContext(ctx, query, time.Now(), id)
	if err != nil {
		return fmt.Errorf("增加踩次数失败: %w", err)
	}
	return nil
}

// IncrementShareCount 增加分享次数
func (r *knowledgeRepository) IncrementShareCount(ctx context.Context, id string) error {
	query := `
		UPDATE knowledge_articles SET 
			share_count = share_count + 1,
			updated_at = $1
		WHERE id = $2 AND deleted_at IS NULL`

	_, err := r.db.ExecContext(ctx, query, time.Now(), id)
	if err != nil {
		return fmt.Errorf("增加分享次数失败: %w", err)
	}
	return nil
}

// IncrementDownloadCount 增加下载次数
func (r *knowledgeRepository) IncrementDownloadCount(ctx context.Context, id string) error {
	query := `
		UPDATE knowledge_articles SET 
			download_count = download_count + 1,
			updated_at = $1
		WHERE id = $2 AND deleted_at IS NULL`

	_, err := r.db.ExecContext(ctx, query, time.Now(), id)
	if err != nil {
		return fmt.Errorf("增加下载次数失败: %w", err)
	}
	return nil
}

// UpdateRating 更新评分
func (r *knowledgeRepository) UpdateRating(ctx context.Context, id string, rating float64) error {
	query := `
		UPDATE knowledge_articles SET 
			rating = $1,
			updated_at = $2
		WHERE id = $3 AND deleted_at IS NULL`

	_, err := r.db.ExecContext(ctx, query, rating, time.Now(), id)
	if err != nil {
		return fmt.Errorf("更新评分失败: %w", err)
	}
	return nil
}

// CreateVersion 创建文章版本
func (r *knowledgeRepository) CreateVersion(ctx context.Context, version *models.KnowledgeVersion) error {
	if version.ID == "" {
		version.ID = uuid.New().String()
	}

	version.CreatedAt = time.Now()

	query := `
		INSERT INTO knowledge_versions (
			id, knowledge_id, version, title, content, change_log, created_by, created_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8
		)`

	_, err := r.db.ExecContext(ctx, query,
			version.ID, version.KnowledgeID, version.Version, version.Title,
			version.Content, version.ChangeLog, version.CreatedBy, version.CreatedAt,
		)

	if err != nil {
		return fmt.Errorf("创建文章版本失败: %w", err)
	}

	return nil
}

// GetVersions 获取文章版本列表
func (r *knowledgeRepository) GetVersions(ctx context.Context, articleID string) ([]*models.KnowledgeVersion, error) {
	query := `
		SELECT id, knowledge_id, version, title, content, change_log, created_by, created_at
		FROM knowledge_versions 
		WHERE knowledge_id = $1
		ORDER BY version DESC`

	rows, err := r.db.QueryContext(ctx, query, articleID)
	if err != nil {
		return nil, fmt.Errorf("获取文章版本列表失败: %w", err)
	}
	defer rows.Close()

	var versions []*models.KnowledgeVersion
	for rows.Next() {
		var version models.KnowledgeVersion
		err := rows.Scan(
			&version.ID, &version.KnowledgeID, &version.Version, &version.Title,
			&version.Content, &version.ChangeLog, &version.CreatedBy, &version.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("扫描版本数据失败: %w", err)
		}
		versions = append(versions, &version)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("遍历版本数据失败: %w", err)
	}

	return versions, nil
}

// GetVersion 获取指定版本
func (r *knowledgeRepository) GetVersion(ctx context.Context, knowledgeID, versionStr string) (*models.KnowledgeVersion, error) {
	var version models.KnowledgeVersion

	query := `
		SELECT id, knowledge_id, version, title, content, change_log, created_by, created_at
		FROM knowledge_versions 
		WHERE knowledge_id = $1 AND version = $2`

	err := r.getExecutor().QueryRowxContext(ctx, query, knowledgeID, versionStr).Scan(
		&version.ID, &version.KnowledgeID, &version.Version, &version.Title,
		&version.Content, &version.ChangeLog, &version.CreatedBy, &version.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("文章版本不存在")
		}
		return nil, fmt.Errorf("获取文章版本失败: %w", err)
	}

	return &version, nil
}

// RestoreVersion 恢复到指定版本
func (r *knowledgeRepository) RestoreVersion(ctx context.Context, knowledgeID, version string) error {
	// 获取指定版本的内容
	versionData, err := r.GetVersion(ctx, knowledgeID, version)
	if err != nil {
		return fmt.Errorf("获取版本数据失败: %w", err)
	}

	// 更新主文章内容为指定版本的内容
	query := `
		UPDATE knowledge_articles SET 
			title = $1,
			content = $2,
			version = $3,
			updated_at = $4
		WHERE id = $5 AND deleted_at IS NULL`

	_, err = r.db.ExecContext(ctx, query,
		versionData.Title, versionData.Content, versionData.Version,
		time.Now(), knowledgeID,
	)
	if err != nil {
		return fmt.Errorf("恢复版本失败: %w", err)
	}

	return nil
}

// CreateCategory 创建分类
func (r *knowledgeRepository) CreateCategory(ctx context.Context, category *models.KnowledgeCategory) error {
	if category.ID == "" {
		category.ID = uuid.New().String()
	}

	now := time.Now()
	category.CreatedAt = now
	category.UpdatedAt = now

	query := `
		INSERT INTO knowledge_categories (
			id, name, description, parent_id, sort_order, is_active, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8
		)`

	_, err := r.db.ExecContext(ctx, query,
		category.ID, category.Name, category.Description,
		category.ParentID, category.SortOrder, category.IsActive, category.CreatedAt, category.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("创建分类失败: %w", err)
	}

	return nil
}

// GetCategories 获取分类列表
func (r *knowledgeRepository) GetCategories(ctx context.Context) ([]*models.KnowledgeCategory, error) {
	query := `
		SELECT id, name, description, parent_id, sort_order, is_active, created_at, updated_at
		FROM knowledge_categories 
		WHERE deleted_at IS NULL
		ORDER BY sort_order ASC, name ASC`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("获取分类列表失败: %w", err)
	}
	defer rows.Close()

	var categories []*models.KnowledgeCategory
	for rows.Next() {
		var category models.KnowledgeCategory
		err := rows.Scan(
			&category.ID, &category.Name, &category.Description,
			&category.ParentID, &category.SortOrder, &category.IsActive, &category.CreatedAt, &category.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("扫描分类数据失败: %w", err)
		}
		categories = append(categories, &category)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("遍历分类数据失败: %w", err)
	}

	return categories, nil
}

// UpdateCategory 更新分类
func (r *knowledgeRepository) UpdateCategory(ctx context.Context, category *models.KnowledgeCategory) error {
	category.UpdatedAt = time.Now()

	query := `
		UPDATE knowledge_categories SET 
			name = $1,
			description = $2,
			parent_id = $3,
			sort_order = $4,
			is_active = $5,
			updated_at = $6
		WHERE id = $7 AND deleted_at IS NULL`

	_, err := r.db.ExecContext(ctx, query,
		category.Name, category.Description, category.ParentID,
		category.SortOrder, category.IsActive, category.UpdatedAt, category.ID,
	)

	if err != nil {
		return fmt.Errorf("更新分类失败: %w", err)
	}

	return nil
}

// DeleteCategory 删除分类
func (r *knowledgeRepository) DeleteCategory(ctx context.Context, id string) error {
	now := time.Now()
	query := `
		UPDATE knowledge_categories SET 
			deleted_at = $1,
			updated_at = $1
		WHERE id = $2 AND deleted_at IS NULL`

	_, err := r.getExecutor().ExecContext(ctx, query, now, id)
	if err != nil {
		return fmt.Errorf("删除分类失败: %w", err)
	}
	return nil
}



// GetTagStats 获取标签统计
func (r *knowledgeRepository) GetTagStats(ctx context.Context) (map[string]int64, error) {
	query := `
		SELECT jsonb_array_elements_text(tags) as tag, COUNT(*) as count
		FROM knowledge_articles 
		WHERE deleted_at IS NULL AND tags IS NOT NULL
		GROUP BY tag
		ORDER BY count DESC, tag`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("获取标签统计失败: %w", err)
	}
	defer rows.Close()

	stats := make(map[string]int64)
	for rows.Next() {
		var tag string
		var count int64
		err := rows.Scan(&tag, &count)
		if err != nil {
			return nil, fmt.Errorf("扫描标签统计数据失败: %w", err)
		}
		stats[tag] = count
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("遍历标签统计数据失败: %w", err)
	}

	return stats, nil
}

// GetPopular 获取热门知识
func (r *knowledgeRepository) GetPopular(ctx context.Context, limit int) ([]*models.Knowledge, error) {
	query := `
		SELECT id, title, content, summary, type, status, visibility, format, 
		       category_id, author_id, team_id, language, tags, keywords, 
		       view_count, like_count, dislike_count, share_count, comment_count, 
		       download_count, rating, rating_count, is_featured, is_template, 
		       template_data, metadata, related_ids, expires_at, 
		       created_at, updated_at, published_at, last_viewed_at
		FROM knowledge 
		WHERE deleted_at IS NULL AND status = $1
		ORDER BY view_count DESC, like_count DESC, created_at DESC
		LIMIT $2`

	rows, err := r.db.QueryContext(ctx, query, models.KnowledgeStatusPublished, limit)
	if err != nil {
		return nil, fmt.Errorf("获取热门知识失败: %w", err)
	}
	defer rows.Close()

	var knowledge []*models.Knowledge
	for rows.Next() {
		var k models.Knowledge
		var tagsJSON, keywordsJSON, templateDataJSON, metadataJSON, relatedIDsJSON sql.NullString

		var dislikeCount, shareCount, commentCount, downloadCount, ratingCount int64
		var rating sql.NullFloat64
		var lastViewedAt sql.NullTime

		err := rows.Scan(
			&k.ID, &k.Title, &k.Content, &k.Summary, &k.Type, &k.Status, &k.Visibility, &k.Format,
			&k.CategoryID, &k.AuthorID, &k.TeamID, &k.Language, &tagsJSON, &keywordsJSON,
			&k.ViewCount, &k.LikeCount, &dislikeCount, &shareCount, &commentCount,
			&downloadCount, &rating, &ratingCount, &k.IsFeatured, &k.IsTemplate,
			&templateDataJSON, &metadataJSON, &relatedIDsJSON, &k.ExpiresAt,
			&k.CreatedAt, &k.UpdatedAt, &k.PublishedAt, &lastViewedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("扫描知识数据失败: %w", err)
		}

		// 设置指标数据
		k.Metrics = &models.KnowledgeMetrics{
			ViewCount:     k.ViewCount,
			LikeCount:     k.LikeCount,
			DislikeCount:  dislikeCount,
			ShareCount:    shareCount,
			CommentCount:  commentCount,
			DownloadCount: downloadCount,
			RatingCount:   ratingCount,
		}
		if rating.Valid {
			k.Metrics.Rating = &rating.Float64
		}
		if lastViewedAt.Valid {
			k.Metrics.LastViewedAt = &lastViewedAt.Time
		}
		if err != nil {
			return nil, fmt.Errorf("扫描知识数据失败: %w", err)
		}

		// 反序列化JSON字段
		if tagsJSON.Valid {
			json.Unmarshal([]byte(tagsJSON.String), &k.Tags)
		}
		if keywordsJSON.Valid {
			json.Unmarshal([]byte(keywordsJSON.String), &k.Keywords)
		}
		if templateDataJSON.Valid {
			json.Unmarshal([]byte(templateDataJSON.String), &k.TemplateData)
		}
		if metadataJSON.Valid {
			json.Unmarshal([]byte(metadataJSON.String), &k.Metadata)
		}
		if relatedIDsJSON.Valid {
			json.Unmarshal([]byte(relatedIDsJSON.String), &k.RelatedIDs)
		}

		knowledge = append(knowledge, &k)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("遍历知识数据失败: %w", err)
	}

	return knowledge, nil
}

// GetRecent 获取最近知识
func (r *knowledgeRepository) GetRecent(ctx context.Context, limit int) ([]*models.Knowledge, error) {
	query := `
		SELECT id, title, content, summary, type, status, visibility, format, 
		       category_id, author_id, team_id, language, tags, keywords, 
		       view_count, like_count, dislike_count, share_count, comment_count, 
		       download_count, rating, rating_count, is_featured, is_template, 
		       template_data, metadata, related_ids, expires_at, 
		       created_at, updated_at, published_at, last_viewed_at
		FROM knowledge 
		WHERE deleted_at IS NULL AND status = $1
		ORDER BY created_at DESC, updated_at DESC
		LIMIT $2`

	rows, err := r.db.QueryContext(ctx, query, models.KnowledgeStatusPublished, limit)
	if err != nil {
		return nil, fmt.Errorf("获取最近知识失败: %w", err)
	}
	defer rows.Close()

	var knowledge []*models.Knowledge
	for rows.Next() {
		var k models.Knowledge
		var tagsJSON, keywordsJSON, templateDataJSON, metadataJSON, relatedIDsJSON sql.NullString
		var dislikeCount, shareCount, commentCount, downloadCount, ratingCount int64
		var rating sql.NullFloat64
		var lastViewedAt sql.NullTime

		err := rows.Scan(
			&k.ID, &k.Title, &k.Content, &k.Summary, &k.Type, &k.Status, &k.Visibility, &k.Format,
			&k.CategoryID, &k.AuthorID, &k.TeamID, &k.Language, &tagsJSON, &keywordsJSON,
			&k.ViewCount, &k.LikeCount, &dislikeCount, &shareCount, &commentCount,
			&downloadCount, &rating, &ratingCount, &k.IsFeatured, &k.IsTemplate,
			&templateDataJSON, &metadataJSON, &relatedIDsJSON, &k.ExpiresAt,
			&k.CreatedAt, &k.UpdatedAt, &k.PublishedAt, &lastViewedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("扫描知识数据失败: %w", err)
		}

		// 设置指标数据
		k.Metrics = &models.KnowledgeMetrics{
			ViewCount:     k.ViewCount,
			LikeCount:     k.LikeCount,
			DislikeCount:  dislikeCount,
			ShareCount:    shareCount,
			CommentCount:  commentCount,
			DownloadCount: downloadCount,
			RatingCount:   ratingCount,
		}
		if rating.Valid {
			k.Metrics.Rating = &rating.Float64
		}
		if lastViewedAt.Valid {
			k.Metrics.LastViewedAt = &lastViewedAt.Time
		}

		// 反序列化JSON字段
		if tagsJSON.Valid {
			json.Unmarshal([]byte(tagsJSON.String), &k.Tags)
		}
		if keywordsJSON.Valid {
			json.Unmarshal([]byte(keywordsJSON.String), &k.Keywords)
		}
		if templateDataJSON.Valid {
			json.Unmarshal([]byte(templateDataJSON.String), &k.TemplateData)
		}
		if metadataJSON.Valid {
			json.Unmarshal([]byte(metadataJSON.String), &k.Metadata)
		}
		if relatedIDsJSON.Valid {
			json.Unmarshal([]byte(relatedIDsJSON.String), &k.RelatedIDs)
		}

		knowledge = append(knowledge, &k)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("遍历知识数据失败: %w", err)
	}

	return knowledge, nil
}

// GetRelated 获取相关知识
func (r *knowledgeRepository) GetRelated(ctx context.Context, knowledgeID string, limit int) ([]*models.Knowledge, error) {
	query := `
		SELECT DISTINCT k.id, k.title, k.content, k.summary, k.type, k.status, k.visibility, k.format, 
		       k.category_id, k.author_id, k.team_id, k.language, k.tags, k.keywords, 
		       k.view_count, k.like_count, 0 as dislike_count, 0 as share_count, 0 as comment_count, 
		       0 as download_count, NULL as rating, 0 as rating_count, k.is_featured, k.is_template, 
		       k.template_data, k.metadata, k.related_ids, k.expires_at, 
		       k.created_at, k.updated_at, k.published_at, NULL as last_viewed_at
		FROM knowledge k
		WHERE k.deleted_at IS NULL 
		  AND k.status = $1 
		  AND k.id != $2
		  AND (
			  k.category_id = (SELECT category_id FROM knowledge WHERE id = $2)
			  OR k.tags && (SELECT tags FROM knowledge WHERE id = $2)
			  OR k.keywords && (SELECT keywords FROM knowledge WHERE id = $2)
		  )
		ORDER BY k.view_count DESC, k.like_count DESC, k.created_at DESC
		LIMIT $3`

	rows, err := r.db.QueryContext(ctx, query, models.KnowledgeStatusPublished, knowledgeID, limit)
	if err != nil {
		return nil, fmt.Errorf("获取相关知识失败: %w", err)
	}
	defer rows.Close()

	var knowledge []*models.Knowledge
	for rows.Next() {
		var k models.Knowledge
		var tagsJSON, keywordsJSON, templateDataJSON, metadataJSON, relatedIDsJSON sql.NullString
		var dislikeCount, shareCount, commentCount, downloadCount, ratingCount int64
		var rating sql.NullFloat64
		var lastViewedAt sql.NullTime

		err := rows.Scan(
			&k.ID, &k.Title, &k.Content, &k.Summary, &k.Type, &k.Status, &k.Visibility, &k.Format,
			&k.CategoryID, &k.AuthorID, &k.TeamID, &k.Language, &tagsJSON, &keywordsJSON,
			&k.ViewCount, &k.LikeCount, &dislikeCount, &shareCount, &commentCount,
			&downloadCount, &rating, &ratingCount, &k.IsFeatured, &k.IsTemplate,
			&templateDataJSON, &metadataJSON, &relatedIDsJSON, &k.ExpiresAt,
			&k.CreatedAt, &k.UpdatedAt, &k.PublishedAt, &lastViewedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("扫描知识数据失败: %w", err)
		}

		// 设置指标数据
		k.Metrics = &models.KnowledgeMetrics{
			ViewCount:     k.ViewCount,
			LikeCount:     k.LikeCount,
			DislikeCount:  dislikeCount,
			ShareCount:    shareCount,
			CommentCount:  commentCount,
			DownloadCount: downloadCount,
			RatingCount:   ratingCount,
		}
		if rating.Valid {
			k.Metrics.Rating = &rating.Float64
		}
		if lastViewedAt.Valid {
			k.Metrics.LastViewedAt = &lastViewedAt.Time
		}

		// 反序列化JSON字段
		if tagsJSON.Valid {
			json.Unmarshal([]byte(tagsJSON.String), &k.Tags)
		}
		if keywordsJSON.Valid {
			json.Unmarshal([]byte(keywordsJSON.String), &k.Keywords)
		}
		if templateDataJSON.Valid {
			json.Unmarshal([]byte(templateDataJSON.String), &k.TemplateData)
		}
		if metadataJSON.Valid {
			json.Unmarshal([]byte(metadataJSON.String), &k.Metadata)
		}
		if relatedIDsJSON.Valid {
			json.Unmarshal([]byte(relatedIDsJSON.String), &k.RelatedIDs)
		}

		knowledge = append(knowledge, &k)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("遍历知识数据失败: %w", err)
	}

	return knowledge, nil
}

// AddAttachment 添加附件
func (r *knowledgeRepository) AddAttachment(ctx context.Context, attachment *models.KnowledgeAttachment) error {
	if attachment.ID == "" {
		attachment.ID = uuid.New().String()
	}

	attachment.CreatedAt = time.Now()

	query := `
		INSERT INTO knowledge_attachments (
			id, article_id, filename, original_filename, file_path, file_size, mime_type, uploaded_by, created_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9
		)`

	_, err := r.getExecutor().ExecContext(ctx, query,
			attachment.ID, attachment.KnowledgeID, attachment.FileName,
			attachment.FilePath, attachment.FileSize, attachment.MimeType, attachment.UploadBy, attachment.CreatedAt,
		)

	if err != nil {
		return fmt.Errorf("添加附件失败: %w", err)
	}

	return nil
}

// GetAttachments 获取文章附件
func (r *knowledgeRepository) GetAttachments(ctx context.Context, articleID string) ([]*models.KnowledgeAttachment, error) {
	query := `
		SELECT id, article_id, filename, original_filename, file_path, file_size, mime_type, uploaded_by, created_at
		FROM knowledge_attachments 
		WHERE article_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, articleID)
	if err != nil {
		return nil, fmt.Errorf("获取文章附件失败: %w", err)
	}
	defer rows.Close()

	var attachments []*models.KnowledgeAttachment
	for rows.Next() {
		var attachment models.KnowledgeAttachment
		err := rows.Scan(
			&attachment.ID, &attachment.KnowledgeID, &attachment.FileName,
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

// RemoveAttachment 删除附件
func (r *knowledgeRepository) RemoveAttachment(ctx context.Context, id string) error {
	now := time.Now()
	query := `
		UPDATE knowledge_attachments SET 
			deleted_at = $1
		WHERE id = $2 AND deleted_at IS NULL`

	_, err := r.db.ExecContext(ctx, query, now, id)
	if err != nil {
		return fmt.Errorf("删除附件失败: %w", err)
	}
	return nil
}

// GetMetrics 获取知识库指标
func (r *knowledgeRepository) GetMetrics(ctx context.Context, period string) (*models.KnowledgeMetrics, error) {
	metrics := &models.KnowledgeMetrics{}

	// 获取总浏览数
	err := sqlx.GetContext(ctx, r.db, &metrics.ViewCount, "SELECT COALESCE(SUM(view_count), 0) FROM knowledge WHERE deleted_at IS NULL")
	if err != nil {
		return nil, fmt.Errorf("获取总浏览数失败: %w", err)
	}

	// 获取总点赞数
	err = sqlx.GetContext(ctx, r.db, &metrics.LikeCount, "SELECT COALESCE(SUM(like_count), 0) FROM knowledge WHERE deleted_at IS NULL")
	if err != nil {
		return nil, fmt.Errorf("获取总点赞数失败: %w", err)
	}

	// 获取总评分
	err = sqlx.GetContext(ctx, r.db, &metrics.Rating, "SELECT COALESCE(AVG(rating), 0) FROM knowledge WHERE deleted_at IS NULL AND rating > 0")
	if err != nil {
		return nil, fmt.Errorf("获取平均评分失败: %w", err)
	}

	// 获取评分数量
	err = sqlx.GetContext(ctx, r.db, &metrics.RatingCount, "SELECT COUNT(*) FROM knowledge WHERE deleted_at IS NULL AND rating > 0")
	if err != nil {
		return nil, fmt.Errorf("获取评分数量失败: %w", err)
	}

	return metrics, nil
}

// GetStats 获取知识库统计
func (r *knowledgeRepository) GetStats(ctx context.Context, filter *models.KnowledgeFilter) (*models.KnowledgeStats, error) {
	stats := &models.KnowledgeStats{
		ByStatus:     make(map[models.KnowledgeStatus]int64),
		ByType:       make(map[models.KnowledgeType]int64),
		ByVisibility: make(map[models.KnowledgeVisibility]int64),
		ByFormat:     make(map[models.KnowledgeFormat]int64),
	}

	// 按状态统计
	statusQuery := `
		SELECT status, COUNT(*) 
		FROM knowledge 
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
		stats.ByStatus[models.KnowledgeStatus(status)] = count
		stats.Total += count
		if status == string(models.KnowledgeStatusPublished) {
			stats.PublishedCount = count
		} else if status == string(models.KnowledgeStatusDraft) {
			stats.DraftCount = count
		}
	}

	// 按类型统计
	typeQuery := `
		SELECT type, COUNT(*) 
		FROM knowledge 
		WHERE deleted_at IS NULL 
		GROUP BY type`

	typeRows, err := r.db.QueryContext(ctx, typeQuery)
	if err != nil {
		return nil, fmt.Errorf("按类型统计失败: %w", err)
	}
	defer typeRows.Close()

	for typeRows.Next() {
		var kType string
		var count int64
		err := typeRows.Scan(&kType, &count)
		if err != nil {
			return nil, fmt.Errorf("扫描类型统计失败: %w", err)
		}
		stats.ByType[models.KnowledgeType(kType)] = count
	}

	// 获取其他统计信息
	err = r.getExecutor().QueryRowxContext(ctx, "SELECT COALESCE(SUM(view_count), 0) FROM knowledge WHERE deleted_at IS NULL").Scan(&stats.TotalViews)
	if err != nil {
		return nil, fmt.Errorf("获取总浏览数失败: %w", err)
	}

	err = r.getExecutor().QueryRowxContext(ctx, "SELECT COALESCE(SUM(like_count), 0) FROM knowledge WHERE deleted_at IS NULL").Scan(&stats.TotalLikes)
	if err != nil {
		return nil, fmt.Errorf("获取总点赞数失败: %w", err)
	}

	err = r.getExecutor().QueryRowxContext(ctx, "SELECT COUNT(*) FROM knowledge WHERE deleted_at IS NULL AND is_featured = true").Scan(&stats.FeaturedCount)
	if err != nil {
		return nil, fmt.Errorf("获取推荐数失败: %w", err)
	}

	err = r.getExecutor().QueryRowxContext(ctx, "SELECT COALESCE(AVG(CASE WHEN rating IS NOT NULL THEN rating ELSE 0 END), 0) FROM knowledge WHERE deleted_at IS NULL").Scan(&stats.AvgRating)
	if err != nil {
		return nil, fmt.Errorf("获取平均评分失败: %w", err)
	}

	return stats, nil
}

// BatchCreate 批量创建文章
func (r *knowledgeRepository) BatchCreate(ctx context.Context, articles []*models.KnowledgeArticle) error {
	if len(articles) == 0 {
		return nil
	}

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("开始事务失败: %w", err)
	}
	defer tx.Rollback()

	for _, article := range articles {
		if article.ID == "" {
			article.ID = uuid.New().String()
		}

		now := time.Now()
		article.CreatedAt = now
		article.UpdatedAt = now
		article.Version = "1"

		if article.Status == "" {
			article.Status = models.KnowledgeStatusDraft
		}

		// 序列化标签和元数据
		tagsJSON, err := json.Marshal(article.Tags)
		if err != nil {
			return fmt.Errorf("序列化标签失败: %w", err)
		}

		metadataJSON, err := json.Marshal(article.Metadata)
		if err != nil {
			return fmt.Errorf("序列化元数据失败: %w", err)
		}

		query := `
			INSERT INTO knowledge_articles (
				id, title, content, summary, category_id, status, type, language,
				author_id, reviewer_id, tags, metadata, version, view_count, like_count,
				is_featured, is_public, created_at, updated_at
			) VALUES (
				$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19
			)`

		_, err = tx.ExecContext(ctx, query,
			article.ID, article.Title, article.Content, article.Summary, article.CategoryID,
			article.Status, article.Type, article.Language, article.AuthorID, article.ReviewerID,
			string(tagsJSON), string(metadataJSON), article.Version, article.ViewCount, article.LikeCount,
			article.IsFeatured, article.IsPublic, article.CreatedAt, article.UpdatedAt,
		)
		if err != nil {
			return fmt.Errorf("批量创建文章失败: %w", err)
		}
	}

	return tx.Commit()
}

// BatchUpdate 批量更新文章
func (r *knowledgeRepository) BatchUpdate(ctx context.Context, articles []*models.KnowledgeArticle) error {
	if len(articles) == 0 {
		return nil
	}

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("开始事务失败: %w", err)
	}
	defer tx.Rollback()

	for _, article := range articles {
		article.UpdatedAt = time.Now()
		// 版本号递增（字符串类型）
		if article.Version == "" {
			article.Version = "1"
		} else {
			// 简单的版本号递增，假设版本号是数字字符串
			version := "1"
			if article.Version != "" {
				version = fmt.Sprintf("%s.1", article.Version)
			}
			article.Version = version
		}

		// 序列化标签和元数据
		tagsJSON, err := json.Marshal(article.Tags)
		if err != nil {
			return fmt.Errorf("序列化标签失败: %w", err)
		}

		metadataJSON, err := json.Marshal(article.Metadata)
		if err != nil {
			return fmt.Errorf("序列化元数据失败: %w", err)
		}

		query := `
			UPDATE knowledge_articles SET 
				title = $1,
				content = $2,
				summary = $3,
				category_id = $4,
				status = $5,
				type = $6,
				language = $7,
				reviewer_id = $8,
				tags = $9,
				metadata = $10,
				version = $11,
				is_featured = $12,
				is_public = $13,
				updated_at = $14
			WHERE id = $15 AND deleted_at IS NULL`

		_, err = tx.ExecContext(ctx, query,
			article.Title, article.Content, article.Summary, article.CategoryID,
			article.Status, article.Type, article.Language, article.ReviewerID,
			string(tagsJSON), string(metadataJSON), article.Version,
			article.IsFeatured, article.IsPublic, article.UpdatedAt, article.ID,
		)
		if err != nil {
			return fmt.Errorf("批量更新文章失败: %w", err)
		}
	}

	return tx.Commit()
}

// BatchDelete 批量删除文章
func (r *knowledgeRepository) BatchDelete(ctx context.Context, ids []string) error {
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
			UPDATE knowledge_articles SET 
				deleted_at = $1,
				updated_at = $1
			WHERE id = $2 AND deleted_at IS NULL`

		_, err = tx.ExecContext(ctx, query, now, id)
		if err != nil {
			return fmt.Errorf("批量删除文章失败: %w", err)
		}
	}

	return tx.Commit()
}

// BatchArchive 批量归档文章
func (r *knowledgeRepository) BatchArchive(ctx context.Context, ids []string) error {
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
			UPDATE knowledge_articles SET 
				status = $1,
				updated_at = $2
			WHERE id = $3 AND deleted_at IS NULL`

		_, err = tx.ExecContext(ctx, query, models.KnowledgeStatusArchived, now, id)
		if err != nil {
			return fmt.Errorf("批量归档文章失败: %w", err)
		}
	}

	return tx.Commit()
}

// BatchPublish 批量发布文章
func (r *knowledgeRepository) BatchPublish(ctx context.Context, ids []string, publisherID string) error {
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
			UPDATE knowledge_articles SET 
				status = $1,
				published_at = $2,
				published_by = $3,
				updated_at = $2
			WHERE id = $4 AND deleted_at IS NULL`

		_, err = tx.ExecContext(ctx, query, models.KnowledgeStatusPublished, now, publisherID, id)
		if err != nil {
			return fmt.Errorf("批量发布文章失败: %w", err)
		}
	}

	return tx.Commit()
}

// CleanupDrafts 清理指定时间之前的草稿
func (r *knowledgeRepository) CleanupDrafts(ctx context.Context, before time.Time) (int64, error) {
	query := `
		DELETE FROM knowledge 
		WHERE status = $1 AND created_at < $2
	`
	
	result, err := r.db.ExecContext(ctx, query, models.KnowledgeStatusDraft, before)
	if err != nil {
		return 0, fmt.Errorf("清理草稿失败: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("获取影响行数失败: %w", err)
	}
	
	return rowsAffected, nil
}

// CleanupExpired 清理过期的知识库文章
func (r *knowledgeRepository) CleanupExpired(ctx context.Context) (int64, error) {
	query := `
		UPDATE knowledge 
		SET status = $1, updated_at = CURRENT_TIMESTAMP
		WHERE status = $2 AND expires_at IS NOT NULL AND expires_at < CURRENT_TIMESTAMP
	`
	
	result, err := r.db.ExecContext(ctx, query, models.KnowledgeStatusExpired, models.KnowledgeStatusPublished)
	if err != nil {
		return 0, fmt.Errorf("清理过期文章失败: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("获取影响行数失败: %w", err)
	}
	
	return rowsAffected, nil
}

// CreateTag 创建知识标签
func (r *knowledgeRepository) CreateTag(ctx context.Context, tag *models.KnowledgeTag) error {
	query := `
		INSERT INTO knowledge_tags (id, name, description, color, usage_count, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	
	now := time.Now()
	_, err := r.db.ExecContext(ctx, query,
		tag.ID, tag.Name, tag.Description, tag.Color, tag.UsageCount, now, now)
	if err != nil {
		return fmt.Errorf("创建知识标签失败: %w", err)
	}
	
	return nil
}

// GetTags 获取所有知识标签
func (r *knowledgeRepository) GetTags(ctx context.Context) ([]*models.KnowledgeTag, error) {
	query := `
		SELECT id, name, description, color, usage_count, created_at, updated_at
		FROM knowledge_tags
		ORDER BY usage_count DESC, name ASC
	`
	
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("获取知识标签失败: %w", err)
	}
	defer rows.Close()
	
	var tags []*models.KnowledgeTag
	for rows.Next() {
		var tag models.KnowledgeTag
		err := rows.Scan(
			&tag.ID, &tag.Name, &tag.Description, &tag.Color,
			&tag.UsageCount, &tag.CreatedAt, &tag.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("扫描知识标签失败: %w", err)
		}
		tags = append(tags, &tag)
	}
	
	return tags, nil
}

// GetTag 根据ID获取知识标签
func (r *knowledgeRepository) GetTag(ctx context.Context, id string) (*models.KnowledgeTag, error) {
	query := `
		SELECT id, name, description, color, usage_count, created_at, updated_at
		FROM knowledge_tags
		WHERE id = $1
	`
	
	var tag models.KnowledgeTag
	err := r.getExecutor().QueryRowxContext(ctx, query, id).Scan(
		&tag.ID, &tag.Name, &tag.Description, &tag.Color,
		&tag.UsageCount, &tag.CreatedAt, &tag.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("获取知识标签失败: %w", err)
	}
	
	return &tag, nil
}

// UpdateTag 更新知识标签
func (r *knowledgeRepository) UpdateTag(ctx context.Context, tag *models.KnowledgeTag) error {
	query := `
		UPDATE knowledge_tags
		SET name = $1, description = $2, color = $3, updated_at = $4
		WHERE id = $5
	`
	
	now := time.Now()
	_, err := r.db.ExecContext(ctx, query,
		tag.Name, tag.Description, tag.Color, now, tag.ID)
	if err != nil {
		return fmt.Errorf("更新知识标签失败: %w", err)
	}
	
	return nil
}

// DeleteTag 删除知识标签
func (r *knowledgeRepository) DeleteTag(ctx context.Context, id string) error {
	query := `DELETE FROM knowledge_tags WHERE id = $1`
	
	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("删除知识标签失败: %w", err)
	}
	
	return nil
}

// GetKnowledgeByTag 根据标签获取知识
func (r *knowledgeRepository) GetKnowledgeByTag(ctx context.Context, tagName string, filter *models.KnowledgeFilter) (*models.KnowledgeList, error) {
	// 这里简化实现，实际应该根据标签关联查询
	filter.Tags = []string{tagName}
	return r.List(ctx, filter)
}

// UpdateTagUsage 更新标签使用次数
func (r *knowledgeRepository) UpdateTagUsage(ctx context.Context, tagName string, delta int64) error {
	query := `
		UPDATE knowledge_tags
		SET usage_count = usage_count + $1, updated_at = CURRENT_TIMESTAMP
		WHERE name = $2
	`
	
	_, err := r.db.ExecContext(ctx, query, delta, tagName)
	if err != nil {
		return fmt.Errorf("更新标签使用次数失败: %w", err)
	}
	
	return nil
}

// DeleteAttachment 删除知识库附件
func (r *knowledgeRepository) DeleteAttachment(ctx context.Context, attachmentID string) error {
	query := `
		UPDATE knowledge_attachments 
		SET deleted_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND deleted_at IS NULL
	`
	
	_, err := r.db.ExecContext(ctx, query, attachmentID)
	if err != nil {
		return fmt.Errorf("删除知识库附件失败: %w", err)
	}
	
	return nil
}

// GetCategory 根据ID获取知识分类
func (r *knowledgeRepository) GetCategory(ctx context.Context, id string) (*models.KnowledgeCategory, error) {
	query := `
		SELECT id, name, description, parent_id, path, level, sort_order, icon, color, is_active, created_at, updated_at
		FROM knowledge_categories
		WHERE id = $1 AND deleted_at IS NULL
	`
	
	var category models.KnowledgeCategory
	err := r.getExecutor().QueryRowxContext(ctx, query, id).Scan(
		&category.ID, &category.Name, &category.Description, &category.ParentID,
		&category.Path, &category.Level, &category.SortOrder, &category.Icon,
		&category.Color, &category.IsActive, &category.CreatedAt, &category.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("获取知识分类失败: %w", err)
	}
	
	return &category, nil
}

// GetFeatured 获取推荐知识列表
func (r *knowledgeRepository) GetFeatured(ctx context.Context, limit int) ([]*models.Knowledge, error) {
	query := `
		SELECT id, title, slug, summary, content, type, status, visibility, format, 
		       category_id, tags, keywords, language, version, author_id, author_name,
		       reviewer_id, reviewer_name, team_id, team_name, priority, sort_order,
		       is_featured, is_template, template_data, metadata, view_count, like_count,
		       related_ids, expires_at, published_at, archived_at, reviewed_at,
		       last_edited_at, created_at, updated_at
		FROM knowledge
		WHERE is_featured = true AND status = $1 AND deleted_at IS NULL
		ORDER BY sort_order ASC, created_at DESC
		LIMIT $2
	`
	
	rows, err := r.db.QueryContext(ctx, query, models.KnowledgeStatusPublished, limit)
	if err != nil {
		return nil, fmt.Errorf("获取推荐知识列表失败: %w", err)
	}
	defer rows.Close()
	
	var knowledgeList []*models.Knowledge
	for rows.Next() {
		var knowledge models.Knowledge
		err := rows.Scan(
			&knowledge.ID, &knowledge.Title, &knowledge.Slug, &knowledge.Summary,
			&knowledge.Content, &knowledge.Type, &knowledge.Status, &knowledge.Visibility,
			&knowledge.Format, &knowledge.CategoryID, &knowledge.Tags, &knowledge.Keywords,
			&knowledge.Language, &knowledge.Version, &knowledge.AuthorID, &knowledge.AuthorName,
			&knowledge.ReviewerID, &knowledge.ReviewerName, &knowledge.TeamID, &knowledge.TeamName,
			&knowledge.Priority, &knowledge.SortOrder, &knowledge.IsFeatured, &knowledge.IsTemplate,
			&knowledge.TemplateData, &knowledge.Metadata, &knowledge.ViewCount, &knowledge.LikeCount,
			&knowledge.RelatedIDs, &knowledge.ExpiresAt, &knowledge.PublishedAt, &knowledge.ArchivedAt,
			&knowledge.ReviewedAt, &knowledge.LastEditedAt, &knowledge.CreatedAt, &knowledge.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("扫描推荐知识数据失败: %w", err)
		}
		knowledgeList = append(knowledgeList, &knowledge)
	}
	
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("遍历推荐知识数据失败: %w", err)
	}
	
	return knowledgeList, nil
}

// GetKnowledgeByCategory 根据分类获取知识列表
func (r *knowledgeRepository) GetKnowledgeByCategory(ctx context.Context, categoryID string, filter *models.KnowledgeFilter) (*models.KnowledgeList, error) {
	// 设置默认值
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 {
		filter.PageSize = 20
	}
	
	offset := (filter.Page - 1) * filter.PageSize
	
	// 获取总数
	countQuery := `
		SELECT COUNT(*)
		FROM knowledge
		WHERE category_id = $1 AND status = $2 AND deleted_at IS NULL
	`
	
	var total int64
	err := r.getExecutor().QueryRowxContext(ctx, countQuery, categoryID, models.KnowledgeStatusPublished).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("获取分类知识总数失败: %w", err)
	}
	
	// 获取知识列表
	query := `
		SELECT id, title, slug, summary, content, type, status, visibility, format, 
		       category_id, tags, keywords, language, version, author_id, author_name,
		       reviewer_id, reviewer_name, team_id, team_name, priority, sort_order,
		       is_featured, is_template, template_data, metadata, view_count, like_count,
		       related_ids, expires_at, published_at, archived_at, reviewed_at,
		       last_edited_at, created_at, updated_at
		FROM knowledge
		WHERE category_id = $1 AND status = $2 AND deleted_at IS NULL
		ORDER BY sort_order ASC, created_at DESC
		LIMIT $3 OFFSET $4
	`
	
	rows, err := r.db.QueryContext(ctx, query, categoryID, models.KnowledgeStatusPublished, filter.PageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("获取分类知识列表失败: %w", err)
	}
	defer rows.Close()
	
	var knowledgeList []*models.Knowledge
	for rows.Next() {
		var knowledge models.Knowledge
		err := rows.Scan(
			&knowledge.ID, &knowledge.Title, &knowledge.Slug, &knowledge.Summary,
			&knowledge.Content, &knowledge.Type, &knowledge.Status, &knowledge.Visibility,
			&knowledge.Format, &knowledge.CategoryID, &knowledge.Tags, &knowledge.Keywords,
			&knowledge.Language, &knowledge.Version, &knowledge.AuthorID, &knowledge.AuthorName,
			&knowledge.ReviewerID, &knowledge.ReviewerName, &knowledge.TeamID, &knowledge.TeamName,
			&knowledge.Priority, &knowledge.SortOrder, &knowledge.IsFeatured, &knowledge.IsTemplate,
			&knowledge.TemplateData, &knowledge.Metadata, &knowledge.ViewCount, &knowledge.LikeCount,
			&knowledge.RelatedIDs, &knowledge.ExpiresAt, &knowledge.PublishedAt, &knowledge.ArchivedAt,
			&knowledge.ReviewedAt, &knowledge.LastEditedAt, &knowledge.CreatedAt, &knowledge.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("扫描分类知识数据失败: %w", err)
		}
		knowledgeList = append(knowledgeList, &knowledge)
	}
	
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("遍历分类知识数据失败: %w", err)
	}
	
	// 计算总页数
	totalPages := int((total + int64(filter.PageSize) - 1) / int64(filter.PageSize))
	
	return &models.KnowledgeList{
		Knowledge:  knowledgeList,
		Total:      total,
		Page:       filter.Page,
		PageSize:   filter.PageSize,
		TotalPages: totalPages,
	}, nil
}

// Search 搜索知识库
func (r *knowledgeRepository) Search(ctx context.Context, query string, filter *models.KnowledgeFilter) (*models.KnowledgeSearchResult, error) {
	var conditions []string
	var args []interface{}
	argIndex := 1
	
	// 基础搜索条件
	conditions = append(conditions, "deleted_at IS NULL")
	
	// 全文搜索
	if query != "" {
		conditions = append(conditions, "(title ILIKE $"+fmt.Sprintf("%d", argIndex)+" OR content ILIKE $"+fmt.Sprintf("%d", argIndex)+")")
		args = append(args, "%"+query+"%")
		argIndex++
	}
	
	// 应用过滤器
	if filter != nil {
		if filter.Status != nil {
			conditions = append(conditions, "status = $"+fmt.Sprintf("%d", argIndex))
			args = append(args, *filter.Status)
			argIndex++
		}
		
		if filter.CategoryID != nil {
			conditions = append(conditions, "category_id = $"+fmt.Sprintf("%d", argIndex))
			args = append(args, *filter.CategoryID)
			argIndex++
		}
		
		if filter.AuthorID != nil {
			conditions = append(conditions, "author_id = $"+fmt.Sprintf("%d", argIndex))
			args = append(args, *filter.AuthorID)
			argIndex++
		}
	}
	
	// 构建查询
	baseQuery := `
		SELECT id, title, slug, summary, content, status, category_id, author_id,
		       view_count, like_count, dislike_count, share_count, download_count,
		       rating, rating_count, featured, created_at, updated_at, published_at
		FROM knowledge
		WHERE ` + strings.Join(conditions, " AND ")
	
	// 排序
	orderBy := "ORDER BY created_at DESC"
	if filter != nil && filter.SortBy != nil {
		switch *filter.SortBy {
		case "title":
			orderBy = "ORDER BY title"
		case "view_count":
			orderBy = "ORDER BY view_count DESC"
		case "rating":
			orderBy = "ORDER BY rating DESC"
		case "updated_at":
			orderBy = "ORDER BY updated_at DESC"
		}
		
		if filter.SortOrder != nil && *filter.SortOrder == "asc" {
			orderBy = strings.Replace(orderBy, "DESC", "ASC", 1)
		}
	}
	
	// 分页
	limit := 20
	offset := 0
	if filter != nil {
		if filter.PageSize > 0 {
			limit = filter.PageSize
		}
		if filter.Page > 0 {
			offset = (filter.Page - 1) * limit
		}
	}
	
	finalQuery := baseQuery + " " + orderBy + " LIMIT $" + fmt.Sprintf("%d", argIndex) + " OFFSET $" + fmt.Sprintf("%d", argIndex+1)
	args = append(args, limit, offset)
	
	// 执行查询
	rows, err := r.db.QueryContext(ctx, finalQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("搜索知识库失败: %w", err)
	}
	defer rows.Close()
	
	var knowledgeList []*models.Knowledge
	for rows.Next() {
		var knowledge models.Knowledge
		err := rows.Scan(
			&knowledge.ID, &knowledge.Title, &knowledge.Slug, &knowledge.Summary,
			&knowledge.Content, &knowledge.Status, &knowledge.CategoryID, &knowledge.AuthorID,
			&knowledge.ViewCount, &knowledge.LikeCount, &knowledge.DislikeCount,
			&knowledge.ShareCount, &knowledge.DownloadCount, &knowledge.Rating,
			&knowledge.RatingCount, &knowledge.Featured, &knowledge.CreatedAt,
			&knowledge.UpdatedAt, &knowledge.PublishedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("扫描知识库记录失败: %w", err)
		}
		knowledgeList = append(knowledgeList, &knowledge)
	}
	
	// 获取总数
	countQuery := "SELECT COUNT(*) FROM knowledge WHERE " + strings.Join(conditions, " AND ")
	var total int64
	err = r.getExecutor().QueryRowxContext(ctx, countQuery, args[:len(args)-2]...).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("获取搜索结果总数失败: %w", err)
	}
	
	// 计算分页信息
	totalPages := (total + int64(limit) - 1) / int64(limit)
	page := 1
	if filter != nil && filter.Page > 0 {
		page = filter.Page
	}
	
	return &models.KnowledgeSearchResult{
		Knowledge:  knowledgeList,
		Total:      total,
		Page:       page,
		PageSize:   limit,
		TotalPages: int(totalPages),
		Query:      query,
	}, nil
}

// SubmitForReview 提交知识库进行审核
func (r *knowledgeRepository) SubmitForReview(ctx context.Context, id string) error {
	// 检查知识库是否存在且为草稿状态
	var currentStatus models.KnowledgeStatus
	query := `SELECT status FROM knowledge WHERE id = $1 AND deleted_at IS NULL`
	err := r.getExecutor().QueryRowxContext(ctx, query, id).Scan(&currentStatus)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("知识库不存在")
		}
		return fmt.Errorf("获取知识库状态失败: %w", err)
	}
	
	// 只有草稿状态的知识库才能提交审核
	if currentStatus != models.KnowledgeStatusDraft {
		return fmt.Errorf("只有草稿状态的知识库才能提交审核")
	}
	
	// 更新状态为审核中
	updateQuery := `
		UPDATE knowledge 
		SET status = $1, updated_at = CURRENT_TIMESTAMP
		WHERE id = $2 AND deleted_at IS NULL`
	
	_, err = r.db.ExecContext(ctx, updateQuery, models.KnowledgeStatusReview, id)
	if err != nil {
		return fmt.Errorf("提交审核失败: %w", err)
	}
	
	return nil
}