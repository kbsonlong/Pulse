import { apiRequest } from './api';

export interface SystemSettings {
  // 基本设置
  site_name: string;
  site_description: string;
  site_logo?: string;
  timezone: string;
  language: string;
  
  // 告警设置
  alert_retention_days: number;
  max_alerts_per_rule: number;
  alert_batch_size: number;
  
  // 通知设置
  email_enabled: boolean;
  email_smtp_host: string;
  email_smtp_port: number;
  email_smtp_username: string;
  email_smtp_password: string;
  email_from_address: string;
  
  sms_enabled: boolean;
  sms_provider: string;
  sms_api_key: string;
  sms_api_secret: string;
  
  webhook_enabled: boolean;
  webhook_timeout: number;
  webhook_retry_count: number;
  
  // 安全设置
  session_timeout: number;
  password_min_length: number;
  password_require_uppercase: boolean;
  password_require_lowercase: boolean;
  password_require_numbers: boolean;
  password_require_symbols: boolean;
  login_attempt_limit: number;
  login_lockout_duration: number;
  
  // 性能设置
  query_timeout: number;
  max_concurrent_queries: number;
  cache_ttl: number;
  log_level: string;
  
  // 备份设置
  backup_enabled: boolean;
  backup_schedule: string;
  backup_retention_days: number;
  backup_storage_path: string;
}

export interface NotificationSettings {
  email_notifications: boolean;
  sms_notifications: boolean;
  webhook_notifications: boolean;
  notification_frequency: 'immediate' | 'hourly' | 'daily';
  quiet_hours_enabled: boolean;
  quiet_hours_start: string;
  quiet_hours_end: string;
  notification_channels: string[];
}

export const systemService = {
  // 获取系统设置
  getSystemSettings: async (): Promise<SystemSettings> => {
    const response = await apiRequest.get<SystemSettings>('/system/settings');
    return response.data;
  },

  // 更新系统设置
  updateSystemSettings: async (settings: Partial<SystemSettings>): Promise<SystemSettings> => {
    const response = await apiRequest.put<SystemSettings>('/system/settings', settings);
    return response.data;
  },

  // 重置系统设置
  resetSystemSettings: async (): Promise<SystemSettings> => {
    const response = await apiRequest.post<SystemSettings>('/system/settings/reset');
    return response.data;
  },

  // 测试邮件配置
  testEmailConfig: async (config: {
    smtp_host: string;
    smtp_port: number;
    smtp_user: string;
    smtp_password: string;
    test_email: string;
  }): Promise<{
    success: boolean;
    message: string;
  }> => {
    const response = await apiRequest.post('/system/test-email', config);
    return response.data as { success: boolean; message: string; };
  },

  // 测试短信配置
  testSmsConfig: async (config: {
    provider: string;
    api_key: string;
    api_secret: string;
    test_phone: string;
  }): Promise<{
    success: boolean;
    message: string;
  }> => {
    const response = await apiRequest.post('/system/test-sms', config);
    return response.data as { success: boolean; message: string; };
  },

  // 测试Webhook配置
  testWebhookConfig: async (config: {
    url: string;
    timeout: number;
    headers?: Record<string, string>;
  }): Promise<{
    success: boolean;
    message: string;
    response_time: number;
  }> => {
    const response = await apiRequest.post('/system/test-webhook', config);
    return response.data;
  },

  // 获取系统状态
  getSystemStatus: async (): Promise<{
    status: 'healthy' | 'warning' | 'error';
    uptime: number;
    version: string;
    database: {
      status: 'connected' | 'disconnected';
      latency: number;
    };
    redis: {
      status: 'connected' | 'disconnected';
      latency: number;
    };
    disk_usage: {
      total: number;
      used: number;
      free: number;
      percentage: number;
    };
    memory_usage: {
      total: number;
      used: number;
      free: number;
      percentage: number;
    };
    cpu_usage: number;
    active_connections: number;
  }> => {
    const response = await apiRequest.get('/system/status');
    return response.data;
  },

  // 获取系统日志
  getSystemLogs: async (params?: {
    level?: 'debug' | 'info' | 'warn' | 'error';
    start_time?: string;
    end_time?: string;
    page?: number;
    limit?: number;
  }): Promise<{
    logs: Array<{
      timestamp: string;
      level: string;
      message: string;
      context?: Record<string, any>;
    }>;
    total: number;
  }> => {
    const response = await apiRequest.get('/system/logs', params);
    return response.data;
  },

  // 清理系统日志
  clearSystemLogs: async (before_date?: string): Promise<{
    deleted_count: number;
  }> => {
    const response = await apiRequest.delete('/system/logs', {
      data: { before_date }
    });
    return response.data;
  },

  // 创建系统备份
  createBackup: async (): Promise<{
    backup_id: string;
    file_path: string;
    size: number;
  }> => {
    const response = await apiRequest.post('/system/backup');
    return response.data;
  },

  // 获取备份列表
  getBackups: async (): Promise<Array<{
    id: string;
    created_at: string;
    size: number;
    file_path: string;
    status: 'completed' | 'failed' | 'in_progress';
  }>> => {
    const response = await apiRequest.get('/system/backups');
    return response.data;
  },

  // 恢复备份
  restoreBackup: async (backup_id: string): Promise<{
    success: boolean;
    message: string;
  }> => {
    const response = await apiRequest.post(`/system/backups/${backup_id}/restore`);
    return response.data;
  },

  // 删除备份
  deleteBackup: async (backup_id: string): Promise<void> => {
    await apiRequest.delete(`/system/backups/${backup_id}`);
  },

  // 获取用户通知设置
  getUserNotificationSettings: async (): Promise<NotificationSettings> => {
    const response = await apiRequest.get<NotificationSettings>('/system/notification-settings');
    return response.data;
  },

  // 更新用户通知设置
  updateUserNotificationSettings: async (settings: Partial<NotificationSettings>): Promise<NotificationSettings> => {
    const response = await apiRequest.put<NotificationSettings>('/system/notification-settings', settings);
    return response.data;
  },

  // 获取系统统计信息
  getSystemStatistics: async (params?: {
    start_time?: string;
    end_time?: string;
  }): Promise<{
    alerts: {
      total: number;
      by_severity: Record<string, number>;
      trend: Array<{
        time: string;
        count: number;
      }>;
    };
    rules: {
      total: number;
      active: number;
      executions: number;
    };
    tickets: {
      total: number;
      by_status: Record<string, number>;
      resolution_time: number;
    };
    users: {
      total: number;
      active: number;
      online: number;
    };
    performance: {
      avg_response_time: number;
      error_rate: number;
      throughput: number;
    };
  }> => {
    const response = await apiRequest.get('/system/statistics', params);
    return response.data;
  },

  // 重启系统服务
  restartService: async (service_name: string): Promise<{
    success: boolean;
    message: string;
  }> => {
    const response = await apiRequest.post(`/system/services/${service_name}/restart`);
    return response.data as { success: boolean; message: string; };
  },

  // 获取系统配置模板
  getConfigTemplate: async (type: string): Promise<{
    template: Record<string, any>;
    schema: Record<string, any>;
  }> => {
    const response = await apiRequest.get(`/system/config-templates/${type}`);
    return response.data as { template: Record<string, any>; schema: Record<string, any>; };
  },
};