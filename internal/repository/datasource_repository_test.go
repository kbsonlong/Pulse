package repository

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	testifymock "github.com/stretchr/testify/mock"

	"Pulse/internal/models"
)

// MockEncryptionService mock加密服务
type MockEncryptionService struct {
	testifymock.Mock
}

func (m *MockEncryptionService) EncryptDataSourceConfig(config *models.DataSourceConfig) error {
	args := m.Called(config)
	return args.Error(0)
}

func (m *MockEncryptionService) DecryptDataSourceConfig(config *models.DataSourceConfig) error {
	args := m.Called(config)
	return args.Error(0)
}

func (m *MockEncryptionService) IsConfigEncrypted(config *models.DataSourceConfig) bool {
	args := m.Called(config)
	return args.Bool(0)
}

// setupTestDB 设置测试数据库
func setupTestDB(t *testing.T) (*sqlx.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("创建mock数据库失败: %v", err)
	}
	sqlxDB := sqlx.NewDb(db, "sqlmock")
	return sqlxDB, mock
}

func TestDataSourceRepository_TestConnection(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func(mock sqlmock.Sqlmock, encMock *MockEncryptionService)
		dataSource  *models.DataSource
		expectError bool
	}{
		{
			name: "成功测试连接",
			setupMock: func(mock sqlmock.Sqlmock, encMock *MockEncryptionService) {
				// Mock解密服务
				encMock.On("DecryptDataSourceConfig", testifymock.AnythingOfType("*models.DataSourceConfig")).Return(nil)
			},
			dataSource: &models.DataSource{
				ID:   "test-id",
				Type: "http",
				Config: models.DataSourceConfig{
					"url": "http://httpbin.org/status/200",
				},
			},
			expectError: false,
		},
		{
			name: "解密失败",
			setupMock: func(mock sqlmock.Sqlmock, encMock *MockEncryptionService) {
				// Mock解密服务失败
				encMock.On("DecryptDataSourceConfig", testifymock.AnythingOfType("*models.DataSourceConfig")).Return(errors.New("decryption failed"))
			},
			dataSource: &models.DataSource{
				ID:   "test-id",
				Type: "http",
				Config: models.DataSourceConfig{
					"url": "http://httpbin.org/status/200",
				},
			},
			expectError: true,
		},
		{
			name: "不支持的数据源类型",
			setupMock: func(mock sqlmock.Sqlmock, encMock *MockEncryptionService) {
				// Mock解密服务
				encMock.On("DecryptDataSourceConfig", testifymock.AnythingOfType("*models.DataSourceConfig")).Return(nil)
			},
			dataSource: &models.DataSource{
				ID:   "test-id",
				Type: "unsupported",
				Config: models.DataSourceConfig{},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, sqlMock := setupTestDB(t)
			defer db.Close()

			encMock := &MockEncryptionService{}
			tt.setupMock(sqlMock, encMock)

			repo := NewDataSourceRepository(db, encMock)
			ctx := context.Background()

			err := repo.TestConnection(ctx, tt.dataSource)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			encMock.AssertExpectations(t)
			assert.NoError(t, sqlMock.ExpectationsWereMet())
		})
	}
}

func TestDataSourceRepository_BatchHealthCheck(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func(mock sqlmock.Sqlmock, encMock *MockEncryptionService)
		dataSources []*models.DataSource
		expectError bool
	}{
		{
			name: "成功批量健康检查",
			setupMock: func(mock sqlmock.Sqlmock, encMock *MockEncryptionService) {
				// Mock解密服务
				encMock.On("DecryptDataSourceConfig", testifymock.AnythingOfType("*models.DataSourceConfig")).Return(nil)
				
				// Mock批量更新健康状态
				mock.ExpectExec(`UPDATE data_sources SET health_status = \$1, last_health_check = \$2, error_message = \$3, updated_at = \$4 WHERE id = \$5`).WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(0, 1))
			},
			dataSources: []*models.DataSource{
				{
					ID:   "test-id-1",
					Type: "http",
					Config: models.DataSourceConfig{
						"url": "http://httpbin.org/status/200",
					},
				},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, sqlMock := setupTestDB(t)
			defer db.Close()

			encMock := &MockEncryptionService{}
			tt.setupMock(sqlMock, encMock)

			repo := NewDataSourceRepository(db, encMock)
			ctx := context.Background()

			err := repo.BatchHealthCheck(ctx, tt.dataSources)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			encMock.AssertExpectations(t)
			assert.NoError(t, sqlMock.ExpectationsWereMet())
		})
	}
}

func TestDataSourceRepository_GetStats(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func(mock sqlmock.Sqlmock)
		expectError bool
	}{
		{
			name: "成功获取统计信息",
			setupMock: func(mock sqlmock.Sqlmock) {
				// Mock统计查询
				rows := sqlmock.NewRows([]string{"total", "active", "inactive", "healthy", "unhealthy"}).
					AddRow(10, 8, 2, 7, 3)
				mock.ExpectQuery(`SELECT COUNT\(\*\) as total`).WillReturnRows(rows)
			},
			expectError: false,
		},
		{
			name: "数据库查询失败",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT COUNT\(\*\) as total`).WillReturnError(errors.New("database error"))
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, sqlMock := setupTestDB(t)
			defer db.Close()

			repo := NewDataSourceRepository(db, &MockEncryptionService{})
			ctx := context.Background()

			stats, err := repo.GetStats(ctx)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, stats)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, stats)
			}

			assert.NoError(t, sqlMock.ExpectationsWereMet())
		})
	}
}

func TestDataSourceRepository_GetActiveCount(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func(mock sqlmock.Sqlmock)
		expectError bool
		expectedCount int64
	}{
		{
			name: "成功获取活跃数量",
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"count"}).AddRow(5)
				mock.ExpectQuery(`SELECT COUNT\(\*\) FROM data_sources WHERE status = 'active' AND deleted_at IS NULL`).WillReturnRows(rows)
			},
			expectError: false,
			expectedCount: 5,
		},
		{
			name: "数据库查询失败",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT COUNT\(\*\) FROM data_sources WHERE status = 'active' AND deleted_at IS NULL`).WillReturnError(errors.New("database error"))
			},
			expectError: true,
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, sqlMock := setupTestDB(t)
			defer db.Close()

			repo := NewDataSourceRepository(db, &MockEncryptionService{})
			ctx := context.Background()

			count, err := repo.GetActiveCount(ctx)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedCount, count)
			}

			assert.NoError(t, sqlMock.ExpectationsWereMet())
		})
	}
}

func TestDataSourceRepository_UpdateMetrics(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func(mock sqlmock.Sqlmock)
		dataSourceID string
		metrics     map[string]interface{}
		expectError bool
	}{
		{
			name: "成功更新指标",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`UPDATE data_sources SET metrics = \$1, updated_at = \$2 WHERE id = \$3`).WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), "test-id").WillReturnResult(sqlmock.NewResult(0, 1))
			},
			dataSourceID: "test-id",
			metrics: map[string]interface{}{"cpu": 80.5, "memory": 60.2},
			expectError: false,
		},
		{
			name: "数据库更新失败",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`UPDATE data_sources SET metrics = \$1, updated_at = \$2 WHERE id = \$3`).WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), "test-id").WillReturnError(errors.New("database error"))
			},
			dataSourceID: "test-id",
			metrics: map[string]interface{}{"cpu": 80.5},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, sqlMock := setupTestDB(t)
			defer db.Close()

			repo := NewDataSourceRepository(db, &MockEncryptionService{})
			ctx := context.Background()

			err := repo.UpdateMetrics(ctx, tt.dataSourceID, tt.metrics)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.NoError(t, sqlMock.ExpectationsWereMet())
		})
	}
}

func createTestDataSource() *models.DataSource {	return &models.DataSource{
		ID:          "test-id",
		Name:        "Test DataSource",
		Description: "Test Description",
		Type:        models.DataSourceTypePrometheus,
		Status:      models.DataSourceStatusActive,
		Config: models.DataSourceConfig{
			URL:      "http://localhost:9090",
			Username: stringPtr("admin"),
			Password: stringPtr("password"),
		},
		CreatedBy: "test-user",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}



func TestDataSourceRepository_Create(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func(mock sqlmock.Sqlmock, encMock *MockEncryptionService)
		dataSource  *models.DataSource
		expectError bool
	}{
		{
			name: "成功创建数据源",
			setupMock: func(mock sqlmock.Sqlmock, encMock *MockEncryptionService) {
				// Mock加密服务
				encMock.On("EncryptDataSourceConfig", testifymock.AnythingOfType("*models.DataSourceConfig")).Return(nil)
				
				// Mock数据库插入
				mock.ExpectExec(`INSERT INTO data_sources`).
					WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			dataSource:  createTestDataSource(),
			expectError: false,
		},
		{
			name: "加密失败",
			setupMock: func(mock sqlmock.Sqlmock, encMock *MockEncryptionService) {
				// Mock加密服务失败
				encMock.On("EncryptDataSourceConfig", testifymock.AnythingOfType("*models.DataSourceConfig")).Return(errors.New("encryption failed"))
			},
			dataSource:  createTestDataSource(),
			expectError: true,
		},
		{
			name: "数据库插入失败",
			setupMock: func(mock sqlmock.Sqlmock, encMock *MockEncryptionService) {
				// Mock加密服务成功
				encMock.On("EncryptDataSourceConfig", testifymock.AnythingOfType("*models.DataSourceConfig")).Return(nil)
				
				// Mock数据库插入失败
			mock.ExpectExec(`INSERT INTO data_sources`).
				WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
				WillReturnError(errors.New("database error"))
			},
			dataSource:  createTestDataSource(),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, sqlMock := setupTestDB(t)
			defer db.Close()

			encMock := &MockEncryptionService{}
			tt.setupMock(sqlMock, encMock)

			repo := NewDataSourceRepository(db, encMock)
			ctx := context.Background()

			err := repo.Create(ctx, tt.dataSource)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, tt.dataSource.ID)
			}

			encMock.AssertExpectations(t)
			assert.NoError(t, sqlMock.ExpectationsWereMet())
		})
	}
}

func TestDataSourceRepository_GetByID(t *testing.T) {
	tests := []struct {
		name         string
		setupMock    func(mock sqlmock.Sqlmock, encMock *MockEncryptionService)
		dataSourceID string
		expectError  bool
		expectNil    bool
	}{
		{
			name: "成功获取数据源",
			setupMock: func(mock sqlmock.Sqlmock, encMock *MockEncryptionService) {
				// Mock数据库查询 - 需要匹配17个字段
				rows := sqlmock.NewRows([]string{"id", "name", "description", "type", "config", "tags", "version", "health_check_url", "health_status", "last_health_check", "error_message", "metrics", "status", "created_by", "updated_by", "created_at", "updated_at"}).
AddRow("test-id", "Test DataSource", "Test Description", "prometheus", `{"url":"http://localhost:9090"}`, `[]`, "1.0", "http://localhost:9090/health", "healthy", time.Now(), "", `{}`, "active", "test-user", "test-user", time.Now(), time.Now())
				mock.ExpectQuery(`SELECT (.+) FROM data_sources WHERE id = \$1`).
					WithArgs("test-id").
					WillReturnRows(rows)
				
				// Mock解密服务
				encMock.On("DecryptDataSourceConfig", testifymock.AnythingOfType("*models.DataSourceConfig")).Return(nil)
			},
			dataSourceID: "test-id",
			expectError:  false,
			expectNil:    false,
		},
		{
			name: "数据源不存在",
			setupMock: func(mock sqlmock.Sqlmock, encMock *MockEncryptionService) {
				// Mock数据库查询返回空结果
				mock.ExpectQuery(`SELECT (.+) FROM data_sources WHERE id = \$1`).
					WithArgs("nonexistent-id").
					WillReturnError(sql.ErrNoRows)
			},
			dataSourceID: "nonexistent-id",
			expectError:  false,
			expectNil:    true,
		},
		{
			name: "解密失败",
			setupMock: func(mock sqlmock.Sqlmock, encMock *MockEncryptionService) {
				// Mock数据库查询 - 需要匹配17个字段
			rows := sqlmock.NewRows([]string{"id", "name", "description", "type", "config", "tags", "version", "health_check_url", "health_status", "last_health_check", "error_message", "metrics", "status", "created_by", "updated_by", "created_at", "updated_at"}).
				AddRow("test-id", "Test DataSource", "Test Description", "prometheus", `{"url":"http://localhost:9090"}`, `[]`, "1.0", "http://localhost:9090/health", "healthy", time.Now(), "", `{}`, "active", "test-user", "test-user", time.Now(), time.Now())
				mock.ExpectQuery(`SELECT (.+) FROM data_sources WHERE id = \$1`).
					WithArgs("test-id").
					WillReturnRows(rows)
				
				// Mock解密服务失败
				encMock.On("DecryptDataSourceConfig", testifymock.AnythingOfType("*models.DataSourceConfig")).Return(errors.New("decryption failed"))
			},
			dataSourceID: "test-id",
			expectError:  true,
			expectNil:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, sqlMock := setupTestDB(t)
			defer db.Close()

			encMock := &MockEncryptionService{}
			tt.setupMock(sqlMock, encMock)

			repo := NewDataSourceRepository(db, encMock)
			ctx := context.Background()

			result, err := repo.GetByID(ctx, tt.dataSourceID)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			if tt.expectNil {
				assert.Nil(t, result)
			} else if !tt.expectError {
				assert.NotNil(t, result)
				assert.Equal(t, tt.dataSourceID, result.ID)
			}

			encMock.AssertExpectations(t)
			assert.NoError(t, sqlMock.ExpectationsWereMet())
		})
	}
}

func TestDataSourceRepository_Update(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func(mock sqlmock.Sqlmock, encMock *MockEncryptionService)
		dataSource  *models.DataSource
		expectError bool
	}{
		{
			name: "成功更新数据源",
			setupMock: func(mock sqlmock.Sqlmock, encMock *MockEncryptionService) {
				// Mock加密服务
				encMock.On("EncryptDataSourceConfig", testifymock.AnythingOfType("*models.DataSourceConfig")).Return(nil)
				
				// Mock数据库更新 - Update方法有9个参数
				mock.ExpectExec(`UPDATE data_sources SET`).
					WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			dataSource:  createTestDataSource(),
			expectError: false,
		},
		{
			name: "加密失败",
			setupMock: func(mock sqlmock.Sqlmock, encMock *MockEncryptionService) {
				// Mock加密服务失败
				encMock.On("EncryptDataSourceConfig", testifymock.AnythingOfType("*models.DataSourceConfig")).Return(errors.New("encryption failed"))
			},
			dataSource:  createTestDataSource(),
			expectError: true,
		},
		{
			name: "数据库更新失败",
			setupMock: func(mock sqlmock.Sqlmock, encMock *MockEncryptionService) {
				// Mock加密服务成功
				encMock.On("EncryptDataSourceConfig", testifymock.AnythingOfType("*models.DataSourceConfig")).Return(nil)
				
				// Mock数据库更新失败 - Update方法有9个参数
				mock.ExpectExec(`UPDATE data_sources SET`).
					WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnError(errors.New("database error"))
			},
			dataSource:  createTestDataSource(),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, sqlMock := setupTestDB(t)
			defer db.Close()

			encMock := &MockEncryptionService{}
			tt.setupMock(sqlMock, encMock)

			repo := NewDataSourceRepository(db, encMock)
			ctx := context.Background()

			err := repo.Update(ctx, tt.dataSource)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			encMock.AssertExpectations(t)
			assert.NoError(t, sqlMock.ExpectationsWereMet())
		})
	}
}

func TestDataSourceRepository_Delete(t *testing.T) {
	tests := []struct {
		name         string
		setupMock    func(mock sqlmock.Sqlmock)
		dataSourceID string
		expectError  bool
	}{
		{
			name: "成功删除数据源",
			setupMock: func(mock sqlmock.Sqlmock) {
				// Mock数据库删除
				mock.ExpectExec(`DELETE FROM data_sources WHERE id = \$1`).
					WithArgs("test-id").
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			dataSourceID: "test-id",
			expectError:  false,
		},
		{
			name: "数据源不存在",
			setupMock: func(mock sqlmock.Sqlmock) {
				// Mock数据库删除，影响行数为0
				mock.ExpectExec(`DELETE FROM data_sources WHERE id = \$1`).
					WithArgs("nonexistent-id").
					WillReturnResult(sqlmock.NewResult(0, 0))
			},
			dataSourceID: "nonexistent-id",
			expectError:  true,
		},
		{
			name: "数据库错误",
			setupMock: func(mock sqlmock.Sqlmock) {
				// Mock数据库删除失败
				mock.ExpectExec(`DELETE FROM data_sources WHERE id = \$1`).
					WithArgs("test-id").
					WillReturnError(errors.New("database error"))
			},
			dataSourceID: "test-id",
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, sqlMock := setupTestDB(t)
			defer db.Close()

			tt.setupMock(sqlMock)

			repo := NewDataSourceRepository(db, &MockEncryptionService{})
			ctx := context.Background()

			err := repo.Delete(ctx, tt.dataSourceID)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.NoError(t, sqlMock.ExpectationsWereMet())
		})
	}
}

func TestDataSourceRepository_List(t *testing.T) {
	tests := []struct {
		name          string
		setupMock     func(mock sqlmock.Sqlmock, encMock *MockEncryptionService)
		filter        *models.DataSourceFilter
		expectedCount int
		expectError   bool
	}{
		{
			name: "成功获取数据源列表",
			setupMock: func(mock sqlmock.Sqlmock, encMock *MockEncryptionService) {
				// Mock总数查询 - 必须先设置，因为List方法先调用Count
				countRows := sqlmock.NewRows([]string{"count"}).AddRow(2)
				mock.ExpectQuery(`SELECT COUNT\(\*\) FROM data_sources WHERE deleted_at IS NULL`).
					WillReturnRows(countRows)
				
				// Mock数据库查询
				rows := sqlmock.NewRows([]string{
					"id", "name", "description", "type", "config", "tags", "version",
					"health_check_url", "health_status", "last_health_check", "error_message",
					"metrics", "status", "created_by", "updated_by", "created_at", "updated_at",
				}).AddRow(
					"test-id-1", "DataSource 1", "Description 1", "prometheus",
					"{}", "[]", "1.0",
					"http://test1.com/health", "healthy", time.Now(), "",
					"{}", "active", "user1", "user1", time.Now(), time.Now(),
				).AddRow(
					"test-id-2", "DataSource 2", "Description 2", "grafana",
					"{}", "[]", "2.0",
					"http://test2.com/health", "healthy", time.Now(), "",
					"{}", "active", "user2", "user2", time.Now(), time.Now(),
				)
				mock.ExpectQuery(`SELECT (.+) FROM data_sources`).
					WillReturnRows(rows)
				
				// Mock解密服务
				encMock.On("DecryptDataSourceConfig", testifymock.AnythingOfType("*models.DataSourceConfig")).Return(nil).Times(2)
			},
			filter: &models.DataSourceFilter{
				Page:     1,
				PageSize: 10,
			},
			expectedCount: 2,
			expectError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, sqlMock := setupTestDB(t)
			defer db.Close()

			encMock := &MockEncryptionService{}
			tt.setupMock(sqlMock, encMock)

			repo := NewDataSourceRepository(db, encMock)
			ctx := context.Background()

			result, err := repo.List(ctx, tt.filter)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Len(t, result.DataSources, tt.expectedCount)
				assert.Equal(t, int64(tt.expectedCount), result.Total)
			}

			encMock.AssertExpectations(t)
			assert.NoError(t, sqlMock.ExpectationsWereMet())
		})
	}
}