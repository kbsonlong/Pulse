import React, { useEffect, useState } from 'react';
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
  Avatar,
  Typography,
  Tooltip,
} from 'antd';
import {
  PlusOutlined,
  SearchOutlined,
  ReloadOutlined,
  EyeOutlined,
  EditOutlined,
  UserOutlined,
} from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import { useTicket, useUI } from '../../hooks';
import { Ticket, TicketStatus, TicketPriority } from '../../types';
import { formatDateTime } from '../../utils/date';
import type { ColumnsType } from 'antd/es/table';
import type { TabsProps } from 'antd';

const { RangePicker } = DatePicker;
const { Option } = Select;
const { Text } = Typography;

const MyTickets: React.FC = () => {
  const navigate = useNavigate();
  const { setBreadcrumbs } = useUI();
  const {
    tickets,
    loading,
    pagination,
    filters,
    fetchTickets,
    setFilters,
    setPage,
    setLimit,
  } = useTicket();

  const [searchText, setSearchText] = useState('');
  const [activeTab, setActiveTab] = useState('assigned');

  useEffect(() => {
    setBreadcrumbs([
      { title: '工单管理' },
      { title: '我的工单' },
    ]);

    // 根据当前标签页设置不同的筛选条件
    handleTabChange(activeTab);
  }, [setBreadcrumbs]);

  // 标签页切换
  const handleTabChange = (key: string) => {
    setActiveTab(key);
    const newFilters = { ...filters };
    
    // 清除之前的筛选条件
    delete newFilters.assignee_id;
    delete newFilters.creator_id;
    
    if (key === 'assigned') {
      // 分配给我的工单
      newFilters.assignee_id = 'current_user'; // 后端会识别这个特殊值
    } else if (key === 'created') {
      // 我创建的工单
      newFilters.creator_id = 'current_user'; // 后端会识别这个特殊值
    }
    
    setFilters(newFilters);
    setPage(1);
  };

  // 搜索
  const handleSearch = () => {
    setFilters({
      ...filters,
      search: searchText,
    });
    setPage(1);
  };

  // 重置搜索
  const handleReset = () => {
    setSearchText('');
    setFilters({});
    setPage(1);
    handleTabChange(activeTab); // 重新应用标签页筛选
  };

  // 刷新数据
  const handleRefresh = () => {
    fetchTickets();
  };

  // 查看详情
  const handleView = (record: Ticket) => {
    navigate(`/tickets/detail/${record.id}`);
  };

  // 编辑工单
  const handleEdit = (record: Ticket) => {
    navigate(`/tickets/edit/${record.id}`);
  };

  // 创建工单
  const handleCreate = () => {
    navigate('/tickets/create');
  };

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

  // 表格列定义
  const columns: ColumnsType<Ticket> = [
    {
      title: 'ID',
      dataIndex: 'id',
      key: 'id',
      width: 80,
      render: (id: string) => (
        <Text code style={{ fontSize: '12px' }}>
          {id.slice(0, 8)}
        </Text>
      ),
    },
    {
      title: '标题',
      dataIndex: 'title',
      key: 'title',
      ellipsis: {
        showTitle: false,
      },
      render: (title: string) => (
        <Tooltip placement="topLeft" title={title}>
          {title}
        </Tooltip>
      ),
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 100,
      render: (status: TicketStatus) => getStatusTag(status),
    },
    {
      title: '优先级',
      dataIndex: 'priority',
      key: 'priority',
      width: 100,
      render: (priority: TicketPriority) => getPriorityTag(priority),
    },
    {
      title: '创建人',
      dataIndex: 'creator',
      key: 'creator',
      width: 120,
      render: (creator: any) => (
        <Space>
          <Avatar size="small" icon={<UserOutlined />} />
          <Text>{creator?.username}</Text>
        </Space>
      ),
    },
    {
      title: '分配人',
      dataIndex: 'assignee',
      key: 'assignee',
      width: 120,
      render: (assignee: any) => (
        assignee ? (
          <Space>
            <Avatar size="small" icon={<UserOutlined />} />
            <Text>{assignee.username}</Text>
          </Space>
        ) : (
          <Text type="secondary">未分配</Text>
        )
      ),
    },
    {
      title: '创建时间',
      dataIndex: 'created_at',
      key: 'created_at',
      width: 180,
      render: (date: string) => formatDateTime(date),
    },
    {
      title: '操作',
      key: 'action',
      width: 120,
      render: (_, record) => (
        <Space size="small">
          <Tooltip title="查看详情">
            <Button
              type="text"
              size="small"
              icon={<EyeOutlined />}
              onClick={() => handleView(record)}
            />
          </Tooltip>
          <Tooltip title="编辑">
            <Button
              type="text"
              size="small"
              icon={<EditOutlined />}
              onClick={() => handleEdit(record)}
            />
          </Tooltip>
        </Space>
      ),
    },
  ];

  // 标签页配置
  const tabItems: TabsProps['items'] = [
    {
      key: 'assigned',
      label: '分配给我的',
    },
    {
      key: 'created',
      label: '我创建的',
    },
  ];

  return (
    <div style={{ padding: '24px' }}>
      <Card>
        <div style={{ marginBottom: '16px' }}>
          <Row gutter={[16, 16]}>
            <Col span={8}>
              <Input
                placeholder="搜索工单标题或描述"
                value={searchText}
                onChange={(e) => setSearchText(e.target.value)}
                onPressEnter={handleSearch}
                suffix={<SearchOutlined />}
              />
            </Col>
            <Col span={4}>
              <Select
                placeholder="状态"
                allowClear
                style={{ width: '100%' }}
                value={filters.status}
                onChange={(value) => setFilters({ ...filters, status: value })}
              >
                <Option value="open">待处理</Option>
                <Option value="in_progress">处理中</Option>
                <Option value="resolved">已解决</Option>
                <Option value="closed">已关闭</Option>
              </Select>
            </Col>
            <Col span={4}>
              <Select
                placeholder="优先级"
                allowClear
                style={{ width: '100%' }}
                value={filters.priority}
                onChange={(value) => setFilters({ ...filters, priority: value })}
              >
                <Option value="low">低</Option>
                <Option value="medium">中</Option>
                <Option value="high">高</Option>
                <Option value="critical">紧急</Option>
              </Select>
            </Col>
            <Col span={8}>
              <Space>
                <Button
                  type="primary"
                  icon={<SearchOutlined />}
                  onClick={handleSearch}
                >
                  搜索
                </Button>
                <Button onClick={handleReset}>
                  重置
                </Button>
                <Button
                  icon={<ReloadOutlined />}
                  onClick={handleRefresh}
                >
                  刷新
                </Button>
                <Button
                  type="primary"
                  icon={<PlusOutlined />}
                  onClick={handleCreate}
                >
                  创建工单
                </Button>
              </Space>
            </Col>
          </Row>
        </div>

        <Tabs
          activeKey={activeTab}
          onChange={handleTabChange}
          items={tabItems}
          style={{ marginBottom: '16px' }}
        />

        <Table
          columns={columns}
          dataSource={tickets}
          rowKey="id"
          loading={loading}
          pagination={{
            current: pagination.page,
            pageSize: pagination.limit,
            total: pagination.total,
            showSizeChanger: true,
            showQuickJumper: true,
            showTotal: (total, range) =>
              `第 ${range[0]}-${range[1]} 条/共 ${total} 条`,
            onChange: (page, pageSize) => {
              setPage(page);
              if (pageSize !== pagination.limit) {
                setLimit(pageSize);
              }
            },
          }}
          scroll={{ x: 1200 }}
        />
      </Card>
    </div>
  );
};

export default MyTickets;