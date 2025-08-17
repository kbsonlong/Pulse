import { useCallback } from 'react';
import { useAppDispatch, useAppSelector } from './index';
import {
  fetchAlerts,
  fetchAlert,
  updateAlertStatus,
  batchUpdateAlertStatus,
  deleteAlert,
  fetchStatistics,
  searchAlerts,
  setFilters,
  clearFilters,
  setPage,
  setLimit,
  clearError,
  clearCurrentAlert,
} from '../store/slices/alertSlice';
import { AlertQuery, AlertStatus, AlertLevel } from '../types';

const useAlert = () => {
  const dispatch = useAppDispatch();
  const {
    alerts,
    currentAlert,
    statistics,
    total,
    page,
    limit,
    loading,
    error,
    filters,
  } = useAppSelector((state) => state.alert);

  const handleFetchAlerts = useCallback(
    (query?: AlertQuery) => {
      return dispatch(fetchAlerts(query));
    },
    [dispatch]
  );

  const handleFetchAlert = useCallback(
    (id: string) => {
      return dispatch(fetchAlert(id));
    },
    [dispatch]
  );

  const handleUpdateStatus = useCallback(
    (id: string, status: AlertStatus) => {
      return dispatch(updateAlertStatus({ id, status }));
    },
    [dispatch]
  );

  const handleBatchUpdateStatus = useCallback(
    (ids: string[], status: AlertStatus) => {
      return dispatch(batchUpdateAlertStatus({ ids, status }));
    },
    [dispatch]
  );

  const handleDeleteAlert = useCallback(
    (id: string) => {
      return dispatch(deleteAlert(id));
    },
    [dispatch]
  );

  const handleFetchStatistics = useCallback(
    (params?: {
      start_time?: string;
      end_time?: string;
      source?: string;
    }) => {
      return dispatch(fetchStatistics(params));
    },
    [dispatch]
  );

  const handleSearchAlerts = useCallback(
    (params: {
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
    }) => {
      return dispatch(searchAlerts(params));
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

  const handleClearCurrentAlert = useCallback(() => {
    dispatch(clearCurrentAlert());
  }, [dispatch]);

  return {
    // 状态
    alerts,
    currentAlert,
    statistics,
    total,
    page,
    limit,
    loading,
    error,
    filters,
    
    // 操作
    fetchAlerts: handleFetchAlerts,
    fetchAlert: handleFetchAlert,
    updateStatus: handleUpdateStatus,
    batchUpdateStatus: handleBatchUpdateStatus,
    deleteAlert: handleDeleteAlert,
    fetchStatistics: handleFetchStatistics,
    searchAlerts: handleSearchAlerts,
    setFilters: handleSetFilters,
    clearFilters: handleClearFilters,
    setPage: handleSetPage,
    setLimit: handleSetLimit,
    clearError: handleClearError,
    clearCurrentAlert: handleClearCurrentAlert,
  };
};

export default useAlert;