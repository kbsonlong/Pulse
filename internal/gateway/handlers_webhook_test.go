package gateway

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"

	"pulse/internal/models"
)

// MockWebhookService is a mock implementation of WebhookService
type MockWebhookService struct {
	mock.Mock
}

func (m *MockWebhookService) Create(ctx context.Context, webhook *models.Webhook) error {
	args := m.Called(ctx, webhook)
	return args.Error(0)
}

func (m *MockWebhookService) GetByID(ctx context.Context, id string) (*models.Webhook, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Webhook), args.Error(1)
}

func (m *MockWebhookService) Update(ctx context.Context, webhook *models.Webhook) error {
	args := m.Called(ctx, webhook)
	return args.Error(0)
}

func (m *MockWebhookService) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockWebhookService) List(ctx context.Context, filter *models.WebhookFilter) (*models.WebhookList, int64, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).(*models.WebhookList), args.Get(1).(int64), args.Error(2)
}

func (m *MockWebhookService) Trigger(ctx context.Context, id string, payload map[string]interface{}) error {
	args := m.Called(ctx, id, payload)
	return args.Error(0)
}

func setupWebhookHandlerTest() (*gin.Engine, *MockWebhookService) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	mockService := &MockWebhookService{}
	logger := zap.NewNop()

	handler := &Handler{
		webhookService: mockService,
		logger:         logger,
	}

	// 注册路由
	v1 := router.Group("/api/v1")
	{
		v1.GET("/webhooks", handler.listWebhooks)
		v1.POST("/webhooks", handler.createWebhook)
		v1.GET("/webhooks/:id", handler.getWebhook)
		v1.PUT("/webhooks/:id", handler.updateWebhook)
		v1.DELETE("/webhooks/:id", handler.deleteWebhook)
		v1.POST("/webhooks/:id/trigger", handler.triggerWebhook)
	}

	return router, mockService
}

func TestListWebhooks(t *testing.T) {
	router, mockService := setupWebhookHandlerTest()

	t.Run("成功获取Webhook列表", func(t *testing.T) {
		expectedList := &models.WebhookList{
			Webhooks: []*models.Webhook{
				{
					ID:   uuid.New(),
					Name: "Test Webhook 1",
					URL:  "https://example1.com/webhook",
				},
				{
					ID:   uuid.New(),
					Name: "Test Webhook 2",
					URL:  "https://example2.com/webhook",
				},
			},
		}
		expectedTotal := int64(2)

		mockService.On("List", mock.Anything, mock.AnythingOfType("*models.WebhookFilter")).Return(expectedList, expectedTotal, nil)

		req := httptest.NewRequest("GET", "/api/v1/webhooks?page=1&page_size=10", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "success", response["status"])
		assert.NotNil(t, response["data"])

		mockService.AssertExpectations(t)
	})

	t.Run("服务层返回错误", func(t *testing.T) {
		mockService.On("List", mock.Anything, mock.AnythingOfType("*models.WebhookFilter")).Return((*models.WebhookList)(nil), int64(0), errors.New("服务错误"))

		req := httptest.NewRequest("GET", "/api/v1/webhooks", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "error", response["status"])

		mockService.AssertExpectations(t)
	})
}

func TestCreateWebhook(t *testing.T) {
	router, mockService := setupWebhookHandlerTest()

	t.Run("成功创建Webhook", func(t *testing.T) {
		webhookData := map[string]interface{}{
			"name":         "Test Webhook",
			"url":          "https://example.com/webhook",
			"method":       "POST",
			"timeout":      30,
			"retry_count":  3,
			"retry_delay":  5,
			"status":       "active",
			"description":  "Test webhook description",
		}

		mockService.On("Create", mock.Anything, mock.AnythingOfType("*models.Webhook")).Return(nil)

		body, _ := json.Marshal(webhookData)
		req := httptest.NewRequest("POST", "/api/v1/webhooks", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "success", response["status"])

		mockService.AssertExpectations(t)
	})

	t.Run("请求体格式错误", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/v1/webhooks", bytes.NewBuffer([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "error", response["status"])
	})

	t.Run("服务层返回错误", func(t *testing.T) {
		webhookData := map[string]interface{}{
			"name": "Test Webhook",
			"url":  "https://example.com/webhook",
		}

		mockService.On("Create", mock.Anything, mock.AnythingOfType("*models.Webhook")).Return(errors.New("创建失败"))

		body, _ := json.Marshal(webhookData)
		req := httptest.NewRequest("POST", "/api/v1/webhooks", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "error", response["status"])

		mockService.AssertExpectations(t)
	})
}

func TestGetWebhook(t *testing.T) {
	router, mockService := setupWebhookHandlerTest()

	t.Run("成功获取Webhook", func(t *testing.T) {
		webhookID := uuid.New()
		expectedWebhook := &models.Webhook{
			ID:          webhookID,
			Name:        "Test Webhook",
			URL:         "https://example.com/webhook",
			Method:      "POST",
			Timeout:     30,
			RetryCount:  3,
			RetryDelay:  5,
			Status:      models.WebhookStatusActive,
			CreatedAt:   time.Now(),
		}

		mockService.On("GetByID", mock.Anything, webhookID.String()).Return(expectedWebhook, nil)

		req := httptest.NewRequest("GET", "/api/v1/webhooks/"+webhookID.String(), nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "success", response["status"])
		assert.NotNil(t, response["data"])

		mockService.AssertExpectations(t)
	})

	t.Run("Webhook不存在", func(t *testing.T) {
		webhookID := uuid.New()
		mockService.On("GetByID", mock.Anything, webhookID.String()).Return(nil, nil)

		req := httptest.NewRequest("GET", "/api/v1/webhooks/"+webhookID.String(), nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "error", response["status"])

		mockService.AssertExpectations(t)
	})

	t.Run("无效的UUID格式", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/webhooks/invalid-uuid", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "error", response["status"])
	})
}

func TestUpdateWebhook(t *testing.T) {
	router, mockService := setupWebhookHandlerTest()

	t.Run("成功更新Webhook", func(t *testing.T) {
		webhookID := uuid.New()
		updateData := map[string]interface{}{
			"name":        "Updated Webhook",
			"url":         "https://updated.example.com/webhook",
			"method":      "PUT",
			"timeout":     60,
			"retry_count": 5,
			"retry_delay": 10,
			"status":      "active",
		}

		mockService.On("Update", mock.Anything, mock.AnythingOfType("*models.Webhook")).Return(nil)

		body, _ := json.Marshal(updateData)
		req := httptest.NewRequest("PUT", "/api/v1/webhooks/"+webhookID.String(), bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "success", response["status"])

		mockService.AssertExpectations(t)
	})

	t.Run("无效的UUID格式", func(t *testing.T) {
		updateData := map[string]interface{}{"name": "Updated Webhook"}
		body, _ := json.Marshal(updateData)
		req := httptest.NewRequest("PUT", "/api/v1/webhooks/invalid-uuid", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "error", response["status"])
	})
}

func TestDeleteWebhook(t *testing.T) {
	router, mockService := setupWebhookHandlerTest()

	t.Run("成功删除Webhook", func(t *testing.T) {
		webhookID := uuid.New()
		mockService.On("Delete", mock.Anything, webhookID.String()).Return(nil)

		req := httptest.NewRequest("DELETE", "/api/v1/webhooks/"+webhookID.String(), nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "success", response["status"])

		mockService.AssertExpectations(t)
	})

	t.Run("无效的UUID格式", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/api/v1/webhooks/invalid-uuid", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "error", response["status"])
	})
}

func TestTriggerWebhook(t *testing.T) {
	router, mockService := setupWebhookHandlerTest()

	t.Run("成功触发Webhook", func(t *testing.T) {
		webhookID := uuid.New()
		payload := map[string]interface{}{
			"message": "test message",
			"data":    map[string]interface{}{"key": "value"},
		}

		mockService.On("Trigger", mock.Anything, webhookID.String(), payload).Return(nil)

		body, _ := json.Marshal(map[string]interface{}{"payload": payload})
		req := httptest.NewRequest("POST", "/api/v1/webhooks/"+webhookID.String()+"/trigger", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "success", response["status"])

		mockService.AssertExpectations(t)
	})

	t.Run("无payload时使用空对象", func(t *testing.T) {
		webhookID := uuid.New()
		emptyPayload := map[string]interface{}{}

		mockService.On("Trigger", mock.Anything, webhookID.String(), emptyPayload).Return(nil)

		body, _ := json.Marshal(map[string]interface{}{})
		req := httptest.NewRequest("POST", "/api/v1/webhooks/"+webhookID.String()+"/trigger", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "success", response["status"])

		mockService.AssertExpectations(t)
	})

	t.Run("无效的UUID格式", func(t *testing.T) {
		body, _ := json.Marshal(map[string]interface{}{"payload": map[string]interface{}{}})
		req := httptest.NewRequest("POST", "/api/v1/webhooks/invalid-uuid/trigger", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "error", response["status"])
	})

	t.Run("触发失败", func(t *testing.T) {
		webhookID := uuid.New()
		payload := map[string]interface{}{"test": "data"}

		mockService.On("Trigger", mock.Anything, webhookID.String(), payload).Return(errors.New("触发失败"))

		body, _ := json.Marshal(map[string]interface{}{"payload": payload})
		req := httptest.NewRequest("POST", "/api/v1/webhooks/"+webhookID.String()+"/trigger", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "error", response["status"])

		mockService.AssertExpectations(t)
	})
}