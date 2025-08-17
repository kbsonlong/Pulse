import { TypedUseSelectorHook, useDispatch, useSelector } from 'react-redux';
import type { RootState, AppDispatch } from '../store';

// 使用类型化的hooks
export const useAppDispatch = () => useDispatch<AppDispatch>();
export const useAppSelector: TypedUseSelectorHook<RootState> = useSelector;

// 导出其他自定义hooks
export { default as useAuth } from './useAuth';
export { default as useAlert } from './useAlert';
export { default as useRule } from './useRule';
export { default as useTicket } from './useTicket';
export { default as useKnowledge } from './useKnowledge';
export { default as useUI } from './useUI';