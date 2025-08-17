import axios, { AxiosInstance, AxiosResponse } from 'axios';
import { ApiResponse } from '../types';

// 创建axios实例
const api: AxiosInstance = axios.create({
  baseURL: process.env.REACT_APP_API_BASE_URL || 'http://localhost:8080/api',
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json',
  },
});

// 请求拦截器 - 添加认证token
api.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('token');
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

// 响应拦截器 - 处理通用错误
api.interceptors.response.use(
  (response: AxiosResponse<ApiResponse<any>>) => {
    return response;
  },
  (error) => {
    if (error.response?.status === 401) {
      // 清除token并跳转到登录页
      localStorage.removeItem('token');
      window.location.href = '/login';
    }
    return Promise.reject(error);
  }
);

export default api;

// 通用API请求方法
export const apiRequest = {
  get: <T>(url: string, params?: any): Promise<ApiResponse<T>> => {
    return api.get(url, { params }).then(res => res.data);
  },
  
  post: <T>(url: string, data?: any): Promise<ApiResponse<T>> => {
    return api.post(url, data).then(res => res.data);
  },
  
  put: <T>(url: string, data?: any): Promise<ApiResponse<T>> => {
    return api.put(url, data).then(res => res.data);
  },
  
  delete: <T>(url: string): Promise<ApiResponse<T>> => {
    return api.delete(url).then(res => res.data);
  },
  
  patch: <T>(url: string, data?: any): Promise<ApiResponse<T>> => {
    return api.patch(url, data).then(res => res.data);
  },
};