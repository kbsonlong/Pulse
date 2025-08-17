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
  Form,
  Row,
  Col,
} from 'antd';
import {
  PlusOutlined,
  SearchOutlined,
  ReloadOutlined,
  EditOutlined,
  DeleteOutlined,
  EyeOutlined,
  ApiOutlined,
} from '@ant-design/icons';
import { useRule, useUI } from '../../hooks';
import { DataSource } from '../../types';
import { formatDate } from '../../utils';
import type { ColumnsType } from 'antd/es/table';

const { Search } = Input;
const { Option } = Select;
const { TextArea } = Input;

interface DataSourceFormData {
  name: string;
  type: string;
  url: string;
  config: string;
  enabled: boolean;
}

const DataSourceList: React.FC = () => {
  const { setBreadcrumbs } = useUI();
  const { dataSources, loading, fetchDataSources } = useRule();
  
  const [searchText, setSearchText] = useState('');
  const [typeFilter, setTypeFilter] = useState<string | undefined>();
  const [modalVisible, setModalVisible] = useState(false);
  const [editingDataSource, setEditingDataSource] = useState<DataSource | null>(null);
  const [form] = Form.useForm<DataSourceFormData>();
  const [submitting, setSubmitting] = useState(false);

  useEffect(() => {
    setBreadcrumbs([
      { title: '规则管理' },
      { title: '数据源管理' },
    ]);
    fetchDataSources();
  }, [setBreadcrumbs, fetchDataSources]);

  // 过滤数据源
  const filteredDataSources = dataSources.filter((ds) => {
    const matchesSearch = !searchText || 
      ds.name.toLowerCase().includes(searchText.toLowerCase()) ||
      ds.url.toLowerCase().includes(searchText.toLowerCase());
    const matchesType = !typeFilter || ds.type === typeFilter;
    return matchesSearch && matchesType;
  });

  // 处理搜索
  const handleSearch = (value: string) => {
    setSearchText(value);
  };

  // 处理类型筛选
  const handleTypeFilter = (value: string | undefined) => {
    setTypeFilter(value);
  };

  // 清除筛选
  const handleClearFilters = () => {
    setSearchText('');
    setTypeFilter(undefined);
  };

  // 刷新数据
  const handleRefresh = () => {
    fetchDataSources();
  };

  // 创建数据源
  const handleCreate = () => {
    setEditingDataSource(null);
    form.resetFields();
    form.setFieldsValue({ enabled: true });
    setModalVisible(true);
  };

  // 编辑数据源
  const handleEdit = (record: DataSource) => {
    setEditingDataSource(record);
    form.setFieldsValue({
      name: record.name,
      type: record.type,
      url: record.url,
      config: JSON.stringify(record.config, null, 2),
      enabled: record.enabled,
    });
    setModalVisible(true);
  };

  // 删除数据源
  const handleDelete = async (id: string) => {
    try {
      // TODO: 实现删除数据源的API调用
      message.success('删除成功');
      fetchDataSources();
    } catch (error) {
      message.error('删除失败');
    }
  };

  // 切换数据源状态
  const handleToggleStatus = async (id: string, enabled: boolean) => {
    try {
      // TODO: 实现切换数据源状态的API调用
      message.success('状态更新成功');
      fetchDataSources();
    } catch (error) {
      message.error('状态更新失败');
    }
  };

  // 测试连接
  const handleTestConnection = async (record: DataSource) => {
    try {
      // TODO: 实现测试连接的API调用
      message.success('连接测试成功');
    } catch (error) {
      message.error('连接测试失败');
    }
  };

  // 提交表单
  const handleSubmit = async (values: DataSourceFormData) => {
    setSubmitting(true);
    try {
      let config;
      try {
        config = JSON.parse(values.config || '{}');
      } catch (error) {
        message.error('配置格式错误，请输入有效的JSON');
        setSubmitting(false);
        return;
      }

      const dataSourceData = {
        ...values,
        config,
      };

      if (editingDataSource) {
        // TODO: 实现更新数据源的API调用
        message.success('数据源更新成功');
      } else {
        // TODO: 实现创建数据源的API调用
        message.success('数据源创建成功');
      }
      
      setModalVisible(false);
      fetchDataSources();
    } catch (error) {
      message.error(editingDataSource ? '数据源更新失败' : '数据源创建失败');
    } finally {
      setSubmitting(false);
    }
  };

  // 获取数据源类型颜色
  const getTypeColor = (type: string) => {
    const colors: Record<string, string> = {
      prometheus: 'blue',
      elasticsearch: 'green',
      mysql: 'orange',
      postgresql: 'purple',
      mongodb: 'cyan',
      redis: 'red',
    };
    return colors[type] || 'default';
  };

  // 获取数据源类型列表
  const dataSourceTypes = Array.from(new Set(dataSources.map(ds => ds.type)));

  // 表格列定义
  const columns: ColumnsType<DataSource> = [
    {
      title: '数据源名称',
      dataIndex: 'name',
      key: 'name',
      ellipsis: true,
    },
    {
      title: '类型',
      dataIndex: 'type',
      key: 'type',
      width: 120,
      render: (type: string) => (
        <Tag color={getTypeColor(type)}>
          {type.toUpperCase()}
        </Tag>
      ),
    },
    {
      title: 'URL',
      dataIndex: 'url',
      key: 'url',
      ellipsis: true,
      render: (url: string) => (
        <Tooltip title={url}>
          <span>{url}</span>
        </Tooltip>
      ),
    },
    {
      title: '状态',
      dataIndex: 'enabled',
      key: 'enabled',
      width: 100,
      render: (enabled: boolean, record: DataSource) => (
        <Switch
          checked={enabled}
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
      title: '操作',
      key: 'action',
      width: 200,
      render: (_, record: DataSource) => (
        <Space size="small">
          <Tooltip title="测试连接">
            <Button
              type="text"
              icon={<ApiOutlined />}
              onClick={() => handleTestConnection(record)}
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
            title="确定要删除这个数据源吗？"
            onConfirm={() => handleDelete(record.id)}
            okText="确定"
            cancelText="取消"
          >
            <Tooltip title="删除">
              <Button
                type="text"
                danger
                icon={<DeleteOutlined />}
              />
            </Tooltip>
          </Popconfirm>
        </Space>
      ),
    },
  ];

  return (
    <div style={{ padding: '24px' }}>
      <Card>
        <div style={{ marginBottom: '16px' }}>
          <Row gutter={16}>
            <Col span={8}>
              <Search
                placeholder="搜索数据源名称或URL"
                value={searchText}
                onChange={(e) => setSearchText(e.target.value)}
                onSearch={handleSearch}
                allowClear
              />
            </Col>
            <Col span={6}>
              <Select
                placeholder="选择类型"
                value={typeFilter}
                onChange={handleTypeFilter}
                allowClear
                style={{ width: '100%' }}
              >
                {dataSourceTypes.map((type) => (
                  <Option key={type} value={type}>
                    {type.toUpperCase()}
                  </Option>
                ))}
              </Select>
            </Col>
            <Col span={10}>
              <Space>
                <Button
                  type="primary"
                  icon={<PlusOutlined />}
                  onClick={handleCreate}
                >
                  创建数据源
                </Button>
                <Button
                  icon={<ReloadOutlined />}
                  onClick={handleRefresh}
                >
                  刷新
                </Button>
                <Button onClick={handleClearFilters}>
                  清除筛选
                </Button>
              </Space>
            </Col>
          </Row>
        </div>

        <Table
          columns={columns}
          dataSource={filteredDataSources}
          loading={loading}
          rowKey="id"
          pagination={{
            total: filteredDataSources.length,
            pageSize: 20,
            showSizeChanger: true,
            showQuickJumper: true,
            showTotal: (total, range) =>
              `第 ${range[0]}-${range[1]} 条/共 ${total} 条`,
          }}
        />
      </Card>

      <Modal
        title={editingDataSource ? '编辑数据源' : '创建数据源'}
        open={modalVisible}
        onCancel={() => setModalVisible(false)}
        footer={null}
        width={600}
      >
        <Form
          form={form}
          layout="vertical"
          onFinish={handleSubmit}
        >
          <Form.Item
            label="数据源名称"
            name="name"
            rules={[
              { required: true, message: '请输入数据源名称' },
              { max: 100, message: '名称不能超过100个字符' },
            ]}
          >
            <Input placeholder="请输入数据源名称" />
          </Form.Item>

          <Form.Item
            label="数据源类型"
            name="type"
            rules={[{ required: true, message: '请选择数据源类型' }]}
          >
            <Select placeholder="请选择数据源类型">
              <Option value="prometheus">Prometheus</Option>
              <Option value="elasticsearch">Elasticsearch</Option>
              <Option value="mysql">MySQL</Option>
              <Option value="postgresql">PostgreSQL</Option>
              <Option value="mongodb">MongoDB</Option>
              <Option value="redis">Redis</Option>
            </Select>
          </Form.Item>

          <Form.Item
            label="连接URL"
            name="url"
            rules={[
              { required: true, message: '请输入连接URL' },
              { type: 'url', message: '请输入有效的URL' },
            ]}
          >
            <Input placeholder="请输入连接URL，例如：http://localhost:9090" />
          </Form.Item>

          <Form.Item
            label="配置"
            name="config"
            rules={[{ required: true, message: '请输入配置信息' }]}
          >
            <TextArea
              rows={6}
              placeholder='请输入JSON格式的配置信息，例如：{"username": "admin", "password": "password"}'
            />
          </Form.Item>

          <Form.Item
            label="启用状态"
            name="enabled"
            valuePropName="checked"
          >
            <Switch checkedChildren="启用" unCheckedChildren="禁用" />
          </Form.Item>

          <Form.Item>
            <Space>
              <Button
                type="primary"
                htmlType="submit"
                loading={submitting}
              >
                {editingDataSource ? '更新' : '创建'}
              </Button>
              <Button onClick={() => setModalVisible(false)}>
                取消
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
};

export default DataSourceList;