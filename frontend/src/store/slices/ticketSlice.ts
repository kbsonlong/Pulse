import { createSlice, createAsyncThunk, PayloadAction } from '@reduxjs/toolkit';
import {
  Ticket,
  TicketStatus,
  TicketPriority,
  ProcessRecord,
  PaginatedResponse,
  User
} from '../../types';
import { ticketService } from '../../services/ticket';

interface TicketState {
  tickets: Ticket[];
  currentTicket: Ticket | null;
  processRecords: ProcessRecord[];
  assignableUsers: User[];
  total: number;
  page: number;
  limit: number;
  loading: boolean;
  error: string | null;
  filters: {
    status?: TicketStatus;
    priority?: TicketPriority;
    assignee_id?: string;
    created_by?: string;
    search?: string;
    start_time?: string;
    end_time?: string;
  };
}

const initialState: TicketState = {
  tickets: [],
  currentTicket: null,
  processRecords: [],
  assignableUsers: [],
  total: 0,
  page: 1,
  limit: 20,
  loading: false,
  error: null,
  filters: {},
};

export const fetchTickets = createAsyncThunk(
  'ticket/fetchTickets',
  async (params: any, { rejectWithValue }) => {
    try {
      const response = await ticketService.getTickets(params);
      return response;
    } catch (error: any) {
      return rejectWithValue(error.response?.data?.message || '获取工单列表失败');
    }
  }
);

export const fetchTicket = createAsyncThunk(
  'ticket/fetchTicket',
  async (id: string, { rejectWithValue }) => {
    try {
      const ticket = await ticketService.getTicket(id);
      return ticket;
    } catch (error: any) {
      return rejectWithValue(error.response?.data?.message || '获取工单详情失败');
    }
  }
);

export const createTicket = createAsyncThunk(
  'ticket/createTicket',
  async (ticketData: any, { rejectWithValue }) => {
    try {
      const ticket = await ticketService.createTicket(ticketData);
      return ticket;
    } catch (error: any) {
      return rejectWithValue(error.response?.data?.message || '创建工单失败');
    }
  }
);

export const updateTicket = createAsyncThunk(
  'ticket/updateTicket',
  async ({ id, data }: { id: string; data: any }, { rejectWithValue }) => {
    try {
      const ticket = await ticketService.updateTicket(id, data);
      return ticket;
    } catch (error: any) {
      return rejectWithValue(error.response?.data?.message || '更新工单失败');
    }
  }
);

export const updateTicketStatus = createAsyncThunk(
  'ticket/updateStatus',
  async ({ id, status, comment }: { id: string; status: TicketStatus; comment?: string }, { rejectWithValue }) => {
    try {
      const ticket = await ticketService.updateTicketStatus(id, status, comment);
      return ticket;
    } catch (error: any) {
      return rejectWithValue(error.response?.data?.message || '更新工单状态失败');
    }
  }
);

export const assignTicket = createAsyncThunk(
  'ticket/assignTicket',
  async ({ id, assignee_id, comment }: { id: string; assignee_id: string; comment?: string }, { rejectWithValue }) => {
    try {
      const ticket = await ticketService.assignTicket(id, assignee_id, comment);
      return ticket;
    } catch (error: any) {
      return rejectWithValue(error.response?.data?.message || '分配工单失败');
    }
  }
);

export const fetchProcessRecords = createAsyncThunk(
  'ticket/fetchProcessRecords',
  async ({ id, params }: { id: string; params?: any }, { rejectWithValue }) => {
    try {
      const response = await ticketService.getTicketProcessRecords(id, params);
      return response.items;
    } catch (error: any) {
      return rejectWithValue(error.response?.data?.message || '获取处理记录失败');
    }
  }
);

export const addProcessRecord = createAsyncThunk(
  'ticket/addProcessRecord',
  async ({ id, data }: { id: string; data: any }, { rejectWithValue }) => {
    try {
      const record = await ticketService.addProcessRecord(id, data);
      return record;
    } catch (error: any) {
      return rejectWithValue(error.response?.data?.message || '添加处理记录失败');
    }
  }
);

export const fetchAssignableUsers = createAsyncThunk(
  'ticket/fetchAssignableUsers',
  async (_, { rejectWithValue }) => {
    try {
      const users = await ticketService.getAssignableUsers();
      return users;
    } catch (error: any) {
      return rejectWithValue(error.response?.data?.message || '获取可分配用户失败');
    }
  }
);

const ticketSlice = createSlice({
  name: 'ticket',
  initialState,
  reducers: {
    setFilters: (state, action: PayloadAction<Partial<TicketState['filters']>>) => {
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
    clearCurrentTicket: (state) => {
      state.currentTicket = null;
    },
  },
  extraReducers: (builder) => {
    builder
      .addCase(fetchTickets.pending, (state) => {
        state.loading = true;
        state.error = null;
      })
      .addCase(fetchTickets.fulfilled, (state, action) => {
        state.loading = false;
        state.tickets = action.payload.items;
        state.total = action.payload.total;
        state.page = action.payload.page;
        state.limit = action.payload.limit;
      })
      .addCase(fetchTickets.rejected, (state, action) => {
        state.loading = false;
        state.error = action.payload as string;
      })
      .addCase(fetchTicket.fulfilled, (state, action) => {
        state.currentTicket = action.payload;
      })
      .addCase(createTicket.fulfilled, (state, action) => {
        state.tickets.unshift(action.payload);
        state.total += 1;
      })
      .addCase(updateTicket.fulfilled, (state, action) => {
        const index = state.tickets.findIndex(ticket => ticket.id === action.payload.id);
        if (index !== -1) {
          state.tickets[index] = action.payload;
        }
        if (state.currentTicket?.id === action.payload.id) {
          state.currentTicket = action.payload;
        }
      })
      .addCase(updateTicketStatus.fulfilled, (state, action) => {
        const index = state.tickets.findIndex(ticket => ticket.id === action.payload.id);
        if (index !== -1) {
          state.tickets[index] = action.payload;
        }
        if (state.currentTicket?.id === action.payload.id) {
          state.currentTicket = action.payload;
        }
      })
      .addCase(assignTicket.fulfilled, (state, action) => {
        const index = state.tickets.findIndex(ticket => ticket.id === action.payload.id);
        if (index !== -1) {
          state.tickets[index] = action.payload;
        }
        if (state.currentTicket?.id === action.payload.id) {
          state.currentTicket = action.payload;
        }
      })
      .addCase(fetchProcessRecords.fulfilled, (state, action) => {
        state.processRecords = action.payload;
      })
      .addCase(addProcessRecord.fulfilled, (state, action) => {
        state.processRecords.unshift(action.payload);
      })
      .addCase(fetchAssignableUsers.fulfilled, (state, action) => {
        state.assignableUsers = action.payload;
      });
  },
});

export const {
  setFilters,
  clearFilters,
  setPage,
  setLimit,
  clearError,
  clearCurrentTicket,
} = ticketSlice.actions;

export default ticketSlice.reducer;