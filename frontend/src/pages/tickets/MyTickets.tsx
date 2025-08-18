import React, { useState, useEffect } from 'react';
import {
  Card,
  Table,
  Button,
  Space,
  Tag,
  Input,
  Select,
  DatePicker,
  Row,
  Col,
  Tabs,
  Badge,
  Tooltip,
  message,
  Popconfirm
} from 'antd';
import {
  PlusOutlined,
  SearchOutlined,
  ReloadOutlined,
  EditOutlined,
  DeleteOutlined,
  EyeOutlined,
  ClockCircleOutlined,
  UserOutlined
} from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import { format } from 'date-fns';
import { useTicket } from '../../hooks/useTicket';
import { useUI } from '../../hooks/useUI';
import { ticketService } from '../../services/ticket';
import type { Ticket, TicketStatus, TicketPriority } from '../../types';
import type { ColumnsType } from 'antd/es/table';

const { RangePicker } = DatePicker;
// const { TabPane } = Tabs; // 已废弃，使用items属性替代

interface TicketQuery {
  page: number;
  limit: number;
  status?: TicketStatus[];
  priority?: TicketPriority[];
  search?: string;
  start_time?: string;
  end_time?: string;
}

const MyTickets: React.FC = () => {
  const navigate = useNavigate();
  const { loading, setLoading } = useUI();
  
  const [assignedTickets, setAssignedTickets] = useState<Ticket[]>([]);
  const [createdTickets, setCreatedTickets] = useState<Ticket[]>([]);
  const [assignedTotal, setAssignedTotal] = useState(0);
  const [createdTotal, setCreatedTotal] = useState(0);
  const [activeTab, setActiveTab] = useState('assigned');
  
  const [query, setQuery] = useState<TicketQuery>({
    page: 1,
    limit: 10
  });

  // 加载分配给我的工单
  const loadAssignedTickets = async () => {
    try {
      setLoading(true);
      const response = await ticketService.getMyTickets(query);
      setAssignedTickets(response.tickets);
      setAssignedTotal(response.total);
    } catch (error) {
      message.error('加载工单失败');
    } finally {
      setLoading(false);
    }
  };

  // 加载我创建的工单
  const loadCreatedTickets = async () => {
    try {
      setLoading(true);
      const response = await ticketService.getMyCreatedTickets(query);
      setCreatedTickets(response.tickets);
      setCreatedTotal(response.total);
    } catch (error) {
      message.error('加载工单失败');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    if (activeTab === 'assigned') {
      loadAssignedTickets();
    } else {
      loadCreatedTickets();
    }
  }, [query, activeTab]);

  // 处理搜索
  const handleSearch = (value: string) => {
    setQuery(prev => ({ ...prev, search: value, page: 1 }));
  };

  // 处理筛选
  const handleFilter = (key: string, value: any) => {
    setQuery(prev => ({ ...prev, [key]: value, page: 1 }));
  };

  // 处理日期范围筛选
  const handleDateRangeChange = (dates: any) => {
    if (dates && dates.length === 2) {
      setQuery(prev => ({
        ...prev,
        start_time: dates[0].toISOString(),
        end_time: dates[1].toISOString(),
        page: 1
      }));
    } else {
      setQuery(prev => {
        const { start_time, end_time, ...rest } = prev;
        return { ...rest, page: 1 };
      });
    }
  };

  // 处理分页
  const handleTableChange = (pagination: any) => {
    setQuery(prev => ({
      ...prev,
      page: pagination.current,
      limit: pagination.pageSize
    }));
  };

  // 快速更新状态
  const handleQuickStatusUpdate = async (id: string, status: TicketStatus) => {
    try {
      await ticketService.updateTicketStatus(id, status);
      message.success('状态更新成功');
      if (activeTab === 'assigned') {
        loadAssignedTickets();
      } else {
        loadCreatedTickets();
      }
    } catch (error) {
      message.error('状态更新失败');
    }
  };

  // 删除工单
  const handleDelete = async (id: string) => {
    try {
      await ticketService.deleteTicket(id);
      message.success('工单删除成功');
      if (activeTab === 'assigned') {
        loadAssignedTickets();
      } else {
        loadCreatedTickets();
      }
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

  const columns: ColumnsType<Ticket> = [
    {
      title: 'ID',
      dataIndex: 'id',
      key: 'id',
      width: 80,
      render: (id: string) => (
        <Button
          type="link"
          onClick={() => navigate(`/tickets/${id}`)}
        >
          #{id.slice(-6)}
        </Button>
      )
    },
    {
      title: '标题',
      dataIndex: 'title',
      key: 'title',
      ellipsis: true,
      render: (title: string, record: Ticket) => (
        <Tooltip title={title}>
          <Button
            type="link"
            onClick={() => navigate(`/tickets/${record.id}`)}
            style={{ padding: 0, height: 'auto' }}
          >
            {title}
          </Button>
        </Tooltip>
      )
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 100,
      render: (status: TicketStatus, record: Ticket) => (
        <Select
          value={status}
          size="small"
          style={{ width: '100%' }}
          onChange={(newStatus) => handleQuickStatusUpdate(record.id, newStatus)}
        >
          <Select.Option value="open">
            <Tag color="blue">待处理</Tag>
          </Select.Option>
          <Select.Option value="in_progress">
            <Tag color="orange">处理中</Tag>
          </Select.Option>
          <Select.Option value="resolved">
            <Tag color="green">已解决</Tag>
          </Select.Option>
          <Select.Option value="closed">
            <Tag color="gray">已关闭</Tag>
          </Select.Option>
        </Select>
      )
    },
    {
      title: '优先级',
      dataIndex: 'priority',
      key: 'priority',
      width: 80,
      render: (priority: TicketPriority) => (
        <Tag color={getPriorityColor(priority)}>
          {getPriorityText(priority)}
        </Tag>
      )
    },
    {
      title: activeTab === 'assigned' ? '创建人' : '分配给',
      key: 'user',
      width: 120,
      render: (_, record: Ticket) => {
        const user = activeTab === 'assigned' ? record.created_by : record.assigned_to;
        return user ? (
          <Space>
            <UserOutlined />
            <span>{user}</span>
          </Space>
        ) : (
          <span style={{ color: '#999' }}>未分配</span>
        );
      }
    },
    {
      title: '创建时间',
      dataIndex: 'created_at',
      key: 'created_at',
      width: 150,
      render: (date: string) => (
        <Space>
          <ClockCircleOutlined />
          {format(new Date(date), 'MM-dd HH:mm')}
        </Space>
      )
    },
    {
      title: '截止时间',
      dataIndex: 'due_date',
      key: 'due_date',
      width: 150,
      render: (date: string) => {
        if (!date) return <span style={{ color: '#999' }}>无</span>;
        
        const dueDate = new Date(date);
        const now = new Date();
        const isOverdue = dueDate < now;
        
        return (
          <Space>
            <ClockCircleOutlined style={{ color: isOverdue ? '#ff4d4f' : undefined }} />
            <span style={{ color: isOverdue ? '#ff4d4f' : undefined }}>
              {format(dueDate, 'MM-dd HH:mm')}
            </span>
            {isOverdue && <Badge status="error" text="逾期" />}
          </Space>
        );
      }
    },
    {
      title: '操作',
      key: 'actions',
      width: 150,
      render: (_, record: Ticket) => (
        <Space>
          <Tooltip title="查看详情">
            <Button
              type="text"
              icon={<EyeOutlined />}
              onClick={() => navigate(`/tickets/${record.id}`)}
            />
          </Tooltip>
          <Tooltip title="编辑">
            <Button
              type="text"
              icon={<EditOutlined />}
              onClick={() => navigate(`/tickets/${record.id}/edit`)}
            />
          </Tooltip>
          {activeTab === 'created' && (
            <Tooltip title="删除">
              <Popconfirm
                title="确定要删除这个工单吗？"
                onConfirm={() => handleDelete(record.id)}
                okText="确定"
                cancelText="取消"
              >
                <Button
                  type="text"
                  danger
                  icon={<DeleteOutlined />}
                />
              </Popconfirm>
            </Tooltip>
          )}
        </Space>
      )
    }
  ];

  const renderFilters = () => (
    <Row gutter={16} style={{ marginBottom: '16px' }}>
      <Col span={6}>
        <Input.Search
          placeholder="搜索工单标题或描述"
          allowClear
          onSearch={handleSearch}
          style={{ width: '100%' }}
        />
      </Col>
      <Col span={4}>
        <Select
          placeholder="状态筛选"
          allowClear
          mode="multiple"
          style={{ width: '100%' }}
          onChange={(value) => handleFilter('status', value)}
        >
          <Select.Option value="open">待处理</Select.Option>
          <Select.Option value="in_progress">处理中</Select.Option>
          <Select.Option value="resolved">已解决</Select.Option>
          <Select.Option value="closed">已关闭</Select.Option>
        </Select>
      </Col>
      <Col span={4}>
        <Select
          placeholder="优先级筛选"
          allowClear
          mode="multiple"
          style={{ width: '100%' }}
          onChange={(value) => handleFilter('priority', value)}
        >
          <Select.Option value="low">低</Select.Option>
          <Select.Option value="medium">中</Select.Option>
          <Select.Option value="high">高</Select.Option>
          <Select.Option value="urgent">紧急</Select.Option>
        </Select>
      </Col>
      <Col span={6}>
        <RangePicker
          style={{ width: '100%' }}
          placeholder={['开始时间', '结束时间']}
          onChange={handleDateRangeChange}
        />
      </Col>
      <Col span={4}>
        <Space>
          <Button
            icon={<ReloadOutlined />}
            onClick={() => {
              if (activeTab === 'assigned') {
                loadAssignedTickets();
              } else {
                loadCreatedTickets();
              }
            }}
          >
            刷新
          </Button>
          <Button
            type="primary"
            icon={<PlusOutlined />}
            onClick={() => navigate('/tickets/create')}
          >
            创建工单
          </Button>
        </Space>
      </Col>
    </Row>
  );

  return (
    <div style={{ padding: '24px' }}>
      <Card>
        <Tabs
          activeKey={activeTab}
          onChange={setActiveTab}
          tabBarExtraContent={renderFilters()}
          items={[
            {
              key: 'assigned',
              label: (
                <Badge count={assignedTotal} offset={[10, 0]}>
                  <span>分配给我的</span>
                </Badge>
              ),
              children: (
                <Table
                  columns={columns}
                  dataSource={assignedTickets}
                  rowKey="id"
                  loading={loading}
                  pagination={{
                    current: query.page,
                    pageSize: query.limit,
                    total: assignedTotal,
                    showSizeChanger: true,
                    showQuickJumper: true,
                    showTotal: (total, range) =>
                      `第 ${range[0]}-${range[1]} 条，共 ${total} 条`
                  }}
                  onChange={handleTableChange}
                  scroll={{ x: 1200 }}
                />
              )
            },
            {
              key: 'created',
              label: (
                <Badge count={createdTotal} offset={[10, 0]}>
                  <span>我创建的</span>
                </Badge>
              ),
              children: (
                <Table
                  columns={columns}
                  dataSource={createdTickets}
                  rowKey="id"
                  loading={loading}
                  pagination={{
                    current: query.page,
                    pageSize: query.limit,
                    total: createdTotal,
                    showSizeChanger: true,
                    showQuickJumper: true,
                    showTotal: (total, range) =>
                      `第 ${range[0]}-${range[1]} 条，共 ${total} 条`
                  }}
                  onChange={handleTableChange}
                  scroll={{ x: 1200 }}
                />
              )
            }
          ]}
        />
      </Card>
    </div>
  );
};

export default MyTickets;