import React, { useEffect, useState } from 'react';
import {
  Card,
  Form,
  Input,
  Button,
  Space,
  message,
  Row,
  Col,
  Typography,
  Avatar,
  Upload,
  Divider,
  Select,
  Switch,
  TimePicker,
  Tabs,
} from 'antd';
import {
  UserOutlined,
  SaveOutlined,
  UploadOutlined,
  LockOutlined,
  SettingOutlined,
  BellOutlined,
} from '@ant-design/icons';
import type { UploadFile } from 'antd/es/upload/interface';
import { useUI } from '../../hooks';
import { User } from '../../types';
import dayjs from 'dayjs';

const { Title, Text } = Typography;
const { Option } = Select;
const { TabPane } = Tabs;

interface ProfileFormData {
  username: string;
  email: string;
  phone?: string;
  real_name?: string;
  department?: string;
  position?: string;
  avatar?: string;
}

interface PasswordFormData {
  current_password: string;
  new_password: string;
  confirm_password: string;
}

interface NotificationSettings {
  email_alerts: boolean;
  sms_alerts: boolean;
  webhook_alerts: boolean;
  alert_levels: string[];
  quiet_hours_enabled: boolean;
  quiet_hours_start: string;
  quiet_hours_end: string;
  digest_enabled: boolean;
  digest_frequency: string;
}

const Profile: React.FC = () => {
  const { setBreadcrumb } = useUI();
  const [profileForm] = Form.useForm<ProfileFormData>();
  const [passwordForm] = Form.useForm<PasswordFormData>();
  const [notificationForm] = Form.useForm<NotificationSettings>();
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  const [avatarFileList, setAvatarFileList] = useState<UploadFile[]>([]);
  const [currentUser, setCurrentUser] = useState<User | null>(null);

  useEffect(() => {
    setBreadcrumb([
      { title: '系统管理' },
      { title: '个人资料' },
    ]);
    fetchUserProfile();
  }, [setBreadcrumb]);

  // 获取用户资料
  const fetchUserProfile = async () => {
    setLoading(true);
    try {
      // 模拟API调用
      const mockUser: User = {
        id: '1',
        username: 'admin',
        email: 'admin@example.com',
        phone: '13800138000',
        real_name: '系统管理员',
        department: 'IT部门',
        position: '系统管理员',
        role: 'admin',
        status: 'active',
        avatar: '',
        created_at: '2024-01-01T00:00:00Z',
        updated_at: '2024-01-01T00:00:00Z',
        last_login: '2024-01-15T10:30:00Z',
      };
      
      setCurrentUser(mockUser);
      profileForm.setFieldsValue({
        username: mockUser.username,
        email: mockUser.email,
        phone: mockUser.phone,
        real_name: mockUser.real_name,
        department: mockUser.department,
        position: mockUser.position,
        avatar: mockUser.avatar,
      });
      
      // 设置通知设置
      const mockNotificationSettings: NotificationSettings = {
        email_alerts: true,
        sms_alerts: false,
        webhook_alerts: false,
        alert_levels: ['critical', 'warning'],
        quiet_hours_enabled: true,
        quiet_hours_start: '22:00',
        quiet_hours_end: '08:00',
        digest_enabled: true,
        digest_frequency: 'daily',
      };
      
      notificationForm.setFieldsValue({
        ...mockNotificationSettings,
        quiet_hours_start: dayjs(mockNotificationSettings.quiet_hours_start, 'HH:mm'),
        quiet_hours_end: dayjs(mockNotificationSettings.quiet_hours_end, 'HH:mm'),
      });
    } catch (error) {
      message.error('获取用户资料失败');
    } finally {
      setLoading(false);
    }
  };

  // 更新个人资料
  const handleUpdateProfile = async (values: ProfileFormData) => {
    setSaving(true);
    try {
      // 这里应该调用更新用户资料的API
      console.log('更新资料:', values);
      message.success('个人资料更新成功');
    } catch (error) {
      message.error('更新个人资料失败');
    } finally {
      setSaving(false);
    }
  };

  // 修改密码
  const handleChangePassword = async (values: PasswordFormData) => {
    setSaving(true);
    try {
      // 这里应该调用修改密码的API
      console.log('修改密码:', values);
      message.success('密码修改成功');
      passwordForm.resetFields();
    } catch (error) {
      message.error('密码修改失败');
    } finally {
      setSaving(false);
    }
  };

  // 更新通知设置
  const handleUpdateNotifications = async (values: NotificationSettings) => {
    setSaving(true);
    try {
      // 转换时间格式
      const formattedValues = {
        ...values,
        quiet_hours_start: values.quiet_hours_start?.format('HH:mm'),
        quiet_hours_end: values.quiet_hours_end?.format('HH:mm'),
      };
      
      // 这里应该调用更新通知设置的API
      console.log('更新通知设置:', formattedValues);
      message.success('通知设置更新成功');
    } catch (error) {
      message.error('更新通知设置失败');
    } finally {
      setSaving(false);
    }
  };

  // 头像上传前处理
  const beforeAvatarUpload = (file: File) => {
    const isJpgOrPng = file.type === 'image/jpeg' || file.type === 'image/png';
    if (!isJpgOrPng) {
      message.error('只能上传 JPG/PNG 格式的图片!');
      return false;
    }
    const isLt2M = file.size / 1024 / 1024 < 2;
    if (!isLt2M) {
      message.error('图片大小不能超过 2MB!');
      return false;
    }
    return false; // 阻止自动上传
  };

  return (
    <div className="profile">
      <Card loading={loading}>
        <Tabs defaultActiveKey="profile">
          {/* 个人资料 */}
          <TabPane
            tab={
              <Space>
                <UserOutlined />
                个人资料
              </Space>
            }
            key="profile"
          >
            <Row gutter={24}>
              <Col span={6}>
                <div style={{ textAlign: 'center' }}>
                  <Avatar
                    size={120}
                    src={currentUser?.avatar}
                    icon={<UserOutlined />}
                    style={{ marginBottom: 16 }}
                  />
                  <br />
                  <Upload
                    showUploadList={false}
                    beforeUpload={beforeAvatarUpload}
                    onChange={({ fileList }) => setAvatarFileList(fileList)}
                  >
                    <Button icon={<UploadOutlined />}>更换头像</Button>
                  </Upload>
                  <div style={{ marginTop: 16, color: '#666' }}>
                    <Text type="secondary">支持 JPG、PNG 格式</Text>
                    <br />
                    <Text type="secondary">文件大小不超过 2MB</Text>
                  </div>
                </div>
              </Col>
              <Col span={18}>
                <Form
                  form={profileForm}
                  layout="vertical"
                  onFinish={handleUpdateProfile}
                >
                  <Row gutter={24}>
                    <Col span={12}>
                      <Form.Item
                        name="username"
                        label="用户名"
                        rules={[{ required: true, message: '请输入用户名' }]}
                      >
                        <Input disabled placeholder="用户名" />
                      </Form.Item>
                    </Col>
                    <Col span={12}>
                      <Form.Item
                        name="email"
                        label="邮箱"
                        rules={[
                          { required: true, message: '请输入邮箱' },
                          { type: 'email', message: '请输入有效的邮箱地址' },
                        ]}
                      >
                        <Input placeholder="邮箱" />
                      </Form.Item>
                    </Col>
                  </Row>
                  
                  <Row gutter={24}>
                    <Col span={12}>
                      <Form.Item
                        name="real_name"
                        label="真实姓名"
                      >
                        <Input placeholder="真实姓名" />
                      </Form.Item>
                    </Col>
                    <Col span={12}>
                      <Form.Item
                        name="phone"
                        label="手机号"
                        rules={[
                          { pattern: /^1[3-9]\d{9}$/, message: '请输入有效的手机号' },
                        ]}
                      >
                        <Input placeholder="手机号" />
                      </Form.Item>
                    </Col>
                  </Row>
                  
                  <Row gutter={24}>
                    <Col span={12}>
                      <Form.Item
                        name="department"
                        label="部门"
                      >
                        <Input placeholder="部门" />
                      </Form.Item>
                    </Col>
                    <Col span={12}>
                      <Form.Item
                        name="position"
                        label="职位"
                      >
                        <Input placeholder="职位" />
                      </Form.Item>
                    </Col>
                  </Row>
                  
                  <Form.Item>
                    <Button
                      type="primary"
                      htmlType="submit"
                      icon={<SaveOutlined />}
                      loading={saving}
                    >
                      保存资料
                    </Button>
                  </Form.Item>
                </Form>
              </Col>
            </Row>
          </TabPane>

          {/* 修改密码 */}
          <TabPane
            tab={
              <Space>
                <LockOutlined />
                修改密码
              </Space>
            }
            key="password"
          >
            <Row justify="center">
              <Col span={12}>
                <Form
                  form={passwordForm}
                  layout="vertical"
                  onFinish={handleChangePassword}
                >
                  <Form.Item
                    name="current_password"
                    label="当前密码"
                    rules={[{ required: true, message: '请输入当前密码' }]}
                  >
                    <Input.Password placeholder="请输入当前密码" />
                  </Form.Item>
                  
                  <Form.Item
                    name="new_password"
                    label="新密码"
                    rules={[
                      { required: true, message: '请输入新密码' },
                      { min: 8, message: '密码长度至少8位' },
                      {
                        pattern: /^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)(?=.*[@$!%*?&])[A-Za-z\d@$!%*?&]/,
                        message: '密码必须包含大小写字母、数字和特殊字符',
                      },
                    ]}
                  >
                    <Input.Password placeholder="请输入新密码" />
                  </Form.Item>
                  
                  <Form.Item
                    name="confirm_password"
                    label="确认新密码"
                    dependencies={['new_password']}
                    rules={[
                      { required: true, message: '请确认新密码' },
                      ({ getFieldValue }) => ({
                        validator(_, value) {
                          if (!value || getFieldValue('new_password') === value) {
                            return Promise.resolve();
                          }
                          return Promise.reject(new Error('两次输入的密码不一致'));
                        },
                      }),
                    ]}
                  >
                    <Input.Password placeholder="请确认新密码" />
                  </Form.Item>
                  
                  <Form.Item>
                    <Button
                      type="primary"
                      htmlType="submit"
                      icon={<SaveOutlined />}
                      loading={saving}
                    >
                      修改密码
                    </Button>
                  </Form.Item>
                </Form>
              </Col>
            </Row>
          </TabPane>

          {/* 通知设置 */}
          <TabPane
            tab={
              <Space>
                <BellOutlined />
                通知设置
              </Space>
            }
            key="notifications"
          >
            <Form
              form={notificationForm}
              layout="vertical"
              onFinish={handleUpdateNotifications}
            >
              <Title level={5}>通知方式</Title>
              <Row gutter={24}>
                <Col span={8}>
                  <Form.Item
                    name="email_alerts"
                    label="邮件通知"
                    valuePropName="checked"
                  >
                    <Switch />
                  </Form.Item>
                </Col>
                <Col span={8}>
                  <Form.Item
                    name="sms_alerts"
                    label="短信通知"
                    valuePropName="checked"
                  >
                    <Switch />
                  </Form.Item>
                </Col>
                <Col span={8}>
                  <Form.Item
                    name="webhook_alerts"
                    label="Webhook通知"
                    valuePropName="checked"
                  >
                    <Switch />
                  </Form.Item>
                </Col>
              </Row>
              
              <Divider />
              
              <Title level={5}>告警级别</Title>
              <Form.Item
                name="alert_levels"
                label="接收的告警级别"
              >
                <Select
                  mode="multiple"
                  placeholder="请选择要接收的告警级别"
                  style={{ width: '100%' }}
                >
                  <Option value="critical">严重</Option>
                  <Option value="warning">警告</Option>
                  <Option value="info">信息</Option>
                </Select>
              </Form.Item>
              
              <Divider />
              
              <Title level={5}>免打扰时间</Title>
              <Row gutter={24}>
                <Col span={8}>
                  <Form.Item
                    name="quiet_hours_enabled"
                    label="启用免打扰"
                    valuePropName="checked"
                  >
                    <Switch />
                  </Form.Item>
                </Col>
                <Col span={8}>
                  <Form.Item
                    name="quiet_hours_start"
                    label="开始时间"
                  >
                    <TimePicker
                      format="HH:mm"
                      placeholder="选择开始时间"
                      style={{ width: '100%' }}
                    />
                  </Form.Item>
                </Col>
                <Col span={8}>
                  <Form.Item
                    name="quiet_hours_end"
                    label="结束时间"
                  >
                    <TimePicker
                      format="HH:mm"
                      placeholder="选择结束时间"
                      style={{ width: '100%' }}
                    />
                  </Form.Item>
                </Col>
              </Row>
              
              <Divider />
              
              <Title level={5}>摘要报告</Title>
              <Row gutter={24}>
                <Col span={8}>
                  <Form.Item
                    name="digest_enabled"
                    label="启用摘要报告"
                    valuePropName="checked"
                  >
                    <Switch />
                  </Form.Item>
                </Col>
                <Col span={8}>
                  <Form.Item
                    name="digest_frequency"
                    label="发送频率"
                  >
                    <Select placeholder="请选择发送频率">
                      <Option value="daily">每日</Option>
                      <Option value="weekly">每周</Option>
                      <Option value="monthly">每月</Option>
                    </Select>
                  </Form.Item>
                </Col>
              </Row>
              
              <Form.Item>
                <Button
                  type="primary"
                  htmlType="submit"
                  icon={<SaveOutlined />}
                  loading={saving}
                >
                  保存设置
                </Button>
              </Form.Item>
            </Form>
          </TabPane>
        </Tabs>
      </Card>
    </div>
  );
};

export default Profile;