import { createSlice, createAsyncThunk, PayloadAction } from '@reduxjs/toolkit';
import {
  Alert,
  AlertQuery,
  AlertList,
  AlertStatistics,
  AlertStatus,
  AlertLevel
} from '../../types';
import { alertService } from '../../services/alert';

interface AlertState {
  alerts: Alert[];
  currentAlert: Alert | null;
  statistics: AlertStatistics | null;
  total: number;
  page: number;
  limit: number;
  loading: boolean;
  error: string | null;
  filters: {
    source?: string;
    level?: AlertLevel;
    status?: AlertStatus;
    start_time?: string;
    end_time?: string;
    search?: string;
  };
}

const initialState: AlertState = {
  alerts: [],
  currentAlert: null,
  statistics: null,
  total: 0,
  page: 1,
  limit: 20,
  loading: false,
  error: null,
  filters: {},
};

// 异步actions
export const fetchAlerts = createAsyncThunk(
  'alert/fetchAlerts',
  async (query?: AlertQuery, { rejectWithValue }) => {
    try {
      const response = await alertService.getAlerts(query);
      return response;
    } catch (error: any) {
      return rejectWithValue(error.response?.data?.message || '获取告警列表失败');
    }
  }
);

export const fetchAlert = createAsyncThunk(
  'alert/fetchAlert',
  async (id: string, { rejectWithValue }) => {
    try {
      const alert = await alertService.getAlert(id);
      return alert;
    } catch (error: any) {
      return rejectWithValue(error.response?.data?.message || '获取告警详情失败');
    }
  }
);

export const updateAlertStatus = createAsyncThunk(
  'alert/updateStatus',
  async ({ id, status }: { id: string; status: AlertStatus }, { rejectWithValue }) => {
    try {
      const alert = await alertService.updateAlertStatus(id, status);
      return alert;
    } catch (error: any) {
      return rejectWithValue(error.response?.data?.message || '更新告警状态失败');
    }
  }
);

export const batchUpdateAlertStatus = createAsyncThunk(
  'alert/batchUpdateStatus',
  async ({ ids, status }: { ids: string[]; status: AlertStatus }, { rejectWithValue }) => {
    try {
      await alertService.batchUpdateStatus(ids, status);
      return { ids, status };
    } catch (error: any) {
      return rejectWithValue(error.response?.data?.message || '批量更新告警状态失败');
    }
  }
);

export const deleteAlert = createAsyncThunk(
  'alert/deleteAlert',
  async (id: string, { rejectWithValue }) => {
    try {
      await alertService.deleteAlert(id);
      return id;
    } catch (error: any) {
      return rejectWithValue(error.response?.data?.message || '删除告警失败');
    }
  }
);

export const fetchStatistics = createAsyncThunk(
  'alert/fetchStatistics',
  async (params?: {
    start_time?: string;
    end_time?: string;
    source?: string;
  }, { rejectWithValue }) => {
    try {
      const statistics = await alertService.getStatistics(params);
      return statistics;
    } catch (error: any) {
      return rejectWithValue(error.response?.data?.message || '获取统计信息失败');
    }
  }
);

export const searchAlerts = createAsyncThunk(
  'alert/searchAlerts',
  async (params: {
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
  }, { rejectWithValue }) => {
    try {
      const response = await alertService.searchAlerts(params);
      return response;
    } catch (error: any) {
      return rejectWithValue(error.response?.data?.message || '搜索告警失败');
    }
  }
);

const alertSlice = createSlice({
  name: 'alert',
  initialState,
  reducers: {
    setFilters: (state, action: PayloadAction<Partial<AlertState['filters']>>) => {
      state.filters = { ...state.filters, ...action.payload };
    },
    clearFilters: (state) => {
      state.filters = {};
    },
    setPage: (state, action: PayloadAction<number>) => {
      state.page = action.payload;
    },
    setLimit: (state, action: PayloadAction<number>) => {
      state.limit = action.payload;
    },
    clearError: (state) => {
      state.error = null;
    },
    clearCurrentAlert: (state) => {
      state.currentAlert = null;
    },
  },
  extraReducers: (builder) => {
    builder
      // 获取告警列表
      .addCase(fetchAlerts.pending, (state) => {
        state.loading = true;
        state.error = null;
      })
      .addCase(fetchAlerts.fulfilled, (state, action) => {
        state.loading = false;
        state.alerts = action.payload.alerts;
        state.total = action.payload.total;
        state.page = action.payload.page;
        state.limit = action.payload.limit;
      })
      .addCase(fetchAlerts.rejected, (state, action) => {
        state.loading = false;
        state.error = action.payload as string;
      })
      // 获取告警详情
      .addCase(fetchAlert.pending, (state) => {
        state.loading = true;
        state.error = null;
      })
      .addCase(fetchAlert.fulfilled, (state, action) => {
        state.loading = false;
        state.currentAlert = action.payload;
      })
      .addCase(fetchAlert.rejected, (state, action) => {
        state.loading = false;
        state.error = action.payload as string;
      })
      // 更新告警状态
      .addCase(updateAlertStatus.fulfilled, (state, action) => {
        const index = state.alerts.findIndex(alert => alert.id === action.payload.id);
        if (index !== -1) {
          state.alerts[index] = action.payload;
        }
        if (state.currentAlert?.id === action.payload.id) {
          state.currentAlert = action.payload;
        }
      })
      .addCase(updateAlertStatus.rejected, (state, action) => {
        state.error = action.payload as string;
      })
      // 批量更新告警状态
      .addCase(batchUpdateAlertStatus.fulfilled, (state, action) => {
        const { ids, status } = action.payload;
        state.alerts = state.alerts.map(alert => 
          ids.includes(alert.id) ? { ...alert, status } : alert
        );
      })
      .addCase(batchUpdateAlertStatus.rejected, (state, action) => {
        state.error = action.payload as string;
      })
      // 删除告警
      .addCase(deleteAlert.fulfilled, (state, action) => {
        state.alerts = state.alerts.filter(alert => alert.id !== action.payload);
        state.total -= 1;
        if (state.currentAlert?.id === action.payload) {
          state.currentAlert = null;
        }
      })
      .addCase(deleteAlert.rejected, (state, action) => {
        state.error = action.payload as string;
      })
      // 获取统计信息
      .addCase(fetchStatistics.fulfilled, (state, action) => {
        state.statistics = action.payload;
      })
      .addCase(fetchStatistics.rejected, (state, action) => {
        state.error = action.payload as string;
      })
      // 搜索告警
      .addCase(searchAlerts.pending, (state) => {
        state.loading = true;
        state.error = null;
      })
      .addCase(searchAlerts.fulfilled, (state, action) => {
        state.loading = false;
        state.alerts = action.payload.alerts;
        state.total = action.payload.total;
        state.page = action.payload.page;
        state.limit = action.payload.limit;
      })
      .addCase(searchAlerts.rejected, (state, action) => {
        state.loading = false;
        state.error = action.payload as string;
      });
  },
});

export const {
  setFilters,
  clearFilters,
  setPage,
  setLimit,
  clearError,
  clearCurrentAlert,
} = alertSlice.actions;

export default alertSlice.reducer;