package service

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"pulse/internal/models"
	"pulse/internal/repository"
)

// ticketService 工单服务实现
type ticketService struct {
	repoManager repository.RepositoryManager
	logger      *zap.Logger
}

// NewTicketService 创建工单服务实例
func NewTicketService(repoManager repository.RepositoryManager, logger *zap.Logger) TicketService {
	return &ticketService{
		repoManager: repoManager,
		logger:      logger,
	}
}

// Create 创建工单
func (s *ticketService) Create(ctx context.Context, ticket *models.Ticket) error {
	if ticket == nil {
		return fmt.Errorf("工单信息不能为空")
	}

	// 验证必填字段
	if ticket.Title == "" {
		return fmt.Errorf("工单标题不能为空")
	}
	if ticket.ReporterID == "" {
		return fmt.Errorf("报告人ID不能为空")
	}

	// 生成工单编号
	if ticket.Number == "" {
		ticket.Number = s.generateTicketNumber()
	}

	// 设置默认值
	if ticket.Status == "" {
		ticket.Status = models.TicketStatusOpen
	}
	if ticket.Priority == "" {
		ticket.Priority = models.TicketPriorityMedium
	}

	// 创建工单
	err := s.repoManager.Ticket().Create(ctx, ticket)
	if err != nil {
		s.logger.Error("创建工单失败", zap.Error(err), zap.String("title", ticket.Title))
		return fmt.Errorf("创建工单失败: %w", err)
	}

	s.logger.Info("工单创建成功", zap.String("id", ticket.ID), zap.String("number", ticket.Number))
	return nil
}

// GetByID 根据ID获取工单
func (s *ticketService) GetByID(ctx context.Context, id string) (*models.Ticket, error) {
	if id == "" {
		return nil, fmt.Errorf("工单ID不能为空")
	}

	ticket, err := s.repoManager.Ticket().GetByID(ctx, id)
	if err != nil {
		s.logger.Error("获取工单失败", zap.Error(err), zap.String("id", id))
		return nil, fmt.Errorf("获取工单失败: %w", err)
	}

	return ticket, nil
}

// List 获取工单列表
func (s *ticketService) List(ctx context.Context, filter *models.TicketFilter) ([]*models.Ticket, int64, error) {
	ticketList, err := s.repoManager.Ticket().List(ctx, filter)
	if err != nil {
		s.logger.Error("获取工单列表失败", zap.Error(err))
		return nil, 0, fmt.Errorf("获取工单列表失败: %w", err)
	}

	return ticketList.Tickets, ticketList.Total, nil
}

// Update 更新工单
func (s *ticketService) Update(ctx context.Context, ticket *models.Ticket) error {
	if ticket == nil {
		return fmt.Errorf("工单信息不能为空")
	}
	if ticket.ID == "" {
		return fmt.Errorf("工单ID不能为空")
	}

	// 检查工单是否存在
	exists, err := s.repoManager.Ticket().Exists(ctx, ticket.ID)
	if err != nil {
		s.logger.Error("检查工单是否存在失败", zap.Error(err), zap.String("id", ticket.ID))
		return fmt.Errorf("检查工单是否存在失败: %w", err)
	}
	if !exists {
		return fmt.Errorf("工单不存在")
	}

	// 更新工单
	err = s.repoManager.Ticket().Update(ctx, ticket)
	if err != nil {
		s.logger.Error("更新工单失败", zap.Error(err), zap.String("id", ticket.ID))
		return fmt.Errorf("更新工单失败: %w", err)
	}

	s.logger.Info("工单更新成功", zap.String("id", ticket.ID))
	return nil
}

// Delete 删除工单
func (s *ticketService) Delete(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("工单ID不能为空")
	}

	// 检查工单是否存在
	exists, err := s.repoManager.Ticket().Exists(ctx, id)
	if err != nil {
		s.logger.Error("检查工单是否存在失败", zap.Error(err), zap.String("id", id))
		return fmt.Errorf("检查工单是否存在失败: %w", err)
	}
	if !exists {
		return fmt.Errorf("工单不存在")
	}

	// 软删除工单
	err = s.repoManager.Ticket().SoftDelete(ctx, id)
	if err != nil {
		s.logger.Error("删除工单失败", zap.Error(err), zap.String("id", id))
		return fmt.Errorf("删除工单失败: %w", err)
	}

	s.logger.Info("工单删除成功", zap.String("id", id))
	return nil
}

// Assign 分配工单
func (s *ticketService) Assign(ctx context.Context, id string, userID string) error {
	if id == "" {
		return fmt.Errorf("工单ID不能为空")
	}
	if userID == "" {
		return fmt.Errorf("用户ID不能为空")
	}

	// 检查工单是否存在
	exists, err := s.repoManager.Ticket().Exists(ctx, id)
	if err != nil {
		s.logger.Error("检查工单是否存在失败", zap.Error(err), zap.String("id", id))
		return fmt.Errorf("检查工单是否存在失败: %w", err)
	}
	if !exists {
		return fmt.Errorf("工单不存在")
	}

	// 分配工单
	err = s.repoManager.Ticket().Assign(ctx, id, userID)
	if err != nil {
		s.logger.Error("分配工单失败", zap.Error(err), zap.String("id", id), zap.String("userID", userID))
		return fmt.Errorf("分配工单失败: %w", err)
	}

	s.logger.Info("工单分配成功", zap.String("id", id), zap.String("userID", userID))
	return nil
}

// Close 关闭工单
func (s *ticketService) Close(ctx context.Context, id string, userID string) error {
	if id == "" {
		return fmt.Errorf("工单ID不能为空")
	}
	if userID == "" {
		return fmt.Errorf("用户ID不能为空")
	}

	// 检查工单是否存在
	exists, err := s.repoManager.Ticket().Exists(ctx, id)
	if err != nil {
		s.logger.Error("检查工单是否存在失败", zap.Error(err), zap.String("id", id))
		return fmt.Errorf("检查工单是否存在失败: %w", err)
	}
	if !exists {
		return fmt.Errorf("工单不存在")
	}

	// 关闭工单
	err = s.repoManager.Ticket().Close(ctx, id, userID)
	if err != nil {
		s.logger.Error("关闭工单失败", zap.Error(err), zap.String("id", id), zap.String("userID", userID))
		return fmt.Errorf("关闭工单失败: %w", err)
	}

	s.logger.Info("工单关闭成功", zap.String("id", id), zap.String("userID", userID))
	return nil
}

// UpdateStatus 更新工单状态
func (s *ticketService) UpdateStatus(ctx context.Context, id string, status models.TicketStatus) error {
	if id == "" {
		return fmt.Errorf("工单ID不能为空")
	}

	// 检查工单是否存在
	exists, err := s.repoManager.Ticket().Exists(ctx, id)
	if err != nil {
		s.logger.Error("检查工单是否存在失败", zap.Error(err), zap.String("id", id))
		return fmt.Errorf("检查工单是否存在失败: %w", err)
	}
	if !exists {
		return fmt.Errorf("工单不存在")
	}

	// 更新工单状态
	err = s.repoManager.Ticket().UpdateStatus(ctx, id, status)
	if err != nil {
		s.logger.Error("更新工单状态失败", zap.Error(err), zap.String("id", id), zap.String("status", string(status)))
		return fmt.Errorf("更新工单状态失败: %w", err)
	}

	s.logger.Info("工单状态更新成功", zap.String("id", id), zap.String("status", string(status)))
	return nil
}

// generateTicketNumber 生成工单编号
func (s *ticketService) generateTicketNumber() string {
	// 使用时间戳生成工单编号，格式：TK-YYYYMMDD-HHMMSS
	now := time.Now()
	return fmt.Sprintf("TK-%s-%s", now.Format("20060102"), now.Format("150405"))
}