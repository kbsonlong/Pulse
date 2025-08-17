import React from 'react';
import { Layout, Card, theme } from 'antd';
import { Outlet } from 'react-router-dom';

const { Content } = Layout;
const { useToken } = theme;

const AuthLayout: React.FC = () => {
  const { token } = useToken();

  return (
    <Layout
      style={{
        minHeight: '100vh',
        background: `linear-gradient(135deg, ${token.colorPrimary}20 0%, ${token.colorPrimary}40 100%)`,
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
      }}
    >
      <Content
        style={{
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          padding: 24,
        }}
      >
        <Card
          style={{
            width: '100%',
            maxWidth: 400,
            boxShadow: '0 4px 12px rgba(0, 0, 0, 0.1)',
          }}
          bodyStyle={{ padding: 32 }}
        >
          <div
            style={{
              textAlign: 'center',
              marginBottom: 32,
            }}
          >
            <h1
              style={{
                fontSize: 32,
                fontWeight: 'bold',
                color: token.colorPrimary,
                margin: 0,
              }}
            >
              Pulse
            </h1>
            <p
              style={{
                fontSize: 16,
                color: token.colorTextSecondary,
                margin: '8px 0 0 0',
              }}
            >
              告警管理平台
            </p>
          </div>
          <Outlet />
        </Card>
      </Content>
    </Layout>
  );
};

export default AuthLayout;