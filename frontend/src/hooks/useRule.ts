import { useCallback } from 'react';
import { useAppDispatch, useAppSelector } from './index';
import {
  fetchRules,
  fetchRule,
  createRule,
  updateRule,
  deleteRule,
  toggleRule,
  fetchDataSources,
  setFilters,
  clearFilters,
  setPage,
  setLimit,
  clearError,
  clearCurrentRule,
} from '../store/slices/ruleSlice';
import { Rule } from '../types';

const useRule = () => {
  const dispatch = useAppDispatch();
  const {
    rules,
    currentRule,
    dataSources,
    total,
    page,
    limit,
    loading,
    error,
    filters,
  } = useAppSelector((state) => state.rule);

  const handleFetchRules = useCallback(
    (params?: {
      page?: number;
      limit?: number;
      search?: string;
      enabled?: boolean;
      data_source_id?: string;
    }) => {
      return dispatch(fetchRules(params));
    },
    [dispatch]
  );

  const handleFetchRule = useCallback(
    (id: string) => {
      return dispatch(fetchRule(id));
    },
    [dispatch]
  );

  const handleCreateRule = useCallback(
    (ruleData: any) => {
      return dispatch(createRule(ruleData));
    },
    [dispatch]
  );

  const handleUpdateRule = useCallback(
    (id: string, data: Partial<Rule>) => {
      return dispatch(updateRule({ id, data }));
    },
    [dispatch]
  );

  const handleDeleteRule = useCallback(
    (id: string) => {
      return dispatch(deleteRule(id));
    },
    [dispatch]
  );

  const handleToggleRule = useCallback(
    (id: string, enabled: boolean) => {
      return dispatch(toggleRule({ id, enabled }));
    },
    [dispatch]
  );

  const handleFetchDataSources = useCallback(() => {
    return dispatch(fetchDataSources());
  }, [dispatch]);

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

  const handleClearCurrentRule = useCallback(() => {
    dispatch(clearCurrentRule());
  }, [dispatch]);

  return {
    // 状态
    rules,
    currentRule,
    dataSources,
    total,
    page,
    limit,
    loading,
    error,
    filters,
    
    // 操作
    fetchRules: handleFetchRules,
    fetchRule: handleFetchRule,
    createRule: handleCreateRule,
    updateRule: handleUpdateRule,
    deleteRule: handleDeleteRule,
    toggleRule: handleToggleRule,
    fetchDataSources: handleFetchDataSources,
    setFilters: handleSetFilters,
    clearFilters: handleClearFilters,
    setPage: handleSetPage,
    setLimit: handleSetLimit,
    clearError: handleClearError,
    clearCurrentRule: handleClearCurrentRule,
  };
};

export default useRule;