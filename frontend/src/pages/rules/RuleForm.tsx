import React, { useEffect, useState } from 'react';
import {
  Form,
  Input,
  Select,
  Button,
  Card,
  Space,
  message,
  Row,
  Col,
  Switch,
  Divider,
  Typography,
  Alert,
} from 'antd';
import {
  PlusOutlined,
  DeleteOutlined,
  SaveOutlined,
  ArrowLeftOutlined,
} from '@ant-design/icons';
import { useNavigate, useParams } from 'react-router-dom';
import { useRule, useUI } from '../../hooks';
import { Rule, RuleCondition, RuleAction } from '../../types';

const { TextArea } = Input;
const { Option } = Select;
const { Title } = Typography;

interface RuleFormData {
  name: string;
  description: string;
  data_source_id: string;
  query: string;
  conditions: RuleCondition[];
  actions: RuleAction[];
  enabled: boolean;
}

const RuleForm: React.FC = () => {
  const navigate = useNavigate();
  const { id } = useParams<{ id: string }>();
  const { setBreadcrumbs } = useUI();
  const {
    currentRule,
    dataSources,
    loading,
    fetchRule,
    createRule,
    updateRule,
    fetchDataSources,
    clearCurrentRule,
  } = useRule();

  const [form] = Form.useForm<RuleFormData>();
  const [submitting, setSubmitting] = useState(false);
  const isEdit = !!id;

  useEffect(() => {
    setBreadcrumbs([
      { title: '规则管理' },
      { title: '规则列表', path: '/rules' },
      { title: isEdit ? '编辑规则' : '创建规则' },
    ]);

    fetchDataSources();

    if (isEdit && id) {
      fetchRule(id);
    }

    return () => {
      clearCurrentRule();
    };
  }, [setBreadcrumbs, isEdit, id, fetchRule, fetchDataSources, clearCurrentRule]);

  useEffect(() => {
    if (currentRule && isEdit) {
      form.setFieldsValue({
        name: currentRule.name,
        description: currentRule.description,
        data_source_id: currentRule.data_source_id,
        query: currentRule.query,
        conditions: currentRule.conditions,
        actions: currentRule.actions,
        enabled: currentRule.enabled,
      });
    }
  }, [currentRule, isEdit, form]);

  const handleSubmit = async (values: RuleFormData) => {
    setSubmitting(true);
    try {
      if (isEdit && id) {
        await updateRule(id, values);
        message.success('规则更新成功');
      } else {
        await createRule(values);
        message.success('规则创建成功');
      }
      navigate('/rules');
    } catch (error) {
      message.error(isEdit ? '规则更新失败' : '规则创建失败');
    } finally {
      setSubmitting(false);
    }
  };

  const handleBack = () => {
    navigate('/rules');
  };

  const conditionOperators = [
    { label: '等于', value: 'eq' },
    { label: '不等于', value: 'ne' },
    { label: '大于', value: 'gt' },
    { label: '大于等于', value: 'gte' },
    { label: '小于', value: 'lt' },
    { label: '小于等于', value: 'lte' },
    { label: '包含', value: 'contains' },
    { label: '不包含', value: 'not_contains' },
    { label: '正则匹配', value: 'regex' },
  ];

  const actionTypes = [
    { label: '发送邮件', value: 'email' },
    { label: '发送短信', value: 'sms' },
    { label: 'Webhook', value: 'webhook' },
    { label: '钉钉通知', value: 'dingtalk' },
    { label: '企业微信', value: 'wechat' },
  ];

  return (
    <div style={{ padding: '24px' }}>
      <Card>
        <div style={{ marginBottom: '24px' }}>
          <Space>
            <Button icon={<ArrowLeftOutlined />} onClick={handleBack}>
              返回
            </Button>
            <Title level={4} style={{ margin: 0 }}>
              {isEdit ? '编辑规则' : '创建规则'}
            </Title>
          </Space>
        </div>

        <Form
          form={form}
          layout="vertical"
          onFinish={handleSubmit}
          initialValues={{
            enabled: true,
            conditions: [{ field: '', operator: 'eq', value: '' }],
            actions: [{ type: 'email', config: {} }],
          }}
        >
          <Row gutter={24}>
            <Col span={12}>
              <Form.Item
                label="规则名称"
                name="name"
                rules={[
                  { required: true, message: '请输入规则名称' },
                  { max: 100, message: '规则名称不能超过100个字符' },
                ]}
              >
                <Input placeholder="请输入规则名称" />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                label="数据源"
                name="data_source_id"
                rules={[{ required: true, message: '请选择数据源' }]}
              >
                <Select placeholder="请选择数据源">
                  {dataSources.map((ds) => (
                    <Option key={ds.id} value={ds.id}>
                      {ds.name} ({ds.type})
                    </Option>
                  ))}
                </Select>
              </Form.Item>
            </Col>
          </Row>

          <Form.Item
            label="规则描述"
            name="description"
            rules={[{ max: 500, message: '描述不能超过500个字符' }]}
          >
            <TextArea
              rows={3}
              placeholder="请输入规则描述"
            />
          </Form.Item>

          <Form.Item
            label="查询语句"
            name="query"
            rules={[{ required: true, message: '请输入查询语句' }]}
          >
            <TextArea
              rows={4}
              placeholder="请输入查询语句，例如：SELECT * FROM metrics WHERE cpu_usage > 80"
            />
          </Form.Item>

          <Divider>触发条件</Divider>
          <Form.List name="conditions">
            {(fields, { add, remove }) => (
              <>
                {fields.map(({ key, name, ...restField }) => (
                  <Card key={key} size="small" style={{ marginBottom: '16px' }}>
                    <Row gutter={16}>
                      <Col span={6}>
                        <Form.Item
                          {...restField}
                          name={[name, 'field']}
                          label="字段"
                          rules={[{ required: true, message: '请输入字段名' }]}
                        >
                          <Input placeholder="字段名" />
                        </Form.Item>
                      </Col>
                      <Col span={6}>
                        <Form.Item
                          {...restField}
                          name={[name, 'operator']}
                          label="操作符"
                          rules={[{ required: true, message: '请选择操作符' }]}
                        >
                          <Select placeholder="选择操作符">
                            {conditionOperators.map((op) => (
                              <Option key={op.value} value={op.value}>
                                {op.label}
                              </Option>
                            ))}
                          </Select>
                        </Form.Item>
                      </Col>
                      <Col span={8}>
                        <Form.Item
                          {...restField}
                          name={[name, 'value']}
                          label="值"
                          rules={[{ required: true, message: '请输入值' }]}
                        >
                          <Input placeholder="值" />
                        </Form.Item>
                      </Col>
                      <Col span={4}>
                        <Form.Item label=" ">
                          <Button
                            type="text"
                            danger
                            icon={<DeleteOutlined />}
                            onClick={() => remove(name)}
                            disabled={fields.length === 1}
                          >
                            删除
                          </Button>
                        </Form.Item>
                      </Col>
                    </Row>
                  </Card>
                ))}
                <Button
                  type="dashed"
                  onClick={() => add()}
                  block
                  icon={<PlusOutlined />}
                >
                  添加条件
                </Button>
              </>
            )}
          </Form.List>

          <Divider>执行动作</Divider>
          <Form.List name="actions">
            {(fields, { add, remove }) => (
              <>
                {fields.map(({ key, name, ...restField }) => (
                  <Card key={key} size="small" style={{ marginBottom: '16px' }}>
                    <Row gutter={16}>
                      <Col span={8}>
                        <Form.Item
                          {...restField}
                          name={[name, 'type']}
                          label="动作类型"
                          rules={[{ required: true, message: '请选择动作类型' }]}
                        >
                          <Select placeholder="选择动作类型">
                            {actionTypes.map((action) => (
                              <Option key={action.value} value={action.value}>
                                {action.label}
                              </Option>
                            ))}
                          </Select>
                        </Form.Item>
                      </Col>
                      <Col span={12}>
                        <Form.Item
                          {...restField}
                          name={[name, 'config']}
                          label="配置"
                          rules={[{ required: true, message: '请输入配置' }]}
                        >
                          <TextArea
                            rows={2}
                            placeholder='请输入JSON格式的配置，例如：{"url": "https://example.com/webhook"}'
                          />
                        </Form.Item>
                      </Col>
                      <Col span={4}>
                        <Form.Item label=" ">
                          <Button
                            type="text"
                            danger
                            icon={<DeleteOutlined />}
                            onClick={() => remove(name)}
                            disabled={fields.length === 1}
                          >
                            删除
                          </Button>
                        </Form.Item>
                      </Col>
                    </Row>
                  </Card>
                ))}
                <Button
                  type="dashed"
                  onClick={() => add()}
                  block
                  icon={<PlusOutlined />}
                >
                  添加动作
                </Button>
              </>
            )}
          </Form.List>

          <Divider />
          <Row gutter={24}>
            <Col span={12}>
              <Form.Item
                label="启用状态"
                name="enabled"
                valuePropName="checked"
              >
                <Switch checkedChildren="启用" unCheckedChildren="禁用" />
              </Form.Item>
            </Col>
          </Row>

          <Form.Item>
            <Space>
              <Button
                type="primary"
                htmlType="submit"
                loading={submitting}
                icon={<SaveOutlined />}
              >
                {isEdit ? '更新规则' : '创建规则'}
              </Button>
              <Button onClick={handleBack}>
                取消
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Card>
    </div>
  );
};

export default RuleForm;