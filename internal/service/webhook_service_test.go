package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"

	"pulse/internal/models"
	"pulse/internal/repository"
)

// MockWebhookRepository 模拟Webhook仓储
type MockWebhookRepository struct {
	mock.Mock
}

func (m *MockWebhookRepository) Create(ctx context.Context, webhook *models.Webhook) error {
	args := m.Called(ctx, webhook)
	return args.Error(0)
}

func (m *MockWebhookRepository) GetByID(ctx context.Context, id string) (*models.Webhook, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*models.Webhook), args.Error(1)
}

func (m *MockWebhookRepository) Update(ctx context.Context, webhook *models.Webhook) error {
	args := m.Called(ctx, webhook)
	return args.Error(0)
}

func (m *MockWebhookRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockWebhookRepository) List(ctx context.Context, filter *models.WebhookFilter) ([]*models.Webhook, int64, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]*models.Webhook), args.Get(1).(int64), args.Error(2)
}

func (m *MockWebhookRepository) GetByName(ctx context.Context, name string) (*models.Webhook, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Webhook), args.Error(1)
}

func (m *MockWebhookRepository) UpdateStatus(ctx context.Context, id string, status models.WebhookStatus) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}

func (m *MockWebhookRepository) UpdateLastTriggered(ctx context.Context, id string, lastTriggered time.Time) error {
	args := m.Called(ctx, id, lastTriggered)
	return args.Error(0)
}

func (m *MockWebhookRepository) BatchCreate(ctx context.Context, webhooks []*models.Webhook) error {
	args := m.Called(ctx, webhooks)
	return args.Error(0)
}

func (m *MockWebhookRepository) BatchDelete(ctx context.Context, ids []string) error {
	args := m.Called(ctx, ids)
	return args.Error(0)
}

// MockRepositoryManager 模拟仓储管理器
type MockRepositoryManager struct {
	mockWebhookRepo *MockWebhookRepository
}

func (m *MockRepositoryManager) User() repository.UserRepository {
	return nil
}

func (m *MockRepositoryManager) Alert() repository.AlertRepository {
	return nil
}

func (m *MockRepositoryManager) Rule() repository.RuleRepository {
	return nil
}

func (m *MockRepositoryManager) DataSource() repository.DataSourceRepository {
	return nil
}

func (m *MockRepositoryManager) Ticket() repository.TicketRepository {
	return nil
}

func (m *MockRepositoryManager) Knowledge() repository.KnowledgeRepository {
	return nil
}

func (m *MockRepositoryManager) Permission() repository.PermissionRepository {
	return nil
}

func (m *MockRepositoryManager) Auth() repository.AuthRepository {
	return nil
}

func (m *MockRepositoryManager) Webhook() repository.WebhookRepository {
	return m.mockWebhookRepo
}

func (m *MockRepositoryManager) Notification() repository.NotificationRepository {
	return nil
}

func (m *MockRepositoryManager) BeginTx(ctx context.Context) (repository.RepositoryManager, error) {
	return m, nil
}

func (m *MockRepositoryManager) Commit() error {
	return nil
}

func (m *MockRepositoryManager) Rollback() error {
	return nil
}

func (m *MockRepositoryManager) Close() error {
	return nil
}

func setupWebhookServiceTest() (*webhookService, *MockWebhookRepository) {
	mockRepo := &MockWebhookRepository{}
	mockRepoManager := &MockRepositoryManager{
		mockWebhookRepo: mockRepo,
	}
	logger := zap.NewNop()
	service := &webhookService{
		repoManager: mockRepoManager,
		logger:      logger,
	}
	return service, mockRepo
}

func TestWebhookService_Create(t *testing.T) {
	service, mockRepo := setupWebhookServiceTest()
	ctx := context.Background()

	t.Run("成功创建Webhook", func(t *testing.T) {
		webhook := &models.Webhook{
			Name:        "Test Webhook",
			URL:         "https://example.com/webhook",
			Timeout:     30,
			RetryCount:  3,
			Status:      models.WebhookStatusActive,
			CreatedBy:   uuid.New(),
		}

		mockRepo.On("Create", ctx, webhook).Return(nil)

		err := service.Create(ctx, webhook)
		assert.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, webhook.ID)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Webhook为空时返回错误", func(t *testing.T) {
		err := service.Create(ctx, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "webhook不能为空")
	})

	t.Run("必填字段为空时返回错误", func(t *testing.T) {
		webhook := &models.Webhook{}
		err := service.Create(ctx, webhook)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "名称不能为空")
	})
}

func TestWebhookService_GetByID(t *testing.T) {
	service, mockRepo := setupWebhookServiceTest()
	ctx := context.Background()

	t.Run("成功获取Webhook", func(t *testing.T) {
		webhookID := uuid.New().String()
		expectedWebhook := &models.Webhook{
			ID:   uuid.MustParse(webhookID),
			Name: "Test Webhook",
			URL:  "https://example.com/webhook",
		}

		mockRepo.On("GetByID", ctx, webhookID).Return(expectedWebhook, nil)

		webhook, err := service.GetByID(ctx, webhookID)
		assert.NoError(t, err)
		assert.Equal(t, expectedWebhook, webhook)
		mockRepo.AssertExpectations(t)
	})

	t.Run("ID为空时返回错误", func(t *testing.T) {
		webhook, err := service.GetByID(ctx, "")
		assert.Error(t, err)
		assert.Nil(t, webhook)
		assert.Contains(t, err.Error(), "ID不能为空")
	})

	t.Run("仓储层返回错误", func(t *testing.T) {
		webhookID := uuid.New().String()
		mockRepo.On("GetByID", ctx, webhookID).Return((*models.Webhook)(nil), errors.New("数据库错误"))

		webhook, err := service.GetByID(ctx, webhookID)
		assert.Error(t, err)
		assert.Nil(t, webhook)
		mockRepo.AssertExpectations(t)
	})
}

func TestWebhookService_Update(t *testing.T) {
	service, mockRepo := setupWebhookServiceTest()
	ctx := context.Background()

	t.Run("成功更新Webhook", func(t *testing.T) {
		webhook := &models.Webhook{
			ID:          uuid.New(),
			Name:        "Updated Webhook",
			URL:         "https://updated.example.com/webhook",
			Timeout:     60,
			RetryCount:  5,
			Status:      models.WebhookStatusActive,
		}

		mockRepo.On("Update", ctx, webhook).Return(nil)

		err := service.Update(ctx, webhook)
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Webhook为空时返回错误", func(t *testing.T) {
		err := service.Update(ctx, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "webhook不能为空")
	})

	t.Run("ID为空时返回错误", func(t *testing.T) {
		webhook := &models.Webhook{
			ID: uuid.Nil,
		}
		err := service.Update(ctx, webhook)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "ID不能为空")
	})
}

func TestWebhookService_Delete(t *testing.T) {
	service, mockRepo := setupWebhookServiceTest()
	ctx := context.Background()

	t.Run("成功删除Webhook", func(t *testing.T) {
		webhookID := uuid.New().String()
		existingWebhook := &models.Webhook{
			ID:   uuid.MustParse(webhookID),
			Name: "Test Webhook",
		}

		mockRepo.On("GetByID", ctx, webhookID).Return(existingWebhook, nil)
		mockRepo.On("Delete", ctx, webhookID).Return(nil)

		err := service.Delete(ctx, webhookID)
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("ID为空时返回错误", func(t *testing.T) {
		err := service.Delete(ctx, "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "ID不能为空")
	})

	t.Run("Webhook不存在时返回错误", func(t *testing.T) {
		webhookID := uuid.New().String()
		mockRepo.On("GetByID", ctx, webhookID).Return((*models.Webhook)(nil), nil)

		err := service.Delete(ctx, webhookID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Webhook不存在")
		mockRepo.AssertExpectations(t)
	})
}

func TestWebhookService_List(t *testing.T) {
	service, mockRepo := setupWebhookServiceTest()
	ctx := context.Background()

	t.Run("成功获取Webhook列表", func(t *testing.T) {
		filter := &models.WebhookFilter{
			Page:     1,
			PageSize: 10,
		}

		expectedList := &models.WebhookList{
			Webhooks: []*models.Webhook{
				{ID: uuid.New(), Name: "Webhook 1"},
				{ID: uuid.New(), Name: "Webhook 2"},
			},
		}
		expectedTotal := int64(2)

		mockRepo.On("List", ctx, filter).Return(expectedList, expectedTotal, nil)

		list, total, err := service.List(ctx, filter)
		assert.NoError(t, err)
		assert.Equal(t, expectedList, list)
		assert.Equal(t, expectedTotal, total)
		mockRepo.AssertExpectations(t)
	})

	t.Run("filter为空时使用默认值", func(t *testing.T) {
		expectedList := &models.WebhookList{Webhooks: []*models.Webhook{}}
		expectedTotal := int64(0)

		mockRepo.On("List", ctx, mock.MatchedBy(func(f *models.WebhookFilter) bool {
			return f.Page == 1 && f.PageSize == 20
		})).Return(expectedList, expectedTotal, nil)

		list, total, err := service.List(ctx, nil)
		assert.NoError(t, err)
		assert.Equal(t, expectedList, list)
		assert.Equal(t, expectedTotal, total)
		mockRepo.AssertExpectations(t)
	})
}

func TestWebhookService_Trigger(t *testing.T) {
	service, mockRepo := setupWebhookServiceTest()
	ctx := context.Background()

	t.Run("ID为空时返回错误", func(t *testing.T) {
		err := service.Trigger(ctx, "", map[string]interface{}{"test": "data"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "ID不能为空")
	})

	t.Run("Webhook不存在时返回错误", func(t *testing.T) {
		webhookID := uuid.New().String()
		mockRepo.On("GetByID", ctx, webhookID).Return((*models.Webhook)(nil), nil)

		err := service.Trigger(ctx, webhookID, map[string]interface{}{"test": "data"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Webhook不存在")
		mockRepo.AssertExpectations(t)
	})

	t.Run("Webhook状态非活跃时返回错误", func(t *testing.T) {
		webhookID := uuid.New().String()
		webhook := &models.Webhook{
			ID:     uuid.MustParse(webhookID),
			Name:   "Test Webhook",
			Status: models.WebhookStatusInactive,
		}

		mockRepo.On("GetByID", ctx, webhookID).Return(webhook, nil)

		err := service.Trigger(ctx, webhookID, map[string]interface{}{"test": "data"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Webhook未激活")
		mockRepo.AssertExpectations(t)
	})
}