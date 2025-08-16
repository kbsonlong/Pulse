package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// Permission 权限定义
type Permission struct {
	Resource string `json:"resource"` // 资源名称，如 "alerts", "users", "datasources"
	Action   string `json:"action"`   // 操作类型，如 "read", "write", "delete", "admin"
}

// Role 角色定义
type Role struct {
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Permissions []Permission `json:"permissions"`
}

// RBACService RBAC服务接口
type RBACService interface {
	GetUserRoles(userID string) ([]string, error)
	GetRolePermissions(roleName string) ([]Permission, error)
	HasPermission(userID string, resource string, action string) (bool, error)
	CheckPermissions(userID string, requiredPermissions []Permission) (bool, error)
}

// DefaultRBACService 默认RBAC服务实现
type DefaultRBACService struct {
	userRoles       map[string][]string              // userID -> roles
	rolePermissions map[string][]Permission          // roleName -> permissions
	defaultRoles    map[string]map[string][]string   // resource -> action -> roles
}

// NewDefaultRBACService 创建默认RBAC服务
func NewDefaultRBACService() *DefaultRBACService {
	service := &DefaultRBACService{
		userRoles:       make(map[string][]string),
		rolePermissions: make(map[string][]Permission),
		defaultRoles:    make(map[string]map[string][]string),
	}

	// 初始化默认角色和权限
	service.initializeDefaultRoles()
	return service
}

// initializeDefaultRoles 初始化默认角色和权限
func (r *DefaultRBACService) initializeDefaultRoles() {
	// 定义默认角色权限
	roles := map[string][]Permission{
		"admin": {
			{Resource: "*", Action: "*"}, // 管理员拥有所有权限
		},
		"operator": {
			{Resource: "alerts", Action: "read"},
			{Resource: "alerts", Action: "write"},
			{Resource: "alerts", Action: "ack"},
			{Resource: "rules", Action: "read"},
			{Resource: "rules", Action: "write"},
			{Resource: "datasources", Action: "read"},
			{Resource: "tickets", Action: "read"},
			{Resource: "tickets", Action: "write"},
			{Resource: "knowledge", Action: "read"},
		},
		"viewer": {
			{Resource: "alerts", Action: "read"},
			{Resource: "rules", Action: "read"},
			{Resource: "datasources", Action: "read"},
			{Resource: "tickets", Action: "read"},
			{Resource: "knowledge", Action: "read"},
			{Resource: "dashboard", Action: "read"},
		},
		"guest": {
			{Resource: "alerts", Action: "read"},
			{Resource: "dashboard", Action: "read"},
		},
	}

	// 设置角色权限
	for roleName, permissions := range roles {
		r.rolePermissions[roleName] = permissions
	}

	// 设置默认用户角色（示例数据）
	r.userRoles["user-1"] = []string{"admin"}
	r.userRoles["user-2"] = []string{"operator"}
	r.userRoles["demo-user"] = []string{"viewer"}
	r.userRoles["guest-user"] = []string{"guest"}
}

// GetUserRoles 获取用户角色
func (r *DefaultRBACService) GetUserRoles(userID string) ([]string, error) {
	if roles, exists := r.userRoles[userID]; exists {
		return roles, nil
	}
	// 默认返回guest角色
	return []string{"guest"}, nil
}

// GetRolePermissions 获取角色权限
func (r *DefaultRBACService) GetRolePermissions(roleName string) ([]Permission, error) {
	if permissions, exists := r.rolePermissions[roleName]; exists {
		return permissions, nil
	}
	return []Permission{}, nil
}

// HasPermission 检查用户是否有特定权限
func (r *DefaultRBACService) HasPermission(userID string, resource string, action string) (bool, error) {
	userRoles, err := r.GetUserRoles(userID)
	if err != nil {
		return false, err
	}

	for _, roleName := range userRoles {
		permissions, err := r.GetRolePermissions(roleName)
		if err != nil {
			continue
		}

		for _, permission := range permissions {
			// 检查通配符权限
			if permission.Resource == "*" && permission.Action == "*" {
				return true, nil
			}
			// 检查资源通配符
			if permission.Resource == "*" && permission.Action == action {
				return true, nil
			}
			// 检查操作通配符
			if permission.Resource == resource && permission.Action == "*" {
				return true, nil
			}
			// 检查精确匹配
			if permission.Resource == resource && permission.Action == action {
				return true, nil
			}
		}
	}

	return false, nil
}

// CheckPermissions 检查用户是否拥有所需的所有权限
func (r *DefaultRBACService) CheckPermissions(userID string, requiredPermissions []Permission) (bool, error) {
	for _, permission := range requiredPermissions {
		hasPermission, err := r.HasPermission(userID, permission.Resource, permission.Action)
		if err != nil {
			return false, err
		}
		if !hasPermission {
			return false, nil
		}
	}
	return true, nil
}

// RequirePermissionMiddleware 权限检查中间件
func RequirePermissionMiddleware(rbacService RBACService, resource string, action string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取用户ID
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "authentication_required",
				"message": "User authentication is required",
			})
			c.Abort()
			return
		}

		userIDStr, ok := userID.(string)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "invalid_user_context",
				"message": "Invalid user context",
			})
			c.Abort()
			return
		}

		// 检查权限
		hasPermission, err := rbacService.HasPermission(userIDStr, resource, action)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "permission_check_failed",
				"message": "Failed to check permissions",
				"details": err.Error(),
			})
			c.Abort()
			return
		}

		if !hasPermission {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "insufficient_permissions",
				"message": "Insufficient permissions to access this resource",
				"required": gin.H{
					"resource": resource,
					"action":   action,
				},
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireRoleMiddleware 角色检查中间件
func RequireRoleMiddleware(rbacService RBACService, requiredRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取用户角色
		userRoles, exists := c.Get("user_roles")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "authentication_required",
				"message": "User authentication is required",
			})
			c.Abort()
			return
		}

		userRolesList, ok := userRoles.([]string)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "invalid_role_context",
				"message": "Invalid role context",
			})
			c.Abort()
			return
		}

		// 检查是否拥有所需角色
		for _, userRole := range userRolesList {
			for _, requiredRole := range requiredRoles {
				if userRole == requiredRole {
					c.Next()
					return
				}
			}
		}

		c.JSON(http.StatusForbidden, gin.H{
			"error":   "insufficient_role",
			"message": "Insufficient role to access this resource",
			"required": requiredRoles,
			"current":  userRolesList,
		})
		c.Abort()
	}
}

// RequireAnyPermissionMiddleware 任一权限检查中间件（满足其中一个权限即可）
func RequireAnyPermissionMiddleware(rbacService RBACService, permissions []Permission) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取用户ID
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "authentication_required",
				"message": "User authentication is required",
			})
			c.Abort()
			return
		}

		userIDStr, ok := userID.(string)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "invalid_user_context",
				"message": "Invalid user context",
			})
			c.Abort()
			return
		}

		// 检查是否拥有任一权限
		for _, permission := range permissions {
			hasPermission, err := rbacService.HasPermission(userIDStr, permission.Resource, permission.Action)
			if err == nil && hasPermission {
				c.Next()
				return
			}
		}

		c.JSON(http.StatusForbidden, gin.H{
			"error":   "insufficient_permissions",
			"message": "Insufficient permissions to access this resource",
			"required": permissions,
		})
		c.Abort()
	}
}

// ExtractResourceFromPath 从路径中提取资源名称
func ExtractResourceFromPath(path string) string {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) >= 3 && parts[0] == "api" && parts[1] == "v1" {
		return parts[2]
	}
	return "unknown"
}

// ExtractActionFromMethod 从HTTP方法中提取操作类型
func ExtractActionFromMethod(method string) string {
	switch strings.ToUpper(method) {
	case "GET":
		return "read"
	case "POST":
		return "write"
	case "PUT", "PATCH":
		return "write"
	case "DELETE":
		return "delete"
	default:
		return "unknown"
	}
}

// DynamicPermissionMiddleware 动态权限检查中间件（根据路径和方法自动判断）
func DynamicPermissionMiddleware(rbacService RBACService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取用户ID
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "authentication_required",
				"message": "User authentication is required",
			})
			c.Abort()
			return
		}

		userIDStr, ok := userID.(string)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "invalid_user_context",
				"message": "Invalid user context",
			})
			c.Abort()
			return
		}

		// 从路径和方法中提取资源和操作
		resource := ExtractResourceFromPath(c.Request.URL.Path)
		action := ExtractActionFromMethod(c.Request.Method)

		// 检查权限
		hasPermission, err := rbacService.HasPermission(userIDStr, resource, action)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "permission_check_failed",
				"message": "Failed to check permissions",
				"details": err.Error(),
			})
			c.Abort()
			return
		}

		if !hasPermission {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "insufficient_permissions",
				"message": "Insufficient permissions to access this resource",
				"required": gin.H{
					"resource": resource,
					"action":   action,
					"path":     c.Request.URL.Path,
					"method":   c.Request.Method,
				},
			})
			c.Abort()
			return
		}

		c.Next()
	}
}