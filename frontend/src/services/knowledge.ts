import { apiRequest } from './api';
import { Knowledge, PaginatedResponse } from '../types';

export const knowledgeService = {
  // 获取知识库列表
  getKnowledgeList: async (params?: {
    page?: number;
    limit?: number;
    category?: string;
    tags?: string[];
    search?: string;
    sort_by?: 'created_at' | 'updated_at' | 'score' | 'title';
    sort_order?: 'asc' | 'desc';
  }): Promise<PaginatedResponse<Knowledge>> => {
    const response = await apiRequest.get<PaginatedResponse<Knowledge>>('/knowledge', params);
    return response.data;
  },

  // 获取知识库详情
  getKnowledge: async (id: string): Promise<Knowledge> => {
    const response = await apiRequest.get<Knowledge>(`/knowledge/${id}`);
    return response.data;
  },

  // 创建知识库文档
  createKnowledge: async (knowledgeData: {
    title: string;
    content: string;
    tags: string[];
    category: string;
  }): Promise<Knowledge> => {
    const response = await apiRequest.post<Knowledge>('/knowledge', knowledgeData);
    return response.data;
  },

  // 更新知识库文档
  updateKnowledge: async (id: string, knowledgeData: {
    title?: string;
    content?: string;
    tags?: string[];
    category?: string;
  }): Promise<Knowledge> => {
    const response = await apiRequest.put<Knowledge>(`/knowledge/${id}`, knowledgeData);
    return response.data;
  },

  // 删除知识库文档
  deleteKnowledge: async (id: string): Promise<void> => {
    await apiRequest.delete(`/knowledge/${id}`);
  },

  // 批量删除知识库文档
  batchDeleteKnowledge: async (ids: string[]): Promise<void> => {
    await apiRequest.delete('/knowledge/batch', { data: { ids } });
  },

  // 搜索知识库
  searchKnowledge: async (params: {
    query: string;
    page?: number;
    limit?: number;
    category?: string;
    tags?: string[];
    min_score?: number;
  }): Promise<PaginatedResponse<Knowledge>> => {
    const response = await apiRequest.post<PaginatedResponse<Knowledge>>('/knowledge/search', params);
    return response.data;
  },

  // 智能搜索（基于AI）
  intelligentSearch: async (params: {
    query: string;
    context?: string;
    page?: number;
    limit?: number;
  }): Promise<{
    results: Knowledge[];
    suggestions: string[];
    total: number;
  }> => {
    const response = await apiRequest.post('/knowledge/intelligent-search', params);
    return response.data;
  },

  // 获取相关知识库文档
  getRelatedKnowledge: async (id: string, limit?: number): Promise<Knowledge[]> => {
    const response = await apiRequest.get<Knowledge[]>(`/knowledge/${id}/related`, { limit });
    return response.data;
  },

  // 获取知识库分类列表
  getCategories: async (): Promise<Array<{
    name: string;
    count: number;
  }>> => {
    const response = await apiRequest.get('/knowledge/categories');
    return response.data;
  },

  // 获取知识库标签列表
  getTags: async (category?: string): Promise<Array<{
    name: string;
    count: number;
  }>> => {
    const response = await apiRequest.get('/knowledge/tags', { category });
    return response.data;
  },

  // 获取热门知识库文档
  getPopularKnowledge: async (limit?: number): Promise<Knowledge[]> => {
    const response = await apiRequest.get<Knowledge[]>('/knowledge/popular', { limit });
    return response.data;
  },

  // 获取最新知识库文档
  getLatestKnowledge: async (limit?: number): Promise<Knowledge[]> => {
    const response = await apiRequest.get<Knowledge[]>('/knowledge/latest', { limit });
    return response.data;
  },

  // 评分知识库文档
  rateKnowledge: async (id: string, score: number, comment?: string): Promise<void> => {
    await apiRequest.post(`/knowledge/${id}/rate`, { score, comment });
  },

  // 获取知识库统计信息
  getKnowledgeStatistics: async (): Promise<{
    total: number;
    by_category: Record<string, number>;
    by_author: Record<string, number>;
    recent_activity: Array<{
      date: string;
      created: number;
      updated: number;
    }>;
    top_tags: Array<{
      name: string;
      count: number;
    }>;
  }> => {
    const response = await apiRequest.get('/knowledge/statistics');
    return response.data;
  },

  // 导入知识库文档
  importKnowledge: async (file: File, options?: {
    category?: string;
    tags?: string[];
    auto_categorize?: boolean;
  }): Promise<{
    success: number;
    failed: number;
    errors: string[];
  }> => {
    const formData = new FormData();
    formData.append('file', file);
    if (options) {
      formData.append('options', JSON.stringify(options));
    }
    const response = await apiRequest.post('/knowledge/import', formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    });
    return response.data;
  },

  // 导出知识库文档
  exportKnowledge: async (params?: {
    ids?: string[];
    category?: string;
    tags?: string[];
    format?: 'json' | 'markdown' | 'pdf';
  }): Promise<Blob> => {
    const response = await apiRequest.post('/knowledge/export', params, {
      responseType: 'blob'
    });
    return response.data;
  },

  // 生成知识库摘要（AI功能）
  generateSummary: async (id: string): Promise<{
    summary: string;
    key_points: string[];
  }> => {
    const response = await apiRequest.post(`/knowledge/${id}/generate-summary`);
    return response.data;
  },

  // 自动标记标签（AI功能）
  autoTagging: async (id: string): Promise<{
    suggested_tags: string[];
    confidence: number;
  }> => {
    const response = await apiRequest.post(`/knowledge/${id}/auto-tag`);
    return response.data;
  },

  // 检查内容重复
  checkDuplicate: async (content: string, title?: string): Promise<{
    is_duplicate: boolean;
    similar_documents: Array<{
      id: string;
      title: string;
      similarity: number;
    }>;
  }> => {
    const response = await apiRequest.post('/knowledge/check-duplicate', {
      content,
      title
    });
    return response.data;
  },
};