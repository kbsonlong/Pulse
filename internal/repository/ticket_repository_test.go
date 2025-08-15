package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
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
		ticket.ID, ticket.Number, ticket.Title, ticket.Description, ticket.Status,
		ticket.Priority, sqlmock.AnyArg(), ticket.Type, ticket.Source, // category
		ticket.ReporterID, sqlmock.AnyArg(), // assignee_id
		sqlmock.AnyArg(), sqlmock.AnyArg(), // tags, custom_fields JSON
		sqlmock.AnyArg(), sqlmock.AnyArg(), // due_date, sla_deadline
		sqlmock.AnyArg(), sqlmock.AnyArg(), // created_at, updated_at
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
		"id", "number", "title", "description", "status", "priority", "category", "type", "source",
		"reporter_id", "assignee_id", "tags", "custom_fields", "due_date", "sla_deadline",
		"resolved_at", "closed_at", "created_at", "updated_at",
	}).AddRow(
		ticket.ID, ticket.Number, ticket.Title, ticket.Description, ticket.Status, ticket.Priority,
		nil, ticket.Type, ticket.Source, ticket.ReporterID, nil,
		`["test","incident"]`, `{"key":"value"}`, nil, nil,
		nil, nil, time.Now(), time.Now(),
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
			ticket.ID, ticket.Title, ticket.Description, ticket.Status, ticket.Priority,
			ticket.Category, ticket.Type, ticket.Source, ticket.AssigneeID, string(tagsJSON),
			string(customFieldsJSON), ticket.DueDate, ticket.SLADeadline, ticket.ResolvedAt,
			ticket.ClosedAt, sqlmock.AnyArg(),
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

	mock.ExpectExec(`DELETE FROM tickets WHERE id`).
		WithArgs(ticketID).
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
			"id", "number", "title", "description", "status", "priority", "category", "type", "source",
			"reporter_id", "assignee_id", "tags", "custom_fields", "due_date", "sla_deadline",
			"resolved_at", "closed_at", "created_at", "updated_at",
		}).AddRow(
			"ticket-1", "T-001", "Test Ticket", "Test Description", models.TicketStatusOpen, models.TicketPriorityMedium,
			nil, models.TicketTypeIncident, models.TicketSourceManual, "user-1", "user-2",
			string(tagsJSON), string(customFieldsJSON), nil, nil,
			nil, nil, time.Now(), time.Now(),
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
		WithArgs(status, sqlmock.AnyArg(), ticketID).
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
	closerID := "user-1"

	mock.ExpectExec(`UPDATE tickets SET status`).
		WithArgs(models.TicketStatusClosed, sqlmock.AnyArg(), closerID, ticketID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.Close(context.Background(), ticketID, closerID)
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
	reopenerID := "user-1"

	mock.ExpectExec(`UPDATE tickets SET status`).
		WithArgs(models.TicketStatusOpen, sqlmock.AnyArg(), ticketID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.Reopen(context.Background(), ticketID, reopenerID)
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
		ID:         "comment-1",
		TicketID:   "ticket-1",
		AuthorID:   "user-1",
		Content:    "Test comment",
		IsInternal: false,
	}

	mock.ExpectExec(`INSERT INTO ticket_comments`).
		WithArgs(
			comment.ID, comment.TicketID, comment.AuthorID, comment.Content,
			comment.IsInternal, sqlmock.AnyArg(), sqlmock.AnyArg(),
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Mock the ticket update query
	mock.ExpectExec(`UPDATE tickets SET updated_at`).
		WithArgs(sqlmock.AnyArg(), comment.TicketID).
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
		ID:               "attachment-1",
		TicketID:         "ticket-1",
		Filename:         "screenshot.png",
		OriginalFilename: "original_screenshot.png",
		FileSize:         1024,
		MimeType:         "image/png",
		FilePath:         "/uploads/screenshot.png",
		UploadBy:         "user-1",
	}

	mock.ExpectExec(`INSERT INTO ticket_attachments`).
		WithArgs(
			attachment.ID, attachment.TicketID, attachment.Filename, attachment.OriginalFilename,
			attachment.FilePath, attachment.FileSize, attachment.MimeType,
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
			"id", "ticket_id", "filename", "original_filename", "file_path",
			"file_size", "mime_type", "upload_by", "created_at",
		}).AddRow(
			"attachment-1", ticketID, "screenshot.png", "original_screenshot.png", "/uploads/screenshot.png",
			1024, "image/png", "user-1", time.Now(),
		))

	attachments, err := repo.GetAttachments(context.Background(), ticketID)
	assert.NoError(t, err)
	assert.Len(t, attachments, 1)
	assert.Equal(t, "attachment-1", attachments[0].ID)
	assert.Equal(t, ticketID, attachments[0].TicketID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTicketRepository_BatchUpdateStatus(t *testing.T) {
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
		WithArgs(string(status), pq.Array(ticketIDs)).
		WillReturnResult(sqlmock.NewResult(2, 2))

	err = repo.BatchUpdateStatus(context.Background(), ticketIDs, status)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTicketRepository_BatchAssign(t *testing.T) {
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
		WithArgs(assigneeID, pq.Array(ticketIDs)).
		WillReturnResult(sqlmock.NewResult(2, 2))

	err = repo.BatchAssign(context.Background(), ticketIDs, assigneeID)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTicketRepository_GetStats(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("创建mock数据库失败: %v", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := NewTicketRepository(sqlxDB)

	filter := &models.TicketFilter{}

	// Mock status query
	mock.ExpectQuery(`SELECT status, COUNT\(\*\) FROM tickets`).
		WillReturnRows(sqlmock.NewRows([]string{"status", "count"}).
			AddRow("open", 5).
			AddRow("closed", 3))

	// Mock priority query
	mock.ExpectQuery(`SELECT priority, COUNT\(\*\) FROM tickets`).
		WillReturnRows(sqlmock.NewRows([]string{"priority", "count"}).
			AddRow("high", 2).
			AddRow("medium", 4).
			AddRow("low", 2))

	// Mock unassigned count query
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM tickets WHERE assignee_id IS NULL`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(3))

	// Mock overdue count query
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM tickets WHERE due_date < \$1`).
		WithArgs(sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	// Mock due soon count query
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM tickets WHERE due_date BETWEEN \$1 AND \$2`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

	stats, err := repo.GetStats(context.Background(), filter)
	assert.NoError(t, err)
	assert.NotNil(t, stats)
	assert.Equal(t, int64(8), stats.Total) // 5 + 3
	assert.Equal(t, int64(5), stats.ByStatus["open"])
	assert.Equal(t, int64(3), stats.ByStatus["closed"])
	assert.Equal(t, int64(2), stats.ByPriority["high"])
	assert.Equal(t, int64(3), stats.Unassigned)
	assert.Equal(t, int64(1), stats.Overdue)
	assert.Equal(t, int64(2), stats.DueSoon)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTicketRepository_GetSLA(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("创建mock数据库失败: %v", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := NewTicketRepository(sqlxDB)

	ticketID := "ticket-1"

	mock.ExpectQuery(`SELECT (.+) FROM ticket_slas`).
		WithArgs(ticketID).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "name", "description", "type", "priority", "severity",
			"response_time", "resolution_time", "escalation_rules", "business_hours",
			"holidays", "enabled", "created_by", "updated_by", "created_at", "updated_at",
		}).AddRow(
			"sla-1", "High Priority SLA", "SLA for high priority tickets", "incident",
			models.TicketPriorityHigh, models.TicketSeverityMajor, 30*time.Minute, 4*time.Hour,
			"{}", "{}", "[]", true, "admin", "admin", time.Now(), time.Now(),
		))

	slaStatus, err := repo.GetSLA(context.Background(), ticketID)
	assert.NoError(t, err)
	assert.NotNil(t, slaStatus)
	assert.Equal(t, models.TicketPriorityHigh, *slaStatus.Priority)
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
	responseTime := 30 * time.Minute
	resolutionTime := 120 * time.Minute
	priority := models.TicketPriorityHigh
	sla := &models.TicketSLA{
		ID:             "sla-1",
		Name:           "Test SLA",
		ResponseTime:   &responseTime,
		ResolutionTime: &resolutionTime,
		Priority:       &priority,
		Enabled:        true,
		CreatedBy:      "admin",
	}

	mock.ExpectExec(`UPDATE tickets SET`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), ticketID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.UpdateSLA(context.Background(), ticketID, sla)
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
			"id", "ticket_id", "action", "field", "old_value",
			"new_value", "changes", "user_id", "user_name", "comment", "created_at",
		}).AddRow(
			"history-1", ticketID, "status_change", "status",
			"open", "in_progress", "{}", "user-1", "Test User", "Status changed", time.Now(),
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
		ID:       "history-1",
		TicketID: "ticket-1",
		Action:   "status_change",
		Field:    stringPtr("status"),
		OldValue: stringPtr("open"),
		NewValue: stringPtr("in_progress"),
		UserID:   "user-1",
		UserName: "Test User",
		Comment:  stringPtr("Status changed"),
	}

	mock.ExpectExec(`INSERT INTO ticket_history`).
		WithArgs(
			history.ID, history.TicketID, history.Action, history.Field,
			history.OldValue, history.NewValue, sqlmock.AnyArg(), // changes JSON
			history.UserID, history.UserName, history.Comment, sqlmock.AnyArg(), // created_at
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