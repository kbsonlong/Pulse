package worker

import (
	"context"
	"sync"
	"time"

	"go.uber.org/zap"

	"Pulse/internal/service"
)

// Manager Worker管理器接口
type Manager interface {
	Start(ctx context.Context) error
	Stop() error
	GetStatus() map[string]WorkerStatus
	RegisterWorker(name string, worker Worker) error
}

// WorkerStatus Worker状态
type WorkerStatus struct {
	Name      string    `json:"name"`
	Status    string    `json:"status"`
	StartTime time.Time `json:"start_time"`
	LastSeen  time.Time `json:"last_seen"`
	Error     string    `json:"error,omitempty"`
}

// Worker Worker接口
type Worker interface {
	Start(ctx context.Context) error
	Stop() error
	Name() string
	Status() WorkerStatus
}

// manager Worker管理器实现
type manager struct {
	serviceManager service.ServiceManager
	logger         *zap.Logger
	workers        map[string]Worker
	mu             sync.RWMutex
	ctx            context.Context
	cancel         context.CancelFunc
	wg             sync.WaitGroup
}

// NewManager 创建新的Worker管理器
func NewManager(serviceManager service.ServiceManager, logger *zap.Logger) Manager {
	return &manager{
		serviceManager: serviceManager,
		logger:         logger,
		workers:        make(map[string]Worker),
	}
}

// Start 启动Worker管理器
func (m *manager) Start(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.ctx, m.cancel = context.WithCancel(ctx)
	m.logger.Info("Starting worker manager")

	// 注册默认的Worker
	if err := m.registerDefaultWorkers(); err != nil {
		return err
	}

	// 启动所有Worker
	for name, worker := range m.workers {
		m.wg.Add(1)
		go func(name string, worker Worker) {
			defer m.wg.Done()
			m.logger.Info("Starting worker", zap.String("name", name))
			if err := worker.Start(m.ctx); err != nil {
				m.logger.Error("Worker failed to start", zap.String("name", name), zap.Error(err))
			}
		}(name, worker)
	}

	m.logger.Info("Worker manager started", zap.Int("workers", len(m.workers)))
	return nil
}

// Stop 停止Worker管理器
func (m *manager) Stop() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.logger.Info("Stopping worker manager")

	if m.cancel != nil {
		m.cancel()
	}

	// 停止所有Worker
	for name, worker := range m.workers {
		m.logger.Info("Stopping worker", zap.String("name", name))
		if err := worker.Stop(); err != nil {
			m.logger.Error("Worker failed to stop", zap.String("name", name), zap.Error(err))
		}
	}

	// 等待所有Worker停止
	m.wg.Wait()

	m.logger.Info("Worker manager stopped")
	return nil
}

// GetStatus 获取所有Worker状态
func (m *manager) GetStatus() map[string]WorkerStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()

	status := make(map[string]WorkerStatus)
	for name, worker := range m.workers {
		status[name] = worker.Status()
	}
	return status
}

// RegisterWorker 注册Worker
func (m *manager) RegisterWorker(name string, worker Worker) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.workers[name]; exists {
		return ErrWorkerAlreadyExists
	}

	m.workers[name] = worker
	m.logger.Info("Worker registered", zap.String("name", name))
	return nil
}

// registerDefaultWorkers 注册默认的Worker
func (m *manager) registerDefaultWorkers() error {
	// 注册通知Worker
	notificationWorker := NewNotificationWorker(m.serviceManager, m.logger)
	if err := m.RegisterWorker("notification", notificationWorker); err != nil {
		return err
	}

	// 注册告警处理Worker
	alertWorker := NewAlertWorker(m.serviceManager, m.logger)
	if err := m.RegisterWorker("alert", alertWorker); err != nil {
		return err
	}

	// 注册数据收集Worker
	collectorWorker := NewCollectorWorker(m.serviceManager, m.logger)
	if err := m.RegisterWorker("collector", collectorWorker); err != nil {
		return err
	}

	return nil
}