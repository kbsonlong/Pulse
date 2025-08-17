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
  Switch,
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
import { useRule, useUI } from '../../hooks';
import { Rule, RuleStatus } from '../../types';
import { formatDate } from '../../utils';
import type { ColumnsType } from 'antd/es/table';

const { Search } = Input;
const { Option } = Select;

const RuleList: React.FC = () => {
  const navigate = useNavigate();
  const { setBreadcrumbs } = useUI();
  const {
    rules,
    total,
    page,
    limit,
    loading,
    filters,
    fetchRules,
    deleteRule,
    updateRuleStatus,
    setFilters,
    clearFilters,
    setPage,
    setLimit,
  } = useRule();

  const [searchText, setSearchText] = useState('');

  useEffect(() => {
    setBreadcrumbs([
      { title: '规则管理' },
      { title: '规则列表' },
    ]);
    fetchRules();
  }, [setBreadcrumbs, fetchRules]);

  // 处理搜索
  const handleSearch = (value: string) => {
    setSearchText(value);
    setFilters({ ...filters, search: value });
    fetchRules({ ...filters, search: value, page: 1 });
  };

  // 处理筛选
  const handleFilterChange = (key: string, value: any) => {
    const newFilters = { ...filters, [key]: value };
    setFilters(newFilters);
    fetchRules({ ...newFilters, page: 1 });
  };

  // 清除筛选
  const handleClearFilters = () => {
    setSearchText('');
    clearFilters();
    fetchRules({ page: 1 });
  };

  // 刷新数据
  const handleRefresh = () => {
    fetchRules({ ...filters, page });
  };

  // 创建规则
  const handleCreate = () => {
    navigate('/rules/create');
  };

  // 查看详情
  const handleViewDetail = (record: Rule) => {
    navigate(`/rules/${record.id}`);
  };

  // 编辑规则
  const handleEdit = (record: Rule) => {
    navigate(`/rules/${record.id}/edit`);
  };

  // 删除规则
  const handleDelete = async (id: string) => {
    try {
      await deleteRule(id);
      message.success('删除成功');
      fetchRules({ ...filters, page });
    } catch (error) {
      message.error('删除失败');
    }
  };

  // 切换规则状态
  const handleToggleStatus = async (id: string, enabled: boolean) => {
    try {
      await updateRuleStatus(id, enabled ? 'enabled' : 'disabled');
      message.success('状态更新成功');
      fetchRules({ ...filters, page });
    } catch (error) {
      message.error('状态更新失败');
    }
  };

  // 获取严重程度颜色
  const getSeverityColor = (severity: string) => {
    const colors: Record<string, string> = {
      critical: 'red',
      high: 'orange',
      medium: 'yellow',
      low: 'blue',
    };
    return colors[severity] || 'default';
  };

  // 表格列定义
  const columns: ColumnsType<Rule> = [
    {
      title: '规则名称',
      dataIndex: 'name',
      key: 'name',
      ellipsis: true,
      render: (text: string, record: Rule) => (
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
      title: '数据源',
      dataIndex: 'datasource_id',
      key: 'datasource_id',
      width: 120,
      ellipsis: true,
    },
    {
      title: '严重程度',
      dataIndex: 'severity',
      key: 'severity',
      width: 100,
      render: (severity: string) => (
        <Tag color={getSeverityColor(severity)}>
          {severity.toUpperCase()}
        </Tag>
      ),
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 100,
      render: (status: RuleStatus, record: Rule) => (
        <Switch
          checked={status === 'enabled'}
          onChange={(checked) => handleToggleStatus(record.id, checked)}
          checkedChildren="启用"
          unCheckedChildren="禁用"
        />
      ),
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
      render: (_, record: Rule) => (
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
          <Popconfirm
            title="确定要删除这个规则吗？"
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
            创建规则
          </Button>
          <Search
            placeholder="搜索规则名称或描述"
            value={searchText}
            onChange={(e) => setSearchText(e.target.value)}
            onSearch={handleSearch}
            style={{ width: 250 }}
            allowClear
          />
          <Select
            placeholder="数据源"
            style={{ width: 150 }}
            value={filters.datasource_id}
            onChange={(value) => handleFilterChange('datasource_id', value)}
            allowClear
          >
            {/* 这里应该从数据源列表中获取选项 */}
            <Option value="prometheus">Prometheus</Option>
            <Option value="elasticsearch">Elasticsearch</Option>
            <Option value="mysql">MySQL</Option>
          </Select>
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
            <Option value="enabled">启用</Option>
            <Option value="disabled">禁用</Option>
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
        dataSource={rules}
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
            fetchRules({ ...filters, page: newPage, limit: newPageSize });
          },
        }}
        scroll={{ x: 1200 }}
      />
    </Card>
  );
};

export default RuleList;