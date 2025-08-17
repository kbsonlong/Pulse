import React, { useEffect, useState } from 'react';
import {
  Table,
  Card,
  Button,
  Input,
  Select,
  Space,
  Modal,
  message,
  Tag,
  Tooltip,
  Badge,
  Popconfirm,
} from 'antd';
import {
  PlusOutlined,
  SearchOutlined,
  ReloadOutlined,
  EditOutlined,
  DeleteOutlined,
  EyeOutlined,
} from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import { useTicket, useUI } from '../../hooks';
import { Ticket, TicketStatus, TicketPriority } from '../../types';
import { formatDate } from '../../utils';
import type { ColumnsType } from 'antd/es/table';

const { Search } = Input;
const { Option } = Select;

const TicketList: React.FC = () => {
  const navigate = useNavigate();
  const { setBreadcrumb } = useUI();
  const {
    tickets,
    total,
    page,
    limit,
    loading,
    filters,
    fetchTickets,
    deleteTicket,
    updateTicketStatus,
    setFilters,
    clearFilters,
    setPage,
    setLimit,
  } = useTicket();

  const [searchText, setSearchText] = useState('');

  useEffect(() => {
    setBreadcrumb([
      { title: '工单管理' },
      { title: '工单列表' },
    ]);
    fetchTickets();
  }, [setBreadcrumb, fetchTickets]);

  // 处理搜索
  const handleSearch = (value: string) => {
    setSearchText(value);
    setFilters({ ...filters, search: value });
    fetchTickets({ ...filters, search: value, page: 1 });
  };

  // 处理筛选
  const handleFilterChange = (key: string, value: any) => {
    const newFilters = { ...filters, [key]: value };
    setFilters(newFilters);
    fetchTickets({ ...newFilters, page: 1 });
  };

  // 清除筛选
  const handleClearFilters = () => {
    setSearchText('');
    clearFilters();
    fetchTickets({ page: 1 });
  };

  // 刷新数据
  const handleRefresh = () => {
    fetchTickets({ ...filters, page });
  };

  // 创建工单
  const handleCreate = () => {
    navigate('/tickets/create');
  };

  // 查看详情
  const handleViewDetail = (record: Ticket) => {
    navigate(`/tickets/${record.id}`);
  };

  // 编辑工单
  const handleEdit = (record: Ticket) => {
    navigate(`/tickets/${record.id}/edit`);
  };

  // 删除工单
  const handleDelete = async (id: string) => {
    try {
      await deleteTicket(id);
      message.success('删除成功');
      fetchTickets({ ...filters, page });
    } catch (error) {
      message.error('删除失败');
    }
  };

  // 更新工单状态
  const handleUpdateStatus = async (id: string, status: TicketStatus) => {
    try {
      await updateTicketStatus(id, status);
      message.success('状态更新成功');
      fetchTickets({ ...filters, page });
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

  // 表格列定义
  const columns: ColumnsType<Ticket> = [
    {
      title: '工单标题',
      dataIndex: 'title',
      key: 'title',
      ellipsis: true,
      render: (text: string, record: Ticket) => (
        <Button
          type="link"
          onClick={() => handleViewDetail(record)}
          style={{ padding: 0, height: 'auto' }}
        >
          {text}
        </Button>
      ),
    },
    {
      title: '优先级',
      dataIndex: 'priority',
      key: 'priority',
      width: 100,
      render: (priority: TicketPriority) => (
        <Tag color={getPriorityColor(priority)}>
          {priority.toUpperCase()}
        </Tag>
      ),
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 100,
      render: (status: TicketStatus) => (
        <Badge
          status={getStatusColor(status) as any}
          text={getStatusText(status)}
        />
      ),
    },
    {
      title: '分配给',
      dataIndex: 'assignee',
      key: 'assignee',
      width: 120,
      ellipsis: true,
      render: (assignee: string) => assignee || '-',
    },
    {
      title: '创建人',
      dataIndex: 'creator',
      key: 'creator',
      width: 120,
      ellipsis: true,
    },
    {
      title: '创建时间',
      dataIndex: 'created_at',
      key: 'created_at',
      width: 160,
      render: (date: string) => formatDate(date),
    },
    {
      title: '更新时间',
      dataIndex: 'updated_at',
      key: 'updated_at',
      width: 160,
      render: (date: string) => formatDate(date),
    },
    {
      title: '操作',
      key: 'action',
      width: 200,
      render: (_, record: Ticket) => (
        <Space size="small">
          <Tooltip title="查看详情">
            <Button
              type="text"
              icon={<EyeOutlined />}
              onClick={() => handleViewDetail(record)}
            />
          </Tooltip>
          <Tooltip title="编辑">
            <Button
              type="text"
              icon={<EditOutlined />}
              onClick={() => handleEdit(record)}
            />
          </Tooltip>
          {record.status === 'open' && (
            <Button
              type="text"
              size="small"
              onClick={() => handleUpdateStatus(record.id, 'in_progress')}
            >
              开始处理
            </Button>
          )}
          {record.status === 'in_progress' && (
            <Button
              type="text"
              size="small"
              onClick={() => handleUpdateStatus(record.id, 'resolved')}
            >
              标记解决
            </Button>
          )}
          <Popconfirm
            title="确定要删除这个工单吗？"
            onConfirm={() => handleDelete(record.id)}
            okText="确定"
            cancelText="取消"
          >
            <Tooltip title="删除">
              <Button
                type="text"
                icon={<DeleteOutlined />}
                danger
              />
            </Tooltip>
          </Popconfirm>
        </Space>
      ),
    },
  ];

  return (
    <Card>
      {/* 操作栏 */}
      <div style={{ marginBottom: 16 }}>
        <Space wrap>
          <Button
            type="primary"
            icon={<PlusOutlined />}
            onClick={handleCreate}
          >
            创建工单
          </Button>
          <Search
            placeholder="搜索工单标题或描述"
            value={searchText}
            onChange={(e) => setSearchText(e.target.value)}
            onSearch={handleSearch}
            style={{ width: 250 }}
            allowClear
          />
          <Select
            placeholder="优先级"
            style={{ width: 120 }}
            value={filters.priority}
            onChange={(value) => handleFilterChange('priority', value)}
            allowClear
          >
            <Option value="critical">Critical</Option>
            <Option value="high">High</Option>
            <Option value="medium">Medium</Option>
            <Option value="low">Low</Option>
          </Select>
          <Select
            placeholder="状态"
            style={{ width: 120 }}
            value={filters.status}
            onChange={(value) => handleFilterChange('status', value)}
            allowClear
          >
            <Option value="open">待处理</Option>
            <Option value="in_progress">处理中</Option>
            <Option value="resolved">已解决</Option>
            <Option value="closed">已关闭</Option>
          </Select>
          <Select
            placeholder="分配给"
            style={{ width: 150 }}
            value={filters.assignee}
            onChange={(value) => handleFilterChange('assignee', value)}
            allowClear
          >
            {/* 这里应该从用户列表中获取选项 */}
            <Option value="admin">管理员</Option>
            <Option value="user1">用户1</Option>
            <Option value="user2">用户2</Option>
          </Select>
          <Button onClick={handleClearFilters}>清除筛选</Button>
          <Button icon={<ReloadOutlined />} onClick={handleRefresh}>
            刷新
          </Button>
        </Space>
      </div>

      {/* 表格 */}
      <Table
        columns={columns}
        dataSource={tickets}
        rowKey="id"
        loading={loading}
        pagination={{
          current: page,
          pageSize: limit,
          total,
          showSizeChanger: true,
          showQuickJumper: true,
          showTotal: (total, range) =>
            `第 ${range[0]}-${range[1]} 条，共 ${total} 条`,
          onChange: (newPage, newPageSize) => {
            setPage(newPage);
            if (newPageSize !== limit) {
              setLimit(newPageSize);
            }
            fetchTickets({ ...filters, page: newPage, limit: newPageSize });
          },
        }}
        scroll={{ x: 1400 }}
      />
    </Card>
  );
};

export default TicketList;