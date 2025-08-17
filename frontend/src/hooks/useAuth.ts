import { useCallback } from 'react';
import { useAppDispatch, useAppSelector } from './index';
import {
  login,
  register,
  logout,
  refreshToken,
  changePassword,
  clearError,
  updateUser,
} from '../store/slices/authSlice';
import { LoginRequest, User } from '../types';

const useAuth = () => {
  const dispatch = useAppDispatch();
  const {
    user,
    token,
    isAuthenticated,
    loading,
    error,
  } = useAppSelector((state) => state.auth);

  const handleLogin = useCallback(
    (credentials: LoginRequest) => {
      return dispatch(login(credentials));
    },
    [dispatch]
  );

  const handleRegister = useCallback(
    (userData: {
      username: string;
      email: string;
      password: string;
      role?: string;
    }) => {
      return dispatch(register(userData));
    },
    [dispatch]
  );

  const handleLogout = useCallback(() => {
    dispatch(logout());
  }, [dispatch]);

  const handleRefreshToken = useCallback(() => {
    return dispatch(refreshToken());
  }, [dispatch]);

  const handleChangePassword = useCallback(
    (data: {
      old_password: string;
      new_password: string;
    }) => {
      return dispatch(changePassword(data));
    },
    [dispatch]
  );

  const handleUpdateUser = useCallback(
    (userData: Partial<User>) => {
      dispatch(updateUser(userData));
    },
    [dispatch]
  );

  const handleClearError = useCallback(() => {
    dispatch(clearError());
  }, [dispatch]);

  // 检查用户权限
  const hasPermission = useCallback(
    (permission: string) => {
      if (!user) return false;
      // 这里可以根据实际的权限系统来实现
      // 例如检查用户角色或权限列表
      return user.role === 'admin' || user.role === permission;
    },
    [user]
  );

  // 检查用户角色
  const hasRole = useCallback(
    (role: string) => {
      return user?.role === role;
    },
    [user]
  );

  return {
    // 状态
    user,
    token,
    isAuthenticated,
    loading,
    error,
    
    // 操作
    login: handleLogin,
    register: handleRegister,
    logout: handleLogout,
    refreshToken: handleRefreshToken,
    changePassword: handleChangePassword,
    updateUser: handleUpdateUser,
    clearError: handleClearError,
    
    // 权限检查
    hasPermission,
    hasRole,
  };
};

export default useAuth;