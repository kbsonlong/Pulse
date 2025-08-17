package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"pulse/internal/models"
	"pulse/internal/repository"
)

// alertService 告警服务实现
type alertService struct {
	alertRepo repository.AlertRepository
	userRepo  repository.UserRepository
	logger    *zap.Logger
}

// NewAlertService 创建告警服务实例
func NewAlertService(alertRepo repository.AlertRepository, userRepo repository.UserRepository, logger *zap.Logger) AlertService {
	return &alertService{
		alertRepo: alertRepo,
		userRepo:  userRepo,
		logger:    logger,
	}
}

// Create 创建告警
func (s *alertService) Create(ctx context.Context, alert *models.Alert) error {
	// 验证告警数据
	if err := alert.Validate(); err != nil {
		s.logger.Error("告警数据验证失败", zap.Error(err))
		return fmt.Errorf("告警数据验证失败: %w", err)
	}

	// 生成告警ID
	if alert.ID == "" {
		alert.ID = uuid.New().String()
	}

	// 设置创建时间
	now := time.Now()
	alert.CreatedAt = now
	alert.UpdatedAt = now

	// 设置默认状态
	if alert.Status == "" {
		alert.Status = models.AlertStatusFiring
	}

	// 设置开始时间
	if alert.StartsAt.IsZero() {
		alert.StartsAt = now
	}

	// 设置最后评估时间
	alert.LastEvalAt = now
	alert.EvalCount = 1

	// 生成指纹
	if alert.Fingerprint == "" {
		alert.Fingerprint = s.generateFingerprint(alert)
	}

	// 创建告警
	if err := s.alertRepo.Create(ctx, alert); err != nil {
		s.logger.Error("创建告警失败", zap.Error(err), zap.String("alert_id", alert.ID))
		return fmt.Errorf("创建告警失败: %w", err)
	}

	// 记录历史
	history := &models.AlertHistory{
		ID:        uuid.New().String(),
		AlertID:   alert.ID,
		Action:    "created",
		NewValue:  s.alertToMap(alert),
		CreatedAt: now,
	}

	if err := s.alertRepo.AddHistory(ctx, history); err != nil {
		s.logger.Warn("记录告警历史失败", zap.Error(err), zap.String("alert_id", alert.ID))
	}

	s.logger.Info("告警创建成功", zap.String("alert_id", alert.ID), zap.String("name", alert.Name))
	return nil
}

// GetByID 根据ID获取告警
func (s *alertService) GetByID(ctx context.Context, id string) (*models.Alert, error) {
	if id == "" {
		return nil, fmt.Errorf("告警ID不能为空")
	}

	alert, err := s.alertRepo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("获取告警失败", zap.Error(err), zap.String("alert_id", id))
		return nil, fmt.Errorf("获取告警失败: %w", err)
	}

	return alert, nil
}

// List 获取告警列表
func (s *alertService) List(ctx context.Context, filter *models.AlertFilter) ([]*models.Alert, int64, error) {
	// 获取告警列表
	alertList, err := s.alertRepo.List(ctx, filter)
	if err != nil {
		s.logger.Error("获取告警列表失败", zap.Error(err))
		return nil, 0, fmt.Errorf("获取告警列表失败: %w", err)
	}

	return alertList.Alerts, alertList.Total, nil
}

// Update 更新告警
func (s *alertService) Update(ctx context.Context, alert *models.Alert) error {
	if alert.ID == "" {
		return fmt.Errorf("告警ID不能为空")
	}

	// 获取原始告警
	originalAlert, err := s.alertRepo.GetByID(ctx, alert.ID)
	if err != nil {
		s.logger.Error("获取原始告警失败", zap.Error(err), zap.String("alert_id", alert.ID))
		return fmt.Errorf("获取原始告警失败: %w", err)
	}

	// 验证告警数据
	if err := alert.Validate(); err != nil {
		s.logger.Error("告警数据验证失败", zap.Error(err))
		return fmt.Errorf("告警数据验证失败: %w", err)
	}

	// 更新时间
	alert.UpdatedAt = time.Now()

	// 更新告警
	if err := s.alertRepo.Update(ctx, alert); err != nil {
		s.logger.Error("更新告警失败", zap.Error(err), zap.String("alert_id", alert.ID))
		return fmt.Errorf("更新告警失败: %w", err)
	}

	// 记录历史
	history := &models.AlertHistory{
		ID:        uuid.New().String(),
		AlertID:   alert.ID,
		Action:    "updated",
		OldValue:  s.alertToMap(originalAlert),
		NewValue:  s.alertToMap(alert),
		CreatedAt: time.Now(),
	}

	if err := s.alertRepo.AddHistory(ctx, history); err != nil {
		s.logger.Warn("记录告警历史失败", zap.Error(err), zap.String("alert_id", alert.ID))
	}

	s.logger.Info("告警更新成功", zap.String("alert_id", alert.ID))
	return nil
}

// Delete 删除告警
func (s *alertService) Delete(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("告警ID不能为空")
	}

	// 获取告警信息
	alert, err := s.alertRepo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("获取告警失败", zap.Error(err), zap.String("alert_id", id))
		return fmt.Errorf("获取告警失败: %w", err)
	}

	// 软删除告警
	if err := s.alertRepo.SoftDelete(ctx, id); err != nil {
		s.logger.Error("删除告警失败", zap.Error(err), zap.String("alert_id", id))
		return fmt.Errorf("删除告警失败: %w", err)
	}

	// 记录历史
	history := &models.AlertHistory{
		ID:        uuid.New().String(),
		AlertID:   id,
		Action:    "deleted",
		OldValue:  s.alertToMap(alert),
		CreatedAt: time.Now(),
	}

	if err := s.alertRepo.AddHistory(ctx, history); err != nil {
		s.logger.Warn("记录告警历史失败", zap.Error(err), zap.String("alert_id", id))
	}

	s.logger.Info("告警删除成功", zap.String("alert_id", id))
	return nil
}

// Acknowledge 确认告警
func (s *alertService) Acknowledge(ctx context.Context, id string, userID string) error {
	if id == "" {
		return fmt.Errorf("告警ID不能为空")
	}
	if userID == "" {
		return fmt.Errorf("用户ID不能为空")
	}

	// 验证用户存在
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		s.logger.Error("获取用户失败", zap.Error(err), zap.String("user_id", userID))
		return fmt.Errorf("获取用户失败: %w", err)
	}

	// 获取告警信息
	alert, err := s.alertRepo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("获取告警失败", zap.Error(err), zap.String("alert_id", id))
		return fmt.Errorf("获取告警失败: %w", err)
	}

	// 检查告警状态
	if alert.Status != models.AlertStatusFiring {
		return fmt.Errorf("只能确认正在触发的告警")
	}

	// 确认告警
	if err := s.alertRepo.Acknowledge(ctx, id, userID, nil); err != nil {
		s.logger.Error("确认告警失败", zap.Error(err), zap.String("alert_id", id), zap.String("user_id", userID))
		return fmt.Errorf("确认告警失败: %w", err)
	}

	// 记录历史
	history := &models.AlertHistory{
		ID:        uuid.New().String(),
		AlertID:   id,
		Action:    "acknowledged",
		UserID:    &userID,
		CreatedAt: time.Now(),
	}

	if err := s.alertRepo.AddHistory(ctx, history); err != nil {
		s.logger.Warn("记录告警历史失败", zap.Error(err), zap.String("alert_id", id))
	}

	s.logger.Info("告警确认成功", zap.String("alert_id", id), zap.String("user_id", userID), zap.String("username", user.Username))
	return nil
}

// Resolve 解决告警
func (s *alertService) Resolve(ctx context.Context, id string, userID string) error {
	if id == "" {
		return fmt.Errorf("告警ID不能为空")
	}
	if userID == "" {
		return fmt.Errorf("用户ID不能为空")
	}

	// 验证用户存在
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		s.logger.Error("获取用户失败", zap.Error(err), zap.String("user_id", userID))
		return fmt.Errorf("获取用户失败: %w", err)
	}

	// 获取告警信息
	alert, err := s.alertRepo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("获取告警失败", zap.Error(err), zap.String("alert_id", id))
		return fmt.Errorf("获取告警失败: %w", err)
	}

	// 检查告警状态
	if alert.Status == models.AlertStatusResolved {
		return fmt.Errorf("告警已经解决")
	}

	// 解决告警
	if err := s.alertRepo.Resolve(ctx, id, userID, nil); err != nil {
		s.logger.Error("解决告警失败", zap.Error(err), zap.String("alert_id", id), zap.String("user_id", userID))
		return fmt.Errorf("解决告警失败: %w", err)
	}

	// 记录历史
	history := &models.AlertHistory{
		ID:        uuid.New().String(),
		AlertID:   id,
		Action:    "resolved",
		UserID:    &userID,
		CreatedAt: time.Now(),
	}

	if err := s.alertRepo.AddHistory(ctx, history); err != nil {
		s.logger.Warn("记录告警历史失败", zap.Error(err), zap.String("alert_id", id))
	}

	s.logger.Info("告警解决成功", zap.String("alert_id", id), zap.String("user_id", userID), zap.String("username", user.Username))
	return nil
}

// 辅助方法

// generateFingerprint 生成告警指纹
func (s *alertService) generateFingerprint(alert *models.Alert) string {
	// 简单的指纹生成逻辑，实际项目中可能需要更复杂的算法
	return fmt.Sprintf("%s-%s-%s", alert.Name, alert.DataSourceID, alert.Expression)
}

// alertToMap 将告警转换为map用于历史记录
func (s *alertService) alertToMap(alert *models.Alert) map[string]interface{} {
	return map[string]interface{}{
		"id":              alert.ID,
		"name":            alert.Name,
		"description":     alert.Description,
		"severity":        alert.Severity,
		"status":          alert.Status,
		"source":          alert.Source,
		"data_source_id":  alert.DataSourceID,
		"expression":      alert.Expression,
		"value":           alert.Value,
		"threshold":       alert.Threshold,
		"starts_at":       alert.StartsAt,
		"ends_at":         alert.EndsAt,
		"acked_by":        alert.AckedBy,
		"acked_at":        alert.AckedAt,
		"resolved_by":     alert.ResolvedBy,
		"resolved_at":     alert.ResolvedAt,
		"updated_at":      alert.UpdatedAt,
	}
}