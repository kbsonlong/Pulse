package service

import (
	"context"
	"fmt"
	"time"

	"pulse/internal/models"
	"pulse/internal/repository"
)

// userService 用户服务实现
type userService struct {
	userRepo repository.UserRepository
}

// NewUserService 创建用户服务实例
func NewUserService(userRepo repository.UserRepository) UserService {
	return &userService{
		userRepo: userRepo,
	}
}

// Create 创建用户
func (s *userService) Create(ctx context.Context, user *models.User) error {
	// 验证用户数据
	if err := user.Validate(); err != nil {
		return fmt.Errorf("用户数据验证失败: %w", err)
	}

	// 检查用户名是否已存在
	exists, err := s.userRepo.ExistsByUsername(ctx, user.Username)
	if err != nil {
		return fmt.Errorf("检查用户名失败: %w", err)
	}
	if exists {
		return fmt.Errorf("用户名已存在")
	}

	// 检查邮箱是否已存在
	exists, err = s.userRepo.ExistsByEmail(ctx, user.Email)
	if err != nil {
		return fmt.Errorf("检查邮箱失败: %w", err)
	}
	if exists {
		return fmt.Errorf("邮箱已存在")
	}

	// 如果密码为空，生成默认密码
	if user.PasswordHash == "" {
		defaultPassword := "password123" // 默认密码，实际应用中应该生成随机密码
		hashedPassword, err := repository.HashPassword(defaultPassword)
		if err != nil {
			return fmt.Errorf("密码加密失败: %w", err)
		}
		user.PasswordHash = hashedPassword
	}

	// 设置默认状态
	if user.Status == "" {
		user.Status = models.UserStatusInactive
	}

	// 保存到数据库
	if err := s.userRepo.Create(ctx, user); err != nil {
		return fmt.Errorf("创建用户失败: %w", err)
	}

	return nil
}

// GetByID 根据ID获取用户
func (s *userService) GetByID(ctx context.Context, id string) (*models.User, error) {
	if id == "" {
		return nil, fmt.Errorf("用户ID不能为空")
	}

	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("获取用户失败: %w", err)
	}

	return user, nil
}

// GetByEmail 根据邮箱获取用户
func (s *userService) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	if email == "" {
		return nil, fmt.Errorf("邮箱不能为空")
	}

	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("获取用户失败: %w", err)
	}

	return user, nil
}

// List 获取用户列表
func (s *userService) List(ctx context.Context, filter *models.UserFilter) ([]*models.User, int64, error) {
	// 设置默认分页参数
	if filter == nil {
		filter = &models.UserFilter{}
	}
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 {
		filter.PageSize = 20
	}

	// 获取用户列表
	userList, err := s.userRepo.List(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("获取用户列表失败: %w", err)
	}

	// 获取总数
	total, err := s.userRepo.Count(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("获取用户总数失败: %w", err)
	}

	return userList.Users, total, nil
}

// Update 更新用户
func (s *userService) Update(ctx context.Context, user *models.User) error {
	// 验证用户数据
	if err := user.Validate(); err != nil {
		return fmt.Errorf("用户数据验证失败: %w", err)
	}

	// 检查用户是否存在
	existingUser, err := s.userRepo.GetByID(ctx, user.ID)
	if err != nil {
		return fmt.Errorf("获取用户失败: %w", err)
	}
	if existingUser == nil {
		return fmt.Errorf("用户不存在")
	}

	// 如果更新了用户名，检查是否与其他用户冲突
	if user.Username != existingUser.Username {
		exists, err := s.userRepo.ExistsByUsername(ctx, user.Username)
		if err != nil {
			return fmt.Errorf("检查用户名失败: %w", err)
		}
		if exists {
			return fmt.Errorf("用户名已存在")
		}
	}

	// 如果更新了邮箱，检查是否与其他用户冲突
	if user.Email != existingUser.Email {
		exists, err := s.userRepo.ExistsByEmail(ctx, user.Email)
		if err != nil {
			return fmt.Errorf("检查邮箱失败: %w", err)
		}
		if exists {
			return fmt.Errorf("邮箱已存在")
		}
	}

	// 保存更新
	if err := s.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("更新用户失败: %w", err)
	}

	return nil
}

// Delete 删除用户
func (s *userService) Delete(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("用户ID不能为空")
	}

	// 检查用户是否存在
	exists, err := s.userRepo.Exists(ctx, id)
	if err != nil {
		return fmt.Errorf("检查用户存在性失败: %w", err)
	}
	if !exists {
		return fmt.Errorf("用户不存在")
	}

	// 执行软删除
	if err := s.userRepo.SoftDelete(ctx, id); err != nil {
		return fmt.Errorf("删除用户失败: %w", err)
	}

	return nil
}

// UpdatePassword 更新用户密码
func (s *userService) UpdatePassword(ctx context.Context, id string, oldPassword, newPassword string) error {
	if id == "" {
		return fmt.Errorf("用户ID不能为空")
	}
	if oldPassword == "" {
		return fmt.Errorf("旧密码不能为空")
	}
	if newPassword == "" {
		return fmt.Errorf("新密码不能为空")
	}

	// 获取用户信息
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("获取用户失败: %w", err)
	}
	if user == nil {
		return fmt.Errorf("用户不存在")
	}

	// 验证旧密码
	if err := repository.VerifyPasswordHash(oldPassword, user.PasswordHash); err != nil {
		return fmt.Errorf("旧密码错误")
	}

	// 加密新密码
	hashedPassword, err := repository.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("密码加密失败: %w", err)
	}

	// 更新密码
	if err := s.userRepo.UpdatePassword(ctx, id, hashedPassword); err != nil {
		return fmt.Errorf("更新密码失败: %w", err)
	}

	return nil
}

// Activate 激活用户
func (s *userService) Activate(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("用户ID不能为空")
	}

	if err := s.userRepo.Activate(ctx, id); err != nil {
		return fmt.Errorf("激活用户失败: %w", err)
	}

	return nil
}

// Deactivate 停用用户
func (s *userService) Deactivate(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("用户ID不能为空")
	}

	if err := s.userRepo.Deactivate(ctx, id); err != nil {
		return fmt.Errorf("停用用户失败: %w", err)
	}

	return nil
}

// UpdateLastLogin 更新最后登录时间
func (s *userService) UpdateLastLogin(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("用户ID不能为空")
	}

	if err := s.userRepo.UpdateLastLogin(ctx, id, time.Now()); err != nil {
		return fmt.Errorf("更新最后登录时间失败: %w", err)
	}

	return nil
}