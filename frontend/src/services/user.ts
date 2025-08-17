import { apiRequest } from './api';
import { User, PaginatedResponse } from '../types';

export const userService = {
  // 获取用户列表
  getUsers: async (params?: {
    page?: number;
    limit?: number;
    search?: string;
    role?: string;
    status?: 'active' | 'inactive';
    department?: string;
  }): Promise<PaginatedResponse<User>> => {
    const response = await apiRequest.get<PaginatedResponse<User>>('/users', params);
    return response.data;
  },

  // 获取用户详情
  getUser: async (id: string): Promise<User> => {
    const response = await apiRequest.get<User>(`/users/${id}`);
    return response.data;
  },

  // 创建用户
  createUser: async (userData: {
    username: string;
    email: string;
    password: string;
    name: string;
    role: string;
    department?: string;
    phone?: string;
  }): Promise<User> => {
    const response = await apiRequest.post<User>('/users', userData);
    return response.data;
  },

  // 更新用户
  updateUser: async (id: string, userData: {
    username?: string;
    email?: string;
    name?: string;
    role?: string;
    department?: string;
    phone?: string;
    status?: 'active' | 'inactive';
  }): Promise<User> => {
    const response = await apiRequest.put<User>(`/users/${id}`, userData);
    return response.data;
  },

  // 删除用户
  deleteUser: async (id: string): Promise<void> => {
    await apiRequest.delete(`/users/${id}`);
  },

  // 批量删除用户
  batchDeleteUsers: async (ids: string[]): Promise<void> => {
    await apiRequest.delete('/users/batch', { data: { ids } });
  },

  // 重置用户密码
  resetPassword: async (id: string, newPassword?: string): Promise<{
    password: string;
  }> => {
    const response = await apiRequest.post(`/users/${id}/reset-password`, {
      new_password: newPassword
    });
    return response.data;
  },

  // 更新用户状态
  updateUserStatus: async (id: string, status: 'active' | 'inactive'): Promise<User> => {
    const response = await apiRequest.patch<User>(`/users/${id}/status`, { status });
    return response.data;
  },

  // 批量更新用户状态
  batchUpdateStatus: async (ids: string[], status: 'active' | 'inactive'): Promise<void> => {
    await apiRequest.patch('/users/batch/status', { ids, status });
  },

  // 获取用户角色列表
  getRoles: async (): Promise<Array<{
    id: string;
    name: string;
    description: string;
    permissions: string[];
  }>> => {
    const response = await apiRequest.get('/users/roles');
    return response.data;
  },

  // 获取部门列表
  getDepartments: async (): Promise<Array<{
    id: string;
    name: string;
    description?: string;
  }>> => {
    const response = await apiRequest.get('/users/departments');
    return response.data;
  },

  // 获取当前用户信息
  getCurrentUser: async (): Promise<User> => {
    const response = await apiRequest.get<User>('/users/me');
    return response.data;
  },

  // 更新当前用户信息
  updateCurrentUser: async (userData: {
    name?: string;
    email?: string;
    phone?: string;
    avatar?: string;
  }): Promise<User> => {
    const response = await apiRequest.put<User>('/users/me', userData);
    return response.data;
  },

  // 修改当前用户密码
  changePassword: async (passwordData: {
    current_password: string;
    new_password: string;
  }): Promise<void> => {
    await apiRequest.post('/users/me/change-password', passwordData);
  },

  // 上传用户头像
  uploadAvatar: async (file: File): Promise<{
    avatar_url: string;
  }> => {
    const formData = new FormData();
    formData.append('avatar', file);
    const response = await apiRequest.post('/users/me/avatar', formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    });
    return response.data;
  },

  // 获取用户统计信息
  getUserStatistics: async (): Promise<{
    total: number;
    active: number;
    inactive: number;
    by_role: Record<string, number>;
    by_department: Record<string, number>;
    recent_logins: Array<{
      date: string;
      count: number;
    }>
  }> => {
    const response = await apiRequest.get('/users/statistics');
    return response.data;
  },

  // 搜索用户
  searchUsers: async (params: {
    query: string;
    page?: number;
    limit?: number;
    filters?: {
      role?: string[];
      department?: string[];
      status?: ('active' | 'inactive')[];
    };
  }): Promise<PaginatedResponse<User>> => {
    const response = await apiRequest.post<PaginatedResponse<User>>('/users/search', params);
    return response.data;
  },

  // 导出用户数据
  exportUsers: async (params?: {
    role?: string;
    department?: string;
    status?: 'active' | 'inactive';
    format?: 'csv' | 'excel';
  }): Promise<Blob> => {
    const response = await apiRequest.get('/users/export', params, {
      responseType: 'blob'
    });
    return response.data;
  },

  // 批量导入用户
  importUsers: async (file: File): Promise<{
    success: number;
    failed: number;
    errors: string[];
  }> => {
    const formData = new FormData();
    formData.append('file', file);
    const response = await apiRequest.post('/users/import', formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    });
    return response.data;
  },
};