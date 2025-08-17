import { apiRequest } from './api';
import { Rule, RuleCondition, RuleAction, DataSource, PaginatedResponse } from '../types';

export const ruleService = {
  // 获取规则列表
  getRules: async (params?: {
    page?: number;
    limit?: number;
    search?: string;
    enabled?: boolean;
    data_source_id?: string;
  }): Promise<PaginatedResponse<Rule>> => {
    const response = await apiRequest.get<PaginatedResponse<Rule>>('/rules', params);
    return response.data;
  },

  // 获取规则详情
  getRule: async (id: string): Promise<Rule> => {
    const response = await apiRequest.get<Rule>(`/rules/${id}`);
    return response.data;
  },

  // 创建规则
  createRule: async (ruleData: {
    name: string;
    description: string;
    data_source_id: string;
    query: string;
    conditions: RuleCondition[];
    actions: RuleAction[];
    enabled?: boolean;
  }): Promise<Rule> => {
    const response = await apiRequest.post<Rule>('/rules', ruleData);
    return response.data;
  },

  // 更新规则
  updateRule: async (id: string, ruleData: Partial<Rule>): Promise<Rule> => {
    const response = await apiRequest.put<Rule>(`/rules/${id}`, ruleData);
    return response.data;
  },

  // 删除规则
  deleteRule: async (id: string): Promise<void> => {
    await apiRequest.delete(`/rules/${id}`);
  },

  // 批量删除规则
  batchDeleteRules: async (ids: string[]): Promise<void> => {
    await apiRequest.delete('/rules/batch', { data: { ids } });
  },

  // 启用/禁用规则
  toggleRule: async (id: string, enabled: boolean): Promise<Rule> => {
    const response = await apiRequest.patch<Rule>(`/rules/${id}/toggle`, { enabled });
    return response.data;
  },

  // 批量启用/禁用规则
  batchToggleRules: async (ids: string[], enabled: boolean): Promise<void> => {
    await apiRequest.patch('/rules/batch/toggle', { ids, enabled });
  },

  // 测试规则
  testRule: async (ruleData: {
    data_source_id: string;
    query: string;
    conditions: RuleCondition[];
  }): Promise<{
    success: boolean;
    result: any;
    message?: string;
  }> => {
    const response = await apiRequest.post('/rules/test', ruleData);
    return response.data;
  },

  // 获取规则执行历史
  getRuleHistory: async (id: string, params?: {
    page?: number;
    limit?: number;
    start_time?: string;
    end_time?: string;
  }): Promise<PaginatedResponse<{
    id: string;
    rule_id: string;
    executed_at: string;
    success: boolean;
    result: any;
    error?: string;
  }>> => {
    const response = await apiRequest.get(`/rules/${id}/history`, params);
    return response.data;
  },

  // 获取规则统计信息
  getRuleStatistics: async (params?: {
    start_time?: string;
    end_time?: string;
  }): Promise<{
    total: number;
    enabled: number;
    disabled: number;
    executions: {
      total: number;
      success: number;
      failed: number;
    };
    trend: Array<{
      time: string;
      executions: number;
      alerts: number;
    }>;
  }> => {
    const response = await apiRequest.get('/rules/statistics', params);
    return response.data;
  },

  // 复制规则
  duplicateRule: async (id: string, name?: string): Promise<Rule> => {
    const response = await apiRequest.post<Rule>(`/rules/${id}/duplicate`, { name });
    return response.data;
  },

  // 导入规则
  importRules: async (file: File): Promise<{
    success: number;
    failed: number;
    errors: string[];
  }> => {
    const formData = new FormData();
    formData.append('file', file);
    const response = await apiRequest.post('/rules/import', formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    });
    return response.data;
  },

  // 导出规则
  exportRules: async (ids?: string[], format: 'json' | 'yaml' = 'json'): Promise<Blob> => {
    const response = await apiRequest.post('/rules/export', { ids, format }, {
      responseType: 'blob'
    });
    return response.data;
  },
};

// 数据源服务
export const dataSourceService = {
  // 获取数据源列表
  getDataSources: async (params?: {
    page?: number;
    limit?: number;
    type?: string;
    enabled?: boolean;
  }): Promise<PaginatedResponse<DataSource>> => {
    const response = await apiRequest.get<PaginatedResponse<DataSource>>('/datasources', params);
    return response.data;
  },

  // 获取数据源详情
  getDataSource: async (id: string): Promise<DataSource> => {
    const response = await apiRequest.get<DataSource>(`/datasources/${id}`);
    return response.data;
  },

  // 创建数据源
  createDataSource: async (dataSourceData: {
    name: string;
    type: string;
    url: string;
    config: Record<string, any>;
    enabled?: boolean;
  }): Promise<DataSource> => {
    const response = await apiRequest.post<DataSource>('/datasources', dataSourceData);
    return response.data;
  },

  // 更新数据源
  updateDataSource: async (id: string, dataSourceData: Partial<DataSource>): Promise<DataSource> => {
    const response = await apiRequest.put<DataSource>(`/datasources/${id}`, dataSourceData);
    return response.data;
  },

  // 删除数据源
  deleteDataSource: async (id: string): Promise<void> => {
    await apiRequest.delete(`/datasources/${id}`);
  },

  // 测试数据源连接
  testDataSource: async (dataSourceData: {
    type: string;
    url: string;
    config: Record<string, any>;
  }): Promise<{
    success: boolean;
    message: string;
    latency?: number;
  }> => {
    const response = await apiRequest.post('/datasources/test', dataSourceData);
    return response.data;
  },

  // 获取数据源类型列表
  getDataSourceTypes: async (): Promise<Array<{
    type: string;
    name: string;
    description: string;
    config_schema: any;
  }>> => {
    const response = await apiRequest.get('/datasources/types');
    return response.data;
  },
};