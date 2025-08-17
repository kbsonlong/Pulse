import { apiRequest } from './api';
import {
  Alert,
  AlertQuery,
  AlertList,
  AlertStatistics,
  AlertLevel,
  AlertStatus,
  PaginatedResponse
} from '../types';

export const alertService = {
  // 获取告警列表
  getAlerts: async (query?: AlertQuery): Promise<AlertList> => {
    const response = await apiRequest.get<AlertList>('/alerts', query);
    return response.data;
  },

  // 获取告警详情
  getAlert: async (id: string): Promise<Alert> => {
    const response = await apiRequest.get<Alert>(`/alerts/${id}`);
    return response.data;
  },

  // 创建告警
  createAlert: async (alertData: Omit<Alert, 'id' | 'created_at' | 'updated_at'>): Promise<Alert> => {
    const response = await apiRequest.post<Alert>('/alerts', alertData);
    return response.data;
  },

  // 更新告警状态
  updateAlertStatus: async (id: string, status: AlertStatus): Promise<Alert> => {
    const response = await apiRequest.patch<Alert>(`/alerts/${id}/status`, { status });
    return response.data;
  },

  // 批量更新告警状态
  batchUpdateStatus: async (ids: string[], status: AlertStatus): Promise<void> => {
    await apiRequest.patch('/alerts/batch/status', { ids, status });
  },

  // 删除告警
  deleteAlert: async (id: string): Promise<void> => {
    await apiRequest.delete(`/alerts/${id}`);
  },

  // 批量删除告警
  batchDeleteAlerts: async (ids: string[]): Promise<void> => {
    await apiRequest.delete('/alerts/batch', { data: { ids } });
  },

  // 获取告警统计信息
  getStatistics: async (params?: {
    start_time?: string;
    end_time?: string;
    source?: string;
  }): Promise<AlertStatistics> => {
    const response = await apiRequest.get<AlertStatistics>('/alerts/statistics', params);
    return response.data;
  },

  // 获取告警趋势数据
  getTrend: async (params: {
    start_time: string;
    end_time: string;
    interval?: string; // hour, day, week
    source?: string;
    level?: AlertLevel;
  }): Promise<Array<{ time: string; count: number }>> => {
    const response = await apiRequest.get('/alerts/trend', params);
    return response.data;
  },

  // 获取告警来源列表
  getSources: async (): Promise<string[]> => {
    const response = await apiRequest.get<string[]>('/alerts/sources');
    return response.data;
  },

  // 搜索告警
  searchAlerts: async (params: {
    query: string;
    page?: number;
    limit?: number;
    filters?: {
      level?: AlertLevel[];
      status?: AlertStatus[];
      source?: string[];
      start_time?: string;
      end_time?: string;
    };
  }): Promise<AlertList> => {
    const response = await apiRequest.post<AlertList>('/alerts/search', params);
    return response.data;
  },

  // 获取相似告警
  getSimilarAlerts: async (id: string, limit?: number): Promise<Alert[]> => {
    const response = await apiRequest.get<Alert[]>(`/alerts/${id}/similar`, { limit });
    return response.data;
  },

  // 抑制告警
  suppressAlert: async (id: string, duration: number, reason?: string): Promise<void> => {
    await apiRequest.post(`/alerts/${id}/suppress`, { duration, reason });
  },

  // 取消抑制
  unsuppressAlert: async (id: string): Promise<void> => {
    await apiRequest.post(`/alerts/${id}/unsuppress`);
  },

  // 确认告警
  acknowledgeAlert: async (id: string, comment?: string): Promise<void> => {
    await apiRequest.post(`/alerts/${id}/acknowledge`, { comment });
  },

  // 导出告警数据
  exportAlerts: async (query?: AlertQuery, format: 'csv' | 'json' = 'csv'): Promise<Blob> => {
    const response = await apiRequest.get('/alerts/export', { ...query, format }, {
      responseType: 'blob'
    });
    return response.data;
  },
};