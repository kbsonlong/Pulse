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
                <Route path="datasources" element={<div>数据源管理</div>} />
              </Route>
              <Route path="tickets">
                <Route index element={<Navigate to="list" replace />} />
                <Route path="list" element={<TicketList />} />
                <Route path=":id" element={<TicketDetail />} />
                <Route path="my" element={<div>我的工单</div>} />
              </Route>
              <Route path="knowledge">
                <Route index element={<Navigate to="list" replace />} />
                <Route path="list" element={<KnowledgeList />} />
                <Route path="categories" element={<div>分类管理</div>} />
              </Route>
              <Route path="system">
                <Route path="users" element={<div>用户管理</div>} />
                <Route path="settings" element={<div>系统设置</div>} />
              </Route>
              <Route path="profile" element={<div>个人资料</div>} />
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
