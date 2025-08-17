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
  Popconfirm,
  Tag,
  Avatar,
  Switch,
  Row,
  Col,
  Typography,
  Tooltip,
} from 'antd';
import {
  PlusOutlined,
  SearchOutlined,
  EditOutlined,
  DeleteOutlined,
  UserOutlined,
  ReloadOutlined,
  KeyOutlined,
  LockOutlined,
  UnlockOutlined,
} from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';
import { useUI } from '../../hooks';
import { User, UserRole, UserStatus } from '../../types';

const { Title } = Typography;
const { Option } = Select;
const { Search } = Input;

interface UserFormData {
  username: string;
  email: string;
  real_name: string;
  phone?: string;
  role: UserRole;
  status: UserStatus;
  password?: string;
  confirm_password?: string;
}

interface UserFilters {
  keyword: string;
  role?: UserRole;
  status?: UserStatus;
}

const UserManagement: React.FC = () => {
  const { setBreadcrumb } = useUI();
  const [form] = Form.useForm<UserFormData>();
  
  // 状态管理
  const [users, setUsers] = useState<User[]>([]);
  const [loading, setLoading] = useState(false);
  const [modalVisible, setModalVisible] = useState(false);
  const [editingUser, setEditingUser] = useState<User | null>(null);
  const [filters, setFilters] = useState<UserFilters>({ keyword: '' });
  const [pagination, setPagination] = useState({
    current: 1,
    pageSize: 10,
    total: 0,
  });

  useEffect(() => {
    setBreadcrumb([
      { title: '系统管理' },
      { title: '用户管理' },
    ]);
    fetchUsers();
  }, [setBreadcrumb]);

  // 获取用户列表
  const fetchUsers = async () => {
    setLoading(true);
    try {
      // 模拟API调用
      const mockUsers: User[] = [
        {
          id: '1',
          username: 'admin',
          email: 'admin@example.com',
          real_name: '系统管理员',
          phone: '13800138000',
          role: UserRole.ADMIN,
          status: UserStatus.ACTIVE,
          avatar: '',
          last_login_at: new Date().toISOString(),
          created_at: new Date().toISOString(),
          updated_at: new Date().toISOString(),
        },
        {
          id: '2',
          username: 'operator',
          email: 'operator@example.com',
          real_name: '运维工程师',
          phone: '13800138001',
          role: UserRole.OPERATOR,
          status: UserStatus.ACTIVE,
          avatar: '',
          last_login_at: new Date().toISOString(),
          created_at: new Date().toISOString(),
          updated_at: new Date().toISOString(),
        },
        {
          id: '3',
          username: 'viewer',
          email: 'viewer@example.com',
          real_name: '普通用户',
          phone: '13800138002',
          role: UserRole.VIEWER,
          status: UserStatus.INACTIVE,
          avatar: '',
          last_login_at: new Date().toISOString(),
          created_at: new Date().toISOString(),
          updated_at: new Date().toISOString(),
        },
      ];
      
      setUsers(mockUsers);
      setPagination(prev => ({ ...prev, total: mockUsers.length }));
    } catch (error) {
      message.error('获取用户列表失败');
    } finally {
      setLoading(false);
    }
  };

  // 角色标签颜色
  const getRoleColor = (role: UserRole) => {
    switch (role) {
      case UserRole.ADMIN:
        return 'red';
      case UserRole.OPERATOR:
        return 'blue';
      case UserRole.VIEWER:
        return 'green';
      default:
        return 'default';
    }
  };

  // 角色标签文本
  const getRoleText = (role: UserRole) => {
    switch (role) {
      case UserRole.ADMIN:
        return '管理员';
      case UserRole.OPERATOR:
        return '运维人员';
      case UserRole.VIEWER:
        return '普通用户';
      default:
        return '未知';
    }
  };

  // 状态标签颜色
  const getStatusColor = (status: UserStatus) => {
    switch (status) {
      case UserStatus.ACTIVE:
        return 'success';
      case UserStatus.INACTIVE:
        return 'default';
      case UserStatus.LOCKED:
        return 'error';
      default:
        return 'default';
    }
  };

  // 状态标签文本
  const getStatusText = (status: UserStatus) => {
    switch (status) {
      case UserStatus.ACTIVE:
        return '正常';
      case UserStatus.INACTIVE:
        return '禁用';
      case UserStatus.LOCKED:
        return '锁定';
      default:
        return '未知';
    }
  };

  // 表格列定义
  const columns: ColumnsType<User> = [
    {
      title: '用户',
      key: 'user',
      render: (_, record) => (
        <Space>
          <Avatar
            size={40}
            src={record.avatar}
            icon={<UserOutlined />}
          />
          <div>
            <div style={{ fontWeight: 500 }}>{record.real_name}</div>
            <div style={{ color: '#666', fontSize: 12 }}>{record.username}</div>
          </div>
        </Space>
      ),
    },
    {
      title: '邮箱',
      dataIndex: 'email',
      key: 'email',
    },
    {
      title: '手机号',
      dataIndex: 'phone',
      key: 'phone',
    },
    {
      title: '角色',
      dataIndex: 'role',
      key: 'role',
      render: (role: UserRole) => (
        <Tag color={getRoleColor(role)}>
          {getRoleText(role)}
        </Tag>
      ),
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (status: UserStatus) => (
        <Tag color={getStatusColor(status)}>
          {getStatusText(status)}
        </Tag>
      ),
    },
    {
      title: '最后登录',
      dataIndex: 'last_login_at',
      key: 'last_login_at',
      render: (date: string) => date ? new Date(date).toLocaleString() : '-',
    },
    {
      title: '操作',
      key: 'actions',
      render: (_, record) => (
        <Space>
          <Tooltip title="编辑">
            <Button
              type="text"
              size="small"
              icon={<EditOutlined />}
              onClick={() => handleEdit(record)}
            />
          </Tooltip>
          <Tooltip title="重置密码">
            <Button
              type="text"
              size="small"
              icon={<KeyOutlined />}
              onClick={() => handleResetPassword(record)}
            />
          </Tooltip>
          <Tooltip title={record.status === UserStatus.ACTIVE ? '禁用' : '启用'}>
            <Button
              type="text"
              size="small"
              icon={record.status === UserStatus.ACTIVE ? <LockOutlined /> : <UnlockOutlined />}
              onClick={() => handleToggleStatus(record)}
            />
          </Tooltip>
          <Popconfirm
            title="确定要删除这个用户吗？"
            onConfirm={() => handleDelete(record.id)}
            okText="删除"
            cancelText="取消"
          >
            <Tooltip title="删除">
              <Button
                type="text"
                size="small"
                icon={<DeleteOutlined />}
                danger
              />
            </Tooltip>
          </Popconfirm>
        </Space>
      ),
    },
  ];

  // 创建用户
  const handleCreate = () => {
    setEditingUser(null);
    form.resetFields();
    form.setFieldsValue({ status: UserStatus.ACTIVE, role: UserRole.VIEWER });
    setModalVisible(true);
  };

  // 编辑用户
  const handleEdit = (user: User) => {
    setEditingUser(user);
    form.setFieldsValue({
      username: user.username,
      email: user.email,
      real_name: user.real_name,
      phone: user.phone,
      role: user.role,
      status: user.status,
    });
    setModalVisible(true);
  };

  // 删除用户
  const handleDelete = async (id: string) => {
    try {
      message.success('删除成功');
      fetchUsers();
    } catch (error) {
      message.error('删除失败');
    }
  };

  // 切换用户状态
  const handleToggleStatus = async (user: User) => {
    try {
      const newStatus = user.status === UserStatus.ACTIVE ? UserStatus.INACTIVE : UserStatus.ACTIVE;
      message.success(`${newStatus === UserStatus.ACTIVE ? '启用' : '禁用'}成功`);
      fetchUsers();
    } catch (error) {
      message.error('操作失败');
    }
  };

  // 重置密码
  const handleResetPassword = async (user: User) => {
    Modal.confirm({
      title: '重置密码',
      content: `确定要重置用户 "${user.real_name}" 的密码吗？新密码将通过邮件发送给用户。`,
      okText: '确定',
      cancelText: '取消',
      onOk: async () => {
        try {
          message.success('密码重置成功，新密码已发送到用户邮箱');
        } catch (error) {
          message.error('密码重置失败');
        }
      },
    });
  };

  // 提交表单
  const handleSubmit = async (values: UserFormData) => {
    try {
      if (editingUser) {
        message.success('更新成功');
      } else {
        message.success('创建成功');
      }
      setModalVisible(false);
      fetchUsers();
    } catch (error) {
      message.error(editingUser ? '更新失败' : '创建失败');
    }
  };

  // 搜索
  const handleSearch = (value: string) => {
    setFilters(prev => ({ ...prev, keyword: value }));
    // 这里应该重新获取数据
  };

  // 筛选
  const handleFilter = (key: keyof UserFilters, value: any) => {
    setFilters(prev => ({ ...prev, [key]: value }));
    // 这里应该重新获取数据
  };

  return (
    <div className="user-management">
      <Card
        title={<Title level={4} style={{ margin: 0 }}>用户管理</Title>}
        extra={
          <Button
            type="primary"
            icon={<PlusOutlined />}
            onClick={handleCreate}
          >
            新建用户
          </Button>
        }
      >
        {/* 搜索和筛选 */}
        <Row gutter={16} style={{ marginBottom: 16 }}>
          <Col span={8}>
            <Search
              placeholder="搜索用户名、姓名、邮箱"
              allowClear
              onSearch={handleSearch}
              style={{ width: '100%' }}
            />
          </Col>
          <Col span={4}>
            <Select
              placeholder="角色"
              allowClear
              style={{ width: '100%' }}
              onChange={(value) => handleFilter('role', value)}
            >
              <Option value={UserRole.ADMIN}>管理员</Option>
              <Option value={UserRole.OPERATOR}>运维人员</Option>
              <Option value={UserRole.VIEWER}>普通用户</Option>
            </Select>
          </Col>
          <Col span={4}>
            <Select
              placeholder="状态"
              allowClear
              style={{ width: '100%' }}
              onChange={(value) => handleFilter('status', value)}
            >
              <Option value={UserStatus.ACTIVE}>正常</Option>
              <Option value={UserStatus.INACTIVE}>禁用</Option>
              <Option value={UserStatus.LOCKED}>锁定</Option>
            </Select>
          </Col>
          <Col span={4}>
            <Button
              icon={<ReloadOutlined />}
              onClick={fetchUsers}
            >
              刷新
            </Button>
          </Col>
        </Row>

        {/* 用户表格 */}
        <Table
          columns={columns}
          dataSource={users}
          rowKey="id"
          loading={loading}
          pagination={{
            ...pagination,
            showSizeChanger: true,
            showQuickJumper: true,
            showTotal: (total) => `共 ${total} 条记录`,
          }}
          onChange={(paginationInfo) => {
            setPagination({
              current: paginationInfo.current || 1,
              pageSize: paginationInfo.pageSize || 10,
              total: pagination.total,
            });
          }}
        />
      </Card>

      {/* 用户表单弹窗 */}
      <Modal
        title={editingUser ? '编辑用户' : '新建用户'}
        open={modalVisible}
        onCancel={() => setModalVisible(false)}
        onOk={() => form.submit()}
        okText={editingUser ? '更新' : '创建'}
        cancelText="取消"
        width={600}
      >
        <Form
          form={form}
          layout="vertical"
          onFinish={handleSubmit}
        >
          <Row gutter={16}>
            <Col span={12}>
              <Form.Item
                name="username"
                label="用户名"
                rules={[
                  { required: true, message: '请输入用户名' },
                  { min: 3, max: 20, message: '用户名长度为3-20个字符' },
                  { pattern: /^[a-zA-Z0-9_]+$/, message: '用户名只能包含字母、数字和下划线' },
                ]}
              >
                <Input placeholder="请输入用户名" disabled={!!editingUser} />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                name="real_name"
                label="真实姓名"
                rules={[
                  { required: true, message: '请输入真实姓名' },
                  { max: 20, message: '姓名不能超过20个字符' },
                ]}
              >
                <Input placeholder="请输入真实姓名" />
              </Form.Item>
            </Col>
          </Row>

          <Row gutter={16}>
            <Col span={12}>
              <Form.Item
                name="email"
                label="邮箱"
                rules={[
                  { required: true, message: '请输入邮箱' },
                  { type: 'email', message: '请输入有效的邮箱地址' },
                ]}
              >
                <Input placeholder="请输入邮箱" />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                name="phone"
                label="手机号"
                rules={[
                  { pattern: /^1[3-9]\d{9}$/, message: '请输入有效的手机号' },
                ]}
              >
                <Input placeholder="请输入手机号" />
              </Form.Item>
            </Col>
          </Row>

          <Row gutter={16}>
            <Col span={12}>
              <Form.Item
                name="role"
                label="角色"
                rules={[{ required: true, message: '请选择角色' }]}
              >
                <Select placeholder="请选择角色">
                  <Option value={UserRole.ADMIN}>管理员</Option>
                  <Option value={UserRole.OPERATOR}>运维人员</Option>
                  <Option value={UserRole.VIEWER}>普通用户</Option>
                </Select>
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                name="status"
                label="状态"
                rules={[{ required: true, message: '请选择状态' }]}
              >
                <Select placeholder="请选择状态">
                  <Option value={UserStatus.ACTIVE}>正常</Option>
                  <Option value={UserStatus.INACTIVE}>禁用</Option>
                  <Option value={UserStatus.LOCKED}>锁定</Option>
                </Select>
              </Form.Item>
            </Col>
          </Row>

          {!editingUser && (
            <>
              <Row gutter={16}>
                <Col span={12}>
                  <Form.Item
                    name="password"
                    label="密码"
                    rules={[
                      { required: true, message: '请输入密码' },
                      { min: 6, message: '密码至少6个字符' },
                    ]}
                  >
                    <Input.Password placeholder="请输入密码" />
                  </Form.Item>
                </Col>
                <Col span={12}>
                  <Form.Item
                    name="confirm_password"
                    label="确认密码"
                    dependencies={['password']}
                    rules={[
                      { required: true, message: '请确认密码' },
                      ({ getFieldValue }) => ({
                        validator(_, value) {
                          if (!value || getFieldValue('password') === value) {
                            return Promise.resolve();
                          }
                          return Promise.reject(new Error('两次输入的密码不一致'));
                        },
                      }),
                    ]}
                  >
                    <Input.Password placeholder="请再次输入密码" />
                  </Form.Item>
                </Col>
              </Row>
            </>
          )}
        </Form>
      </Modal>
    </div>
  );
};

export default UserManagement;