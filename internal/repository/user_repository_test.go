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
	"golang.org/x/crypto/bcrypt"

	"Pulse/internal/models"
)

func setupUserRepositoryTest(t *testing.T) (*userRepository, sqlmock.Sqlmock, func()) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := NewUserRepository(sqlxDB)

	cleanup := func() {
		db.Close()
	}

	return repo.(*userRepository), mock, cleanup
}

func TestUserRepository_Create(t *testing.T) {
	repo, mock, cleanup := setupUserRepositoryTest(t)
	defer cleanup()

	user := &models.User{
		ID:           uuid.New().String(),
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: "hashedpassword",
		DisplayName:  "Test User",
		Role:         models.UserRoleOperator,
		Status:       models.UserStatusActive,
		Phone:        stringPtr("1234567890"),
		Department:   stringPtr("IT"),
	}

	mock.ExpectExec(`INSERT INTO users`).WithArgs(
		user.ID, user.Username, user.Email, user.PasswordHash,
		user.DisplayName, user.Role, user.Status, user.Phone,
		user.Avatar, user.Department, sqlmock.AnyArg(), sqlmock.AnyArg(),
	).WillReturnResult(sqlmock.NewResult(1, 1))

	err := repo.Create(context.Background(), user)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_GetByID(t *testing.T) {
	repo, mock, cleanup := setupUserRepositoryTest(t)
	defer cleanup()

	userID := uuid.New().String()
	expectedUser := &models.User{
		ID:          userID,
		Username:    "testuser",
		Email:       "test@example.com",
		DisplayName: "Test User",
		Role:        models.UserRoleOperator,
		Status:      models.UserStatusActive,
	}

	rows := sqlmock.NewRows([]string{
		"id", "username", "email", "password_hash", "display_name",
		"role", "status", "phone", "avatar", "department",
		"last_login_at", "created_at", "updated_at",
	}).AddRow(
		expectedUser.ID, expectedUser.Username, expectedUser.Email, "hashedpassword",
		expectedUser.DisplayName, expectedUser.Role, expectedUser.Status,
		null, null, null, null, time.Now(), time.Now(),
	)

	mock.ExpectQuery(`SELECT .+ FROM users WHERE id = \$1 AND deleted_at IS NULL`).WithArgs(userID).WillReturnRows(rows)

	user, err := repo.GetByID(context.Background(), userID)
	assert.NoError(t, err)
	assert.Equal(t, expectedUser.ID, user.ID)
	assert.Equal(t, expectedUser.Username, user.Username)
	assert.Equal(t, expectedUser.Email, user.Email)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_GetByUsername(t *testing.T) {
	repo, mock, cleanup := setupUserRepositoryTest(t)
	defer cleanup()

	username := "testuser"
	expectedUser := &models.User{
		ID:          uuid.New().String(),
		Username:    username,
		Email:       "test@example.com",
		DisplayName: "Test User",
		Role:        models.UserRoleOperator,
		Status:      models.UserStatusActive,
	}

	rows := sqlmock.NewRows([]string{
		"id", "username", "email", "password_hash", "display_name",
		"role", "status", "phone", "avatar", "department",
		"last_login_at", "created_at", "updated_at",
	}).AddRow(
		expectedUser.ID, expectedUser.Username, expectedUser.Email, "hashedpassword",
		expectedUser.DisplayName, expectedUser.Role, expectedUser.Status,
		null, null, null, null, time.Now(), time.Now(),
	)

	mock.ExpectQuery(`SELECT .+ FROM users WHERE username = \$1 AND deleted_at IS NULL`).WithArgs(username).WillReturnRows(rows)

	user, err := repo.GetByUsername(context.Background(), username)
	assert.NoError(t, err)
	assert.Equal(t, expectedUser.Username, user.Username)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_Update(t *testing.T) {
	repo, mock, cleanup := setupUserRepositoryTest(t)
	defer cleanup()

	user := &models.User{
		ID:          uuid.New().String(),
		Username:    "updateduser",
		Email:       "updated@example.com",
		DisplayName: "Updated User",
		Role:        models.UserRoleViewer,
		Status:      models.UserStatusActive,
	}

	mock.ExpectExec(`UPDATE users SET`).WithArgs(
		user.Username, user.Email, user.PasswordHash, user.DisplayName,
		user.Role, user.Status, user.Phone, user.Avatar, user.Department,
		user.LastLoginAt, sqlmock.AnyArg(), user.ID,
	).WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Update(context.Background(), user)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_Delete(t *testing.T) {
	repo, mock, cleanup := setupUserRepositoryTest(t)
	defer cleanup()

	userID := uuid.New().String()

	mock.ExpectExec(`DELETE FROM users WHERE id = \$1`).WithArgs(userID).WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Delete(context.Background(), userID)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_SoftDelete(t *testing.T) {
	repo, mock, cleanup := setupUserRepositoryTest(t)
	defer cleanup()

	userID := uuid.New().String()

	mock.ExpectExec(`UPDATE users SET deleted_at = \$1, updated_at = \$2 WHERE id = \$3 AND deleted_at IS NULL`).WithArgs(
		sqlmock.AnyArg(), sqlmock.AnyArg(), userID,
	).WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.SoftDelete(context.Background(), userID)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_List(t *testing.T) {
	repo, mock, cleanup := setupUserRepositoryTest(t)
	defer cleanup()

	filter := &models.UserFilter{
		Page:     1,
		PageSize: 10,
	}

	// Mock count query
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM users WHERE deleted_at IS NULL`).WillReturnRows(
		sqlmock.NewRows([]string{"count"}).AddRow(2),
	)

	// Mock list query
	rows := sqlmock.NewRows([]string{
		"id", "username", "email", "display_name", "role", "status",
		"phone", "avatar", "department", "last_login_at", "created_at", "updated_at",
	}).AddRow(
		uuid.New().String(), "user1", "user1@example.com", "User 1",
		models.UserRoleOperator, models.UserStatusActive,
		null, null, null, null, time.Now(), time.Now(),
	).AddRow(
		uuid.New().String(), "user2", "user2@example.com", "User 2",
		models.UserRoleViewer, models.UserStatusActive,
		null, null, null, null, time.Now(), time.Now(),
	)

	mock.ExpectQuery(`SELECT .+ FROM users WHERE deleted_at IS NULL ORDER BY created_at DESC LIMIT \$1 OFFSET \$2`).WithArgs(
		10, 0,
	).WillReturnRows(rows)

	userList, err := repo.List(context.Background(), filter)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), userList.Total)
	assert.Len(t, userList.Users, 2)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_Exists(t *testing.T) {
	repo, mock, cleanup := setupUserRepositoryTest(t)
	defer cleanup()

	userID := uuid.New().String()

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM users WHERE id = \$1 AND deleted_at IS NULL`).WithArgs(userID).WillReturnRows(
		sqlmock.NewRows([]string{"count"}).AddRow(1),
	)

	exists, err := repo.Exists(context.Background(), userID)
	assert.NoError(t, err)
	assert.True(t, exists)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_ExistsByUsername(t *testing.T) {
	repo, mock, cleanup := setupUserRepositoryTest(t)
	defer cleanup()

	username := "testuser"

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM users WHERE username = \$1 AND deleted_at IS NULL`).WithArgs(username).WillReturnRows(
		sqlmock.NewRows([]string{"count"}).AddRow(1),
	)

	exists, err := repo.ExistsByUsername(context.Background(), username)
	assert.NoError(t, err)
	assert.True(t, exists)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_VerifyPassword(t *testing.T) {
	repo, mock, cleanup := setupUserRepositoryTest(t)
	defer cleanup()

	username := "testuser"
	password := "testpassword"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	expectedUser := &models.User{
		ID:           uuid.New().String(),
		Username:     username,
		Email:        "test@example.com",
		PasswordHash: string(hashedPassword),
		DisplayName:  "Test User",
		Role:         models.UserRoleOperator,
		Status:       models.UserStatusActive,
	}

	rows := sqlmock.NewRows([]string{
		"id", "username", "email", "password_hash", "display_name",
		"role", "status", "phone", "avatar", "department",
		"last_login_at", "created_at", "updated_at",
	}).AddRow(
		expectedUser.ID, expectedUser.Username, expectedUser.Email, expectedUser.PasswordHash,
		expectedUser.DisplayName, expectedUser.Role, expectedUser.Status,
		null, null, null, null, time.Now(), time.Now(),
	)

	mock.ExpectQuery(`SELECT .+ FROM users WHERE username = \$1 AND deleted_at IS NULL`).WithArgs(username).WillReturnRows(rows)

	user, err := repo.VerifyPassword(context.Background(), username, password)
	assert.NoError(t, err)
	assert.Equal(t, expectedUser.Username, user.Username)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_UpdatePassword(t *testing.T) {
	repo, mock, cleanup := setupUserRepositoryTest(t)
	defer cleanup()

	userID := uuid.New().String()
	newHashedPassword := "newhashedpassword"

	mock.ExpectExec(`UPDATE users SET password_hash = \$1, updated_at = \$2 WHERE id = \$3 AND deleted_at IS NULL`).WithArgs(
		newHashedPassword, sqlmock.AnyArg(), userID,
	).WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.UpdatePassword(context.Background(), userID, newHashedPassword)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_UpdateStatus(t *testing.T) {
	repo, mock, cleanup := setupUserRepositoryTest(t)
	defer cleanup()

	userID := uuid.New().String()
	newStatus := models.UserStatusInactive

	mock.ExpectExec(`UPDATE users SET status = \$1, updated_at = \$2 WHERE id = \$3 AND deleted_at IS NULL`).WithArgs(
		newStatus, sqlmock.AnyArg(), userID,
	).WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.UpdateStatus(context.Background(), userID, newStatus)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_Activate(t *testing.T) {
	repo, mock, cleanup := setupUserRepositoryTest(t)
	defer cleanup()

	userID := uuid.New().String()

	mock.ExpectExec(`UPDATE users SET status = \$1, updated_at = \$2 WHERE id = \$3 AND deleted_at IS NULL`).WithArgs(
		models.UserStatusActive, sqlmock.AnyArg(), userID,
	).WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Activate(context.Background(), userID)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_Deactivate(t *testing.T) {
	repo, mock, cleanup := setupUserRepositoryTest(t)
	defer cleanup()

	userID := uuid.New().String()

	mock.ExpectExec(`UPDATE users SET status = \$1, updated_at = \$2 WHERE id = \$3 AND deleted_at IS NULL`).WithArgs(
		models.UserStatusInactive, sqlmock.AnyArg(), userID,
	).WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Deactivate(context.Background(), userID)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_BatchCreate(t *testing.T) {
	repo, mock, cleanup := setupUserRepositoryTest(t)
	defer cleanup()

	users := []*models.User{
		{
			Username:    "user1",
			Email:       "user1@example.com",
			DisplayName: "User 1",
			Role:        models.UserRoleOperator,
		},
		{
			Username:    "user2",
			Email:       "user2@example.com",
			DisplayName: "User 2",
			Role:        models.UserRoleViewer,
		},
	}

	mock.ExpectBegin()
	for range users {
		mock.ExpectExec(`INSERT INTO users`).WithArgs(
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
		).WillReturnResult(sqlmock.NewResult(1, 1))
	}
	mock.ExpectCommit()

	err := repo.BatchCreate(context.Background(), users)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// Helper functions are now in test_helpers.go