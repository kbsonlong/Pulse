import React, { useEffect } from 'react';
import { Form, Input, Button, Checkbox, message } from 'antd';
import { UserOutlined, LockOutlined } from '@ant-design/icons';
import { Link, useNavigate } from 'react-router-dom';
import { useAuth } from '../hooks';

interface LoginForm {
  username: string;
  password: string;
  remember: boolean;
}

const Login: React.FC = () => {
  const [form] = Form.useForm();
  const navigate = useNavigate();
  const { login, loading, error, isAuthenticated } = useAuth();

  useEffect(() => {
    if (isAuthenticated) {
      navigate('/dashboard');
    }
  }, [isAuthenticated, navigate]);

  useEffect(() => {
    if (error) {
      message.error(error);
    }
  }, [error]);

  const handleSubmit = async (values: LoginForm) => {
    try {
      await login({
        username: values.username,
        password: values.password,
      });
      message.success('登录成功');
      navigate('/dashboard');
    } catch (err) {
      // 错误已在useAuth中处理
    }
  };

  return (
    <Form
      form={form}
      name="login"
      onFinish={handleSubmit}
      autoComplete="off"
      size="large"
    >
      <Form.Item
        name="username"
        rules={[
          { required: true, message: '请输入用户名' },
          { min: 3, message: '用户名至少3个字符' },
        ]}
      >
        <Input
          prefix={<UserOutlined />}
          placeholder="用户名"
        />
      </Form.Item>

      <Form.Item
        name="password"
        rules={[
          { required: true, message: '请输入密码' },
          { min: 6, message: '密码至少6个字符' },
        ]}
      >
        <Input.Password
          prefix={<LockOutlined />}
          placeholder="密码"
        />
      </Form.Item>

      <Form.Item name="remember" valuePropName="checked">
        <Checkbox>记住我</Checkbox>
      </Form.Item>

      <Form.Item>
        <Button
          type="primary"
          htmlType="submit"
          loading={loading}
          style={{ width: '100%' }}
        >
          登录
        </Button>
      </Form.Item>

      <Form.Item style={{ textAlign: 'center', marginBottom: 0 }}>
        <span>还没有账号？</span>
        <Link to="/register" style={{ marginLeft: 8 }}>
          立即注册
        </Link>
      </Form.Item>
    </Form>
  );
};

export default Login;