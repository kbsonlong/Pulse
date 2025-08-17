import React, { useState, useEffect } from 'react';
import {
  Form,
  Input,
  Select,
  Button,
  Card,
  Row,
  Col,
  Space,
  DatePicker,
  Upload,
  message,
  Spin,
  Tag,
  Divider,
  Alert
} from 'antd';
import {
  SaveOutlined,
  ArrowLeftOutlined,
  PlusOutlined,
  DeleteOutlined,
  UploadOutlined
} from '@ant-design/icons';
import { useNavigate, useParams } from 'react-router-dom';
import { useTicket } from '../../hooks/useTicket';
import { useUI } from '../../hooks/useUI';
import { ticketService } from '../../services/ticket';
import { userService } from '../../services/user';
import type { Ticket, TicketPriority, TicketStatus } from '../../types';
import type { UploadFile } from 'antd';

const { TextArea } = Input;
const { Option } = Select;

interface TicketFormData {
  title: string;
  description: string;
  priority: TicketPriority;
  assigned_to?: string;
  due_date?: string;
  tags: string[];
  attachments?: UploadFile[];
  alert_id?: string;
}

interface User {
  id: string;
  username: string;
  email: string;
  role: string;
}

const TicketForm: React.FC = () => {
  const [form] = Form.useForm<TicketFormData>();
  const navigate = useNavigate();
  const { id } = useParams<{ id: string }>();
  const isEdit = Boolean(id);
  
  const { loading, setLoading } = useUI();
  const [users, setUsers] = useState<User[]>([]);
  const [ticket, setTicket] = useState<Ticket | null>(null);
  const [attachments, setAttachments] = useState<UploadFile[]>([]);
  const [tags, setTags] = useState<string[]>([]);
  const [newTag, setNewTag] = useState('');

  // 加载用户列表
  const loadUsers = async () => {
    try {
      const response = await userService.getUsers({ page: 1, limit: 100 });
      setUsers(response.users);
    } catch (error) {
      console.error('Failed to load users:', error);
    }
  };

  // 加载工单详情（编辑模式）
  const loadTicket = async () => {
    if (!id) return;
    
    try {
      setLoading(true);
      const ticketData = await ticketService.getTicket(id);
      setTicket(ticketData);
      
      // 填充表单
      form.setFieldsValue({
        title: ticketData.title,
        description: ticketData.description,
        priority: ticketData.priority,
        assigned_to: ticketData.assigned_to,
        due_date: ticketData.due_date,
        tags: ticketData.tags || []
      });
      
      setTags(ticketData.tags || []);
    } catch (error) {
      message.error('加载工单详情失败');
      navigate('/tickets');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadUsers();
    if (isEdit) {
      loadTicket();
    }
  }, [id, isEdit]);

  // 处理表单提交
  const handleSubmit = async (values: TicketFormData) => {
    try {
      setLoading(true);
      
      const formData = {
        ...values,
        tags,
        due_date: values.due_date || undefined
      };

      if (isEdit && id) {
        await ticketService.updateTicket(id, formData);
        message.success('工单更新成功');
      } else {
        await ticketService.createTicket(formData);
        message.success('工单创建成功');
      }
      
      navigate('/tickets');
    } catch (error) {
      message.error(isEdit ? '工单更新失败' : '工单创建失败');
    } finally {
      setLoading(false);
    }
  };

  // 添加标签
  const handleAddTag = () => {
    if (newTag && !tags.includes(newTag)) {
      const newTags = [...tags, newTag];
      setTags(newTags);
      form.setFieldValue('tags', newTags);
      setNewTag('');
    }
  };

  // 删除标签
  const handleRemoveTag = (tagToRemove: string) => {
    const newTags = tags.filter(tag => tag !== tagToRemove);
    setTags(newTags);
    form.setFieldValue('tags', newTags);
  };

  // 文件上传处理
  const handleUploadChange = (info: any) => {
    setAttachments(info.fileList);
  };

  const priorityOptions = [
    { value: 'low', label: '低', color: 'green' },
    { value: 'medium', label: '中', color: 'orange' },
    { value: 'high', label: '高', color: 'red' },
    { value: 'urgent', label: '紧急', color: 'purple' }
  ];

  if (loading && isEdit) {
    return (
      <div style={{ textAlign: 'center', padding: '50px' }}>
        <Spin size="large" />
      </div>
    );
  }

  return (
    <div style={{ padding: '24px' }}>
      <Card>
        <div style={{ marginBottom: '24px' }}>
          <Space>
            <Button
              icon={<ArrowLeftOutlined />}
              onClick={() => navigate('/tickets')}
            >
              返回
            </Button>
            <h2 style={{ margin: 0 }}>
              {isEdit ? '编辑工单' : '创建工单'}
            </h2>
          </Space>
        </div>

        {ticket?.alert_id && (
          <Alert
            message="此工单由告警自动创建"
            description={`关联告警ID: ${ticket.alert_id}`}
            type="info"
            showIcon
            style={{ marginBottom: '24px' }}
          />
        )}

        <Form
          form={form}
          layout="vertical"
          onFinish={handleSubmit}
          initialValues={{
            priority: 'medium',
            tags: []
          }}
        >
          <Row gutter={24}>
            <Col span={24}>
              <Form.Item
                name="title"
                label="工单标题"
                rules={[
                  { required: true, message: '请输入工单标题' },
                  { max: 200, message: '标题不能超过200个字符' }
                ]}
              >
                <Input placeholder="请输入工单标题" />
              </Form.Item>
            </Col>
          </Row>

          <Row gutter={24}>
            <Col span={12}>
              <Form.Item
                name="priority"
                label="优先级"
                rules={[{ required: true, message: '请选择优先级' }]}
              >
                <Select placeholder="请选择优先级">
                  {priorityOptions.map(option => (
                    <Option key={option.value} value={option.value}>
                      <Tag color={option.color}>{option.label}</Tag>
                    </Option>
                  ))}
                </Select>
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                name="assigned_to"
                label="分配给"
              >
                <Select
                  placeholder="请选择处理人"
                  allowClear
                  showSearch
                  filterOption={(input, option) =>
                    (option?.children as string)?.toLowerCase().includes(input.toLowerCase())
                  }
                >
                  {users.map(user => (
                    <Option key={user.id} value={user.id}>
                      {user.username} ({user.email})
                    </Option>
                  ))}
                </Select>
              </Form.Item>
            </Col>
          </Row>

          <Row gutter={24}>
            <Col span={12}>
              <Form.Item
                name="due_date"
                label="截止日期"
              >
                <DatePicker
                  style={{ width: '100%' }}
                  placeholder="请选择截止日期"
                  showTime
                />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item label="标签">
                <Space direction="vertical" style={{ width: '100%' }}>
                  <Space wrap>
                    {tags.map(tag => (
                      <Tag
                        key={tag}
                        closable
                        onClose={() => handleRemoveTag(tag)}
                      >
                        {tag}
                      </Tag>
                    ))}
                  </Space>
                  <Space.Compact style={{ width: '100%' }}>
                    <Input
                      placeholder="添加标签"
                      value={newTag}
                      onChange={(e) => setNewTag(e.target.value)}
                      onPressEnter={handleAddTag}
                    />
                    <Button
                      type="primary"
                      icon={<PlusOutlined />}
                      onClick={handleAddTag}
                      disabled={!newTag}
                    >
                      添加
                    </Button>
                  </Space.Compact>
                </Space>
              </Form.Item>
            </Col>
          </Row>

          <Row gutter={24}>
            <Col span={24}>
              <Form.Item
                name="description"
                label="工单描述"
                rules={[
                  { required: true, message: '请输入工单描述' },
                  { min: 10, message: '描述至少需要10个字符' }
                ]}
              >
                <TextArea
                  rows={6}
                  placeholder="请详细描述问题或需求..."
                  showCount
                  maxLength={2000}
                />
              </Form.Item>
            </Col>
          </Row>

          <Row gutter={24}>
            <Col span={24}>
              <Form.Item label="附件">
                <Upload
                  fileList={attachments}
                  onChange={handleUploadChange}
                  beforeUpload={() => false} // 阻止自动上传
                  multiple
                >
                  <Button icon={<UploadOutlined />}>选择文件</Button>
                </Upload>
              </Form.Item>
            </Col>
          </Row>

          <Divider />

          <Form.Item>
            <Space>
              <Button
                type="primary"
                htmlType="submit"
                icon={<SaveOutlined />}
                loading={loading}
                size="large"
              >
                {isEdit ? '更新工单' : '创建工单'}
              </Button>
              <Button
                size="large"
                onClick={() => navigate('/tickets')}
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