import React from 'react';
import { Layout, Card, theme } from 'antd';
import { Outlet } from 'react-router-dom';
import { Logo, PulseBeat } from '../components';

const { Content } = Layout;
const { useToken } = theme;

const AuthLayout: React.FC = () => {
  const { token } = useToken();

  return (
    <Layout
      style={{
        minHeight: '100vh',
        background: 'linear-gradient(135deg, #2C3E50 0%, #34495E 50%, #2C3E50 100%)',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        position: 'relative',
        overflow: 'hidden',
      }}
    >
      {/* 背景动画元素 */}
      <div
        style={{
          position: 'absolute',
          top: '20%',
          left: '10%',
          width: '100px',
          height: '100px',
          background: 'radial-gradient(circle, #FFD700 0%, transparent 70%)',
          borderRadius: '50%',
          opacity: 0.1,
          animation: 'float 6s ease-in-out infinite',
        }}
      />
      <div
        style={{
          position: 'absolute',
          bottom: '30%',
          right: '15%',
          width: '80px',
          height: '80px',
          background: 'radial-gradient(circle, #FF4757 0%, transparent 70%)',
          borderRadius: '50%',
          opacity: 0.1,
          animation: 'float 8s ease-in-out infinite reverse',
        }}
      />
      <style>
        {`
          @keyframes float {
            0%, 100% { transform: translateY(0px) rotate(0deg); }
            50% { transform: translateY(-20px) rotate(180deg); }
          }
        `}
      </style>
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
          styles={{ body: { padding: 32 } }}
        >
          <div
            style={{
              textAlign: 'center',
              marginBottom: 32,
            }}
          >
            <div style={{ marginBottom: 16 }}>
              <Logo 
                variant="full" 
                size="lg" 
                animated={true}
                className="mx-auto"
              />
            </div>
            <div 
              style={{
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                marginTop: 12,
              }}
            >
              <PulseBeat className="mr-2" />
              <span
                style={{
                  fontSize: 14,
                  color: '#95A5A6',
                  fontWeight: 500,
                }}
              >
                感知每一次脉搏，响应每一个告警
              </span>
            </div>
          </div>
          <Outlet />
        </Card>
      </Content>
    </Layout>
  );
};

export default AuthLayout;