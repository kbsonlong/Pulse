package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"

	"Pulse/internal/models"
)

// userRepository 用户仓储实现
type userRepository struct {
	db *sqlx.DB
}

// NewUserRepository 创建用户仓储实例
func NewUserRepository(db *sqlx.DB) UserRepository {
	return &userRepository{
		db: db,
	}
}

// Create 创建用户
func (r *userRepository) Create(ctx context.Context, user *models.User) error {
	// 生成用户ID
	if user.ID == "" {
		user.ID = uuid.New().String()
	}

	// 设置创建时间
	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now

	// 设置默认状态
	if user.Status == "" {
		user.Status = models.UserStatusInactive
	}

	query := `
		INSERT INTO users (
			id, username, email, password_hash, display_name, role, status,
			phone, avatar, department, created_at, updated_at
		) VALUES (
			:id, :username, :email, :password_hash, :display_name, :role, :status,
			:phone, :avatar, :department, :created_at, :updated_at
		)`

	_, err := r.db.NamedExecContext(ctx, query, user)
	if err != nil {
		return fmt.Errorf("创建用户失败: %w", err)
	}

	return nil
}

// GetByID 根据ID获取用户
func (r *userRepository) GetByID(ctx context.Context, id string) (*models.User, error) {
	var user models.User
	query := `
		SELECT id, username, email, password_hash, display_name, role, status,
		       phone, avatar, department, last_login_at, created_at, updated_at, deleted_at
		FROM users 
		WHERE id = $1 AND deleted_at IS NULL`

	err := r.db.GetContext(ctx, &user, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("用户不存在")
		}
		return nil, fmt.Errorf("获取用户失败: %w", err)
	}

	return &user, nil
}

// GetByUsername 根据用户名获取用户
func (r *userRepository) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	var user models.User
	query := `
		SELECT id, username, email, password_hash, display_name, role, status,
		       phone, avatar, department, last_login_at, created_at, updated_at, deleted_at
		FROM users 
		WHERE username = $1 AND deleted_at IS NULL`

	err := r.db.GetContext(ctx, &user, query, username)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("用户不存在")
		}
		return nil, fmt.Errorf("获取用户失败: %w", err)
	}

	return &user, nil
}

// GetByEmail 根据邮箱获取用户
func (r *userRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	query := `
		SELECT id, username, email, password_hash, display_name, role, status,
		       phone, avatar, department, last_login_at, created_at, updated_at, deleted_at
		FROM users 
		WHERE email = $1 AND deleted_at IS NULL`

	err := r.db.GetContext(ctx, &user, query, email)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("用户不存在")
		}
		return nil, fmt.Errorf("获取用户失败: %w", err)
	}

	return &user, nil
}

// Update 更新用户
func (r *userRepository) Update(ctx context.Context, user *models.User) error {
	user.UpdatedAt = time.Now()

	query := `
		UPDATE users SET 
			username = :username,
			email = :email,
			password_hash = :password_hash,
			display_name = :display_name,
			role = :role,
			status = :status,
			phone = :phone,
			avatar = :avatar,
			department = :department,
			last_login_at = :last_login_at,
			updated_at = :updated_at
		WHERE id = :id AND deleted_at IS NULL`

	result, err := r.db.NamedExecContext(ctx, query, user)
	if err != nil {
		return fmt.Errorf("更新用户失败: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("获取更新结果失败: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("用户不存在或已被删除")
	}

	return nil
}

// Delete 硬删除用户
func (r *userRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM users WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("删除用户失败: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("获取删除结果失败: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("用户不存在")
	}

	return nil
}

// SoftDelete 软删除用户
func (r *userRepository) SoftDelete(ctx context.Context, id string) error {
	now := time.Now()
	query := `
		UPDATE users SET 
			deleted_at = $1,
			updated_at = $1
		WHERE id = $2 AND deleted_at IS NULL`

	result, err := r.db.ExecContext(ctx, query, now, id)
	if err != nil {
		return fmt.Errorf("软删除用户失败: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("获取删除结果失败: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("用户不存在或已被删除")
	}

	return nil
}

// List 获取用户列表
func (r *userRepository) List(ctx context.Context, filter *models.UserFilter) (*models.UserList, error) {
	if filter == nil {
		filter = &models.UserFilter{Page: 1, PageSize: 20}
	}

	// 设置默认值
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 {
		filter.PageSize = 20
	}
	if filter.PageSize > 100 {
		filter.PageSize = 100
	}

	// 构建查询条件
	var conditions []string
	var args []interface{}
	argIndex := 1

	conditions = append(conditions, "deleted_at IS NULL")

	if filter.Role != nil {
		conditions = append(conditions, fmt.Sprintf("role = $%d", argIndex))
		args = append(args, *filter.Role)
		argIndex++
	}

	if filter.Status != nil {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argIndex))
		args = append(args, *filter.Status)
		argIndex++
	}

	if filter.Department != nil && *filter.Department != "" {
		conditions = append(conditions, fmt.Sprintf("department = $%d", argIndex))
		args = append(args, *filter.Department)
		argIndex++
	}

	if filter.Keyword != nil && *filter.Keyword != "" {
		keyword := "%" + *filter.Keyword + "%"
		conditions = append(conditions, fmt.Sprintf("(username ILIKE $%d OR email ILIKE $%d OR display_name ILIKE $%d)", argIndex, argIndex, argIndex))
		args = append(args, keyword)
		argIndex++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// 获取总数
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM users %s", whereClause)
	var total int64
	err := r.db.GetContext(ctx, &total, countQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("获取用户总数失败: %w", err)
	}

	// 获取用户列表
	offset := (filter.Page - 1) * filter.PageSize
	listQuery := fmt.Sprintf(`
		SELECT id, username, email, display_name, role, status,
		       phone, avatar, department, last_login_at, created_at, updated_at
		FROM users %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d`, whereClause, argIndex, argIndex+1)

	args = append(args, filter.PageSize, offset)

	var users []*models.User
	err = r.db.SelectContext(ctx, &users, listQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("获取用户列表失败: %w", err)
	}

	totalPages := int((total + int64(filter.PageSize) - 1) / int64(filter.PageSize))

	return &models.UserList{
		Users:      users,
		Total:      total,
		Page:       filter.Page,
		PageSize:   filter.PageSize,
		TotalPages: totalPages,
	}, nil
}

// Count 获取用户总数
func (r *userRepository) Count(ctx context.Context, filter *models.UserFilter) (int64, error) {
	var conditions []string
	var args []interface{}
	argIndex := 1

	conditions = append(conditions, "deleted_at IS NULL")

	if filter != nil {
		if filter.Role != nil {
			conditions = append(conditions, fmt.Sprintf("role = $%d", argIndex))
			args = append(args, *filter.Role)
			argIndex++
		}

		if filter.Status != nil {
			conditions = append(conditions, fmt.Sprintf("status = $%d", argIndex))
			args = append(args, *filter.Status)
			argIndex++
		}

		if filter.Department != nil && *filter.Department != "" {
			conditions = append(conditions, fmt.Sprintf("department = $%d", argIndex))
			args = append(args, *filter.Department)
			argIndex++
		}

		if filter.Keyword != nil && *filter.Keyword != "" {
			keyword := "%" + *filter.Keyword + "%"
			conditions = append(conditions, fmt.Sprintf("(username ILIKE $%d OR email ILIKE $%d OR display_name ILIKE $%d)", argIndex, argIndex, argIndex))
			args = append(args, keyword)
			argIndex++
		}
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	query := fmt.Sprintf("SELECT COUNT(*) FROM users %s", whereClause)
	var count int64
	err := r.db.GetContext(ctx, &count, query, args...)
	if err != nil {
		return 0, fmt.Errorf("获取用户总数失败: %w", err)
	}

	return count, nil
}

// Exists 检查用户是否存在
func (r *userRepository) Exists(ctx context.Context, id string) (bool, error) {
	var count int
	query := `SELECT COUNT(*) FROM users WHERE id = $1 AND deleted_at IS NULL`
	err := r.db.GetContext(ctx, &count, query, id)
	if err != nil {
		return false, fmt.Errorf("检查用户存在性失败: %w", err)
	}
	return count > 0, nil
}

// ExistsByUsername 检查用户名是否存在
func (r *userRepository) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	var count int
	query := `SELECT COUNT(*) FROM users WHERE username = $1 AND deleted_at IS NULL`
	err := r.db.GetContext(ctx, &count, query, username)
	if err != nil {
		return false, fmt.Errorf("检查用户名存在性失败: %w", err)
	}
	return count > 0, nil
}

// ExistsByEmail 检查邮箱是否存在
func (r *userRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	var count int
	query := `SELECT COUNT(*) FROM users WHERE email = $1 AND deleted_at IS NULL`
	err := r.db.GetContext(ctx, &count, query, email)
	if err != nil {
		return false, fmt.Errorf("检查邮箱存在性失败: %w", err)
	}
	return count > 0, nil
}

// VerifyPassword 验证用户密码
func (r *userRepository) VerifyPassword(ctx context.Context, username, password string) (*models.User, error) {
	user, err := r.GetByUsername(ctx, username)
	if err != nil {
		return nil, err
	}

	// 检查用户是否可以登录
	if !user.CanLogin() {
		return nil, fmt.Errorf("用户账户已被禁用或锁定")
	}

	// 验证密码
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return nil, fmt.Errorf("用户名或密码错误")
	}

	return user, nil
}

// UpdatePassword 更新用户密码
func (r *userRepository) UpdatePassword(ctx context.Context, id, hashedPassword string) error {
	query := `
		UPDATE users SET 
			password_hash = $1,
			updated_at = $2
		WHERE id = $3 AND deleted_at IS NULL`

	result, err := r.db.ExecContext(ctx, query, hashedPassword, time.Now(), id)
	if err != nil {
		return fmt.Errorf("更新密码失败: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("获取更新结果失败: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("用户不存在或已被删除")
	}

	return nil
}

// UpdateLastLogin 更新最后登录时间
func (r *userRepository) UpdateLastLogin(ctx context.Context, id string, loginTime time.Time) error {
	query := `
		UPDATE users SET 
			last_login_at = $1,
			updated_at = $2
		WHERE id = $3 AND deleted_at IS NULL`

	result, err := r.db.ExecContext(ctx, query, loginTime, time.Now(), id)
	if err != nil {
		return fmt.Errorf("更新最后登录时间失败: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("获取更新结果失败: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("用户不存在或已被删除")
	}

	return nil
}

// UpdateStatus 更新用户状态
func (r *userRepository) UpdateStatus(ctx context.Context, id string, status models.UserStatus) error {
	query := `
		UPDATE users SET 
			status = $1,
			updated_at = $2
		WHERE id = $3 AND deleted_at IS NULL`

	result, err := r.db.ExecContext(ctx, query, status, time.Now(), id)
	if err != nil {
		return fmt.Errorf("更新用户状态失败: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("获取更新结果失败: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("用户不存在或已被删除")
	}

	return nil
}

// Activate 激活用户
func (r *userRepository) Activate(ctx context.Context, id string) error {
	return r.UpdateStatus(ctx, id, models.UserStatusActive)
}

// Deactivate 停用用户
func (r *userRepository) Deactivate(ctx context.Context, id string) error {
	return r.UpdateStatus(ctx, id, models.UserStatusInactive)
}

// BatchCreate 批量创建用户
func (r *userRepository) BatchCreate(ctx context.Context, users []*models.User) error {
	if len(users) == 0 {
		return nil
	}

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("开始事务失败: %w", err)
	}
	defer tx.Rollback()

	for _, user := range users {
		if user.ID == "" {
			user.ID = uuid.New().String()
		}

		now := time.Now()
		user.CreatedAt = now
		user.UpdatedAt = now

		if user.Status == "" {
			user.Status = models.UserStatusInactive
		}

		query := `
			INSERT INTO users (
				id, username, email, password_hash, display_name, role, status,
				phone, avatar, department, created_at, updated_at
			) VALUES (
				:id, :username, :email, :password_hash, :display_name, :role, :status,
				:phone, :avatar, :department, :created_at, :updated_at
			)`

		_, err := tx.NamedExecContext(ctx, query, user)
		if err != nil {
			return fmt.Errorf("批量创建用户失败: %w", err)
		}
	}

	return tx.Commit()
}

// BatchUpdate 批量更新用户
func (r *userRepository) BatchUpdate(ctx context.Context, users []*models.User) error {
	if len(users) == 0 {
		return nil
	}

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("开始事务失败: %w", err)
	}
	defer tx.Rollback()

	for _, user := range users {
		user.UpdatedAt = time.Now()

		query := `
			UPDATE users SET 
				username = :username,
				email = :email,
				password_hash = :password_hash,
				display_name = :display_name,
				role = :role,
				status = :status,
				phone = :phone,
				avatar = :avatar,
				department = :department,
				last_login_at = :last_login_at,
				updated_at = :updated_at
			WHERE id = :id AND deleted_at IS NULL`

		_, err := tx.NamedExecContext(ctx, query, user)
		if err != nil {
			return fmt.Errorf("批量更新用户失败: %w", err)
		}
	}

	return tx.Commit()
}

// BatchDelete 批量删除用户
func (r *userRepository) BatchDelete(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return nil
	}

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("开始事务失败: %w", err)
	}
	defer tx.Rollback()

	now := time.Now()
	for _, id := range ids {
		query := `
			UPDATE users SET 
				deleted_at = $1,
				updated_at = $1
			WHERE id = $2 AND deleted_at IS NULL`

		_, err := tx.ExecContext(ctx, query, now, id)
		if err != nil {
			return fmt.Errorf("批量删除用户失败: %w", err)
		}
	}

	return tx.Commit()
}

// HashPassword 密码加密
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("密码加密失败: %w", err)
	}
	return string(hash), nil
}

// VerifyPasswordHash 验证密码哈希
func VerifyPasswordHash(password, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}