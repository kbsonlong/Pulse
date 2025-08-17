import React, { useEffect, useState } from 'react';
import {
  Card,
  Descriptions,
  Button,
  Space,
  Tag,
  Badge,
  Timeline,
  Modal,
  Form,
  Input,
  Select,
  message,
  Spin,
  Divider,
} from 'antd';
import {
  ArrowLeftOutlined,
  EditOutlined,
  ReloadOutlined,
  MessageOutlined,
} from '@ant-design/icons';
import { useParams, useNavigate } from 'react-router-dom';
import { useTicket, useUI } from '../../hooks';
import { Ticket, TicketStatus, TicketPriority } from '../../types';
import { formatDate } from '../../utils';

const { TextArea } = Input;
const { Option } = Select;

interface CommentForm {
  content: string;
}

interface StatusUpdateForm {
  status: TicketStatus;
  comment: string;
}

const TicketDetail: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { setBreadcrumb } = useUI();
  const {
    currentTicket,
    loading,
    fetchTicketById,
    updateTicketStatus,
    addTicketComment,
  } = useTicket();

  const [commentModalVisible, setCommentModalVisible] = useState(false);
  const [statusModalVisible, setStatusModalVisible] = useState(false);
  const [commentForm] = Form.useForm<CommentForm>();
  const [statusForm] = Form.useForm<StatusUpdateForm>();

  useEffect(() => {
    if (id) {
      fetchTicketById(id);
    }
  }, [id, fetchTicketById]);

  useEffect(() => {
    if (currentTicket) {
      setBreadcrumb([
        { title: '工单管理' },
        { title: '工单列表', path: '/tickets' },
        { title: currentTicket.title },
      ]);
    }
  }, [currentTicket, setBreadcrumb]);

  // 返回列表
  const handleBack = () => {
    navigate('/tickets');
  };

  // 编辑工单
  const handleEdit = () => {
    navigate(`/tickets/${id}/edit`);
  };

  // 刷新数据
  const handleRefresh = () => {
    if (id) {
      fetchTicketById(id);
    }
  };

  // 添加评论
  const handleAddComment = async (values: CommentForm) => {
    try {
      if (id) {
        await addTicketComment(id, values.content);
        message.success('评论添加成功');
        setCommentModalVisible(false);
        commentForm.resetFields();
        fetchTicketById(id);
      }
    } catch (error) {
      message.error('评论添加失败');
    }
  };

  // 更新状态
  const handleUpdateStatus = async (values: StatusUpdateForm) => {
    try {
      if (id) {
        await updateTicketStatus(id, values.status, values.comment);
        message.success('状态更新成功');
        setStatusModalVisible(false);
        statusForm.resetFields();
        fetchTicketById(id);
      }
    } catch (error) {
      message.error('状态更新失败');
    }
  };

  // 获取优先级颜色
  const getPriorityColor = (priority: TicketPriority) => {
    const colors: Record<TicketPriority, string> = {
      critical: 'red',
      high: 'orange',
      medium: 'yellow',
      low: 'blue',
    };
    return colors[priority] || 'default';
  };

  // 获取状态颜色
  const getStatusColor = (status: TicketStatus) => {
    const colors: Record<TicketStatus, string> = {
      open: 'processing',
      in_progress: 'warning',
      resolved: 'success',
      closed: 'default',
    };
    return colors[status] || 'default';
  };

  // 获取状态文本
  const getStatusText = (status: TicketStatus) => {
    const texts: Record<TicketStatus, string> = {
      open: '待处理',
      in_progress: '处理中',
      resolved: '已解决',
      closed: '已关闭',
    };
    return texts[status] || status;
  };

  if (loading || !currentTicket) {
    return (
      <div style={{ textAlign: 'center', padding: '50px' }}>
        <Spin size="large" />
      </div>
    );
  }

  return (
    <div>
      {/* 操作栏 */}
      <Card style={{ marginBottom: 16 }}>
        <Space>
          <Button icon={<ArrowLeftOutlined />} onClick={handleBack}>
            返回列表
          </Button>
          <Button icon={<EditOutlined />} onClick={handleEdit}>
            编辑工单
          </Button>
          <Button icon={<ReloadOutlined />} onClick={handleRefresh}>
            刷新
          </Button>
          <Button
            icon={<MessageOutlined />}
            onClick={() => setCommentModalVisible(true)}
          >
            添加评论
          </Button>
          <Button
            type="primary"
            onClick={() => setStatusModalVisible(true)}
          >
            更新状态
          </Button>
        </Space>
      </Card>

      {/* 工单基本信息 */}
      <Card title="工单信息" style={{ marginBottom: 16 }}>
        <Descriptions column={2} bordered>
          <Descriptions.Item label="工单标题" span={2}>
            {currentTicket.title}
          </Descriptions.Item>
          <Descriptions.Item label="优先级">
            <Tag color={getPriorityColor(currentTicket.priority)}>
              {currentTicket.priority.toUpperCase()}
            </Tag>
          </Descriptions.Item>
          <Descriptions.Item label="状态">
            <Badge
              status={getStatusColor(currentTicket.status) as any}
              text={getStatusText(currentTicket.status)}
            />
          </Descriptions.Item>
          <Descriptions.Item label="分配给">
            {currentTicket.assignee || '未分配'}
          </Descriptions.Item>
          <Descriptions.Item label="创建人">
            {currentTicket.creator}
          </Descriptions.Item>
          <Descriptions.Item label="创建时间">
            {formatDate(currentTicket.created_at)}
          </Descriptions.Item>
          <Descriptions.Item label="更新时间">
            {formatDate(currentTicket.updated_at)}
          </Descriptions.Item>
          <Descriptions.Item label="工单描述" span={2}>
            <div style={{ whiteSpace: 'pre-wrap' }}>
              {currentTicket.description}
            </div>
          </Descriptions.Item>
        </Descriptions>
      </Card>

      {/* 处理历史 */}
      <Card title="处理历史">
        <Timeline>
          {currentTicket.comments?.map((comment, index) => (
            <Timeline.Item key={index}>
              <div>
                <div style={{ fontWeight: 'bold', marginBottom: 4 }}>
                  {comment.author} - {formatDate(comment.created_at)}
                </div>
                <div style={{ whiteSpace: 'pre-wrap' }}>
                  {comment.content}
                </div>
              </div>
            </Timeline.Item>
          )) || (
            <Timeline.Item>
              <div style={{ color: '#999' }}>暂无处理记录</div>
            </Timeline.Item>
          )}
        </Timeline>
      </Card>

      {/* 添加评论模态框 */}
      <Modal
        title="添加评论"
        open={commentModalVisible}
        onCancel={() => {
          setCommentModalVisible(false);
          commentForm.resetFields();
        }}
        footer={null}
      >
        <Form
          form={commentForm}
          layout="vertical"
          onFinish={handleAddComment}
        >
          <Form.Item
            name="content"
            label="评论内容"
            rules={[
              { required: true, message: '请输入评论内容' },
              { min: 1, message: '评论内容不能为空' },
            ]}
          >
            <TextArea
              rows={4}
              placeholder="请输入评论内容..."
              maxLength={1000}
              showCount
            />
          </Form.Item>
          <Form.Item style={{ marginBottom: 0, textAlign: 'right' }}>
            <Space>
              <Button
                onClick={() => {
                  setCommentModalVisible(false);
                  commentForm.resetFields();
                }}
              >
                取消
              </Button>
              <Button type="primary" htmlType="submit">
                添加评论
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>

      {/* 更新状态模态框 */}
      <Modal
        title="更新工单状态"
        open={statusModalVisible}
        onCancel={() => {
          setStatusModalVisible(false);
          statusForm.resetFields();
        }}
        footer={null}
      >
        <Form
          form={statusForm}
          layout="vertical"
          onFinish={handleUpdateStatus}
          initialValues={{ status: currentTicket.status }}
        >
          <Form.Item
            name="status"
            label="新状态"
            rules={[{ required: true, message: '请选择新状态' }]}
          >
            <Select placeholder="请选择新状态">
              <Option value="open">待处理</Option>
              <Option value="in_progress">处理中</Option>
              <Option value="resolved">已解决</Option>
              <Option value="closed">已关闭</Option>
            </Select>
          </Form.Item>
          <Form.Item
            name="comment"
            label="备注说明"
            rules={[{ required: true, message: '请输入备注说明' }]}
          >
            <TextArea
              rows={3}
              placeholder="请输入状态更新的备注说明..."
              maxLength={500}
              showCount
            />
          </Form.Item>
          <Form.Item style={{ marginBottom: 0, textAlign: 'right' }}>
            <Space>
              <Button
                onClick={() => {
                  setStatusModalVisible(false);
                  statusForm.resetFields();
                }}
              >
                取消
              </Button>
              <Button type="primary" htmlType="submit">
                更新状态
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
};

export default TicketDetail;