import React, { useEffect, useState } from 'react';
import {
  Card,
  Form,
  Input,
  InputNumber,
  Switch,
  Button,
  Space,
  message,
  Divider,
  Row,
  Col,
  Typography,
  Select,
  TimePicker,
  Upload,
  Image,
} from 'antd';
import {
  SaveOutlined,
  ReloadOutlined,
  UploadOutlined,
  SettingOutlined,
} from '@ant-design/icons';
import type { UploadFile } from 'antd/es/upload/interface';
import { useUI } from '../../hooks';
import dayjs from 'dayjs';

const { Title, Text } = Typography;
const { Option } = Select;
const { TextArea } = Input;

interface SystemConfig {
  // 基本设置
  system_name: string;
  system_description: string;
  system_logo?: string;
  timezone: string;
  language: string;
  
  // 告警设置
  alert_retention_days: number;
  max_alerts_per_page: number;
  auto_resolve_timeout: number;
  enable_alert_grouping: boolean;
  
  // 通知设置
  email_enabled: boolean;
  email_smtp_host: string;
  email_smtp_port: number;
  email_smtp_username: string;
  email_smtp_password: string;
  email_from_address: string;
  
  sms_enabled: boolean;
  sms_provider: string;
  sms_api_key: string;
  sms_api_secret: string;
  
  webhook_enabled: boolean;
  webhook_url: string;
  webhook_timeout: number;
  
  // 安全设置
  session_timeout: number;
  password_min_length: number;
  password_require_special: boolean;
  login_max_attempts: number;
  login_lockout_duration: number;
  
  // 性能设置
  max_concurrent_requests: number;
  request_timeout: number;
  cache_enabled: boolean;
  cache_ttl: number;
  
  // 备份设置
  backup_enabled: boolean;
  backup_schedule: string;
  backup_retention_days: number;
  backup_storage_path: string;
}

const SystemSettings: React.FC = () => {
  const { setBreadcrumb } = useUI();
  const [form] = Form.useForm<SystemConfig>();
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  const [logoFileList, setLogoFileList] = useState<UploadFile[]>([]);

  useEffect(() => {
    setBreadcrumb([
      { title: '系统管理' },
      { title: '系统设置' },
    ]);
    fetchSystemConfig();
  }, [setBreadcrumb]);

  // 获取系统配置
  const fetchSystemConfig = async () => {
    setLoading(true);
    try {
      // 模拟API调用
      const mockConfig: SystemConfig = {
        system_name: 'Pulse 告警管理平台',
        system_description: '企业级告警监控与管理平台',
        system_logo: '',
        timezone: 'Asia/Shanghai',
        language: 'zh-CN',
        
        alert_retention_days: 30,
        max_alerts_per_page: 50,
        auto_resolve_timeout: 24,
        enable_alert_grouping: true,
        
        email_enabled: true,
        email_smtp_host: 'smtp.example.com',
        email_smtp_port: 587,
        email_smtp_username: 'alerts@example.com',
        email_smtp_password: '',
        email_from_address: 'alerts@example.com',
        
        sms_enabled: false,
        sms_provider: 'aliyun',
        sms_api_key: '',
        sms_api_secret: '',
        
        webhook_enabled: false,
        webhook_url: '',
        webhook_timeout: 30,
        
        session_timeout: 8,
        password_min_length: 8,
        password_require_special: true,
        login_max_attempts: 5,
        login_lockout_duration: 30,
        
        max_concurrent_requests: 100,
        request_timeout: 30,
        cache_enabled: true,
        cache_ttl: 300,
        
        backup_enabled: true,
        backup_schedule: '0 2 * * *',
        backup_retention_days: 7,
        backup_storage_path: '/data/backups',
      };
      
      form.setFieldsValue(mockConfig);
    } catch (error) {
      message.error('获取系统配置失败');
    } finally {
      setLoading(false);
    }
  };

  // 保存系统配置
  const handleSave = async (values: SystemConfig) => {
    setSaving(true);
    try {
      // 这里应该调用保存配置的API
      console.log('保存配置:', values);
      message.success('保存成功');
    } catch (error) {
      message.error('保存失败');
    } finally {
      setSaving(false);
    }
  };

  // 重置配置
  const handleReset = () => {
    form.resetFields();
    fetchSystemConfig();
  };

  // 测试邮件配置
  const handleTestEmail = async () => {
    try {
      const emailConfig = form.getFieldsValue([
        'email_smtp_host',
        'email_smtp_port',
        'email_smtp_username',
        'email_smtp_password',
        'email_from_address',
      ]);
      
      // 这里应该调用测试邮件的API
      message.success('测试邮件发送成功');
    } catch (error) {
      message.error('测试邮件发送失败');
    }
  };

  // 测试短信配置
  const handleTestSMS = async () => {
    try {
      const smsConfig = form.getFieldsValue([
        'sms_provider',
        'sms_api_key',
        'sms_api_secret',
      ]);
      
      // 这里应该调用测试短信的API
      message.success('测试短信发送成功');
    } catch (error) {
      message.error('测试短信发送失败');
    }
  };

  // 测试Webhook配置
  const handleTestWebhook = async () => {
    try {
      const webhookConfig = form.getFieldsValue([
        'webhook_url',
        'webhook_timeout',
      ]);
      
      // 这里应该调用测试Webhook的API
      message.success('Webhook测试成功');
    } catch (error) {
      message.error('Webhook测试失败');
    }
  };

  return (
    <div className="system-settings">
      <Card
        title={
          <Space>
            <SettingOutlined />
            <Title level={4} style={{ margin: 0 }}>系统设置</Title>
          </Space>
        }
        extra={
          <Space>
            <Button
              icon={<ReloadOutlined />}
              onClick={handleReset}
            >
              重置
            </Button>
            <Button
              type="primary"
              icon={<SaveOutlined />}
              loading={saving}
              onClick={() => form.submit()}
            >
              保存设置
            </Button>
          </Space>
        }
        loading={loading}
      >
        <Form
          form={form}
          layout="vertical"
          onFinish={handleSave}
        >
          {/* 基本设置 */}
          <Title level={5}>基本设置</Title>
          <Row gutter={24}>
            <Col span={12}>
              <Form.Item
                name="system_name"
                label="系统名称"
                rules={[{ required: true, message: '请输入系统名称' }]}
              >
                <Input placeholder="请输入系统名称" />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                name="timezone"
                label="时区"
                rules={[{ required: true, message: '请选择时区' }]}
              >
                <Select placeholder="请选择时区">
                  <Option value="Asia/Shanghai">Asia/Shanghai (UTC+8)</Option>
                  <Option value="UTC">UTC (UTC+0)</Option>
                  <Option value="America/New_York">America/New_York (UTC-5)</Option>
                  <Option value="Europe/London">Europe/London (UTC+0)</Option>
                </Select>
              </Form.Item>
            </Col>
          </Row>
          
          <Row gutter={24}>
            <Col span={12}>
              <Form.Item
                name="language"
                label="语言"
                rules={[{ required: true, message: '请选择语言' }]}
              >
                <Select placeholder="请选择语言">
                  <Option value="zh-CN">简体中文</Option>
                  <Option value="en-US">English</Option>
                </Select>
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                name="system_logo"
                label="系统Logo"
              >
                <Upload
                  listType="picture-card"
                  fileList={logoFileList}
                  onChange={({ fileList }) => setLogoFileList(fileList)}
                  beforeUpload={() => false}
                  maxCount={1}
                >
                  {logoFileList.length === 0 && (
                    <div>
                      <UploadOutlined />
                      <div style={{ marginTop: 8 }}>上传Logo</div>
                    </div>
                  )}
                </Upload>
              </Form.Item>
            </Col>
          </Row>
          
          <Form.Item
            name="system_description"
            label="系统描述"
          >
            <TextArea
              placeholder="请输入系统描述"
              rows={3}
            />
          </Form.Item>

          <Divider />

          {/* 告警设置 */}
          <Title level={5}>告警设置</Title>
          <Row gutter={24}>
            <Col span={8}>
              <Form.Item
                name="alert_retention_days"
                label="告警保留天数"
                rules={[{ required: true, message: '请输入告警保留天数' }]}
              >
                <InputNumber
                  min={1}
                  max={365}
                  placeholder="天数"
                  style={{ width: '100%' }}
                />
              </Form.Item>
            </Col>
            <Col span={8}>
              <Form.Item
                name="max_alerts_per_page"
                label="每页最大告警数"
                rules={[{ required: true, message: '请输入每页最大告警数' }]}
              >
                <InputNumber
                  min={10}
                  max={200}
                  placeholder="条数"
                  style={{ width: '100%' }}
                />
              </Form.Item>
            </Col>
            <Col span={8}>
              <Form.Item
                name="auto_resolve_timeout"
                label="自动解决超时（小时）"
                rules={[{ required: true, message: '请输入自动解决超时时间' }]}
              >
                <InputNumber
                  min={1}
                  max={168}
                  placeholder="小时"
                  style={{ width: '100%' }}
                />
              </Form.Item>
            </Col>
          </Row>
          
          <Row gutter={24}>
            <Col span={8}>
              <Form.Item
                name="enable_alert_grouping"
                label="启用告警分组"
                valuePropName="checked"
              >
                <Switch />
              </Form.Item>
            </Col>
          </Row>

          <Divider />

          {/* 邮件通知设置 */}
          <Title level={5}>邮件通知设置</Title>
          <Row gutter={24}>
            <Col span={8}>
              <Form.Item
                name="email_enabled"
                label="启用邮件通知"
                valuePropName="checked"
              >
                <Switch />
              </Form.Item>
            </Col>
            <Col span={16}>
              <Space>
                <Button onClick={handleTestEmail}>测试邮件配置</Button>
              </Space>
            </Col>
          </Row>
          
          <Row gutter={24}>
            <Col span={12}>
              <Form.Item
                name="email_smtp_host"
                label="SMTP服务器"
              >
                <Input placeholder="smtp.example.com" />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                name="email_smtp_port"
                label="SMTP端口"
              >
                <InputNumber
                  min={1}
                  max={65535}
                  placeholder="587"
                  style={{ width: '100%' }}
                />
              </Form.Item>
            </Col>
          </Row>
          
          <Row gutter={24}>
            <Col span={12}>
              <Form.Item
                name="email_smtp_username"
                label="SMTP用户名"
              >
                <Input placeholder="username@example.com" />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                name="email_smtp_password"
                label="SMTP密码"
              >
                <Input.Password placeholder="请输入SMTP密码" />
              </Form.Item>
            </Col>
          </Row>
          
          <Form.Item
            name="email_from_address"
            label="发件人地址"
          >
            <Input placeholder="alerts@example.com" />
          </Form.Item>

          <Divider />

          {/* 短信通知设置 */}
          <Title level={5}>短信通知设置</Title>
          <Row gutter={24}>
            <Col span={8}>
              <Form.Item
                name="sms_enabled"
                label="启用短信通知"
                valuePropName="checked"
              >
                <Switch />
              </Form.Item>
            </Col>
            <Col span={16}>
              <Space>
                <Button onClick={handleTestSMS}>测试短信配置</Button>
              </Space>
            </Col>
          </Row>
          
          <Row gutter={24}>
            <Col span={8}>
              <Form.Item
                name="sms_provider"
                label="短信服务商"
              >
                <Select placeholder="请选择短信服务商">
                  <Option value="aliyun">阿里云</Option>
                  <Option value="tencent">腾讯云</Option>
                  <Option value="huawei">华为云</Option>
                </Select>
              </Form.Item>
            </Col>
            <Col span={8}>
              <Form.Item
                name="sms_api_key"
                label="API Key"
              >
                <Input placeholder="请输入API Key" />
              </Form.Item>
            </Col>
            <Col span={8}>
              <Form.Item
                name="sms_api_secret"
                label="API Secret"
              >
                <Input.Password placeholder="请输入API Secret" />
              </Form.Item>
            </Col>
          </Row>

          <Divider />

          {/* Webhook设置 */}
          <Title level={5}>Webhook设置</Title>
          <Row gutter={24}>
            <Col span={8}>
              <Form.Item
                name="webhook_enabled"
                label="启用Webhook"
                valuePropName="checked"
              >
                <Switch />
              </Form.Item>
            </Col>
            <Col span={16}>
              <Space>
                <Button onClick={handleTestWebhook}>测试Webhook</Button>
              </Space>
            </Col>
          </Row>
          
          <Row gutter={24}>
            <Col span={18}>
              <Form.Item
                name="webhook_url"
                label="Webhook URL"
              >
                <Input placeholder="https://example.com/webhook" />
              </Form.Item>
            </Col>
            <Col span={6}>
              <Form.Item
                name="webhook_timeout"
                label="超时时间（秒）"
              >
                <InputNumber
                  min={1}
                  max={300}
                  placeholder="30"
                  style={{ width: '100%' }}
                />
              </Form.Item>
            </Col>
          </Row>

          <Divider />

          {/* 安全设置 */}
          <Title level={5}>安全设置</Title>
          <Row gutter={24}>
            <Col span={8}>
              <Form.Item
                name="session_timeout"
                label="会话超时（小时）"
                rules={[{ required: true, message: '请输入会话超时时间' }]}
              >
                <InputNumber
                  min={1}
                  max={24}
                  placeholder="8"
                  style={{ width: '100%' }}
                />
              </Form.Item>
            </Col>
            <Col span={8}>
              <Form.Item
                name="password_min_length"
                label="密码最小长度"
                rules={[{ required: true, message: '请输入密码最小长度' }]}
              >
                <InputNumber
                  min={6}
                  max={32}
                  placeholder="8"
                  style={{ width: '100%' }}
                />
              </Form.Item>
            </Col>
            <Col span={8}>
              <Form.Item
                name="password_require_special"
                label="密码需要特殊字符"
                valuePropName="checked"
              >
                <Switch />
              </Form.Item>
            </Col>
          </Row>
          
          <Row gutter={24}>
            <Col span={8}>
              <Form.Item
                name="login_max_attempts"
                label="最大登录尝试次数"
                rules={[{ required: true, message: '请输入最大登录尝试次数' }]}
              >
                <InputNumber
                  min={3}
                  max={10}
                  placeholder="5"
                  style={{ width: '100%' }}
                />
              </Form.Item>
            </Col>
            <Col span={8}>
              <Form.Item
                name="login_lockout_duration"
                label="锁定时长（分钟）"
                rules={[{ required: true, message: '请输入锁定时长' }]}
              >
                <InputNumber
                  min={5}
                  max={1440}
                  placeholder="30"
                  style={{ width: '100%' }}
                />
              </Form.Item>
            </Col>
          </Row>

          <Divider />

          {/* 性能设置 */}
          <Title level={5}>性能设置</Title>
          <Row gutter={24}>
            <Col span={8}>
              <Form.Item
                name="max_concurrent_requests"
                label="最大并发请求数"
                rules={[{ required: true, message: '请输入最大并发请求数' }]}
              >
                <InputNumber
                  min={10}
                  max={1000}
                  placeholder="100"
                  style={{ width: '100%' }}
                />
              </Form.Item>
            </Col>
            <Col span={8}>
              <Form.Item
                name="request_timeout"
                label="请求超时（秒）"
                rules={[{ required: true, message: '请输入请求超时时间' }]}
              >
                <InputNumber
                  min={5}
                  max={300}
                  placeholder="30"
                  style={{ width: '100%' }}
                />
              </Form.Item>
            </Col>
            <Col span={8}>
              <Form.Item
                name="cache_enabled"
                label="启用缓存"
                valuePropName="checked"
              >
                <Switch />
              </Form.Item>
            </Col>
          </Row>
          
          <Row gutter={24}>
            <Col span={8}>
              <Form.Item
                name="cache_ttl"
                label="缓存TTL（秒）"
              >
                <InputNumber
                  min={60}
                  max={3600}
                  placeholder="300"
                  style={{ width: '100%' }}
                />
              </Form.Item>
            </Col>
          </Row>

          <Divider />

          {/* 备份设置 */}
          <Title level={5}>备份设置</Title>
          <Row gutter={24}>
            <Col span={8}>
              <Form.Item
                name="backup_enabled"
                label="启用自动备份"
                valuePropName="checked"
              >
                <Switch />
              </Form.Item>
            </Col>
            <Col span={8}>
              <Form.Item
                name="backup_retention_days"
                label="备份保留天数"
              >
                <InputNumber
                  min={1}
                  max={365}
                  placeholder="7"
                  style={{ width: '100%' }}
                />
              </Form.Item>
            </Col>
          </Row>
          
          <Row gutter={24}>
            <Col span={12}>
              <Form.Item
                name="backup_schedule"
                label="备份计划（Cron表达式）"
              >
                <Input placeholder="0 2 * * *" />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                name="backup_storage_path"
                label="备份存储路径"
              >
                <Input placeholder="/data/backups" />
              </Form.Item>
            </Col>
          </Row>
        </Form>
      </Card>
    </div>
  );
};

export default SystemSettings;