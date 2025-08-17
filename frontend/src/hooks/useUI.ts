import { useCallback } from 'react';
import { useAppDispatch, useAppSelector } from './index';
import {
  toggleSidebar,
  setTheme,
  setLoading,
  addNotification,
  removeNotification,
  clearNotifications,
  setBreadcrumbs,
} from '../store/slices/uiSlice';
import { Theme, NotificationType } from '../types';

const useUI = () => {
  const dispatch = useAppDispatch();
  const {
    sidebarCollapsed,
    theme,
    loading,
    notifications,
    breadcrumbs,
  } = useAppSelector((state) => state.ui);

  const handleToggleSidebar = useCallback(() => {
    dispatch(toggleSidebar());
  }, [dispatch]);

  const handleSetTheme = useCallback(
    (newTheme: Theme) => {
      dispatch(setTheme(newTheme));
    },
    [dispatch]
  );

  const handleSetLoading = useCallback(
    (isLoading: boolean) => {
      dispatch(setLoading(isLoading));
    },
    [dispatch]
  );

  const handleAddNotification = useCallback(
    (notification: {
      type: NotificationType;
      title: string;
      message?: string;
      duration?: number;
    }) => {
      dispatch(addNotification(notification));
    },
    [dispatch]
  );

  const handleRemoveNotification = useCallback(
    (id: string) => {
      dispatch(removeNotification(id));
    },
    [dispatch]
  );

  const handleClearNotifications = useCallback(() => {
    dispatch(clearNotifications());
  }, [dispatch]);

  const handleSetBreadcrumbs = useCallback(
    (items: Array<{ title: string; path?: string }>) => {
      dispatch(setBreadcrumbs(items));
    },
    [dispatch]
  );

  const handleClearBreadcrumbs = useCallback(() => {
    dispatch(setBreadcrumbs([]));
  }, [dispatch]);

  // 便捷的通知方法
  const showSuccess = useCallback(
    (title: string, message?: string, duration?: number) => {
      handleAddNotification({
        type: 'success',
        title,
        message,
        duration,
      });
    },
    [handleAddNotification]
  );

  const showError = useCallback(
    (title: string, message?: string, duration?: number) => {
      handleAddNotification({
        type: 'error',
        title,
        message,
        duration,
      });
    },
    [handleAddNotification]
  );

  const showWarning = useCallback(
    (title: string, message?: string, duration?: number) => {
      handleAddNotification({
        type: 'warning',
        title,
        message,
        duration,
      });
    },
    [handleAddNotification]
  );

  const showInfo = useCallback(
    (title: string, message?: string, duration?: number) => {
      handleAddNotification({
        type: 'info',
        title,
        message,
        duration,
      });
    },
    [handleAddNotification]
  );

  return {
    // 状态
    sidebarCollapsed,
    theme,
    loading,
    notifications,
    breadcrumbs,
    
    // 操作
    toggleSidebar: handleToggleSidebar,
    setTheme: handleSetTheme,
    setLoading: handleSetLoading,
    addNotification: handleAddNotification,
    removeNotification: handleRemoveNotification,
    clearNotifications: handleClearNotifications,
    setBreadcrumbs: handleSetBreadcrumbs,
    clearBreadcrumbs: handleClearBreadcrumbs,
    
    // 便捷方法
    showSuccess,
    showError,
    showWarning,
    showInfo,
  };
};

export default useUI;
export { useUI };