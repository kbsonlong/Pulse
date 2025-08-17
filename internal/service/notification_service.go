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

// notificationService 通知服务实现
type notificationService struct {
	repoManager repository.RepositoryManager
	logger      *zap.Logger
}

// NewNotificationService 创建通知服务实例
func NewNotificationService(repoManager repository.RepositoryManager, logger *zap.Logger) NotificationService {
	return &notificationService{
		repoManager: repoManager,
		logger:      logger,
	}
}

// Send 发送通知
func (s *notificationService) Send(ctx context.Context, notification *models.Notification) error {
	if notification == nil {
		return fmt.Errorf("通知对象不能为空")
	}

	// 验证必填字段
	if notification.Type == "" {
		return fmt.Errorf("通知类型不能为空")
	}
	if notification.Recipient == "" {
		return fmt.Errorf("接收者不能为空")
	}
	if notification.Content == "" {
		return fmt.Errorf("通知内容不能为空")
	}

	// 设置默认值
	if notification.ID == uuid.Nil {
		notification.ID = uuid.New()
	}
	if notification.Status == "" {
		notification.Status = models.NotificationStatusPending
	}
	if notification.MaxRetries == 0 {
		notification.MaxRetries = 3
	}
	notification.CreatedAt = time.Now()
	notification.UpdatedAt = time.Now()

	// 保存通知记录
	notificationRepo := s.repoManager.Notification()
	if err := notificationRepo.Create(ctx, notification); err != nil {
		s.logger.Error("保存通知记录失败", zap.Error(err), zap.String("notification_id", notification.ID.String()))
		return fmt.Errorf("保存通知记录失败: %w", err)
	}

	// 根据通知类型发送通知
	var err error
	switch notification.Type {
	case models.NotificationTypeEmail:
		err = s.sendEmail(ctx, notification)
	case models.NotificationTypeSMS:
		err = s.sendSMS(ctx, notification)
	case models.NotificationTypeDingTalk:
		err = s.sendDingTalk(ctx, notification)
	case models.NotificationTypeWeChat:
		err = s.sendWeChat(ctx, notification)
	case models.NotificationTypeSlack:
		err = s.sendSlack(ctx, notification)
	case models.NotificationTypeWebhook:
		err = s.sendWebhook(ctx, notification)
	default:
		err = fmt.Errorf("不支持的通知类型: %s", notification.Type)
	}

	// 更新通知状态
	if err != nil {
		notification.Status = models.NotificationStatusFailed
		notification.LastError = func() *string { msg := err.Error(); return &msg }()
		s.logger.Error("发送通知失败", zap.Error(err), zap.String("notification_id", notification.ID.String()))
	} else {
		notification.Status = models.NotificationStatusSent
		now := time.Now()
		notification.SentAt = &now
		s.logger.Info("通知发送成功", zap.String("notification_id", notification.ID.String()))
	}

	notification.UpdatedAt = time.Now()
	if updateErr := notificationRepo.Update(ctx, notification); updateErr != nil {
		s.logger.Error("更新通知状态失败", zap.Error(updateErr), zap.String("notification_id", notification.ID.String()))
	}

	return err
}

// GetByID 根据ID获取通知
func (s *notificationService) GetByID(ctx context.Context, id string) (*models.Notification, error) {
	if id == "" {
		return nil, fmt.Errorf("通知ID不能为空")
	}

	notificationRepo := s.repoManager.Notification()
	notification, err := notificationRepo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("获取通知失败", zap.Error(err), zap.String("notification_id", id))
		return nil, fmt.Errorf("获取通知失败: %w", err)
	}

	return notification, nil
}

// SendBatch 批量发送通知
func (s *notificationService) SendBatch(ctx context.Context, notifications []*models.Notification) error {
	if len(notifications) == 0 {
		return nil
	}

	var errors []error
	for i, notification := range notifications {
		if err := s.Send(ctx, notification); err != nil {
			s.logger.Error("批量发送通知失败", zap.Error(err), zap.Int("index", i), zap.String("notification_id", notification.ID.String()))
			errors = append(errors, fmt.Errorf("通知 %d 发送失败: %w", i, err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("批量发送中有 %d 个通知失败", len(errors))
	}

	s.logger.Info("批量通知发送完成", zap.Int("total", len(notifications)))
	return nil
}

// GetTemplates 获取通知模板列表
func (s *notificationService) GetTemplates(ctx context.Context) ([]*models.NotificationTemplate, error) {
	notificationRepo := s.repoManager.Notification()
	templates, err := notificationRepo.GetTemplates(ctx)
	if err != nil {
		s.logger.Error("获取通知模板列表失败", zap.Error(err))
		return nil, fmt.Errorf("获取通知模板列表失败: %w", err)
	}

	return templates, nil
}

// CreateTemplate 创建通知模板
func (s *notificationService) CreateTemplate(ctx context.Context, template *models.NotificationTemplate) error {
	if template == nil {
		return fmt.Errorf("通知模板对象不能为空")
	}

	// 验证必填字段
	if template.Name == "" {
		return fmt.Errorf("模板名称不能为空")
	}
	if template.Type == "" {
		return fmt.Errorf("模板类型不能为空")
	}
	if template.Content == "" {
		return fmt.Errorf("模板内容不能为空")
	}

	// 设置默认值
	if template.ID == uuid.Nil {
		template.ID = uuid.New()
	}
	template.CreatedAt = time.Now()
	template.UpdatedAt = time.Now()

	// 保存模板
	notificationRepo := s.repoManager.Notification()
	if err := notificationRepo.CreateTemplate(ctx, template); err != nil {
		s.logger.Error("创建通知模板失败", zap.Error(err), zap.String("template_name", template.Name))
		return fmt.Errorf("创建通知模板失败: %w", err)
	}

	s.logger.Info("通知模板创建成功", zap.String("template_id", template.ID.String()), zap.String("template_name", template.Name))
	return nil
}

// 私有方法：各种通知类型的具体发送实现

// sendEmail 发送邮件通知
func (s *notificationService) sendEmail(ctx context.Context, notification *models.Notification) error {
	// TODO: 集成邮件服务提供商 (如 SMTP, SendGrid, AWS SES 等)
	s.logger.Info("发送邮件通知", zap.String("recipient", notification.Recipient), zap.String("subject", notification.Subject))
	return nil // 暂时返回成功，实际需要集成邮件服务
}

// sendSMS 发送短信通知
func (s *notificationService) sendSMS(ctx context.Context, notification *models.Notification) error {
	// TODO: 集成短信服务提供商 (如 Twilio, 阿里云短信等)
	s.logger.Info("发送短信通知", zap.String("recipient", notification.Recipient))
	return nil // 暂时返回成功，实际需要集成短信服务
}

// sendDingTalk 发送钉钉通知
func (s *notificationService) sendDingTalk(ctx context.Context, notification *models.Notification) error {
	// TODO: 集成钉钉机器人API
	s.logger.Info("发送钉钉通知", zap.String("recipient", notification.Recipient))
	return nil // 暂时返回成功，实际需要集成钉钉API
}

// sendWeChat 发送微信通知
func (s *notificationService) sendWeChat(ctx context.Context, notification *models.Notification) error {
	// TODO: 集成企业微信API
	s.logger.Info("发送微信通知", zap.String("recipient", notification.Recipient))
	return nil // 暂时返回成功，实际需要集成微信API
}

// sendSlack 发送Slack通知
func (s *notificationService) sendSlack(ctx context.Context, notification *models.Notification) error {
	// TODO: 集成Slack API
	s.logger.Info("发送Slack通知", zap.String("recipient", notification.Recipient))
	return nil // 暂时返回成功，实际需要集成Slack API
}

// sendWebhook 发送Webhook通知
func (s *notificationService) sendWebhook(ctx context.Context, notification *models.Notification) error {
	// TODO: 发送HTTP请求到指定的Webhook URL
	s.logger.Info("发送Webhook通知", zap.String("recipient", notification.Recipient))
	return nil // 暂时返回成功，实际需要发送HTTP请求
}