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

	"Pulse/internal/models"
)

func setupDataSourceRepositoryTest(t *testing.T) (*dataSourceRepository, sqlmock.Sqlmock, func()) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := NewDataSourceRepository(sqlxDB)

	cleanup := func() {
		db.Close()
	}

	return repo.(*dataSourceRepository), mock, cleanup
}

func TestDataSourceRepository_Create(t *testing.T) {
	repo, mock, cleanup := setupDataSourceRepositoryTest(t)
	defer cleanup()

	ds := &models.DataSource{
		ID:          uuid.New().String(),
		Name:        "Test DataSource",
		Description: "Test datasource description",
		Type:        models.DataSourceTypePrometheus,
		Config: models.DataSourceConfig{
			URL:      "http://localhost:9090",
			Username: stringPtr("admin"),
			Password: stringPtr("password"),
			Timeout: durationPtr(30 * time.Second),
		},
		Tags:              []string{"env:test"},
		Status:            models.DataSourceStatusActive,
		HealthCheckURL:    stringPtr("http://localhost:9090/api/v1/query"),
		HealthStatus:      stringPtr(string(models.DataSourceHealthStatusHealthy)),
		CreatedBy:         uuid.New().String(),
	}

	mock.ExpectExec(`INSERT INTO data_sources`).WithArgs(
		sqlmock.AnyArg(), // id
		ds.Name,
		ds.Description,
		ds.Type,
		sqlmock.AnyArg(), // config JSON
		sqlmock.AnyArg(), // tags JSON
		ds.Status,
		ds.HealthCheckURL,
		ds.LastHealthCheck,
		ds.HealthStatus,
		ds.CreatedBy,
		sqlmock.AnyArg(), // created_at
		sqlmock.AnyArg(), // updated_at
	).WillReturnResult(sqlmock.NewResult(1, 1))

	err := repo.Create(context.Background(), ds)
	assert.NoError(t, err)
	assert.NotEmpty(t, ds.ID)
	assert.False(t, ds.CreatedAt.IsZero())
	assert.False(t, ds.UpdatedAt.IsZero())
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDataSourceRepository_GetByID(t *testing.T) {
	repo, mock, cleanup := setupDataSourceRepositoryTest(t)
	defer cleanup()

	dsID := uuid.New().String()
	expectedDS := &models.DataSource{
		ID:          dsID,
		Name:        "Test DataSource",
		Description: "Test description",
		Type:        models.DataSourceTypePrometheus,
		Status:      models.DataSourceStatusActive,
	}

	rows := sqlmock.NewRows([]string{
		"id", "name", "description", "type", "config", "tags", "status",
		"health_check_url", "last_health_check", "health_status",
		"created_by", "created_at", "updated_at",
	}).AddRow(
		expectedDS.ID, expectedDS.Name, expectedDS.Description, expectedDS.Type,
		"{\"url\":\"http://localhost:9090\"}", "[]", expectedDS.Status,
		(*string)(nil), (*time.Time)(nil), (*string)(nil),
		"test-user", time.Now(), time.Now(),
	)

	mock.ExpectQuery(`SELECT .+ FROM data_sources WHERE id = \$1 AND deleted_at IS NULL`).WithArgs(dsID).WillReturnRows(rows)

	ds, err := repo.GetByID(context.Background(), dsID)
	assert.NoError(t, err)
	assert.Equal(t, expectedDS.ID, ds.ID)
	assert.Equal(t, expectedDS.Name, ds.Name)
	assert.Equal(t, expectedDS.Type, ds.Type)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDataSourceRepository_GetByName(t *testing.T) {
	repo, mock, cleanup := setupDataSourceRepositoryTest(t)
	defer cleanup()

	dsName := "Test DataSource"
	expectedDS := &models.DataSource{
		ID:          uuid.New().String(),
		Name:        dsName,
		Description: "Test description",
		Type:        models.DataSourceTypePrometheus,
		Status:      models.DataSourceStatusActive,
	}

	rows := sqlmock.NewRows([]string{
		"id", "name", "description", "type", "config", "tags", "status",
		"health_check_url", "last_health_check", "health_status",
		"created_by", "created_at", "updated_at",
	}).AddRow(
		expectedDS.ID, expectedDS.Name, expectedDS.Description, expectedDS.Type,
		"{\"url\":\"http://localhost:9090\"}", "[]", expectedDS.Status,
		(*string)(nil), (*time.Time)(nil), (*string)(nil),
		"test-user", time.Now(), time.Now(),
	)

	mock.ExpectQuery(`SELECT .+ FROM data_sources WHERE name = \$1 AND deleted_at IS NULL`).WithArgs(dsName).WillReturnRows(rows)

	ds, err := repo.GetByName(context.Background(), dsName)
	assert.NoError(t, err)
	assert.Equal(t, expectedDS.Name, ds.Name)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDataSourceRepository_Update(t *testing.T) {
	repo, mock, cleanup := setupDataSourceRepositoryTest(t)
	defer cleanup()

	ds := &models.DataSource{
		ID:          uuid.New().String(),
		Name:        "Updated DataSource",
		Description: "Updated description",
		Type:        models.DataSourceTypeInfluxDB,
		Config: models.DataSourceConfig{
			URL:     "http://localhost:8086",
			Timeout: durationPtr(60 * time.Second),
		},
		Tags:           []string{"env:prod"},
		Status:         models.DataSourceStatusActive,
		HealthCheckURL: stringPtr("http://localhost:8086/ping"),
	}

	mock.ExpectExec(`UPDATE data_sources SET`).WithArgs(
		ds.ID,
		ds.Name,
		ds.Description,
		ds.Type,
		sqlmock.AnyArg(), // config JSON
		sqlmock.AnyArg(), // tags JSON
		ds.Status,
		ds.HealthCheckURL,
		sqlmock.AnyArg(), // updated_at
	).WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Update(context.Background(), ds)
	assert.NoError(t, err)
	assert.False(t, ds.UpdatedAt.IsZero())
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDataSourceRepository_Delete(t *testing.T) {
	repo, mock, cleanup := setupDataSourceRepositoryTest(t)
	defer cleanup()

	dsID := uuid.New().String()

	mock.ExpectExec(`DELETE FROM data_sources WHERE id = \$1`).WithArgs(dsID).WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Delete(context.Background(), dsID)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDataSourceRepository_SoftDelete(t *testing.T) {
	repo, mock, cleanup := setupDataSourceRepositoryTest(t)
	defer cleanup()

	dsID := uuid.New().String()

	mock.ExpectExec(`UPDATE data_sources SET deleted_at = \$1, updated_at = \$2 WHERE id = \$3 AND deleted_at IS NULL`).WithArgs(
		sqlmock.AnyArg(), sqlmock.AnyArg(), dsID,
	).WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.SoftDelete(context.Background(), dsID)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDataSourceRepository_Activate(t *testing.T) {
	repo, mock, cleanup := setupDataSourceRepositoryTest(t)
	defer cleanup()

	dsID := uuid.New().String()

	mock.ExpectExec(`UPDATE data_sources SET status = \$1, updated_at = NOW\(\) WHERE id = \$2 AND deleted_at IS NULL`).WithArgs(
		models.DataSourceStatusActive, dsID,
	).WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Activate(context.Background(), dsID)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDataSourceRepository_Deactivate(t *testing.T) {
	repo, mock, cleanup := setupDataSourceRepositoryTest(t)
	defer cleanup()

	dsID := uuid.New().String()

	mock.ExpectExec(`UPDATE data_sources SET status = \$1, updated_at = NOW\(\) WHERE id = \$2 AND deleted_at IS NULL`).WithArgs(
		models.DataSourceStatusInactive, dsID,
	).WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Deactivate(context.Background(), dsID)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDataSourceRepository_GetActiveCount(t *testing.T) {
	repo, mock, cleanup := setupDataSourceRepositoryTest(t)
	defer cleanup()

	expectedCount := int64(3)
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM data_sources WHERE status = 'active' AND deleted_at IS NULL`).WillReturnRows(
		sqlmock.NewRows([]string{"count"}).AddRow(expectedCount),
	)

	count, err := repo.GetActiveCount(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, expectedCount, count)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDataSourceRepository_GetHealthyCount(t *testing.T) {
	repo, mock, cleanup := setupDataSourceRepositoryTest(t)
	defer cleanup()

	expectedCount := int64(2)
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM data_sources WHERE health_status = 'healthy' AND deleted_at IS NULL`).WillReturnRows(
		sqlmock.NewRows([]string{"count"}).AddRow(expectedCount),
	)

	count, err := repo.GetHealthyCount(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, expectedCount, count)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDataSourceRepository_GetUnhealthyCount(t *testing.T) {
	repo, mock, cleanup := setupDataSourceRepositoryTest(t)
	defer cleanup()

	expectedCount := int64(1)
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM data_sources WHERE health_status != 'healthy' AND deleted_at IS NULL`).WillReturnRows(
		sqlmock.NewRows([]string{"count"}).AddRow(expectedCount),
	)

	count, err := repo.GetUnhealthyCount(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, expectedCount, count)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDataSourceRepository_BatchHealthCheck(t *testing.T) {
	repo, mock, cleanup := setupDataSourceRepositoryTest(t)
	defer cleanup()

	ids := []string{uuid.New().String(), uuid.New().String()}

	mock.ExpectExec(`UPDATE data_sources SET health_status = 'checking', last_health_check = NOW\(\), updated_at = NOW\(\) WHERE id IN \(\$1, \$2\) AND deleted_at IS NULL`).WithArgs(
		ids[0], ids[1],
	).WillReturnResult(sqlmock.NewResult(0, 2))

	err := repo.BatchHealthCheck(context.Background(), ids)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDataSourceRepository_GetStats(t *testing.T) {
	repo, mock, cleanup := setupDataSourceRepositoryTest(t)
	defer cleanup()

	filter := &models.DataSourceFilter{}

	// Mock total count query
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM data_sources WHERE deleted_at IS NULL`).WillReturnRows(
		sqlmock.NewRows([]string{"count"}).AddRow(10),
	)

	// Mock active count query
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM data_sources WHERE status = \$1 AND deleted_at IS NULL`).WithArgs(
		models.DataSourceStatusActive,
	).WillReturnRows(
		sqlmock.NewRows([]string{"count"}).AddRow(8),
	)

	// Mock healthy count query
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM data_sources WHERE health_status = \$1 AND deleted_at IS NULL`).WithArgs(
		models.DataSourceHealthStatusHealthy,
	).WillReturnRows(
		sqlmock.NewRows([]string{"count"}).AddRow(7),
	)

	stats, err := repo.GetStats(context.Background(), filter)
	assert.NoError(t, err)
	assert.Equal(t, int64(10), stats.Total)
	assert.Equal(t, int64(7), stats.HealthyCount)
	assert.Equal(t, int64(8), stats.ByStatus[models.DataSourceStatusActive])
	assert.NoError(t, mock.ExpectationsWereMet())
}

// 辅助函数已在test_helpers.go中定义

func TestDataSourceRepository_GetMetrics(t *testing.T) {
	repo, mock, cleanup := setupDataSourceRepositoryTest(t)
	defer cleanup()

	dsID := uuid.New().String()

	_ = mock // 避免未使用变量警告
	// GetMetrics 方法返回模拟数据，不需要数据库查询
	metrics, err := repo.GetMetrics(context.Background(), dsID)
	assert.NoError(t, err)
	assert.NotNil(t, metrics)
	assert.Equal(t, int64(10), metrics.ConnectionCount)
	assert.Equal(t, int64(1000), metrics.QueryCount)
	assert.Equal(t, int64(5), metrics.ErrorCount)
	assert.Equal(t, 150.5, metrics.AvgResponseTime)
	assert.NotNil(t, metrics.LastQueryAt)
}