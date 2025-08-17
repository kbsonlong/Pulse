import { createSlice, createAsyncThunk, PayloadAction } from '@reduxjs/toolkit';
import { Knowledge, PaginatedResponse } from '../../types';
import { knowledgeService } from '../../services/knowledge';

interface KnowledgeState {
  knowledgeList: Knowledge[];
  currentKnowledge: Knowledge | null;
  categories: Array<{ name: string; count: number }>;
  tags: Array<{ name: string; count: number }>;
  total: number;
  page: number;
  limit: number;
  loading: boolean;
  error: string | null;
  filters: {
    category?: string;
    tags?: string[];
    search?: string;
    sort_by?: 'created_at' | 'updated_at' | 'score' | 'title';
    sort_order?: 'asc' | 'desc';
  };
}

const initialState: KnowledgeState = {
  knowledgeList: [],
  currentKnowledge: null,
  categories: [],
  tags: [],
  total: 0,
  page: 1,
  limit: 20,
  loading: false,
  error: null,
  filters: {},
};

export const fetchKnowledgeList = createAsyncThunk(
  'knowledge/fetchKnowledgeList',
  async (params?: any, { rejectWithValue }) => {
    try {
      const response = await knowledgeService.getKnowledgeList(params);
      return response;
    } catch (error: any) {
      return rejectWithValue(error.response?.data?.message || '获取知识库列表失败');
    }
  }
);

export const fetchKnowledge = createAsyncThunk(
  'knowledge/fetchKnowledge',
  async (id: string, { rejectWithValue }) => {
    try {
      const knowledge = await knowledgeService.getKnowledge(id);
      return knowledge;
    } catch (error: any) {
      return rejectWithValue(error.response?.data?.message || '获取知识库详情失败');
    }
  }
);

export const createKnowledge = createAsyncThunk(
  'knowledge/createKnowledge',
  async (knowledgeData: any, { rejectWithValue }) => {
    try {
      const knowledge = await knowledgeService.createKnowledge(knowledgeData);
      return knowledge;
    } catch (error: any) {
      return rejectWithValue(error.response?.data?.message || '创建知识库文档失败');
    }
  }
);

export const updateKnowledge = createAsyncThunk(
  'knowledge/updateKnowledge',
  async ({ id, data }: { id: string; data: any }, { rejectWithValue }) => {
    try {
      const knowledge = await knowledgeService.updateKnowledge(id, data);
      return knowledge;
    } catch (error: any) {
      return rejectWithValue(error.response?.data?.message || '更新知识库文档失败');
    }
  }
);

export const deleteKnowledge = createAsyncThunk(
  'knowledge/deleteKnowledge',
  async (id: string, { rejectWithValue }) => {
    try {
      await knowledgeService.deleteKnowledge(id);
      return id;
    } catch (error: any) {
      return rejectWithValue(error.response?.data?.message || '删除知识库文档失败');
    }
  }
);

export const searchKnowledge = createAsyncThunk(
  'knowledge/searchKnowledge',
  async (params: any, { rejectWithValue }) => {
    try {
      const response = await knowledgeService.searchKnowledge(params);
      return response;
    } catch (error: any) {
      return rejectWithValue(error.response?.data?.message || '搜索知识库失败');
    }
  }
);

export const fetchCategories = createAsyncThunk(
  'knowledge/fetchCategories',
  async (_, { rejectWithValue }) => {
    try {
      const categories = await knowledgeService.getCategories();
      return categories;
    } catch (error: any) {
      return rejectWithValue(error.response?.data?.message || '获取分类列表失败');
    }
  }
);

export const fetchTags = createAsyncThunk(
  'knowledge/fetchTags',
  async (category?: string, { rejectWithValue }) => {
    try {
      const tags = await knowledgeService.getTags(category);
      return tags;
    } catch (error: any) {
      return rejectWithValue(error.response?.data?.message || '获取标签列表失败');
    }
  }
);

const knowledgeSlice = createSlice({
  name: 'knowledge',
  initialState,
  reducers: {
    setFilters: (state, action: PayloadAction<Partial<KnowledgeState['filters']>>) => {
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
    clearCurrentKnowledge: (state) => {
      state.currentKnowledge = null;
    },
  },
  extraReducers: (builder) => {
    builder
      .addCase(fetchKnowledgeList.pending, (state) => {
        state.loading = true;
        state.error = null;
      })
      .addCase(fetchKnowledgeList.fulfilled, (state, action) => {
        state.loading = false;
        state.knowledgeList = action.payload.items;
        state.total = action.payload.total;
        state.page = action.payload.page;
        state.limit = action.payload.limit;
      })
      .addCase(fetchKnowledgeList.rejected, (state, action) => {
        state.loading = false;
        state.error = action.payload as string;
      })
      .addCase(fetchKnowledge.fulfilled, (state, action) => {
        state.currentKnowledge = action.payload;
      })
      .addCase(createKnowledge.fulfilled, (state, action) => {
        state.knowledgeList.unshift(action.payload);
        state.total += 1;
      })
      .addCase(updateKnowledge.fulfilled, (state, action) => {
        const index = state.knowledgeList.findIndex(item => item.id === action.payload.id);
        if (index !== -1) {
          state.knowledgeList[index] = action.payload;
        }
        if (state.currentKnowledge?.id === action.payload.id) {
          state.currentKnowledge = action.payload;
        }
      })
      .addCase(deleteKnowledge.fulfilled, (state, action) => {
        state.knowledgeList = state.knowledgeList.filter(item => item.id !== action.payload);
        state.total -= 1;
      })
      .addCase(searchKnowledge.pending, (state) => {
        state.loading = true;
        state.error = null;
      })
      .addCase(searchKnowledge.fulfilled, (state, action) => {
        state.loading = false;
        state.knowledgeList = action.payload.items;
        state.total = action.payload.total;
        state.page = action.payload.page;
        state.limit = action.payload.limit;
      })
      .addCase(searchKnowledge.rejected, (state, action) => {
        state.loading = false;
        state.error = action.payload as string;
      })
      .addCase(fetchCategories.fulfilled, (state, action) => {
        state.categories = action.payload;
      })
      .addCase(fetchTags.fulfilled, (state, action) => {
        state.tags = action.payload;
      });
  },
});

export const {
  setFilters,
  clearFilters,
  setPage,
  setLimit,
  clearError,
  clearCurrentKnowledge,
} = knowledgeSlice.actions;

export default knowledgeSlice.reducer;