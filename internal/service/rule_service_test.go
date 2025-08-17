package service

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"

	"pulse/internal/models"
)

// MockRuleRepository mock规则仓储
type MockRuleRepository struct {
	mock.Mock
}

func (m *MockRuleRepository) Create(ctx context.Context, rule *models.Rule) error {
	args := m.Called(ctx, rule)
	return args.Error(0)
}

func (m *MockRuleRepository) GetByID(ctx context.Context, id string) (*models.Rule, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Rule), args.Error(1)
}

func (m *MockRuleRepository) GetByName(ctx context.Context, name string) (*models.Rule, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Rule), args.Error(1)
}

func (m *MockRuleRepository) List(ctx context.Context, filter *models.RuleFilter) ([]*models.Rule, int64, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]*models.Rule), args.Get(1).(int64), args.Error(2)
}

func (m *MockRuleRepository) Update(ctx context.Context, rule *models.Rule) error {
	args := m.Called(ctx, rule)
	return args.Error(0)
}

func (m *MockRuleRepository) SoftDelete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockRuleRepository) Activate(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockRuleRepository) Deactivate(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockRuleRepository) GetActiveRules(ctx context.Context) ([]*models.Rule, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Rule), args.Error(1)
}

func (m *MockRuleRepository) TestRule(ctx context.Context, rule *models.Rule) (*models.RuleTestResult, error) {
	args := m.Called(ctx, rule)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.RuleTestResult), args.Error(1)
}

func (m *MockRuleRepository) GetByDataSourceID(ctx context.Context, dataSourceID string) ([]*models.Rule, error) {
	args := m.Called(ctx, dataSourceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Rule), args.Error(1)
}

func (m *MockRuleRepository) BatchActivate(ctx context.Context, ids []string) error {
	args := m.Called(ctx, ids)
	return args.Error(0)
}

func (m *MockRuleRepository) BatchCreate(ctx context.Context, rules []*models.Rule) error {
	args := m.Called(ctx, rules)
	return args.Error(0)
}

func (m *MockRuleRepository) BatchDeactivate(ctx context.Context, ids []string) error {
	args := m.Called(ctx, ids)
	return args.Error(0)
}

func (m *MockRuleRepository) BatchUpdate(ctx context.Context, rules []*models.Rule) error {
	args := m.Called(ctx, rules)
	return args.Error(0)
}

func (m *MockRuleRepository) Count(ctx context.Context, filter *models.RuleFilter) (int64, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockRuleRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}







// 创建测试用的规则服务
func setupRuleService() (*ruleService, *MockRepositoryManager, *MockRuleRepository) {
	mockRuleRepo := &MockRuleRepository{}
	mockRepoManager := &MockRepositoryManager{
		mockRuleRepo: mockRuleRepo,
	}
	logger := zap.NewNop()
	service := &ruleService{
		repoManager: mockRepoManager,
		logger:      logger,
	}
	return service, mockRepoManager, mockRuleRepo
}

// 创建测试用的规则对象
func createTestRule() *models.Rule {
	return &models.Rule{
		ID:                 "test-rule-id",
		DataSourceID:       "test-datasource-id",
		Name:               "Test Rule",
		Description:        "Test rule description",
		Type:               models.RuleTypeMetric,
		Severity:     models.AlertSeverityMedium,
		Expression:         "cpu_usage > 80",
		Conditions:         []models.RuleCondition{},
		Actions:            []models.RuleAction{},
		Labels:             map[string]string{"env": "test"},
		Annotations:        map[string]string{"description": "test rule"},
		EvaluationInterval: 5 * time.Minute,
		ForDuration:        1 * time.Minute,
		Enabled:            true,
		Status:             models.RuleStatusActive,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}
}

func TestRuleService_Create(t *testing.T) {
	service, _, mockRepo := setupRuleService()
	ctx := context.Background()

	t.Run("成功创建规则", func(t *testing.T) {
		rule := createTestRule()
		rule.ID = "" // 创建时ID为空

		// Mock 检查名称是否存在
		mockRepo.On("GetByName", ctx, rule.Name).Return(nil, sql.ErrNoRows).Once()
		// Mock 创建规则
		mockRepo.On("Create", ctx, mock.AnythingOfType("*models.Rule")).Return(nil).Once()

		err := service.Create(ctx, rule)
		assert.NoError(t, err)
		assert.NotEmpty(t, rule.ID)
		mockRepo.AssertExpectations(t)
	})

	t.Run("规则名称已存在", func(t *testing.T) {
		rule := createTestRule()
		existingRule := createTestRule()

		// Mock 检查名称已存在
		mockRepo.On("GetByName", ctx, rule.Name).Return(existingRule, nil).Once()

		err := service.Create(ctx, rule)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "规则名称已存在")
		mockRepo.AssertExpectations(t)
	})

	t.Run("规则验证失败", func(t *testing.T) {
		rule := createTestRule()
		rule.Name = "" // 无效的名称

		err := service.Create(ctx, rule)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "规则名称不能为空")
	})
}

func TestRuleService_GetByID(t *testing.T) {
	service, _, mockRepo := setupRuleService()
	ctx := context.Background()

	t.Run("成功获取规则", func(t *testing.T) {
		rule := createTestRule()
		mockRepo.On("GetByID", ctx, rule.ID).Return(rule, nil).Once()

		result, err := service.GetByID(ctx, rule.ID)
		assert.NoError(t, err)
		assert.Equal(t, rule, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("规则不存在", func(t *testing.T) {
		ruleID := "non-existent-id"
		mockRepo.On("GetByID", ctx, ruleID).Return(nil, sql.ErrNoRows).Once()

		result, err := service.GetByID(ctx, ruleID)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "规则不存在")
		mockRepo.AssertExpectations(t)
	})

	t.Run("ID为空", func(t *testing.T) {
		result, err := service.GetByID(ctx, "")
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "规则ID不能为空")
	})
}

func TestRuleService_List(t *testing.T) {
	service, _, mockRepo := setupRuleService()
	ctx := context.Background()

	t.Run("成功获取规则列表", func(t *testing.T) {
		rules := []*models.Rule{createTestRule()}
		filter := &models.RuleFilter{
			Page:     1,
			PageSize: 10,
		}

		mockRepo.On("List", ctx, filter).Return(rules, int64(1), nil).Once()

		result, total, err := service.List(ctx, filter)
		assert.NoError(t, err)
		assert.Equal(t, rules, result)
		assert.Equal(t, int64(1), total)
		mockRepo.AssertExpectations(t)
	})

	t.Run("使用默认分页参数", func(t *testing.T) {
		rules := []*models.Rule{}
		filter := &models.RuleFilter{}

		// 期望使用默认值
		expectedFilter := &models.RuleFilter{
			Page:     1,
			PageSize: 20,
		}

		mockRepo.On("List", ctx, expectedFilter).Return(rules, int64(0), nil).Once()

		result, total, err := service.List(ctx, filter)
		assert.NoError(t, err)
		assert.Equal(t, rules, result)
		assert.Equal(t, int64(0), total)
		mockRepo.AssertExpectations(t)
	})
}

func TestRuleService_Update(t *testing.T) {
	service, _, mockRepo := setupRuleService()
	ctx := context.Background()

	t.Run("成功更新规则", func(t *testing.T) {
		rule := createTestRule()
		existingRule := createTestRule()

		// Mock 检查规则是否存在
		mockRepo.On("GetByID", ctx, rule.ID).Return(existingRule, nil).Once()
		// Mock 检查名称冲突（没有冲突）
		mockRepo.On("GetByName", ctx, rule.Name).Return(nil, sql.ErrNoRows).Once()
		// Mock 更新规则
		mockRepo.On("Update", ctx, rule).Return(nil).Once()
		// Mock 获取更新后的规则
		mockRepo.On("GetByID", ctx, rule.ID).Return(rule, nil).Once()

		err := service.Update(ctx, rule)
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("规则不存在", func(t *testing.T) {
		rule := createTestRule()
		mockRepo.On("GetByID", ctx, rule.ID).Return(nil, sql.ErrNoRows).Once()

		err := service.Update(ctx, rule)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "规则不存在")
		mockRepo.AssertExpectations(t)
	})

	t.Run("名称与其他规则冲突", func(t *testing.T) {
		rule := createTestRule()
		existingRule := createTestRule()
		conflictRule := createTestRule()
		conflictRule.ID = "different-id"

		// Mock 检查规则是否存在
		mockRepo.On("GetByID", ctx, rule.ID).Return(existingRule, nil).Once()
		// Mock 检查名称冲突（有冲突）
		mockRepo.On("GetByName", ctx, rule.Name).Return(conflictRule, nil).Once()

		err := service.Update(ctx, rule)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "规则名称已被其他规则使用")
		mockRepo.AssertExpectations(t)
	})
}

func TestRuleService_Delete(t *testing.T) {
	service, _, mockRepo := setupRuleService()
	ctx := context.Background()

	t.Run("成功删除规则", func(t *testing.T) {
		rule := createTestRule()
		mockRepo.On("GetByID", ctx, rule.ID).Return(rule, nil).Once()
		mockRepo.On("SoftDelete", ctx, rule.ID).Return(nil).Once()

		err := service.Delete(ctx, rule.ID)
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("规则不存在", func(t *testing.T) {
		ruleID := "non-existent-id"
		mockRepo.On("GetByID", ctx, ruleID).Return(nil, sql.ErrNoRows).Once()

		err := service.Delete(ctx, ruleID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "规则不存在")
		mockRepo.AssertExpectations(t)
	})

	t.Run("ID为空", func(t *testing.T) {
		err := service.Delete(ctx, "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "规则ID不能为空")
	})
}

func TestRuleService_Enable(t *testing.T) {
	service, _, mockRepo := setupRuleService()
	ctx := context.Background()

	t.Run("成功启用规则", func(t *testing.T) {
		rule := createTestRule()
		rule.Enabled = false // 设置为禁用状态

		mockRepo.On("GetByID", ctx, rule.ID).Return(rule, nil).Once()
		mockRepo.On("Activate", ctx, rule.ID).Return(nil).Once()

		err := service.Enable(ctx, rule.ID)
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("规则已启用", func(t *testing.T) {
		rule := createTestRule()
		rule.Enabled = true // 已启用状态

		mockRepo.On("GetByID", ctx, rule.ID).Return(rule, nil).Once()

		err := service.Enable(ctx, rule.ID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "规则已启用")
		mockRepo.AssertExpectations(t)
	})

	t.Run("规则不存在", func(t *testing.T) {
		ruleID := "non-existent-id"
		mockRepo.On("GetByID", ctx, ruleID).Return(nil, sql.ErrNoRows).Once()

		err := service.Enable(ctx, ruleID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "规则不存在")
		mockRepo.AssertExpectations(t)
	})
}

func TestRuleService_Disable(t *testing.T) {
	service, _, mockRepo := setupRuleService()
	ctx := context.Background()

	t.Run("成功禁用规则", func(t *testing.T) {
		rule := createTestRule()
		rule.Enabled = true // 设置为启用状态

		mockRepo.On("GetByID", ctx, rule.ID).Return(rule, nil).Once()
		mockRepo.On("Deactivate", ctx, rule.ID).Return(nil).Once()

		err := service.Disable(ctx, rule.ID)
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("规则已禁用", func(t *testing.T) {
		rule := createTestRule()
		rule.Enabled = false // 已禁用状态

		mockRepo.On("GetByID", ctx, rule.ID).Return(rule, nil).Once()

		err := service.Disable(ctx, rule.ID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "规则已禁用")
		mockRepo.AssertExpectations(t)
	})

	t.Run("规则不存在", func(t *testing.T) {
		ruleID := "non-existent-id"
		mockRepo.On("GetByID", ctx, ruleID).Return(nil, sql.ErrNoRows).Once()

		err := service.Disable(ctx, ruleID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "规则不存在")
		mockRepo.AssertExpectations(t)
	})
}