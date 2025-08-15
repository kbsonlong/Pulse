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
		session.IPAddress, sqlmock.AnyArg(), session.ExpiresAt, sqlmock.AnyArg(), sqlmock.AnyArg(),
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
		"id", "user_id", "session_token", "user_agent", "ip_address", "last_activity", "expires_at", "created_at", "updated_at",
	}).AddRow(
		expectedSession.ID, expectedSession.UserID, expectedSession.SessionToken,
		expectedSession.UserAgent, expectedSession.IPAddress, time.Now(), expectedSession.ExpiresAt,
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
		ID:           uuid.New().String(),
		UserID:       uuid.New().String(),
		SessionToken: token,
		UserAgent:    "Mozilla/5.0",
		IPAddress:    "192.168.1.1",
		ExpiresAt:    time.Now().Add(24 * time.Hour),
	}

	rows := sqlmock.NewRows([]string{
		"id", "user_id", "session_token", "user_agent", "ip_address", "last_activity", "expires_at", "created_at", "updated_at",
	}).AddRow(
		expectedSession.ID, expectedSession.UserID, expectedSession.SessionToken,
		expectedSession.UserAgent, expectedSession.IPAddress, time.Now(), expectedSession.ExpiresAt,
		time.Now(), time.Now(),
	)

	mock.ExpectQuery(`SELECT .+ FROM user_sessions WHERE session_token = \$1`).WithArgs(token).WillReturnRows(rows)

	session, err := repo.GetSessionByToken(context.Background(), token)
	assert.NoError(t, err)
	assert.Equal(t, expectedSession.SessionToken, session.SessionToken)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAuthRepository_UpdateSessionLastActivity(t *testing.T) {
	repo, mock, cleanup := setupAuthRepositoryTest(t)
	defer cleanup()

	sessionID := uuid.New().String()
	lastActivity := time.Now()

	mock.ExpectExec(`UPDATE user_sessions SET`).WithArgs(
		lastActivity, sqlmock.AnyArg(), sessionID,
	).WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.UpdateSessionLastActivity(context.Background(), sessionID, lastActivity)
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
		refreshToken.ExpiresAt, refreshToken.RevokedAt, sqlmock.AnyArg(), sqlmock.AnyArg(),
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
		RevokedAt: nil,
	}

	rows := sqlmock.NewRows([]string{
		"id", "user_id", "token", "expires_at", "revoked_at", "created_at", "updated_at",
	}).AddRow(
		expectedToken.ID, expectedToken.UserID, expectedToken.Token,
		expectedToken.ExpiresAt, expectedToken.RevokedAt, time.Now(), time.Now(),
	)

	mock.ExpectQuery(`SELECT .+ FROM refresh_tokens WHERE token = \$1`).WithArgs(token).WillReturnRows(rows)

	refreshToken, err := repo.GetRefreshToken(context.Background(), token)
	assert.NoError(t, err)
	assert.Equal(t, expectedToken.Token, refreshToken.Token)
	assert.False(t, refreshToken.IsRevoked())
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAuthRepository_RevokeRefreshToken(t *testing.T) {
	repo, mock, cleanup := setupAuthRepositoryTest(t)
	defer cleanup()

	token := "refresh_token_123"

	mock.ExpectExec(`UPDATE refresh_tokens SET revoked_at = \$1, updated_at = \$2 WHERE token = \$3`).WithArgs(
		sqlmock.AnyArg(), sqlmock.AnyArg(), token,
	).WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.RevokeRefreshToken(context.Background(), token)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAuthRepository_RevokeUserRefreshTokens(t *testing.T) {
	repo, mock, cleanup := setupAuthRepositoryTest(t)
	defer cleanup()

	userID := uuid.New().String()

	mock.ExpectExec(`UPDATE refresh_tokens SET revoked_at = \$1, updated_at = \$2 WHERE user_id = \$3`).WithArgs(
		sqlmock.AnyArg(), sqlmock.AnyArg(), userID,
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
		ID:         uuid.New().String(),
		Identifier: "test@example.com",
		IPAddress:  "192.168.1.1",
		UserAgent:  "Mozilla/5.0",
		Success:    true,
	}

	mock.ExpectExec(`INSERT INTO login_attempts`).WithArgs(
		attempt.ID, attempt.Identifier, attempt.IPAddress, attempt.UserAgent,
		attempt.Success, attempt.FailReason, sqlmock.AnyArg(),
	).WillReturnResult(sqlmock.NewResult(1, 1))

	err := repo.CreateLoginAttempt(context.Background(), attempt)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAuthRepository_GetLoginAttempts(t *testing.T) {
	repo, mock, cleanup := setupAuthRepositoryTest(t)
	defer cleanup()

	identifier := "test@example.com"
	since := time.Now().Add(-24 * time.Hour)
	failReason := "invalid password"

	rows := sqlmock.NewRows([]string{"id", "identifier", "ip_address", "user_agent", "success", "fail_reason", "created_at"}).
		AddRow(uuid.New().String(), identifier, "192.168.1.1", "Mozilla/5.0", true, nil, time.Now()).
		AddRow(uuid.New().String(), identifier, "192.168.1.1", "Mozilla/5.0", false, &failReason, time.Now())

	mock.ExpectQuery(`SELECT (.+) FROM login_attempts`).WithArgs(identifier, since).WillReturnRows(rows)

	attempts, err := repo.GetLoginAttempts(context.Background(), identifier, since)
	assert.NoError(t, err)
	assert.Len(t, attempts, 2)
	assert.Equal(t, identifier, attempts[0].Identifier)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAuthRepository_GetFailedLoginAttempts(t *testing.T) {
	repo, mock, cleanup := setupAuthRepositoryTest(t)
	defer cleanup()

	identifier := "test@example.com"
	since := time.Now().Add(-1 * time.Hour)

	rows := sqlmock.NewRows([]string{"count"}).
		AddRow(3)

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM login_attempts`).WithArgs(identifier, since).WillReturnRows(rows)

	count, err := repo.GetFailedLoginAttempts(context.Background(), identifier, since)
	assert.NoError(t, err)
	assert.Equal(t, 3, count)
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