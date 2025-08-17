import { useCallback } from 'react';
import { useAppDispatch, useAppSelector } from './index';
import {
  fetchKnowledgeList,
  fetchKnowledge,
  createKnowledge,
  updateKnowledge,
  deleteKnowledge,
  searchKnowledge,
  fetchCategories,
  fetchTags,
  setFilters,
  clearFilters,
  setPage,
  setLimit,
  clearError,
  clearCurrentKnowledge,
} from '../store/slices/knowledgeSlice';

const useKnowledge = () => {
  const dispatch = useAppDispatch();
  const {
    knowledgeList,
    currentKnowledge,
    categories,
    tags,
    total,
    page,
    limit,
    loading,
    error,
    filters,
  } = useAppSelector((state) => state.knowledge);

  const handleFetchKnowledgeList = useCallback(
    (params?: any) => {
      return dispatch(fetchKnowledgeList(params));
    },
    [dispatch]
  );

  const handleFetchKnowledge = useCallback(
    (id: string) => {
      return dispatch(fetchKnowledge(id));
    },
    [dispatch]
  );

  const handleCreateKnowledge = useCallback(
    (knowledgeData: any) => {
      return dispatch(createKnowledge(knowledgeData));
    },
    [dispatch]
  );

  const handleUpdateKnowledge = useCallback(
    (id: string, data: any) => {
      return dispatch(updateKnowledge({ id, data }));
    },
    [dispatch]
  );

  const handleDeleteKnowledge = useCallback(
    (id: string) => {
      return dispatch(deleteKnowledge(id));
    },
    [dispatch]
  );

  const handleSearchKnowledge = useCallback(
    (params: any) => {
      return dispatch(searchKnowledge(params));
    },
    [dispatch]
  );

  const handleFetchCategories = useCallback(() => {
    return dispatch(fetchCategories());
  }, [dispatch]);

  const handleFetchTags = useCallback(
    (category?: string) => {
      return dispatch(fetchTags(category));
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

  const handleClearCurrentKnowledge = useCallback(() => {
    dispatch(clearCurrentKnowledge());
  }, [dispatch]);

  return {
    // 状态
    knowledgeList,
    currentKnowledge,
    categories,
    tags,
    total,
    page,
    limit,
    loading,
    error,
    filters,
    
    // 操作
    fetchKnowledgeList: handleFetchKnowledgeList,
    fetchKnowledge: handleFetchKnowledge,
    createKnowledge: handleCreateKnowledge,
    updateKnowledge: handleUpdateKnowledge,
    deleteKnowledge: handleDeleteKnowledge,
    searchKnowledge: handleSearchKnowledge,
    fetchCategories: handleFetchCategories,
    fetchTags: handleFetchTags,
    setFilters: handleSetFilters,
    clearFilters: handleClearFilters,
    setPage: handleSetPage,
    setLimit: handleSetLimit,
    clearError: handleClearError,
    clearCurrentKnowledge: handleClearCurrentKnowledge,
  };
};

export default useKnowledge;