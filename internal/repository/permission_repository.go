package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"

	"Pulse/internal/models"
)

// permissionRepository 权限仓储实现
type permissionRepository struct {
	db *sqlx.DB
	tx *sqlx.Tx
}

// NewPermissionRepository 创建权限仓储实例
func NewPermissionRepository(db *sqlx.DB) PermissionRepository {
	return &permissionRepository{
		db: db,
	}
}

// NewPermissionRepositoryWithTx 创建带事务的权限仓储实例
func NewPermissionRepositoryWithTx(tx *sqlx.Tx) PermissionRepository {
	return &permissionRepository{
		tx: tx,
	}
}

// getExecutor 获取数据库执行器（事务或普通连接）
func (r *permissionRepository) getExecutor() sqlx.ExtContext {
	if r.tx != nil {
		return r.tx
	}
	return r.db
}

// CheckPermission 检查用户是否有指定权限
func (r *permissionRepository) CheckPermission(ctx context.Context, userID string, permission models.Permission) (bool, error) {
	// 获取用户信息
	var user models.User
	query := `SELECT id, role, status FROM users WHERE id = $1 AND deleted_at IS NULL`
	err := sqlx.GetContext(ctx, r.getExecutor(), &user, query, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, fmt.Errorf("用户不存在")
		}
		return false, fmt.Errorf("获取用户信息失败: %w", err)
	}

	// 检查用户状态
	if !user.IsActive() {
		return false, nil
	}

	// 检查角色权限
	if models.HasRolePermission(user.Role, permission) {
		// 检查是否有权限覆盖撤销了该权限
		override, err := r.getActivePermissionOverride(ctx, userID, permission)
		if err != nil {
			return false, err
		}
		if override != nil && !override.Granted {
			return false, nil // 权限被撤销
		}
		return true, nil
	}

	// 检查是否有权限覆盖授予了该权限
	override, err := r.getActivePermissionOverride(ctx, userID, permission)
	if err != nil {
		return false, err
	}
	if override != nil && override.Granted {
		return true, nil // 权限被授予
	}

	return false, nil
}

// CheckPermissions 批量检查用户权限
func (r *permissionRepository) CheckPermissions(ctx context.Context, userID string, permissions []models.Permission) (map[models.Permission]bool, error) {
	result := make(map[models.Permission]bool)

	for _, permission := range permissions {
		allowed, err := r.CheckPermission(ctx, userID, permission)
		if err != nil {
			return nil, err
		}
		result[permission] = allowed
	}

	return result, nil
}

// GetUserPermissions 获取用户的所有权限
func (r *permissionRepository) GetUserPermissions(ctx context.Context, userID string) ([]models.Permission, error) {
	// 获取用户信息
	var user models.User
	query := `SELECT id, role, status FROM users WHERE id = $1 AND deleted_at IS NULL`
	err := sqlx.GetContext(ctx, r.getExecutor(), &user, query, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("用户不存在")
		}
		return nil, fmt.Errorf("获取用户信息失败: %w", err)
	}

	// 检查用户状态
	if !user.IsActive() {
		return []models.Permission{}, nil
	}

	// 获取角色权限
	rolePermissions := models.GetRolePermissions(user.Role)
	permissionMap := make(map[models.Permission]bool)

	// 添加角色权限
	for _, perm := range rolePermissions {
		permissionMap[perm] = true
	}

	// 获取权限覆盖
	overrides, err := r.GetUserPermissionOverrides(ctx, userID)
	if err != nil {
		return nil, err
	}

	// 应用权限覆盖
	for _, override := range overrides {
		if override.IsActive() {
			if override.Granted {
				permissionMap[override.Permission] = true
			} else {
				delete(permissionMap, override.Permission)
			}
		}
	}

	// 转换为切片
	var permissions []models.Permission
	for perm := range permissionMap {
		permissions = append(permissions, perm)
	}

	return permissions, nil
}

// CreatePermissionGroup 创建权限组
func (r *permissionRepository) CreatePermissionGroup(ctx context.Context, group *models.PermissionGroup) error {
	// 生成ID
	if group.ID == "" {
		group.ID = uuid.New().String()
	}

	// 设置时间
	now := time.Now()
	group.CreatedAt = now
	group.UpdatedAt = now

	// 验证权限组
	if err := group.Validate(); err != nil {
		return err
	}

	// 转换权限为字符串数组
	permissions := make([]string, len(group.Permissions))
	for i, perm := range group.Permissions {
		permissions[i] = string(perm)
	}

	query := `
		INSERT INTO permission_groups (
			id, name, description, permissions, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6
		)`

	_, err := r.getExecutor().ExecContext(ctx, query, group.ID, group.Name, group.Description, pq.Array(permissions), group.CreatedAt, group.UpdatedAt)
	if err != nil {
		return fmt.Errorf("创建权限组失败: %w", err)
	}

	return nil
}

// GetPermissionGroup 获取权限组
func (r *permissionRepository) GetPermissionGroup(ctx context.Context, id string) (*models.PermissionGroup, error) {
	var group models.PermissionGroup
	var permissions pq.StringArray

	query := `
		SELECT id, name, description, permissions, created_at, updated_at, deleted_at
		FROM permission_groups 
		WHERE id = $1 AND deleted_at IS NULL`

	err := r.getExecutor().QueryRowxContext(ctx, query, id).Scan(
		&group.ID, &group.Name, &group.Description, &permissions,
		&group.CreatedAt, &group.UpdatedAt, &group.DeletedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("权限组不存在")
		}
		return nil, fmt.Errorf("获取权限组失败: %w", err)
	}

	// 转换权限
	group.Permissions = make([]models.Permission, len(permissions))
	for i, perm := range permissions {
		group.Permissions[i] = models.Permission(perm)
	}

	return &group, nil
}

// UpdatePermissionGroup 更新权限组
func (r *permissionRepository) UpdatePermissionGroup(ctx context.Context, group *models.PermissionGroup) error {
	group.UpdatedAt = time.Now()

	// 验证权限组
	if err := group.Validate(); err != nil {
		return err
	}

	// 转换权限为字符串数组
	permissions := make([]string, len(group.Permissions))
	for i, perm := range group.Permissions {
		permissions[i] = string(perm)
	}

	query := `
		UPDATE permission_groups SET 
			name = $1,
			description = $2,
			permissions = $3,
			updated_at = $4
		WHERE id = $5 AND deleted_at IS NULL`

	result, err := r.db.ExecContext(ctx, query, group.Name, group.Description, pq.Array(permissions), group.UpdatedAt, group.ID)
	if err != nil {
		return fmt.Errorf("更新权限组失败: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("获取更新结果失败: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("权限组不存在或已被删除")
	}

	return nil
}

// DeletePermissionGroup 删除权限组
func (r *permissionRepository) DeletePermissionGroup(ctx context.Context, id string) error {
	now := time.Now()
	query := `
		UPDATE permission_groups SET 
			deleted_at = $1,
			updated_at = $1
		WHERE id = $2 AND deleted_at IS NULL`

	result, err := r.db.ExecContext(ctx, query, now, id)
	if err != nil {
		return fmt.Errorf("删除权限组失败: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("获取删除结果失败: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("权限组不存在或已被删除")
	}

	return nil
}

// ListPermissionGroups 获取权限组列表
func (r *permissionRepository) ListPermissionGroups(ctx context.Context) ([]*models.PermissionGroup, error) {
	query := `
		SELECT id, name, description, permissions, created_at, updated_at
		FROM permission_groups 
		WHERE deleted_at IS NULL
		ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("获取权限组列表失败: %w", err)
	}
	defer rows.Close()

	var groups []*models.PermissionGroup
	for rows.Next() {
		var group models.PermissionGroup
		var permissions pq.StringArray

		err := rows.Scan(
			&group.ID, &group.Name, &group.Description, &permissions,
			&group.CreatedAt, &group.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("扫描权限组数据失败: %w", err)
		}

		// 转换权限
		group.Permissions = make([]models.Permission, len(permissions))
		for i, perm := range permissions {
			group.Permissions[i] = models.Permission(perm)
		}

		groups = append(groups, &group)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("遍历权限组数据失败: %w", err)
	}

	return groups, nil
}

// CreatePermissionOverride 创建用户权限覆盖
func (r *permissionRepository) CreatePermissionOverride(ctx context.Context, override *models.UserPermissionOverride) error {
	// 生成ID
	if override.ID == "" {
		override.ID = uuid.New().String()
	}

	// 设置时间
	now := time.Now()
	override.CreatedAt = now
	override.UpdatedAt = now

	// 验证权限覆盖
	if err := override.Validate(); err != nil {
		return err
	}

	query := `
		INSERT INTO user_permission_overrides (
			id, user_id, permission, granted, granted_by, reason, expires_at, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9
		)`

	_, err := r.db.ExecContext(ctx, query,
		override.ID, override.UserID, string(override.Permission), override.Granted,
		override.GrantedBy, override.Reason, override.ExpiresAt,
		override.CreatedAt, override.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("创建权限覆盖失败: %w", err)
	}

	return nil
}

// GetPermissionOverride 获取权限覆盖
func (r *permissionRepository) GetPermissionOverride(ctx context.Context, id string) (*models.UserPermissionOverride, error) {
	var override models.UserPermissionOverride
	var permission string

	query := `
		SELECT id, user_id, permission, granted, granted_by, reason, expires_at, created_at, updated_at, deleted_at
		FROM user_permission_overrides 
		WHERE id = $1 AND deleted_at IS NULL`

	err := r.getExecutor().QueryRowxContext(ctx, query, id).Scan(
		&override.ID, &override.UserID, &permission, &override.Granted,
		&override.GrantedBy, &override.Reason, &override.ExpiresAt,
		&override.CreatedAt, &override.UpdatedAt, &override.DeletedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("权限覆盖不存在")
		}
		return nil, fmt.Errorf("获取权限覆盖失败: %w", err)
	}

	override.Permission = models.Permission(permission)
	return &override, nil
}

// UpdatePermissionOverride 更新权限覆盖
func (r *permissionRepository) UpdatePermissionOverride(ctx context.Context, override *models.UserPermissionOverride) error {
	override.UpdatedAt = time.Now()

	// 验证权限覆盖
	if err := override.Validate(); err != nil {
		return err
	}

	query := `
		UPDATE user_permission_overrides SET 
			permission = $1,
			granted = $2,
			granted_by = $3,
			reason = $4,
			expires_at = $5,
			updated_at = $6
		WHERE id = $7 AND deleted_at IS NULL`

	result, err := r.db.ExecContext(ctx, query,
		string(override.Permission), override.Granted, override.GrantedBy,
		override.Reason, override.ExpiresAt, override.UpdatedAt, override.ID,
	)
	if err != nil {
		return fmt.Errorf("更新权限覆盖失败: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("获取更新结果失败: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("权限覆盖不存在或已被删除")
	}

	return nil
}

// DeletePermissionOverride 删除权限覆盖
func (r *permissionRepository) DeletePermissionOverride(ctx context.Context, id string) error {
	now := time.Now()
	query := `
		UPDATE user_permission_overrides SET 
			deleted_at = $1,
			updated_at = $1
		WHERE id = $2 AND deleted_at IS NULL`

	result, err := r.db.ExecContext(ctx, query, now, id)
	if err != nil {
		return fmt.Errorf("删除权限覆盖失败: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("获取删除结果失败: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("权限覆盖不存在或已被删除")
	}

	return nil
}

// GetUserPermissionOverrides 获取用户的权限覆盖列表
func (r *permissionRepository) GetUserPermissionOverrides(ctx context.Context, userID string) ([]*models.UserPermissionOverride, error) {
	query := `
		SELECT id, user_id, permission, granted, granted_by, reason, expires_at, created_at, updated_at
		FROM user_permission_overrides 
		WHERE user_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("获取用户权限覆盖失败: %w", err)
	}
	defer rows.Close()

	var overrides []*models.UserPermissionOverride
	for rows.Next() {
		var override models.UserPermissionOverride
		var permission string

		err := rows.Scan(
			&override.ID, &override.UserID, &permission, &override.Granted,
			&override.GrantedBy, &override.Reason, &override.ExpiresAt,
			&override.CreatedAt, &override.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("扫描权限覆盖数据失败: %w", err)
		}

		override.Permission = models.Permission(permission)
		overrides = append(overrides, &override)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("遍历权限覆盖数据失败: %w", err)
	}

	return overrides, nil
}

// GrantPermission 授予用户权限
func (r *permissionRepository) GrantPermission(ctx context.Context, userID string, permission models.Permission, grantedBy, reason string, expiresAt *time.Time) error {
	// 检查是否已存在相同的权限覆盖
	existingOverride, err := r.getActivePermissionOverride(ctx, userID, permission)
	if err != nil {
		return err
	}

	if existingOverride != nil {
		// 更新现有的权限覆盖
		existingOverride.Granted = true
		existingOverride.GrantedBy = grantedBy
		existingOverride.Reason = reason
		existingOverride.ExpiresAt = expiresAt
		return r.UpdatePermissionOverride(ctx, existingOverride)
	}

	// 创建新的权限覆盖
	override := &models.UserPermissionOverride{
		UserID:     userID,
		Permission: permission,
		Granted:    true,
		GrantedBy:  grantedBy,
		Reason:     reason,
		ExpiresAt:  expiresAt,
	}

	return r.CreatePermissionOverride(ctx, override)
}

// RevokePermission 撤销用户权限
func (r *permissionRepository) RevokePermission(ctx context.Context, userID string, permission models.Permission, revokedBy, reason string) error {
	// 检查是否已存在相同的权限覆盖
	existingOverride, err := r.getActivePermissionOverride(ctx, userID, permission)
	if err != nil {
		return err
	}

	if existingOverride != nil {
		// 更新现有的权限覆盖
		existingOverride.Granted = false
		existingOverride.GrantedBy = revokedBy
		existingOverride.Reason = reason
		existingOverride.ExpiresAt = nil // 撤销权限不设置过期时间
		return r.UpdatePermissionOverride(ctx, existingOverride)
	}

	// 创建新的权限覆盖（撤销）
	override := &models.UserPermissionOverride{
		UserID:     userID,
		Permission: permission,
		Granted:    false,
		GrantedBy:  revokedBy,
		Reason:     reason,
	}

	return r.CreatePermissionOverride(ctx, override)
}

// CleanupExpiredOverrides 清理过期的权限覆盖
func (r *permissionRepository) CleanupExpiredOverrides(ctx context.Context) (int64, error) {
	now := time.Now()
	query := `
		UPDATE user_permission_overrides SET 
			deleted_at = $1,
			updated_at = $1
		WHERE expires_at IS NOT NULL 
		  AND expires_at < $1 
		  AND deleted_at IS NULL`

	result, err := r.db.ExecContext(ctx, query, now)
	if err != nil {
		return 0, fmt.Errorf("清理过期权限覆盖失败: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("获取清理结果失败: %w", err)
	}

	return rowsAffected, nil
}

// getActivePermissionOverride 获取用户的活跃权限覆盖
func (r *permissionRepository) getActivePermissionOverride(ctx context.Context, userID string, permission models.Permission) (*models.UserPermissionOverride, error) {
	var override models.UserPermissionOverride
	var permissionStr string

	query := `
		SELECT id, user_id, permission, granted, granted_by, reason, expires_at, created_at, updated_at
		FROM user_permission_overrides 
		WHERE user_id = $1 
		  AND permission = $2 
		  AND deleted_at IS NULL
		  AND (expires_at IS NULL OR expires_at > $3)
		ORDER BY created_at DESC
		LIMIT 1`

	err := r.getExecutor().QueryRowxContext(ctx, query, userID, string(permission), time.Now()).Scan(
		&override.ID, &override.UserID, &permissionStr, &override.Granted,
		&override.GrantedBy, &override.Reason, &override.ExpiresAt,
		&override.CreatedAt, &override.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // 没有找到活跃的权限覆盖
		}
		return nil, fmt.Errorf("获取权限覆盖失败: %w", err)
	}

	override.Permission = models.Permission(permissionStr)
	return &override, nil
}