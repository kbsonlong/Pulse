package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"go.uber.org/zap"

	"pulse/internal/models"
	"pulse/internal/repository"
)

// ruleService 规则服务实现
type ruleService struct {
	repoManager repository.RepositoryManager
	logger      *zap.Logger
}

// NewRuleService 创建规则服务实例
func NewRuleService(repoManager repository.RepositoryManager, logger *zap.Logger) RuleService {
	return &ruleService{
		repoManager: repoManager,
		logger:      logger,
	}
}

// Create 创建规则
func (s *ruleService) Create(ctx context.Context, rule *models.Rule) error {
	s.logger.Info("创建规则", zap.String("name", rule.Name))

	// 验证规则
	if err := rule.Validate(); err != nil {
		s.logger.Error("规则验证失败", zap.Error(err))
		return fmt.Errorf("规则验证失败: %w", err)
	}

	// 检查规则名称是否已存在
	existingRule, err := s.repoManager.Rule().GetByName(ctx, rule.Name)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		s.logger.Error("检查规则名称失败", zap.Error(err))
		return fmt.Errorf("检查规则名称失败: %w", err)
	}
	if existingRule != nil {
		return fmt.Errorf("规则名称 '%s' 已存在", rule.Name)
	}

	// 创建规则
	if err := s.repoManager.Rule().Create(ctx, rule); err != nil {
		s.logger.Error("创建规则失败", zap.Error(err))
		return fmt.Errorf("创建规则失败: %w", err)
	}

	s.logger.Info("规则创建成功", zap.String("id", rule.ID), zap.String("name", rule.Name))
	return nil
}

// GetByID 根据ID获取规则
func (s *ruleService) GetByID(ctx context.Context, id string) (*models.Rule, error) {
	s.logger.Debug("获取规则", zap.String("id", id))

	// 验证ID格式
	if id == "" {
		return nil, fmt.Errorf("规则ID不能为空")
	}

	// 获取规则
	rule, err := s.repoManager.Rule().GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			s.logger.Warn("规则不存在", zap.String("id", id))
			return nil, fmt.Errorf("规则不存在")
		}
		s.logger.Error("获取规则失败", zap.String("id", id), zap.Error(err))
		return nil, fmt.Errorf("获取规则失败: %w", err)
	}

	s.logger.Debug("规则获取成功", zap.String("id", id), zap.String("name", rule.Name))
	return rule, nil
}

// List 获取规则列表
func (s *ruleService) List(ctx context.Context, filter *models.RuleFilter) ([]*models.Rule, int64, error) {
	s.logger.Debug("获取规则列表")

	// 设置默认分页参数
	if filter == nil {
		filter = &models.RuleFilter{
			Page:     1,
			PageSize: 20,
		}
	} else {
		if filter.Page <= 0 {
			filter.Page = 1
		}
		if filter.PageSize <= 0 {
			filter.PageSize = 20
		}
		if filter.PageSize > 100 {
			filter.PageSize = 100
		}
	}

	// 获取规则列表
	ruleList, err := s.repoManager.Rule().List(ctx, filter)
	if err != nil {
		s.logger.Error("获取规则列表失败", zap.Error(err))
		return nil, 0, fmt.Errorf("获取规则列表失败: %w", err)
	}

	s.logger.Debug("规则列表获取成功", 
		zap.Int("count", len(ruleList.Rules)), 
		zap.Int64("total", ruleList.Total))
	return ruleList.Rules, ruleList.Total, nil
}

// Update 更新规则
func (s *ruleService) Update(ctx context.Context, rule *models.Rule) error {
	s.logger.Info("更新规则", zap.String("id", rule.ID), zap.String("name", rule.Name))

	// 验证规则
	if err := rule.Validate(); err != nil {
		s.logger.Error("规则验证失败", zap.Error(err))
		return fmt.Errorf("规则验证失败: %w", err)
	}

	// 检查规则是否存在
	existingRule, err := s.repoManager.Rule().GetByID(ctx, rule.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("规则不存在")
		}
		s.logger.Error("检查规则存在性失败", zap.Error(err))
		return fmt.Errorf("检查规则存在性失败: %w", err)
	}

	// 检查名称是否与其他规则冲突
	if existingRule.Name != rule.Name {
		nameConflictRule, err := s.repoManager.Rule().GetByName(ctx, rule.Name)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			s.logger.Error("检查规则名称冲突失败", zap.Error(err))
			return fmt.Errorf("检查规则名称冲突失败: %w", err)
		}
		if nameConflictRule != nil && nameConflictRule.ID != rule.ID {
			return fmt.Errorf("规则名称 '%s' 已被其他规则使用", rule.Name)
		}
	}

	// 更新规则
	if err := s.repoManager.Rule().Update(ctx, rule); err != nil {
		s.logger.Error("更新规则失败", zap.Error(err))
		return fmt.Errorf("更新规则失败: %w", err)
	}

	s.logger.Info("规则更新成功", zap.String("id", rule.ID), zap.String("name", rule.Name))
	return nil
}

// Delete 删除规则
func (s *ruleService) Delete(ctx context.Context, id string) error {
	s.logger.Info("删除规则", zap.String("id", id))

	// 验证ID
	if id == "" {
		return fmt.Errorf("规则ID不能为空")
	}

	// 检查规则是否存在
	existingRule, err := s.repoManager.Rule().GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("规则不存在")
		}
		s.logger.Error("检查规则存在性失败", zap.Error(err))
		return fmt.Errorf("检查规则存在性失败: %w", err)
	}

	// 软删除规则
	if err := s.repoManager.Rule().SoftDelete(ctx, id); err != nil {
		s.logger.Error("删除规则失败", zap.Error(err))
		return fmt.Errorf("删除规则失败: %w", err)
	}

	s.logger.Info("规则删除成功", zap.String("id", id), zap.String("name", existingRule.Name))
	return nil
}

// Enable 启用规则
func (s *ruleService) Enable(ctx context.Context, id string) error {
	s.logger.Info("启用规则", zap.String("id", id))

	// 验证ID
	if id == "" {
		return fmt.Errorf("规则ID不能为空")
	}

	// 检查规则是否存在
	existingRule, err := s.repoManager.Rule().GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("规则不存在")
		}
		s.logger.Error("检查规则存在性失败", zap.Error(err))
		return fmt.Errorf("检查规则存在性失败: %w", err)
	}

	// 检查规则是否已启用
	if existingRule.Enabled {
		s.logger.Info("规则已处于启用状态", zap.String("id", id))
		return nil
	}

	// 启用规则
	if err := s.repoManager.Rule().Activate(ctx, id); err != nil {
		s.logger.Error("启用规则失败", zap.Error(err))
		return fmt.Errorf("启用规则失败: %w", err)
	}

	s.logger.Info("规则启用成功", zap.String("id", id), zap.String("name", existingRule.Name))
	return nil
}

// Disable 禁用规则
func (s *ruleService) Disable(ctx context.Context, id string) error {
	s.logger.Info("禁用规则", zap.String("id", id))

	// 验证ID
	if id == "" {
		return fmt.Errorf("规则ID不能为空")
	}

	// 检查规则是否存在
	existingRule, err := s.repoManager.Rule().GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("规则不存在")
		}
		s.logger.Error("检查规则存在性失败", zap.Error(err))
		return fmt.Errorf("检查规则存在性失败: %w", err)
	}

	// 检查规则是否已禁用
	if !existingRule.Enabled {
		s.logger.Info("规则已处于禁用状态", zap.String("id", id))
		return nil
	}

	// 禁用规则
	if err := s.repoManager.Rule().Deactivate(ctx, id); err != nil {
		s.logger.Error("禁用规则失败", zap.Error(err))
		return fmt.Errorf("禁用规则失败: %w", err)
	}

	s.logger.Info("规则禁用成功", zap.String("id", id), zap.String("name", existingRule.Name))
	return nil
}