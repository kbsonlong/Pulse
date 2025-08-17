import { apiRequest } from './api';
import {
  Ticket,
  TicketStatus,
  TicketPriority,
  ProcessRecord,
  PaginatedResponse,
  User
} from '../types';

export const ticketService = {
  // 获取工单列表
  getTickets: async (params?: {
    page?: number;
    limit?: number;
    status?: TicketStatus;
    priority?: TicketPriority;
    assignee_id?: string;
    created_by?: string;
    search?: string;
    start_time?: string;
    end_time?: string;
  }): Promise<PaginatedResponse<Ticket>> => {
    const response = await apiRequest.get<PaginatedResponse<Ticket>>('/tickets', params);
    return response.data;
  },

  // 获取工单详情
  getTicket: async (id: string): Promise<Ticket> => {
    const response = await apiRequest.get<Ticket>(`/tickets/${id}`);
    return response.data;
  },

  // 创建工单
  createTicket: async (ticketData: {
    title: string;
    description: string;
    alert_id?: string;
    assignee_id?: string;
    priority: TicketPriority;
  }): Promise<Ticket> => {
    const response = await apiRequest.post<Ticket>('/tickets', ticketData);
    return response.data;
  },

  // 更新工单
  updateTicket: async (id: string, ticketData: {
    title?: string;
    description?: string;
    assignee_id?: string;
    priority?: TicketPriority;
    status?: TicketStatus;
  }): Promise<Ticket> => {
    const response = await apiRequest.put<Ticket>(`/tickets/${id}`, ticketData);
    return response.data;
  },

  // 删除工单
  deleteTicket: async (id: string): Promise<void> => {
    await apiRequest.delete(`/tickets/${id}`);
  },

  // 分配工单
  assignTicket: async (id: string, assignee_id: string, comment?: string): Promise<Ticket> => {
    const response = await apiRequest.post<Ticket>(`/tickets/${id}/assign`, {
      assignee_id,
      comment
    });
    return response.data;
  },

  // 更新工单状态
  updateTicketStatus: async (id: string, status: TicketStatus, comment?: string): Promise<Ticket> => {
    const response = await apiRequest.patch<Ticket>(`/tickets/${id}/status`, {
      status,
      comment
    });
    return response.data;
  },

  // 更新工单优先级
  updateTicketPriority: async (id: string, priority: TicketPriority, comment?: string): Promise<Ticket> => {
    const response = await apiRequest.patch<Ticket>(`/tickets/${id}/priority`, {
      priority,
      comment
    });
    return response.data;
  },

  // 批量更新工单状态
  batchUpdateStatus: async (ids: string[], status: TicketStatus): Promise<void> => {
    await apiRequest.patch('/tickets/batch/status', { ids, status });
  },

  // 批量分配工单
  batchAssign: async (ids: string[], assignee_id: string): Promise<void> => {
    await apiRequest.patch('/tickets/batch/assign', { ids, assignee_id });
  },

  // 获取工单处理记录
  getTicketProcessRecords: async (id: string, params?: {
    page?: number;
    limit?: number;
  }): Promise<PaginatedResponse<ProcessRecord>> => {
    const response = await apiRequest.get<PaginatedResponse<ProcessRecord>>(`/tickets/${id}/records`, params);
    return response.data;
  },

  // 添加工单处理记录
  addProcessRecord: async (id: string, recordData: {
    action: string;
    description: string;
  }): Promise<ProcessRecord> => {
    const response = await apiRequest.post<ProcessRecord>(`/tickets/${id}/records`, recordData);
    return response.data;
  },

  // 获取工单统计信息
  getTicketStatistics: async (params?: {
    start_time?: string;
    end_time?: string;
  }): Promise<{
    total: number;
    by_status: Record<TicketStatus, number>;
    by_priority: Record<TicketPriority, number>;
    by_assignee: Record<string, number>;
    resolution_time: {
      average: number;
      median: number;
    };
    trend: Array<{
      date: string;
      count: number;
    }>;
  }> => {
    const response = await apiRequest.get('/tickets/statistics', params);
    return response.data as { total: number; by_status: Record<TicketStatus, number>; by_priority: Record<TicketPriority, number>; by_assignee: Record<string, number>; resolution_time: { average: number; median: number; }; trend: Array<{ date: string; count: number; }>; };
  },

  // 获取我的工单
  getMyTickets: async (params?: {
    page?: number;
    limit?: number;
    status?: TicketStatus;
    priority?: TicketPriority;
  }): Promise<PaginatedResponse<Ticket>> => {
    const response = await apiRequest.get<PaginatedResponse<Ticket>>('/tickets/my', params);
    return response.data;
  },

  // 获取我创建的工单
  getMyCreatedTickets: async (params?: {
    page?: number;
    limit?: number;
    status?: TicketStatus;
    priority?: TicketPriority;
  }): Promise<PaginatedResponse<Ticket>> => {
    const response = await apiRequest.get<PaginatedResponse<Ticket>>('/tickets/created', params);
    return response.data;
  },

  // 搜索工单
  searchTickets: async (params: {
    query: string;
    page?: number;
    limit?: number;
    filters?: {
      status?: TicketStatus[];
      priority?: TicketPriority[];
      assignee_id?: string[];
      start_time?: string;
      end_time?: string;
    };
  }): Promise<PaginatedResponse<Ticket>> => {
    const response = await apiRequest.post<PaginatedResponse<Ticket>>('/tickets/search', params);
    return response.data;
  },

  // 获取可分配的用户列表
  getAssignableUsers: async (): Promise<User[]> => {
    const response = await apiRequest.get<User[]>('/tickets/assignable-users');
    return response.data;
  },

  // 导出工单数据
  exportTickets: async (params?: {
    format?: 'csv' | 'excel';
    status?: TicketStatus;
    priority?: TicketPriority;
    assignee?: string;
    start_time?: string;
    end_time?: string;
  }): Promise<Blob> => {
    const response = await apiRequest.get('/tickets/export', {
      params,
      responseType: 'blob'
    });
    return response.data as Blob;
  },

  // 从告警创建工单
  createTicketFromAlert: async (alert_id: string, ticketData: {
    title?: string;
    description?: string;
    assignee_id?: string;
    priority: TicketPriority;
  }): Promise<Ticket> => {
    const response = await apiRequest.post<Ticket>('/tickets/from-alert', {
      alert_id,
      ...ticketData
    });
    return response.data;
  },
};