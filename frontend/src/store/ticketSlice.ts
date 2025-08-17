import { createSlice, createAsyncThunk, PayloadAction } from '@reduxjs/toolkit';
import { ticketService } from '../services/ticket';
import { Ticket, TicketStatus, TicketPriority, ProcessRecord, PaginatedResponse } from '../types';

interface TicketState {
  tickets: Ticket[];
  currentTicket: Ticket | null;
  processRecords: ProcessRecord[];
  total: number;
  page: number;
  limit: number;
  loading: boolean;
  error: string | null;
  filters: {
    status?: TicketStatus;
    priority?: TicketPriority;
    assignee_id?: string;
    search?: string;
    start_time?: string;
    end_time?: string;
  };
  statistics: {
    total: number;
    by_status: Record<TicketStatus, number>;
    by_priority: Record<TicketPriority, number>;
    by_assignee: Record<string, number>;
    resolution_time: {
      average: number;
      median: number;
    };
    trend: Array<{
      time: string;
      created: number;
      resolved: number;
    }>;
  } | null;
}

const initialState: TicketState = {
  tickets: [],
  currentTicket: null,
  processRecords: [],
  total: 0,
  page: 1,
  limit: 20,
  loading: false,
  error: null,
  filters: {},
  statistics: null,
};

// 异步thunks
export const fetchTickets = createAsyncThunk(
  'ticket/fetchTickets',
  async (params?: {
    page?: number;
    limit?: number;
    status?: TicketStatus;
    priority?: TicketPriority;
    assignee_id?: string;
    search?: string;
    start_time?: string;
    end_time?: string;
  }) => {
    const response = await ticketService.getTickets(params);
    return response;
  }
);

export const fetchTicket = createAsyncThunk(
  'ticket/fetchTicket',
  async (id: string) => {
    const response = await ticketService.getTicket(id);
    return response;
  }
);

export const createTicket = createAsyncThunk(
  'ticket/createTicket',
  async (ticketData: {
    title: string;
    description: string;
    alert_id?: string;
    assignee_id?: string;
    priority: TicketPriority;
  }) => {
    const response = await ticketService.createTicket(ticketData);
    return response;
  }
);

export const updateTicket = createAsyncThunk(
  'ticket/updateTicket',
  async ({
    id,
    ticketData,
  }: {
    id: string;
    ticketData: {
      title?: string;
      description?: string;
      assignee_id?: string;
      priority?: TicketPriority;
      status?: TicketStatus;
    };
  }) => {
    const response = await ticketService.updateTicket(id, ticketData);
    return response;
  }
);

export const deleteTicket = createAsyncThunk(
  'ticket/deleteTicket',
  async (id: string) => {
    await ticketService.deleteTicket(id);
    return id;
  }
);

export const updateTicketStatus = createAsyncThunk(
  'ticket/updateTicketStatus',
  async ({
    id,
    status,
    comment,
  }: {
    id: string;
    status: TicketStatus;
    comment?: string;
  }) => {
    const response = await ticketService.updateTicketStatus(id, status, comment);
    return response;
  }
);

export const assignTicket = createAsyncThunk(
  'ticket/assignTicket',
  async ({
    id,
    assignee_id,
    comment,
  }: {
    id: string;
    assignee_id: string;
    comment?: string;
  }) => {
    const response = await ticketService.assignTicket(id, assignee_id, comment);
    return response;
  }
);

export const fetchProcessRecords = createAsyncThunk(
  'ticket/fetchProcessRecords',
  async ({
    id,
    params,
  }: {
    id: string;
    params?: {
      page?: number;
      limit?: number;
    };
  }) => {
    const response = await ticketService.getTicketProcessRecords(id, params);
    return response;
  }
);

export const addProcessRecord = createAsyncThunk(
  'ticket/addProcessRecord',
  async ({
    id,
    recordData,
  }: {
    id: string;
    recordData: {
      action: string;
      description: string;
    };
  }) => {
    const response = await ticketService.addProcessRecord(id, recordData);
    return response;
  }
);

export const fetchTicketStatistics = createAsyncThunk(
  'ticket/fetchTicketStatistics',
  async (params?: {
    start_time?: string;
    end_time?: string;
    assignee_id?: string;
  }) => {
    const response = await ticketService.getTicketStatistics(params);
    return response;
  }
);

export const searchTickets = createAsyncThunk(
  'ticket/searchTickets',
  async (params: {
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
  }) => {
    const response = await ticketService.searchTickets(params);
    return response;
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
    clearProcessRecords: (state) => {
      state.processRecords = [];
    },
  },
  extraReducers: (builder) => {
    builder
      // fetchTickets
      .addCase(fetchTickets.pending, (state) => {
        state.loading = true;
        state.error = null;
      })
      .addCase(fetchTickets.fulfilled, (state, action) => {
        state.loading = false;
        state.tickets = action.payload.data;
        state.total = action.payload.total;
        state.page = action.payload.page;
        state.limit = action.payload.limit;
      })
      .addCase(fetchTickets.rejected, (state, action) => {
        state.loading = false;
        state.error = action.error.message || '获取工单列表失败';
      })
      // fetchTicket
      .addCase(fetchTicket.pending, (state) => {
        state.loading = true;
        state.error = null;
      })
      .addCase(fetchTicket.fulfilled, (state, action) => {
        state.loading = false;
        state.currentTicket = action.payload;
      })
      .addCase(fetchTicket.rejected, (state, action) => {
        state.loading = false;
        state.error = action.error.message || '获取工单详情失败';
      })
      // createTicket
      .addCase(createTicket.pending, (state) => {
        state.loading = true;
        state.error = null;
      })
      .addCase(createTicket.fulfilled, (state, action) => {
        state.loading = false;
        state.tickets.unshift(action.payload);
        state.total += 1;
      })
      .addCase(createTicket.rejected, (state, action) => {
        state.loading = false;
        state.error = action.error.message || '创建工单失败';
      })
      // updateTicket
      .addCase(updateTicket.pending, (state) => {
        state.loading = true;
        state.error = null;
      })
      .addCase(updateTicket.fulfilled, (state, action) => {
        state.loading = false;
        const index = state.tickets.findIndex(ticket => ticket.id === action.payload.id);
        if (index !== -1) {
          state.tickets[index] = action.payload;
        }
        if (state.currentTicket?.id === action.payload.id) {
          state.currentTicket = action.payload;
        }
      })
      .addCase(updateTicket.rejected, (state, action) => {
        state.loading = false;
        state.error = action.error.message || '更新工单失败';
      })
      // deleteTicket
      .addCase(deleteTicket.pending, (state) => {
        state.loading = true;
        state.error = null;
      })
      .addCase(deleteTicket.fulfilled, (state, action) => {
        state.loading = false;
        state.tickets = state.tickets.filter(ticket => ticket.id !== action.payload);
        state.total -= 1;
        if (state.currentTicket?.id === action.payload) {
          state.currentTicket = null;
        }
      })
      .addCase(deleteTicket.rejected, (state, action) => {
        state.loading = false;
        state.error = action.error.message || '删除工单失败';
      })
      // updateTicketStatus
      .addCase(updateTicketStatus.fulfilled, (state, action) => {
        const index = state.tickets.findIndex(ticket => ticket.id === action.payload.id);
        if (index !== -1) {
          state.tickets[index] = action.payload;
        }
        if (state.currentTicket?.id === action.payload.id) {
          state.currentTicket = action.payload;
        }
      })
      // assignTicket
      .addCase(assignTicket.fulfilled, (state, action) => {
        const index = state.tickets.findIndex(ticket => ticket.id === action.payload.id);
        if (index !== -1) {
          state.tickets[index] = action.payload;
        }
        if (state.currentTicket?.id === action.payload.id) {
          state.currentTicket = action.payload;
        }
      })
      // fetchProcessRecords
      .addCase(fetchProcessRecords.fulfilled, (state, action) => {
        state.processRecords = action.payload.data;
      })
      // addProcessRecord
      .addCase(addProcessRecord.fulfilled, (state, action) => {
        state.processRecords.unshift(action.payload);
      })
      // fetchTicketStatistics
      .addCase(fetchTicketStatistics.fulfilled, (state, action) => {
        state.statistics = action.payload;
      })
      // searchTickets
      .addCase(searchTickets.fulfilled, (state, action) => {
        state.tickets = action.payload.data;
        state.total = action.payload.total;
        state.page = action.payload.page;
        state.limit = action.payload.limit;
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
  clearProcessRecords,
} = ticketSlice.actions;

export default ticketSlice.reducer;