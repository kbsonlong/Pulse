import React, { useEffect } from 'react';
import { Layout, Menu, Avatar, Dropdown, Button, Breadcrumb, theme } from 'antd';
import {
  MenuFoldOutlined,
  MenuUnfoldOutlined,
  DashboardOutlined,
  AlertOutlined,
  SettingOutlined,
  FileTextOutlined,
  CustomerServiceOutlined,
  BookOutlined,
  UserOutlined,
  LogoutOutlined,
} from '@ant-design/icons';
import { Outlet, useNavigate, useLocation } from 'react-router-dom';
import { useAuth, useUI } from '../hooks';
import { LogoIcon } from '../components';

const { Header, Sider, Content } = Layout;
const { useToken } = theme;

const MainLayout: React.FC = () => {
  const navigate = useNavigate();
  const location = useLocation();
  const { user, logout } = useAuth();
  const { sidebarCollapsed, toggleSidebar, breadcrumb } = useUI();
  const { token } = useToken();

  // 菜单项配置
  const menuItems = [
    {
      key: '/dashboard',
      icon: <DashboardOutlined />,
      label: '仪表盘',
    },
    {
      key: '/alerts',
      icon: <AlertOutlined />,
      label: '告警管理',
      children: [
        {
          key: '/alerts/list',
          label: '告警列表',
        },
        {
          key: '/alerts/statistics',
          label: '告警统计',
        },
      ],
    },
    {
      key: '/rules',
      icon: <SettingOutlined />,
      label: '规则管理',
      children: [
        {
          key: '/rules/list',
          label: '规则列表',
        },
        {
          key: '/rules/datasources',
          label: '数据源管理',
        },
      ],
    },
    {
      key: '/tickets',
      icon: <CustomerServiceOutlined />,
      label: '工单管理',
      children: [
        {
          key: '/tickets/list',
          label: '工单列表',
        },
        {
          key: '/tickets/my',
          label: '我的工单',
        },
      ],
    },
    {
      key: '/knowledge',
      icon: <BookOutlined />,
      label: '知识库',
      children: [
        {
          key: '/knowledge/list',
          label: '知识库列表',
        },
        {
          key: '/knowledge/categories',
          label: '分类管理',
        },
      ],
    },
    {
      key: '/system',
      icon: <SettingOutlined />,
      label: '系统管理',
      children: [
        {
          key: '/system/users',
          label: '用户管理',
        },
        {
          key: '/system/settings',
          label: '系统设置',
        },
      ],
    },
  ];

  // 用户下拉菜单
  const userMenuItems = [
    {
      key: 'profile',
      icon: <UserOutlined />,
      label: '个人资料',
      onClick: () => navigate('/profile'),
    },
    {
      type: 'divider' as const,
    },
    {
      key: 'logout',
      icon: <LogoutOutlined />,
      label: '退出登录',
      onClick: () => {
        logout();
        navigate('/login');
      },
    },
  ];

  // 获取当前选中的菜单项
  const getSelectedKeys = () => {
    const path = location.pathname;
    // 查找匹配的菜单项
    for (const item of menuItems) {
      if (item.children) {
        for (const child of item.children) {
          if (path.startsWith(child.key)) {
            return [child.key];
          }
        }
      } else if (path.startsWith(item.key)) {
        return [item.key];
      }
    }
    return [path];
  };

  // 获取展开的菜单项
  const getOpenKeys = () => {
    const path = location.pathname;
    const openKeys: string[] = [];
    
    for (const item of menuItems) {
      if (item.children) {
        for (const child of item.children) {
          if (path.startsWith(child.key)) {
            openKeys.push(item.key);
            break;
          }
        }
      }
    }
    return openKeys;
  };

  const handleMenuClick = ({ key }: { key: string }) => {
    navigate(key);
  };

  return (
    <Layout style={{ minHeight: '100vh' }}>
      <Sider
        trigger={null}
        collapsible
        collapsed={sidebarCollapsed}
        style={{
          background: token.colorBgContainer,
          borderRight: `1px solid ${token.colorBorder}`,
        }}
      >
        <div
          style={{
            height: 64,
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            borderBottom: `1px solid ${token.colorBorder}`,
            padding: '0 16px',
          }}
        >
          <LogoIcon 
            size={sidebarCollapsed ? 'sm' : 'md'} 
            animated={true}
            className="cursor-pointer hover:scale-105 transition-transform"
          />
          {!sidebarCollapsed && (
            <span 
              style={{
                marginLeft: 8,
                fontSize: 18,
                fontWeight: 'bold',
                color: '#2C3E50',
                background: 'linear-gradient(45deg, #FFD700 0%, #FF4757 100%)',
                WebkitBackgroundClip: 'text',
                WebkitTextFillColor: 'transparent',
                backgroundClip: 'text',
              }}
            >
              Pulse
            </span>
          )}
        </div>
        <Menu
          mode="inline"
          selectedKeys={getSelectedKeys()}
          defaultOpenKeys={getOpenKeys()}
          items={menuItems}
          onClick={handleMenuClick}
          style={{ borderRight: 0 }}
        />
      </Sider>
      <Layout>
        <Header
          style={{
            padding: '0 16px',
            background: token.colorBgContainer,
            borderBottom: `1px solid ${token.colorBorder}`,
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between',
          }}
        >
          <div style={{ display: 'flex', alignItems: 'center' }}>
            <Button
              type="text"
              icon={sidebarCollapsed ? <MenuUnfoldOutlined /> : <MenuFoldOutlined />}
              onClick={toggleSidebar}
              style={{ fontSize: 16, width: 64, height: 64 }}
            />
            {breadcrumb && breadcrumb.length > 0 && (
              <Breadcrumb
                style={{ marginLeft: 16 }}
                items={breadcrumb.map((item) => ({
                  title: item.path ? (
                    <a onClick={() => navigate(item.path!)}>{item.title}</a>
                  ) : (
                    item.title
                  ),
                }))}
              />
            )}
          </div>
          <div style={{ display: 'flex', alignItems: 'center' }}>
            <Dropdown menu={{ items: userMenuItems }} placement="bottomRight">
              <div
                style={{
                  display: 'flex',
                  alignItems: 'center',
                  cursor: 'pointer',
                  padding: '0 8px',
                }}
              >
                <Avatar
                  size="small"
                  icon={<UserOutlined />}
                  style={{ marginRight: 8 }}
                />
                <span>{user?.username || '用户'}</span>
              </div>
            </Dropdown>
          </div>
        </Header>
        <Content
          style={{
            margin: 16,
            padding: 24,
            background: token.colorBgContainer,
            borderRadius: token.borderRadius,
            minHeight: 280,
          }}
        >
          <Outlet />
        </Content>
      </Layout>
    </Layout>
  );
};

export default MainLayout;