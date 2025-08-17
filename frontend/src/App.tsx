import React from 'react';
import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import { Provider } from 'react-redux';
import { ConfigProvider } from 'antd';
import zhCN from 'antd/locale/zh_CN';
import { store } from './store';
import { MainLayout, AuthLayout } from './layouts';
import { 
  Login, 
  Register, 
  Dashboard, 
  AlertList, 
  AlertDetail,
  RuleList,
  TicketList,
  TicketDetail,
  KnowledgeList
} from './pages';
import DataSourceList from './pages/rules/DataSourceList';
import RuleForm from './pages/rules/RuleForm';
import DataSourceManagement from './pages/rules/DataSourceManagement';
import TicketForm from './pages/tickets/TicketForm';
import MyTickets from './pages/tickets/MyTickets';
import CategoryManagement from './pages/knowledge/CategoryManagement';
import KnowledgeForm from './pages/knowledge/KnowledgeForm';
import KnowledgeDetail from './pages/knowledge/KnowledgeDetail';
import UserManagement from './pages/system/UserManagement';
import SystemSettings from './pages/system/SystemSettings';
import Profile from './pages/system/Profile';
import 'antd/dist/reset.css';

const App: React.FC = () => {
  return (
    <Provider store={store}>
      <ConfigProvider locale={zhCN}>
        <Router>
          <Routes>
            {/* 认证路由 */}
            <Route path="/" element={<AuthLayout />}>
              <Route index element={<Navigate to="/login" replace />} />
              <Route path="login" element={<Login />} />
              <Route path="register" element={<Register />} />
            </Route>
            
            {/* 主应用路由 */}
            <Route path="/" element={<MainLayout />}>
              <Route path="dashboard" element={<Dashboard />} />
              <Route path="alerts">
                <Route index element={<Navigate to="list" replace />} />
                <Route path="list" element={<AlertList />} />
                <Route path=":id" element={<AlertDetail />} />
                <Route path="statistics" element={<div>告警统计</div>} />
              </Route>
              <Route path="rules">
                <Route index element={<Navigate to="list" replace />} />
                <Route path="list" element={<RuleList />} />
                <Route path="create" element={<RuleForm />} />
                <Route path="edit/:id" element={<RuleForm />} />
                <Route path="datasources" element={<DataSourceManagement />} />
              </Route>
              <Route path="tickets">
                <Route index element={<Navigate to="list" replace />} />
                <Route path="list" element={<TicketList />} />
                <Route path="create" element={<TicketForm />} />
                <Route path="edit/:id" element={<TicketForm />} />
                <Route path="detail/:id" element={<TicketDetail />} />
                <Route path="my" element={<MyTickets />} />
              </Route>
              <Route path="knowledge">
                <Route index element={<Navigate to="list" replace />} />
                <Route path="list" element={<KnowledgeList />} />
                <Route path="create" element={<KnowledgeForm />} />
                <Route path="edit/:id" element={<KnowledgeForm />} />
                <Route path="detail/:id" element={<KnowledgeDetail />} />
                <Route path="categories" element={<CategoryManagement />} />
              </Route>
              <Route path="system">
                <Route path="users" element={<UserManagement />} />
                <Route path="settings" element={<SystemSettings />} />
              </Route>
              <Route path="profile" element={<Profile />} />
            </Route>
            
            {/* 404 页面 */}
            <Route path="*" element={<Navigate to="/dashboard" replace />} />
          </Routes>
        </Router>
      </ConfigProvider>
    </Provider>
  );
};

export default App;
