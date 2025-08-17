import { apiRequest } from './api';
import { User, LoginRequest, LoginResponse } from '../types';

export const authService = {
  // 用户登录
  login: async (credentials: LoginRequest): Promise<LoginResponse> => {
    const response = await apiRequest.post<LoginResponse>('/auth/login', credentials);
    if (response.data.token) {
      localStorage.setItem('token', response.data.token);
      localStorage.setItem('user', JSON.stringify(response.data.user));
    }
    return response.data;
  },

  // 用户注册
  register: async (userData: {
    username: string;
    email: string;
    password: string;
    role?: string;
  }): Promise<User> => {
    const response = await apiRequest.post<User>('/auth/register', userData);
    return response.data;
  },

  // 用户登出
  logout: (): void => {
    localStorage.removeItem('token');
    localStorage.removeItem('user');
    window.location.href = '/login';
  },

  // 获取当前用户信息
  getCurrentUser: (): User | null => {
    const userStr = localStorage.getItem('user');
    return userStr ? JSON.parse(userStr) : null;
  },

  // 检查是否已登录
  isAuthenticated: (): boolean => {
    return !!localStorage.getItem('token');
  },

  // 获取token
  getToken: (): string | null => {
    return localStorage.getItem('token');
  },

  // 刷新token
  refreshToken: async (): Promise<string> => {
    const response = await apiRequest.post<{ token: string }>('/auth/refresh');
    if (response.data.token) {
      localStorage.setItem('token', response.data.token);
    }
    return response.data.token;
  },

  // 修改密码
  changePassword: async (data: {
    old_password: string;
    new_password: string;
  }): Promise<void> => {
    await apiRequest.post('/auth/change-password', data);
  },

  // 获取用户列表（管理员功能）
  getUsers: async (params?: {
    page?: number;
    limit?: number;
    search?: string;
  }): Promise<{
    users: User[];
    total: number;
    page: number;
    limit: number;
  }> => {
    const response = await apiRequest.get('/auth/users', params);
    return response.data;
  },

  // 创建用户（管理员功能）
  createUser: async (userData: {
    username: string;
    email: string;
    password: string;
    role: string;
  }): Promise<User> => {
    const response = await apiRequest.post<User>('/auth/users', userData);
    return response.data;
  },

  // 更新用户（管理员功能）
  updateUser: async (id: string, userData: Partial<User>): Promise<User> => {
    const response = await apiRequest.put<User>(`/auth/users/${id}`, userData);
    return response.data;
  },

  // 删除用户（管理员功能）
  deleteUser: async (id: string): Promise<void> => {
    await apiRequest.delete(`/auth/users/${id}`);
  },
};