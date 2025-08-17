import { configureStore } from '@reduxjs/toolkit';
import authSlice from './slices/authSlice';
import alertSlice from './slices/alertSlice';
import ruleSlice from './slices/ruleSlice';
import ticketSlice from './ticketSlice';
import knowledgeSlice from './slices/knowledgeSlice';
import uiSlice from './slices/uiSlice';

export const store = configureStore({
  reducer: {
    auth: authSlice,
    alert: alertSlice,
    rule: ruleSlice,
    ticket: ticketSlice,
    knowledge: knowledgeSlice,
    ui: uiSlice,
  },
  middleware: (getDefaultMiddleware) =>
    getDefaultMiddleware({
      serializableCheck: {
        ignoredActions: ['persist/PERSIST'],
      },
    }),
});

export type RootState = ReturnType<typeof store.getState>;
export type AppDispatch = typeof store.dispatch;