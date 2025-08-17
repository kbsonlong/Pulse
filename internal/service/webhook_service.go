package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"pulse/internal/models"
	"pulse/internal/repository"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// webhookService Webhook服务实现
type webhookService struct {
	repoManager repository.RepositoryManager
	logger      *zap.Logger
}

// NewWebhookService 创建Webhook服务实例
func NewWebhookService(repoManager repository.RepositoryManager, logger *zap.Logger) WebhookService {
	return &webhookService{
		repoManager: repoManager,
		logger:      logger,
	}
}

// Create 创建Webhook
func (s *webhookService) Create(ctx context.Context, webhook *models.Webhook) error {
	if webhook == nil {
		return fmt.Errorf("webhook不能为空")
	}

	// 验证必填字段
	if webhook.Name == "" {
		return fmt.Errorf("webhook名称不能为空")
	}
	if webhook.URL == "" {
		return fmt.Errorf("webhook URL不能为空")
	}

	// 调用仓储层创建Webhook
	err := s.repoManager.Webhook().Create(ctx, webhook)
	if err != nil {
		s.logger.Error("创建Webhook失败", zap.Error(err), zap.String("name", webhook.Name))
		return fmt.Errorf("创建Webhook失败: %w", err)
	}

	s.logger.Info("Webhook创建成功", zap.String("id", webhook.ID.String()), zap.String("name", webhook.Name))
	return nil
}

// GetByID 根据ID获取Webhook
func (s *webhookService) GetByID(ctx context.Context, id string) (*models.Webhook, error) {
	if id == "" {
		return nil, fmt.Errorf("webhook ID不能为空")
	}

	// 调用仓储层获取Webhook
	webhook, err := s.repoManager.Webhook().GetByID(ctx, id)
	if err != nil {
		s.logger.Error("获取Webhook失败", zap.Error(err), zap.String("id", id))
		return nil, fmt.Errorf("获取Webhook失败: %w", err)
	}

	return webhook, nil
}

// List 获取Webhook列表
func (s *webhookService) List(ctx context.Context, filter *models.WebhookFilter) ([]*models.Webhook, int64, error) {
	if filter == nil {
		filter = &models.WebhookFilter{}
	}

	// 设置默认分页参数
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 {
		filter.PageSize = 20
	}

	// 调用仓储层获取Webhook列表
	webhooks, err := s.repoManager.Webhook().List(ctx, filter)
	if err != nil {
		s.logger.Error("获取Webhook列表失败", zap.Error(err))
		return nil, 0, fmt.Errorf("获取Webhook列表失败: %w", err)
	}

	// 获取总数
	total, err := s.repoManager.Webhook().Count(ctx, filter)
	if err != nil {
		s.logger.Error("获取Webhook总数失败", zap.Error(err))
		return nil, 0, fmt.Errorf("获取Webhook总数失败: %w", err)
	}

	return webhooks.Webhooks, total, nil
}

// Update 更新Webhook
func (s *webhookService) Update(ctx context.Context, webhook *models.Webhook) error {
	if webhook == nil {
		return fmt.Errorf("webhook不能为空")
	}
	if webhook.ID == uuid.Nil {
		return fmt.Errorf("webhook ID不能为空")
	}

	// 验证必填字段
	if webhook.Name == "" {
		return fmt.Errorf("webhook名称不能为空")
	}
	if webhook.URL == "" {
		return fmt.Errorf("webhook URL不能为空")
	}

	// 调用仓储层更新Webhook
	err := s.repoManager.Webhook().Update(ctx, webhook)
	if err != nil {
		s.logger.Error("更新Webhook失败", zap.Error(err), zap.String("id", webhook.ID.String()))
		return fmt.Errorf("更新Webhook失败: %w", err)
	}

	s.logger.Info("Webhook更新成功", zap.String("id", webhook.ID.String()), zap.String("name", webhook.Name))
	return nil
}

// Delete 删除Webhook
func (s *webhookService) Delete(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("webhook ID不能为空")
	}

	// 检查Webhook是否存在
	_, err := s.repoManager.Webhook().GetByID(ctx, id)
	if err != nil {
		s.logger.Error("获取Webhook失败", zap.Error(err), zap.String("id", id))
		return fmt.Errorf("Webhook不存在或获取失败: %w", err)
	}

	// 调用仓储层删除Webhook
	err = s.repoManager.Webhook().Delete(ctx, id)
	if err != nil {
		s.logger.Error("删除Webhook失败", zap.Error(err), zap.String("id", id))
		return fmt.Errorf("删除Webhook失败: %w", err)
	}

	s.logger.Info("Webhook删除成功", zap.String("id", id))
	return nil
}

// Trigger 触发Webhook
func (s *webhookService) Trigger(ctx context.Context, id string, payload interface{}) error {
	if id == "" {
		return fmt.Errorf("webhook ID不能为空")
	}

	// 获取Webhook配置
	webhook, err := s.repoManager.Webhook().GetByID(ctx, id)
	if err != nil {
		s.logger.Error("获取Webhook配置失败", zap.Error(err), zap.String("id", id))
		return fmt.Errorf("获取Webhook配置失败: %w", err)
	}

	// 检查Webhook状态
	if webhook.Status != models.WebhookStatusActive {
		s.logger.Warn("Webhook未激活", zap.String("id", id))
		return fmt.Errorf("Webhook未激活")
	}

	// 序列化payload
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		s.logger.Error("序列化payload失败", zap.Error(err), zap.String("id", id))
		return fmt.Errorf("序列化payload失败: %w", err)
	}

	// 执行Webhook调用
	return s.executeWebhook(ctx, webhook, payloadBytes)
}

// executeWebhook 执行Webhook HTTP调用
func (s *webhookService) executeWebhook(ctx context.Context, webhook *models.Webhook, payload []byte) error {
	start := time.Now()
	var lastErr error

	// 重试逻辑
	for attempt := 0; attempt <= webhook.RetryCount; attempt++ {
		// 创建HTTP请求
		req, err := http.NewRequestWithContext(ctx, "POST", webhook.URL, bytes.NewBuffer(payload))
		if err != nil {
			s.logger.Error("创建HTTP请求失败", zap.Error(err), zap.String("webhook_id", webhook.ID.String()))
			lastErr = err
			continue
		}

		// 设置请求头
		req.Header.Set("Content-Type", "application/json")
		if webhook.Secret != nil && *webhook.Secret != "" {
			req.Header.Set("X-Webhook-Secret", *webhook.Secret)
		}

		// 添加自定义头部
		if webhook.Headers != nil {
			for key, value := range webhook.Headers {
				req.Header.Set(key, value)
			}
		}

		// 创建HTTP客户端
		client := &http.Client{
			Timeout: time.Duration(webhook.Timeout) * time.Second,
		}

		// 执行请求
		resp, err := client.Do(req)
		if err != nil {
			s.logger.Warn("Webhook调用失败", zap.Error(err), zap.String("webhook_id", webhook.ID.String()), zap.Int("attempt", attempt+1))
			lastErr = err
			// 记录失败日志
			s.logWebhookCall(ctx, webhook.ID.String(), payload, 0, "", err.Error(), time.Since(start))
			continue
		}
		defer resp.Body.Close()

		// 读取响应
		var responseBody bytes.Buffer
		_, err = responseBody.ReadFrom(resp.Body)
		if err != nil {
			s.logger.Warn("读取响应失败", zap.Error(err), zap.String("webhook_id", webhook.ID.String()))
		}

		// 检查响应状态码
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			// 成功
			s.logger.Info("Webhook调用成功", zap.String("webhook_id", webhook.ID.String()), zap.Int("status_code", resp.StatusCode))
			// 记录成功日志
			s.logWebhookCall(ctx, webhook.ID.String(), payload, resp.StatusCode, responseBody.String(), "", time.Since(start))
			// 更新成功计数
			s.repoManager.Webhook().IncrementSuccessCount(ctx, webhook.ID.String())
			s.repoManager.Webhook().UpdateLastTriggered(ctx, webhook.ID.String())
			return nil
		} else {
			// 失败
			lastErr = fmt.Errorf("HTTP状态码: %d", resp.StatusCode)
			s.logger.Warn("Webhook调用返回错误状态码", zap.String("webhook_id", webhook.ID.String()), zap.Int("status_code", resp.StatusCode))
			// 记录失败日志
			s.logWebhookCall(ctx, webhook.ID.String(), payload, resp.StatusCode, responseBody.String(), lastErr.Error(), time.Since(start))
		}

		// 如果不是最后一次尝试，等待一段时间再重试
		if attempt < webhook.RetryCount {
			time.Sleep(time.Duration(attempt+1) * time.Second)
		}
	}

	// 所有重试都失败了
	s.logger.Error("Webhook调用最终失败", zap.String("webhook_id", webhook.ID.String()), zap.Error(lastErr))
	// 更新失败计数
	s.repoManager.Webhook().IncrementFailureCount(ctx, webhook.ID.String())
	return fmt.Errorf("Webhook调用失败: %w", lastErr)
}

// logWebhookCall 记录Webhook调用日志
func (s *webhookService) logWebhookCall(ctx context.Context, webhookID string, payload []byte, statusCode int, response string, errorMsg string, duration time.Duration) {
	webhookUUID, err := uuid.Parse(webhookID)
	if err != nil {
		s.logger.Error("解析Webhook ID失败", zap.Error(err), zap.String("webhook_id", webhookID))
		return
	}

	log := &models.WebhookLog{
		ID:         uuid.New(),
		WebhookID:  webhookUUID,
		Payload:    string(payload),
		StatusCode: statusCode,
		Duration:   duration.Milliseconds(),
		CreatedAt:  time.Now(),
	}

	if response != "" {
		log.Response = &response
	}
	if errorMsg != "" {
		log.Error = &errorMsg
	}

	if err := s.repoManager.Webhook().CreateLog(ctx, log); err != nil {
		s.logger.Error("记录Webhook日志失败", zap.Error(err))
	}
}