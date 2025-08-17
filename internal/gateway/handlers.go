package gateway

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"pulse/internal/models"
)

// 健康检查处理函数
func (g *Gateway) healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"service": "alert-management-platform",
		"version": "1.0.0",
	})
}

// 状态检查处理函数
func (g *Gateway) statusCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "ok",
		"timestamp": "2024-01-01T00:00:00Z",
		"uptime":    "0s",
	})
}

// 认证相关处理函数
func (g *Gateway) login(c *gin.Context) {
	// TODO: 实现登录逻辑
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented yet"})
}

func (g *Gateway) logout(c *gin.Context) {
	// TODO: 实现登出逻辑
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented yet"})
}

func (g *Gateway) refreshToken(c *gin.Context) {
	// TODO: 实现刷新令牌逻辑
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented yet"})
}

func (g *Gateway) resetPassword(c *gin.Context) {
	// TODO: 实现重置密码逻辑
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented yet"})
}

// 告警相关处理函数
func (g *Gateway) listAlerts(c *gin.Context) {
	// 解析查询参数
	filter := &models.AlertFilter{
		Page:     1,
		PageSize: 20,
	}

	// 解析分页参数
	if pageStr := c.Query("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil && page > 0 {
			filter.Page = page
		}
	}

	if pageSizeStr := c.Query("page_size"); pageSizeStr != "" {
		if pageSize, err := strconv.Atoi(pageSizeStr); err == nil && pageSize > 0 && pageSize <= 100 {
			filter.PageSize = pageSize
		}
	}

	// 解析过滤参数
	if ruleID := c.Query("rule_id"); ruleID != "" {
		filter.RuleID = &ruleID
	}

	if dataSourceID := c.Query("data_source_id"); dataSourceID != "" {
		filter.DataSourceID = &dataSourceID
	}

	if severityStr := c.Query("severity"); severityStr != "" {
		severity := models.AlertSeverity(severityStr)
		if severity.IsValid() {
			filter.Severity = &severity
		}
	}

	if statusStr := c.Query("status"); statusStr != "" {
		status := models.AlertStatus(statusStr)
		if status.IsValid() {
			filter.Status = &status
		}
	}

	if sourceStr := c.Query("source"); sourceStr != "" {
		source := models.AlertSource(sourceStr)
		if source.IsValid() {
			filter.Source = &source
		}
	}

	if keyword := c.Query("keyword"); keyword != "" {
		filter.Keyword = &keyword
	}

	// 解析时间范围
	if startTimeStr := c.Query("start_time"); startTimeStr != "" {
		if startTime, err := time.Parse(time.RFC3339, startTimeStr); err == nil {
			filter.StartTime = &startTime
		}
	}

	if endTimeStr := c.Query("end_time"); endTimeStr != "" {
		if endTime, err := time.Parse(time.RFC3339, endTimeStr); err == nil {
			filter.EndTime = &endTime
		}
	}

	// 解析排序参数
	if sortBy := c.Query("sort_by"); sortBy != "" {
		filter.SortBy = &sortBy
	}

	if sortOrder := c.Query("sort_order"); sortOrder != "" {
		if sortOrder == "asc" || sortOrder == "desc" {
			filter.SortOrder = &sortOrder
		}
	}

	// 调用告警服务获取列表
	alerts, total, err := g.serviceManager.Alert().List(c.Request.Context(), filter)
	if err != nil {
		g.logger.WithError(err).Error("获取告警列表失败")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "获取告警列表失败",
			"message": err.Error(),
		})
		return
	}

	// 计算总页数
	totalPages := int(total) / filter.PageSize
	if int(total)%filter.PageSize > 0 {
		totalPages++
	}

	// 构造响应
	response := &models.AlertList{
		Alerts:     alerts,
		Total:      total,
		Page:       filter.Page,
		PageSize:   filter.PageSize,
		TotalPages: totalPages,
	}

	c.JSON(http.StatusOK, response)
}

func (g *Gateway) createAlert(c *gin.Context) {
	// 解析请求体
	var req models.AlertCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		g.logger.WithError(err).Error("解析创建告警请求失败")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "请求参数无效",
			"message": err.Error(),
		})
		return
	}

	// 验证请求数据
	if err := req.Validate(); err != nil {
		g.logger.WithError(err).Error("创建告警请求验证失败")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "请求数据验证失败",
			"message": err.Error(),
		})
		return
	}

	// 构造告警对象
	alert := &models.Alert{
		RuleID:       req.RuleID,
		DataSourceID: req.DataSourceID,
		Name:         req.Name,
		Description:  req.Description,
		Severity:     req.Severity,
		Source:       req.Source,
		Labels:       req.Labels,
		Annotations:  req.Annotations,
		Value:        req.Value,
		Threshold:    req.Threshold,
		Expression:   req.Expression,
		GeneratorURL: req.GeneratorURL,
	}

	// 设置开始时间
	if req.StartsAt != nil {
		alert.StartsAt = *req.StartsAt
	} else {
		alert.StartsAt = time.Now()
	}

	// 调用告警服务创建告警
	if err := g.serviceManager.Alert().Create(c.Request.Context(), alert); err != nil {
		g.logger.WithError(err).Error("创建告警失败")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "创建告警失败",
			"message": err.Error(),
		})
		return
	}

	g.logger.WithField("alert_id", alert.ID).Info("告警创建成功")
	c.JSON(http.StatusCreated, alert)
}

func (g *Gateway) getAlert(c *gin.Context) {
	// 获取告警ID
	alertID := c.Param("id")
	if alertID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "告警ID不能为空",
			"message": "请提供有效的告警ID",
		})
		return
	}

	// 调用告警服务获取告警详情
	alert, err := g.serviceManager.Alert().GetByID(c.Request.Context(), alertID)
	if err != nil {
		g.logger.WithError(err).WithField("alert_id", alertID).Error("获取告警详情失败")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "获取告警详情失败",
			"message": err.Error(),
		})
		return
	}

	// 检查告警是否存在
	if alert == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "告警不存在",
			"message": "指定的告警ID不存在",
		})
		return
	}

	c.JSON(http.StatusOK, alert)
}

func (g *Gateway) updateAlert(c *gin.Context) {
	// 获取告警ID
	alertID := c.Param("id")
	if alertID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "告警ID不能为空",
			"message": "请提供有效的告警ID",
		})
		return
	}

	// 解析请求体
	var req models.AlertUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		g.logger.WithError(err).Error("解析更新告警请求失败")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "请求参数无效",
			"message": err.Error(),
		})
		return
	}

	// 验证请求数据
	if err := req.Validate(); err != nil {
		g.logger.WithError(err).Error("更新告警请求验证失败")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "请求数据验证失败",
			"message": err.Error(),
		})
		return
	}

	// 获取当前告警
	alert, err := g.serviceManager.Alert().GetByID(c.Request.Context(), alertID)
	if err != nil {
		g.logger.WithError(err).WithField("alert_id", alertID).Error("获取告警失败")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "获取告警失败",
			"message": err.Error(),
		})
		return
	}

	if alert == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "告警不存在",
			"message": "指定的告警ID不存在",
		})
		return
	}

	// 更新告警字段
	if req.Name != nil {
		alert.Name = *req.Name
	}
	if req.Description != nil {
		alert.Description = *req.Description
	}
	if req.Severity != nil {
		alert.Severity = *req.Severity
	}
	if req.Labels != nil {
		alert.Labels = req.Labels
	}
	if req.Annotations != nil {
		alert.Annotations = req.Annotations
	}
	if req.Value != nil {
		alert.Value = req.Value
	}
	if req.Threshold != nil {
		alert.Threshold = req.Threshold
	}
	if req.Expression != nil {
		alert.Expression = *req.Expression
	}
	if req.GeneratorURL != nil {
		alert.GeneratorURL = req.GeneratorURL
	}

	// 调用告警服务更新告警
	if err := g.serviceManager.Alert().Update(c.Request.Context(), alert); err != nil {
		g.logger.WithError(err).WithField("alert_id", alertID).Error("更新告警失败")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "更新告警失败",
			"message": err.Error(),
		})
		return
	}

	g.logger.WithField("alert_id", alertID).Info("告警更新成功")
	c.JSON(http.StatusOK, alert)
}

func (g *Gateway) deleteAlert(c *gin.Context) {
	// 获取告警ID
	alertID := c.Param("id")
	if alertID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "告警ID不能为空",
			"message": "请提供有效的告警ID",
		})
		return
	}

	// 检查告警是否存在
	alert, err := g.serviceManager.Alert().GetByID(c.Request.Context(), alertID)
	if err != nil {
		g.logger.WithError(err).WithField("alert_id", alertID).Error("获取告警失败")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "获取告警失败",
			"message": err.Error(),
		})
		return
	}

	if alert == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "告警不存在",
			"message": "指定的告警ID不存在",
		})
		return
	}

	// 调用告警服务删除告警
	if err := g.serviceManager.Alert().Delete(c.Request.Context(), alertID); err != nil {
		g.logger.WithError(err).WithField("alert_id", alertID).Error("删除告警失败")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "删除告警失败",
			"message": err.Error(),
		})
		return
	}

	g.logger.WithField("alert_id", alertID).Info("告警删除成功")
	c.JSON(http.StatusOK, gin.H{
		"message": "告警删除成功",
		"id":      alertID,
	})
}

func (g *Gateway) acknowledgeAlert(c *gin.Context) {
	// 获取告警ID
	alertID := c.Param("id")
	if alertID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "告警ID不能为空",
			"message": "请提供有效的告警ID",
		})
		return
	}

	// 解析请求体
	var req models.AlertAckRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		g.logger.WithError(err).Error("解析确认告警请求失败")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "请求参数无效",
			"message": err.Error(),
		})
		return
	}

	// 验证请求数据
	if err := req.Validate(); err != nil {
		g.logger.WithError(err).Error("确认告警请求验证失败")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "请求数据验证失败",
			"message": err.Error(),
		})
		return
	}

	// 调用告警服务确认告警
	if err := g.serviceManager.Alert().Acknowledge(c.Request.Context(), alertID, req.UserID); err != nil {
		g.logger.WithError(err).WithField("alert_id", alertID).Error("确认告警失败")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "确认告警失败",
			"message": err.Error(),
		})
		return
	}

	g.logger.WithField("alert_id", alertID).WithField("user_id", req.UserID).Info("告警确认成功")
	c.JSON(http.StatusOK, gin.H{
		"message": "告警确认成功",
		"id":      alertID,
	})
}

func (g *Gateway) resolveAlert(c *gin.Context) {
	// 获取告警ID
	alertID := c.Param("id")
	if alertID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "告警ID不能为空",
			"message": "请提供有效的告警ID",
		})
		return
	}

	// 解析请求体
	var req models.AlertResolveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		g.logger.WithError(err).Error("解析解决告警请求失败")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "请求参数无效",
			"message": err.Error(),
		})
		return
	}

	// 验证请求数据
	if err := req.Validate(); err != nil {
		g.logger.WithError(err).Error("解决告警请求验证失败")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "请求数据验证失败",
			"message": err.Error(),
		})
		return
	}

	// 调用告警服务解决告警
	if err := g.serviceManager.Alert().Resolve(c.Request.Context(), alertID, req.UserID); err != nil {
		g.logger.WithError(err).WithField("alert_id", alertID).Error("解决告警失败")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "解决告警失败",
			"message": err.Error(),
		})
		return
	}

	g.logger.WithField("alert_id", alertID).WithField("user_id", req.UserID).Info("告警解决成功")
	c.JSON(http.StatusOK, gin.H{
		"message": "告警解决成功",
		"id":      alertID,
	})
}

// 规则相关处理函数
func (g *Gateway) listRules(c *gin.Context) {
	// 解析查询参数
	filter := &models.RuleFilter{
		Page:     1,
		PageSize: 20,
	}

	// 解析分页参数
	if pageStr := c.Query("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil && page > 0 {
			filter.Page = page
		}
	}

	if pageSizeStr := c.Query("page_size"); pageSizeStr != "" {
		if pageSize, err := strconv.Atoi(pageSizeStr); err == nil && pageSize > 0 && pageSize <= 100 {
			filter.PageSize = pageSize
		}
	}

	// 解析过滤参数
	if dataSourceID := c.Query("data_source_id"); dataSourceID != "" {
		filter.DataSourceID = &dataSourceID
	}

	if keyword := c.Query("keyword"); keyword != "" {
		filter.Keyword = &keyword
	}

	if enabledStr := c.Query("enabled"); enabledStr != "" {
		if enabled, err := strconv.ParseBool(enabledStr); err == nil {
			filter.Enabled = &enabled
		}
	}

	// 解析排序参数
	if sortBy := c.Query("sort_by"); sortBy != "" {
		filter.SortBy = &sortBy
	}

	if sortOrder := c.Query("sort_order"); sortOrder != "" {
		if sortOrder == "asc" || sortOrder == "desc" {
			filter.SortOrder = &sortOrder
		}
	}

	// 调用规则服务获取列表
	rules, total, err := g.serviceManager.Rule().List(c.Request.Context(), filter)
	if err != nil {
		g.logger.WithError(err).Error("获取规则列表失败")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "获取规则列表失败",
			"message": err.Error(),
		})
		return
	}

	// 计算总页数
	totalPages := int(total) / filter.PageSize
	if int(total)%filter.PageSize > 0 {
		totalPages++
	}

	// 构造响应
	response := gin.H{
		"rules":       rules,
		"total":       total,
		"page":        filter.Page,
		"page_size":   filter.PageSize,
		"total_pages": totalPages,
	}

	c.JSON(http.StatusOK, response)
}

func (g *Gateway) createRule(c *gin.Context) {
	// 解析请求体
	var req models.RuleCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		g.logger.WithError(err).Error("解析创建规则请求失败")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "请求参数无效",
			"message": err.Error(),
		})
		return
	}

	// 验证请求数据
	if err := req.Validate(); err != nil {
		g.logger.WithError(err).Error("创建规则请求验证失败")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "请求数据验证失败",
			"message": err.Error(),
		})
		return
	}

	// 构造规则对象
	rule := &models.Rule{
		ID:           uuid.New().String(),
		DataSourceID: req.DataSourceID,
		Name:         req.Name,
		Description:  req.Description,
		Expression:   req.Expression,
		Conditions:   req.Conditions,
		Actions:      req.Actions,
		Severity:     req.Severity,
		Enabled:      true, // 默认启用
		Labels:       req.Labels,
		Annotations:  req.Annotations,
	}

	// 调用规则服务创建规则
	if err := g.serviceManager.Rule().Create(c.Request.Context(), rule); err != nil {
		g.logger.WithError(err).Error("创建规则失败")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "创建规则失败",
			"message": err.Error(),
		})
		return
	}

	g.logger.WithField("rule_id", rule.ID).WithField("rule_name", rule.Name).Info("规则创建成功")
	c.JSON(http.StatusCreated, gin.H{
		"message": "规则创建成功",
		"data":    rule,
	})
}

func (g *Gateway) getRule(c *gin.Context) {
	// 获取规则ID
	ruleID := c.Param("id")
	if ruleID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "规则ID不能为空",
			"message": "请提供有效的规则ID",
		})
		return
	}

	// 调用规则服务获取规则
	rule, err := g.serviceManager.Rule().GetByID(c.Request.Context(), ruleID)
	if err != nil {
		g.logger.WithError(err).WithField("rule_id", ruleID).Error("获取规则失败")
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "获取规则失败",
			"message": err.Error(),
		})
		return
	}

	g.logger.WithField("rule_id", ruleID).Info("获取规则成功")
	c.JSON(http.StatusOK, gin.H{
		"data": rule,
	})
}

func (g *Gateway) updateRule(c *gin.Context) {
	// 获取规则ID
	ruleID := c.Param("id")
	if ruleID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "规则ID不能为空",
			"message": "请提供有效的规则ID",
		})
		return
	}

	// 解析请求体
	var req models.RuleUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		g.logger.WithError(err).Error("解析更新规则请求失败")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "请求参数无效",
			"message": err.Error(),
		})
		return
	}

	// 构造规则对象，处理指针类型字段
	rule := &models.Rule{
		ID: ruleID,
	}

	// 只更新非空字段
	if req.Name != nil {
		rule.Name = *req.Name
	}
	if req.Description != nil {
		rule.Description = *req.Description
	}
	if req.Expression != nil {
		rule.Expression = *req.Expression
	}
	if req.Conditions != nil {
		rule.Conditions = *req.Conditions
	}
	if req.Actions != nil {
		rule.Actions = *req.Actions
	}
	if req.Severity != nil {
		rule.Severity = *req.Severity
	}
	if req.Type != nil {
		rule.Type = *req.Type
	}
	if req.Status != nil {
		rule.Status = *req.Status
	}
	if req.Labels != nil {
		rule.Labels = *req.Labels
	}
	if req.Annotations != nil {
		rule.Annotations = *req.Annotations
	}
	if req.EvaluationInterval != nil {
		rule.EvaluationInterval = *req.EvaluationInterval
	}
	if req.ForDuration != nil {
		rule.ForDuration = *req.ForDuration
	}
	if req.Threshold != nil {
		rule.Threshold = req.Threshold
	}
	if req.RecoveryThreshold != nil {
		rule.RecoveryThreshold = req.RecoveryThreshold
	}

	// 调用规则服务更新规则
	if err := g.serviceManager.Rule().Update(c.Request.Context(), rule); err != nil {
		g.logger.WithError(err).WithField("rule_id", ruleID).Error("更新规则失败")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "更新规则失败",
			"message": err.Error(),
		})
		return
	}

	g.logger.WithField("rule_id", ruleID).WithField("rule_name", rule.Name).Info("规则更新成功")
	c.JSON(http.StatusOK, gin.H{
		"message": "规则更新成功",
		"data":    rule,
	})
}

func (g *Gateway) deleteRule(c *gin.Context) {
	// 获取规则ID
	ruleID := c.Param("id")
	if ruleID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "规则ID不能为空",
			"message": "请提供有效的规则ID",
		})
		return
	}

	// 调用规则服务删除规则
	if err := g.serviceManager.Rule().Delete(c.Request.Context(), ruleID); err != nil {
		g.logger.WithError(err).WithField("rule_id", ruleID).Error("删除规则失败")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "删除规则失败",
			"message": err.Error(),
		})
		return
	}

	g.logger.WithField("rule_id", ruleID).Info("规则删除成功")
	c.JSON(http.StatusOK, gin.H{
		"message": "规则删除成功",
		"id":      ruleID,
	})
}

func (g *Gateway) enableRule(c *gin.Context) {
	// 获取规则ID
	ruleID := c.Param("id")
	if ruleID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "规则ID不能为空",
			"message": "请提供有效的规则ID",
		})
		return
	}

	// 调用规则服务启用规则
	if err := g.serviceManager.Rule().Enable(c.Request.Context(), ruleID); err != nil {
		g.logger.WithError(err).WithField("rule_id", ruleID).Error("启用规则失败")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "启用规则失败",
			"message": err.Error(),
		})
		return
	}

	g.logger.WithField("rule_id", ruleID).Info("规则启用成功")
	c.JSON(http.StatusOK, gin.H{
		"message": "规则启用成功",
		"id":      ruleID,
	})
}

func (g *Gateway) disableRule(c *gin.Context) {
	// 获取规则ID
	ruleID := c.Param("id")
	if ruleID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "规则ID不能为空",
			"message": "请提供有效的规则ID",
		})
		return
	}

	// 调用规则服务禁用规则
	if err := g.serviceManager.Rule().Disable(c.Request.Context(), ruleID); err != nil {
		g.logger.WithError(err).WithField("rule_id", ruleID).Error("禁用规则失败")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "禁用规则失败",
			"message": err.Error(),
		})
		return
	}

	g.logger.WithField("rule_id", ruleID).Info("规则禁用成功")
	c.JSON(http.StatusOK, gin.H{
		"message": "规则禁用成功",
		"id":      ruleID,
	})
}

// 数据源相关处理函数
func (g *Gateway) listDataSources(c *gin.Context) {
	// 解析查询参数
	filter := &models.DataSourceFilter{}
	
	// 分页参数
	if pageStr := c.Query("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil && page > 0 {
			filter.Page = page
		}
	}
	if pageSizeStr := c.Query("page_size"); pageSizeStr != "" {
		if pageSize, err := strconv.Atoi(pageSizeStr); err == nil && pageSize > 0 {
			filter.PageSize = pageSize
		}
	}
	
	// 过滤参数
	if keyword := c.Query("keyword"); keyword != "" {
		filter.Keyword = &keyword
	}
	if dsType := c.Query("type"); dsType != "" {
		filter.Type = (*models.DataSourceType)(&dsType)
	}
	if status := c.Query("status"); status != "" {
		filter.Status = (*models.DataSourceStatus)(&status)
	}
	
	// 调用服务层
	dataSources, total, err := g.serviceManager.DataSource().List(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	// 返回结果
	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"data_sources": dataSources,
			"total":        total,
			"page":         filter.Page,
			"page_size":    filter.PageSize,
		},
	})
}

func (g *Gateway) createDataSource(c *gin.Context) {
	var dataSource models.DataSource
	if err := c.ShouldBindJSON(&dataSource); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}
	
	// 调用服务层创建数据源
	if err := g.serviceManager.DataSource().Create(c.Request.Context(), &dataSource); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusCreated, gin.H{"data": dataSource})
}

func (g *Gateway) getDataSource(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "数据源ID不能为空"})
		return
	}
	
	// 调用服务层获取数据源
	dataSource, err := g.serviceManager.DataSource().GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"data": dataSource})
}

func (g *Gateway) updateDataSource(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "数据源ID不能为空"})
		return
	}
	
	var dataSource models.DataSource
	if err := c.ShouldBindJSON(&dataSource); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}
	
	// 设置ID
	dataSource.ID = id
	
	// 调用服务层更新数据源
	if err := g.serviceManager.DataSource().Update(c.Request.Context(), &dataSource); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"data": dataSource})
}

func (g *Gateway) deleteDataSource(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "数据源ID不能为空"})
		return
	}
	
	// 调用服务层删除数据源
	if err := g.serviceManager.DataSource().Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "数据源删除成功"})
}

func (g *Gateway) testDataSource(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "数据源ID不能为空"})
		return
	}
	
	// 调用服务层测试数据源连接
	if err := g.serviceManager.DataSource().TestConnection(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "数据源连接测试成功"})
}

// 工单相关处理函数
func (g *Gateway) listTickets(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented yet"})
}

func (g *Gateway) createTicket(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented yet"})
}

func (g *Gateway) getTicket(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented yet"})
}

func (g *Gateway) updateTicket(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented yet"})
}

func (g *Gateway) deleteTicket(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented yet"})
}

func (g *Gateway) assignTicket(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented yet"})
}

// 知识库相关处理函数
func (g *Gateway) listKnowledge(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented yet"})
}

func (g *Gateway) createKnowledge(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented yet"})
}

func (g *Gateway) getKnowledge(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented yet"})
}

func (g *Gateway) updateKnowledge(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented yet"})
}

func (g *Gateway) deleteKnowledge(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented yet"})
}

func (g *Gateway) searchKnowledge(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented yet"})
}

// 用户相关处理函数
func (g *Gateway) listUsers(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented yet"})
}

func (g *Gateway) createUser(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented yet"})
}

func (g *Gateway) getUser(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented yet"})
}

func (g *Gateway) updateUser(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented yet"})
}

func (g *Gateway) deleteUser(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented yet"})
}

// Webhook相关处理函数
func (g *Gateway) listWebhooks(c *gin.Context) {
	// 解析查询参数
	filter := &models.WebhookFilter{
		Page:     1,
		PageSize: 20,
	}

	// 解析分页参数
	if pageStr := c.Query("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil && page > 0 {
			filter.Page = page
		}
	}

	if pageSizeStr := c.Query("page_size"); pageSizeStr != "" {
		if pageSize, err := strconv.Atoi(pageSizeStr); err == nil && pageSize > 0 && pageSize <= 100 {
			filter.PageSize = pageSize
		}
	}

	// 解析过滤参数
	if name := c.Query("name"); name != "" {
		filter.Name = &name
	}

	if statusStr := c.Query("status"); statusStr != "" {
		status := models.WebhookStatus(statusStr)
		filter.Status = &status
	}

	if createdByStr := c.Query("created_by"); createdByStr != "" {
		if createdByUUID, err := uuid.Parse(createdByStr); err == nil {
			filter.CreatedBy = &createdByUUID
		}
	}

	// 调用Webhook服务获取列表
	webhooks, total, err := g.serviceManager.Webhook().List(c.Request.Context(), filter)
	if err != nil {
		g.logger.WithError(err).Error("获取Webhook列表失败")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "获取Webhook列表失败",
			"message": err.Error(),
		})
		return
	}

	// 计算总页数
	totalPages := int(total) / filter.PageSize
	if int(total)%filter.PageSize > 0 {
		totalPages++
	}

	// 构造响应
	response := gin.H{
		"webhooks":    webhooks,
		"total":       total,
		"page":        filter.Page,
		"page_size":   filter.PageSize,
		"total_pages": totalPages,
	}

	c.JSON(http.StatusOK, response)
}

func (g *Gateway) createWebhook(c *gin.Context) {
	// 解析请求体
	var webhook models.Webhook
	if err := c.ShouldBindJSON(&webhook); err != nil {
		g.logger.WithError(err).Error("解析创建Webhook请求失败")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "请求参数无效",
			"message": err.Error(),
		})
		return
	}

	// 调用Webhook服务创建Webhook
	if err := g.serviceManager.Webhook().Create(c.Request.Context(), &webhook); err != nil {
		g.logger.WithError(err).Error("创建Webhook失败")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "创建Webhook失败",
			"message": err.Error(),
		})
		return
	}

	g.logger.WithField("webhook_id", webhook.ID).Info("Webhook创建成功")
	c.JSON(http.StatusCreated, webhook)
}

func (g *Gateway) getWebhook(c *gin.Context) {
	// 获取Webhook ID
	webhookID := c.Param("id")
	if webhookID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Webhook ID不能为空",
			"message": "请提供有效的Webhook ID",
		})
		return
	}

	// 调用Webhook服务获取Webhook详情
	webhook, err := g.serviceManager.Webhook().GetByID(c.Request.Context(), webhookID)
	if err != nil {
		g.logger.WithError(err).WithField("webhook_id", webhookID).Error("获取Webhook详情失败")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "获取Webhook详情失败",
			"message": err.Error(),
		})
		return
	}

	// 检查Webhook是否存在
	if webhook == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Webhook不存在",
			"message": "指定的Webhook ID不存在",
		})
		return
	}

	c.JSON(http.StatusOK, webhook)
}

func (g *Gateway) updateWebhook(c *gin.Context) {
	// 获取Webhook ID
	webhookID := c.Param("id")
	if webhookID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Webhook ID不能为空",
			"message": "请提供有效的Webhook ID",
		})
		return
	}

	// 解析请求体
	var webhook models.Webhook
	if err := c.ShouldBindJSON(&webhook); err != nil {
		g.logger.WithError(err).Error("解析更新Webhook请求失败")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "请求参数无效",
			"message": err.Error(),
		})
		return
	}

	// 设置ID
	webhookUUID, err := uuid.Parse(webhookID)
	if err != nil {
		g.logger.WithError(err).Error("解析Webhook ID失败")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Webhook ID格式无效",
			"message": err.Error(),
		})
		return
	}
	webhook.ID = webhookUUID

	// 调用Webhook服务更新Webhook
	if err := g.serviceManager.Webhook().Update(c.Request.Context(), &webhook); err != nil {
		g.logger.WithError(err).WithField("webhook_id", webhookID).Error("更新Webhook失败")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "更新Webhook失败",
			"message": err.Error(),
		})
		return
	}

	g.logger.WithField("webhook_id", webhookID).Info("Webhook更新成功")
	c.JSON(http.StatusOK, webhook)
}

func (g *Gateway) deleteWebhook(c *gin.Context) {
	// 获取Webhook ID
	webhookID := c.Param("id")
	if webhookID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Webhook ID不能为空",
			"message": "请提供有效的Webhook ID",
		})
		return
	}

	// 调用Webhook服务删除Webhook
	if err := g.serviceManager.Webhook().Delete(c.Request.Context(), webhookID); err != nil {
		g.logger.WithError(err).WithField("webhook_id", webhookID).Error("删除Webhook失败")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "删除Webhook失败",
			"message": err.Error(),
		})
		return
	}

	g.logger.WithField("webhook_id", webhookID).Info("Webhook删除成功")
	c.JSON(http.StatusOK, gin.H{
		"message": "Webhook删除成功",
		"id":      webhookID,
	})
}

// 配置相关处理函数
func (g *Gateway) listConfig(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented yet"})
}

func (g *Gateway) setConfig(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented yet"})
}

func (g *Gateway) deleteConfig(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented yet"})
}

func (g *Gateway) triggerWebhook(c *gin.Context) {
	// 获取Webhook ID
	webhookID := c.Param("id")
	if webhookID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Webhook ID不能为空",
			"message": "请提供有效的Webhook ID",
		})
		return
	}

	// 解析请求体获取payload
	var payload interface{}
	if err := c.ShouldBindJSON(&payload); err != nil {
		// 如果没有payload，使用空对象
		payload = map[string]interface{}{}
	}

	// 调用Webhook服务触发Webhook
	if err := g.serviceManager.Webhook().Trigger(c.Request.Context(), webhookID, payload); err != nil {
		g.logger.WithError(err).WithField("webhook_id", webhookID).Error("触发Webhook失败")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "触发Webhook失败",
			"message": err.Error(),
		})
		return
	}

	g.logger.WithField("webhook_id", webhookID).Info("Webhook触发成功")
	c.JSON(http.StatusOK, gin.H{
		"message":    "Webhook触发成功",
		"webhook_id": webhookID,
		"status":     "triggered",
	})
}

// Worker状态处理函数
func (g *Gateway) getWorkerStatus(c *gin.Context) {
	// TODO: 实现获取Worker状态逻辑
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented yet"})
}