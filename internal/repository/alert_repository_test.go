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

func setupAlertRepositoryTest(t *testing.T) (AlertRepository, sqlmock.Sqlmock, func()) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := NewAlertRepository(sqlxDB)

	cleanup := func() {
		db.Close()
	}

	return repo, mock, cleanup
}

func TestAlertRepository_Create(t *testing.T) {
	repo, mock, cleanup := setupAlertRepositoryTest(t)
	defer cleanup()

	alert := &models.Alert{
		ID:           uuid.New().String(),
		RuleID:       stringPtr(uuid.New().String()),
		DataSourceID: uuid.New().String(),
		Name:         "Test Alert",
		Description:  "Test alert description",
		Severity:     models.AlertSeverityCritical,
		Status:       models.AlertStatusFiring,
		Source:       models.AlertSourcePrometheus,
		Labels:       map[string]string{"env": "test"},
		Annotations:  map[string]string{"summary": "Test alert"},
		Value:        float64Ptr(100.0),
		Threshold:    float64Ptr(80.0),
		Expression:   "cpu_usage > 80",
		StartsAt:     time.Now(),
		LastEvalAt:   time.Now(),
		EvalCount:    1,
		Fingerprint:  "test-fingerprint",
	}

	mock.ExpectExec(`INSERT INTO alerts`).WithArgs(
		sqlmock.AnyArg(), // id
		alert.RuleID,
		alert.DataSourceID,
		alert.Name,
		alert.Description,
		alert.Severity,
		alert.Status,
		alert.Source,
		sqlmock.AnyArg(), // labels JSON
		sqlmock.AnyArg(), // annotations JSON
		alert.Value,
		alert.Threshold,
		alert.Expression,
		alert.StartsAt,
		alert.EndsAt,
		alert.LastEvalAt,
		alert.EvalCount,
		alert.Fingerprint,
		alert.GeneratorURL,
		alert.SilenceID,
		alert.AckedBy,
		alert.AckedAt,
		alert.ResolvedBy,
		alert.ResolvedAt,
		sqlmock.AnyArg(), // created_at
		sqlmock.AnyArg(), // updated_at
	).WillReturnResult(sqlmock.NewResult(1, 1))

	err := repo.Create(context.Background(), alert)
	assert.NoError(t, err)
	assert.NotEmpty(t, alert.ID)
	assert.False(t, alert.CreatedAt.IsZero())
	assert.False(t, alert.UpdatedAt.IsZero())
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAlertRepository_GetByID(t *testing.T) {
	repo, mock, cleanup := setupAlertRepositoryTest(t)
	defer cleanup()

	alertID := uuid.New().String()
	expectedAlert := &models.Alert{
		ID:          alertID,
		Name:        "Test Alert",
		Description: "Test description",
		Severity:    models.AlertSeverityCritical,
		Status:      models.AlertStatusFiring,
		Source:      models.AlertSourcePrometheus,
	}

	rows := sqlmock.NewRows([]string{
		"id", "rule_id", "data_source_id", "name", "description", "severity", "status", "source",
		"labels", "annotations", "value", "threshold", "expression", "starts_at", "ends_at",
		"last_eval_at", "eval_count", "fingerprint", "generator_url",
		"silence_id", "acked_by", "acked_at", "resolved_by", "resolved_at",
		"created_at", "updated_at", "deleted_at",
	}).AddRow(
		expectedAlert.ID, (*string)(nil), "datasource-1", expectedAlert.Name, expectedAlert.Description,
		expectedAlert.Severity, expectedAlert.Status, expectedAlert.Source,
		"{}", "{}", (*float64)(nil), (*float64)(nil), "test-expression", time.Now(), (*time.Time)(nil),
		time.Now(), int64(1), "test-fingerprint", (*string)(nil),
		(*string)(nil), (*string)(nil), (*time.Time)(nil), (*string)(nil), (*time.Time)(nil),
		time.Now(), time.Now(), (*time.Time)(nil),
	)

	mock.ExpectQuery(`SELECT .+ FROM alerts WHERE id = \$1 AND deleted_at IS NULL`).WithArgs(alertID).WillReturnRows(rows)

	alert, err := repo.GetByID(context.Background(), alertID)
	assert.NoError(t, err)
	assert.Equal(t, expectedAlert.ID, alert.ID)
	assert.Equal(t, expectedAlert.Name, alert.Name)
	assert.Equal(t, expectedAlert.Severity, alert.Severity)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAlertRepository_Update(t *testing.T) {
	repo, mock, cleanup := setupAlertRepositoryTest(t)
	defer cleanup()

	alert := &models.Alert{
		ID:          uuid.New().String(),
		Name:        "Updated Alert",
		Description: "Updated description",
		Severity:     models.AlertSeverityMedium,
		Status:      models.AlertStatusResolved,
		Labels:      map[string]string{"env": "prod"},
		Annotations: map[string]string{"summary": "Updated alert"},
	}

	mock.ExpectExec(`UPDATE alerts SET`).WithArgs(
		alert.RuleID,
		alert.DataSourceID,
		alert.Name,
		alert.Description,
		alert.Severity,
		alert.Status,
		alert.Source,
		sqlmock.AnyArg(), // labels JSON
		sqlmock.AnyArg(), // annotations JSON
		alert.Value,
		alert.Threshold,
		alert.Expression,
		alert.StartsAt,
		alert.EndsAt,
		alert.LastEvalAt,
		alert.EvalCount,
		alert.Fingerprint,
		alert.GeneratorURL,
		alert.SilenceID,
		alert.AckedBy,
		alert.AckedAt,
		alert.ResolvedBy,
		alert.ResolvedAt,
		sqlmock.AnyArg(), // updated_at
		alert.ID,
	).WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Update(context.Background(), alert)
	assert.NoError(t, err)
	assert.False(t, alert.UpdatedAt.IsZero())
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAlertRepository_Delete(t *testing.T) {
	repo, mock, cleanup := setupAlertRepositoryTest(t)
	defer cleanup()

	alertID := uuid.New().String()

	mock.ExpectExec(`DELETE FROM alerts WHERE id = \$1`).WithArgs(alertID).WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Delete(context.Background(), alertID)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAlertRepository_SoftDelete(t *testing.T) {
	repo, mock, cleanup := setupAlertRepositoryTest(t)
	defer cleanup()

	alertID := uuid.New().String()

	mock.ExpectExec(`UPDATE alerts SET deleted_at = \$1, updated_at = \$1 WHERE id = \$2 AND deleted_at IS NULL`).WithArgs(
		sqlmock.AnyArg(), alertID,
	).WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.SoftDelete(context.Background(), alertID)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAlertRepository_GetActiveCount(t *testing.T) {
	repo, mock, cleanup := setupAlertRepositoryTest(t)
	defer cleanup()

	expectedCount := int64(5)
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM alerts WHERE status IN \('firing', 'pending'\) AND deleted_at IS NULL`).WillReturnRows(
		sqlmock.NewRows([]string{"count"}).AddRow(expectedCount),
	)

	count, err := repo.GetActiveCount(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, expectedCount, count)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAlertRepository_GetCriticalCount(t *testing.T) {
	repo, mock, cleanup := setupAlertRepositoryTest(t)
	defer cleanup()

	expectedCount := int64(2)
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM alerts WHERE severity = 'critical' AND status IN \('firing', 'pending'\) AND deleted_at IS NULL`).WillReturnRows(
		sqlmock.NewRows([]string{"count"}).AddRow(expectedCount),
	)

	count, err := repo.GetCriticalCount(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, expectedCount, count)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAlertRepository_CleanupResolved(t *testing.T) {
	repo, mock, cleanup := setupAlertRepositoryTest(t)
	defer cleanup()

	before := time.Now().Add(-24 * time.Hour)
	expectedRows := int64(3)

	mock.ExpectExec(`DELETE FROM alerts WHERE status = \$1 AND resolved_at < \$2`).WithArgs(
		models.AlertStatusResolved, before,
	).WillReturnResult(sqlmock.NewResult(0, expectedRows))

	rowsAffected, err := repo.CleanupResolved(context.Background(), before)
	assert.NoError(t, err)
	assert.Equal(t, expectedRows, rowsAffected)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAlertRepository_CleanupExpired(t *testing.T) {
	repo, mock, cleanup := setupAlertRepositoryTest(t)
	defer cleanup()

	expectedRows := int64(1)

	mock.ExpectExec(`DELETE FROM alerts WHERE ends_at IS NOT NULL AND ends_at < NOW\(\) - INTERVAL '7 days'`).WillReturnResult(
		sqlmock.NewResult(0, expectedRows),
	)

	rowsAffected, err := repo.CleanupExpired(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, expectedRows, rowsAffected)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// Helper functions are now in test_helpers.go