package service

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"pulse/internal/models"
	"pulse/internal/repository"
)

// dataSourceService 数据源服务实现
type dataSourceService struct {
	repoManager repository.RepositoryManager
	logger      *zap.Logger
}

// NewDataSourceService 创建数据源服务实例
func NewDataSourceService(repoManager repository.RepositoryManager, logger *zap.Logger) DataSourceService {
	return &dataSourceService{
		repoManager: repoManager,
		logger:      logger,
	}
}

// Create 创建数据源
func (s *dataSourceService) Create(ctx context.Context, dataSource *models.DataSource) error {
	s.logger.Info("创建数据源", zap.String("name", dataSource.Name), zap.String("type", string(dataSource.Type)))
	
	// 验证数据源配置
	if err := dataSource.Validate(); err != nil {
		s.logger.Error("数据源配置验证失败", zap.Error(err))
		return fmt.Errorf("数据源配置验证失败: %w", err)
	}
	
	// 设置默认状态
	if dataSource.Status == "" {
		dataSource.Status = models.DataSourceStatusActive
	}
	
	// 调用仓储层创建数据源
	if err := s.repoManager.DataSource().Create(ctx, dataSource); err != nil {
		s.logger.Error("创建数据源失败", zap.Error(err))
		return fmt.Errorf("创建数据源失败: %w", err)
	}
	
	s.logger.Info("数据源创建成功", zap.String("id", dataSource.ID))
	return nil
}

// GetByID 根据ID获取数据源
func (s *dataSourceService) GetByID(ctx context.Context, id string) (*models.DataSource, error) {
	s.logger.Debug("获取数据源", zap.String("id", id))
	
	if id == "" {
		return nil, fmt.Errorf("数据源ID不能为空")
	}
	
	dataSource, err := s.repoManager.DataSource().GetByID(ctx, id)
	if err != nil {
		s.logger.Error("获取数据源失败", zap.String("id", id), zap.Error(err))
		return nil, fmt.Errorf("获取数据源失败: %w", err)
	}
	
	if dataSource == nil {
		s.logger.Warn("数据源不存在", zap.String("id", id))
		return nil, fmt.Errorf("数据源不存在: %s", id)
	}
	
	return dataSource, nil
}

// List 获取数据源列表
func (s *dataSourceService) List(ctx context.Context, filter *models.DataSourceFilter) ([]*models.DataSource, int64, error) {
	s.logger.Info("开始获取数据源列表")

	if filter == nil {
		filter = &models.DataSourceFilter{}
	}

	// 设置默认分页参数
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 {
		filter.PageSize = 20
	}

	// 调用仓储层获取数据源列表
	dataSourceList, err := s.repoManager.DataSource().List(ctx, filter)
	if err != nil {
		s.logger.Error("获取数据源列表失败", zap.Error(err))
		return nil, 0, fmt.Errorf("获取数据源列表失败: %w", err)
	}

	if dataSourceList == nil {
		return []*models.DataSource{}, 0, nil
	}

	s.logger.Info("成功获取数据源列表", zap.Int("count", len(dataSourceList.DataSources)), zap.Int64("total", dataSourceList.Total))
	return dataSourceList.DataSources, dataSourceList.Total, nil
}

// Update 更新数据源
func (s *dataSourceService) Update(ctx context.Context, dataSource *models.DataSource) error {
	s.logger.Info("更新数据源", zap.String("id", dataSource.ID))
	
	if dataSource.ID == "" {
		return fmt.Errorf("数据源ID不能为空")
	}
	
	// 验证数据源是否存在
	exists, err := s.repoManager.DataSource().Exists(ctx, dataSource.ID)
	if err != nil {
		s.logger.Error("检查数据源是否存在失败", zap.String("id", dataSource.ID), zap.Error(err))
		return fmt.Errorf("检查数据源是否存在失败: %w", err)
	}
	if !exists {
		s.logger.Warn("数据源不存在", zap.String("id", dataSource.ID))
		return fmt.Errorf("数据源不存在: %s", dataSource.ID)
	}
	
	// 验证数据源配置
	if err := dataSource.Validate(); err != nil {
		s.logger.Error("数据源配置验证失败", zap.Error(err))
		return fmt.Errorf("数据源配置验证失败: %w", err)
	}
	
	// 调用仓储层更新数据源
	if err := s.repoManager.DataSource().Update(ctx, dataSource); err != nil {
		s.logger.Error("更新数据源失败", zap.String("id", dataSource.ID), zap.Error(err))
		return fmt.Errorf("更新数据源失败: %w", err)
	}
	
	s.logger.Info("数据源更新成功", zap.String("id", dataSource.ID))
	return nil
}

// Delete 删除数据源
func (s *dataSourceService) Delete(ctx context.Context, id string) error {
	s.logger.Info("删除数据源", zap.String("id", id))
	
	if id == "" {
		return fmt.Errorf("数据源ID不能为空")
	}
	
	// 验证数据源是否存在
	exists, err := s.repoManager.DataSource().Exists(ctx, id)
	if err != nil {
		s.logger.Error("检查数据源是否存在失败", zap.String("id", id), zap.Error(err))
		return fmt.Errorf("检查数据源是否存在失败: %w", err)
	}
	if !exists {
		s.logger.Warn("数据源不存在", zap.String("id", id))
		return fmt.Errorf("数据源不存在: %s", id)
	}
	
	// 使用软删除
	if err := s.repoManager.DataSource().SoftDelete(ctx, id); err != nil {
		s.logger.Error("删除数据源失败", zap.String("id", id), zap.Error(err))
		return fmt.Errorf("删除数据源失败: %w", err)
	}
	
	s.logger.Info("数据源删除成功", zap.String("id", id))
	return nil
}

// TestConnection 测试数据源连接
func (s *dataSourceService) TestConnection(ctx context.Context, id string) error {
	s.logger.Info("测试数据源连接", zap.String("id", id))
	
	if id == "" {
		return fmt.Errorf("数据源ID不能为空")
	}
	
	// 获取数据源信息
	dataSource, err := s.repoManager.DataSource().GetByID(ctx, id)
	if err != nil {
		s.logger.Error("获取数据源失败", zap.String("id", id), zap.Error(err))
		return fmt.Errorf("获取数据源失败: %w", err)
	}
	
	if dataSource == nil {
		s.logger.Warn("数据源不存在", zap.String("id", id))
		return fmt.Errorf("数据源不存在: %s", id)
	}
	
	// 调用仓储层测试连接
	testResult, err := s.repoManager.DataSource().TestConnection(ctx, dataSource)
	if err != nil {
		s.logger.Error("数据源连接测试失败", zap.String("id", id), zap.Error(err))
		return fmt.Errorf("数据源连接测试失败: %w", err)
	}
	
	// 检查测试结果
	if testResult != nil && !testResult.Success {
		errorMsg := "连接测试失败"
		if testResult.Error != nil {
			errorMsg = *testResult.Error
		}
		s.logger.Error("数据源连接测试失败", zap.String("id", id), zap.String("error", errorMsg))
		return fmt.Errorf("数据源连接测试失败: %s", errorMsg)
	}
	
	s.logger.Info("数据源连接测试成功", zap.String("id", id))
	return nil
}