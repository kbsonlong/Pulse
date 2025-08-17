import React, { useEffect, useState } from 'react';
import {
  Table,
  Card,
  Button,
  Input,
  Select,
  DatePicker,
  Tag,
  Space,
  Modal,
  message,
  Tooltip,
  Badge,
} from 'antd';
import {
  SearchOutlined,
  ReloadOutlined,
  ExclamationCircleOutlined,
  EyeOutlined,
  CheckOutlined,
  CloseOutlined,
} from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import { useAlert, useUI } from '../../hooks';
import { Alert, AlertSeverity, AlertStatus } from '../../types';
import { formatDate, getPriorityColor } from '../../utils';
import type { ColumnsType } from 'antd/es/table';
import dayjs from 'dayjs';

const { Search } = Input;
const { Option } = Select;
const { RangePicker } = DatePicker;
const { confirm } = Modal;

const AlertList: React.FC = () => {
  const navigate = useNavigate();
  const { setBreadcrumb } = useUI();
  const {
    alerts,
    total,
    page,
    limit,
    loading,
    filters,
    fetchAlerts,
    updateStatus,
    batchUpdateStatus,
    setFilters,
    clearFilters,
    setPage,
    setLimit,
  } = useAlert();

  const [selectedRowKeys, setSelectedRowKeys] = useState<string[]>([]);
  const [searchText, setSearchText] = useState('');

  useEffect(() => {
    setBreadcrumb([
      { title: '告警管理' },
      { title: '告警列表' },
    ]);
    fetchAlerts();
  }, [setBreadcrumb, fetchAlerts]);

  // 处理搜索
  const handleSearch = (value: string) => {
    setSearchText(value);
    setFilters({ ...filters, search: value });
    fetchAlerts({ ...filters, search: value, page: 1 });
  };

  // 处理筛选
  const handleFilterChange = (key: string, value: any) => {
    const newFilters = { ...filters, [key]: value };
    setFilters(newFilters);
    fetchAlerts({ ...newFilters, page: 1 });
  };

  // 处理时间范围筛选
  const handleDateRangeChange = (dates: any) => {
    const newFilters = {
      ...filters,
      start_time: dates?.[0]?.format('YYYY-MM-DD HH:mm:ss'),
      end_time: dates?.[1]?.format('YYYY-MM-DD HH:mm:ss'),
    };
    setFilters(newFilters);
    fetchAlerts({ ...newFilters, page: 1 });
  };

  // 清除筛选
  const handleClearFilters = () => {
    setSearchText('');
    clearFilters();
    fetchAlerts({ page: 1 });
  };

  // 刷新数据
  const handleRefresh = () => {
    fetchAlerts({ ...filters, page });
  };

  // 查看详情
  const handleViewDetail = (record: Alert) => {
    navigate(`/alerts/${record.id}`);
  };

  // 更新告警状态
  const handleUpdateStatus = (id: string, status: AlertStatus) => {
    confirm({
      title: '确认操作',
      icon: <ExclamationCircleOutlined />,
      content: `确定要将告警状态更新为"${status === 'resolved' ? '已解决' : '未解决'}"吗？`,
      onOk: async () => {
        try {
          await updateStatus(id, status);
          message.success('状态更新成功');
          fetchAlerts({ ...filters, page });
        } catch (error) {
          message.error('状态更新失败');
        }
      },
    });
  };

  // 批量更新状态
  const handleBatchUpdateStatus = (status: AlertStatus) => {
    if (selectedRowKeys.length === 0) {
      message.warning('请选择要操作的告警');
      return;
    }

    confirm({
      title: '确认批量操作',
      icon: <ExclamationCircleOutlined />,
      content: `确定要将选中的 ${selectedRowKeys.length} 个告警状态更新为"${status === 'resolved' ? '已解决' : '未解决'}"吗？`,
      onOk: async () => {
        try {
          await batchUpdateStatus(selectedRowKeys, status);
          message.success('批量更新成功');
          setSelectedRowKeys([]);
          fetchAlerts({ ...filters, page });
        } catch (error) {
          message.error('批量更新失败');
        }
      },
    });
  };

  // 表格列定义
  const columns: ColumnsType<Alert> = [
    {
      title: '告警名称',
      dataIndex: 'name',
      key: 'name',
      ellipsis: true,
      render: (text: string, record: Alert) => (
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
      title: '严重程度',
      dataIndex: 'severity',
      key: 'severity',
      width: 100,
      render: (severity: AlertSeverity) => (
        <Tag color={getPriorityColor(severity)}>
          {severity.toUpperCase()}
        </Tag>
      ),
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 100,
      render: (status: AlertStatus) => (
        <Badge
          status={status === 'resolved' ? 'success' : 'error'}
          text={status === 'resolved' ? '已解决' : '未解决'}
        />
      ),
    },
    {
      title: '数据源',
      dataIndex: 'source',
      key: 'source',
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
      width: 150,
      render: (_, record: Alert) => (
        <Space size="small">
          <Tooltip title="查看详情">
            <Button
              type="text"
              icon={<EyeOutlined />}
              onClick={() => handleViewDetail(record)}
            />
          </Tooltip>
          {record.status !== 'resolved' && (
            <Tooltip title="标记为已解决">
              <Button
                type="text"
                icon={<CheckOutlined />}
                onClick={() => handleUpdateStatus(record.id, 'resolved')}
              />
            </Tooltip>
          )}
          {record.status === 'resolved' && (
            <Tooltip title="标记为未解决">
              <Button
                type="text"
                icon={<CloseOutlined />}
                onClick={() => handleUpdateStatus(record.id, 'active')}
              />
            </Tooltip>
          )}
        </Space>
      ),
    },
  ];

  // 行选择配置
  const rowSelection = {
    selectedRowKeys,
    onChange: (keys: React.Key[]) => {
      setSelectedRowKeys(keys as string[]);
    },
  };

  return (
    <Card>
      {/* 筛选区域 */}
      <div style={{ marginBottom: 16 }}>
        <Space wrap>
          <Search
            placeholder="搜索告警名称或描述"
            value={searchText}
            onChange={(e) => setSearchText(e.target.value)}
            onSearch={handleSearch}
            style={{ width: 250 }}
            allowClear
          />
          <Select
            placeholder="严重程度"
            style={{ width: 120 }}
            value={filters.severity}
            onChange={(value) => handleFilterChange('severity', value)}
            allowClear
          >
            <Option value="critical">Critical</Option>
            <Option value="high">High</Option>
            <Option value="medium">Medium</Option>
            <Option value="low">Low</Option>
          </Select>
          <Select
            placeholder="状态"
            style={{ width: 100 }}
            value={filters.status}
            onChange={(value) => handleFilterChange('status', value)}
            allowClear
          >
            <Option value="active">未解决</Option>
            <Option value="resolved">已解决</Option>
          </Select>
          <RangePicker
            placeholder={['开始时间', '结束时间']}
            onChange={handleDateRangeChange}
            showTime
            format="YYYY-MM-DD HH:mm:ss"
          />
          <Button onClick={handleClearFilters}>清除筛选</Button>
          <Button icon={<ReloadOutlined />} onClick={handleRefresh}>
            刷新
          </Button>
        </Space>
      </div>

      {/* 批量操作 */}
      {selectedRowKeys.length > 0 && (
        <div style={{ marginBottom: 16 }}>
          <Space>
            <span>已选择 {selectedRowKeys.length} 项</span>
            <Button
              type="primary"
              size="small"
              onClick={() => handleBatchUpdateStatus('resolved')}
            >
              批量标记为已解决
            </Button>
            <Button
              size="small"
              onClick={() => handleBatchUpdateStatus('active')}
            >
              批量标记为未解决
            </Button>
          </Space>
        </div>
      )}

      {/* 表格 */}
      <Table
        columns={columns}
        dataSource={alerts}
        rowKey="id"
        rowSelection={rowSelection}
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
            fetchAlerts({ ...filters, page: newPage, limit: newPageSize });
          },
        }}
        scroll={{ x: 1200 }}
      />
    </Card>
  );
};

export default AlertList;