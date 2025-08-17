package repository

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"pulse/internal/models"
)

func setupWebhookRepositoryTest(t *testing.T) (*webhookRepository, sqlmock.Sqlmock, func()) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := NewWebhookRepository(sqlxDB)

	cleanup := func() {
		db.Close()
	}

	return repo.(*webhookRepository), mock, cleanup
}

func TestWebhookRepository_Create(t *testing.T) {
	repo, mock, cleanup := setupWebhookRepositoryTest(t)
	defer cleanup()

	webhook := &models.Webhook{
		ID:           uuid.New(),
		Name:         "Test Webhook",
		URL:          "https://example.com/webhook",
		Timeout:      30,
		RetryCount:   3,
		Status:       models.WebhookStatusActive,
		CreatedBy:    uuid.New(),
	}

	mock.ExpectExec("INSERT INTO webhooks").
		WithArgs(sqlmock.AnyArg(), webhook.Name, webhook.URL, webhook.Secret,
			sqlmock.AnyArg(), sqlmock.AnyArg(), webhook.Timeout, webhook.RetryCount,
			webhook.Status, webhook.CreatedBy, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := repo.Create(context.Background(), webhook)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestWebhookRepository_GetByID(t *testing.T) {
	repo, mock, cleanup := setupWebhookRepositoryTest(t)
	defer cleanup()

	webhookID := uuid.New()
	expectedWebhook := &models.Webhook{
		ID:          webhookID,
		Name:        "Test Webhook",
		URL:         "https://example.com/webhook",
		Status:      models.WebhookStatusActive,
		Timeout:     30,
		RetryCount:  3,
		CreatedBy:   uuid.New(),
	}

	mock.ExpectQuery("SELECT (.+) FROM webhooks WHERE id = \\$1").
		WithArgs(webhookID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "url", "secret", "events", "headers", "timeout", "retry_count", "status", "last_triggered", "created_by", "created_at", "updated_at"}).
			AddRow(expectedWebhook.ID, expectedWebhook.Name, expectedWebhook.URL,
				expectedWebhook.Secret, `[]`, `{}`, expectedWebhook.Timeout, expectedWebhook.RetryCount,
				expectedWebhook.Status, nil, expectedWebhook.CreatedBy, time.Now(), time.Now()))

	webhook, err := repo.GetByID(context.Background(), webhookID.String())
	assert.NoError(t, err)
	assert.Equal(t, expectedWebhook.ID, webhook.ID)
	assert.Equal(t, expectedWebhook.Name, webhook.Name)
	assert.Equal(t, expectedWebhook.URL, webhook.URL)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestWebhookRepository_Update(t *testing.T) {
	repo, mock, cleanup := setupWebhookRepositoryTest(t)
	defer cleanup()

	webhook := &models.Webhook{
		ID:         uuid.New(),
		Name:       "Updated Webhook",
		URL:        "https://updated.example.com/webhook",
		Secret:     stringPtr("newsecret"),
		Timeout:    60,
		RetryCount: 5,
		Status:     models.WebhookStatusInactive,
	}

	mock.ExpectExec(`UPDATE webhooks SET`).WithArgs(
		webhook.ID, webhook.Name, webhook.URL, webhook.Secret,
		sqlmock.AnyArg(), sqlmock.AnyArg(), webhook.Timeout, webhook.RetryCount,
		webhook.Status, sqlmock.AnyArg(),
	).WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Update(context.Background(), webhook)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestWebhookRepository_Delete(t *testing.T) {
	repo, mock, cleanup := setupWebhookRepositoryTest(t)
	defer cleanup()

	webhookID := uuid.New().String()

	mock.ExpectExec(`DELETE FROM webhooks WHERE id = \$1`).WithArgs(webhookID).WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Delete(context.Background(), webhookID)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestWebhookRepository_List(t *testing.T) {
	repo, mock, cleanup := setupWebhookRepositoryTest(t)
	defer cleanup()

	// Mock COUNT query - 匹配实际的复杂查询结构
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM \( SELECT id, name, url, secret, events, headers, timeout, retry_count, status, last_triggered, created_by, created_at, updated_at FROM webhooks WHERE deleted_at IS NULL \) as count_query`).WillReturnRows(
		sqlmock.NewRows([]string{"count"}).AddRow(1),
	)

	// Mock SELECT query
	mock.ExpectQuery(`SELECT id, name, url, secret, events, headers, timeout, retry_count, status, last_triggered, created_by, created_at, updated_at FROM webhooks WHERE deleted_at IS NULL ORDER BY created_at DESC`).WillReturnRows(
		sqlmock.NewRows([]string{"id", "name", "url", "secret", "events", "headers", "timeout", "retry_count", "status", "last_triggered", "created_by", "created_at", "updated_at"}).AddRow(
			"550e8400-e29b-41d4-a716-446655440000", "Test Webhook", "https://example.com/webhook", "secret123", `["alert.created"]`, `{"Content-Type":"application/json"}`, 30, 3, "active", nil, uuid.New(), time.Now(), time.Now(),
		),
	)

	filter := &models.WebhookFilter{}
	result, err := repo.List(context.Background(), filter)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.Webhooks, 1)
	assert.Equal(t, int64(1), result.Total)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestWebhookRepository_IncrementSuccessCount(t *testing.T) {
	repo, mock, cleanup := setupWebhookRepositoryTest(t)
	defer cleanup()

	webhookID := uuid.New()

	mock.ExpectExec(`UPDATE webhooks SET success_count = success_count \+ 1, updated_at = \$2 WHERE id = \$1`).WithArgs(
		webhookID, sqlmock.AnyArg(),
	).WillReturnResult(sqlmock.NewResult(1, 1))

	err := repo.IncrementSuccessCount(context.Background(), webhookID.String())
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestWebhookRepository_IncrementFailureCount(t *testing.T) {
	repo, mock, cleanup := setupWebhookRepositoryTest(t)
	defer cleanup()

	webhookID := uuid.New()

	mock.ExpectExec(`UPDATE webhooks SET failure_count = failure_count \+ 1, updated_at = \$2 WHERE id = \$1`).WithArgs(
		webhookID, sqlmock.AnyArg(),
	).WillReturnResult(sqlmock.NewResult(1, 1))

	err := repo.IncrementFailureCount(context.Background(), webhookID.String())
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestWebhookRepository_UpdateLastTriggered(t *testing.T) {
	repo, mock, cleanup := setupWebhookRepositoryTest(t)
	defer cleanup()

	webhookID := uuid.New()

	mock.ExpectExec(`UPDATE webhooks SET last_triggered = \$2, updated_at = \$3 WHERE id = \$1`).WithArgs(
		webhookID, sqlmock.AnyArg(), sqlmock.AnyArg(),
	).WillReturnResult(sqlmock.NewResult(1, 1))

	err := repo.UpdateLastTriggered(context.Background(), webhookID.String())
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestWebhookRepository_CreateLog(t *testing.T) {
	repo, mock, cleanup := setupWebhookRepositoryTest(t)
	defer cleanup()

	response := "success"
	errorMsg := ""
	log := &models.WebhookLog{
		ID:         uuid.New(),
		WebhookID:  uuid.New(),
		Event:      "test.event",
		Payload:    "test payload",
		StatusCode: 200,
		Response:   &response,
		Error:      &errorMsg,
		Duration:   100,
	}

	mock.ExpectExec(`INSERT INTO webhook_logs`).WithArgs(
		log.ID, log.WebhookID, log.Event, log.Payload,
		log.StatusCode, log.Response, log.Error, log.Duration, sqlmock.AnyArg(),
	).WillReturnResult(sqlmock.NewResult(1, 1))

	err := repo.CreateLog(context.Background(), log)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}