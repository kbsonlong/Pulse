import { format, parseISO, isValid } from 'date-fns';
import { zhCN } from 'date-fns/locale';

/**
 * 格式化日期
 */
export const formatDate = (
  date: string | Date,
  formatStr: string = 'yyyy-MM-dd HH:mm:ss'
): string => {
  try {
    const dateObj = typeof date === 'string' ? parseISO(date) : date;
    if (!isValid(dateObj)) {
      return '-';
    }
    return format(dateObj, formatStr, { locale: zhCN });
  } catch {
    return '-';
  }
};

/**
 * 格式化相对时间
 */
export const formatRelativeTime = (date: string | Date): string => {
  try {
    const dateObj = typeof date === 'string' ? parseISO(date) : date;
    if (!isValid(dateObj)) {
      return '-';
    }
    const now = new Date();
    const diff = now.getTime() - dateObj.getTime();
    const seconds = Math.floor(diff / 1000);
    const minutes = Math.floor(seconds / 60);
    const hours = Math.floor(minutes / 60);
    const days = Math.floor(hours / 24);

    if (days > 0) {
      return `${days}天前`;
    } else if (hours > 0) {
      return `${hours}小时前`;
    } else if (minutes > 0) {
      return `${minutes}分钟前`;
    } else {
      return '刚刚';
    }
  } catch {
    return '-';
  }
};

/**
 * 防抖函数
 */
export const debounce = <T extends (...args: any[]) => any>(
  func: T,
  wait: number
): ((...args: Parameters<T>) => void) => {
  let timeout: NodeJS.Timeout;
  return (...args: Parameters<T>) => {
    clearTimeout(timeout);
    timeout = setTimeout(() => func(...args), wait);
  };
};

/**
 * 节流函数
 */
export const throttle = <T extends (...args: any[]) => any>(
  func: T,
  wait: number
): ((...args: Parameters<T>) => void) => {
  let inThrottle: boolean;
  return (...args: Parameters<T>) => {
    if (!inThrottle) {
      func(...args);
      inThrottle = true;
      setTimeout(() => (inThrottle = false), wait);
    }
  };
};

/**
 * 深拷贝
 */
export const deepClone = <T>(obj: T): T => {
  if (obj === null || typeof obj !== 'object') {
    return obj;
  }
  if (obj instanceof Date) {
    return new Date(obj.getTime()) as unknown as T;
  }
  if (obj instanceof Array) {
    return obj.map((item) => deepClone(item)) as unknown as T;
  }
  if (typeof obj === 'object') {
    const clonedObj = {} as { [key: string]: any };
    for (const key in obj) {
      if (obj.hasOwnProperty(key)) {
        clonedObj[key] = deepClone(obj[key]);
      }
    }
    return clonedObj as T;
  }
  return obj;
};

/**
 * 生成唯一ID
 */
export const generateId = (): string => {
  return Math.random().toString(36).substr(2, 9);
};

/**
 * 格式化文件大小
 */
export const formatFileSize = (bytes: number): string => {
  if (bytes === 0) return '0 Bytes';
  const k = 1024;
  const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
};

/**
 * 获取文件扩展名
 */
export const getFileExtension = (filename: string): string => {
  return filename.slice(((filename.lastIndexOf('.') - 1) >>> 0) + 2);
};

/**
 * 验证邮箱格式
 */
export const isValidEmail = (email: string): boolean => {
  const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
  return emailRegex.test(email);
};

/**
 * 验证手机号格式
 */
export const isValidPhone = (phone: string): boolean => {
  const phoneRegex = /^1[3-9]\d{9}$/;
  return phoneRegex.test(phone);
};

/**
 * 获取URL参数
 */
export const getUrlParams = (url?: string): Record<string, string> => {
  const params: Record<string, string> = {};
  const searchParams = new URLSearchParams(url || window.location.search);
  searchParams.forEach((value, key) => {
    params[key] = value;
  });
  return params;
};

/**
 * 设置URL参数
 */
export const setUrlParams = (params: Record<string, string>): string => {
  const searchParams = new URLSearchParams();
  Object.entries(params).forEach(([key, value]) => {
    if (value) {
      searchParams.set(key, value);
    }
  });
  return searchParams.toString();
};

/**
 * 下载文件
 */
export const downloadFile = (blob: Blob, filename: string): void => {
  const url = window.URL.createObjectURL(blob);
  const link = document.createElement('a');
  link.href = url;
  link.download = filename;
  document.body.appendChild(link);
  link.click();
  document.body.removeChild(link);
  window.URL.revokeObjectURL(url);
};

/**
 * 复制到剪贴板
 */
export const copyToClipboard = async (text: string): Promise<boolean> => {
  try {
    if (navigator.clipboard) {
      await navigator.clipboard.writeText(text);
      return true;
    } else {
      // 降级方案
      const textArea = document.createElement('textarea');
      textArea.value = text;
      document.body.appendChild(textArea);
      textArea.select();
      document.execCommand('copy');
      document.body.removeChild(textArea);
      return true;
    }
  } catch {
    return false;
  }
};

/**
 * 获取随机颜色
 */
export const getRandomColor = (): string => {
  const colors = [
    '#f56565', '#ed8936', '#ecc94b', '#48bb78', '#38b2ac',
    '#4299e1', '#667eea', '#9f7aea', '#ed64a6', '#a0aec0'
  ];
  return colors[Math.floor(Math.random() * colors.length)];
};

/**
 * 获取状态颜色
 */
export const getStatusColor = (status: string): string => {
  const statusColors: Record<string, string> = {
    success: '#52c41a',
    error: '#ff4d4f',
    warning: '#faad14',
    info: '#1890ff',
    processing: '#1890ff',
    default: '#d9d9d9',
  };
  return statusColors[status] || statusColors.default;
};

/**
 * 获取优先级颜色
 */
export const getPriorityColor = (priority: string): string => {
  const priorityColors: Record<string, string> = {
    critical: '#ff4d4f',
    high: '#fa8c16',
    medium: '#faad14',
    low: '#52c41a',
    info: '#1890ff',
  };
  return priorityColors[priority.toLowerCase()] || priorityColors.info;
};