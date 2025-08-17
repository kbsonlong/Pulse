import React, { useState, useEffect } from 'react';
import {
  Card,
  Form,
  Input,
  Button,
  Avatar,
  Upload,
  message,
  Tabs,
  Switch,
  Select,
  Divider,
  Row,
  Col,
  Modal,
  List,
  Tag,
  Space,
  Alert
} from 'antd';
import {
  UserOutlined,
  CameraOutlined,
  SaveOutlined,
  LockOutlined,
  BellOutlined,
  SecurityScanOutlined,
  HistoryOutlined,
  LogoutOutlined,
  ExclamationCircleOutlined
} from '@ant-design/icons';
import { useAuth, useUI } from '../../hooks';
import { userService } from '../../services';

const { TabPane } = Tabs;
const { Option } = Select;
const { TextArea } = Input;

interface UserProfile {
  id: string;
  username: string;
  email: string;
  fullName: string;
  phone: string;
  avatar: string;
  bio: string;
  department: string;
  position: string;
  timezone: string;
  language: string;
}

interface NotificationSettings {
  emailNotifications: boolean;
  smsNotifications: boolean;
  pushNotifications: boolean;
  alertNotifications: boolean;
  ticketNotifications: boolean;
  systemNotifications: boolean;
  weeklyReport: boolean;
  monthlyReport: boolean;
}

interface SecuritySettings {
  twoFactorEnabled: boolean;
  loginAlerts: boolean;
  sessionTimeout: number;
}

interface LoginHistory {
  id: string;
  loginTime: string;
  ipAddress: string;
  userAgent: string;
  location: string;
  success: boolean;
}

const ProfilePage: React.FC = () => {
  const { user, logout } = useAuth();
  const { loading, setLoading } = useUI();
  const [profile, setProfile] = useState<UserProfile | null>(null);
  const [notifications, setNotifications] = useState<NotificationSettings | null>(null);
  const [security, setSecurity] = useState<SecuritySettings | null>(null);
  const [loginHistory, setLoginHistory] = useState<LoginHistory[]>([]);
  const [activeTab, setActiveTab] = useState('profile');
  const [avatarUrl, setAvatarUrl] = useState<string>('');
  const [form] = Form.useForm();
  const [passwordForm] = Form.useForm();
  const [notificationForm] = Form.useForm();

  useEffect(() => {
    loadProfile();
    loadNotificationSettings();
    loadSecuritySettings();
    loadLoginHistory();
  }, []);

  const loadProfile = async () => {
    try {
      setLoading(true);
      const response = await userService.getCurrentUser();
      setProfile(response.data);
      setAvatarUrl(response.data.avatar);
      form.setFieldsValue(response.data);
    } catch (error) {
      message.error('加载用户信息失败');
    } finally {
      setLoading(false);
    }
  };

  const loadNotificationSettings = async () => {
    try {
      const response = await userService.getNotificationSettings();
      setNotifications(response.data);
      notificationForm.setFieldsValue(response.data);
    } catch (error) {
      message.error('加载通知设置失败');
    }
  };

  const loadSecuritySettings = async () => {
    try {
      const response = await userService.getSecuritySettings();
      setSecurity(response.data);
    } catch (error) {
      message.error('加载安全设置失败');
    }
  };

  const loadLoginHistory = async () => {
    try {
      const response = await userService.getLoginHistory();
      setLoginHistory(response.data);
    } catch (error) {
      message.error('加载登录历史失败');
    }
  };

  const handleProfileUpdate = async (values: UserProfile) => {
    try {
      await userService.updateProfile(values);
      message.success('个人信息更新成功');
      setProfile({ ...profile, ...values });
    } catch (error) {
      message.error('个人信息更新失败');
    }
  };

  const handlePasswordChange = async (values: any) => {
    try {
      await userService.changePassword({
        currentPassword: values.currentPassword,
        newPassword: values.newPassword
      });
      message.success('密码修改成功');
      passwordForm.resetFields();
    } catch (error) {
      message.error('密码修改失败');
    }
  };

  const handleNotificationUpdate = async (values: NotificationSettings) => {
    try {
      await userService.updateNotificationSettings(values);
      message.success('通知设置更新成功');
      setNotifications(values);
    } catch (error) {
      message.error('通知设置更新失败');
    }
  };

  const handleAvatarUpload = async (file: File) => {
    try {
      const formData = new FormData();
      formData.append('avatar', file);
      const response = await userService.uploadAvatar(formData);
      setAvatarUrl(response.data.url);
      message.success('头像上传成功');
    } catch (error) {
      message.error('头像上传失败');
    }
    return false;
  };

  const handleLogout = () => {
    Modal.confirm({
      title: '确认退出',
      icon: <ExclamationCircleOutlined />,
      content: '确定要退出登录吗？',
      okText: '确定',
      cancelText: '取消',
      onOk: () => {
        logout();
        message.success('已退出登录');
      }
    });
  };

  const handleToggle2FA = async () => {
    try {
      if (security?.twoFactorEnabled) {
        await userService.disable2FA();
        message.success('双因子认证已禁用');
      } else {
        await userService.enable2FA();
        message.success('双因子认证已启用');
      }
      loadSecuritySettings();
    } catch (error) {
      message.error('双因子认证设置失败');
    }
  };

  return (
    <div style={{ padding: '24px' }}>
      <Card>
        <Tabs activeKey={activeTab} onChange={setActiveTab}>
          <TabPane tab="个人信息" key="profile">
            <Row gutter={24}>
              <Col span={6}>
                <div style={{ textAlign: 'center', marginBottom: '24px' }}>
                  <Avatar
                    size={120}
                    src={avatarUrl}
                    icon={<UserOutlined />}
                    style={{ marginBottom: '16px' }}
                  />
                  <br />
                  <Upload
                    accept="image/*"
                    showUploadList={false}
                    beforeUpload={handleAvatarUpload}
                  >
                    <Button icon={<CameraOutlined />}>
                      更换头像
                    </Button>
                  </Upload>
                </div>
              </Col>
              <Col span={18}>
                <Form
                  form={form}
                  layout="vertical"
                  onFinish={handleProfileUpdate}
                  initialValues={profile}
                >
                  <Row gutter={16}>
                    <Col span={12}>
                      <Form.Item
                        name="fullName"
                        label="姓名"
                        rules={[{ required: true, message: '请输入姓名' }]}
                      >
                        <Input placeholder="请输入姓名" />
                      </Form.Item>
                    </Col>
                    <Col span={12}>
                      <Form.Item
                        name="username"
                        label="用户名"
                        rules={[{ required: true, message: '请输入用户名' }]}
                      >
                        <Input placeholder="请输入用户名" disabled />
                      </Form.Item>
                    </Col>
                  </Row>

                  <Row gutter={16}>
                    <Col span={12}>
                      <Form.Item
                        name="email"
                        label="邮箱"
                        rules={[
                          { required: true, message: '请输入邮箱' },
                          { type: 'email', message: '请输入有效的邮箱地址' }
                        ]}
                      >
                        <Input placeholder="请输入邮箱" />
                      </Form.Item>
                    </Col>
                    <Col span={12}>
                      <Form.Item
                        name="phone"
                        label="手机号"
                        rules={[{ required: true, message: '请输入手机号' }]}
                      >
                        <Input placeholder="请输入手机号" />
                      </Form.Item>
                    </Col>
                  </Row>

                  <Row gutter={16}>
                    <Col span={12}>
                      <Form.Item
                        name="department"
                        label="部门"
                      >
                        <Input placeholder="请输入部门" />
                      </Form.Item>
                    </Col>
                    <Col span={12}>
                      <Form.Item
                        name="position"
                        label="职位"
                      >
                        <Input placeholder="请输入职位" />
                      </Form.Item>
                    </Col>
                  </Row>

                  <Row gutter={16}>
                    <Col span={12}>
                      <Form.Item
                        name="timezone"
                        label="时区"
                      >
                        <Select placeholder="请选择时区">
                          <Option value="Asia/Shanghai">Asia/Shanghai</Option>
                          <Option value="UTC">UTC</Option>
                          <Option value="America/New_York">America/New_York</Option>
                        </Select>
                      </Form.Item>
                    </Col>
                    <Col span={12}>
                      <Form.Item
                        name="language"
                        label="语言"
                      >
                        <Select placeholder="请选择语言">
                          <Option value="zh-CN">简体中文</Option>
                          <Option value="en-US">English</Option>
                        </Select>
                      </Form.Item>
                    </Col>
                  </Row>

                  <Form.Item
                    name="bio"
                    label="个人简介"
                  >
                    <TextArea rows={4} placeholder="请输入个人简介" />
                  </Form.Item>

                  <Form.Item>
                    <Button type="primary" icon={<SaveOutlined />} htmlType="submit">
                      保存更改
                    </Button>
                  </Form.Item>
                </Form>
              </Col>
            </Row>
          </TabPane>

          <TabPane tab="密码安全" key="security">
            <Row gutter={24}>
              <Col span={12}>
                <Card title="修改密码" style={{ marginBottom: '16px' }}>
                  <Form
                    form={passwordForm}
                    layout="vertical"
                    onFinish={handlePasswordChange}
                  >
                    <Form.Item
                      name="currentPassword"
                      label="当前密码"
                      rules={[{ required: true, message: '请输入当前密码' }]}
                    >
                      <Input.Password placeholder="请输入当前密码" />
                    </Form.Item>

                    <Form.Item
                      name="newPassword"
                      label="新密码"
                      rules={[
                        { required: true, message: '请输入新密码' },
                        { min: 6, message: '密码长度至少6位' }
                      ]}
                    >
                      <Input.Password placeholder="请输入新密码" />
                    </Form.Item>

                    <Form.Item
                      name="confirmPassword"
                      label="确认新密码"
                      dependencies={['newPassword']}
                      rules={[
                        { required: true, message: '请确认新密码' },
                        ({ getFieldValue }) => ({
                          validator(_, value) {
                            if (!value || getFieldValue('newPassword') === value) {
                              return Promise.resolve();
                            }
                            return Promise.reject(new Error('两次输入的密码不一致'));
                          }
                        })
                      ]}
                    >
                      <Input.Password placeholder="请确认新密码" />
                    </Form.Item>

                    <Form.Item>
                      <Button type="primary" icon={<LockOutlined />} htmlType="submit">
                        修改密码
                      </Button>
                    </Form.Item>
                  </Form>
                </Card>
              </Col>

              <Col span={12}>
                <Card title="安全设置">
                  <Space direction="vertical" style={{ width: '100%' }}>
                    <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                      <span>双因子认证</span>
                      <Switch
                        checked={security?.twoFactorEnabled}
                        onChange={handleToggle2FA}
                      />
                    </div>
                    <Divider />
                    <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                      <span>登录提醒</span>
                      <Switch checked={security?.loginAlerts} />
                    </div>
                    <Divider />
                    <div>
                      <span>会话超时: {security?.sessionTimeout} 分钟</span>
                    </div>
                  </Space>
                </Card>
              </Col>
            </Row>
          </TabPane>

          <TabPane tab="通知设置" key="notifications">
            <Form
              form={notificationForm}
              layout="vertical"
              onFinish={handleNotificationUpdate}
              initialValues={notifications}
            >
              <Card title="通知方式" style={{ marginBottom: '16px' }}>
                <Row gutter={16}>
                  <Col span={8}>
                    <Form.Item
                      name="emailNotifications"
                      label="邮件通知"
                      valuePropName="checked"
                    >
                      <Switch />
                    </Form.Item>
                  </Col>
                  <Col span={8}>
                    <Form.Item
                      name="smsNotifications"
                      label="短信通知"
                      valuePropName="checked"
                    >
                      <Switch />
                    </Form.Item>
                  </Col>
                  <Col span={8}>
                    <Form.Item
                      name="pushNotifications"
                      label="推送通知"
                      valuePropName="checked"
                    >
                      <Switch />
                    </Form.Item>
                  </Col>
                </Row>
              </Card>

              <Card title="通知类型" style={{ marginBottom: '16px' }}>
                <Row gutter={16}>
                  <Col span={12}>
                    <Form.Item
                      name="alertNotifications"
                      label="告警通知"
                      valuePropName="checked"
                    >
                      <Switch />
                    </Form.Item>
                  </Col>
                  <Col span={12}>
                    <Form.Item
                      name="ticketNotifications"
                      label="工单通知"
                      valuePropName="checked"
                    >
                      <Switch />
                    </Form.Item>
                  </Col>
                </Row>
                <Row gutter={16}>
                  <Col span={12}>
                    <Form.Item
                      name="systemNotifications"
                      label="系统通知"
                      valuePropName="checked"
                    >
                      <Switch />
                    </Form.Item>
                  </Col>
                </Row>
              </Card>

              <Card title="报告订阅" style={{ marginBottom: '16px' }}>
                <Row gutter={16}>
                  <Col span={12}>
                    <Form.Item
                      name="weeklyReport"
                      label="周报"
                      valuePropName="checked"
                    >
                      <Switch />
                    </Form.Item>
                  </Col>
                  <Col span={12}>
                    <Form.Item
                      name="monthlyReport"
                      label="月报"
                      valuePropName="checked"
                    >
                      <Switch />
                    </Form.Item>
                  </Col>
                </Row>
              </Card>

              <Form.Item>
                <Button type="primary" icon={<BellOutlined />} htmlType="submit">
                  保存通知设置
                </Button>
              </Form.Item>
            </Form>
          </TabPane>

          <TabPane tab="登录历史" key="history">
            <Card title="最近登录记录">
              <List
                dataSource={loginHistory}
                renderItem={(item) => (
                  <List.Item>
                    <List.Item.Meta
                      avatar={<HistoryOutlined />}
                      title={
                        <Space>
                          <span>{item.loginTime}</span>
                          <Tag color={item.success ? 'green' : 'red'}>
                            {item.success ? '成功' : '失败'}
                          </Tag>
                        </Space>
                      }
                      description={
                        <div>
                          <div>IP地址: {item.ipAddress}</div>
                          <div>位置: {item.location}</div>
                          <div>设备: {item.userAgent}</div>
                        </div>
                      }
                    />
                  </List.Item>
                )}
              />
            </Card>
          </TabPane>
        </Tabs>

        <Divider />
        
        <div style={{ textAlign: 'center' }}>
          <Button danger icon={<LogoutOutlined />} onClick={handleLogout}>
            退出登录
          </Button>
        </div>
      </Card>
    </div>
  );
};

export default ProfilePage;