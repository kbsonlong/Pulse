import React, { useEffect, useState } from 'react';
import {
  Card,
  Tree,
  Button,
  Space,
  Modal,
  Form,
  Input,
  Select,
  message,
  Popconfirm,
  Row,
  Col,
  Typography,
} from 'antd';
import {
  PlusOutlined,
  EditOutlined,
  DeleteOutlined,
  FolderOutlined,
  FolderOpenOutlined,
  ArrowLeftOutlined,
} from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import { useKnowledge, useUI } from '../../hooks';
import { KnowledgeCategory } from '../../types';
import type { DataNode, TreeProps } from 'antd/es/tree';

const { Title } = Typography;
const { TextArea } = Input;
const { Option } = Select;

interface CategoryFormData {
  name: string;
  description?: string;
  parent_id?: string;
  sort_order: number;
}

const CategoryManagement: React.FC = () => {
  const navigate = useNavigate();
  const { setBreadcrumb } = useUI();
  const {
    categories,
    loading,
    fetchCategories,
  } = useKnowledge();

  const [form] = Form.useForm<CategoryFormData>();
  const [modalVisible, setModalVisible] = useState(false);
  const [editingCategory, setEditingCategory] = useState<KnowledgeCategory | null>(null);
  const [expandedKeys, setExpandedKeys] = useState<React.Key[]>([]);
  const [selectedKeys, setSelectedKeys] = useState<React.Key[]>([]);
  const [autoExpandParent, setAutoExpandParent] = useState(true);

  useEffect(() => {
    setBreadcrumb([
      { title: '知识库管理' },
      { title: '文档列表', path: '/knowledge' },
      { title: '分类管理' },
    ]);
    fetchCategories();
  }, [setBreadcrumb, fetchCategories]);

  // 构建分类树数据
  const buildCategoryTree = (categories: KnowledgeCategory[]): DataNode[] => {
    const categoryMap = new Map<string, KnowledgeCategory & { children: KnowledgeCategory[] }>();
    
    // 初始化所有分类
    categories.forEach(category => {
      categoryMap.set(category.id, { ...category, children: [] });
    });
    
    // 构建树结构
    const rootCategories: (KnowledgeCategory & { children: KnowledgeCategory[] })[] = [];
    
    categories.forEach(category => {
      const categoryWithChildren = categoryMap.get(category.id)!;
      if (category.parent_id) {
        const parent = categoryMap.get(category.parent_id);
        if (parent) {
          parent.children.push(categoryWithChildren);
        }
      } else {
        rootCategories.push(categoryWithChildren);
      }
    });
    
    // 转换为 Tree 组件需要的格式
    const convertToTreeData = (cats: (KnowledgeCategory & { children: KnowledgeCategory[] })[]): DataNode[] => {
      return cats.map(cat => ({
        key: cat.id,
        title: (
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
            <span>{cat.name}</span>
            <Space size="small">
              <Button
                type="text"
                size="small"
                icon={<EditOutlined />}
                onClick={(e) => {
                  e.stopPropagation();
                  handleEdit(cat);
                }}
              />
              <Popconfirm
                title="确定要删除这个分类吗？"
                description="删除分类会同时删除其下的所有子分类"
                onConfirm={(e) => {
                  e?.stopPropagation();
                  handleDelete(cat.id);
                }}
                okText="删除"
                cancelText="取消"
              >
                <Button
                  type="text"
                  size="small"
                  icon={<DeleteOutlined />}
                  danger
                  onClick={(e) => e.stopPropagation()}
                />
              </Popconfirm>
            </Space>
          </div>
        ),
        icon: ({ expanded }: { expanded: boolean }) => 
          expanded ? <FolderOpenOutlined /> : <FolderOutlined />,
        children: cat.children.length > 0 ? convertToTreeData(cat.children) : undefined,
      }));
    };
    
    return convertToTreeData(rootCategories);
  };

  // 获取可选的父分类（排除自己和子分类）
  const getAvailableParents = (excludeId?: string): KnowledgeCategory[] => {
    if (!excludeId) return categories as KnowledgeCategory[];
    
    const getDescendants = (parentId: string): string[] => {
      const descendants: string[] = [parentId];
      const children = (categories as KnowledgeCategory[]).filter(cat => cat.parent_id === parentId);
      children.forEach(child => {
        descendants.push(...getDescendants(child.id));
      });
      return descendants;
    };
    
    const excludeIds = getDescendants(excludeId);
    return (categories as KnowledgeCategory[]).filter(cat => !excludeIds.includes(cat.id));
  };

  // 处理树节点展开
  const onExpand: TreeProps['onExpand'] = (expandedKeysValue) => {
    setExpandedKeys(expandedKeysValue);
    setAutoExpandParent(false);
  };

  // 处理树节点选择
  const onSelect: TreeProps['onSelect'] = (selectedKeysValue) => {
    setSelectedKeys(selectedKeysValue);
  };

  // 创建分类
  const handleCreate = () => {
    setEditingCategory(null);
    form.resetFields();
    form.setFieldsValue({ sort_order: 0 });
    setModalVisible(true);
  };

  // 编辑分类
  const handleEdit = (category: KnowledgeCategory) => {
    setEditingCategory(category);
    form.setFieldsValue({
      name: category.name,
      description: category.description,
      parent_id: category.parent_id,
      sort_order: category.sort_order,
    });
    setModalVisible(true);
  };

  // 删除分类
  const handleDelete = async (id: string) => {
    try {
      // 这里应该调用删除分类的API
      message.success('删除成功');
      fetchCategories();
    } catch (error) {
      message.error('删除失败');
    }
  };

  // 提交表单
  const handleSubmit = async (values: CategoryFormData) => {
    try {
      if (editingCategory) {
        // 更新分类
        message.success('更新成功');
      } else {
        // 创建分类
        message.success('创建成功');
      }
      setModalVisible(false);
      fetchCategories();
    } catch (error) {
      message.error(editingCategory ? '更新失败' : '创建失败');
    }
  };

  // 返回列表
  const handleBack = () => {
    navigate('/knowledge');
  };

  return (
    <div className="category-management">
      <Card
        title={
          <Space>
            <Button
              icon={<ArrowLeftOutlined />}
              onClick={handleBack}
            >
              返回
            </Button>
            <Title level={4} style={{ margin: 0 }}>分类管理</Title>
          </Space>
        }
        extra={
          <Button
            type="primary"
            icon={<PlusOutlined />}
            onClick={handleCreate}
          >
            新建分类
          </Button>
        }
      >
        <Row gutter={24}>
          <Col span={24}>
            {categories && categories.length > 0 ? (
              <Tree
                showIcon
                expandedKeys={expandedKeys}
                autoExpandParent={autoExpandParent}
                selectedKeys={selectedKeys}
                onExpand={onExpand}
                onSelect={onSelect}
                treeData={buildCategoryTree(categories as KnowledgeCategory[])}
                style={{ background: '#fafafa', padding: 16, borderRadius: 6 }}
              />
            ) : (
              <div style={{ textAlign: 'center', padding: 60, color: '#999' }}>
                <FolderOutlined style={{ fontSize: 48, marginBottom: 16 }} />
                <div>暂无分类，点击上方按钮创建第一个分类</div>
              </div>
            )}
          </Col>
        </Row>
      </Card>

      {/* 分类表单弹窗 */}
      <Modal
        title={editingCategory ? '编辑分类' : '新建分类'}
        open={modalVisible}
        onCancel={() => setModalVisible(false)}
        onOk={() => form.submit()}
        okText={editingCategory ? '更新' : '创建'}
        cancelText="取消"
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
              { max: 50, message: '分类名称不能超过50个字符' },
            ]}
          >
            <Input placeholder="请输入分类名称" />
          </Form.Item>

          <Form.Item
            name="description"
            label="分类描述"
            rules={[
              { max: 200, message: '分类描述不能超过200个字符' },
            ]}
          >
            <TextArea
              placeholder="请输入分类描述（可选）"
              rows={3}
            />
          </Form.Item>

          <Form.Item
            name="parent_id"
            label="父分类"
          >
            <Select
              placeholder="请选择父分类（可选）"
              allowClear
            >
              {getAvailableParents(editingCategory?.id).map(category => (
                <Option key={category.id} value={category.id}>
                  {category.name}
                </Option>
              ))}
            </Select>
          </Form.Item>

          <Form.Item
            name="sort_order"
            label="排序"
            rules={[
              { required: true, message: '请输入排序值' },
              { type: 'number', min: 0, message: '排序值不能小于0' },
            ]}
          >
            <Input
              type="number"
              placeholder="请输入排序值（数字越小越靠前）"
            />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
};

export default CategoryManagement;