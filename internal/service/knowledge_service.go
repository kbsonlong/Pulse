package service

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"pulse/internal/models"
	"pulse/internal/repository"
)

// knowledgeService 知识库服务实现
type knowledgeService struct {
	repoManager repository.RepositoryManager
	logger      *zap.Logger
}

// NewKnowledgeService 创建知识库服务实例
func NewKnowledgeService(repoManager repository.RepositoryManager, logger *zap.Logger) KnowledgeService {
	return &knowledgeService{
		repoManager: repoManager,
		logger:      logger,
	}
}

// Create 创建知识库条目
func (s *knowledgeService) Create(ctx context.Context, knowledge *models.Knowledge) error {
	if knowledge == nil {
		return fmt.Errorf("知识库条目信息不能为空")
	}

	// 验证必填字段
	if knowledge.Title == "" {
		return fmt.Errorf("标题不能为空")
	}
	if knowledge.Content == "" {
		return fmt.Errorf("内容不能为空")
	}
	if knowledge.AuthorID == "" {
		return fmt.Errorf("作者ID不能为空")
	}

	// 设置默认值
	if knowledge.Status == "" {
		knowledge.Status = models.KnowledgeStatusDraft
	}
	if knowledge.Type == "" {
		knowledge.Type = models.KnowledgeTypeArticle
	}
	if knowledge.Language == "" {
		knowledge.Language = "zh-CN"
	}
	if knowledge.Visibility == "" {
		knowledge.Visibility = models.KnowledgeVisibilityPublic
	}

	err := s.repoManager.Knowledge().Create(ctx, knowledge)
	if err != nil {
		s.logger.Error("创建知识库条目失败", zap.Error(err), zap.String("title", knowledge.Title))
		return fmt.Errorf("创建知识库条目失败: %w", err)
	}

	s.logger.Info("知识库条目创建成功", zap.String("id", knowledge.ID), zap.String("title", knowledge.Title))
	return nil
}

// GetByID 根据ID获取知识库条目
func (s *knowledgeService) GetByID(ctx context.Context, id string) (*models.Knowledge, error) {
	if id == "" {
		return nil, fmt.Errorf("知识库条目ID不能为空")
	}

	knowledge, err := s.repoManager.Knowledge().GetByID(ctx, id)
	if err != nil {
		s.logger.Error("获取知识库条目失败", zap.Error(err), zap.String("id", id))
		return nil, fmt.Errorf("获取知识库条目失败: %w", err)
	}

	return knowledge, nil
}

// List 获取知识库条目列表
func (s *knowledgeService) List(ctx context.Context, filter *models.KnowledgeFilter) ([]*models.Knowledge, int64, error) {
	result, err := s.repoManager.Knowledge().List(ctx, filter)
	if err != nil {
		s.logger.Error("获取知识库条目列表失败", zap.Error(err))
		return nil, 0, fmt.Errorf("获取知识库条目列表失败: %w", err)
	}

	return result.Knowledge, result.Total, nil
}

// Update 更新知识库条目
func (s *knowledgeService) Update(ctx context.Context, knowledge *models.Knowledge) error {
	if knowledge == nil {
		return fmt.Errorf("知识库条目信息不能为空")
	}
	if knowledge.ID == "" {
		return fmt.Errorf("知识库条目ID不能为空")
	}

	// 检查知识库条目是否存在
	existing, err := s.repoManager.Knowledge().GetByID(ctx, knowledge.ID)
	if err != nil {
		s.logger.Error("检查知识库条目存在性失败", zap.Error(err), zap.String("id", knowledge.ID))
		return fmt.Errorf("知识库条目不存在")
	}

	// 保留原有的创建信息
	knowledge.CreatedAt = existing.CreatedAt
	knowledge.AuthorID = existing.AuthorID

	err = s.repoManager.Knowledge().Update(ctx, knowledge)
	if err != nil {
		s.logger.Error("更新知识库条目失败", zap.Error(err), zap.String("id", knowledge.ID))
		return fmt.Errorf("更新知识库条目失败: %w", err)
	}

	s.logger.Info("知识库条目更新成功", zap.String("id", knowledge.ID), zap.String("title", knowledge.Title))
	return nil
}

// Delete 删除知识库条目
func (s *knowledgeService) Delete(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("知识库条目ID不能为空")
	}

	// 检查知识库条目是否存在
	_, err := s.repoManager.Knowledge().GetByID(ctx, id)
	if err != nil {
		s.logger.Error("检查知识库条目存在性失败", zap.Error(err), zap.String("id", id))
		return fmt.Errorf("知识库条目不存在")
	}

	err = s.repoManager.Knowledge().SoftDelete(ctx, id)
	if err != nil {
		s.logger.Error("删除知识库条目失败", zap.Error(err), zap.String("id", id))
		return fmt.Errorf("删除知识库条目失败: %w", err)
	}

	s.logger.Info("知识库条目删除成功", zap.String("id", id))
	return nil
}

// Search 搜索知识库条目
func (s *knowledgeService) Search(ctx context.Context, query string) ([]*models.Knowledge, error) {
	if query == "" {
		return nil, fmt.Errorf("搜索关键词不能为空")
	}

	// 构建搜索过滤器
	filter := &models.KnowledgeFilter{
		Keyword: &query,
		// 只搜索已发布的内容
		Status: func() *models.KnowledgeStatus {
			status := models.KnowledgeStatusPublished
			return &status
		}(),
	}

	result, err := s.repoManager.Knowledge().Search(ctx, query, filter)
	if err != nil {
		s.logger.Error("搜索知识库条目失败", zap.Error(err), zap.String("query", query))
		return nil, fmt.Errorf("搜索知识库条目失败: %w", err)
	}

	return result.Knowledge, nil
}