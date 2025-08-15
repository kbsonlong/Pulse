package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"Pulse/internal/models"
)

// AuthRepository 认证仓储接口
type AuthRepository interface {
	// 会话管理
	CreateSession(ctx context.Context, session *models.UserSession) error
	GetSession(ctx context.Context, sessionID string) (*models.UserSession, error)
	GetUserSessions(ctx context.Context, userID string) ([]*models.UserSession, error)
	UpdateSessionLastActivity(ctx context.Context, sessionID string, lastActivity time.Time) error
	DeleteSession(ctx context.Context, sessionID string) error
	DeleteUserSessions(ctx context.Context, userID string) error
	CleanupExpiredSessions(ctx context.Context) (int64, error)

	// 刷新令牌管理
	CreateRefreshToken(ctx context.Context, token *models.RefreshToken) error
	GetRefreshToken(ctx context.Context, tokenID string) (*models.RefreshToken, error)
	GetUserRefreshTokens(ctx context.Context, userID string) ([]*models.RefreshToken, error)
	RevokeRefreshToken(ctx context.Context, tokenID string) error
	RevokeUserRefreshTokens(ctx context.Context, userID string) error
	CleanupExpiredRefreshTokens(ctx context.Context) (int64, error)

	// 登录尝试记录
	CreateLoginAttempt(ctx context.Context, attempt *models.LoginAttempt) error
	GetLoginAttempts(ctx context.Context, identifier string, since time.Time) ([]*models.LoginAttempt, error)
	GetFailedLoginAttempts(ctx context.Context, identifier string, since time.Time) (int, error)
	CleanupOldLoginAttempts(ctx context.Context, before time.Time) (int64, error)
}



// authRepository 认证仓储实现
type authRepository struct {
	db *sqlx.DB
	tx *sqlx.Tx
}

// NewAuthRepository 创建认证仓储实例
func NewAuthRepository(db *sqlx.DB) AuthRepository {
	return &authRepository{
		db: db,
	}
}

// NewAuthRepositoryWithTx 创建带事务的认证仓储实例
func NewAuthRepositoryWithTx(tx *sqlx.Tx) AuthRepository {
	return &authRepository{
		tx: tx,
	}
}

// getExecutor 获取数据库执行器（事务或普通连接）
func (r *authRepository) getExecutor() sqlx.ExtContext {
	if r.tx != nil {
		return r.tx
	}
	return r.db
}

// CreateSession 创建用户会话
func (r *authRepository) CreateSession(ctx context.Context, session *models.UserSession) error {
	if session.ID == "" {
		session.ID = uuid.New().String()
	}

	now := time.Now()
	session.CreatedAt = now
	session.UpdatedAt = now
	session.LastActivity = now

	query := `
		INSERT INTO user_sessions (
			id, user_id, session_token, ip_address, user_agent, 
			last_activity, expires_at, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9
		)`

	_, err := r.getExecutor().ExecContext(ctx, query,
		session.ID, session.UserID, session.SessionToken, session.IPAddress,
		session.UserAgent, session.LastActivity, session.ExpiresAt,
		session.CreatedAt, session.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("创建用户会话失败: %w", err)
	}

	return nil
}

// GetSession 获取用户会话
func (r *authRepository) GetSession(ctx context.Context, sessionID string) (*models.UserSession, error) {
	var session models.UserSession
	query := `
		SELECT id, user_id, session_token, ip_address, user_agent,
		       last_activity, expires_at, created_at, updated_at
		FROM user_sessions 
		WHERE id = $1`

	err := sqlx.GetContext(ctx, r.getExecutor(), &session, query, sessionID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("会话不存在")
		}
		return nil, fmt.Errorf("获取用户会话失败: %w", err)
	}

	return &session, nil
}

// GetUserSessions 获取用户的所有会话
func (r *authRepository) GetUserSessions(ctx context.Context, userID string) ([]*models.UserSession, error) {
	query := `
		SELECT id, user_id, session_token, ip_address, user_agent,
		       last_activity, expires_at, created_at, updated_at
		FROM user_sessions 
		WHERE user_id = $1
		ORDER BY last_activity DESC`

	rows, err := r.getExecutor().QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("获取用户会话列表失败: %w", err)
	}
	defer rows.Close()

	var sessions []*models.UserSession
	for rows.Next() {
		var session models.UserSession
		err := rows.Scan(
			&session.ID, &session.UserID, &session.SessionToken,
			&session.IPAddress, &session.UserAgent, &session.LastActivity,
			&session.ExpiresAt, &session.CreatedAt, &session.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("扫描会话数据失败: %w", err)
		}
		sessions = append(sessions, &session)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("遍历会话数据失败: %w", err)
	}

	return sessions, nil
}

// UpdateSessionLastActivity 更新会话最后活动时间
func (r *authRepository) UpdateSessionLastActivity(ctx context.Context, sessionID string, lastActivity time.Time) error {
	query := `
		UPDATE user_sessions SET 
			last_activity = $1,
			updated_at = $2
		WHERE id = $3`

	result, err := r.getExecutor().ExecContext(ctx, query, lastActivity, time.Now(), sessionID)
	if err != nil {
		return fmt.Errorf("更新会话活动时间失败: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("获取更新结果失败: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("会话不存在")
	}

	return nil
}

// DeleteSession 删除用户会话
func (r *authRepository) DeleteSession(ctx context.Context, sessionID string) error {
	query := `DELETE FROM user_sessions WHERE id = $1`

	result, err := r.getExecutor().ExecContext(ctx, query, sessionID)
	if err != nil {
		return fmt.Errorf("删除用户会话失败: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("获取删除结果失败: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("会话不存在")
	}

	return nil
}

// DeleteUserSessions 删除用户的所有会话
func (r *authRepository) DeleteUserSessions(ctx context.Context, userID string) error {
	query := `DELETE FROM user_sessions WHERE user_id = $1`

	_, err := r.getExecutor().ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("删除用户会话失败: %w", err)
	}

	return nil
}

// CleanupExpiredSessions 清理过期会话
func (r *authRepository) CleanupExpiredSessions(ctx context.Context) (int64, error) {
	now := time.Now()
	query := `DELETE FROM user_sessions WHERE expires_at < $1`

	result, err := r.getExecutor().ExecContext(ctx, query, now)
	if err != nil {
		return 0, fmt.Errorf("清理过期会话失败: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("获取清理结果失败: %w", err)
	}

	return rowsAffected, nil
}

// CreateRefreshToken 创建刷新令牌
func (r *authRepository) CreateRefreshToken(ctx context.Context, token *models.RefreshToken) error {
	if token.ID == "" {
		token.ID = uuid.New().String()
	}

	now := time.Now()
	token.CreatedAt = now
	token.UpdatedAt = now

	query := `
		INSERT INTO refresh_tokens (
			id, user_id, token, expires_at, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6
		)`

	_, err := r.getExecutor().ExecContext(ctx, query,
		token.ID, token.UserID, token.Token, token.ExpiresAt,
		token.CreatedAt, token.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("创建刷新令牌失败: %w", err)
	}

	return nil
}

// GetRefreshToken 获取刷新令牌
func (r *authRepository) GetRefreshToken(ctx context.Context, tokenID string) (*models.RefreshToken, error) {
	var token models.RefreshToken
	query := `
		SELECT id, user_id, token, expires_at, revoked_at, created_at, updated_at
		FROM refresh_tokens 
		WHERE id = $1`

	err := sqlx.GetContext(ctx, r.getExecutor(), &token, query, tokenID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("刷新令牌不存在")
		}
		return nil, fmt.Errorf("获取刷新令牌失败: %w", err)
	}

	return &token, nil
}

// GetUserRefreshTokens 获取用户的刷新令牌列表
func (r *authRepository) GetUserRefreshTokens(ctx context.Context, userID string) ([]*models.RefreshToken, error) {
	query := `
		SELECT id, user_id, token, expires_at, revoked_at, created_at, updated_at
		FROM refresh_tokens 
		WHERE user_id = $1
		ORDER BY created_at DESC`

	rows, err := r.getExecutor().QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("获取用户刷新令牌失败: %w", err)
	}
	defer rows.Close()

	var tokens []*models.RefreshToken
	for rows.Next() {
		var token models.RefreshToken
		err := rows.Scan(
			&token.ID, &token.UserID, &token.Token, &token.ExpiresAt,
			&token.RevokedAt, &token.CreatedAt, &token.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("扫描刷新令牌数据失败: %w", err)
		}
		tokens = append(tokens, &token)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("遍历刷新令牌数据失败: %w", err)
	}

	return tokens, nil
}

// RevokeRefreshToken 撤销刷新令牌
func (r *authRepository) RevokeRefreshToken(ctx context.Context, tokenID string) error {
	now := time.Now()
	query := `
		UPDATE refresh_tokens SET 
			revoked_at = $1,
			updated_at = $1
		WHERE id = $2 AND revoked_at IS NULL`

	result, err := r.getExecutor().ExecContext(ctx, query, now, tokenID)
	if err != nil {
		return fmt.Errorf("撤销刷新令牌失败: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("获取撤销结果失败: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("刷新令牌不存在或已被撤销")
	}

	return nil
}

// RevokeUserRefreshTokens 撤销用户的所有刷新令牌
func (r *authRepository) RevokeUserRefreshTokens(ctx context.Context, userID string) error {
	now := time.Now()
	query := `
		UPDATE refresh_tokens SET 
			revoked_at = $1,
			updated_at = $1
		WHERE user_id = $2 AND revoked_at IS NULL`

	_, err := r.getExecutor().ExecContext(ctx, query, now, userID)
	if err != nil {
		return fmt.Errorf("撤销用户刷新令牌失败: %w", err)
	}

	return nil
}

// CleanupExpiredRefreshTokens 清理过期的刷新令牌
func (r *authRepository) CleanupExpiredRefreshTokens(ctx context.Context) (int64, error) {
	now := time.Now()
	query := `DELETE FROM refresh_tokens WHERE expires_at < $1 OR revoked_at IS NOT NULL`

	result, err := r.getExecutor().ExecContext(ctx, query, now)
	if err != nil {
		return 0, fmt.Errorf("清理过期刷新令牌失败: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("获取清理结果失败: %w", err)
	}

	return rowsAffected, nil
}

// CreateLoginAttempt 创建登录尝试记录
func (r *authRepository) CreateLoginAttempt(ctx context.Context, attempt *models.LoginAttempt) error {
	if attempt.ID == "" {
		attempt.ID = uuid.New().String()
	}

	attempt.CreatedAt = time.Now()

	query := `
		INSERT INTO login_attempts (
			id, identifier, ip_address, user_agent, success, fail_reason, created_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7
		)`

	_, err := r.getExecutor().ExecContext(ctx, query,
		attempt.ID, attempt.Identifier, attempt.IPAddress, attempt.UserAgent,
		attempt.Success, attempt.FailReason, attempt.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("创建登录尝试记录失败: %w", err)
	}

	return nil
}

// GetLoginAttempts 获取登录尝试记录
func (r *authRepository) GetLoginAttempts(ctx context.Context, identifier string, since time.Time) ([]*models.LoginAttempt, error) {
	query := `
		SELECT id, identifier, ip_address, user_agent, success, fail_reason, created_at
		FROM login_attempts 
		WHERE identifier = $1 AND created_at >= $2
		ORDER BY created_at DESC`

	rows, err := r.getExecutor().QueryContext(ctx, query, identifier, since)
	if err != nil {
		return nil, fmt.Errorf("获取登录尝试记录失败: %w", err)
	}
	defer rows.Close()

	var attempts []*models.LoginAttempt
	for rows.Next() {
		var attempt models.LoginAttempt
		err := rows.Scan(
			&attempt.ID, &attempt.Identifier, &attempt.IPAddress,
			&attempt.UserAgent, &attempt.Success, &attempt.FailReason,
			&attempt.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("扫描登录尝试数据失败: %w", err)
		}
		attempts = append(attempts, &attempt)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("遍历登录尝试数据失败: %w", err)
	}

	return attempts, nil
}

// GetFailedLoginAttempts 获取失败的登录尝试次数
func (r *authRepository) GetFailedLoginAttempts(ctx context.Context, identifier string, since time.Time) (int, error) {
	var count int
	query := `
		SELECT COUNT(*) 
		FROM login_attempts 
		WHERE identifier = $1 AND created_at >= $2 AND success = false`

	err := sqlx.GetContext(ctx, r.getExecutor(), &count, query, identifier, since)
	if err != nil {
		return 0, fmt.Errorf("获取失败登录尝试次数失败: %w", err)
	}

	return count, nil
}

// CleanupOldLoginAttempts 清理旧的登录尝试记录
func (r *authRepository) CleanupOldLoginAttempts(ctx context.Context, before time.Time) (int64, error) {
	query := `DELETE FROM login_attempts WHERE created_at < $1`

	result, err := r.getExecutor().ExecContext(ctx, query, before)
	if err != nil {
		return 0, fmt.Errorf("清理旧登录尝试记录失败: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("获取清理结果失败: %w", err)
	}

	return rowsAffected, nil
}