package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"pulse/internal/models"
	"pulse/internal/repository"
)

// authService 认证服务实现
type authService struct {
	userRepo repository.UserRepository
	authRepo repository.AuthRepository
	jwtSecret string
	tokenExpiration time.Duration
	refreshTokenExpiration time.Duration
}

// NewAuthService 创建认证服务实例
func NewAuthService(userRepo repository.UserRepository, authRepo repository.AuthRepository, jwtSecret string) AuthService {
	return &authService{
		userRepo: userRepo,
		authRepo: authRepo,
		jwtSecret: jwtSecret,
		tokenExpiration: 24 * time.Hour, // 访问令牌24小时过期
		refreshTokenExpiration: 7 * 24 * time.Hour, // 刷新令牌7天过期
	}
}

// Login 用户登录
func (s *authService) Login(ctx context.Context, email, password string) (*models.AuthToken, error) {
	// 验证输入参数
	if email == "" {
		return nil, fmt.Errorf("邮箱不能为空")
	}
	if password == "" {
		return nil, fmt.Errorf("密码不能为空")
	}

	// 记录登录尝试
	attempt := &models.LoginAttempt{
		ID: uuid.New().String(),
		Identifier: email,
		IPAddress: "", // 这里应该从context中获取IP地址
		UserAgent: "", // 这里应该从context中获取User-Agent
		Success: false,
		CreatedAt: time.Now(),
	}

	// 获取用户信息
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		failReason := "用户不存在"
		attempt.FailReason = &failReason
		s.authRepo.CreateLoginAttempt(ctx, attempt)
		return nil, fmt.Errorf("邮箱或密码错误")
	}

	// 检查用户状态
	if user.Status != models.UserStatusActive {
		failReason := "用户已被禁用"
		attempt.FailReason = &failReason
		s.authRepo.CreateLoginAttempt(ctx, attempt)
		return nil, fmt.Errorf("用户已被禁用")
	}

	// 验证密码
	if err := repository.VerifyPasswordHash(password, user.PasswordHash); err != nil {
		failReason := "密码错误"
		attempt.FailReason = &failReason
		s.authRepo.CreateLoginAttempt(ctx, attempt)
		return nil, fmt.Errorf("邮箱或密码错误")
	}

	// 登录成功，记录成功的登录尝试
	attempt.Success = true
	attempt.FailReason = nil
	s.authRepo.CreateLoginAttempt(ctx, attempt)

	// 更新用户最后登录时间
	if err := s.userRepo.UpdateLastLogin(ctx, user.ID, time.Now()); err != nil {
		// 不返回错误，因为登录已经成功
	}

	// 生成访问令牌
	accessToken, err := s.generateAccessToken(user)
	if err != nil {
		return nil, fmt.Errorf("生成访问令牌失败: %w", err)
	}

	// 生成刷新令牌
	refreshTokenStr, err := s.generateRandomToken()
	if err != nil {
		return nil, fmt.Errorf("生成刷新令牌失败: %w", err)
	}

	// 保存刷新令牌到数据库
	refreshToken := &models.RefreshToken{
		ID: uuid.New().String(),
		UserID: user.ID,
		Token: refreshTokenStr,
		ExpiresAt: time.Now().Add(s.refreshTokenExpiration),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.authRepo.CreateRefreshToken(ctx, refreshToken); err != nil {
		return nil, fmt.Errorf("保存刷新令牌失败: %w", err)
	}

	// 创建认证令牌响应
	authToken := &models.AuthToken{
		ID: uuid.New(),
		UserID: uuid.MustParse(user.ID),
		Token: accessToken,
		TokenType: "Bearer",
		Scope: "read write",
		ExpiresAt: time.Now().Add(s.tokenExpiration),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	return authToken, nil
}

// RefreshToken 刷新访问令牌
func (s *authService) RefreshToken(ctx context.Context, refreshTokenStr string) (*models.AuthToken, error) {
	if refreshTokenStr == "" {
		return nil, fmt.Errorf("刷新令牌不能为空")
	}

	// 获取刷新令牌
	refreshToken, err := s.authRepo.GetRefreshToken(ctx, refreshTokenStr)
	if err != nil {
		return nil, fmt.Errorf("无效的刷新令牌")
	}

	// 检查刷新令牌是否有效
	if !refreshToken.IsValid() {
		return nil, fmt.Errorf("刷新令牌已过期或被撤销")
	}

	// 获取用户信息
	user, err := s.userRepo.GetByID(ctx, refreshToken.UserID)
	if err != nil {
		return nil, fmt.Errorf("用户不存在")
	}

	// 检查用户状态
	if user.Status != models.UserStatusActive {
		return nil, fmt.Errorf("用户已被禁用")
	}

	// 生成新的访问令牌
	accessToken, err := s.generateAccessToken(user)
	if err != nil {
		return nil, fmt.Errorf("生成访问令牌失败: %w", err)
	}

	// 创建认证令牌响应
	authToken := &models.AuthToken{
		ID: uuid.New(),
		UserID: uuid.MustParse(user.ID),
		Token: accessToken,
		TokenType: "Bearer",
		Scope: "read write",
		ExpiresAt: time.Now().Add(s.tokenExpiration),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	return authToken, nil
}

// Logout 用户登出
func (s *authService) Logout(ctx context.Context, token string) error {
	if token == "" {
		return fmt.Errorf("令牌不能为空")
	}

	// 解析JWT令牌获取用户ID
	claims, err := s.parseToken(token)
	if err != nil {
		return fmt.Errorf("无效的令牌")
	}

	// 撤销用户的所有刷新令牌
	if err := s.authRepo.RevokeUserRefreshTokens(ctx, claims.UserID); err != nil {
		return fmt.Errorf("撤销刷新令牌失败: %w", err)
	}

	// 删除用户的所有会话
	if err := s.authRepo.DeleteUserSessions(ctx, claims.UserID); err != nil {
		return fmt.Errorf("删除用户会话失败: %w", err)
	}

	return nil
}

// ValidateToken 验证访问令牌
func (s *authService) ValidateToken(ctx context.Context, token string) (*models.User, error) {
	if token == "" {
		return nil, fmt.Errorf("令牌不能为空")
	}

	// 解析JWT令牌
	claims, err := s.parseToken(token)
	if err != nil {
		return nil, fmt.Errorf("无效的令牌: %w", err)
	}

	// 检查令牌是否过期
	if time.Now().After(claims.ExpiresAt.Time) {
		return nil, fmt.Errorf("令牌已过期")
	}

	// 获取用户信息
	user, err := s.userRepo.GetByID(ctx, claims.UserID)
	if err != nil {
		return nil, fmt.Errorf("用户不存在")
	}

	// 检查用户状态
	if user.Status != models.UserStatusActive {
		return nil, fmt.Errorf("用户已被禁用")
	}

	return user, nil
}

// ResetPassword 重置密码
func (s *authService) ResetPassword(ctx context.Context, email string) error {
	if email == "" {
		return fmt.Errorf("邮箱不能为空")
	}

	// 检查用户是否存在
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		// 为了安全考虑，即使用户不存在也返回成功
		return nil
	}

	// 检查用户状态
	if user.Status != models.UserStatusActive {
		return fmt.Errorf("用户已被禁用")
	}

	// 生成重置令牌
	resetToken, err := s.generateRandomToken()
	if err != nil {
		return fmt.Errorf("生成重置令牌失败: %w", err)
	}

	// TODO: 这里应该发送重置密码邮件
	// 暂时只是记录日志
	fmt.Printf("密码重置令牌: %s (用户: %s)\n", resetToken, email)

	return nil
}

// generateAccessToken 生成访问令牌
func (s *authService) generateAccessToken(user *models.User) (string, error) {
	claims := jwt.MapClaims{
		"user_id": user.ID,
		"username": user.Username,
		"email": user.Email,
		"role": user.Role,
		"exp": time.Now().Add(s.tokenExpiration).Unix(),
		"iat": time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", fmt.Errorf("签名令牌失败: %w", err)
	}

	return tokenString, nil
}

// parseToken 解析JWT令牌
func (s *authService) parseToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.jwtSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("无效的令牌")
}

// generateRandomToken 生成随机令牌
func (s *authService) generateRandomToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// JWTClaims JWT声明结构
type JWTClaims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}