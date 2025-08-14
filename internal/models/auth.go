package models

import (
	"time"
	"fmt"
	"strings"
)

// UserSession 用户会话模型
type UserSession struct {
	ID           string    `json:"id" db:"id"`
	UserID       string    `json:"user_id" db:"user_id"`
	SessionToken string    `json:"session_token" db:"session_token"`
	IPAddress    string    `json:"ip_address" db:"ip_address"`
	UserAgent    string    `json:"user_agent" db:"user_agent"`
	LastActivity time.Time `json:"last_activity" db:"last_activity"`
	ExpiresAt    time.Time `json:"expires_at" db:"expires_at"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// RefreshToken 刷新令牌模型
type RefreshToken struct {
	ID        string     `json:"id" db:"id"`
	UserID    string     `json:"user_id" db:"user_id"`
	Token     string     `json:"token" db:"token"`
	ExpiresAt time.Time  `json:"expires_at" db:"expires_at"`
	RevokedAt *time.Time `json:"revoked_at" db:"revoked_at"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
}

// LoginAttempt 登录尝试记录模型
type LoginAttempt struct {
	ID         string    `json:"id" db:"id"`
	Identifier string    `json:"identifier" db:"identifier"` // 用户名、邮箱或IP地址
	IPAddress  string    `json:"ip_address" db:"ip_address"`
	UserAgent  string    `json:"user_agent" db:"user_agent"`
	Success    bool      `json:"success" db:"success"`
	FailReason string    `json:"fail_reason" db:"fail_reason"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}

// AuthRequest 认证请求
type AuthRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// AuthResponse 认证响应
type AuthResponse struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	TokenType    string    `json:"token_type"`
	ExpiresIn    int64     `json:"expires_in"`
	User         *User     `json:"user"`
	ExpiresAt    time.Time `json:"expires_at"`
}

// RefreshTokenRequest 刷新令牌请求
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// SessionInfo 会话信息
type SessionInfo struct {
	SessionID    string    `json:"session_id"`
	UserID       string    `json:"user_id"`
	IPAddress    string    `json:"ip_address"`
	UserAgent    string    `json:"user_agent"`
	LastActivity time.Time `json:"last_activity"`
	ExpiresAt    time.Time `json:"expires_at"`
	IsActive     bool      `json:"is_active"`
}

// UserSessionList 用户会话列表
type UserSessionList struct {
	Sessions []*SessionInfo `json:"sessions"`
	Total    int            `json:"total"`
}

// IsExpired 检查会话是否过期
func (s *UserSession) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}

// IsActive 检查会话是否活跃
func (s *UserSession) IsActive(maxInactivity time.Duration) bool {
	return !s.IsExpired() && time.Since(s.LastActivity) <= maxInactivity
}

// Validate 验证用户会话
func (s *UserSession) Validate() error {
	if strings.TrimSpace(s.UserID) == "" {
		return fmt.Errorf("用户ID不能为空")
	}
	if strings.TrimSpace(s.SessionToken) == "" {
		return fmt.Errorf("会话令牌不能为空")
	}
	if s.ExpiresAt.IsZero() {
		return fmt.Errorf("过期时间不能为空")
	}
	if s.ExpiresAt.Before(time.Now()) {
		return fmt.Errorf("过期时间不能早于当前时间")
	}
	return nil
}

// IsExpired 检查刷新令牌是否过期
func (r *RefreshToken) IsExpired() bool {
	return time.Now().After(r.ExpiresAt)
}

// IsRevoked 检查刷新令牌是否被撤销
func (r *RefreshToken) IsRevoked() bool {
	return r.RevokedAt != nil
}

// IsValid 检查刷新令牌是否有效
func (r *RefreshToken) IsValid() bool {
	return !r.IsExpired() && !r.IsRevoked()
}

// Validate 验证刷新令牌
func (r *RefreshToken) Validate() error {
	if strings.TrimSpace(r.UserID) == "" {
		return fmt.Errorf("用户ID不能为空")
	}
	if strings.TrimSpace(r.Token) == "" {
		return fmt.Errorf("令牌不能为空")
	}
	if r.ExpiresAt.IsZero() {
		return fmt.Errorf("过期时间不能为空")
	}
	if r.ExpiresAt.Before(time.Now()) {
		return fmt.Errorf("过期时间不能早于当前时间")
	}
	return nil
}

// Validate 验证登录尝试记录
func (l *LoginAttempt) Validate() error {
	if strings.TrimSpace(l.Identifier) == "" {
		return fmt.Errorf("标识符不能为空")
	}
	if strings.TrimSpace(l.IPAddress) == "" {
		return fmt.Errorf("IP地址不能为空")
	}
	if !l.Success && strings.TrimSpace(l.FailReason) == "" {
		return fmt.Errorf("失败原因不能为空")
	}
	return nil
}

// Validate 验证认证请求
func (a *AuthRequest) Validate() error {
	if strings.TrimSpace(a.Username) == "" {
		return fmt.Errorf("用户名不能为空")
	}
	if strings.TrimSpace(a.Password) == "" {
		return fmt.Errorf("密码不能为空")
	}
	if len(a.Password) < 6 {
		return fmt.Errorf("密码长度不能少于6位")
	}
	return nil
}

// Validate 验证刷新令牌请求
func (r *RefreshTokenRequest) Validate() error {
	if strings.TrimSpace(r.RefreshToken) == "" {
		return fmt.Errorf("刷新令牌不能为空")
	}
	return nil
}

// ToSessionInfo 转换为会话信息
func (s *UserSession) ToSessionInfo(maxInactivity time.Duration) *SessionInfo {
	return &SessionInfo{
		SessionID:    s.ID,
		UserID:       s.UserID,
		IPAddress:    s.IPAddress,
		UserAgent:    s.UserAgent,
		LastActivity: s.LastActivity,
		ExpiresAt:    s.ExpiresAt,
		IsActive:     s.IsActive(maxInactivity),
	}
}