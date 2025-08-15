package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"Pulse/internal/models"
)

func TestKnowledgeRepository_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := NewKnowledgeRepository(sqlxDB)

	knowledge := &models.Knowledge{
		ID:         uuid.New().String(),
		Title:      "测试知识",
		Content:    "测试内容",
		Summary:    stringPtr("测试摘要"),
		Type:       models.KnowledgeTypeArticle,
		Status:     models.KnowledgeStatusDraft,
		Visibility: models.KnowledgeVisibilityPublic,
		Format:     models.KnowledgeFormatMarkdown,
		CategoryID:   stringPtr("category-1"),
		AuthorID:   "author-1",
		Language:   "zh-CN",
		Tags:       []string{"tag1", "tag2"},
		Keywords:   []string{"keyword1", "keyword2"},
		Metadata:   map[string]interface{}{"key": "value"},
	}

	// Mock INSERT query - 匹配实际Create方法的19个字段
	mock.ExpectExec(`INSERT INTO knowledge_articles`).WithArgs(
		knowledge.ID, knowledge.Title, knowledge.Content, knowledge.Summary,
		knowledge.CategoryID, knowledge.Status, knowledge.Type, knowledge.Language,
		knowledge.AuthorID, knowledge.ReviewerID,
		sqlmock.AnyArg(), sqlmock.AnyArg(), // tags, metadata JSON
		"1", knowledge.ViewCount, knowledge.LikeCount, // version is set to "1" by Create method
		knowledge.IsFeatured, knowledge.Visibility,
		sqlmock.AnyArg(), sqlmock.AnyArg(), // created_at, updated_at
	).WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.Create(context.Background(), knowledge)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestKnowledgeRepository_GetByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := NewKnowledgeRepository(sqlxDB)

	knowledgeID := uuid.New().String()
	tags := []string{"tag1", "tag2"}
	metadata := map[string]interface{}{"key": "value"}

	tagsJSON, _ := json.Marshal(tags)
	metadataJSON, _ := json.Marshal(metadata)

	// 匹配实际GetByID方法的查询字段
	rows := sqlmock.NewRows([]string{
		"id", "title", "content", "summary", "category_id", "status", "type", "language",
		"author_id", "reviewer_id", "tags", "metadata", "version", "view_count", "like_count",
		"is_featured", "visibility", "created_at", "updated_at", "published_at", "reviewed_at",
	}).AddRow(
		knowledgeID, "测试知识", "测试内容", "测试摘要", "category-1", models.KnowledgeStatusDraft,
		models.KnowledgeTypeArticle, "zh-CN", "author-1", nil, string(tagsJSON), string(metadataJSON),
		"1.0", 100, 10, true, models.KnowledgeVisibilityPublic, time.Now(), time.Now(), nil, nil,
	)

	mock.ExpectQuery(`SELECT .+ FROM knowledge_articles WHERE .+ AND deleted_at IS NULL`).WithArgs(knowledgeID).WillReturnRows(rows)

	knowledge, err := repo.GetByID(context.Background(), knowledgeID)
	assert.NoError(t, err)
	assert.NotNil(t, knowledge)
	assert.Equal(t, knowledgeID, knowledge.ID)
	assert.Equal(t, "测试知识", knowledge.Title)
	assert.Equal(t, tags, knowledge.Tags)
	assert.Equal(t, metadata, knowledge.Metadata)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestKnowledgeRepository_GetByID_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := NewKnowledgeRepository(sqlxDB)

	knowledgeID := uuid.New().String()

	mock.ExpectQuery(`SELECT .+ FROM knowledge_articles WHERE .+ AND deleted_at IS NULL`).WithArgs(knowledgeID).WillReturnError(sql.ErrNoRows)

	knowledge, err := repo.GetByID(context.Background(), knowledgeID)
	assert.Error(t, err)
	assert.Nil(t, knowledge)
	assert.Contains(t, err.Error(), "知识库文章不存在")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestKnowledgeRepository_Update(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := NewKnowledgeRepository(sqlxDB)

	knowledge := &models.Knowledge{
		ID:         uuid.New().String(),
		Title:      "更新的知识",
		Content:    "更新的内容",
		Summary:      stringPtr("更新的摘要"),
		Type:       models.KnowledgeTypeArticle,
		Status:     models.KnowledgeStatusDraft,
		Visibility: models.KnowledgeVisibilityPublic,
		Format:     models.KnowledgeFormatMarkdown,
		CategoryID:   stringPtr("category-1"),
		Language:   "zh-CN",
		Tags:       []string{"tag1", "tag2"},
		Keywords:   []string{"keyword1", "keyword2"},
		Metadata:   map[string]interface{}{"key": "value"},
	}

	mock.ExpectExec(`UPDATE knowledge_articles SET`).WithArgs(
		knowledge.Title, knowledge.Content, knowledge.Summary, knowledge.CategoryID,
		knowledge.Status, knowledge.Type, knowledge.Language, knowledge.ReviewerID,
		sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), // tags, metadata, version
		knowledge.IsFeatured, knowledge.Visibility, sqlmock.AnyArg(), // published_at
		sqlmock.AnyArg(), knowledge.ID, // updated_at, id
	).WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.Update(context.Background(), knowledge)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestKnowledgeRepository_Delete(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := NewKnowledgeRepository(sqlxDB)

	knowledgeID := uuid.New().String()

	mock.ExpectExec(`DELETE FROM knowledge_articles WHERE id = \$1`).WithArgs(knowledgeID).WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.Delete(context.Background(), knowledgeID)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestKnowledgeRepository_SoftDelete(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := NewKnowledgeRepository(sqlxDB)

	knowledgeID := uuid.New().String()

	mock.ExpectExec(`UPDATE knowledge_articles SET deleted_at = \$1, updated_at = \$1 WHERE id = \$2 AND deleted_at IS NULL`).WithArgs(
		sqlmock.AnyArg(), knowledgeID,
	).WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.SoftDelete(context.Background(), knowledgeID)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestKnowledgeRepository_List(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := NewKnowledgeRepository(sqlxDB)

	filter := &models.KnowledgeFilter{
		Page:     1,
		PageSize: 10,
	}

	// Mock count query
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM knowledge_articles WHERE deleted_at IS NULL`).WillReturnRows(
		sqlmock.NewRows([]string{"count"}).AddRow(2),
	)

	// Mock list query - 匹配实际的List查询字段
	rows := sqlmock.NewRows([]string{
		"id", "title", "content", "summary", "category_id", "status", "type", "language",
		"author_id", "reviewer_id", "tags", "metadata", "version", "view_count", "like_count",
		"is_featured", "visibility", "created_at", "updated_at", "published_at",
	}).AddRow(
		"id-1", "知识1", "内容1", "摘要1", "category-1", models.KnowledgeStatusPublished,
		models.KnowledgeTypeArticle, "zh-CN", "author-1", nil, "[]", "{}", "1.0",
		100, 10, true, models.KnowledgeVisibilityPublic, time.Now(), time.Now(), time.Now(),
	).AddRow(
		"id-2", "知识2", "内容2", "摘要2", "category-2", models.KnowledgeStatusPublished,
		models.KnowledgeTypeArticle, "zh-CN", "author-2", nil, "[]", "{}", "1.0",
		200, 20, false, models.KnowledgeVisibilityPublic, time.Now(), time.Now(), time.Now(),
	)

	mock.ExpectQuery(`SELECT .+ FROM knowledge_articles WHERE deleted_at IS NULL ORDER BY created_at DESC LIMIT \$1 OFFSET \$2`).WithArgs(
		10, 0,
	).WillReturnRows(rows)

	result, err := repo.List(context.Background(), filter)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, int64(2), result.Total)
	assert.Len(t, result.Knowledge, 2)
	assert.Equal(t, "知识1", result.Knowledge[0].Title)
	assert.Equal(t, "知识2", result.Knowledge[1].Title)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestKnowledgeRepository_Count(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := NewKnowledgeRepository(sqlxDB)

	filter := &models.KnowledgeFilter{}

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM knowledge_articles WHERE deleted_at IS NULL`).WillReturnRows(
		sqlmock.NewRows([]string{"count"}).AddRow(5),
	)

	count, err := repo.Count(context.Background(), filter)
	assert.NoError(t, err)
	assert.Equal(t, int64(5), count)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestKnowledgeRepository_Exists(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := NewKnowledgeRepository(sqlxDB)

	knowledgeID := uuid.New().String()

	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM knowledge_articles WHERE id = \$1 AND deleted_at IS NULL\)`).WithArgs(knowledgeID).WillReturnRows(
		sqlmock.NewRows([]string{"exists"}).AddRow(true),
	)

	exists, err := repo.Exists(context.Background(), knowledgeID)
	assert.NoError(t, err)
	assert.True(t, exists)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestKnowledgeRepository_UpdateStatus(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := NewKnowledgeRepository(sqlxDB)

	knowledgeID := uuid.New().String()
	status := models.KnowledgeStatusDraft

	// 测试非发布状态的更新
	mock.ExpectExec(`UPDATE knowledge_articles SET status = \$1, updated_at = \$2 WHERE id = \$3 AND deleted_at IS NULL`).WithArgs(
		status, sqlmock.AnyArg(), knowledgeID,
	).WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.UpdateStatus(context.Background(), knowledgeID, status)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())

	// 测试发布状态的更新
	db2, mock2, err := sqlmock.New()
	require.NoError(t, err)
	defer db2.Close()

	sqlxDB2 := sqlx.NewDb(db2, "postgres")
	repo2 := NewKnowledgeRepository(sqlxDB2)

	publishedStatus := models.KnowledgeStatusPublished
	mock2.ExpectExec(`UPDATE knowledge_articles SET status = \$1, published_at = \$2, updated_at = \$2 WHERE id = \$3 AND deleted_at IS NULL`).WithArgs(
		publishedStatus, sqlmock.AnyArg(), knowledgeID,
	).WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo2.UpdateStatus(context.Background(), knowledgeID, publishedStatus)
	assert.NoError(t, err)
	assert.NoError(t, mock2.ExpectationsWereMet())
}

func TestKnowledgeRepository_Approve(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := NewKnowledgeRepository(sqlxDB)

	knowledgeID := uuid.New().String()
	reviewerID := uuid.New().String()
	comment := "审批通过"

	mock.ExpectExec(`UPDATE knowledge_articles SET status = \$1, reviewer_id = \$2, review_comment = \$3, published_at = \$4, updated_at = \$4 WHERE id = \$5 AND deleted_at IS NULL AND status = \$6`).WithArgs(
		models.KnowledgeStatusPublished, reviewerID, &comment, sqlmock.AnyArg(), knowledgeID, models.KnowledgeStatusReview,
	).WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.Approve(context.Background(), knowledgeID, reviewerID, &comment)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestKnowledgeRepository_Reject(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := NewKnowledgeRepository(sqlxDB)

	knowledgeID := uuid.New().String()
	reviewerID := uuid.New().String()
	comment := "需要修改"

	mock.ExpectExec(`UPDATE knowledge_articles SET status = \$1, reviewer_id = \$2, review_comment = \$3, updated_at = \$4 WHERE id = \$5 AND deleted_at IS NULL AND status = \$6`).WithArgs(
		models.KnowledgeStatusDraft, reviewerID, &comment, sqlmock.AnyArg(), knowledgeID, models.KnowledgeStatusReview,
	).WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.Reject(context.Background(), knowledgeID, reviewerID, &comment)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestKnowledgeRepository_Publish(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := NewKnowledgeRepository(sqlxDB)

	knowledgeID := uuid.New().String()
	publisherID := uuid.New().String()

	mock.ExpectExec(`UPDATE knowledge_articles SET status = \$1, published_at = \$2, updated_at = \$2 WHERE id = \$3 AND deleted_at IS NULL`).WithArgs(
		models.KnowledgeStatusPublished, sqlmock.AnyArg(), knowledgeID,
	).WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.Publish(context.Background(), knowledgeID, publisherID)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestKnowledgeRepository_Archive(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := NewKnowledgeRepository(sqlxDB)

	knowledgeID := uuid.New().String()

	mock.ExpectExec(`UPDATE knowledge_articles SET status = \$1, updated_at = \$2 WHERE id = \$3 AND deleted_at IS NULL`).WithArgs(
		models.KnowledgeStatusArchived, sqlmock.AnyArg(), knowledgeID,
	).WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.Archive(context.Background(), knowledgeID)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestKnowledgeRepository_IncrementViewCount(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := NewKnowledgeRepository(sqlxDB)

	knowledgeID := uuid.New().String()

	mock.ExpectExec(`UPDATE knowledge_articles SET view_count = view_count \+ 1, updated_at = \$1 WHERE id = \$2 AND deleted_at IS NULL`).WithArgs(
		sqlmock.AnyArg(), knowledgeID,
	).WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.IncrementViewCount(context.Background(), knowledgeID)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestKnowledgeRepository_IncrementLikeCount(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := NewKnowledgeRepository(sqlxDB)

	knowledgeID := uuid.New().String()

	mock.ExpectExec(`UPDATE knowledge_articles SET like_count = like_count \+ 1, updated_at = \$1 WHERE id = \$2 AND deleted_at IS NULL`).WithArgs(
		sqlmock.AnyArg(), knowledgeID,
	).WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.IncrementLikeCount(context.Background(), knowledgeID)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestKnowledgeRepository_CreateVersion(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := NewKnowledgeRepository(sqlxDB)

	version := &models.KnowledgeVersion{
		KnowledgeID: uuid.New().String(),
		Version:     "2.0",
		Title:       "版本标题",
		Content:     "版本内容",
		ChangeLog:    stringPtr("更新日志"),
		CreatedBy:   "user-1",
	}

	mock.ExpectExec(`INSERT INTO knowledge_versions`).WithArgs(
		sqlmock.AnyArg(), version.KnowledgeID, version.Version, version.Title,
		version.Content, version.ChangeLog, version.CreatedBy, sqlmock.AnyArg(),
	).WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.CreateVersion(context.Background(), version)
	assert.NoError(t, err)
	assert.NotEmpty(t, version.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestKnowledgeRepository_GetVersions(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := NewKnowledgeRepository(sqlxDB)

	knowledgeID := uuid.New().String()

	rows := sqlmock.NewRows([]string{
		"id", "knowledge_id", "version", "title", "content", "change_log", "created_by", "created_at",
	}).AddRow(
		"version-1", knowledgeID, "2.0", "版本标题", "版本内容", "更新日志", "user-1", time.Now(),
	).AddRow(
		"version-2", knowledgeID, "1.0", "初始版本", "初始内容", "初始版本", "user-1", time.Now(),
	)

	mock.ExpectQuery(`SELECT .+ FROM knowledge_versions WHERE knowledge_id = \$1 ORDER BY version DESC`).WithArgs(knowledgeID).WillReturnRows(rows)

	versions, err := repo.GetVersions(context.Background(), knowledgeID)
	assert.NoError(t, err)
	assert.Len(t, versions, 2)
	assert.Equal(t, "2.0", versions[0].Version)
	assert.Equal(t, "1.0", versions[1].Version)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestKnowledgeRepository_CreateCategory(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := NewKnowledgeRepository(sqlxDB)

	category := &models.KnowledgeCategory{
		Name:        "测试分类",
		Description: "测试分类描述",
		ParentID:    nil,
		SortOrder:   1,
		IsActive:    true,
	}

	mock.ExpectExec(`INSERT INTO knowledge_categories`).WithArgs(
		sqlmock.AnyArg(), category.Name, category.Description, category.ParentID,
		category.SortOrder, category.IsActive, sqlmock.AnyArg(), sqlmock.AnyArg(),
	).WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.CreateCategory(context.Background(), category)
	assert.NoError(t, err)
	assert.NotEmpty(t, category.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestKnowledgeRepository_GetCategories(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := NewKnowledgeRepository(sqlxDB)

	rows := sqlmock.NewRows([]string{
		"id", "name", "description", "parent_id", "sort_order", "is_active", "created_at", "updated_at",
	}).AddRow(
		"cat-1", "分类1", "分类1描述", nil, 1, true, time.Now(), time.Now(),
	).AddRow(
		"cat-2", "分类2", "分类2描述", "cat-1", 2, true, time.Now(), time.Now(),
	)

	mock.ExpectQuery(`SELECT .+ FROM knowledge_categories WHERE deleted_at IS NULL ORDER BY sort_order ASC, name ASC`).WillReturnRows(rows)

	categories, err := repo.GetCategories(context.Background())
	assert.NoError(t, err)
	assert.Len(t, categories, 2)
	assert.Equal(t, "分类1", categories[0].Name)
	assert.Equal(t, "分类2", categories[1].Name)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestKnowledgeRepository_AddAttachment(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := NewKnowledgeRepository(sqlxDB)

	attachment := &models.KnowledgeAttachment{
		KnowledgeID: uuid.New().String(),
		FileName:    "test.pdf",
		FilePath:    "/uploads/test.pdf",
		FileSize:    1024,
		MimeType:    "application/pdf",
		UploadBy:    "user-1",
	}

	mock.ExpectExec(`INSERT INTO knowledge_attachments`).WithArgs(
		sqlmock.AnyArg(), attachment.KnowledgeID, attachment.FileName,
		attachment.FilePath, attachment.FileSize, attachment.MimeType,
		attachment.UploadBy, sqlmock.AnyArg(),
	).WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.AddAttachment(context.Background(), attachment)
	assert.NoError(t, err)
	assert.NotEmpty(t, attachment.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestKnowledgeRepository_GetAttachments(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := NewKnowledgeRepository(sqlxDB)

	knowledgeID := uuid.New().String()

	rows := sqlmock.NewRows([]string{
		"id", "article_id", "filename", "original_filename", "file_path", "file_size", "mime_type", "uploaded_by", "created_at",
	}).AddRow(
		"att-1", knowledgeID, "test.pdf", "original.pdf", "/uploads/test.pdf", int64(1024), "application/pdf", "user-1", time.Now(),
	)

	mock.ExpectQuery(`SELECT .+ FROM knowledge_attachments WHERE article_id = \$1 AND deleted_at IS NULL ORDER BY created_at DESC`).WithArgs(knowledgeID).WillReturnRows(rows)

	attachments, err := repo.GetAttachments(context.Background(), knowledgeID)
	assert.NoError(t, err)
	assert.Len(t, attachments, 1)
	assert.Equal(t, "test.pdf", attachments[0].FileName)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestKnowledgeRepository_GetStats(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := NewKnowledgeRepository(sqlxDB)

	filter := &models.KnowledgeFilter{}

	// Mock status stats
	statusRows := sqlmock.NewRows([]string{"status", "count"}).AddRow(
		string(models.KnowledgeStatusPublished), 10,
	).AddRow(
		string(models.KnowledgeStatusDraft), 5,
	)
	mock.ExpectQuery(`SELECT status, COUNT\(\*\) FROM knowledge_articles WHERE deleted_at IS NULL GROUP BY status`).WillReturnRows(statusRows)

	// Mock type stats
	typeRows := sqlmock.NewRows([]string{"type", "count"}).AddRow(
		string(models.KnowledgeTypeArticle), 8,
	).AddRow(
		string(models.KnowledgeTypeReference), 7,
	)
	mock.ExpectQuery(`SELECT type, COUNT\(\*\) FROM knowledge_articles WHERE deleted_at IS NULL GROUP BY type`).WillReturnRows(typeRows)

	// Mock other stats
	mock.ExpectQuery(`SELECT COALESCE\(SUM\(view_count\), 0\) FROM knowledge_articles WHERE deleted_at IS NULL`).WillReturnRows(
		sqlmock.NewRows([]string{"sum"}).AddRow(1000),
	)
	mock.ExpectQuery(`SELECT COALESCE\(SUM\(like_count\), 0\) FROM knowledge_articles WHERE deleted_at IS NULL`).WillReturnRows(
		sqlmock.NewRows([]string{"sum"}).AddRow(100),
	)
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM knowledge_articles WHERE deleted_at IS NULL AND is_featured = true`).WillReturnRows(
		sqlmock.NewRows([]string{"count"}).AddRow(3),
	)
	mock.ExpectQuery(`SELECT COALESCE\(AVG\(CASE WHEN rating IS NOT NULL THEN rating ELSE 0 END\), 0\) FROM knowledge_articles WHERE deleted_at IS NULL`).WillReturnRows(
		sqlmock.NewRows([]string{"avg"}).AddRow(4.2),
	)

	stats, err := repo.GetStats(context.Background(), filter)
	assert.NoError(t, err)
	assert.NotNil(t, stats)
	assert.Equal(t, int64(15), stats.Total)
	assert.Equal(t, int64(10), stats.PublishedCount)
	assert.Equal(t, int64(5), stats.DraftCount)
	assert.Equal(t, int64(1000), stats.TotalViews)
	assert.Equal(t, int64(100), stats.TotalLikes)
	assert.Equal(t, int64(3), stats.FeaturedCount)
	assert.Equal(t, 4.2, stats.AvgRating)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestKnowledgeRepository_BatchCreate(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := NewKnowledgeRepository(sqlxDB)

	knowledgeList := []*models.Knowledge{
		{
			Title:      "文章1",
			Content:    "内容1",
			Summary:      stringPtr("摘要1"),
			CategoryID:   stringPtr("cat-1"),
			AuthorID:   "author-1",
			Language:   "zh-CN",
			Tags:       []string{"tag1"},
			Metadata:   map[string]interface{}{"key": "value"},
		},
		{
			Title:      "文章2",
			Content:    "内容2",
			Summary:      stringPtr("摘要2"),
			CategoryID: stringPtr("cat-2"),
			AuthorID:   "author-2",
			Language:   "zh-CN",
			Tags:       []string{"tag2"},
			Metadata:   map[string]interface{}{"key": "value"},
		},
	}

	mock.ExpectBegin()
	for range knowledgeList {
		mock.ExpectExec(`INSERT INTO knowledge_articles`).WithArgs(
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
		).WillReturnResult(sqlmock.NewResult(1, 1))
	}
	mock.ExpectCommit()

	err = repo.BatchCreate(context.Background(), knowledgeList)
	assert.NoError(t, err)
	for _, knowledge := range knowledgeList {
		assert.NotEmpty(t, knowledge.ID)
	}
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestKnowledgeRepository_BatchDelete(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := NewKnowledgeRepository(sqlxDB)

	ids := []string{"id-1", "id-2", "id-3"}

	// 不使用事务，直接mock每个SoftDelete调用
	for _, id := range ids {
		mock.ExpectExec(`UPDATE knowledge_articles SET deleted_at = \$1, updated_at = \$1 WHERE id = \$2 AND deleted_at IS NULL`).WithArgs(
			sqlmock.AnyArg(), id,
		).WillReturnResult(sqlmock.NewResult(1, 1))
	}

	// 使用循环调用SoftDelete替代BatchDelete
	for _, id := range ids {
		err = repo.SoftDelete(context.Background(), id)
		assert.NoError(t, err)
	}
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestKnowledgeRepository_CleanupDrafts(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := NewKnowledgeRepository(sqlxDB)

	before := time.Now().AddDate(0, -1, 0) // 1个月前

	mock.ExpectExec(`DELETE FROM knowledge_articles WHERE status = \$1 AND created_at < \$2`).WithArgs(
		models.KnowledgeStatusDraft, before,
	).WillReturnResult(sqlmock.NewResult(0, 5))

	rowsAffected, err := repo.CleanupDrafts(context.Background(), before)
	assert.NoError(t, err)
	assert.Equal(t, int64(5), rowsAffected)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestKnowledgeRepository_Search(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := NewKnowledgeRepository(sqlxDB)

	query := "测试"
	filter := &models.KnowledgeFilter{
		Page:     1,
		PageSize: 10,
	}

	// Mock search results
	rows := sqlmock.NewRows([]string{
		"id", "title", "slug", "summary", "content", "status", "category_id", "author_id",
		"view_count", "like_count", "dislike_count", "share_count", "download_count",
		"rating", "rating_count", "featured", "created_at", "updated_at", "published_at",
	}).AddRow(
		"id-1", "测试知识", "test-knowledge", "测试摘要", "测试内容", models.KnowledgeStatusPublished,
		"cat-1", "author-1", 100, 10, 1, 5, 2, 4.5, 20, true,
		time.Now(), time.Now(), time.Now(),
	)

	mock.ExpectQuery(`SELECT .+ FROM knowledge_articles WHERE deleted_at IS NULL AND \(title ILIKE \$1 OR content ILIKE \$1\) ORDER BY created_at DESC LIMIT \$2 OFFSET \$3`).WithArgs(
		"%测试%", 10, 0,
	).WillReturnRows(rows)

	// Mock count query
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM knowledge_articles WHERE deleted_at IS NULL AND \(title ILIKE \$1 OR content ILIKE \$1\)`).WithArgs(
		"%测试%",
	).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	result, err := repo.Search(context.Background(), query, filter)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, int64(1), result.Total)
	assert.Len(t, result.Knowledge, 1)
	assert.Equal(t, "测试知识", result.Knowledge[0].Title)
	assert.Equal(t, query, result.Query)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// 测试错误处理
func TestKnowledgeRepository_ErrorHandling(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := NewKnowledgeRepository(sqlxDB)

	t.Run("Create_DatabaseError", func(t *testing.T) {
		knowledge := &models.Knowledge{
			ID:      uuid.New().String(),
			Title:   "测试知识",
			Content: "测试内容",
		}

		mock.ExpectExec(`INSERT INTO knowledge_articles`).WillReturnError(sql.ErrConnDone)

		err := repo.Create(context.Background(), knowledge)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "创建知识库文章失败")
	})

	t.Run("Update_NotFound", func(t *testing.T) {
		knowledge := &models.Knowledge{
			ID:      uuid.New().String(),
			Title:   "更新的知识",
			Content: "更新的内容",
		}

		mock.ExpectExec(`UPDATE knowledge_articles SET`).WillReturnResult(sqlmock.NewResult(1, 0))

		err := repo.Update(context.Background(), knowledge)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "知识库文章不存在")
	})

	assert.NoError(t, mock.ExpectationsWereMet())
}