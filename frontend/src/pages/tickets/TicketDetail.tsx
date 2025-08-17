import React, { useEffect, useState } from 'react';
import {
  Card,
  Descriptions,
  Button,
  Space,
  Tag,
  Timeline,
  Modal,
  Form,
  Input,
  Select,
  message,
  Divider,
  Row,
  Col,
  Avatar,
  Typography,
  Popconfirm,
} from 'antd';
import {
  EditOutlined,
  DeleteOutlined,
  UserOutlined,
  ClockCircleOutlined,
  CheckCircleOutlined,
  ExclamationCircleOutlined,
} from '@ant-design/icons';
import { useNavigate, useParams } from 'react-router-dom';
import { useTicket, useUI } from '../../hooks';
import { TicketStatus, TicketPriority, ProcessRecord } from '../../types';
import { formatDateTime } from '../../utils/date';

const { TextArea } = Input;
const { Option } = Select;
const { Title, Text } = Typography;

interface ProcessRecordForm {
  content: string;
  status?: TicketStatus;
}

const TicketDetail: React.FC = () => {
  const navigate = useNavigate();
  const { id } = useParams<{ id: string }>();
  const { setBreadcrumbs } = useUI();
  const {
    currentTicket,
    processRecords,
    loading,
    fetchTicket,
    updateTicketStatus,
    assignTicket,
    deleteTicket,
    fetchProcessRecords,
    addProcessRecord,
  } = useTicket();

  const [form] = Form.useForm<ProcessRecordForm>();
  const [processModalVisible, setProcessModalVisible] = useState(false);
  const [assignModalVisible, setAssignModalVisible] = useState(false);
  const [statusModalVisible, setStatusModalVisible] = useState(false);
  const [submitting, setSubmitting] = useState(false);

  useEffect(() => {
    setBreadcrumbs([
      { title: '工单管理' },
      { title: '工单列表', path: '/tickets/list' },
      { title: '工单详情' },
    ]);

    if (id) {
      fetchTicket(id);
      fetchProcessRecords(id);
    }
  }, [setBreadcrumbs, id, fetchTicket, fetchProcessRecords]);

  // 获取状态标签
  const getStatusTag = (status: TicketStatus) => {
    const statusConfig = {
      open: { color: 'blue', text: '待处理' },
      in_progress: { color: 'orange', text: '处理中' },
      resolved: { color: 'green', text: '已解决' },
      closed: { color: 'default', text: '已关闭' },
    };
    const config = statusConfig[status];
    return <Tag color={config.color}>{config.text}</Tag>;
  };

  // 获取优先级标签
  const getPriorityTag = (priority: TicketPriority) => {
    const priorityConfig = {
      low: { color: 'blue', text: '低' },
      medium: { color: 'yellow', text: '中' },
      high: { color: 'orange', text: '高' },
      critical: { color: 'red', text: '紧急' },
    };
    const config = priorityConfig[priority];
    return <Tag color={config.color}>{config.text}</Tag>;
  };

  // 编辑工单
  const handleEdit = () => {
    navigate(`/tickets/edit/${id}`);
  };

  // 删除工单
  const handleDelete = async () => {
    if (!id) return;
    try {
      await deleteTicket(id);
      message.success('工单删除成功');
      navigate('/tickets/list');
    } catch (error) {
      message.error('工单删除失败');
    }
  };

  // 更新状态
  const handleUpdateStatus = async (values: { status: TicketStatus }) => {
    if (!id) return;
    setSubmitting(true);
    try {
      await updateTicketStatus(id, values.status);
      message.success('状态更新成功');
      setStatusModalVisible(false);
    } catch (error) {
      message.error('状态更新失败');
    } finally {
      setSubmitting(false);
    }
  };

  // 分配工单
  const handleAssign = async (values: { assignee_id: string }) => {
    if (!id) return;
    setSubmitting(true);
    try {
      await assignTicket(id, values.assignee_id);
      message.success('工单分配成功');
      setAssignModalVisible(false);
    } catch (error) {
      message.error('工单分配失败');
    } finally {
      setSubmitting(false);
    }
  };

  // 添加处理记录
  const handleAddProcessRecord = async (values: ProcessRecordForm) => {
    if (!id) return;
    setSubmitting(true);
    try {
      await addProcessRecord(id, values);
      message.success('处理记录添加成功');
      setProcessModalVisible(false);
      form.resetFields();
    } catch (error) {
      message.error('处理记录添加失败');
    } finally {
      setSubmitting(false);
    }
  };

  // 获取时间线图标
  const getTimelineIcon = (record: ProcessRecord) => {
    if (record.status) {
      switch (record.status) {
        case 'resolved':
          return <CheckCircleOutlined style={{ color: '#52c41a' }} />;
        case 'closed':
          return <CheckCircleOutlined style={{ color: '#8c8c8c' }} />;
        case 'in_progress':
          return <ClockCircleOutlined style={{ color: '#faad14' }} />;
        default:
          return <ExclamationCircleOutlined style={{ color: '#1890ff' }} />;
      }
    }
    return <ClockCircleOutlined style={{ color: '#8c8c8c' }} />;
  };

  if (!currentTicket) {
    return <div>工单不存在</div>;
  }

  return (
    <div style={{ padding: '24px' }}>
      <Row gutter={24}>
        <Col span={16}>
          <Card
            title={`工单详情 - ${currentTicket.title}`}
            extra={
              <Space>
                <Button
                  type="primary"
                  icon={<EditOutlined />}
                  onClick={handleEdit}
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
            }
            loading={loading}
          >
            <Descriptions column={2} bordered>
              <Descriptions.Item label="工单ID">
                {currentTicket.id}
              </Descriptions.Item>
              <Descriptions.Item label="状态">
                {getStatusTag(currentTicket.status)}
              </Descriptions.Item>
              <Descriptions.Item label="优先级">
                {getPriorityTag(currentTicket.priority)}
              </Descriptions.Item>
              <Descriptions.Item label="创建人">
                <Space>
                  <Avatar size="small" icon={<UserOutlined />} />
                  {currentTicket.creator?.username}
                </Space>
              </Descriptions.Item>
              <Descriptions.Item label="分配人">
                {currentTicket.assignee ? (
                  <Space>
                    <Avatar size="small" icon={<UserOutlined />} />
                    {currentTicket.assignee.username}
                  </Space>
                ) : (
                  <Text type="secondary">未分配</Text>
                )}
              </Descriptions.Item>
              <Descriptions.Item label="创建时间">
                {formatDateTime(currentTicket.created_at)}
              </Descriptions.Item>
              <Descriptions.Item label="更新时间">
                {formatDateTime(currentTicket.updated_at)}
              </Descriptions.Item>
              <Descriptions.Item label="关联告警">
                {currentTicket.alert_id || <Text type="secondary">无</Text>}
              </Descriptions.Item>
            </Descriptions>

            <Divider orientation="left">工单描述</Divider>
            <div style={{ padding: '16px', backgroundColor: '#fafafa', borderRadius: '6px' }}>
              <Text>{currentTicket.description}</Text>
            </div>
          </Card>
        </Col>

        <Col span={8}>
          <Card
            title="操作"
            size="small"
            style={{ marginBottom: '16px' }}
          >
            <Space direction="vertical" style={{ width: '100%' }}>
              <Button
                block
                onClick={() => setStatusModalVisible(true)}
              >
                更新状态
              </Button>
              <Button
                block
                onClick={() => setAssignModalVisible(true)}
              >
                分配工单
              </Button>
              <Button
                block
                type="primary"
                onClick={() => setProcessModalVisible(true)}
              >
                添加处理记录
              </Button>
            </Space>
          </Card>

          <Card title="处理记录" size="small">
            {processRecords.length > 0 ? (
              <Timeline>
                {processRecords.map((record) => (
                  <Timeline.Item
                    key={record.id}
                    dot={getTimelineIcon(record)}
                  >
                    <div>
                      <div style={{ marginBottom: '4px' }}>
                        <Space>
                          <Avatar size="small" icon={<UserOutlined />} />
                          <Text strong>{record.operator?.username}</Text>
                          {record.status && getStatusTag(record.status)}
                        </Space>
                      </div>
                      <Text type="secondary" style={{ fontSize: '12px' }}>
                        {formatDateTime(record.created_at)}
                      </Text>
                      <div style={{ marginTop: '8px' }}>
                        <Text>{record.content}</Text>
                      </div>
                    </div>
                  </Timeline.Item>
                ))}
              </Timeline>
            ) : (
              <Text type="secondary">暂无处理记录</Text>
            )}
          </Card>
        </Col>
      </Row>

      {/* 添加处理记录模态框 */}
      <Modal
        title="添加处理记录"
        open={processModalVisible}
        onCancel={() => setProcessModalVisible(false)}
        footer={null}
      >
        <Form
          form={form}
          layout="vertical"
          onFinish={handleAddProcessRecord}
        >
          <Form.Item
            label="处理内容"
            name="content"
            rules={[{ required: true, message: '请输入处理内容' }]}
          >
            <TextArea
              rows={4}
              placeholder="请描述处理过程和结果..."
            />
          </Form.Item>
          <Form.Item
            label="更新状态"
            name="status"
          >
            <Select placeholder="选择新状态（可选）" allowClear>
              <Option value="in_progress">处理中</Option>
              <Option value="resolved">已解决</Option>
              <Option value="closed">已关闭</Option>
            </Select>
          </Form.Item>
          <Form.Item>
            <Space>
              <Button
                type="primary"
                htmlType="submit"
                loading={submitting}
              >
                添加记录
              </Button>
              <Button onClick={() => setProcessModalVisible(false)}>
                取消
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>

      {/* 更新状态模态框 */}
      <Modal
        title="更新工单状态"
        open={statusModalVisible}
        onCancel={() => setStatusModalVisible(false)}
        footer={null}
      >
        <Form
          layout="vertical"
          onFinish={handleUpdateStatus}
          initialValues={{ status: currentTicket.status }}
        >
          <Form.Item
            label="新状态"
            name="status"
            rules={[{ required: true, message: '请选择状态' }]}
          >
            <Select>
              <Option value="open">待处理</Option>
              <Option value="in_progress">处理中</Option>
              <Option value="resolved">已解决</Option>
              <Option value="closed">已关闭</Option>
            </Select>
          </Form.Item>
          <Form.Item>
            <Space>
              <Button
                type="primary"
                htmlType="submit"
                loading={submitting}
              >
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
        <Form
          layout="vertical"
          onFinish={handleAssign}
          initialValues={{ assignee_id: currentTicket.assignee?.id }}
        >
          <Form.Item
            label="分配给"
            name="assignee_id"
            rules={[{ required: true, message: '请选择分配人员' }]}
          >
            <Select placeholder="请选择分配人员">
              {/* 这里需要从API获取用户列表 */}
            </Select>
          </Form.Item>
          <Form.Item>
            <Space>
              <Button
                type="primary"
                htmlType="submit"
                loading={submitting}
              >
                分配工单
              </Button>
              <Button onClick={() => setAssignModalVisible(false)}>
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