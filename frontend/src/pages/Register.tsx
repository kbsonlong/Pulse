import React, { useEffect } from 'react';
import { Form, Input, Button, message } from 'antd';
import { UserOutlined, LockOutlined, MailOutlined } from '@ant-design/icons';
import { Link, useNavigate } from 'react-router-dom';
import { useAuth } from '../hooks';
import { isValidEmail } from '../utils';

interface RegisterForm {
  username: string;
  email: string;
  password: string;
  confirmPassword: string;
}

const Register: React.FC = () => {
  const [form] = Form.useForm();
  const navigate = useNavigate();
  const { register, loading, error, isAuthenticated } = useAuth();

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

  const handleSubmit = async (values: RegisterForm) => {
    try {
      await register({
        username: values.username,
        email: values.email,
        password: values.password,
      });
      message.success('注册成功，请登录');
      navigate('/login');
    } catch (err) {
      // 错误已在useAuth中处理
    }
  };

  return (
    <Form
      form={form}
      name="register"
      onFinish={handleSubmit}
      autoComplete="off"
      size="large"
    >
      <Form.Item
        name="username"
        rules={[
          { required: true, message: '请输入用户名' },
          { min: 3, message: '用户名至少3个字符' },
          { max: 20, message: '用户名最多20个字符' },
          {
            pattern: /^[a-zA-Z0-9_]+$/,
            message: '用户名只能包含字母、数字和下划线',
          },
        ]}
      >
        <Input
          prefix={<UserOutlined />}
          placeholder="用户名"
        />
      </Form.Item>

      <Form.Item
        name="email"
        rules={[
          { required: true, message: '请输入邮箱' },
          {
            validator: (_, value) => {
              if (!value || isValidEmail(value)) {
                return Promise.resolve();
              }
              return Promise.reject(new Error('请输入有效的邮箱地址'));
            },
          },
        ]}
      >
        <Input
          prefix={<MailOutlined />}
          placeholder="邮箱"
        />
      </Form.Item>

      <Form.Item
        name="password"
        rules={[
          { required: true, message: '请输入密码' },
          { min: 6, message: '密码至少6个字符' },
          { max: 50, message: '密码最多50个字符' },
        ]}
      >
        <Input.Password
          prefix={<LockOutlined />}
          placeholder="密码"
        />
      </Form.Item>

      <Form.Item
        name="confirmPassword"
        dependencies={['password']}
        rules={[
          { required: true, message: '请确认密码' },
          ({ getFieldValue }) => ({
            validator(_, value) {
              if (!value || getFieldValue('password') === value) {
                return Promise.resolve();
              }
              return Promise.reject(new Error('两次输入的密码不一致'));
            },
          }),
        ]}
      >
        <Input.Password
          prefix={<LockOutlined />}
          placeholder="确认密码"
        />
      </Form.Item>

      <Form.Item>
        <Button
          type="primary"
          htmlType="submit"
          loading={loading}
          style={{ width: '100%' }}
        >
          注册
        </Button>
      </Form.Item>

      <Form.Item style={{ textAlign: 'center', marginBottom: 0 }}>
        <span>已有账号？</span>
        <Link to="/login" style={{ marginLeft: 8 }}>
          立即登录
        </Link>
      </Form.Item>
    </Form>
  );
};

export default Register;