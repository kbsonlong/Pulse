package models

import (
	"time"
	"errors"
	"regexp"
	"strings"
)

// UserRole 用户角色枚举
type UserRole string

const (
	UserRoleAdmin     UserRole = "admin"     // 管理员
	UserRoleOperator  UserRole = "operator"  // 运维工程师
	UserRoleViewer    UserRole = "viewer"    // 只读用户
	UserRoleDeveloper UserRole = "developer" // 开发者
	UserRoleGuest     UserRole = "guest"     // 访客用户
)

// UserStatus 用户状态枚举
type UserStatus string

const (
	UserStatusActive   UserStatus = "active"   // 激活
	UserStatusInactive UserStatus = "inactive" // 未激活
	UserStatusDisabled UserStatus = "disabled" // 禁用
	UserStatusLocked   UserStatus = "locked"   // 锁定
)

// User 用户模型
type User struct {
	ID          string     `json:"id" db:"id"`
	Username    string     `json:"username" db:"username"`
	Email       string     `json:"email" db:"email"`
	PasswordHash string    `json:"-" db:"password_hash"` // 不在JSON中暴露密码
	DisplayName string     `json:"display_name" db:"display_name"`
	Role        UserRole   `json:"role" db:"role"`
	Status      UserStatus `json:"status" db:"status"`
	Phone       *string    `json:"phone,omitempty" db:"phone"`
	Avatar      *string    `json:"avatar,omitempty" db:"avatar"`
	Department  *string    `json:"department,omitempty" db:"department"`
	LastLoginAt *time.Time `json:"last_login_at,omitempty" db:"last_login_at"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

// UserCreateRequest 创建用户请求
type UserCreateRequest struct {
	Username    string   `json:"username" binding:"required,min=3,max=50"`
	Email       string   `json:"email" binding:"required,email"`
	Password    string   `json:"password" binding:"required,min=8"`
	DisplayName string   `json:"display_name" binding:"required,min=1,max=100"`
	Role        UserRole `json:"role" binding:"required"`
	Phone       *string  `json:"phone,omitempty"`
	Department  *string  `json:"department,omitempty"`
}

// UserUpdateRequest 更新用户请求
type UserUpdateRequest struct {
	DisplayName *string   `json:"display_name,omitempty" binding:"omitempty,min=1,max=100"`
	Role        *UserRole `json:"role,omitempty"`
	Status      *UserStatus `json:"status,omitempty"`
	Phone       *string   `json:"phone,omitempty"`
	Avatar      *string   `json:"avatar,omitempty"`
	Department  *string   `json:"department,omitempty"`
}

// UserLoginRequest 用户登录请求
type UserLoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// UserChangePasswordRequest 修改密码请求
type UserChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

// UserFilter 用户查询过滤器
type UserFilter struct {
	Role       *UserRole   `json:"role,omitempty"`
	Status     *UserStatus `json:"status,omitempty"`
	Department *string     `json:"department,omitempty"`
	Keyword    *string     `json:"keyword,omitempty"` // 搜索用户名、邮箱、显示名
	Page       int         `json:"page" binding:"min=1"`
	PageSize   int         `json:"page_size" binding:"min=1,max=100"`
}

// UserList 用户列表响应
type UserList struct {
	Users      []*User `json:"users"`
	Total      int64   `json:"total"`
	Page       int     `json:"page"`
	PageSize   int     `json:"page_size"`
	TotalPages int     `json:"total_pages"`
}

// 验证方法

// Validate 验证用户数据
func (u *User) Validate() error {
	if strings.TrimSpace(u.Username) == "" {
		return errors.New("用户名不能为空")
	}
	
	if len(u.Username) < 3 || len(u.Username) > 50 {
		return errors.New("用户名长度必须在3-50个字符之间")
	}
	
	// 用户名只能包含字母、数字、下划线和连字符
	usernameRegex := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	if !usernameRegex.MatchString(u.Username) {
		return errors.New("用户名只能包含字母、数字、下划线和连字符")
	}
	
	if strings.TrimSpace(u.Email) == "" {
		return errors.New("邮箱不能为空")
	}
	
	// 简单的邮箱格式验证
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(u.Email) {
		return errors.New("邮箱格式不正确")
	}
	
	if strings.TrimSpace(u.DisplayName) == "" {
		return errors.New("显示名不能为空")
	}
	
	if len(u.DisplayName) > 100 {
		return errors.New("显示名长度不能超过100个字符")
	}
	
	if !u.Role.IsValid() {
		return errors.New("无效的用户角色")
	}
	
	if !u.Status.IsValid() {
		return errors.New("无效的用户状态")
	}
	
	// 验证手机号格式（如果提供）
	if u.Phone != nil && *u.Phone != "" {
		phoneRegex := regexp.MustCompile(`^1[3-9]\d{9}$`)
		if !phoneRegex.MatchString(*u.Phone) {
			return errors.New("手机号格式不正确")
		}
	}
	
	return nil
}

// IsValid 检查用户角色是否有效
func (r UserRole) IsValid() bool {
	switch r {
	case UserRoleAdmin, UserRoleOperator, UserRoleViewer, UserRoleDeveloper, UserRoleGuest:
		return true
	default:
		return false
	}
}

// IsValid 检查用户状态是否有效
func (s UserStatus) IsValid() bool {
	switch s {
	case UserStatusActive, UserStatusInactive, UserStatusDisabled, UserStatusLocked:
		return true
	default:
		return false
	}
}

// HasPermission 检查用户是否有指定权限
func (u *User) HasPermission(permission string) bool {
	// 管理员拥有所有权限
	if u.Role == UserRoleAdmin {
		return true
	}
	
	// 根据角色和权限进行判断
	switch permission {
	case "user:read":
		return u.Role == UserRoleOperator || u.Role == UserRoleViewer || u.Role == UserRoleDeveloper
	case "user:write":
		return u.Role == UserRoleOperator
	case "alert:read":
		return u.Role == UserRoleOperator || u.Role == UserRoleViewer || u.Role == UserRoleDeveloper
	case "alert:write":
		return u.Role == UserRoleOperator
	case "rule:read":
		return u.Role == UserRoleOperator || u.Role == UserRoleViewer || u.Role == UserRoleDeveloper
	case "rule:write":
		return u.Role == UserRoleOperator || u.Role == UserRoleDeveloper
	case "datasource:read":
		return u.Role == UserRoleOperator || u.Role == UserRoleViewer || u.Role == UserRoleDeveloper
	case "datasource:write":
		return u.Role == UserRoleOperator || u.Role == UserRoleDeveloper
	case "ticket:read":
		return u.Role == UserRoleOperator || u.Role == UserRoleViewer || u.Role == UserRoleDeveloper
	case "ticket:write":
		return u.Role == UserRoleOperator
	case "knowledge:read":
		return u.Role == UserRoleOperator || u.Role == UserRoleViewer || u.Role == UserRoleDeveloper
	case "knowledge:write":
		return u.Role == UserRoleOperator
	default:
		return false
	}
}

// IsActive 检查用户是否处于活跃状态
func (u *User) IsActive() bool {
	return u.Status == UserStatusActive
}

// CanLogin 检查用户是否可以登录
func (u *User) CanLogin() bool {
	return u.Status == UserStatusActive || u.Status == UserStatusInactive
}

// Validate 验证创建用户请求
func (req *UserCreateRequest) Validate() error {
	if strings.TrimSpace(req.Username) == "" {
		return errors.New("用户名不能为空")
	}
	
	if len(req.Username) < 3 || len(req.Username) > 50 {
		return errors.New("用户名长度必须在3-50个字符之间")
	}
	
	usernameRegex := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	if !usernameRegex.MatchString(req.Username) {
		return errors.New("用户名只能包含字母、数字、下划线和连字符")
	}
	
	if strings.TrimSpace(req.Email) == "" {
		return errors.New("邮箱不能为空")
	}
	
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(req.Email) {
		return errors.New("邮箱格式不正确")
	}
	
	if len(req.Password) < 8 {
		return errors.New("密码长度不能少于8个字符")
	}
	
	if strings.TrimSpace(req.DisplayName) == "" {
		return errors.New("显示名不能为空")
	}
	
	if len(req.DisplayName) > 100 {
		return errors.New("显示名长度不能超过100个字符")
	}
	
	if !req.Role.IsValid() {
		return errors.New("无效的用户角色")
	}
	
	// 验证手机号格式（如果提供）
	if req.Phone != nil && *req.Phone != "" {
		phoneRegex := regexp.MustCompile(`^1[3-9]\d{9}$`)
		if !phoneRegex.MatchString(*req.Phone) {
			return errors.New("手机号格式不正确")
		}
	}
	
	return nil
}

// Validate 验证修改密码请求
func (req *UserChangePasswordRequest) Validate() error {
	if strings.TrimSpace(req.OldPassword) == "" {
		return errors.New("原密码不能为空")
	}
	
	if len(req.NewPassword) < 8 {
		return errors.New("新密码长度不能少于8个字符")
	}
	
	if req.OldPassword == req.NewPassword {
		return errors.New("新密码不能与原密码相同")
	}
	
	return nil
}