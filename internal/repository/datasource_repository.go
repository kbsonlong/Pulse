package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"Pulse/internal/models"
)

type dataSourceRepository struct {
	db *sqlx.DB
	tx *sqlx.Tx
}

// NewDataSourceRepository 创建数据源仓储实例
func NewDataSourceRepository(db *sqlx.DB) DataSourceRepository {
	return &dataSourceRepository{
		db: db,
	}
}

// NewDataSourceRepositoryWithTx 创建带事务的数据源仓储实例
func NewDataSourceRepositoryWithTx(tx *sqlx.Tx) DataSourceRepository {
	return &dataSourceRepository{
		tx: tx,
	}
}

// getExecutor 获取数据库执行器（事务或普通连接）
func (r *dataSourceRepository) getExecutor() sqlx.ExtContext {
	if r.tx != nil {
		return r.tx
	}
	return r.db
}

// Create 创建数据源
func (r *dataSourceRepository) Create(ctx context.Context, ds *models.DataSource) error {
	if ds.ID == "" {
		ds.ID = uuid.New().String()
	}

	now := time.Now()
	ds.CreatedAt = now
	ds.UpdatedAt = now

	if ds.Status == "" {
		ds.Status = models.DataSourceStatusActive
	}

	// 序列化配置
	configJSON, err := json.Marshal(ds.Config)
	if err != nil {
		return fmt.Errorf("序列化配置失败: %w", err)
	}

	// 序列化标签
	tagsJSON, err := json.Marshal(ds.Tags)
	if err != nil {
		return fmt.Errorf("序列化标签失败: %w", err)
	}

	query := `
		INSERT INTO data_sources (
			id, name, description, type, url, config, labels, status, enabled,
			timeout, retry_count, health_check_interval, last_health_check,
			health_status, created_by, created_at, updated_at
		) VALUES (
			:id, :name, :description, :type, :url, :config, :labels, :status, :enabled,
			:timeout, :retry_count, :health_check_interval, :last_health_check,
			:health_status, :created_by, :created_at, :updated_at
		)`

	_, err = sqlx.NamedExecContext(ctx, r.getExecutor(), query, map[string]interface{}{
		"id":                    ds.ID,
		"name":                  ds.Name,
		"description":           ds.Description,
		"type":                  ds.Type,
		"url":                   ds.Config.URL,
		"config":                string(configJSON),
		"tags":                  string(tagsJSON),
		"status":                ds.Status,
		"health_check_url":      ds.HealthCheckURL,
		"last_health_check":     ds.LastHealthCheck,
		"health_status":         ds.HealthStatus,
		"created_by":            ds.CreatedBy,
		"created_at":            ds.CreatedAt,
		"updated_at":            ds.UpdatedAt,
	})

	if err != nil {
		return fmt.Errorf("创建数据源失败: %w", err)
	}

	return nil
}

// Activate 激活数据源
func (r *dataSourceRepository) Activate(ctx context.Context, id string) error {
	query := `
		UPDATE data_sources 
		SET status = $1, updated_at = NOW() 
		WHERE id = $2 AND deleted_at IS NULL`

	_, err := r.getExecutor().ExecContext(ctx, query, models.DataSourceStatusActive, id)
	if err != nil {
		return fmt.Errorf("激活数据源失败: %w", err)
	}

	return nil
}

// Deactivate 停用数据源
func (r *dataSourceRepository) Deactivate(ctx context.Context, id string) error {
	query := `
		UPDATE data_sources 
		SET status = $1, updated_at = NOW() 
		WHERE id = $2 AND deleted_at IS NULL`

	_, err := r.getExecutor().ExecContext(ctx, query, models.DataSourceStatusInactive, id)
	if err != nil {
		return fmt.Errorf("停用数据源失败: %w", err)
	}

	return nil
}

// BatchHealthCheck 批量健康检查
func (r *dataSourceRepository) BatchHealthCheck(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return nil
	}

	// 构建占位符
	placeholders := make([]string, len(ids))
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = id
	}

	query := fmt.Sprintf(`
		UPDATE data_sources 
		SET health_status = 'checking', last_health_check = NOW(), updated_at = NOW()
		WHERE id IN (%s) AND deleted_at IS NULL
	`, strings.Join(placeholders, ", "))

	_, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("批量健康检查失败: %w", err)
	}

	return nil
}

// GetActiveCount 获取活跃数据源数量
func (r *dataSourceRepository) GetActiveCount(ctx context.Context) (int64, error) {
	var count int64
	query := `SELECT COUNT(*) FROM data_sources WHERE status = 'active' AND deleted_at IS NULL`
	err := r.db.GetContext(ctx, &count, query)
	if err != nil {
		return 0, fmt.Errorf("获取活跃数据源数量失败: %w", err)
	}
	return count, nil
}

// GetHealthyCount 获取健康数据源数量
func (r *dataSourceRepository) GetHealthyCount(ctx context.Context) (int64, error) {
	var count int64
	query := `SELECT COUNT(*) FROM data_sources WHERE health_status = 'healthy' AND deleted_at IS NULL`
	err := r.db.GetContext(ctx, &count, query)
	if err != nil {
		return 0, fmt.Errorf("获取健康数据源数量失败: %w", err)
	}
	return count, nil
}

// GetUnhealthyCount 获取不健康数据源数量
func (r *dataSourceRepository) GetUnhealthyCount(ctx context.Context) (int64, error) {
	var count int64
	query := `SELECT COUNT(*) FROM data_sources WHERE health_status != 'healthy' AND deleted_at IS NULL`
	err := r.db.GetContext(ctx, &count, query)
	if err != nil {
		return 0, fmt.Errorf("获取不健康数据源数量失败: %w", err)
	}
	return count, nil
}

// Query 执行数据源查询
func (r *dataSourceRepository) Query(ctx context.Context, id string, query *models.DataSourceQuery) (*models.DataSourceQueryResult, error) {
	// 获取数据源信息
	dataSource, err := r.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("获取数据源失败: %w", err)
	}
	
	if !dataSource.IsActive() {
		return nil, errors.New("数据源未激活")
	}
	
	if !dataSource.IsHealthy() {
		return nil, errors.New("数据源不健康")
	}
	
	// 验证查询请求
	if err := query.Validate(); err != nil {
		return nil, fmt.Errorf("查询请求验证失败: %w", err)
	}
	
	// 这里应该根据数据源类型执行实际的查询
	// 目前返回模拟结果
	result := &models.DataSourceQueryResult{
		Success:   true,
		Data:      []map[string]interface{}{},
		Columns:   []string{},
		RowCount:  0,
		QueryTime: 100 * time.Millisecond,
		Metadata:  map[string]interface{}{"executed_at": time.Now()},
	}
	
	return result, nil
}

// GetMetrics 获取数据源指标
func (r *dataSourceRepository) GetMetrics(ctx context.Context, id string) (*models.DataSourceMetrics, error) {
	// 模拟返回数据源指标
	now := time.Now()
	return &models.DataSourceMetrics{
		ConnectionCount:  10,
		QueryCount:      1000,
		ErrorCount:      5,
		AvgResponseTime: 150.5,
		LastQueryAt:     &now,
	}, nil
}

// GetStats 获取数据源统计信息
func (r *dataSourceRepository) GetStats(ctx context.Context, filter *models.DataSourceFilter) (*models.DataSourceStats, error) {
	// 获取总数
	totalQuery := `SELECT COUNT(*) FROM data_sources WHERE deleted_at IS NULL`
	var total int64
	err := r.db.GetContext(ctx, &total, totalQuery)
	if err != nil {
		return nil, fmt.Errorf("获取数据源总数失败: %w", err)
	}

	// 获取活跃数量
	activeQuery := `SELECT COUNT(*) FROM data_sources WHERE status = $1 AND deleted_at IS NULL`
	var active int64
	err = r.db.GetContext(ctx, &active, activeQuery, models.DataSourceStatusActive)
	if err != nil {
		return nil, fmt.Errorf("获取活跃数据源数量失败: %w", err)
	}

	// 获取健康数量
	healthyQuery := `SELECT COUNT(*) FROM data_sources WHERE health_status = $1 AND deleted_at IS NULL`
	var healthy int64
	err = r.db.GetContext(ctx, &healthy, healthyQuery, models.DataSourceHealthStatusHealthy)
	if err != nil {
		return nil, fmt.Errorf("获取健康数据源数量失败: %w", err)
	}

	stats := &models.DataSourceStats{
		Total:        total,
		HealthyCount: healthy,
		ByStatus: map[models.DataSourceStatus]int64{
			models.DataSourceStatusActive: active,
		},
		ByType: make(map[models.DataSourceType]int64),
	}

	return stats, nil
}

// GetByID 根据ID获取数据源
func (r *dataSourceRepository) GetByID(ctx context.Context, id string) (*models.DataSource, error) {
	var ds models.DataSource
	var configJSON, tagsJSON string

	query := `
		SELECT id, name, description, type, config, tags, status,
		       health_check_url, last_health_check, health_status, 
		       created_by, created_at, updated_at
		FROM data_sources 
		WHERE id = $1 AND deleted_at IS NULL`

	err := r.getExecutor().QueryRowxContext(ctx, query, id).Scan(
		&ds.ID, &ds.Name, &ds.Description, &ds.Type, &configJSON, &tagsJSON,
		&ds.Status, &ds.HealthCheckURL, &ds.LastHealthCheck, &ds.HealthStatus, 
		&ds.CreatedBy, &ds.CreatedAt, &ds.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("数据源不存在")
		}
		return nil, fmt.Errorf("获取数据源失败: %w", err)
	}

	// 反序列化配置
	if configJSON != "" {
		err = json.Unmarshal([]byte(configJSON), &ds.Config)
		if err != nil {
			return nil, fmt.Errorf("反序列化配置失败: %w", err)
		}
	}

	// 反序列化标签
	if tagsJSON != "" {
		err = json.Unmarshal([]byte(tagsJSON), &ds.Tags)
		if err != nil {
			return nil, fmt.Errorf("反序列化标签失败: %w", err)
		}
	}

	return &ds, nil
}

// GetByName 根据名称获取数据源
func (r *dataSourceRepository) GetByName(ctx context.Context, name string) (*models.DataSource, error) {
	var ds models.DataSource
	var configJSON, tagsJSON string

	query := `
		SELECT id, name, description, type, config, tags, status,
		       health_check_url, last_health_check, health_status, 
		       created_by, created_at, updated_at
		FROM data_sources 
		WHERE name = $1 AND deleted_at IS NULL`

	err := r.getExecutor().QueryRowxContext(ctx, query, name).Scan(
		&ds.ID, &ds.Name, &ds.Description, &ds.Type, &configJSON, &tagsJSON,
		&ds.Status, &ds.HealthCheckURL, &ds.LastHealthCheck, &ds.HealthStatus, 
		&ds.CreatedBy, &ds.CreatedAt, &ds.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("数据源不存在")
		}
		return nil, fmt.Errorf("获取数据源失败: %w", err)
	}

	// 反序列化配置
	if configJSON != "" {
		err = json.Unmarshal([]byte(configJSON), &ds.Config)
		if err != nil {
			return nil, fmt.Errorf("反序列化配置失败: %w", err)
		}
	}

	// 反序列化标签
	if tagsJSON != "" {
		err = json.Unmarshal([]byte(tagsJSON), &ds.Tags)
		if err != nil {
			return nil, fmt.Errorf("反序列化标签失败: %w", err)
		}
	}

	return &ds, nil
}

// Update 更新数据源
func (r *dataSourceRepository) Update(ctx context.Context, ds *models.DataSource) error {
	ds.UpdatedAt = time.Now()

	// 序列化配置
	configJSON, err := json.Marshal(ds.Config)
	if err != nil {
		return fmt.Errorf("序列化配置失败: %w", err)
	}

	// 序列化标签
	tagsJSON, err := json.Marshal(ds.Tags)
	if err != nil {
		return fmt.Errorf("序列化标签失败: %w", err)
	}

	query := `
		UPDATE data_sources SET 
			name = :name,
			description = :description,
			type = :type,
			config = :config,
			tags = :tags,
			status = :status,
			health_check_url = :health_check_url,
			updated_at = :updated_at
		WHERE id = :id AND deleted_at IS NULL`

	_, err = r.db.NamedExecContext(ctx, query, map[string]interface{}{
		"id":                ds.ID,
		"name":              ds.Name,
		"description":       ds.Description,
		"type":              ds.Type,
		"config":            string(configJSON),
		"tags":              string(tagsJSON),
		"status":            ds.Status,
		"health_check_url":  ds.HealthCheckURL,
		"updated_at":        ds.UpdatedAt,
	})

	if err != nil {
		return fmt.Errorf("更新数据源失败: %w", err)
	}

	return nil
}

// Delete 硬删除数据源
func (r *dataSourceRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM data_sources WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("删除数据源失败: %w", err)
	}
	return nil
}

// SoftDelete 软删除数据源
func (r *dataSourceRepository) SoftDelete(ctx context.Context, id string) error {
	now := time.Now()
	query := `
		UPDATE data_sources SET 
			deleted_at = $1,
			updated_at = $1
		WHERE id = $2 AND deleted_at IS NULL`

	_, err := r.db.ExecContext(ctx, query, now, id)
	if err != nil {
		return fmt.Errorf("软删除数据源失败: %w", err)
	}
	return nil
}

// List 获取数据源列表
func (r *dataSourceRepository) List(ctx context.Context, filter *models.DataSourceFilter) (*models.DataSourceList, error) {
	var conditions []string
	var args []interface{}
	argIndex := 1

	conditions = append(conditions, "deleted_at IS NULL")

	if filter != nil {
		if filter.Type != nil {
			conditions = append(conditions, fmt.Sprintf("type = $%d", argIndex))
			args = append(args, *filter.Type)
			argIndex++
		}

		if filter.Status != nil {
			conditions = append(conditions, fmt.Sprintf("status = $%d", argIndex))
			args = append(args, *filter.Status)
			argIndex++
		}

		// Enabled字段已移除，跳过此过滤条件

		if filter.HealthStatus != nil {
			conditions = append(conditions, fmt.Sprintf("health_status = $%d", argIndex))
			args = append(args, *filter.HealthStatus)
			argIndex++
		}

		if filter.Keyword != nil && *filter.Keyword != "" {
			conditions = append(conditions, fmt.Sprintf("(name ILIKE $%d OR description ILIKE $%d)", argIndex, argIndex))
			args = append(args, "%"+*filter.Keyword+"%")
			argIndex++
		}
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// 获取总数
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM data_sources %s", whereClause)
	var total int64
	err := r.db.GetContext(ctx, &total, countQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("获取数据源总数失败: %w", err)
	}

	// 构建查询
	query := fmt.Sprintf(`
		SELECT id, name, description, type, config, tags, status,
		       health_check_url, last_health_check, health_status, 
		       created_by, created_at, updated_at
		FROM data_sources %s
		ORDER BY created_at DESC`, whereClause)

	// 添加分页
	if filter != nil && filter.Page > 0 && filter.PageSize > 0 {
		offset := (filter.Page - 1) * filter.PageSize
		query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
		args = append(args, filter.PageSize, offset)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("查询数据源列表失败: %w", err)
	}
	defer rows.Close()

	var dataSources []*models.DataSource
	for rows.Next() {
		var ds models.DataSource
		var configJSON, tagsJSON string

		err := rows.Scan(
			&ds.ID, &ds.Name, &ds.Description, &ds.Type, &configJSON, &tagsJSON,
			&ds.Status, &ds.HealthCheckURL, &ds.LastHealthCheck, &ds.HealthStatus, 
			&ds.CreatedBy, &ds.CreatedAt, &ds.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("扫描数据源数据失败: %w", err)
		}

		// 反序列化配置
		if configJSON != "" {
			err = json.Unmarshal([]byte(configJSON), &ds.Config)
			if err != nil {
				return nil, fmt.Errorf("反序列化配置失败: %w", err)
			}
		}

		// 反序列化标签
		if tagsJSON != "" {
			err = json.Unmarshal([]byte(tagsJSON), &ds.Tags)
			if err != nil {
				return nil, fmt.Errorf("反序列化标签失败: %w", err)
			}
		}

		dataSources = append(dataSources, &ds)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("遍历数据源数据失败: %w", err)
	}

	// 计算分页信息
	var totalPages int64 = 1
	if filter != nil && filter.PageSize > 0 {
		totalPages = (total + int64(filter.PageSize) - 1) / int64(filter.PageSize)
	}

	page := 1
	pageSize := len(dataSources)
	if filter != nil {
		if filter.Page > 0 {
			page = filter.Page
		}
		if filter.PageSize > 0 {
			pageSize = filter.PageSize
		}
	}

	return &models.DataSourceList{
		DataSources: dataSources,
		Total:       total,
		Page:        page,
		PageSize:    pageSize,
		TotalPages:  int(totalPages),
	}, nil
}

// Count 获取数据源总数
func (r *dataSourceRepository) Count(ctx context.Context, filter *models.DataSourceFilter) (int64, error) {
	var conditions []string
	var args []interface{}
	argIndex := 1

	conditions = append(conditions, "deleted_at IS NULL")

	if filter != nil {
		if filter.Type != nil {
			conditions = append(conditions, fmt.Sprintf("type = $%d", argIndex))
			args = append(args, *filter.Type)
			argIndex++
		}

		if filter.Status != nil {
			conditions = append(conditions, fmt.Sprintf("status = $%d", argIndex))
			args = append(args, *filter.Status)
			argIndex++
		}

		// Enabled字段已移除，跳过此过滤条件
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	query := fmt.Sprintf("SELECT COUNT(*) FROM data_sources %s", whereClause)
	var count int64
	err := r.db.GetContext(ctx, &count, query, args...)
	if err != nil {
		return 0, fmt.Errorf("获取数据源总数失败: %w", err)
	}

	return count, nil
}

// Exists 检查数据源是否存在
func (r *dataSourceRepository) Exists(ctx context.Context, id string) (bool, error) {
	var count int
	query := `SELECT COUNT(*) FROM data_sources WHERE id = $1 AND deleted_at IS NULL`
	err := r.db.GetContext(ctx, &count, query, id)
	if err != nil {
		return false, fmt.Errorf("检查数据源是否存在失败: %w", err)
	}
	return count > 0, nil
}

// ExistsByName 检查数据源名称是否存在
func (r *dataSourceRepository) ExistsByName(ctx context.Context, name string) (bool, error) {
	var count int
	query := `SELECT COUNT(*) FROM data_sources WHERE name = $1 AND deleted_at IS NULL`
	err := r.db.GetContext(ctx, &count, query, name)
	if err != nil {
		return false, fmt.Errorf("检查数据源名称是否存在失败: %w", err)
	}
	return count > 0, nil
}

// TestConnection 测试数据源连接
func (r *dataSourceRepository) TestConnection(ctx context.Context, ds *models.DataSource) (*models.DataSourceTestResult, error) {
	start := time.Now()
	
	// 这里应该根据不同的数据源类型进行实际的连接测试
	// 模拟连接测试逻辑
	time.Sleep(50 * time.Millisecond)
	
	version := "1.0.0"
	// 目前返回一个模拟结果
	result := &models.DataSourceTestResult{
		Success:      true,
		Message:      "连接成功",
		ResponseTime: time.Since(start),
		Version:      &version,
		Metadata:     make(map[string]interface{}),
	}

	// 注意：这个方法只是测试连接，不更新数据库记录
	// 如果需要更新健康状态，应该使用 UpdateHealthStatus 方法

	return result, nil
}

// UpdateHealthStatus 更新数据源健康状态
func (r *dataSourceRepository) UpdateHealthStatus(ctx context.Context, id string, isHealthy bool, errorMsg string) error {
	var healthStatus models.DataSourceHealthStatus
	if isHealthy {
		healthStatus = models.DataSourceHealthStatusHealthy
	} else {
		healthStatus = models.DataSourceHealthStatusUnhealthy
	}
	
	query := `
		UPDATE data_sources 
		SET health_status = $1, last_error = $2, updated_at = CURRENT_TIMESTAMP
		WHERE id = $3 AND deleted_at IS NULL
	`
	
	_, err := r.db.ExecContext(ctx, query, healthStatus, errorMsg, id)
	if err != nil {
		return fmt.Errorf("更新数据源健康状态失败: %w", err)
	}
	
	return nil
}

// UpdateLastHealthCheck 更新最后健康检查时间
func (r *dataSourceRepository) UpdateLastHealthCheck(ctx context.Context, id string, checkTime time.Time) error {
	query := `
		UPDATE data_sources 
		SET last_health_check = $1, updated_at = CURRENT_TIMESTAMP
		WHERE id = $2 AND deleted_at IS NULL
	`
	
	_, err := r.db.ExecContext(ctx, query, checkTime, id)
	if err != nil {
		return fmt.Errorf("更新最后健康检查时间失败: %w", err)
	}
	
	return nil
}

// UpdateMetrics 更新数据源指标
func (r *dataSourceRepository) UpdateMetrics(ctx context.Context, id string, metrics *models.DataSourceMetrics) error {
	metricsJSON, err := json.Marshal(metrics)
	if err != nil {
		return fmt.Errorf("序列化指标数据失败: %w", err)
	}
	
	query := `
		UPDATE data_sources 
		SET metrics = $1, updated_at = CURRENT_TIMESTAMP
		WHERE id = $2 AND deleted_at IS NULL
	`
	
	_, err = r.db.ExecContext(ctx, query, string(metricsJSON), id)
	if err != nil {
		return fmt.Errorf("更新数据源指标失败: %w", err)
	}
	
	return nil
}

// Enable 启用数据源
func (r *dataSourceRepository) Enable(ctx context.Context, id string) error {
	now := time.Now()
	query := `
		UPDATE data_sources SET 
			enabled = true,
			status = $1,
			updated_at = $2
		WHERE id = $3 AND deleted_at IS NULL`

	_, err := r.db.ExecContext(ctx, query, models.DataSourceStatusActive, now, id)
	if err != nil {
		return fmt.Errorf("启用数据源失败: %w", err)
	}
	return nil
}

// Disable 禁用数据源
func (r *dataSourceRepository) Disable(ctx context.Context, id string) error {
	now := time.Now()
	query := `
		UPDATE data_sources SET 
			enabled = false,
			status = $1,
			updated_at = $2
		WHERE id = $3 AND deleted_at IS NULL`

	_, err := r.db.ExecContext(ctx, query, models.DataSourceStatusInactive, now, id)
	if err != nil {
		return fmt.Errorf("禁用数据源失败: %w", err)
	}
	return nil
}

// GetByType 根据类型获取数据源列表
func (r *dataSourceRepository) GetByType(ctx context.Context, dsType models.DataSourceType) ([]*models.DataSource, error) {
	query := `
		SELECT id, name, description, type, status, config, tags, version,
		       health_check_url, health_status, last_health_check, error_message,
		       metrics, created_by, updated_by, created_at, updated_at
		FROM data_sources 
		WHERE type = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, dsType)
	if err != nil {
		return nil, fmt.Errorf("根据类型获取数据源失败: %w", err)
	}
	defer rows.Close()

	var dataSources []*models.DataSource
	for rows.Next() {
		var ds models.DataSource
		var configJSON, tagsJSON, metricsJSON string

		err := rows.Scan(
			&ds.ID, &ds.Name, &ds.Description, &ds.Type, &ds.Status, &configJSON, &tagsJSON,
			&ds.Version, &ds.HealthCheckURL, &ds.HealthStatus, &ds.LastHealthCheck,
			&ds.ErrorMessage, &metricsJSON, &ds.CreatedBy, &ds.UpdatedBy, &ds.CreatedAt, &ds.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("扫描数据源数据失败: %w", err)
		}

		// 反序列化配置
		if configJSON != "" {
			err = json.Unmarshal([]byte(configJSON), &ds.Config)
			if err != nil {
				return nil, fmt.Errorf("反序列化配置失败: %w", err)
			}
		}

		// 反序列化标签
		if tagsJSON != "" {
			err = json.Unmarshal([]byte(tagsJSON), &ds.Tags)
			if err != nil {
				return nil, fmt.Errorf("反序列化标签失败: %w", err)
			}
		}

		// 反序列化指标
		if metricsJSON != "" {
			err = json.Unmarshal([]byte(metricsJSON), &ds.Metrics)
			if err != nil {
				return nil, fmt.Errorf("反序列化指标失败: %w", err)
			}
		}

		dataSources = append(dataSources, &ds)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("遍历数据源数据失败: %w", err)
	}

	return dataSources, nil
}

// GetActiveDataSources 获取活跃数据源列表
func (r *dataSourceRepository) GetActiveDataSources(ctx context.Context) ([]*models.DataSource, error) {
	query := `
		SELECT id, name, description, type, status, config, tags, version,
		       health_check_url, health_status, last_health_check, error_message,
		       metrics, created_by, updated_by, created_at, updated_at
		FROM data_sources 
		WHERE status = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, models.DataSourceStatusActive)
	if err != nil {
		return nil, fmt.Errorf("获取活跃数据源失败: %w", err)
	}
	defer rows.Close()

	var dataSources []*models.DataSource
	for rows.Next() {
		var ds models.DataSource
		var configJSON, tagsJSON, metricsJSON string

		err := rows.Scan(
			&ds.ID, &ds.Name, &ds.Description, &ds.Type, &ds.Status, &configJSON, &tagsJSON,
			&ds.Version, &ds.HealthCheckURL, &ds.HealthStatus, &ds.LastHealthCheck,
			&ds.ErrorMessage, &metricsJSON, &ds.CreatedBy, &ds.UpdatedBy, &ds.CreatedAt, &ds.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("扫描数据源数据失败: %w", err)
		}

		// 反序列化配置
		if configJSON != "" {
			err = json.Unmarshal([]byte(configJSON), &ds.Config)
			if err != nil {
				return nil, fmt.Errorf("反序列化配置失败: %w", err)
			}
		}

		// 反序列化标签
		if tagsJSON != "" {
			err = json.Unmarshal([]byte(tagsJSON), &ds.Tags)
			if err != nil {
				return nil, fmt.Errorf("反序列化标签失败: %w", err)
			}
		}

		// 反序列化指标
		if metricsJSON != "" {
			err = json.Unmarshal([]byte(metricsJSON), &ds.Metrics)
			if err != nil {
				return nil, fmt.Errorf("反序列化指标失败: %w", err)
			}
		}

		dataSources = append(dataSources, &ds)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("遍历数据源数据失败: %w", err)
	}

	return dataSources, nil
}

// BatchCreate 批量创建数据源
func (r *dataSourceRepository) BatchCreate(ctx context.Context, dataSources []*models.DataSource) error {
	if len(dataSources) == 0 {
		return nil
	}

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("开始事务失败: %w", err)
	}
	defer tx.Rollback()

	for _, ds := range dataSources {
		if ds.ID == "" {
			ds.ID = uuid.New().String()
		}

		now := time.Now()
		ds.CreatedAt = now
		ds.UpdatedAt = now

		if ds.Status == "" {
			ds.Status = models.DataSourceStatusActive
		}

		// 序列化配置、标签和指标
		configJSON, err := json.Marshal(ds.Config)
		if err != nil {
			return fmt.Errorf("序列化配置失败: %w", err)
		}

		tagsJSON, err := json.Marshal(ds.Tags)
		if err != nil {
			return fmt.Errorf("序列化标签失败: %w", err)
		}

		var metricsJSON []byte
		if ds.Metrics != nil {
			metricsJSON, err = json.Marshal(ds.Metrics)
			if err != nil {
				return fmt.Errorf("序列化指标失败: %w", err)
			}
		}

		query := `
			INSERT INTO data_sources (
				id, name, description, type, status, config, tags, version,
				health_check_url, health_status, last_health_check, error_message,
				metrics, created_by, updated_by, created_at, updated_at
			) VALUES (
				$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17
			)`

		_, err = tx.ExecContext(ctx, query,
			ds.ID, ds.Name, ds.Description, ds.Type, ds.Status, string(configJSON), string(tagsJSON),
			ds.Version, ds.HealthCheckURL, ds.HealthStatus, ds.LastHealthCheck,
			ds.ErrorMessage, string(metricsJSON), ds.CreatedBy, ds.UpdatedBy, ds.CreatedAt, ds.UpdatedAt,
		)
		if err != nil {
			return fmt.Errorf("批量创建数据源失败: %w", err)
		}
	}

	return tx.Commit()
}

// BatchUpdate 批量更新数据源
func (r *dataSourceRepository) BatchUpdate(ctx context.Context, dataSources []*models.DataSource) error {
	if len(dataSources) == 0 {
		return nil
	}

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("开始事务失败: %w", err)
	}
	defer tx.Rollback()

	for _, ds := range dataSources {
		ds.UpdatedAt = time.Now()

		// 序列化配置、标签和指标
		configJSON, err := json.Marshal(ds.Config)
		if err != nil {
			return fmt.Errorf("序列化配置失败: %w", err)
		}

		tagsJSON, err := json.Marshal(ds.Tags)
		if err != nil {
			return fmt.Errorf("序列化标签失败: %w", err)
		}

		var metricsJSON []byte
		if ds.Metrics != nil {
			metricsJSON, err = json.Marshal(ds.Metrics)
			if err != nil {
				return fmt.Errorf("序列化指标失败: %w", err)
			}
		}

		query := `
			UPDATE data_sources SET 
				name = $1,
				description = $2,
				type = $3,
				config = $4,
				tags = $5,
				status = $6,
				version = $7,
				health_check_url = $8,
				health_status = $9,
				error_message = $10,
				metrics = $11,
				updated_by = $12,
				updated_at = $13
			WHERE id = $14 AND deleted_at IS NULL`

		_, err = tx.ExecContext(ctx, query,
			ds.Name, ds.Description, ds.Type, string(configJSON), string(tagsJSON),
			ds.Status, ds.Version, ds.HealthCheckURL, ds.HealthStatus,
			ds.ErrorMessage, string(metricsJSON), ds.UpdatedBy, ds.UpdatedAt, ds.ID,
		)
		if err != nil {
			return fmt.Errorf("批量更新数据源失败: %w", err)
		}
	}

	return tx.Commit()
}

// BatchDelete 批量删除数据源
func (r *dataSourceRepository) BatchDelete(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return nil
	}

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("开始事务失败: %w", err)
	}
	defer tx.Rollback()

	now := time.Now()
	for _, id := range ids {
		query := `
			UPDATE data_sources SET 
				deleted_at = $1,
				updated_at = $1
			WHERE id = $2 AND deleted_at IS NULL`

		_, err := tx.ExecContext(ctx, query, now, id)
		if err != nil {
			return fmt.Errorf("批量删除数据源失败: %w", err)
		}
	}

	return tx.Commit()
}