import React, { useEffect, useState } from 'react';
import {
  Card,
  Table,
  Button,
  Space,
  Input,
  Select,
  Modal,
  Form,
  message,
  Tag,
  Tooltip,
  Popconfirm,
  Typography,
  Row,
  Col,
  Alert,
  Divider,
} from 'antd';
import {
  PlusOutlined,
  EditOutlined,
  DeleteOutlined,
  ReloadOutlined,
  SearchOutlined,
  LinkOutlined,
  DisconnectOutlined,
  SettingOutlined,
} from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';
import { useUI } from '../../hooks';
import { DataSource } from '../../types';

const { Title, Text } = Typography;
const { Option } = Select;
const { TextArea } = Input;

interface DataSourceFormData {
  name: string;
  type: string;
  url: string;
  username?: string;
  password?: string;
  database?: string;
  timeout?: number;
  max_connections?: number;
  ssl_enabled?: boolean;
  description?: string;
}

const DataSourceManagement: React.FC = () => {
  const { setBreadcrumbs } = useUI();
  const [dataSources, setDataSources] = useState<DataSource[]>([]);
  const [loading, setLoading] = useState(false);
  const [searchText, setSearchText] = useState('');
  const [statusFilter, setStatusFilter] = useState<string>('');
  const [typeFilter, setTypeFilter] = useState<string>('');
  const [modalVisible, setModalVisible] = useState(false);
  const [editingDataSource, setEditingDataSource] = useState<DataSource | null>(null);
  const [testingConnection, setTestingConnection] = useState<string | null>(null);
  const [form] = Form.useForm<DataSourceFormData>();

  useEffect(() => {
    setBreadcrumbs([
      { title: '规则管理', path: '/rules' },
      { title: '数据源管理' },
    ]);
    fetchDataSources();
  }, [setBreadcrumbs]);

  // 获取数据源列表
  const fetchDataSources = async () => {
    setLoading(true);
    try {
      // 模拟API调用
      const mockDataSources: DataSource[] = [
        {
          id: '1',
          name: 'Prometheus Main',
          type: 'prometheus',
          url: 'http://prometheus:9090',
          status: 'active',
          description: '主要的 Prometheus 监控数据源',
          created_at: '2024-01-01T00:00:00Z',
          updated_at: '2024-01-15T10:30:00Z',
        },
        {
          id: '2',
          name: 'InfluxDB Metrics',
          type: 'influxdb',
          url: 'http://influxdb:8086',
          status: 'active',
          description: '时序数据库，存储应用指标',
          created_at: '2024-01-02T00:00:00Z',
          updated_at: '2024-01-15T09:15:00Z',
        },
        {
          id: '3',
          name: 'Elasticsearch Logs',
          type: 'elasticsearch',
          url: 'http://elasticsearch:9200',
          status: 'inactive',
          description: '日志搜索和分析',
          created_at: '2024-01-03T00:00:00Z',
          updated_at: '2024-01-10T14:20:00Z',
        },
        {
          id: '4',
          name: 'MySQL Database',
          type: 'mysql',
          url: 'mysql://mysql:3306/monitoring',
          status: 'error',
          description: '业务数据库监控',
          created_at: '2024-01-04T00:00:00Z',
          updated_at: '2024-01-14T16:45:00Z',
        },
      ];
      setDataSources(mockDataSources);
    } catch (error) {
      message.error('获取数据源列表失败');
    } finally {
      setLoading(false);
    }
  };

  // 测试连接
  const handleTestConnection = async (dataSource: DataSource) => {
    setTestingConnection(dataSource.id);
    try {
      // 模拟连接测试
      await new Promise(resolve => setTimeout(resolve, 2000));
      
      const success = Math.random() > 0.3; // 70% 成功率
      if (success) {
        message.success(`数据源 "${dataSource.name}" 连接测试成功`);
        // 更新状态
        setDataSources(prev => prev.map(ds => 
          ds.id === dataSource.id ? { ...ds, status: 'active' } : ds
        ));
      } else {
        message.error(`数据源 "${dataSource.name}" 连接测试失败`);
        setDataSources(prev => prev.map(ds => 
          ds.id === dataSource.id ? { ...ds, status: 'error' } : ds
        ));
      }
    } catch (error) {
      message.error('连接测试失败');
    } finally {
      setTestingConnection(null);
    }
  };

  // 删除数据源
  const handleDelete = async (id: string) => {
    try {
      setDataSources(prev => prev.filter(ds => ds.id !== id));
      message.success('数据源删除成功');
    } catch (error) {
      message.error('数据源删除失败');
    }
  };

  // 打开编辑模态框
  const handleEdit = (dataSource: DataSource) => {
    setEditingDataSource(dataSource);
    form.setFieldsValue({
      name: dataSource.name,
      type: dataSource.type,
      url: dataSource.url,
      description: dataSource.description,
    });
    setModalVisible(true);
  };

  // 保存数据源
  const handleSave = async (values: DataSourceFormData) => {
    try {
      if (editingDataSource) {
        // 更新数据源
        setDataSources(prev => prev.map(ds => 
          ds.id === editingDataSource.id 
            ? { ...ds, ...values, updated_at: new Date().toISOString() }
            : ds
        ));
        message.success('数据源更新成功');
      } else {
        // 创建数据源
        const newDataSource: DataSource = {
          id: Date.now().toString(),
          ...values,
          status: 'inactive',
          created_at: new Date().toISOString(),
          updated_at: new Date().toISOString(),
        };
        setDataSources(prev => [newDataSource, ...prev]);
        message.success('数据源创建成功');
      }
      
      setModalVisible(false);
      setEditingDataSource(null);
      form.resetFields();
    } catch (error) {
      message.error(editingDataSource ? '数据源更新失败' : '数据源创建失败');
    }
  };

  // 关闭模态框
  const handleCancel = () => {
    setModalVisible(false);
    setEditingDataSource(null);
    form.resetFields();
  };

  // 获取状态标签
  const getStatusTag = (status: string) => {
    const statusConfig = {
      active: { color: 'green', text: '正常' },
      inactive: { color: 'orange', text: '未激活' },
      error: { color: 'red', text: '错误' },
    };
    const config = statusConfig[status as keyof typeof statusConfig] || { color: 'default', text: '未知' };
    return <Tag color={config.color}>{config.text}</Tag>;
  };

  // 获取类型标签
  const getTypeTag = (type: string) => {
    const typeConfig = {
      prometheus: { color: 'blue', text: 'Prometheus' },
      influxdb: { color: 'purple', text: 'InfluxDB' },
      elasticsearch: { color: 'cyan', text: 'Elasticsearch' },
      mysql: { color: 'orange', text: 'MySQL' },
      postgresql: { color: 'geekblue', text: 'PostgreSQL' },
      redis: { color: 'red', text: 'Redis' },
    };
    const config = typeConfig[type as keyof typeof typeConfig] || { color: 'default', text: type };
    return <Tag color={config.color}>{config.text}</Tag>;
  };

  // 过滤数据
  const filteredDataSources = dataSources.filter(ds => {
    const matchesSearch = !searchText || 
      ds.name.toLowerCase().includes(searchText.toLowerCase()) ||
      ds.url.toLowerCase().includes(searchText.toLowerCase()) ||
      (ds.description && ds.description.toLowerCase().includes(searchText.toLowerCase()));
    
    const matchesStatus = !statusFilter || ds.status === statusFilter;
    const matchesType = !typeFilter || ds.type === typeFilter;
    
    return matchesSearch && matchesStatus && matchesType;
  });

  const columns: ColumnsType<DataSource> = [
    {
      title: '名称',
      dataIndex: 'name',
      key: 'name',
      render: (text, record) => (
        <div>
          <div style={{ fontWeight: 500 }}>{text}</div>
          {record.description && (
            <div style={{ fontSize: '12px', color: '#666', marginTop: 2 }}>
              {record.description}
            </div>
          )}
        </div>
      ),
    },
    {
      title: '类型',
      dataIndex: 'type',
      key: 'type',
      width: 120,
      render: (type) => getTypeTag(type),
    },
    {
      title: 'URL',
      dataIndex: 'url',
      key: 'url',
      render: (url) => (
        <Tooltip title={url}>
          <Text code style={{ maxWidth: 200 }} ellipsis>
            {url}
          </Text>
        </Tooltip>
      ),
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 100,
      render: (status) => getStatusTag(status),
    },
    {
      title: '更新时间',
      dataIndex: 'updated_at',
      key: 'updated_at',
      width: 180,
      render: (time) => new Date(time).toLocaleString(),
    },
    {
      title: '操作',
      key: 'actions',
      width: 200,
      render: (_, record) => (
        <Space>
          <Tooltip title="测试连接">
            <Button
              type="text"
              icon={<LinkOutlined />}
              loading={testingConnection === record.id}
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
            description="删除后将无法恢复，且可能影响相关规则。"
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
    <div className="datasource-management">
      <Card
        title={
          <Space>
            <SettingOutlined />
            <Title level={4} style={{ margin: 0 }}>数据源管理</Title>
          </Space>
        }
        extra={
          <Space>
            <Button
              icon={<ReloadOutlined />}
              onClick={fetchDataSources}
              loading={loading}
            >
              刷新
            </Button>
            <Button
              type="primary"
              icon={<PlusOutlined />}
              onClick={() => setModalVisible(true)}
            >
              新建数据源
            </Button>
          </Space>
        }
      >
        {/* 搜索和筛选 */}
        <Row gutter={16} style={{ marginBottom: 16 }}>
          <Col span={8}>
            <Input
              placeholder="搜索数据源名称、URL或描述"
              prefix={<SearchOutlined />}
              value={searchText}
              onChange={(e) => setSearchText(e.target.value)}
              allowClear
            />
          </Col>
          <Col span={4}>
            <Select
              placeholder="状态筛选"
              value={statusFilter}
              onChange={setStatusFilter}
              allowClear
              style={{ width: '100%' }}
            >
              <Option value="active">正常</Option>
              <Option value="inactive">未激活</Option>
              <Option value="error">错误</Option>
            </Select>
          </Col>
          <Col span={4}>
            <Select
              placeholder="类型筛选"
              value={typeFilter}
              onChange={setTypeFilter}
              allowClear
              style={{ width: '100%' }}
            >
              <Option value="prometheus">Prometheus</Option>
              <Option value="influxdb">InfluxDB</Option>
              <Option value="elasticsearch">Elasticsearch</Option>
              <Option value="mysql">MySQL</Option>
              <Option value="postgresql">PostgreSQL</Option>
              <Option value="redis">Redis</Option>
            </Select>
          </Col>
        </Row>

        {/* 统计信息 */}
        <Row gutter={16} style={{ marginBottom: 16 }}>
          <Col span={6}>
            <Card size="small">
              <div style={{ textAlign: 'center' }}>
                <div style={{ fontSize: '24px', fontWeight: 'bold', color: '#1890ff' }}>
                  {dataSources.length}
                </div>
                <div>总数据源</div>
              </div>
            </Card>
          </Col>
          <Col span={6}>
            <Card size="small">
              <div style={{ textAlign: 'center' }}>
                <div style={{ fontSize: '24px', fontWeight: 'bold', color: '#52c41a' }}>
                  {dataSources.filter(ds => ds.status === 'active').length}
                </div>
                <div>正常运行</div>
              </div>
            </Card>
          </Col>
          <Col span={6}>
            <Card size="small">
              <div style={{ textAlign: 'center' }}>
                <div style={{ fontSize: '24px', fontWeight: 'bold', color: '#faad14' }}>
                  {dataSources.filter(ds => ds.status === 'inactive').length}
                </div>
                <div>未激活</div>
              </div>
            </Card>
          </Col>
          <Col span={6}>
            <Card size="small">
              <div style={{ textAlign: 'center' }}>
                <div style={{ fontSize: '24px', fontWeight: 'bold', color: '#ff4d4f' }}>
                  {dataSources.filter(ds => ds.status === 'error').length}
                </div>
                <div>连接错误</div>
              </div>
            </Card>
          </Col>
        </Row>

        {/* 数据源表格 */}
        <Table
          columns={columns}
          dataSource={filteredDataSources}
          rowKey="id"
          loading={loading}
          pagination={{
            total: filteredDataSources.length,
            pageSize: 10,
            showSizeChanger: true,
            showQuickJumper: true,
            showTotal: (total) => `共 ${total} 条记录`,
          }}
        />
      </Card>

      {/* 创建/编辑数据源模态框 */}
      <Modal
        title={editingDataSource ? '编辑数据源' : '新建数据源'}
        open={modalVisible}
        onCancel={handleCancel}
        footer={null}
        width={600}
      >
        <Form
          form={form}
          layout="vertical"
          onFinish={handleSave}
          initialValues={{
            timeout: 30,
            max_connections: 10,
            ssl_enabled: false,
          }}
        >
          <Row gutter={16}>
            <Col span={12}>
              <Form.Item
                name="name"
                label="数据源名称"
                rules={[{ required: true, message: '请输入数据源名称' }]}
              >
                <Input placeholder="请输入数据源名称" />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                name="type"
                label="数据源类型"
                rules={[{ required: true, message: '请选择数据源类型' }]}
              >
                <Select placeholder="请选择数据源类型">
                  <Option value="prometheus">Prometheus</Option>
                  <Option value="influxdb">InfluxDB</Option>
                  <Option value="elasticsearch">Elasticsearch</Option>
                  <Option value="mysql">MySQL</Option>
                  <Option value="postgresql">PostgreSQL</Option>
                  <Option value="redis">Redis</Option>
                </Select>
              </Form.Item>
            </Col>
          </Row>

          <Form.Item
            name="url"
            label="连接URL"
            rules={[
              { required: true, message: '请输入连接URL' },
              { type: 'url', message: '请输入有效的URL' },
            ]}
          >
            <Input placeholder="例如: http://prometheus:9090" />
          </Form.Item>

          <Form.Item
            name="description"
            label="描述"
          >
            <TextArea
              placeholder="请输入数据源描述"
              rows={3}
            />
          </Form.Item>

          <Divider />

          <Row gutter={16}>
            <Col span={12}>
              <Form.Item
                name="username"
                label="用户名"
              >
                <Input placeholder="用户名（可选）" />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                name="password"
                label="密码"
              >
                <Input.Password placeholder="密码（可选）" />
              </Form.Item>
            </Col>
          </Row>

          <Row gutter={16}>
            <Col span={12}>
              <Form.Item
                name="timeout"
                label="超时时间（秒）"
              >
                <Input placeholder="30" />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                name="max_connections"
                label="最大连接数"
              >
                <Input placeholder="10" />
              </Form.Item>
            </Col>
          </Row>

          <div style={{ textAlign: 'right' }}>
            <Space>
              <Button onClick={handleCancel}>
                取消
              </Button>
              <Button type="primary" htmlType="submit">
                {editingDataSource ? '更新' : '创建'}
              </Button>
            </Space>
          </div>
        </Form>
      </Modal>
    </div>
  );
};

export default DataSourceManagement;