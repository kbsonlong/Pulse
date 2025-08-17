import React, { useState, useEffect } from 'react';
import {
  Card,
  Row,
  Col,
  Button,
  Space,
  Tag,
  Descriptions,
  Timeline,
  Modal,
  Form,
  Input,
  Select,
  message,
  Spin,
  Divider,
  Avatar,
  Typography,
  Upload,
  List,
  Popconfirm,
  Badge
} from 'antd';
import {
  ArrowLeftOutlined,
  EditOutlined,
  DeleteOutlined,
  UserOutlined,
  ClockCircleOutlined,
  FileTextOutlined,
  PaperClipOutlined,
  SendOutlined,
  UploadOutlined
} from '@ant-design/icons';
import { useNavigate, useParams } from 'react-router-dom';
import { format } from 'date-fns';
import { useTicket } from '../../hooks/useTicket';
import { useUI } from '../../hooks/useUI';
import { ticketService } from '../../services/ticket';
import { userService } from '../../services/user';
import type { Ticket, TicketStatus, TicketPriority, ProcessRecord } from '../../types';
import type { UploadFile } from 'antd';

const { TextArea } = Input;
const { Text, Title } = Typography;

interface User {
  id: string;
  username: string;
  email: string;
  avatar?: string;
}

const TicketDetail: React.FC = () => {
  const navigate = useNavigate();
  const { id } = useParams<{ id: string }>();
  const [form] = Form.useForm();
  
  const { loading, setLoading } = useUI();
  const [ticket, setTicket] = useState<Ticket | null>(null);
  const [processRecords, setProcessRecords] = useState<ProcessRecord[]>([]);
  const [users, setUsers] = useState<User[]>([]);
  const [statusModalVisible, setStatusModalVisible] = useState(false);
  const [assignModalVisible, setAssignModalVisible] = useState(false);
  const [commentModalVisible, setCommentModalVisible] = useState(false);
  const [attachments, setAttachments] = useState<UploadFile[]>([]);

  // 加载工单详情
  const loadTicket = async () => {
    if (!id) return;
    
    try {
      setLoading(true);
      const [ticketData, recordsData] = await Promise.all([
        ticketService.getTicket(id),
        ticketService.getProcessRecords(id)
      ]);
      
      setTicket(ticketData);
      setProcessRecords(recordsData);
    } catch (error) {
      message.error('加载工单详情失败');
      navigate('/tickets');
    } finally {
      setLoading(false);
    }
  };

  // 加载用户列表
  const loadUsers = async () => {
    try {
      const response = await userService.getUsers({ page: 1, limit: 100 });
      setUsers(response.users);
    } catch (error) {
      console.error('Failed to load users:', error);
    }
  };

  useEffect(() => {
    loadTicket();
    loadUsers();
  }, [id]);

  // 更新工单状态
  const handleStatusUpdate = async (values: { status: TicketStatus; comment?: string }) => {
    if (!id) return;
    
    try {
      await ticketService.updateTicketStatus(id, values.status);
      
      if (values.comment) {
        await ticketService.addProcessRecord(id, {
          action: 'status_change',
          comment: values.comment,
          attachments: attachments.map(file => file.name)
        });
      }
      
      message.success('状态更新成功');
      setStatusModalVisible(false);
      form.resetFields();
      setAttachments([]);
      loadTicket();
    } catch (error) {
      message.error('状态更新失败');
    }
  };

  // 分配工单
  const handleAssign = async (values: { assigned_to: string; comment?: string }) => {
    if (!id) return;
    
    try {
      await ticketService.assignTicket(id, values.assigned_to);
      
      if (values.comment) {
        await ticketService.addProcessRecord(id, {
          action: 'assign',
          comment: values.comment
        });
      }
      
      message.success('分配成功');
      setAssignModalVisible(false);
      form.resetFields();
      loadTicket();
    } catch (error) {
      message.error('分配失败');
    }
  };

  // 添加处理记录
  const handleAddComment = async (values: { comment: string }) => {
    if (!id) return;
    
    try {
      await ticketService.addProcessRecord(id, {
        action: 'comment',
        comment: values.comment,
        attachments: attachments.map(file => file.name)
      });
      
      message.success('评论添加成功');
      setCommentModalVisible(false);
      form.resetFields();
      setAttachments([]);
      loadTicket();
    } catch (error) {
      message.error('评论添加失败');
    }
  };

  // 删除工单
  const handleDelete = async () => {
    if (!id) return;
    
    try {
      await ticketService.deleteTicket(id);
      message.success('工单删除成功');
      navigate('/tickets');
    } catch (error) {
      message.error('工单删除失败');
    }
  };

  const getStatusColor = (status: TicketStatus) => {
    const colors = {
      open: 'blue',
      in_progress: 'orange',
      resolved: 'green',
      closed: 'gray'
    };
    return colors[status] || 'default';
  };

  const getPriorityColor = (priority: TicketPriority) => {
    const colors = {
      low: 'green',
      medium: 'orange',
      high: 'red',
      urgent: 'purple'
    };
    return colors[priority] || 'default';
  };

  const getStatusText = (status: TicketStatus) => {
    const texts = {
      open: '待处理',
      in_progress: '处理中',
      resolved: '已解决',
      closed: '已关闭'
    };
    return texts[status] || status;
  };

  const getPriorityText = (priority: TicketPriority) => {
    const texts = {
      low: '低',
      medium: '中',
      high: '高',
      urgent: '紧急'
    };
    return texts[priority] || priority;
  };

  const getActionText = (action: string) => {
    const texts = {
      create: '创建',
      status_change: '状态变更',
      assign: '分配',
      comment: '评论',
      update: '更新'
    };
    return texts[action] || action;
  };

  if (loading) {
    return (
      <div style={{ textAlign: 'center', padding: '50px' }}>
        <Spin size="large" />
      </div>
    );
  }

  if (!ticket) {
    return (
      <div style={{ textAlign: 'center', padding: '50px' }}>
        <Text>工单不存在</Text>
      </div>
    );
  }

  return (
    <div style={{ padding: '24px' }}>
      {/* 头部操作栏 */}
      <Card style={{ marginBottom: '24px' }}>
        <Row justify="space-between" align="middle">
          <Col>
            <Space>
              <Button
                icon={<ArrowLeftOutlined />}
                onClick={() => navigate('/tickets')}
              >
                返回
              </Button>
              <Title level={3} style={{ margin: 0 }}>
                工单详情 #{ticket.id}
              </Title>
              <Tag color={getStatusColor(ticket.status)}>
                {getStatusText(ticket.status)}
              </Tag>
              <Tag color={getPriorityColor(ticket.priority)}>
                {getPriorityText(ticket.priority)}
              </Tag>
            </Space>
          </Col>
          <Col>
            <Space>
              <Button
                type="primary"
                onClick={() => setStatusModalVisible(true)}
              >
                更新状态
              </Button>
              <Button
                onClick={() => setAssignModalVisible(true)}
              >
                分配工单
              </Button>
              <Button
                icon={<EditOutlined />}
                onClick={() => navigate(`/tickets/${id}/edit`)}
              >
                编辑
              </Button>
              <Popconfirm
                title="确定要删除这个工单吗？"
                onConfirm={handleDelete}
                okText="确定"
                cancelText="取消"
              >
                <Button
                  danger
                  icon={<DeleteOutlined />}
                >
                  删除
                </Button>
              </Popconfirm>
            </Space>
          </Col>
        </Row>
      </Card>

      <Row gutter={24}>
        {/* 左侧：工单信息 */}
        <Col span={16}>
          <Card title="工单信息" style={{ marginBottom: '24px' }}>
            <Descriptions column={2} bordered>
              <Descriptions.Item label="标题" span={2}>
                {ticket.title}
              </Descriptions.Item>
              <Descriptions.Item label="状态">
                <Tag color={getStatusColor(ticket.status)}>
                  {getStatusText(ticket.status)}
                </Tag>
              </Descriptions.Item>
              <Descriptions.Item label="优先级">
                <Tag color={getPriorityColor(ticket.priority)}>
                  {getPriorityText(ticket.priority)}
                </Tag>
              </Descriptions.Item>
              <Descriptions.Item label="创建人">
                <Space>
                  <Avatar size="small" icon={<UserOutlined />} />
                  {ticket.created_by}
                </Space>
              </Descriptions.Item>
              <Descriptions.Item label="分配给">
                {ticket.assigned_to ? (
                  <Space>
                    <Avatar size="small" icon={<UserOutlined />} />
                    {ticket.assigned_to}
                  </Space>
                ) : (
                  <Text type="secondary">未分配</Text>
                )}
              </Descriptions.Item>
              <Descriptions.Item label="创建时间">
                <Space>
                  <ClockCircleOutlined />
                  {format(new Date(ticket.created_at), 'yyyy-MM-dd HH:mm:ss')}
                </Space>
              </Descriptions.Item>
              <Descriptions.Item label="更新时间">
                <Space>
                  <ClockCircleOutlined />
                  {format(new Date(ticket.updated_at), 'yyyy-MM-dd HH:mm:ss')}
                </Space>
              </Descriptions.Item>
              {ticket.due_date && (
                <Descriptions.Item label="截止时间" span={2}>
                  <Space>
                    <ClockCircleOutlined />
                    {format(new Date(ticket.due_date), 'yyyy-MM-dd HH:mm:ss')}
                  </Space>
                </Descriptions.Item>
              )}
              {ticket.tags && ticket.tags.length > 0 && (
                <Descriptions.Item label="标签" span={2}>
                  <Space wrap>
                    {ticket.tags.map(tag => (
                      <Tag key={tag}>{tag}</Tag>
                    ))}
                  </Space>
                </Descriptions.Item>
              )}
              <Descriptions.Item label="描述" span={2}>
                <div style={{ whiteSpace: 'pre-wrap' }}>
                  {ticket.description}
                </div>
              </Descriptions.Item>
            </Descriptions>
          </Card>

          {/* 处理记录 */}
          <Card
            title="处理记录"
            extra={
              <Button
                type="primary"
                icon={<FileTextOutlined />}
                onClick={() => setCommentModalVisible(true)}
              >
                添加评论
              </Button>
            }
          >
            <Timeline>
              {processRecords.map(record => (
                <Timeline.Item key={record.id}>
                  <div>
                    <Space>
                      <Text strong>{getActionText(record.action)}</Text>
                      <Text type="secondary">
                        {format(new Date(record.created_at), 'yyyy-MM-dd HH:mm:ss')}
                      </Text>
                      <Text type="secondary">by {record.created_by}</Text>
                    </Space>
                    {record.comment && (
                      <div style={{ marginTop: '8px', whiteSpace: 'pre-wrap' }}>
                        {record.comment}
                      </div>
                    )}
                    {record.attachments && record.attachments.length > 0 && (
                      <div style={{ marginTop: '8px' }}>
                        <Space wrap>
                          {record.attachments.map(attachment => (
                            <Tag key={attachment} icon={<PaperClipOutlined />}>
                              {attachment}
                            </Tag>
                          ))}
                        </Space>
                      </div>
                    )}
                  </div>
                </Timeline.Item>
              ))}
            </Timeline>
          </Card>
        </Col>

        {/* 右侧：相关信息 */}
        <Col span={8}>
          {ticket.alert_id && (
            <Card title="关联告警" style={{ marginBottom: '24px' }}>
              <Space direction="vertical" style={{ width: '100%' }}>
                <Text>告警ID: {ticket.alert_id}</Text>
                <Button
                  type="link"
                  onClick={() => navigate(`/alerts/${ticket.alert_id}`)}
                >
                  查看告警详情
                </Button>
              </Space>
            </Card>
          )}

          <Card title="统计信息">
            <Space direction="vertical" style={{ width: '100%' }}>
              <div>
                <Text type="secondary">处理记录数量：</Text>
                <Text strong>{processRecords.length}</Text>
              </div>
              <div>
                <Text type="secondary">创建时长：</Text>
                <Text strong>
                  {Math.floor((Date.now() - new Date(ticket.created_at).getTime()) / (1000 * 60 * 60 * 24))} 天
                </Text>
              </div>
              {ticket.resolved_at && (
                <div>
                  <Text type="secondary">解决时长：</Text>
                  <Text strong>
                    {Math.floor((new Date(ticket.resolved_at).getTime() - new Date(ticket.created_at).getTime()) / (1000 * 60 * 60))} 小时
                  </Text>
                </div>
              )}
            </Space>
          </Card>
        </Col>
      </Row>

      {/* 状态更新模态框 */}
      <Modal
        title="更新工单状态"
        open={statusModalVisible}
        onCancel={() => setStatusModalVisible(false)}
        footer={null}
      >
        <Form form={form} onFinish={handleStatusUpdate} layout="vertical">
          <Form.Item
            name="status"
            label="新状态"
            rules={[{ required: true, message: '请选择状态' }]}
          >
            <Select placeholder="请选择状态">
              <Select.Option value="open">待处理</Select.Option>
              <Select.Option value="in_progress">处理中</Select.Option>
              <Select.Option value="resolved">已解决</Select.Option>
              <Select.Option value="closed">已关闭</Select.Option>
            </Select>
          </Form.Item>
          <Form.Item name="comment" label="备注">
            <TextArea rows={4} placeholder="请输入状态变更说明..." />
          </Form.Item>
          <Form.Item label="附件">
            <Upload
              fileList={attachments}
              onChange={({ fileList }) => setAttachments(fileList)}
              beforeUpload={() => false}
              multiple
            >
              <Button icon={<UploadOutlined />}>选择文件</Button>
            </Upload>
          </Form.Item>
          <Form.Item>
            <Space>
              <Button type="primary" htmlType="submit" icon={<SendOutlined />}>
                更新状态
              </Button>
              <Button onClick={() => setStatusModalVisible(false)}>
                取消
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>

      {/* 分配工单模态框 */}
      <Modal
        title="分配工单"
        open={assignModalVisible}
        onCancel={() => setAssignModalVisible(false)}
        footer={null}
      >
        <Form form={form} onFinish={handleAssign} layout="vertical">
          <Form.Item
            name="assigned_to"
            label="分配给"
            rules={[{ required: true, message: '请选择处理人' }]}
          >
            <Select
              placeholder="请选择处理人"
              showSearch
              filterOption={(input, option) =>
                (option?.children as string)?.toLowerCase().includes(input.toLowerCase())
              }
            >
              {users.map(user => (
                <Select.Option key={user.id} value={user.id}>
                  {user.username} ({user.email})
                </Select.Option>
              ))}
            </Select>
          </Form.Item>
          <Form.Item name="comment" label="备注">
            <TextArea rows={4} placeholder="请输入分配说明..." />
          </Form.Item>
          <Form.Item>
            <Space>
              <Button type="primary" htmlType="submit" icon={<SendOutlined />}>
                分配工单
              </Button>
              <Button onClick={() => setAssignModalVisible(false)}>
                取消
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>

      {/* 添加评论模态框 */}
      <Modal
        title="添加评论"
        open={commentModalVisible}
        onCancel={() => setCommentModalVisible(false)}
        footer={null}
      >
        <Form form={form} onFinish={handleAddComment} layout="vertical">
          <Form.Item
            name="comment"
            label="评论内容"
            rules={[{ required: true, message: '请输入评论内容' }]}
          >
            <TextArea rows={6} placeholder="请输入评论内容..." />
          </Form.Item>
          <Form.Item label="附件">
            <Upload
              fileList={attachments}
              onChange={({ fileList }) => setAttachments(fileList)}
              beforeUpload={() => false}
              multiple
            >
              <Button icon={<UploadOutlined />}>选择文件</Button>
            </Upload>
          </Form.Item>
          <Form.Item>
            <Space>
              <Button type="primary" htmlType="submit" icon={<SendOutlined />}>
                添加评论
              </Button>
              <Button onClick={() => setCommentModalVisible(false)}>
                取消
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
};

export default TicketDetail;