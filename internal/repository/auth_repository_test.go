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

func setupAuthRepositoryTest(t *testing.T) (*authRepository, sqlmock.Sqlmock, func()) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := NewAuthRepository(sqlxDB)

	cleanup := func() {
		db.Close()
	}

	return repo.(*authRepository), mock, cleanup
}

func TestAuthRepository_CreateSession(t *testing.T) {
	repo, mock, cleanup := setupAuthRepositoryTest(t)
	defer cleanup()

	session := &models.UserSession{
		ID:           uuid.New().String(),
		UserID:       uuid.New().String(),
		SessionToken: "session_token_123",
		UserAgent:    "Mozilla/5.0",
		IPAddress:    "192.168.1.1",
		ExpiresAt:    time.Now().Add(24 * time.Hour),
	}

	mock.ExpectExec(`INSERT INTO user_sessions`).WithArgs(
		session.ID, session.UserID, session.SessionToken, session.UserAgent,
		session.IPAddress, session.ExpiresAt, sqlmock.AnyArg(), sqlmock.AnyArg(),
	).WillReturnResult(sqlmock.NewResult(1, 1))

	err := repo.CreateSession(context.Background(), session)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAuthRepository_GetSession(t *testing.T) {
	repo, mock, cleanup := setupAuthRepositoryTest(t)
	defer cleanup()

	sessionID := uuid.New().String()
	expectedSession := &models.UserSession{
		ID:           sessionID,
		UserID:       uuid.New().String(),
		SessionToken: "session_token_123",
		UserAgent:    "Mozilla/5.0",
		IPAddress:    "192.168.1.1",
		ExpiresAt:    time.Now().Add(24 * time.Hour),
	}

	rows := sqlmock.NewRows([]string{
		"id", "user_id", "token", "user_agent", "ip_address", "expires_at", "created_at", "updated_at",
	}).AddRow(
		expectedSession.ID, expectedSession.UserID, expectedSession.SessionToken,
		expectedSession.UserAgent, expectedSession.IPAddress, expectedSession.ExpiresAt,
		time.Now(), time.Now(),
	)

	mock.ExpectQuery(`SELECT .+ FROM user_sessions WHERE id = \$1`).WithArgs(sessionID).WillReturnRows(rows)

	session, err := repo.GetSession(context.Background(), sessionID)
	assert.NoError(t, err)
	assert.Equal(t, expectedSession.ID, session.ID)
	assert.Equal(t, expectedSession.SessionToken, session.SessionToken)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAuthRepository_GetSessionByToken(t *testing.T) {
	repo, mock, cleanup := setupAuthRepositoryTest(t)
	defer cleanup()

	token := "session_token_123"
	expectedSession := &models.UserSession{
		ID:        uuid.New().String(),
		UserID:    uuid.New().String(),
		Token:     token,
		UserAgent: "Mozilla/5.0",
		IPAddress: "192.168.1.1",
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	rows := sqlmock.NewRows([]string{
		"id", "user_id", "token", "user_agent", "ip_address", "expires_at", "created_at", "updated_at",
	}).AddRow(
		expectedSession.ID, expectedSession.UserID, expectedSession.Token,
		expectedSession.UserAgent, expectedSession.IPAddress, expectedSession.ExpiresAt,
		time.Now(), time.Now(),
	)

	mock.ExpectQuery(`SELECT .+ FROM user_sessions WHERE token = \$1`).WithArgs(token).WillReturnRows(rows)

	session, err := repo.GetSessionByToken(context.Background(), token)
	assert.NoError(t, err)
	assert.Equal(t, expectedSession.Token, session.Token)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAuthRepository_UpdateSession(t *testing.T) {
	repo, mock, cleanup := setupAuthRepositoryTest(t)
	defer cleanup()

	session := &models.UserSession{
		ID:        uuid.New().String(),
		UserID:    uuid.New().String(),
		Token:     "updated_token_123",
		UserAgent: "Mozilla/5.0 Updated",
		IPAddress: "192.168.1.2",
		ExpiresAt: time.Now().Add(48 * time.Hour),
	}

	mock.ExpectExec(`UPDATE user_sessions SET`).WithArgs(
		session.UserID, session.Token, session.UserAgent, session.IPAddress,
		session.ExpiresAt, sqlmock.AnyArg(), session.ID,
	).WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.UpdateSession(context.Background(), session)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAuthRepository_DeleteSession(t *testing.T) {
	repo, mock, cleanup := setupAuthRepositoryTest(t)
	defer cleanup()

	sessionID := uuid.New().String()

	mock.ExpectExec(`DELETE FROM user_sessions WHERE id = \$1`).WithArgs(sessionID).WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.DeleteSession(context.Background(), sessionID)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAuthRepository_DeleteUserSessions(t *testing.T) {
	repo, mock, cleanup := setupAuthRepositoryTest(t)
	defer cleanup()

	userID := uuid.New().String()

	mock.ExpectExec(`DELETE FROM user_sessions WHERE user_id = \$1`).WithArgs(userID).WillReturnResult(sqlmock.NewResult(0, 3))

	err := repo.DeleteUserSessions(context.Background(), userID)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAuthRepository_CleanupExpiredSessions(t *testing.T) {
	repo, mock, cleanup := setupAuthRepositoryTest(t)
	defer cleanup()

	mock.ExpectExec(`DELETE FROM user_sessions WHERE expires_at < \$1`).WithArgs(
		sqlmock.AnyArg(),
	).WillReturnResult(sqlmock.NewResult(0, 5))

	count, err := repo.CleanupExpiredSessions(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, int64(5), count)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAuthRepository_CreateRefreshToken(t *testing.T) {
	repo, mock, cleanup := setupAuthRepositoryTest(t)
	defer cleanup()

	refreshToken := &models.RefreshToken{
		ID:        uuid.New().String(),
		UserID:    uuid.New().String(),
		Token:     "refresh_token_123",
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}

	mock.ExpectExec(`INSERT INTO refresh_tokens`).WithArgs(
		refreshToken.ID, refreshToken.UserID, refreshToken.Token,
		refreshToken.ExpiresAt, false, sqlmock.AnyArg(), sqlmock.AnyArg(),
	).WillReturnResult(sqlmock.NewResult(1, 1))

	err := repo.CreateRefreshToken(context.Background(), refreshToken)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAuthRepository_GetRefreshToken(t *testing.T) {
	repo, mock, cleanup := setupAuthRepositoryTest(t)
	defer cleanup()

	token := "refresh_token_123"
	expectedToken := &models.RefreshToken{
		ID:        uuid.New().String(),
		UserID:    uuid.New().String(),
		Token:     token,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
		Revoked:   false,
	}

	rows := sqlmock.NewRows([]string{
		"id", "user_id", "token", "expires_at", "revoked", "created_at", "updated_at",
	}).AddRow(
		expectedToken.ID, expectedToken.UserID, expectedToken.Token,
		expectedToken.ExpiresAt, expectedToken.Revoked, time.Now(), time.Now(),
	)

	mock.ExpectQuery(`SELECT .+ FROM refresh_tokens WHERE token = \$1`).WithArgs(token).WillReturnRows(rows)

	refreshToken, err := repo.GetRefreshToken(context.Background(), token)
	assert.NoError(t, err)
	assert.Equal(t, expectedToken.Token, refreshToken.Token)
	assert.False(t, refreshToken.Revoked)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAuthRepository_RevokeRefreshToken(t *testing.T) {
	repo, mock, cleanup := setupAuthRepositoryTest(t)
	defer cleanup()

	token := "refresh_token_123"

	mock.ExpectExec(`UPDATE refresh_tokens SET revoked = true, updated_at = \$1 WHERE token = \$2`).WithArgs(
		sqlmock.AnyArg(), token,
	).WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.RevokeRefreshToken(context.Background(), token)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAuthRepository_RevokeUserRefreshTokens(t *testing.T) {
	repo, mock, cleanup := setupAuthRepositoryTest(t)
	defer cleanup()

	userID := uuid.New().String()

	mock.ExpectExec(`UPDATE refresh_tokens SET revoked = true, updated_at = \$1 WHERE user_id = \$2`).WithArgs(
		sqlmock.AnyArg(), userID,
	).WillReturnResult(sqlmock.NewResult(0, 2))

	err := repo.RevokeUserRefreshTokens(context.Background(), userID)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAuthRepository_CleanupExpiredRefreshTokens(t *testing.T) {
	repo, mock, cleanup := setupAuthRepositoryTest(t)
	defer cleanup()

	mock.ExpectExec(`DELETE FROM refresh_tokens WHERE expires_at < \$1`).WithArgs(
		sqlmock.AnyArg(),
	).WillReturnResult(sqlmock.NewResult(0, 4))

	count, err := repo.CleanupExpiredRefreshTokens(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, int64(4), count)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAuthRepository_CreateLoginAttempt(t *testing.T) {
	repo, mock, cleanup := setupAuthRepositoryTest(t)
	defer cleanup()

	attempt := &models.LoginAttempt{
		ID:        uuid.New().String(),
		Username:  "testuser",
		IPAddress: "192.168.1.1",
		UserAgent: "Mozilla/5.0",
		Success:   true,
		UserID:    func() *string { s := uuid.New().String(); return &s }(),
	}

	mock.ExpectExec(`INSERT INTO login_attempts`).WithArgs(
		attempt.ID, attempt.Username, attempt.IPAddress, attempt.UserAgent,
		attempt.Success, attempt.UserID, attempt.FailureReason, sqlmock.AnyArg(),
	).WillReturnResult(sqlmock.NewResult(1, 1))

	err := repo.CreateLoginAttempt(context.Background(), attempt)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAuthRepository_GetLoginAttempts(t *testing.T) {
	repo, mock, cleanup := setupAuthRepositoryTest(t)
	defer cleanup()

	username := "testuser"
	limit := 10

	rows := sqlmock.NewRows([]string{
		"id", "username", "ip_address", "user_agent", "success", "user_id", "failure_reason", "created_at",
	}).AddRow(
		uuid.New().String(), username, "192.168.1.1", "Mozilla/5.0",
		true, uuid.New().String(), null, time.Now(),
	).AddRow(
		uuid.New().String(), username, "192.168.1.2", "Chrome/90.0",
		false, null, "Invalid password", time.Now(),
	)

	mock.ExpectQuery(`SELECT .+ FROM login_attempts WHERE username = \$1 ORDER BY created_at DESC LIMIT \$2`).WithArgs(
		username, limit,
	).WillReturnRows(rows)

	attempts, err := repo.GetLoginAttempts(context.Background(), username, limit)
	assert.NoError(t, err)
	assert.Len(t, attempts, 2)
	assert.True(t, attempts[0].Success)
	assert.False(t, attempts[1].Success)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAuthRepository_GetRecentFailedAttempts(t *testing.T) {
	repo, mock, cleanup := setupAuthRepositoryTest(t)
	defer cleanup()

	username := "testuser"
	since := time.Now().Add(-1 * time.Hour)

	rows := sqlmock.NewRows([]string{
		"id", "username", "ip_address", "user_agent", "success", "user_id", "failure_reason", "created_at",
	}).AddRow(
		uuid.New().String(), username, "192.168.1.1", "Mozilla/5.0",
		false, null, "Invalid password", time.Now(),
	).AddRow(
		uuid.New().String(), username, "192.168.1.2", "Chrome/90.0",
		false, null, "Account locked", time.Now(),
	)

	mock.ExpectQuery(`SELECT .+ FROM login_attempts WHERE username = \$1 AND success = false AND created_at > \$2 ORDER BY created_at DESC`).WithArgs(
		username, since,
	).WillReturnRows(rows)

	attempts, err := repo.GetRecentFailedAttempts(context.Background(), username, since)
	assert.NoError(t, err)
	assert.Len(t, attempts, 2)
	assert.False(t, attempts[0].Success)
	assert.False(t, attempts[1].Success)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAuthRepository_CleanupOldLoginAttempts(t *testing.T) {
	repo, mock, cleanup := setupAuthRepositoryTest(t)
	defer cleanup()

	before := time.Now().Add(-30 * 24 * time.Hour) // 30 days ago

	mock.ExpectExec(`DELETE FROM login_attempts WHERE created_at < \$1`).WithArgs(
		before,
	).WillReturnResult(sqlmock.NewResult(0, 100))

	count, err := repo.CleanupOldLoginAttempts(context.Background(), before)
	assert.NoError(t, err)
	assert.Equal(t, int64(100), count)
	assert.NoError(t, mock.ExpectationsWereMet())
}