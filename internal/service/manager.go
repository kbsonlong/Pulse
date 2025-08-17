package service

import (
	"go.uber.org/zap"

	"pulse/internal/config"
	"pulse/internal/repository"
)

// ServiceManager 服务管理器接口
type ServiceManager interface {
	Alert() AlertService
	Rule() RuleService
	DataSource() DataSourceService
	Ticket() TicketService
	Knowledge() KnowledgeService
	User() UserService
	Auth() AuthService
	Notification() NotificationService
	Webhook() WebhookService
	Config() ConfigService
}

// serviceManager 服务管理器实现
type serviceManager struct {
	repoManager repository.RepositoryManager
	logger      *zap.Logger

	// 服务实例
	alertService        AlertService
	ruleService         RuleService
	dataSourceService   DataSourceService
	ticketService       TicketService
	knowledgeService    KnowledgeService
	userService         UserService
	authService         AuthService
	notificationService NotificationService
	webhookService      WebhookService
	configService       ConfigService
}

// NewServiceManager 创建新的服务管理器
func NewServiceManager(repoManager repository.RepositoryManager, logger *zap.Logger, cfg *config.Config) ServiceManager {
	// 初始化服务
	alertService := NewAlertService(repoManager.Alert(), repoManager.User(), logger)
	ruleService := NewRuleService(repoManager, logger)
	dataSourceService := NewDataSourceService(repoManager, logger)
	ticketService := NewTicketService(repoManager, logger)
	knowledgeService := NewKnowledgeService(repoManager, logger)
	notificationService := NewNotificationService(repoManager, logger)

	return &serviceManager{
		repoManager: repoManager,
		logger:      logger,
		alertService:        alertService,
		ruleService:         ruleService,
		dataSourceService:   dataSourceService,
		ticketService:       ticketService,
		knowledgeService:    knowledgeService,
		userService:         NewUserService(repoManager.User()),
		authService:         NewAuthService(repoManager.User(), repoManager.Auth(), cfg.JWT.Secret),
		notificationService: notificationService,
		webhookService:      NewWebhookService(repoManager, logger),
		configService:       NewConfigService(repoManager, logger),
	}
}

// Alert 获取告警服务
func (s *serviceManager) Alert() AlertService {
	return s.alertService
}

// Rule 获取规则服务
func (s *serviceManager) Rule() RuleService {
	return s.ruleService
}

// DataSource 获取数据源服务
func (s *serviceManager) DataSource() DataSourceService {
	return s.dataSourceService
}

// Ticket 获取工单服务
func (s *serviceManager) Ticket() TicketService {
	return s.ticketService
}

// Knowledge 获取知识库服务
func (s *serviceManager) Knowledge() KnowledgeService {
	return s.knowledgeService
}

// User 获取用户服务
func (s *serviceManager) User() UserService {
	return s.userService
}

// Auth 获取认证服务
func (s *serviceManager) Auth() AuthService {
	return s.authService
}

// Notification 获取通知服务
func (s *serviceManager) Notification() NotificationService {
	return s.notificationService
}

// Webhook 获取Webhook服务
func (s *serviceManager) Webhook() WebhookService {
	return s.webhookService
}

// Config 获取配置服务
func (s *serviceManager) Config() ConfigService {
	return s.configService
}