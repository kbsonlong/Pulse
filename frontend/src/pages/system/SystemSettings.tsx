import React, { useState, useEffect } from 'react';
import {
  Card,
  Form,
  Input,
  Button,
  Switch,
  Select,
  InputNumber,
  message,
  Tabs,
  Space,
  Divider,
  Row,
  Col,
  Alert,
  Modal,
  Upload,
  Progress
} from 'antd';
import {
  SaveOutlined,
  ReloadOutlined,
  PlayCircleOutlined,
  DownloadOutlined,
  UploadOutlined,
  DeleteOutlined,
  ExclamationCircleOutlined
} from '@ant-design/icons';
import { useUI } from '../../hooks';
import { systemService } from '../../services';

const { TabPane } = Tabs;
const { Option } = Select;
const { TextArea } = Input;

interface SystemSettings {
  // 基本设置
  siteName: string;
  siteDescription: string;
  adminEmail: string;
  timezone: string;
  language: string;
  
  // 告警设置
  alertRetentionDays: number;
  maxAlertsPerPage: number;
  autoResolveHours: number;
  enableAlertGrouping: boolean;
  
  // 邮件设置
  emailEnabled: boolean;
  smtpHost: string;
  smtpPort: number;
  smtpUsername: string;
  smtpPassword: string;
  smtpSsl: boolean;
  
  // 短信设置
  smsEnabled: boolean;
  smsProvider: string;
  smsApiKey: string;
  smsApiSecret: string;
  
  // Webhook设置
  webhookEnabled: boolean;
  webhookUrl: string;
  webhookSecret: string;
  webhookTimeout: number;
  
  // 安全设置
  sessionTimeout: number;
  passwordMinLength: number;
  passwordRequireSpecial: boolean;
  enableTwoFactor: boolean;
  maxLoginAttempts: number;
  
  // 性能设置
  cacheEnabled: boolean;
  cacheTtl: number;
  maxConcurrentUsers: number;
  enableCompression: boolean;
}

const SystemSettingsPage: React.FC = () => {
  const { loading, setLoading } = useUI();
  const [settings, setSettings] = useState<SystemSettings | null>(null);
  const [activeTab, setActiveTab] = useState('basic');
  const [testResults, setTestResults] = useState<Record<string, any>>({});
  const [backupProgress, setBackupProgress] = useState(0);
  const [form] = Form.useForm();

  useEffect(() => {
    loadSettings();
  }, []);

  const loadSettings = async () => {
    try {
      setLoading(true);
      const response = await systemService.getSettings();
      setSettings(response.data);
      form.setFieldsValue(response.data);
    } catch (error) {
      message.error('加载系统设置失败');
    } finally {
      setLoading(false);
    }
  };

  const handleSave = async (values: SystemSettings) => {
    try {
      await systemService.updateSettings(values);
      message.success('设置保存成功');
      setSettings(values);
    } catch (error) {
      message.error('设置保存失败');
    }
  };

  const handleReset = () => {
    Modal.confirm({
      title: '确认重置',
      icon: <ExclamationCircleOutlined />,
      content: '确定要重置所有设置到默认值吗？此操作不可恢复。',
      okText: '确定',
      cancelText: '取消',
      onOk: async () => {
        try {
          await systemService.resetSettings();
          message.success('设置重置成功');
          loadSettings();
        } catch (error) {
          message.error('设置重置失败');
        }
      }
    });
  };

  const handleTestEmail = async () => {
    try {
      const values = form.getFieldsValue();
      const result = await systemService.testEmail({
        smtpHost: values.smtpHost,
        smtpPort: values.smtpPort,
        smtpUsername: values.smtpUsername,
        smtpPassword: values.smtpPassword,
        smtpSsl: values.smtpSsl
      });
      setTestResults({ ...testResults, email: result.data });
      message.success('邮件测试成功');
    } catch (error) {
      message.error('邮件测试失败');
    }
  };

  const handleTestSms = async () => {
    try {
      const values = form.getFieldsValue();
      const result = await systemService.testSms({
        smsProvider: values.smsProvider,
        smsApiKey: values.smsApiKey,
        smsApiSecret: values.smsApiSecret
      });
      setTestResults({ ...testResults, sms: result.data });
      message.success('短信测试成功');
    } catch (error) {
      message.error('短信测试失败');
    }
  };

  const handleTestWebhook = async () => {
    try {
      const values = form.getFieldsValue();
      const result = await systemService.testWebhook({
        webhookUrl: values.webhookUrl,
        webhookSecret: values.webhookSecret,
        webhookTimeout: values.webhookTimeout
      });
      setTestResults({ ...testResults, webhook: result.data });
      message.success('Webhook测试成功');
    } catch (error) {
      message.error('Webhook测试失败');
    }
  };

  const handleBackup = async () => {
    try {
      setBackupProgress(0);
      const response = await systemService.createBackup();
      
      // 模拟进度更新
      const interval = setInterval(() => {
        setBackupProgress(prev => {
          if (prev >= 100) {
            clearInterval(interval);
            return 100;
          }
          return prev + 10;
        });
      }, 500);
      
      message.success('备份创建成功');
    } catch (error) {
      message.error('备份创建失败');
    }
  };

  const handleRestart = () => {
    Modal.confirm({
      title: '确认重启',
      icon: <ExclamationCircleOutlined />,
      content: '确定要重启系统服务吗？这将导致短暂的服务中断。',
      okText: '确定',
      cancelText: '取消',
      onOk: async () => {
        try {
          await systemService.restart();
          message.success('系统重启指令已发送');
        } catch (error) {
          message.error('系统重启失败');
        }
      }
    });
  };

  return (
    <div style={{ padding: '24px' }}>
      <Card>
        <Form
          form={form}
          layout="vertical"
          onFinish={handleSave}
          initialValues={settings}
        >
          <div style={{ marginBottom: '16px' }}>
            <Space>
              <Button type="primary" icon={<SaveOutlined />} htmlType="submit">
                保存设置
              </Button>
              <Button icon={<ReloadOutlined />} onClick={loadSettings}>
                重新加载
              </Button>
              <Button danger icon={<DeleteOutlined />} onClick={handleReset}>
                重置设置
              </Button>
            </Space>
          </div>

          <Tabs activeKey={activeTab} onChange={setActiveTab}>
            <TabPane tab="基本设置" key="basic">
              <Row gutter={24}>
                <Col span={12}>
                  <Form.Item
                    name="siteName"
                    label="站点名称"
                    rules={[{ required: true, message: '请输入站点名称' }]}
                  >
                    <Input placeholder="请输入站点名称" />
                  </Form.Item>
                </Col>
                <Col span={12}>
                  <Form.Item
                    name="adminEmail"
                    label="管理员邮箱"
                    rules={[
                      { required: true, message: '请输入管理员邮箱' },
                      { type: 'email', message: '请输入有效的邮箱地址' }
                    ]}
                  >
                    <Input placeholder="请输入管理员邮箱" />
                  </Form.Item>
                </Col>
              </Row>

              <Form.Item
                name="siteDescription"
                label="站点描述"
              >
                <TextArea rows={3} placeholder="请输入站点描述" />
              </Form.Item>

              <Row gutter={24}>
                <Col span={12}>
                  <Form.Item
                    name="timezone"
                    label="时区"
                    rules={[{ required: true, message: '请选择时区' }]}
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
                    label="默认语言"
                    rules={[{ required: true, message: '请选择默认语言' }]}
                  >
                    <Select placeholder="请选择默认语言">
                      <Option value="zh-CN">简体中文</Option>
                      <Option value="en-US">English</Option>
                    </Select>
                  </Form.Item>
                </Col>
              </Row>
            </TabPane>

            <TabPane tab="告警设置" key="alert">
              <Row gutter={24}>
                <Col span={12}>
                  <Form.Item
                    name="alertRetentionDays"
                    label="告警保留天数"
                    rules={[{ required: true, message: '请输入告警保留天数' }]}
                  >
                    <InputNumber min={1} max={365} style={{ width: '100%' }} />
                  </Form.Item>
                </Col>
                <Col span={12}>
                  <Form.Item
                    name="maxAlertsPerPage"
                    label="每页最大告警数"
                    rules={[{ required: true, message: '请输入每页最大告警数' }]}
                  >
                    <InputNumber min={10} max={100} style={{ width: '100%' }} />
                  </Form.Item>
                </Col>
              </Row>

              <Row gutter={24}>
                <Col span={12}>
                  <Form.Item
                    name="autoResolveHours"
                    label="自动解决时间(小时)"
                  >
                    <InputNumber min={0} max={168} style={{ width: '100%' }} />
                  </Form.Item>
                </Col>
                <Col span={12}>
                  <Form.Item
                    name="enableAlertGrouping"
                    label="启用告警分组"
                    valuePropName="checked"
                  >
                    <Switch />
                  </Form.Item>
                </Col>
              </Row>
            </TabPane>

            <TabPane tab="邮件设置" key="email">
              <Form.Item
                name="emailEnabled"
                label="启用邮件通知"
                valuePropName="checked"
              >
                <Switch />
              </Form.Item>

              <Row gutter={24}>
                <Col span={12}>
                  <Form.Item
                    name="smtpHost"
                    label="SMTP服务器"
                    rules={[{ required: true, message: '请输入SMTP服务器' }]}
                  >
                    <Input placeholder="请输入SMTP服务器" />
                  </Form.Item>
                </Col>
                <Col span={12}>
                  <Form.Item
                    name="smtpPort"
                    label="SMTP端口"
                    rules={[{ required: true, message: '请输入SMTP端口' }]}
                  >
                    <InputNumber min={1} max={65535} style={{ width: '100%' }} />
                  </Form.Item>
                </Col>
              </Row>

              <Row gutter={24}>
                <Col span={12}>
                  <Form.Item
                    name="smtpUsername"
                    label="SMTP用户名"
                    rules={[{ required: true, message: '请输入SMTP用户名' }]}
                  >
                    <Input placeholder="请输入SMTP用户名" />
                  </Form.Item>
                </Col>
                <Col span={12}>
                  <Form.Item
                    name="smtpPassword"
                    label="SMTP密码"
                    rules={[{ required: true, message: '请输入SMTP密码' }]}
                  >
                    <Input.Password placeholder="请输入SMTP密码" />
                  </Form.Item>
                </Col>
              </Row>

              <Form.Item
                name="smtpSsl"
                label="启用SSL"
                valuePropName="checked"
              >
                <Switch />
              </Form.Item>

              <Form.Item>
                <Button icon={<PlayCircleOutlined />} onClick={handleTestEmail}>
                  测试邮件配置
                </Button>
              </Form.Item>

              {testResults.email && (
                <Alert
                  message="邮件测试结果"
                  description={JSON.stringify(testResults.email, null, 2)}
                  type={testResults.email.success ? 'success' : 'error'}
                  style={{ marginTop: '16px' }}
                />
              )}
            </TabPane>

            <TabPane tab="短信设置" key="sms">
              <Form.Item
                name="smsEnabled"
                label="启用短信通知"
                valuePropName="checked"
              >
                <Switch />
              </Form.Item>

              <Row gutter={24}>
                <Col span={12}>
                  <Form.Item
                    name="smsProvider"
                    label="短信服务商"
                    rules={[{ required: true, message: '请选择短信服务商' }]}
                  >
                    <Select placeholder="请选择短信服务商">
                      <Option value="aliyun">阿里云</Option>
                      <Option value="tencent">腾讯云</Option>
                      <Option value="huawei">华为云</Option>
                    </Select>
                  </Form.Item>
                </Col>
                <Col span={12}>
                  <Form.Item
                    name="smsApiKey"
                    label="API Key"
                    rules={[{ required: true, message: '请输入API Key' }]}
                  >
                    <Input placeholder="请输入API Key" />
                  </Form.Item>
                </Col>
              </Row>

              <Form.Item
                name="smsApiSecret"
                label="API Secret"
                rules={[{ required: true, message: '请输入API Secret' }]}
              >
                <Input.Password placeholder="请输入API Secret" />
              </Form.Item>

              <Form.Item>
                <Button icon={<PlayCircleOutlined />} onClick={handleTestSms}>
                  测试短信配置
                </Button>
              </Form.Item>

              {testResults.sms && (
                <Alert
                  message="短信测试结果"
                  description={JSON.stringify(testResults.sms, null, 2)}
                  type={testResults.sms.success ? 'success' : 'error'}
                  style={{ marginTop: '16px' }}
                />
              )}
            </TabPane>

            <TabPane tab="Webhook设置" key="webhook">
              <Form.Item
                name="webhookEnabled"
                label="启用Webhook"
                valuePropName="checked"
              >
                <Switch />
              </Form.Item>

              <Form.Item
                name="webhookUrl"
                label="Webhook URL"
                rules={[{ required: true, message: '请输入Webhook URL' }]}
              >
                <Input placeholder="请输入Webhook URL" />
              </Form.Item>

              <Row gutter={24}>
                <Col span={12}>
                  <Form.Item
                    name="webhookSecret"
                    label="Webhook密钥"
                  >
                    <Input.Password placeholder="请输入Webhook密钥" />
                  </Form.Item>
                </Col>
                <Col span={12}>
                  <Form.Item
                    name="webhookTimeout"
                    label="超时时间(秒)"
                    rules={[{ required: true, message: '请输入超时时间' }]}
                  >
                    <InputNumber min={1} max={60} style={{ width: '100%' }} />
                  </Form.Item>
                </Col>
              </Row>

              <Form.Item>
                <Button icon={<PlayCircleOutlined />} onClick={handleTestWebhook}>
                  测试Webhook配置
                </Button>
              </Form.Item>

              {testResults.webhook && (
                <Alert
                  message="Webhook测试结果"
                  description={JSON.stringify(testResults.webhook, null, 2)}
                  type={testResults.webhook.success ? 'success' : 'error'}
                  style={{ marginTop: '16px' }}
                />
              )}
            </TabPane>

            <TabPane tab="安全设置" key="security">
              <Row gutter={24}>
                <Col span={12}>
                  <Form.Item
                    name="sessionTimeout"
                    label="会话超时时间(分钟)"
                    rules={[{ required: true, message: '请输入会话超时时间' }]}
                  >
                    <InputNumber min={5} max={1440} style={{ width: '100%' }} />
                  </Form.Item>
                </Col>
                <Col span={12}>
                  <Form.Item
                    name="passwordMinLength"
                    label="密码最小长度"
                    rules={[{ required: true, message: '请输入密码最小长度' }]}
                  >
                    <InputNumber min={6} max={32} style={{ width: '100%' }} />
                  </Form.Item>
                </Col>
              </Row>

              <Row gutter={24}>
                <Col span={12}>
                  <Form.Item
                    name="passwordRequireSpecial"
                    label="密码需要特殊字符"
                    valuePropName="checked"
                  >
                    <Switch />
                  </Form.Item>
                </Col>
                <Col span={12}>
                  <Form.Item
                    name="enableTwoFactor"
                    label="启用双因子认证"
                    valuePropName="checked"
                  >
                    <Switch />
                  </Form.Item>
                </Col>
              </Row>

              <Form.Item
                name="maxLoginAttempts"
                label="最大登录尝试次数"
                rules={[{ required: true, message: '请输入最大登录尝试次数' }]}
              >
                <InputNumber min={3} max={10} style={{ width: '100%' }} />
              </Form.Item>
            </TabPane>

            <TabPane tab="系统维护" key="maintenance">
              <Card title="备份管理" style={{ marginBottom: '16px' }}>
                <Space direction="vertical" style={{ width: '100%' }}>
                  <Button icon={<DownloadOutlined />} onClick={handleBackup}>
                    创建备份
                  </Button>
                  {backupProgress > 0 && (
                    <Progress percent={backupProgress} status={backupProgress === 100 ? 'success' : 'active'} />
                  )}
                  <Upload
                    accept=".zip,.tar.gz"
                    showUploadList={false}
                    beforeUpload={() => false}
                  >
                    <Button icon={<UploadOutlined />}>
                      恢复备份
                    </Button>
                  </Upload>
                </Space>
              </Card>

              <Card title="系统操作">
                <Space>
                  <Button danger onClick={handleRestart}>
                    重启系统
                  </Button>
                </Space>
              </Card>
            </TabPane>
          </Tabs>
        </Form>
      </Card>
    </div>
  );
};

export default SystemSettingsPage;