package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// JWTClaims JWT声明结构
type JWTClaims struct {
	UserID   string   `json:"user_id"`
	Username string   `json:"username"`
	Roles    []string `json:"roles"`
	Email    string   `json:"email"`
	jwt.RegisteredClaims
}

// AuthService 认证服务接口
type AuthService interface {
	GenerateToken(userID, username, email string, roles []string) (string, error)
	ValidateToken(tokenString string) (*JWTClaims, error)
	ValidateAPIKey(apiKey string) (string, error)
}

// JWTAuthService JWT认证服务实现
type JWTAuthService struct {
	secret  []byte
	apiKeys map[string]string // apiKey -> userID
}

// NewJWTAuthService 创建JWT认证服务
func NewJWTAuthService(secret string, expiration time.Duration) *JWTAuthService {
	return &JWTAuthService{
		secret:  []byte(secret),
		apiKeys: make(map[string]string), // 初始化空的API Keys映射
	}
}

// GenerateToken 生成JWT Token
func (j *JWTAuthService) GenerateToken(userID, username, email string, roles []string) (string, error) {
	now := time.Now()
	claims := JWTClaims{
		UserID:   userID,
		Username: username,
		Roles:    roles,
		Email:    email,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "alert-management-platform",
			Subject:   userID,
			Audience:  []string{"alert-management-api"},
			ExpiresAt: jwt.NewNumericDate(now.Add(24 * time.Hour)), // 24小时过期
			NotBefore: jwt.NewNumericDate(now),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.secret)
}

// ValidateToken 验证JWT Token
func (j *JWTAuthService) ValidateToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return j.secret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

// ValidateAPIKey 验证API Key
func (j *JWTAuthService) ValidateAPIKey(apiKey string) (string, error) {
	if userID, exists := j.apiKeys[apiKey]; exists {
		return userID, nil
	}
	return "", fmt.Errorf("invalid API key")
}

// JWTAuthMiddleware JWT认证中间件
func JWTAuthMiddleware(authService AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从Authorization头获取token
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "missing_authorization_header",
				"message": "Authorization header is required",
			})
			c.Abort()
			return
		}

		// 检查Bearer token格式
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "invalid_authorization_format",
				"message": "Authorization header must be in format: Bearer <token>",
			})
			c.Abort()
			return
		}

		tokenString := parts[1]
		claims, err := authService.ValidateToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "invalid_token",
				"message": "Invalid or expired token",
				"details": err.Error(),
			})
			c.Abort()
			return
		}

		// 将用户信息存储到上下文中
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("user_roles", claims.Roles)
		c.Set("user_email", claims.Email)
		c.Set("jwt_claims", claims)

		c.Next()
	}
}

// APIKeyAuthMiddleware API Key认证中间件
func APIKeyAuthMiddleware(authService AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从X-API-Key头获取API Key
		apiKey := c.GetHeader("X-API-Key")
		if apiKey == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "missing_api_key",
				"message": "X-API-Key header is required",
			})
			c.Abort()
			return
		}

		userID, err := authService.ValidateAPIKey(apiKey)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "invalid_api_key",
				"message": "Invalid API key",
			})
			c.Abort()
			return
		}

		// 将用户信息存储到上下文中
		c.Set("user_id", userID)
		c.Set("auth_method", "api_key")

		c.Next()
	}
}

// OptionalAuthMiddleware 可选认证中间件（支持JWT和API Key）
func OptionalAuthMiddleware(authService AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 尝试JWT认证
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) == 2 && parts[0] == "Bearer" {
				tokenString := parts[1]
				if claims, err := authService.ValidateToken(tokenString); err == nil {
					c.Set("user_id", claims.UserID)
					c.Set("username", claims.Username)
					c.Set("user_roles", claims.Roles)
					c.Set("user_email", claims.Email)
					c.Set("auth_method", "jwt")
					c.Set("jwt_claims", claims)
					c.Next()
					return
				}
			}
		}

		// 尝试API Key认证
		apiKey := c.GetHeader("X-API-Key")
		if apiKey != "" {
			if userID, err := authService.ValidateAPIKey(apiKey); err == nil {
				c.Set("user_id", userID)
				c.Set("auth_method", "api_key")
				c.Next()
				return
			}
		}

		// 无认证信息，继续处理（匿名访问）
		c.Set("auth_method", "anonymous")
		c.Next()
	}
}

// RequireAuthMiddleware 强制认证中间件（支持JWT和API Key）
func RequireAuthMiddleware(authService AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 尝试JWT认证
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) == 2 && parts[0] == "Bearer" {
				tokenString := parts[1]
				if claims, err := authService.ValidateToken(tokenString); err == nil {
					c.Set("user_id", claims.UserID)
					c.Set("username", claims.Username)
					c.Set("user_roles", claims.Roles)
					c.Set("user_email", claims.Email)
					c.Set("auth_method", "jwt")
					c.Set("jwt_claims", claims)
					c.Next()
					return
				}
			}
		}

		// 尝试API Key认证
		apiKey := c.GetHeader("X-API-Key")
		if apiKey != "" {
			if userID, err := authService.ValidateAPIKey(apiKey); err == nil {
				c.Set("user_id", userID)
				c.Set("auth_method", "api_key")
				c.Next()
				return
			}
		}

		// 认证失败
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "authentication_required",
			"message": "Valid JWT token or API key is required",
		})
		c.Abort()
	}
}