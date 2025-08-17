import { createSlice, createAsyncThunk, PayloadAction } from '@reduxjs/toolkit';
import { Rule, DataSource, PaginatedResponse } from '../../types';
import { ruleService, dataSourceService } from '../../services/rule';

interface RuleState {
  rules: Rule[];
  currentRule: Rule | null;
  dataSources: DataSource[];
  total: number;
  page: number;
  limit: number;
  loading: boolean;
  error: string | null;
  filters: {
    search?: string;
    enabled?: boolean;
    data_source_id?: string;
  };
}

const initialState: RuleState = {
  rules: [],
  currentRule: null,
  dataSources: [],
  total: 0,
  page: 1,
  limit: 20,
  loading: false,
  error: null,
  filters: {},
};

export const fetchRules = createAsyncThunk(
  'rule/fetchRules',
  async (params?: {
    page?: number;
    limit?: number;
    search?: string;
    enabled?: boolean;
    data_source_id?: string;
  }, { rejectWithValue }) => {
    try {
      const response = await ruleService.getRules(params);
      return response;
    } catch (error: any) {
      return rejectWithValue(error.response?.data?.message || '获取规则列表失败');
    }
  }
);

export const fetchRule = createAsyncThunk(
  'rule/fetchRule',
  async (id: string, { rejectWithValue }) => {
    try {
      const rule = await ruleService.getRule(id);
      return rule;
    } catch (error: any) {
      return rejectWithValue(error.response?.data?.message || '获取规则详情失败');
    }
  }
);

export const createRule = createAsyncThunk(
  'rule/createRule',
  async (ruleData: any, { rejectWithValue }) => {
    try {
      const rule = await ruleService.createRule(ruleData);
      return rule;
    } catch (error: any) {
      return rejectWithValue(error.response?.data?.message || '创建规则失败');
    }
  }
);

export const updateRule = createAsyncThunk(
  'rule/updateRule',
  async ({ id, data }: { id: string; data: Partial<Rule> }, { rejectWithValue }) => {
    try {
      const rule = await ruleService.updateRule(id, data);
      return rule;
    } catch (error: any) {
      return rejectWithValue(error.response?.data?.message || '更新规则失败');
    }
  }
);

export const deleteRule = createAsyncThunk(
  'rule/deleteRule',
  async (id: string, { rejectWithValue }) => {
    try {
      await ruleService.deleteRule(id);
      return id;
    } catch (error: any) {
      return rejectWithValue(error.response?.data?.message || '删除规则失败');
    }
  }
);

export const toggleRule = createAsyncThunk(
  'rule/toggleRule',
  async ({ id, enabled }: { id: string; enabled: boolean }, { rejectWithValue }) => {
    try {
      const rule = await ruleService.toggleRule(id, enabled);
      return rule;
    } catch (error: any) {
      return rejectWithValue(error.response?.data?.message || '切换规则状态失败');
    }
  }
);

export const fetchDataSources = createAsyncThunk(
  'rule/fetchDataSources',
  async (_, { rejectWithValue }) => {
    try {
      const response = await dataSourceService.getDataSources();
      return response.items;
    } catch (error: any) {
      return rejectWithValue(error.response?.data?.message || '获取数据源列表失败');
    }
  }
);

const ruleSlice = createSlice({
  name: 'rule',
  initialState,
  reducers: {
    setFilters: (state, action: PayloadAction<Partial<RuleState['filters']>>) => {
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
    clearCurrentRule: (state) => {
      state.currentRule = null;
    },
  },
  extraReducers: (builder) => {
    builder
      .addCase(fetchRules.pending, (state) => {
        state.loading = true;
        state.error = null;
      })
      .addCase(fetchRules.fulfilled, (state, action) => {
        state.loading = false;
        state.rules = action.payload.items;
        state.total = action.payload.total;
        state.page = action.payload.page;
        state.limit = action.payload.limit;
      })
      .addCase(fetchRules.rejected, (state, action) => {
        state.loading = false;
        state.error = action.payload as string;
      })
      .addCase(fetchRule.fulfilled, (state, action) => {
        state.currentRule = action.payload;
      })
      .addCase(createRule.fulfilled, (state, action) => {
        state.rules.unshift(action.payload);
        state.total += 1;
      })
      .addCase(updateRule.fulfilled, (state, action) => {
        const index = state.rules.findIndex(rule => rule.id === action.payload.id);
        if (index !== -1) {
          state.rules[index] = action.payload;
        }
        if (state.currentRule?.id === action.payload.id) {
          state.currentRule = action.payload;
        }
      })
      .addCase(deleteRule.fulfilled, (state, action) => {
        state.rules = state.rules.filter(rule => rule.id !== action.payload);
        state.total -= 1;
      })
      .addCase(toggleRule.fulfilled, (state, action) => {
        const index = state.rules.findIndex(rule => rule.id === action.payload.id);
        if (index !== -1) {
          state.rules[index] = action.payload;
        }
      })
      .addCase(fetchDataSources.fulfilled, (state, action) => {
        state.dataSources = action.payload;
      });
  },
});

export const {
  setFilters,
  clearFilters,
  setPage,
  setLimit,
  clearError,
  clearCurrentRule,
} = ruleSlice.actions;

export default ruleSlice.reducer;