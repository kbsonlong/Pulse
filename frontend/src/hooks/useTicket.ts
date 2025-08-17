import { useCallback } from 'react';
import { useAppDispatch, useAppSelector } from './index';
import {
  fetchTickets,
  fetchTicket,
  createTicket,
  updateTicket,
  updateTicketStatus,
  assignTicket,
  fetchProcessRecords,
  addProcessRecord,
  setFilters,
  clearFilters,
  setPage,
  setLimit,
  clearError,
  clearCurrentTicket,
} from '../store/ticketSlice';
import { TicketStatus, TicketPriority } from '../types';

const useTicket = () => {
  const dispatch = useAppDispatch();
  const {
    tickets,
    currentTicket,
    processRecords,
    total,
    page,
    limit,
    loading,
    error,
    filters,
  } = useAppSelector((state) => state.ticket);

  const handleFetchTickets = useCallback(
    (params?: any) => {
      return dispatch(fetchTickets(params));
    },
    [dispatch]
  );

  const handleFetchTicket = useCallback(
    (id: string) => {
      return dispatch(fetchTicket(id));
    },
    [dispatch]
  );

  const handleCreateTicket = useCallback(
    (ticketData: any) => {
      return dispatch(createTicket(ticketData));
    },
    [dispatch]
  );

  const handleUpdateTicket = useCallback(
    (id: string, ticketData: any) => {
      return dispatch(updateTicket({ id, ticketData }));
    },
    [dispatch]
  );

  const handleUpdateStatus = useCallback(
    (id: string, status: TicketStatus, comment?: string) => {
      return dispatch(updateTicketStatus({ id, status, comment }));
    },
    [dispatch]
  );

  const handleAssignTicket = useCallback(
    (id: string, assignee_id: string, comment?: string) => {
      return dispatch(assignTicket({ id, assignee_id, comment }));
    },
    [dispatch]
  );

  const handleFetchProcessRecords = useCallback(
    (id: string, params?: any) => {
      return dispatch(fetchProcessRecords({ id, params }));
    },
    [dispatch]
  );

  const handleAddProcessRecord = useCallback(
    (id: string, recordData: any) => {
      return dispatch(addProcessRecord({ id, recordData }));
    },
    [dispatch]
  );



  const handleSetFilters = useCallback(
    (newFilters: Partial<typeof filters>) => {
      dispatch(setFilters(newFilters));
    },
    [dispatch]
  );

  const handleClearFilters = useCallback(() => {
    dispatch(clearFilters());
  }, [dispatch]);

  const handleSetPage = useCallback(
    (newPage: number) => {
      dispatch(setPage(newPage));
    },
    [dispatch]
  );

  const handleSetLimit = useCallback(
    (newLimit: number) => {
      dispatch(setLimit(newLimit));
    },
    [dispatch]
  );

  const handleClearError = useCallback(() => {
    dispatch(clearError());
  }, [dispatch]);

  const handleClearCurrentTicket = useCallback(() => {
    dispatch(clearCurrentTicket());
  }, [dispatch]);

  return {
    // 状态
    tickets,
    currentTicket,
    processRecords,
    total,
    page,
    limit,
    loading,
    error,
    filters,
    
    // 操作
    fetchTickets: handleFetchTickets,
    fetchTicket: handleFetchTicket,
    createTicket: handleCreateTicket,
    updateTicket: handleUpdateTicket,
    updateStatus: handleUpdateStatus,
    assignTicket: handleAssignTicket,
    fetchProcessRecords: handleFetchProcessRecords,
    addProcessRecord: handleAddProcessRecord,

    setFilters: handleSetFilters,
    clearFilters: handleClearFilters,
    setPage: handleSetPage,
    setLimit: handleSetLimit,
    clearError: handleClearError,
    clearCurrentTicket: handleClearCurrentTicket,
  };
};

export default useTicket;