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

	"demo02/internal/models"
)

func setupPermissionRepositoryTest(t *testing.T) (*permissionRepository, sqlmock.Sqlmock, func()) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := NewPermissionRepository(sqlxDB)

	cleanup := func() {
		db.Close()
	}

	return repo.(*permissionRepository), mock, cleanup
}

func TestPermissionRepository_CheckPermission(t *testing.T) {
	repo, mock, cleanup := setupPermissionRepositoryTest(t)
	defer cleanup()

	userID := uuid.New().String()
	permission := models.PermissionUserRead

	// Mock user query
	userRows := sqlmock.NewRows([]string{
		"id", "username", "email", "password_hash", "display_name",
		"role", "status", "phone", "avatar", "department",
		"last_login_at", "created_at", "updated_at",
	}).AddRow(
		userID, "testuser", "test@example.com", "hashedpassword",
		"Test User", models.UserRoleOperator, models.UserStatusActive,
		null, null, null, null, time.Now(), time.Now(),
	)

	mock.ExpectQuery(`SELECT .+ FROM users WHERE id = \$1 AND deleted_at IS NULL`).WithArgs(userID).WillReturnRows(userRows)

	// Mock permission override query (no overrides)
	mock.ExpectQuery(`SELECT .+ FROM user_permission_overrides WHERE user_id = \$1 AND permission = \$2`).WithArgs(
		userID, permission,
	).WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "permission", "granted", "granted_by", "reason", "expires_at", "created_at", "updated_at"}))

	hasPermission, err := repo.CheckPermission(context.Background(), userID, permission)
	assert.NoError(t, err)
	assert.True(t, hasPermission) // UserRoleOperator should have UserRead permission
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPermissionRepository_CheckPermissions(t *testing.T) {
	repo, mock, cleanup := setupPermissionRepositoryTest(t)
	defer cleanup()

	userID := uuid.New().String()
	permissions := []models.Permission{models.PermissionUserRead, models.PermissionAlertRead}

	// Mock user query
	userRows := sqlmock.NewRows([]string{
		"id", "username", "email", "password_hash", "display_name",
		"role", "status", "phone", "avatar", "department",
		"last_login_at", "created_at", "updated_at",
	}).AddRow(
		userID, "testuser", "test@example.com", "hashedpassword",
		"Test User", models.UserRoleOperator, models.UserStatusActive,
		null, null, null, null, time.Now(), time.Now(),
	)

	mock.ExpectQuery(`SELECT .+ FROM users WHERE id = \$1 AND deleted_at IS NULL`).WithArgs(userID).WillReturnRows(userRows)

	// Mock permission override queries (no overrides)
	for _, permission := range permissions {
		mock.ExpectQuery(`SELECT .+ FROM user_permission_overrides WHERE user_id = \$1 AND permission = \$2`).WithArgs(
			userID, permission,
		).WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "permission", "granted", "granted_by", "reason", "expires_at", "created_at", "updated_at"}))
	}

	results, err := repo.CheckPermissions(context.Background(), userID, permissions)
	assert.NoError(t, err)
	assert.Len(t, results, 2)
	assert.True(t, results[models.PermissionUserRead])
	assert.True(t, results[models.PermissionAlertRead])
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPermissionRepository_GetUserPermissions(t *testing.T) {
	repo, mock, cleanup := setupPermissionRepositoryTest(t)
	defer cleanup()

	userID := uuid.New().String()

	// Mock user query
	userRows := sqlmock.NewRows([]string{
		"id", "username", "email", "password_hash", "display_name",
		"role", "status", "phone", "avatar", "department",
		"last_login_at", "created_at", "updated_at",
	}).AddRow(
		userID, "testuser", "test@example.com", "hashedpassword",
		"Test User", models.UserRoleOperator, models.UserStatusActive,
		null, null, null, null, time.Now(), time.Now(),
	)

	mock.ExpectQuery(`SELECT .+ FROM users WHERE id = \$1 AND deleted_at IS NULL`).WithArgs(userID).WillReturnRows(userRows)

	// Mock permission overrides query
	overrideRows := sqlmock.NewRows([]string{
		"id", "user_id", "permission", "granted", "granted_by", "reason", "expires_at", "created_at", "updated_at",
	}).AddRow(
		uuid.New().String(), userID, models.PermissionSystemManage, true,
		uuid.New().String(), "Special access", null, time.Now(), time.Now(),
	)

	mock.ExpectQuery(`SELECT .+ FROM user_permission_overrides WHERE user_id = \$1`).WithArgs(userID).WillReturnRows(overrideRows)

	permissions, err := repo.GetUserPermissions(context.Background(), userID)
	assert.NoError(t, err)
	assert.NotNil(t, permissions)
	assert.Equal(t, models.UserRoleOperator, permissions.Role)
	assert.Len(t, permissions.Overrides, 1)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPermissionRepository_CreatePermissionGroup(t *testing.T) {
	repo, mock, cleanup := setupPermissionRepositoryTest(t)
	defer cleanup()

	group := &models.PermissionGroup{
		ID:          uuid.New().String(),
		Name:        "Test Group",
		Description: "Test permission group",
		Permissions: []models.Permission{models.PermissionUserRead, models.PermissionAlertRead},
	}

	mock.ExpectExec(`INSERT INTO permission_groups`).WithArgs(
		group.ID, group.Name, group.Description, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
	).WillReturnResult(sqlmock.NewResult(1, 1))

	err := repo.CreatePermissionGroup(context.Background(), group)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPermissionRepository_GetPermissionGroup(t *testing.T) {
	repo, mock, cleanup := setupPermissionRepositoryTest(t)
	defer cleanup()

	groupID := uuid.New().String()
	expectedGroup := &models.PermissionGroup{
		ID:          groupID,
		Name:        "Test Group",
		Description: "Test permission group",
		Permissions: []models.Permission{models.PermissionUserRead, models.PermissionAlertRead},
	}

	rows := sqlmock.NewRows([]string{
		"id", "name", "description", "permissions", "created_at", "updated_at",
	}).AddRow(
		expectedGroup.ID, expectedGroup.Name, expectedGroup.Description,
		`["user:read","alert:read"]`, time.Now(), time.Now(),
	)

	mock.ExpectQuery(`SELECT .+ FROM permission_groups WHERE id = \$1`).WithArgs(groupID).WillReturnRows(rows)

	group, err := repo.GetPermissionGroup(context.Background(), groupID)
	assert.NoError(t, err)
	assert.Equal(t, expectedGroup.ID, group.ID)
	assert.Equal(t, expectedGroup.Name, group.Name)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPermissionRepository_UpdatePermissionGroup(t *testing.T) {
	repo, mock, cleanup := setupPermissionRepositoryTest(t)
	defer cleanup()

	group := &models.PermissionGroup{
		ID:          uuid.New().String(),
		Name:        "Updated Group",
		Description: "Updated permission group",
		Permissions: []models.Permission{models.PermissionUserRead, models.PermissionUserWrite},
	}

	mock.ExpectExec(`UPDATE permission_groups SET`).WithArgs(
		group.Name, group.Description, sqlmock.AnyArg(), sqlmock.AnyArg(), group.ID,
	).WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.UpdatePermissionGroup(context.Background(), group)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPermissionRepository_DeletePermissionGroup(t *testing.T) {
	repo, mock, cleanup := setupPermissionRepositoryTest(t)
	defer cleanup()

	groupID := uuid.New().String()

	mock.ExpectExec(`DELETE FROM permission_groups WHERE id = \$1`).WithArgs(groupID).WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.DeletePermissionGroup(context.Background(), groupID)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPermissionRepository_ListPermissionGroups(t *testing.T) {
	repo, mock, cleanup := setupPermissionRepositoryTest(t)
	defer cleanup()

	rows := sqlmock.NewRows([]string{
		"id", "name", "description", "permissions", "created_at", "updated_at",
	}).AddRow(
		uuid.New().String(), "Group 1", "First group",
		`["user:read"]`, time.Now(), time.Now(),
	).AddRow(
		uuid.New().String(), "Group 2", "Second group",
		`["alert:read"]`, time.Now(), time.Now(),
	)

	mock.ExpectQuery(`SELECT .+ FROM permission_groups ORDER BY created_at DESC`).WillReturnRows(rows)

	groups, err := repo.ListPermissionGroups(context.Background())
	assert.NoError(t, err)
	assert.Len(t, groups, 2)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPermissionRepository_CreatePermissionOverride(t *testing.T) {
	repo, mock, cleanup := setupPermissionRepositoryTest(t)
	defer cleanup()

	override := &models.UserPermissionOverride{
		ID:         uuid.New().String(),
		UserID:     uuid.New().String(),
		Permission: models.PermissionSystemManage,
		Granted:    true,
		GrantedBy:  uuid.New().String(),
		Reason:     "Special access required",
	}

	mock.ExpectExec(`INSERT INTO user_permission_overrides`).WithArgs(
		override.ID, override.UserID, override.Permission, override.Granted,
		override.GrantedBy, override.Reason, override.ExpiresAt,
		sqlmock.AnyArg(), sqlmock.AnyArg(),
	).WillReturnResult(sqlmock.NewResult(1, 1))

	err := repo.CreatePermissionOverride(context.Background(), override)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPermissionRepository_GetPermissionOverride(t *testing.T) {
	repo, mock, cleanup := setupPermissionRepositoryTest(t)
	defer cleanup()

	overrideID := uuid.New().String()
	expectedOverride := &models.UserPermissionOverride{
		ID:         overrideID,
		UserID:     uuid.New().String(),
		Permission: models.PermissionSystemManage,
		Granted:    true,
		GrantedBy:  uuid.New().String(),
		Reason:     "Special access",
	}

	rows := sqlmock.NewRows([]string{
		"id", "user_id", "permission", "granted", "granted_by", "reason", "expires_at", "created_at", "updated_at",
	}).AddRow(
		expectedOverride.ID, expectedOverride.UserID, expectedOverride.Permission,
		expectedOverride.Granted, expectedOverride.GrantedBy, expectedOverride.Reason,
		null, time.Now(), time.Now(),
	)

	mock.ExpectQuery(`SELECT .+ FROM user_permission_overrides WHERE id = \$1`).WithArgs(overrideID).WillReturnRows(rows)

	override, err := repo.GetPermissionOverride(context.Background(), overrideID)
	assert.NoError(t, err)
	assert.Equal(t, expectedOverride.ID, override.ID)
	assert.Equal(t, expectedOverride.Permission, override.Permission)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPermissionRepository_GrantPermission(t *testing.T) {
	repo, mock, cleanup := setupPermissionRepositoryTest(t)
	defer cleanup()

	userID := uuid.New().String()
	permission := models.PermissionSystemManage
	grantedBy := uuid.New().String()
	reason := "Special access required"

	// Mock check for existing override
	mock.ExpectQuery(`SELECT .+ FROM user_permission_overrides WHERE user_id = \$1 AND permission = \$2`).WithArgs(
		userID, permission,
	).WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "permission", "granted", "granted_by", "reason", "expires_at", "created_at", "updated_at"}))

	// Mock insert new override
	mock.ExpectExec(`INSERT INTO user_permission_overrides`).WithArgs(
		sqlmock.AnyArg(), userID, permission, true, grantedBy, reason, null, sqlmock.AnyArg(), sqlmock.AnyArg(),
	).WillReturnResult(sqlmock.NewResult(1, 1))

	err := repo.GrantPermission(context.Background(), userID, permission, grantedBy, reason, nil)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPermissionRepository_RevokePermission(t *testing.T) {
	repo, mock, cleanup := setupPermissionRepositoryTest(t)
	defer cleanup()

	userID := uuid.New().String()
	permission := models.PermissionSystemManage
	revokedBy := uuid.New().String()
	reason := "Access no longer needed"

	// Mock check for existing override
	mock.ExpectQuery(`SELECT .+ FROM user_permission_overrides WHERE user_id = \$1 AND permission = \$2`).WithArgs(
		userID, permission,
	).WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "permission", "granted", "granted_by", "reason", "expires_at", "created_at", "updated_at"}))

	// Mock insert new override
	mock.ExpectExec(`INSERT INTO user_permission_overrides`).WithArgs(
		sqlmock.AnyArg(), userID, permission, false, revokedBy, reason, null, sqlmock.AnyArg(), sqlmock.AnyArg(),
	).WillReturnResult(sqlmock.NewResult(1, 1))

	err := repo.RevokePermission(context.Background(), userID, permission, revokedBy, reason)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPermissionRepository_CleanupExpiredOverrides(t *testing.T) {
	repo, mock, cleanup := setupPermissionRepositoryTest(t)
	defer cleanup()

	mock.ExpectExec(`DELETE FROM user_permission_overrides WHERE expires_at IS NOT NULL AND expires_at < \$1`).WithArgs(
		sqlmock.AnyArg(),
	).WillReturnResult(sqlmock.NewResult(0, 3))

	count, err := repo.CleanupExpiredOverrides(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, int64(3), count)
	assert.NoError(t, mock.ExpectationsWereMet())
}