import React, { useState, useEffect } from 'react';
import {
  Card,
  Table,
  Button,
  Space,
  Input,
  Form,
  Modal,
  Tree,
  message,
  Popconfirm,
  Tag,
  Tooltip,
  Row,
  Col,
  Select
} from 'antd';
import {
  PlusOutlined,
  EditOutlined,
  DeleteOutlined,
  FolderOutlined,
  FileTextOutlined,
  SearchOutlined,
  ReloadOutlined
} from '@ant-design/icons';
import { useKnowledge } from '../../hooks/useKnowledge';
import { useUI } from '../../hooks/useUI';
import { knowledgeService } from '../../services/knowledge';
import type { KnowledgeCategory } from '../../types';
import type { ColumnsType } from 'antd/es/table';
import type { DataNode } from 'antd/es/tree';

interface CategoryFormData {
  name: string;
  description?: string;
  parent_id?: string;
}

const CategoryManagement: React.FC = () => {
  const { loading, setLoading } = useUI();
  const [form] = Form.useForm<CategoryFormData>();
  
  const [categories, setCategories] = useState<KnowledgeCategory[]>([]);
  const [treeData, setTreeData] = useState<DataNode[]>([]);
  const [isModalVisible, setIsModalVisible] = useState(false);
  const [editingCategory, setEditingCategory] = useState<KnowledgeCategory | null>(null);
  const [searchText, setSearchText] = useState('');
  const [expandedKeys, setExpandedKeys] = useState<React.Key[]>([]);
  const [selectedKeys, setSelectedKeys] = useState<React.Key[]>([]);
  const [viewMode, setViewMode] = useState<'table' | 'tree'>('table');

  // 加载分类列表
  const loadCategories = async () => {
    try {
      setLoading(true);
      const response = await knowledgeService.getCategories();
      setCategories(response);
      buildTreeData(response);
    } catch (error) {
      message.error('加载分类失败');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadCategories();
  }, []);

  // 构建树形数据
  const buildTreeData = (categories: KnowledgeCategory[]) => {
    const categoryMap = new Map<string, KnowledgeCategory>();
    categories.forEach(cat => categoryMap.set(cat.id, cat));

    const buildNode = (category: KnowledgeCategory): DataNode => {
      const children = categories
        .filter(cat => cat.parent_id === category.id)
        .map(buildNode);

      return {
        key: category.id,
        title: (
          <Space>
            <FolderOutlined />
            <span>{category.name}</span>
            <Tag color="blue">{category.document_count || 0}</Tag>
          </Space>
        ),
        children: children.length > 0 ? children : undefined
      };
    };

    const rootCategories = categories.filter(cat => !cat.parent_id);
    const treeData = rootCategories.map(buildNode);
    setTreeData(treeData);

    // 默认展开所有节点
    const allKeys = categories.map(cat => cat.id);
    setExpandedKeys(allKeys);
  };

  // 打开创建/编辑模态框
  const openModal = (category?: KnowledgeCategory) => {
    setEditingCategory(category || null);
    setIsModalVisible(true);
    
    if (category) {
      form.setFieldsValue({
        name: category.name,
        description: category.description,
        parent_id: category.parent_id
      });
    } else {
      form.resetFields();
    }
  };

  // 关闭模态框
  const closeModal = () => {
    setIsModalVisible(false);
    setEditingCategory(null);
    form.resetFields();
  };

  // 处理表单提交
  const handleSubmit = async (values: CategoryFormData) => {
    try {
      if (editingCategory) {
        await knowledgeService.updateCategory(editingCategory.id, values);
        message.success('分类更新成功');
      } else {
        await knowledgeService.createCategory(values);
        message.success('分类创建成功');
      }
      
      closeModal();
      loadCategories();
    } catch (error) {
      message.error(editingCategory ? '分类更新失败' : '分类创建失败');
    }
  };

  // 删除分类
  const handleDelete = async (id: string) => {
    try {
      await knowledgeService.deleteCategory(id);
      message.success('分类删除成功');
      loadCategories();
    } catch (error) {
      message.error('分类删除失败');
    }
  };

  // 获取分类路径
  const getCategoryPath = (categoryId: string): string => {
    const category = categories.find(cat => cat.id === categoryId);
    if (!category) return '';
    
    if (category.parent_id) {
      const parentPath = getCategoryPath(category.parent_id);
      return parentPath ? `${parentPath} / ${category.name}` : category.name;
    }
    
    return category.name;
  };

  // 过滤分类
  const filteredCategories = categories.filter(category =>
    category.name.toLowerCase().includes(searchText.toLowerCase()) ||
    (category.description && category.description.toLowerCase().includes(searchText.toLowerCase()))
  );

  const columns: ColumnsType<KnowledgeCategory> = [
    {
      title: '分类名称',
      dataIndex: 'name',
      key: 'name',
      render: (name: string, record: KnowledgeCategory) => (
        <Space>
          <FolderOutlined />
          <span>{name}</span>
        </Space>
      )
    },
    {
      title: '路径',
      key: 'path',
      render: (_, record: KnowledgeCategory) => (
        <span style={{ color: '#666' }}>
          {getCategoryPath(record.id)}
        </span>
      )
    },
    {
      title: '描述',
      dataIndex: 'description',
      key: 'description',
      ellipsis: true,
      render: (description: string) => (
        description ? (
          <Tooltip title={description}>
            <span>{description}</span>
          </Tooltip>
        ) : (
          <span style={{ color: '#999' }}>无描述</span>
        )
      )
    },
    {
      title: '文档数量',
      dataIndex: 'document_count',
      key: 'document_count',
      width: 100,
      render: (count: number) => (
        <Tag color="blue">{count || 0}</Tag>
      )
    },
    {
      title: '创建时间',
      dataIndex: 'created_at',
      key: 'created_at',
      width: 180,
      render: (date: string) => (
        new Date(date).toLocaleString()
      )
    },
    {
      title: '操作',
      key: 'actions',
      width: 150,
      render: (_, record: KnowledgeCategory) => (
        <Space>
          <Tooltip title="编辑">
            <Button
              type="text"
              icon={<EditOutlined />}
              onClick={() => openModal(record)}
            />
          </Tooltip>
          <Tooltip title="添加子分类">
            <Button
              type="text"
              icon={<PlusOutlined />}
              onClick={() => {
                form.setFieldValue('parent_id', record.id);
                openModal();
              }}
            />
          </Tooltip>
          <Tooltip title="删除">
            <Popconfirm
              title="确定要删除这个分类吗？"
              description="删除分类会同时删除其下的所有子分类和文档"
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
        </Space>
      )
    }
  ];

  const renderTreeView = () => (
    <Card>
      <Tree
        treeData={treeData}
        expandedKeys={expandedKeys}
        selectedKeys={selectedKeys}
        onExpand={setExpandedKeys}
        onSelect={setSelectedKeys}
        showLine
        showIcon
        titleRender={(nodeData) => (
          <div
            style={{
              display: 'flex',
              justifyContent: 'space-between',
              alignItems: 'center',
              width: '100%'
            }}
          >
            <span>{nodeData.title}</span>
            <Space>
              <Button
                type="text"
                size="small"
                icon={<EditOutlined />}
                onClick={(e) => {
                  e.stopPropagation();
                  const category = categories.find(cat => cat.id === nodeData.key);
                  if (category) openModal(category);
                }}
              />
              <Button
                type="text"
                size="small"
                icon={<PlusOutlined />}
                onClick={(e) => {
                  e.stopPropagation();
                  form.setFieldValue('parent_id', nodeData.key);
                  openModal();
                }}
              />
              <Popconfirm
                title="确定要删除这个分类吗？"
                onConfirm={(e) => {
                  e?.stopPropagation();
                  handleDelete(nodeData.key as string);
                }}
                onClick={(e) => e?.stopPropagation()}
              >
                <Button
                  type="text"
                  size="small"
                  danger
                  icon={<DeleteOutlined />}
                  onClick={(e) => e.stopPropagation()}
                />
              </Popconfirm>
            </Space>
          </div>
        )}
      />
    </Card>
  );

  const renderTableView = () => (
    <Card>
      <Table
        columns={columns}
        dataSource={filteredCategories}
        rowKey="id"
        loading={loading}
        pagination={{
          showSizeChanger: true,
          showQuickJumper: true,
          showTotal: (total, range) =>
            `第 ${range[0]}-${range[1]} 条，共 ${total} 条`
        }}
      />
    </Card>
  );

  return (
    <div style={{ padding: '24px' }}>
      <Row gutter={16} style={{ marginBottom: '16px' }}>
        <Col span={12}>
          <Input.Search
            placeholder="搜索分类名称或描述"
            allowClear
            value={searchText}
            onChange={(e) => setSearchText(e.target.value)}
            style={{ width: '100%' }}
          />
        </Col>
        <Col span={12}>
          <Space style={{ float: 'right' }}>
            <Button.Group>
              <Button
                type={viewMode === 'table' ? 'primary' : 'default'}
                onClick={() => setViewMode('table')}
              >
                表格视图
              </Button>
              <Button
                type={viewMode === 'tree' ? 'primary' : 'default'}
                onClick={() => setViewMode('tree')}
              >
                树形视图
              </Button>
            </Button.Group>
            <Button
              icon={<ReloadOutlined />}
              onClick={loadCategories}
            >
              刷新
            </Button>
            <Button
              type="primary"
              icon={<PlusOutlined />}
              onClick={() => openModal()}
            >
              创建分类
            </Button>
          </Space>
        </Col>
      </Row>

      {viewMode === 'table' ? renderTableView() : renderTreeView()}

      <Modal
        title={editingCategory ? '编辑分类' : '创建分类'}
        open={isModalVisible}
        onCancel={closeModal}
        footer={null}
        width={600}
      >
        <Form
          form={form}
          layout="vertical"
          onFinish={handleSubmit}
        >
          <Form.Item
            name="name"
            label="分类名称"
            rules={[
              { required: true, message: '请输入分类名称' },
              { max: 50, message: '分类名称不能超过50个字符' }
            ]}
          >
            <Input placeholder="请输入分类名称" />
          </Form.Item>

          <Form.Item
            name="parent_id"
            label="父分类"
          >
            <Select
              placeholder="请选择父分类（可选）"
              allowClear
            >
              {categories
                .filter(cat => !editingCategory || cat.id !== editingCategory.id)
                .map(category => (
                  <Select.Option key={category.id} value={category.id}>
                    {getCategoryPath(category.id)}
                  </Select.Option>
                ))
              }
            </Select>
          </Form.Item>

          <Form.Item
            name="description"
            label="分类描述"
          >
            <Input.TextArea
              rows={3}
              placeholder="请输入分类描述（可选）"
              maxLength={200}
            />
          </Form.Item>

          <Form.Item style={{ marginBottom: 0, textAlign: 'right' }}>
            <Space>
              <Button onClick={closeModal}>
                取消
              </Button>
              <Button type="primary" htmlType="submit">
                {editingCategory ? '更新' : '创建'}
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
};

export default CategoryManagement;