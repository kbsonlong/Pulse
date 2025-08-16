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
	a