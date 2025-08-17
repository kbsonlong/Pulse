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
  Divider,
} from 'antd';
import { useNavigate, useParams } from 'react-router-dom';
import { useTicket, useUI } from '../../hooks';
import { TicketPriority, User } from '../../types';
import { ticketService } from '../../services/ticket';

const { TextArea } = Input;
const { Option } = Select;

interface TicketFormData {
  title: string;
  description: string;
  priority: TicketPriority;
  assignee_id?: string;
  alert_id?: string;
}

const TicketForm: React.FC = () => {
  const navigate = useNavigate();
  const { id } = useParams<{ id: string }>();
  const { setBreadcrumbs } = useUI();
  const { currentTicket, loading, fetchTicket, createTicket, updateTicket } = useTicket();
  
  const [form] = Form.useForm<TicketFormData>();
  const [submitting, setSubmitting] = useState(false);
  const [assignableUsers, setAssignableUsers] = useState<User[]>([]);
  const [loadingUsers, setLoadingUsers] = useState(false);

  const isEdit = !!id;

  useEffect(() => {
    setBreadcrumbs([
      { title: '工单管理' },
      { title: '工单列表', path: '/tickets/list' },
      { title: isEdit ? '编辑工单' : '创建工单' },
    ]);

    if (isEdit && id) {
      fetchTicket(id);
    }

    // 获取可分配的用户列表
    fetchAssignableUsers();
  }, [setBreadcrumbs, isEdit, id, fetchTicket]);

  useEffect(() => {
    if (isEdit && currentTicket) {
      form.setFieldsValue({
        title: currentTicket.title,
        description: currentTicket.description,
        priority: currentTicket.priority,
        assignee_id: currentTicket.assignee?.id,
        alert_id: currentTicket.alert_id,
      });
    }
  }, [isEdit, currentTicket, form]);

  // 获取可分配的用户列表
  const fetchAssignableUsers = async () => {
    setLoadingUsers(true);
    try {
      const users = await ticketService.getAssignableUsers();
      setAssignableUsers(users);
    } catch (error) {
      message.error('获取用户列表失败');
    } finally {
      setLoadingUsers(false);
    }
  };

  // 提交表单
  const handleSubmit = async (values: TicketFormData) => {
    setSubmitting(true);
    try {
      if (isEdit && id) {
        await updateTicket(id, values);
        message.success('工单更新成功');
      } else {
        await createTicket(values);
        message.success('工单创建成功');
      }
      navigate('/tickets/list');
    } catch (error) {
      message.error(isEdit ? '工单更新失败' : '工单创建失败');
    } finally {
      setSubmitting(false);
    }
  };

  // 取消操作
  const handleCancel = () => {
    navigate('/tickets/list');
  };

  // 获取优先级选项
  const priorityOptions = [
    { value: 'low', label: '低', color: 'blue' },
    { value: 'medium', label: '中', color: 'yellow' },
    { value: 'high', label: '高', color: 'orange' },
    { value: 'critical', label: '紧急', color: 'red' },
  ];

  return (
    <div style={{ padding: '24px' }}>
      <Card title={isEdit ? '编辑工单' : '创建工单'}>
        <Form
          form={form}
          layout="vertical"
          onFinish={handleSubmit}
          initialValues={{
            priority: 'medium',
          }}
        >
          <Row gutter={24}>
            <Col span={24}>
              <Form.Item
                label="工单标题"
                name="title"
                rules={[
                  { required: true, message: '请输入工单标题' },
                  { max: 200, message: '标题不能超过200个字符' },
                ]}
              >
                <Input
                  placeholder="请输入工单标题"
                  maxLength={200}
                  showCount
                />
              </Form.Item>
            </Col>
          </Row>

          <Row gutter={24}>
            <Col span={12}>
              <Form.Item
                label="优先级"
                name="priority"
                rules={[{ required: true, message: '请选择优先级' }]}
              >
                <Select placeholder="请选择优先级">
                  {priorityOptions.map((option) => (
                    <Option key={option.value} value={option.value}>
                      <span style={{ color: option.color }}>●</span>
                      <span style={{ marginLeft: 8 }}>{option.label}</span>
                    </Option>
                  ))}
                </Select>
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                label="分配给"
                name="assignee_id"
              >
                <Select
                  placeholder="请选择分配人员"
                  allowClear
                  loading={loadingUsers}
                  showSearch
                  filterOption={(input, option) =>
                    (option?.children as string)
                      ?.toLowerCase()
                      .includes(input.toLowerCase())
                  }
                >
                  {assignableUsers.map((user) => (
                    <Option key={user.id} value={user.id}>
                      {user.username} ({user.email})
                    </Option>
                  ))}
                </Select>
              </Form.Item>
            </Col>
          </Row>

          <Row gutter={24}>
            <Col span={24}>
              <Form.Item
                label="关联告警ID"
                name="alert_id"
              >
                <Input
                  placeholder="请输入关联的告警ID（可选）"
                  disabled={isEdit} // 编辑时不允许修改关联告警
                />
              </Form.Item>
            </Col>
          </Row>

          <Row gutter={24}>
            <Col span={24}>
              <Form.Item
                label="工单描述"
                name="description"
                rules={[
                  { required: true, message: '请输入工单描述' },
                  { max: 2000, message: '描述不能超过2000个字符' },
                ]}
              >
                <TextArea
                  rows={8}
                  placeholder="请详细描述问题或需求..."
                  maxLength={2000}
                  showCount
                />
              </Form.Item>
            </Col>
          </Row>

          <Divider />

          <Form.Item>
            <Space>
              <Button
                type="primary"
                htmlType="submit"
                loading={submitting || loading}
                size="large"
              >
                {isEdit ? '更新工单' : '创建工单'}
              </Button>
              <Button
                size="large"
                onClick={handleCancel}
              >
                取消
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Card>
    </div>
  );
};

export default TicketForm;