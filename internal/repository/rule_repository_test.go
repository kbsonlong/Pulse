package repository

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"Pulse/internal/models"
)

func setupRuleRepositoryTest(t *testing.T) (*ruleRepository, sqlmock.Sqlmock, func()) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := NewRuleRepository(sqlxDB)

	cleanup := func() {
		db.Close()
	}

	return repo.(*ruleRepository), mock, cleanup
}

// Helper functions are now in test_helpers.go

func TestRuleRepository_Create(t *testing.T) {
	repo, mock, cleanup := setupRuleRepositoryTest(t)
	defer cleanup()

	rule := &models.Rule{
		Name:        "Test Rule",
		Description: "Test rule description",
		Type:        models.RuleTypeMetric,
		Severity:    models.AlertSeverityCritical,
		Status:      models.RuleStatusActive,
		Expression:  "cpu_usage > 80",
		EvaluationInterval: 5 * time.Minute,
		Labels:      map[string]string{"team": "ops"},
		Annotations: map[string]string{"summary": "High CPU usage"},
		CreatedBy:   "user-1",
	}

	// Mock INSERT query
	mock.ExpectExec(`INSERT INTO rules`).WithArgs(
		sqlmock.AnyArg(), rule.Name, rule.Description, rule.Type,
		rule.Severity, rule.Status, rule.Expression, rule.EvaluationInterval,
		sqlmock.AnyArg(), sqlmock.AnyArg(), // labels, annotations JSON
		rule.CreatedBy, sqlmock.AnyArg(), sqlmock.AnyArg(), // created_at, updated_at
	).WillReturnResult(sqlmock.NewResult(1, 1))

	err := repo.Create(context.Background(), rule)
	assert.NoError(t, err)
	assert.NotEmpty(t, rule.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRuleRepository_GetByID(t *testing.T) {
	repo, mock, cleanup := setupRuleRepositoryTest(t)
	defer cleanup()

	ruleID := uuid.New().String()

	rows := sqlmock.NewRows([]string{
		"id", "name", "description", "type", "severity", "status", "enabled", "expression",
		"conditions", "actions", "labels", "annotations", "data_source_id",
		"evaluation_interval", "for_duration", "keep_firing_for", "threshold",
		"recovery_threshold", "no_data_state", "exec_err_state",
		"last_eval_at", "last_eval_result", "eval_count", "alert_count",
		"created_by", "updated_by", "created_at", "updated_at",
	}).AddRow(
		ruleID, "Test Rule", "Test description", models.RuleTypeMetric,
		models.AlertSeverityCritical, models.RuleStatusActive, true, "cpu_usage > 80",
		`[]`, `[]`, `{"team":"ops"}`, `{"summary":"High CPU"}`, "datasource-1",
		5*time.Minute, time.Duration(0), time.Duration(0), nil,
		nil, nil, nil,
		nil, nil, int64(0), int64(0),
		"user-1", nil, time.Now(), time.Now(),
	)

	mock.ExpectQuery(`SELECT .+ FROM rules WHERE id = \$1 AND deleted_at IS NULL`).WithArgs(ruleID).WillReturnRows(rows)

	rule, err := repo.GetByID(context.Background(), ruleID)
	assert.NoError(t, err)
	assert.NotNil(t, rule)
	assert.Equal(t, ruleID, rule.ID)
	assert.Equal(t, "Test Rule", rule.Name)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRuleRepository_GetByID_NotFound(t *testing.T) {
	repo, mock, cleanup := setupRuleRepositoryTest(t)
	defer cleanup()

	ruleID := uuid.New().String()

	mock.ExpectQuery(`SELECT .+ FROM rules WHERE id = \$1 AND deleted_at IS NULL`).WithArgs(ruleID).WillReturnError(sql.ErrNoRows)

	rule, err := repo.GetByID(context.Background(), ruleID)
	assert.Error(t, err)
	assert.Nil(t, rule)
	assert.Contains(t, err.Error(), "规则不存在")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRuleRepository_Update(t *testing.T) {
	repo, mock, cleanup := setupRuleRepositoryTest(t)
	defer cleanup()

	rule := &models.Rule{
		ID:          uuid.New().String(),
		Name:        "Updated Rule",
		Description: "Updated description",
		Type:        models.RuleTypeMetric,
		Severity:    models.AlertSeverityMedium,
		Status:      models.RuleStatusActive,
		Expression:  "memory_usage > 90",
		EvaluationInterval: 10 * time.Minute,
		Labels:      map[string]string{"team": "dev"},
		Annotations: map[string]string{"summary": "High memory usage"},
	}

	mock.ExpectExec(`UPDATE rules SET`).WithArgs(
		rule.Name, rule.Description, rule.Type, rule.Severity,
		rule.Status, rule.Expression, rule.EvaluationInterval,
		sqlmock.AnyArg(), sqlmock.AnyArg(), // labels, annotations JSON
		sqlmock.AnyArg(), rule.ID, // updated_at, id
	).WillReturnResult(sqlmock.NewResult(1, 1))

	err := repo.Update(context.Background(), rule)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRuleRepository_Delete(t *testing.T) {
	repo, mock, cleanup := setupRuleRepositoryTest(t)
	defer cleanup()

	ruleID := uuid.New().String()

	mock.ExpectExec(`DELETE FROM rules WHERE id = \$1`).WithArgs(ruleID).WillReturnResult(sqlmock.NewResult(1, 1))

	err := repo.Delete(context.Background(), ruleID)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRuleRepository_SoftDelete(t *testing.T) {
	repo, mock, cleanup := setupRuleRepositoryTest(t)
	defer cleanup()

	ruleID := uuid.New().String()

	mock.ExpectExec(`UPDATE rules SET deleted_at = \$1, updated_at = \$1 WHERE id = \$2 AND deleted_at IS NULL`).WithArgs(
		sqlmock.AnyArg(), ruleID,
	).WillReturnResult(sqlmock.NewResult(1, 1))

	err := repo.SoftDelete(context.Background(), ruleID)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRuleRepository_List(t *testing.T) {
	repo, mock, cleanup := setupRuleRepositoryTest(t)
	defer cleanup()

	filter := &models.RuleFilter{
		Page:     1,
		PageSize: 10,
	}

	// Mock count query
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM rules WHERE deleted_at IS NULL`).WillReturnRows(
		sqlmock.NewRows([]string{"count"}).AddRow(2),
	)

	// Mock list query
	rows := sqlmock.NewRows([]string{
		"id", "name", "description", "type", "severity", "status", "enabled", "expression",
		"conditions", "actions", "labels", "annotations", "data_source_id",
		"evaluation_interval", "for_duration", "keep_firing_for", "threshold",
		"recovery_threshold", "no_data_state", "exec_err_state",
		"last_eval_at", "last_eval_result", "eval_count", "alert_count",
		"created_by", "updated_by", "created_at", "updated_at",
	}).AddRow(
		"rule-1", "Rule 1", "Description 1", models.RuleTypeMetric,
		models.AlertSeverityCritical, models.RuleStatusActive, true, "cpu_usage > 80",
		`[]`, `[]`, `{}`, `{}`, "datasource-1",
		5*time.Minute, time.Duration(0), time.Duration(0), nil,
		nil, nil, nil,
		nil, nil, int64(0), int64(0),
		"user-1", nil, time.Now(), time.Now(),
	).AddRow(
		"rule-2", "Rule 2", "Description 2", models.RuleTypeLog,
		models.AlertSeverityMedium, models.RuleStatusActive, true, "error_rate > 0.1",
		`[]`, `[]`, `{}`, `{}`, "datasource-2",
		10*time.Minute, time.Duration(0), time.Duration(0), nil,
		nil, nil, nil,
		nil, nil, int64(0), int64(0),
		"user-2", nil, time.Now(), time.Now(),
	)

	mock.ExpectQuery(`SELECT .+ FROM rules WHERE deleted_at IS NULL ORDER BY created_at DESC LIMIT \$1 OFFSET \$2`).WithArgs(
		10, 0,
	).WillReturnRows(rows)

	result, err := repo.List(context.Background(), filter)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, int64(2), result.Total)
	assert.Len(t, result.Rules, 2)
	assert.Equal(t, "Rule 1", result.Rules[0].Name)
	assert.Equal(t, "Rule 2", result.Rules[1].Name)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRuleRepository_Count(t *testing.T) {
	repo, mock, cleanup := setupRuleRepositoryTest(t)
	defer cleanup()

	filter := &models.RuleFilter{}

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM rules WHERE deleted_at IS NULL`).WillReturnRows(
		sqlmock.NewRows([]string{"count"}).AddRow(5),
	)

	count, err := repo.Count(context.Background(), filter)
	assert.NoError(t, err)
	assert.Equal(t, int64(5), count)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRuleRepository_Exists(t *testing.T) {
	repo, mock, cleanup := setupRuleRepositoryTest(t)
	defer cleanup()

	ruleID := uuid.New().String()

	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM rules WHERE id = \$1 AND deleted_at IS NULL\)`).WithArgs(ruleID).WillReturnRows(
		sqlmock.NewRows([]string{"exists"}).AddRow(true),
	)

	exists, err := repo.Exists(context.Background(), ruleID)
	assert.NoError(t, err)
	assert.True(t, exists)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRuleRepository_Activate(t *testing.T) {
	repo, mock, cleanup := setupRuleRepositoryTest(t)
	defer cleanup()

	ruleID := uuid.New().String()

	mock.ExpectExec(`UPDATE rules SET status = \$1, updated_at = \$2 WHERE id = \$3 AND deleted_at IS NULL`).WithArgs(
		models.RuleStatusActive, sqlmock.AnyArg(), ruleID,
	).WillReturnResult(sqlmock.NewResult(1, 1))

	err := repo.Activate(context.Background(), ruleID)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRuleRepository_GetByName(t *testing.T) {
	repo, mock, cleanup := setupRuleRepositoryTest(t)
	defer cleanup()

	ruleName := "Test Rule"

	rows := sqlmock.NewRows([]string{
		"id", "name", "description", "type", "severity", "status", "enabled", "expression",
		"conditions", "actions", "labels", "annotations", "data_source_id",
		"evaluation_interval", "for_duration", "keep_firing_for", "threshold",
		"recovery_threshold", "no_data_state", "exec_err_state",
		"last_eval_at", "last_eval_result", "eval_count", "alert_count",
		"created_by", "updated_by", "created_at", "updated_at",
	}).AddRow(
		"rule-1", ruleName, "Test description", models.RuleTypeMetric,
		models.AlertSeverityCritical, models.RuleStatusActive, true, "cpu_usage > 80",
		`[]`, `[]`, `{}`, `{}`, "datasource-1",
		5*time.Minute, time.Duration(0), time.Duration(0), nil,
		nil, nil, nil,
		nil, nil, int64(0), int64(0),
		"user-1", nil, time.Now(), time.Now(),
	)

	mock.ExpectQuery(`SELECT .+ FROM rules WHERE name = \$1 AND deleted_at IS NULL`).WithArgs(ruleName).WillReturnRows(rows)

	rule, err := repo.GetByName(context.Background(), ruleName)
	assert.NoError(t, err)
	assert.NotNil(t, rule)
	assert.Equal(t, ruleName, rule.Name)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRuleRepository_GetActiveRules(t *testing.T) {
	repo, mock, cleanup := setupRuleRepositoryTest(t)
	defer cleanup()

	rows := sqlmock.NewRows([]string{
		"id", "name", "description", "type", "severity", "status", "enabled", "expression",
		"conditions", "actions", "labels", "annotations", "data_source_id",
		"evaluation_interval", "for_duration", "keep_firing_for", "threshold",
		"recovery_threshold", "no_data_state", "exec_err_state",
		"last_eval_at", "last_eval_result", "eval_count", "alert_count",
		"created_by", "updated_by", "created_at", "updated_at",
	}).AddRow(
		"rule-1", "Active Rule 1", "Description 1", models.RuleTypeMetric,
		models.AlertSeverityCritical, models.RuleStatusActive, true, "cpu_usage > 80",
		`[]`, `[]`, `{}`, `{}`, "datasource-1",
		5*time.Minute, time.Duration(0), time.Duration(0), nil,
		nil, nil, nil,
		nil, nil, int64(0), int64(0),
		"user-1", nil, time.Now(), time.Now(),
	).AddRow(
		"rule-2", "Active Rule 2", "Description 2", models.RuleTypeLog,
		models.AlertSeverityMedium, models.RuleStatusActive, true, "error_rate > 0.1",
		`[]`, `[]`, `{}`, `{}`, "datasource-2",
		10*time.Minute, time.Duration(0), time.Duration(0), nil,
		nil, nil, nil,
		nil, nil, int64(0), int64(0),
		"user-2", nil, time.Now(), time.Now(),
	)

	mock.ExpectQuery(`SELECT .+ FROM rules WHERE enabled = true AND status = \$1 AND deleted_at IS NULL ORDER BY created_at DESC`).WithArgs(
		models.RuleStatusActive,
	).WillReturnRows(rows)

	rules, err := repo.GetActiveRules(context.Background())
	assert.NoError(t, err)
	assert.Len(t, rules, 2)
	assert.Equal(t, "Active Rule 1", rules[0].Name)
	assert.Equal(t, "Active Rule 2", rules[1].Name)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRuleRepository_BatchDeactivate(t *testing.T) {
	repo, mock, cleanup := setupRuleRepositoryTest(t)
	defer cleanup()

	ruleIDs := []string{"rule-1", "rule-2", "rule-3"}

	mock.ExpectBegin()
	for _, id := range ruleIDs {
		mock.ExpectExec(`UPDATE rules SET status = \$1, updated_at = \$2 WHERE id = \$3 AND deleted_at IS NULL`).WithArgs(
			models.RuleStatusInactive, sqlmock.AnyArg(), id,
		).WillReturnResult(sqlmock.NewResult(1, 1))
	}
	mock.ExpectCommit()

	err := repo.BatchDeactivate(context.Background(), ruleIDs)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRuleRepository_BatchDelete(t *testing.T) {
	repo, mock, cleanup := setupRuleRepositoryTest(t)
	defer cleanup()

	ruleIDs := []string{"rule-1", "rule-2", "rule-3"}

	mock.ExpectBegin()
	for _, id := range ruleIDs {
		mock.ExpectExec(`UPDATE rules SET deleted_at = \$1, updated_at = \$1 WHERE id = \$2 AND deleted_at IS NULL`).WithArgs(
			sqlmock.AnyArg(), id,
		).WillReturnResult(sqlmock.NewResult(1, 1))
	}
	mock.ExpectCommit()

	err := repo.BatchDelete(context.Background(), ruleIDs)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRuleRepository_GetStats(t *testing.T) {
	repo, mock, cleanup := setupRuleRepositoryTest(t)
	defer cleanup()

	filter := &models.RuleFilter{}

	// Mock stats query
	mock.ExpectQuery(`SELECT COUNT\(\*\) as total, COUNT\(CASE WHEN status = 'active' THEN 1 END\) as active, COUNT\(CASE WHEN status = 'inactive' THEN 1 END\) as inactive, COUNT\(CASE WHEN status = 'disabled' THEN 1 END\) as disabled FROM rules WHERE deleted_at IS NULL`).WillReturnRows(
		sqlmock.NewRows([]string{"total", "active", "inactive", "disabled"}).AddRow(10, 8, 2, 0),
	)



	stats, err := repo.GetStats(context.Background(), filter)
	assert.NoError(t, err)
	assert.NotNil(t, stats)
	assert.Equal(t, int64(10), stats.Total)
	assert.Equal(t, int64(8), stats.ActiveRules)
	assert.Equal(t, int64(0), stats.Disabled)
	assert.NotNil(t, stats.ByType)
	assert.NotNil(t, stats.BySeverity)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// 测试错误处理
func TestRuleRepository_ErrorHandling(t *testing.T) {
	repo, mock, cleanup := setupRuleRepositoryTest(t)
	defer cleanup()

	t.Run("Create_DatabaseError", func(t *testing.T) {
		rule := &models.Rule{
			Name:       "Test Rule",
			Expression: "cpu_usage > 80",
		}

		mock.ExpectExec(`INSERT INTO rules`).WillReturnError(sql.ErrConnDone)

		err := repo.Create(context.Background(), rule)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "创建规则失败")
	})

	t.Run("Update_NotFound", func(t *testing.T) {
		rule := &models.Rule{
			ID:         uuid.New().String(),
			Name:       "Updated Rule",
			Expression: "memory_usage > 90",
		}

		mock.ExpectExec(`UPDATE rules SET`).WillReturnResult(sqlmock.NewResult(1, 0))

		err := repo.Update(context.Background(), rule)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "规则不存在")
	})

	assert.NoError(t, mock.ExpectationsWereMet())
}