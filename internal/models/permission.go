package models

import (
	"time"
)

// Permission 权限定义
type Permission string

const (
	// 用户管理权限
	PermissionUserRead   Permission = "user:read"
	PermissionUserWrite  Permission = "user:write"
	PermissionUserDelete Permission = "user:delete"

	// 告警管理权限
	PermissionAlertRead   Permission = "alert:read"
	PermissionAlertWrite  Permission = "alert:write"
	PermissionAlertDelete Permission = "alert:delete"
	PermissionAlertAck    Permission = "alert:acknowledge"
	PermissionAlertClose  Permission = "alert:close"

	// 规则管理权限
	PermissionRuleRead   Permission = "rule:read"
	PermissionRuleWrite  Permission = "rule:write"
	PermissionRuleDelete Permission = "rule:delete"
	PermissionRuleEnable Permission = "rule:enable"

	// 数据源管理权限
	PermissionDataSourceRead   Permission = "datasource:read"
	PermissionDataSourceWrite  Permission = "datasource:write"
	PermissionDataSourceDelete Permission = "datasource:delete"
	PermissionDataSourceTest   Permission = "datasource:test"

	// 工单管理权限
	PermissionTicketRead     Permission = "ticket:read"
	PermissionTicketWrite    Permission = "ticket:write"
	PermissionTicketDelete   Permission = "ticket:delete"
	PermissionTicketAssign   Permission = "ticket:assign"
	PermissionTicketComment  Permission = "ticket:comment"
	PermissionTicketResolve  Permission = "ticket:resolve"

	// 知识库管理权限
	PermissionKnowledgeRead   Permission = "knowledge:read"
	PermissionKnowledgeWrite  Permission = "knowledge:write"
	PermissionKnowledgeDelete Permission = "knowledge:delete"

	// 系统管理权限
	PermissionSystemConfig Permission = "system:config"
	PermissionSystemMonitor Permission = "system:monitor"
	PermissionSystemAudit   Permission = "system:audit"
)

// RolePermissions 角色权限映射
var RolePermissions = map[UserRole][]Permission{
	UserRoleAdmin: {
		// 用户管理
		PermissionUserRead, PermissionUserWrite, PermissionUserDelete,
		// 告警管理
		PermissionAlertRead, PermissionAlertWrite, PermissionAlertDelete,
		PermissionAlertAck, PermissionAlertClose,
		// 规则管理
		PermissionRuleRead, PermissionRuleWrite, PermissionRuleDelete, PermissionRuleEnable,
		// 数据源管理
		PermissionDataSourceRead, PermissionDataSourceWrite, PermissionDataSourceDelete, PermissionDataSourceTest,
		// 工单管理
		PermissionTicketRead, PermissionTicketWrite, PermissionTicketDelete,
		PermissionTicketAssign, PermissionTicketComment, PermissionTicketResolve,
		// 知识库管理
		PermissionKnowledgeRead, PermissionKnowledgeWrite, PermissionKnowledgeDelete,
		// 系统管理
		PermissionSystemConfig, PermissionSystemMonitor, PermissionSystemAudit,
	},
	UserRoleOperator: {
		// 用户管理（只读）
		PermissionUserRead,
		// 告警管理
		PermissionAlertRead, PermissionAlertWrite, PermissionAlertAck, PermissionAlertClose,
		// 规则管理
		PermissionRuleRead, PermissionRuleWrite, PermissionRuleEnable,
		// 数据源管理
		PermissionDataSourceRead, PermissionDataSourceWrite, PermissionDataSourceTest,
		// 工单管理
		PermissionTicketRead, PermissionTicketWrite, PermissionTicketAssign,
		PermissionTicketComment, PermissionTicketResolve,
		// 知识库管理
		PermissionKnowledgeRead, PermissionKnowledgeWrite,
		// 系统监控
		PermissionSystemMonitor,
	},
	UserRoleViewer: {
		// 只读权限
		PermissionUserRead,
		PermissionAlertRead,
		PermissionRuleRead,
		PermissionDataSourceRead,
		PermissionTicketRead, PermissionTicketComment,
		PermissionKnowledgeRead,
		PermissionSystemMonitor,
	},
	UserRoleGuest: {
		// 最基础的只读权限
		PermissionAlertRead,
		PermissionKnowledgeRead,
	},
}

// PermissionGroup 权限组
type PermissionGroup struct {
	ID          string       `json:"id" db:"id"`
	Name        string       `json:"name" db:"name"`
	Description string       `json:"description" db:"description"`
	Permissions []Permission `json:"permissions" db:"permissions"`
	CreatedAt   time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at" db:"updated_at"`
	DeletedAt   *time.Time   `json:"deleted_at,omitempty" db:"deleted_at"`
}

// UserPermissionOverride 用户权限覆盖
type UserPermissionOverride struct {
	ID         string       `json:"id" db:"id"`
	UserID     string       `json:"user_id" db:"user_id"`
	Permission Permission   `json:"permission" db:"permission"`
	Granted    bool         `json:"granted" db:"granted"` // true=授予，false=撤销
	GrantedBy  string       `json:"granted_by" db:"granted_by"`
	Reason     string       `json:"reason" db:"reason"`
	ExpiresAt  *time.Time   `json:"expires_at,omitempty" db:"expires_at"`
	CreatedAt  time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time    `json:"updated_at" db:"updated_at"`
	DeletedAt  *time.Time   `json:"deleted_at,omitempty" db:"deleted_at"`
}

// PermissionCheckRequest 权限检查请求
type PermissionCheckRequest struct {
	UserID     string     `json:"user_id" validate:"required"`
	Permission Permission `json:"permission" validate:"required"`
	Resource   string     `json:"resource,omitempty"` // 可选的资源标识
}

// PermissionCheckResponse 权限检查响应
type PermissionCheckResponse struct {
	Allowed bool   `json:"allowed"`
	Reason  string `json:"reason,omitempty"`
}

// UserPermissionsResponse 用户权限响应
type UserPermissionsResponse struct {
	UserID      string       `json:"user_id"`
	Role        UserRole     `json:"role"`
	Permissions []Permission `json:"permissions"`
	Overrides   []UserPermissionOverride `json:"overrides,omitempty"`
}

// IsValid 检查权限是否有效
func (p Permission) IsValid() bool {
	allPermissions := []Permission{
		// 用户管理权限
		PermissionUserRead, PermissionUserWrite, PermissionUserDelete,
		// 告警管理权限
		PermissionAlertRead, PermissionAlertWrite, PermissionAlertDelete,
		PermissionAlertAck, PermissionAlertClose,
		// 规则管理权限
		PermissionRuleRead, PermissionRuleWrite, PermissionRuleDelete, PermissionRuleEnable,
		// 数据源管理权限
		PermissionDataSourceRead, PermissionDataSourceWrite, PermissionDataSourceDelete, PermissionDataSourceTest,
		// 工单管理权限
		PermissionTicketRead, PermissionTicketWrite, PermissionTicketDelete,
		PermissionTicketAssign, PermissionTicketComment, PermissionTicketResolve,
		// 知识库管理权限
		PermissionKnowledgeRead, PermissionKnowledgeWrite, PermissionKnowledgeDelete,
		// 系统管理权限
		PermissionSystemConfig, PermissionSystemMonitor, PermissionSystemAudit,
	}

	for _, perm := range allPermissions {
		if p == perm {
			return true
		}
	}
	return false
}

// String 返回权限的字符串表示
func (p Permission) String() string {
	return string(p)
}

// GetRolePermissions 获取角色的所有权限
func GetRolePermissions(role UserRole) []Permission {
	if permissions, exists := RolePermissions[role]; exists {
		return permissions
	}
	return []Permission{}
}

// HasRolePermission 检查角色是否有指定权限
func HasRolePermission(role UserRole, permission Permission) bool {
	permissions := GetRolePermissions(role)
	for _, p := range permissions {
		if p == permission {
			return true
		}
	}
	return false
}

// Validate 验证权限组
func (pg *PermissionGroup) Validate() error {
	if pg.Name == "" {
		return fmt.Errorf("权限组名称不能为空")
	}

	if len(pg.Name) > 50 {
		return fmt.Errorf("权限组名称长度不能超过50个字符")
	}

	if len(pg.Description) > 200 {
		return fmt.Errorf("权限组描述长度不能超过200个字符")
	}

	// 验证权限是否有效
	for _, permission := range pg.Permissions {
		if !permission.IsValid() {
			return fmt.Errorf("无效的权限: %s", permission)
		}
	}

	return nil
}

// Validate 验证用户权限覆盖
func (upo *UserPermissionOverride) Validate() error {
	if upo.UserID == "" {
		return fmt.Errorf("用户ID不能为空")
	}

	if !upo.Permission.IsValid() {
		return fmt.Errorf("无效的权限: %s", upo.Permission)
	}

	if upo.GrantedBy == "" {
		return fmt.Errorf("授权人不能为空")
	}

	if len(upo.Reason) > 200 {
		return fmt.Errorf("授权原因长度不能超过200个字符")
	}

	// 如果设置了过期时间，检查是否在未来
	if upo.ExpiresAt != nil && upo.ExpiresAt.Before(time.Now()) {
		return fmt.Errorf("过期时间不能早于当前时间")
	}

	return nil
}

// IsExpired 检查权限覆盖是否已过期
func (upo *UserPermissionOverride) IsExpired() bool {
	if upo.ExpiresAt == nil {
		return false
	}
	return upo.ExpiresAt.Before(time.Now())
}

// IsActive 检查权限覆盖是否有效
func (upo *UserPermissionOverride) IsActive() bool {
	return upo.DeletedAt == nil && !upo.IsExpired()
}