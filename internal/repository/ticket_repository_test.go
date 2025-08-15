package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"Pulse/internal/models"
)

func TestTicketRepository_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("创建mock数据库失败: %v", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := NewTicketRepository(sqlxDB)

	ticket := &models.Ticket{
		ID:          "ticket-1",
		Number:      "T-001",
		Title:       "测试工单",
		Description: "这是一个测试工单",
		Type:        models.TicketTypeIncident,
		Status:      models.TicketStatusOpen,
		Priority:    models.TicketPriorityMedium,
		Severity:    models.TicketSeverityMinor,
		Source:      models.TicketSourceManual,
		ReporterID:  "user-1",
		ReporterName: "测试用户",
		Tags:        []string{"test", "incident"},
		Labels:      map[string]string{"env": "test"},
		CustomFields: map[string]interface{}{"key": "value"},
	}

	mock.ExpectExec("INSERT INTO tickets").WithArgs(
		ticket.ID, ticket.Number, ticket.Title, ticket.Description,
		ticket.Type, ticket.Status, ticket.Priority, ticket.Severity,
		ticket.Source, ticket.ReporterID, ticket.ReporterName,
		sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
	).WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.Create(context.Background(), ticket)
	assert.NoError(t, err)
	assert.Equal(t, "ticket-1", ticket.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTicketRepository_GetByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("创建mock数据库失败: %v", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := NewTicketRepository(sqlxDB)

	ticket := &models.Ticket{
		ID:          "ticket-1",
		Number:      "T-001",
		Title:       "测试工单",
		Description: "这是一个测试工单",
		Type:        models.TicketTypeIncident,
		Status:      models.TicketStatusOpen,
		Priority:    models.TicketPriorityMedium,
		Severity:    models.TicketSeverityMinor,
		Source:      models.TicketSourceManual,
		ReporterID:  "user-1",
		ReporterName: "测试用户",
		Tags:        []string{"test", "incident"},
		Labels:      map[string]string{"env": "test"},
		CustomFields: map[string]interface{}{"key": "value"},
	}

	rows := sqlmock.NewRows([]string{
		"id", "number", "title", "description", "type", "status",
		"priority", "severity", "source", "reporter_id", "reporter_name",
		"tags", "labels", "custom_fields", "created_at", "updated_at",
	}).AddRow(
		ticket.ID, ticket.Number, ticket.Title, ticket.Description,
		ticket.Type, ticket.Status, ticket.Priority, ticket.Severity,
		ticket.Source, ticket.ReporterID, ticket.ReporterName,
		`["test","incident"]`, `{"env":"test"}`, `{"key":"value"}`,
		time.Now(), time.Now(),
	)

	mock.ExpectQuery("SELECT .+ FROM tickets WHERE id = \\$1").WithArgs("ticket-1").WillReturnRows(rows)

	result, err := repo.GetByID(context.Background(), "ticket-1")
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "ticket-1", result.ID)
	assert.Equal(t, "T-001", result.Number)
	assert.Equal(t, "测试工单", result.Title)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTicketRepository_Update(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := NewTicketRepository(sqlxDB)

	ticket := &models.Ticket{
		ID:          "ticket-1",
		Title:       "Updated Ticket",
		Description: "Updated Description",
		Status:      models.TicketStatusInProgress,
		Priority:    models.TicketPriorityHigh,
		Type:        models.TicketTypeIncident,
		ReporterID:  "user-1",
		AssigneeID:  stringPtr("user-2"),
		TeamID:      stringPtr("team-1"),
		Tags:        []string{"bug", "critical"},
		CustomFields: map[string]interface{}{"source": "api"},
	}

	tagsJSON, _ := json.Marshal(ticket.Tags)
	customFieldsJSON, _ := json.Marshal(ticket.CustomFields)

	mock.ExpectExec(`UPDATE tickets SET`).
		WithArgs(
			ticket.Title, ticket.Description, ticket.Status, ticket.Priority,
			ticket.Type, ticket.AssigneeID, ticket.TeamID, string(tagsJSON),
			string(customFieldsJSON), sqlmock.AnyArg(), ticket.ID,
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.Update(context.Background(), ticket)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTicketRepository_Delete(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := NewTicketRepository(sqlxDB)

	ticketID := "ticket-1"

	mock.ExpectExec(`UPDATE tickets SET deleted_at`).
		WithArgs(sqlmock.AnyArg(), ticketID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.Delete(context.Background(), ticketID)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTicketRepository_List(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("创建mock数据库失败: %v", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := NewTicketRepository(sqlxDB)

	filter := &models.TicketFilter{
		Status:   &[]models.TicketStatus{models.TicketStatusOpen}[0],
		Page:     1,
		PageSize: 10,
	}

	tags := []string{"bug"}
	customFields := map[string]interface{}{"source": "web"}
	tagsJSON, _ := json.Marshal(tags)
	customFieldsJSON, _ := json.Marshal(customFields)

	// Mock count query
	mock.ExpectQuery(`SELECT COUNT`).
		WithArgs(models.TicketStatusOpen).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	// Mock list query
	mock.ExpectQuery(`SELECT (.+) FROM tickets`).
		WithArgs(models.TicketStatusOpen, 10, 0).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "title", "description", "status", "priority", "type", "reporter_id",
			"assignee_id", "team_id", "tags", "custom_fields", "created_at", "updated_at",
			"resolved_at", "closed_at", "due_date", "sla_breach_at",
		}).AddRow(
			"ticket-1", "Test Ticket", "Test Description", models.TicketStatusOpen,
			models.TicketPriorityMedium, models.TicketTypeIncident, "user-1",
			"user-2", "team-1", string(tagsJSON), string(customFieldsJSON),
			time.Now(), time.Now(), nil, nil, nil, nil,
		))

	result, err := repo.List(context.Background(), filter)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, int64(1), result.Total)
	assert.Len(t, result.Tickets, 1)
	assert.Equal(t, "ticket-1", result.Tickets[0].ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTicketRepository_Assign(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := NewTicketRepository(sqlxDB)

	ticketID := "ticket-1"
	assigneeID := "user-2"

	mock.ExpectExec(`UPDATE tickets SET assignee_id`).
		WithArgs(assigneeID, sqlmock.AnyArg(), ticketID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.Assign(context.Background(), ticketID, assigneeID)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTicketRepository_UpdateStatus(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := NewTicketRepository(sqlxDB)

	ticketID := "ticket-1"
	status := models.TicketStatusResolved

	mock.ExpectExec(`UPDATE tickets SET status`).
		WithArgs(status, sqlmock.AnyArg(), sqlmock.AnyArg(), ticketID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.UpdateStatus(context.Background(), ticketID, status)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTicketRepository_Close(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := NewTicketRepository(sqlxDB)

	ticketID := "ticket-1"
	resolution := "Fixed the issue"

	mock.ExpectExec(`UPDATE tickets SET status`).
		WithArgs(models.TicketStatusClosed, resolution, sqlmock.AnyArg(), sqlmock.AnyArg(), ticketID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.Close(context.Background(), ticketID, resolution)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTicketRepository_Reopen(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := NewTicketRepository(sqlxDB)

	ticketID := "ticket-1"
	reason := "Issue not fully resolved"

	mock.ExpectExec(`UPDATE tickets SET status`).
		WithArgs(models.TicketStatusOpen, reason, sqlmock.AnyArg(), ticketID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.Reopen(context.Background(), ticketID, reason)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTicketRepository_AddComment(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("创建mock数据库失败: %v", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := NewTicketRepository(sqlxDB)

	comment := &models.TicketComment{
		ID:       "comment-1",
		TicketID: "ticket-1",
		UserID:   "user-1",
		Content:  "Test comment",
		IsPrivate: false,
	}

	mock.ExpectExec(`INSERT INTO ticket_comments`).
		WithArgs(
			comment.ID, comment.TicketID, comment.UserID, comment.Content,
			comment.IsPrivate, sqlmock.AnyArg(),
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.AddComment(context.Background(), comment)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTicketRepository_GetComments(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("创建mock数据库失败: %v", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := NewTicketRepository(sqlxDB)

	ticketID := "ticket-1"

	mock.ExpectQuery(`SELECT (.+) FROM ticket_comments`).
		WithArgs(ticketID).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "ticket_id", "user_id", "content", "is_private", "created_at", "updated_at",
		}).AddRow(
			"comment-1", ticketID, "user-1", "This is a comment",
			false, time.Now(), time.Now(),
		))

	comments, err := repo.GetComments(context.Background(), ticketID)
	assert.NoError(t, err)
	assert.Len(t, comments, 1)
	assert.Equal(t, "comment-1", comments[0].ID)
	assert.Equal(t, ticketID, comments[0].TicketID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTicketRepository_AddAttachment(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("创建mock数据库失败: %v", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := NewTicketRepository(sqlxDB)

	attachment := &models.TicketAttachment{
		ID:       "attachment-1",
		TicketID: "ticket-1",
		Filename: "screenshot.png",
		FileSize: 1024,
		MimeType: "image/png",
		FilePath: "/uploads/screenshot.png",
		UploadBy: "user-1",
	}

	mock.ExpectExec(`INSERT INTO ticket_attachments`).
		WithArgs(
			attachment.ID, attachment.TicketID, attachment.Filename,
			attachment.FileSize, attachment.MimeType, attachment.FilePath,
			attachment.UploadBy, sqlmock.AnyArg(),
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.AddAttachment(context.Background(), attachment)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTicketRepository_GetAttachments(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("创建mock数据库失败: %v", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := NewTicketRepository(sqlxDB)

	ticketID := "ticket-1"

	mock.ExpectQuery(`SELECT (.+) FROM ticket_attachments`).
		WithArgs(ticketID).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "ticket_id", "filename", "file_size", "mime_type",
			"file_path", "upload_by", "created_at",
		}).AddRow(
			"attachment-1", ticketID, "screenshot.png", 1024, "image/png",
			"/uploads/screenshot.png", "user-1", time.Now(),
		))

	attachments, err := repo.GetAttachments(context.Background(), ticketID)
	assert.NoError(t, err)
	assert.Len(t, attachments, 1)
	assert.Equal(t, "attachment-1", attachments[0].ID)
	assert.Equal(t, ticketID, attachments[0].TicketID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTicketRepository_BulkUpdateStatus(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("创建mock数据库失败: %v", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := NewTicketRepository(sqlxDB)

	ticketIDs := []string{"ticket-1", "ticket-2"}
	status := models.TicketStatusClosed

	mock.ExpectExec(`UPDATE tickets SET status`).
		WithArgs(status, sqlmock.AnyArg(), "ticket-1", "ticket-2").
		WillReturnResult(sqlmock.NewResult(2, 2))

	err = repo.BulkUpdateStatus(context.Background(), ticketIDs, status)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTicketRepository_BulkAssign(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("创建mock数据库失败: %v", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := NewTicketRepository(sqlxDB)

	ticketIDs := []string{"ticket-1", "ticket-2"}
	assigneeID := "user-2"

	mock.ExpectExec(`UPDATE tickets SET assignee_id`).
		WithArgs(assigneeID, sqlmock.AnyArg(), "ticket-1", "ticket-2").
		WillReturnResult(sqlmock.NewResult(2, 2))

	err = repo.BulkAssign(context.Background(), ticketIDs, assigneeID)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTicketRepository_GetStatistics(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("创建mock数据库失败: %v", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := NewTicketRepository(sqlxDB)

	filter := &models.TicketFilter{
		TeamID: stringPtr("team-1"),
	}

	mock.ExpectQuery(`SELECT status, COUNT`).
		WithArgs("team-1").
		WillReturnRows(sqlmock.NewRows([]string{"status", "count"}).
			AddRow(models.TicketStatusOpen, 5).
			AddRow(models.TicketStatusInProgress, 3).
			AddRow(models.TicketStatusClosed, 10))

	mock.ExpectQuery(`SELECT priority, COUNT`).
		WithArgs("team-1").
		WillReturnRows(sqlmock.NewRows([]string{"priority", "count"}).
			AddRow(models.TicketPriorityHigh, 2).
			AddRow(models.TicketPriorityMedium, 8).
			AddRow(models.TicketPriorityLow, 8))

	mock.ExpectQuery(`SELECT type, COUNT`).
		WithArgs("team-1").
		WillReturnRows(sqlmock.NewRows([]string{"type", "count"}).
			AddRow(models.TicketTypeIncident, 12).
			AddRow(models.TicketTypeRequest, 6))

	stats, err := repo.GetStatistics(context.Background(), filter)
	assert.NoError(t, err)
	assert.NotNil(t, stats)
	assert.Equal(t, int64(18), stats.Total)
	assert.Equal(t, int64(5), stats.ByStatus[models.TicketStatusOpen])
	assert.Equal(t, int64(2), stats.ByPriority[models.TicketPriorityHigh])
	assert.Equal(t, int64(12), stats.ByType[models.TicketTypeIncident])
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTicketRepository_GetSLAStatus(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("创建mock数据库失败: %v", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := NewTicketRepository(sqlxDB)

	ticketID := "ticket-1"

	mock.ExpectQuery(`SELECT (.+) FROM tickets`).
		WithArgs(ticketID).
		WillReturnRows(sqlmock.NewRows([]string{
			"priority", "created_at", "resolved_at", "sla_breach_at",
		}).AddRow(
			models.TicketPriorityHigh, time.Now().Add(-2*time.Hour), nil, nil,
		))

	slaStatus, err := repo.GetSLAStatus(context.Background(), ticketID)
	assert.NoError(t, err)
	assert.NotNil(t, slaStatus)
	assert.Equal(t, models.TicketPriorityHigh, slaStatus.Priority)
	assert.False(t, slaStatus.IsBreached)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTicketRepository_UpdateSLA(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("创建mock数据库失败: %v", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := NewTicketRepository(sqlxDB)

	ticketID := "ticket-1"
	dueDate := time.Now().Add(24 * time.Hour)
	slaBreachAt := time.Now().Add(48 * time.Hour)

	mock.ExpectExec(`UPDATE tickets SET due_date`).
		WithArgs(dueDate, slaBreachAt, sqlmock.AnyArg(), ticketID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.UpdateSLA(context.Background(), ticketID, dueDate, slaBreachAt)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTicketRepository_GetHistory(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("创建mock数据库失败: %v", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := NewTicketRepository(sqlxDB)

	ticketID := "ticket-1"

	mock.ExpectQuery(`SELECT (.+) FROM ticket_history`).
		WithArgs(ticketID).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "ticket_id", "action", "field_name", "old_value",
			"new_value", "changed_by", "created_at",
		}).AddRow(
			"history-1", ticketID, "status_change", "status",
			"open", "in_progress", "user-1", time.Now(),
		))

	history, err := repo.GetHistory(context.Background(), ticketID)
	assert.NoError(t, err)
	assert.Len(t, history, 1)
	assert.Equal(t, "history-1", history[0].ID)
	assert.Equal(t, ticketID, history[0].TicketID)
	assert.Equal(t, "status_change", history[0].Action)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTicketRepository_AddHistory(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("创建mock数据库失败: %v", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := NewTicketRepository(sqlxDB)

	history := &models.TicketHistory{
		ID:        "history-1",
		TicketID:  "ticket-1",
		Action:    "status_change",
		FieldName: stringPtr("status"),
		OldValue:  stringPtr("open"),
		NewValue:  stringPtr("in_progress"),
		ChangedBy: "user-1",
	}

	mock.ExpectExec(`INSERT INTO ticket_history`).
		WithArgs(
			history.ID, history.TicketID, history.Action, history.FieldName,
			history.OldValue, history.NewValue, history.ChangedBy, sqlmock.AnyArg(),
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.AddHistory(context.Background(), history)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTicketRepository_Exists(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("创建mock数据库失败: %v", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := NewTicketRepository(sqlxDB)

	ticketID := "ticket-1"

	mock.ExpectQuery(`SELECT COUNT`).
		WithArgs(ticketID).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	exists, err := repo.Exists(context.Background(), ticketID)
	assert.NoError(t, err)
	assert.True(t, exists)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTicketRepository_Count(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("创建mock数据库失败: %v", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := NewTicketRepository(sqlxDB)

	filter := &models.TicketFilter{
		Status: &[]models.TicketStatus{models.TicketStatusOpen}[0],
	}

	mock.ExpectQuery(`SELECT COUNT`).
		WithArgs(models.TicketStatusOpen).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))

	count, err := repo.Count(context.Background(), filter)
	assert.NoError(t, err)
	assert.Equal(t, int64(5), count)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// 错误处理测试
func TestTicketRepository_GetByID_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("创建mock数据库失败: %v", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := NewTicketRepository(sqlxDB)

	ticketID := "nonexistent"

	mock.ExpectQuery(`SELECT (.+) FROM tickets WHERE`).
		WithArgs(ticketID).
		WillReturnError(sql.ErrNoRows)

	ticket, err := repo.GetByID(context.Background(), ticketID)
	assert.Error(t, err)
	assert.Nil(t, ticket)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTicketRepository_Create_DatabaseError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("创建mock数据库失败: %v", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := NewTicketRepository(sqlxDB)

	ticket := &models.Ticket{
		ID:          "ticket-1",
		Title:       "Test Ticket",
		Description: "Test Description",
		Status:      models.TicketStatusOpen,
		Priority:    models.TicketPriorityMedium,
		Type:        models.TicketTypeIncident,
		ReporterID:  "user-1",
	}

	mock.ExpectExec(`INSERT INTO tickets`).
		WillReturnError(assert.AnError)

	err = repo.Create(context.Background(), ticket)
	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}