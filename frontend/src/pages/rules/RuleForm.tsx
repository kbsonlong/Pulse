import React, { useEffect, useState } from 'react';
import {
  Card,
  Form,
  Input,
  Select,
  InputNumber,
  Switch,
  Button,
  Space,
  message,
  Row,
  Col,
  Typography,
  Divider,
  Tag,
  Alert,
  Collapse,
} from 'antd';
import {
  SaveOutlined,
  ArrowLeftOutlined,
  PlayCircleOutlined,
  InfoCircleOutlined,
} from '@ant-design/icons';
import { useNavigate, useParams } from 'react-router-dom';
import { useRule, useUI } from '../../hooks';
import { Rule, DataSource } from '../../types';

const { Title, Text } = Typography;
const { Option } = Select;
const { TextArea } = Input;
const { Panel } = Collapse;

interface RuleFormData {
  name: string;
  description?: string;
  datasource_id: string;
  query: string;
  condition: string;
  threshold: number;
  severity: 'info' | 'warning' | 'critical';
  enabled: boolean;
  interval: number;
  timeout: number;
  labels?: Record<string, string>;
  annotations?: Record<string, string>;
  notification_channels?: string[];
}

const RuleForm: React.FC = () => {
  const navigate = useNavigate();
  const { id } = useParams<{ id: string }>();
  const { setBreadcrumb } = useUI();
  const {
    currentRule,
    loading,
    error,
    fetchRule,
    createRule,
    updateRule,
    clearCurrentRule,
    clearError,
  } = useRule();
  
  const [form] = Form.useForm<RuleFormData>();
  const [saving, setSaving] = useState(false);
  const [testing, setTesting] = useState(false);
  const [testResult, setTestResult] = useState<any>(null);
  const [dataSources, setDataSources] = useState<DataSource[]>([]);
  const [labelInputs, setLabelInputs] = useState<Array<{ key: string; value: string }>>([]);
  const [annotationInputs, setAnnotationInputs] = useState<Array<{ key: string; value: string }>>([]);

  const isEdit = !!id;

  useEffect(() => {
    setBreadcrumb([
      { title: '规则管理', path: '/rules' },
      { title: isEdit ? '编辑规则' : '创建规则' },
    ]);

    if (isEdit && id) {
      fetchRule(id);
    }

    fetchDataSources();

    return () => {
      clearCurrentRule();
      clearError();
    };
  }, [setBreadcrumb, isEdit, id, fetchRule, clearCurrentRule, clearError]);

  useEffect(() => {
    if (currentRule && isEdit) {
      form.setFieldsValue({
        name: currentRule.name,
        description: currentRule.description,
        datasource_id: currentRule.datasource_id,
        query: currentRule.query,
        condition: currentRule.condition,
        threshold: currentRule.threshold,
        severity: currentRule.severity,
        enabled: currentRule.enabled,
        interval: currentRule.interval,
        timeout: currentRule.timeout,
        notification_channels: currentRule.notification_channels,
      });

      // 设置标签和注释
      if (currentRule.labels) {
        const labels = Object.entries(currentRule.labels).map(([key, value]) => ({ key, value }));
        setLabelInputs(labels);
      }
      if (currentRule.annotations) {
        const annotations = Object.entries(currentRule.annotations).map(([key, value]) => ({ key, value }));
        setAnnotationInputs(annotations);
      }
    }
  }, [currentRule, isEdit, form]);

  // 获取数据源列表
  const fetchDataSources = async () => {
    try {
      // 模拟API调用
      const mockDataSources: DataSource[] = [
        {
          id: '1',
          name: 'Prometheus',
          type: 'prometheus',
          url: 'http://prometheus:9090',
          status: 'active',
          created_at: '2024-01-01T00:00:00Z',
          updated_at: '2024-01-01T00:00:00Z',
        },
        {
          id: '2',
          name: 'InfluxDB',
          type: 'influxdb',
          url: 'http://influxdb:8086',
          status: 'active',
          created_at: '2024-01-01T00:00:00Z',
          updated_at: '2024-01-01T00:00:00Z',
        },
      ];
      setDataSources(mockDataSources);
    } catch (error) {
      message.error('获取数据源列表失败');
    }
  };

  // 保存规则
  const handleSave = async (values: RuleFormData) => {
    setSaving(true);
    try {
      // 处理标签和注释
      const labels: Record<string, string> = {};
      const annotations: Record<string, string> = {};
      
      labelInputs.forEach(({ key, value }) => {
        if (key && value) {
          labels[key] = value;
        }
      });
      
      annotationInputs.forEach(({ key, value }) => {
        if (key && value) {
          annotations[key] = value;
        }
      });

      const ruleData = {
        ...values,
        labels: Object.keys(labels).length > 0 ? labels : undefined,
        annotations: Object.keys(annotations).length > 0 ? annotations : undefined,
      };

      if (isEdit && id) {
        await updateRule(id, ruleData);
        message.success('规则更新成功');
      } else {
        await createRule(ruleData);
        message.success('规则创建成功');
      }
      
      navigate('/rules/list');
    } catch (error) {
      message.error(isEdit ? '规则更新失败' : '规则创建失败');
    } finally {
      setSaving(false);
    }
  };

  // 测试规则
  const handleTestRule = async () => {
    setTesting(true);
    try {
      const values = await form.validateFields(['datasource_id', 'query', 'condition', 'threshold']);
      
      // 模拟测试结果
      const mockResult = {
        success: true,
        matches: 3,
        sample_data: [
          { timestamp: '2024-01-15T10:30:00Z', value: 85.2, labels: { instance: 'server-01' } },
          { timestamp: '2024-01-15T10:31:00Z', value: 87.1, labels: { instance: 'server-02' } },
          { timestamp: '2024-01-15T10:32:00Z', value: 92.3, labels: { instance: 'server-03' } },
        ],
      };
      
      setTestResult(mockResult);
      message.success('规则测试完成');
    } catch (error) {
      message.error('规则测试失败');
      setTestResult({ success: false, error: '测试失败' });
    } finally {
      setTesting(false);
    }
  };

  // 添加标签
  const addLabel = () => {
    setLabelInputs([...labelInputs, { key: '', value: '' }]);
  };

  // 删除标签
  const removeLabel = (index: number) => {
    const newLabels = labelInputs.filter((_, i) => i !== index);
    setLabelInputs(newLabels);
  };

  // 更新标签
  const updateLabel = (index: number, field: 'key' | 'value', value: string) => {
    const newLabels = [...labelInputs];
    newLabels[index][field] = value;
    setLabelInputs(newLabels);
  };

  // 添加注释
  const addAnnotation = () => {
    setAnnotationInputs([...annotationInputs, { key: '', value: '' }]);
  };

  // 删除注释
  const removeAnnotation = (index: number) => {
    const newAnnotations = annotationInputs.filter((_, i) => i !== index);
    setAnnotationInputs(newAnnotations);
  };

  // 更新注释
  const updateAnnotation = (index: number, field: 'key' | 'value', value: string) => {
    const newAnnotations = [...annotationInputs];
    newAnnotations[index][field] = value;
    setAnnotationInputs(newAnnotations);
  };

  return (
    <div className="rule-form">
      <Card
        title={
          <Space>
            <Button
              icon={<ArrowLeftOutlined />}
              onClick={() => navigate('/rules/list')}
            >
              返回
            </Button>
            <Title level={4} style={{ margin: 0 }}>
              {isEdit ? '编辑规则' : '创建规则'}
            </Title>
          </Space>
        }
        extra={
          <Space>
            <Button
              icon={<PlayCircleOutlined />}
              onClick={handleTestRule}
              loading={testing}
            >
              测试规则
            </Button>
            <Button
              type="primary"
              icon={<SaveOutlined />}
              loading={saving}
              onClick={() => form.submit()}
            >
              {isEdit ? '更新规则' : '创建规则'}
            </Button>
          </Space>
        }
        loading={loading}
      >
        {error && (
          <Alert
            message="错误"
            description={error}
            type="error"
            closable
            onClose={clearError}
            style={{ marginBottom: 16 }}
          />
        )}

        <Form
          form={form}
          layout="vertical"
          onFinish={handleSave}
          initialValues={{
            enabled: true,
            severity: 'warning',
            interval: 60,
            timeout: 30,
          }}
        >
          {/* 基本信息 */}
          <Title level={5}>基本信息</Title>
          <Row gutter={24}>
            <Col span={12}>
              <Form.Item
                name="name"
                label="规则名称"
                rules={[{ required: true, message: '请输入规则名称' }]}
              >
                <Input placeholder="请输入规则名称" />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                name="severity"
                label="严重级别"
                rules={[{ required: true, message: '请选择严重级别' }]}
              >
                <Select placeholder="请选择严重级别">
                  <Option value="info">
                    <Tag color="blue">信息</Tag>
                  </Option>
                  <Option value="warning">
                    <Tag color="orange">警告</Tag>
                  </Option>
                  <Option value="critical">
                    <Tag color="red">严重</Tag>
                  </Option>
                </Select>
              </Form.Item>
            </Col>
          </Row>

          <Form.Item
            name="description"
            label="规则描述"
          >
            <TextArea
              placeholder="请输入规则描述"
              rows={3}
            />
          </Form.Item>

          <Divider />

          {/* 数据源和查询 */}
          <Title level={5}>数据源和查询</Title>
          <Row gutter={24}>
            <Col span={12}>
              <Form.Item
                name="datasource_id"
                label="数据源"
                rules={[{ required: true, message: '请选择数据源' }]}
              >
                <Select placeholder="请选择数据源">
                  {dataSources.map(ds => (
                    <Option key={ds.id} value={ds.id}>
                      <Space>
                        <Tag color={ds.status === 'active' ? 'green' : 'red'}>
                          {ds.type}
                        </Tag>
                        {ds.name}
                      </Space>
                    </Option>
                  ))}
                </Select>
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                name="interval"
                label="检查间隔（秒）"
                rules={[{ required: true, message: '请输入检查间隔' }]}
              >
                <InputNumber
                  min={10}
                  max={3600}
                  placeholder="60"
                  style={{ width: '100%' }}
                />
              </Form.Item>
            </Col>
          </Row>

          <Form.Item
            name="query"
            label="查询语句"
            rules={[{ required: true, message: '请输入查询语句' }]}
            extra="支持 PromQL 或其他数据源查询语法"
          >
            <TextArea
              placeholder="例如: up == 0 或 cpu_usage > 80"
              rows={4}
            />
          </Form.Item>

          <Divider />

          {/* 告警条件 */}
          <Title level={5}>告警条件</Title>
          <Row gutter={24}>
            <Col span={12}>
              <Form.Item
                name="condition"
                label="条件表达式"
                rules={[{ required: true, message: '请输入条件表达式' }]}
              >
                <Select placeholder="请选择条件">
                  <Option value=">">&gt; 大于</Option>
                  <Option value=">=">&gt;= 大于等于</Option>
                  <Option value="<">&lt; 小于</Option>
                  <Option value="<=">&lt;= 小于等于</Option>
                  <Option value="==">== 等于</Option>
                  <Option value="!=">!= 不等于</Option>
                </Select>
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                name="threshold"
                label="阈值"
                rules={[{ required: true, message: '请输入阈值' }]}
              >
                <InputNumber
                  placeholder="请输入阈值"
                  style={{ width: '100%' }}
                />
              </Form.Item>
            </Col>
          </Row>

          <Row gutter={24}>
            <Col span={12}>
              <Form.Item
                name="timeout"
                label="超时时间（秒）"
                rules={[{ required: true, message: '请输入超时时间' }]}
              >
                <InputNumber
                  min={5}
                  max={300}
                  placeholder="30"
                  style={{ width: '100%' }}
                />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                name="enabled"
                label="启用规则"
                valuePropName="checked"
              >
                <Switch />
              </Form.Item>
            </Col>
          </Row>

          <Divider />

          {/* 高级配置 */}
          <Collapse ghost>
            <Panel header="高级配置" key="advanced">
              {/* 标签 */}
              <Title level={5}>标签</Title>
              <div style={{ marginBottom: 16 }}>
                {labelInputs.map((label, index) => (
                  <Row key={index} gutter={8} style={{ marginBottom: 8 }}>
                    <Col span={10}>
                      <Input
                        placeholder="标签名"
                        value={label.key}
                        onChange={(e) => updateLabel(index, 'key', e.target.value)}
                      />
                    </Col>
                    <Col span={10}>
                      <Input
                        placeholder="标签值"
                        value={label.value}
                        onChange={(e) => updateLabel(index, 'value', e.target.value)}
                      />
                    </Col>
                    <Col span={4}>
                      <Button
                        type="link"
                        danger
                        onClick={() => removeLabel(index)}
                      >
                        删除
                      </Button>
                    </Col>
                  </Row>
                ))}
                <Button type="dashed" onClick={addLabel}>
                  添加标签
                </Button>
              </div>

              {/* 注释 */}
              <Title level={5}>注释</Title>
              <div style={{ marginBottom: 16 }}>
                {annotationInputs.map((annotation, index) => (
                  <Row key={index} gutter={8} style={{ marginBottom: 8 }}>
                    <Col span={10}>
                      <Input
                        placeholder="注释名"
                        value={annotation.key}
                        onChange={(e) => updateAnnotation(index, 'key', e.target.value)}
                      />
                    </Col>
                    <Col span={10}>
                      <Input
                        placeholder="注释值"
                        value={annotation.value}
                        onChange={(e) => updateAnnotation(index, 'value', e.target.value)}
                      />
                    </Col>
                    <Col span={4}>
                      <Button
                        type="link"
                        danger
                        onClick={() => removeAnnotation(index)}
                      >
                        删除
                      </Button>
                    </Col>
                  </Row>
                ))}
                <Button type="dashed" onClick={addAnnotation}>
                  添加注释
                </Button>
              </div>

              {/* 通知渠道 */}
              <Title level={5}>通知渠道</Title>
              <Form.Item
                name="notification_channels"
                label="通知渠道"
              >
                <Select
                  mode="multiple"
                  placeholder="请选择通知渠道"
                  style={{ width: '100%' }}
                >
                  <Option value="email">邮件</Option>
                  <Option value="sms">短信</Option>
                  <Option value="webhook">Webhook</Option>
                  <Option value="slack">Slack</Option>
                </Select>
              </Form.Item>
            </Panel>
          </Collapse>
        </Form>

        {/* 测试结果 */}
        {testResult && (
          <Card
            title="测试结果"
            style={{ marginTop: 16 }}
            size="small"
          >
            {testResult.success ? (
              <div>
                <Alert
                  message={`测试成功，找到 ${testResult.matches} 条匹配记录`}
                  type="success"
                  style={{ marginBottom: 16 }}
                />
                {testResult.sample_data && (
                  <div>
                    <Text strong>样本数据：</Text>
                    <pre style={{ background: '#f5f5f5', padding: 8, marginTop: 8 }}>
                      {JSON.stringify(testResult.sample_data, null, 2)}
                    </pre>
                  </div>
                )}
              </div>
            ) : (
              <Alert
                message="测试失败"
                description={testResult.error}
                type="error"
              />
            )}
          </Card>
        )}
      </Card>
    </div>
  );
};

export default RuleForm;