package models

import "errors"

// 通用错误定义
var (
	// 用户相关错误
	ErrUserNotFound     = errors.New("用户不存在")
	ErrUserExists       = errors.New("用户已存在")
	ErrInvalidPassword  = errors.New("密码无效")
	ErrUserDisabled     = errors.New("用户已禁用")

	// 数据源相关错误
	ErrDataSourceNotFound = errors.New("数据源不存在")
	ErrDataSourceExists   = errors.New("数据源已存在")
	ErrDataSourceOffline  = errors.New("数据源离线")

	// 规则相关错误
	ErrRuleNotFound     = errors.New("规则不存在")
	ErrRuleExists       = errors.New("规则已存在")
	ErrRuleDisabled     = errors.New("规则已禁用")
	ErrRuleEvalFailed   = errors.New("规则评估失败")

	// 告警相关错误
	ErrAlertNotFound    = errors.New("告警不存在")
	ErrAlertExists      = errors.New("告警已存在")
	ErrAlertResolved    = errors.New("告警已解决")

	// 工单相关错误
	ErrTicketNotFound   = errors.New("工单不存在")
	ErrTicketExists     = errors.New("工单已存在")
	ErrTicketClosed     = errors.New("工单已关闭")

	// 知识库相关错误
	ErrKnowledgeNotFound = errors.New("知识库文章不存在")
	ErrKnowledgeExists   = errors.New("知识库文章已存在")
	ErrVersionNotFound   = errors.New("版本不存在")

	// 权限相关错误
	ErrPermissionDenied  = errors.New("权限不足")
	ErrInvalidToken      = errors.New("无效的令牌")
	ErrTokenExpired      = errors.New("令牌已过期")

	// 通用错误
	ErrInvalidInput      = errors.New("输入参数无效")
	ErrInternalError     = errors.New("内部服务器错误")
	ErrDatabaseError     = errors.New("数据库错误")
	ErrNetworkError      = errors.New("网络错误")
	ErrTimeout           = errors.New("操作超时")
	ErrNotImplemented    = errors.New("功能未实现")
)