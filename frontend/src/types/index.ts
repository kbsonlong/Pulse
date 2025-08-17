// 用户相关类型
export interface User {
  id: string;
  username: string;
  email: string;
  role: string;
  created_at: string;
  updated_at: string;
}

export interface LoginRequest {
  username: string;
  password: string;
}

export interface LoginResponse {
  token: string;
  user: User;
}

// 告警相关类型
export interface Alert {
  id: string;
  source: string;
  level: AlertLevel;
  title: string;
  description: string;
  labels: Record<string, string>;
  annotations: Record<string, string>;
  status: AlertStatus;
  starts_at: string;
  ends_at?: string;
  created_at: string;
  updated_at: string;
}

export enum AlertLevel {
  INFO = 'info',
  WARNING = 'warning',
  ERROR = 'error',
  CRITICAL = 'critical'
}

export enum AlertStatus {
  FIRING = 'firing',
  RESOLVED = 'resolved',
  SUPPRESSED = 'suppressed'
}

export interface AlertQuery {
  page?: number;
  limit?: number;
  source?: string;
  level?: AlertLevel;
  status?: AlertStatus;
  start_time?: string;
  end_time?: string;
  search?: string;
}

export interface AlertList {
  alerts: Alert[];
  total: number;
  page: number;
  limit: number;
}

// 规则相关类型
export interface Rule {
  id: string;
  name: string;
  description: string;
  data_source_id: string;
  query: string;
  conditions: RuleCondition[];
  actions: RuleAction[];
  enabled: boolean;
  created_at: string;
  updated_at: string;
}

export interface RuleCondition {
  field: string;
  operator: string;
  value: string;
}

export interface RuleAction {
  type: string;
  config: Record<string, any>;
}

// 数据源相关类型
export interface DataSource {
  id: string;
  name: string;
  type: string;
  url: string;
  config: Record<string, any>;
  enabled: boolean;
  created_at: string;
  updated_at: string;
}

// 工单相关类型
export interface Ticket {
  id: string;
  title: string;
  description: string;
  alert_id?: string;
  assignee_id?: string;
  status: TicketStatus;
  priority: TicketPriority;
  created_by: string;
  created_at: string;
  updated_at: string;
}

export enum TicketStatus {
  OPEN = 'open',
  IN_PROGRESS = 'in_progress',
  RESOLVED = 'resolved',
  CLOSED = 'closed'
}

export enum TicketPriority {
  LOW = 'low',
  MEDIUM = 'medium',
  HIGH = 'high',
  URGENT = 'urgent'
}

export interface ProcessRecord {
  id: string;
  ticket_id: string;
  user_id: string;
  action: string;
  description: string;
  created_at: string;
}

// 知识库相关类型
export interface Knowledge {
  id: string;
  title: string;
  content: string;
  tags: string[];
  category: string;
  score: number;
  created_by: string;
  created_at: string;
  updated_at: string;
}

// API响应类型
export interface ApiResponse<T> {
  code: number;
  message: string;
  data: T;
}

export interface PaginatedResponse<T> {
  items: T[];
  total: number;
  page: number;
  limit: number;
}

// 统计相关类型
export interface AlertStatistics {
  total: number;
  by_level: Record<AlertLevel, number>;
  by_status: Record<AlertStatus, number>;
  by_source: Record<string, number>;
  trend: Array<{
    time: string;
    count: number;
  }>;
}