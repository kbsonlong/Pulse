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
  Popconfirm,
  Tree,
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
  FolderOutlined,
  FileTextOutlined,
} from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import { useKnowledge, useUI } from '../../hooks';
import { KnowledgeBase, KnowledgeCategory } from '../../types';
import { formatDate } from '../../utils';
import type { ColumnsType } from 'antd/es/table';
import type { DataNode } from 'antd/es/tree';

const { Search } = Input;
const { Option } = Select;

const KnowledgeList: React.FC = () => {
  const navigate = useNavigate();
  const { setBreadcrumbs } = useUI();
  const {
    knowledgeList,
    categories,
    total,
    page,
    limit,
    loading,
    filters,
    fetchKnowledgeList,
    fetchCategories,
    deleteKnowledge,
    setFilters,
    clearFilters,
    setPage,
    setLimit,
  } = useKnowledge();

  const [searchText, setSearchText] = useState('');
  const [selectedCategory, setSelectedCategory] = useState<string | undefined>();

  useEffect(() => {
    setBreadcrumbs([
      { title: '知识库管理' },
      { title: '文档列表' },
    ]);
    fetchKnowledgeList();
    fetchCategories();
  }, [setBreadcrumbs, fetchKnowledgeList, fetchCategories]);

  // 处理搜索
  const handleSearch = (value: string) => {
    setSearchText(value);
    setFilters({ ...filters, search: value });
    fetchKnowledgeList({ ...filters, search: value, page: 1 });
  };

  // 处理分类筛选
  const handleCategoryChange = (categoryId: string | undefined) => {
    setSelectedCategory(categoryId);
    const newFilters = { ...filters, category_id: categoryId };
    setFilters(newFilters);
    fetchKnowledgeList({ ...newFilters, page: 1 });
  };

  // 清除筛选
  const handleClearFilters = () => {
    setSearchText('');
    setSelectedCategory(undefined);
    clearFilters();
    fetchKnowledgeList({ page: 1 });
  };

  // 刷新数据
  const handleRefresh = () => {
    fetchKnowledgeList({ ...filters, page });
    fetchCategories();
  };

  // 创建文档
  const handleCreate = () => {
    navigate('/knowledge/create');
  };

  // 查看详情
  const handleViewDetail = (record: KnowledgeBase) => {
    navigate(`/knowledge/${record.id}`);
  };

  // 编辑文档
  const handleEdit = (record: KnowledgeBase) => {
    navigate(`/knowledge/${record.id}/edit`);
  };

  // 删除文档
  const handleDelete = async (id: string) => {
    try {
      await deleteKnowledge(id);
      message.success('删除成功');
      fetchKnowledgeList({ ...filters, page });
    } catch (error) {
      message.error('删除失败');
    }
  };

  // 管理分类
  const handleManageCategories = () => {
    navigate('/knowledge/categories');
  };

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
        title: cat.name,
        icon: <FolderOutlined />,
        children: cat.children.length > 0 ? convertToTreeData(cat.children) : undefined,
      }));
    };
    
    return convertToTreeData(rootCategories);
  };

  // 表格列定义
  const columns: ColumnsType<KnowledgeBase> = [
    {
      title: '文档标题',
      dataIndex: 'title',
      key: 'title',
      ellipsis: true,
      render: (text: string, record: KnowledgeBase) => (
        <Space>
          <FileTextOutlined />
          <Button
            type="link"
            onClick={() => handleViewDetail(record)}
            style={{ padding: 0, height: 'auto' }}
          >
            {text}
          </Button>
        </Space>
      ),
    },
    {
      title: '分类',
      dataIndex: 'category',
      key: 'category',
      width: 120,
      render: (category: KnowledgeCategory) => (
        <Tag color="blue">{category?.name || '未分类'}</Tag>
      ),
    },
    {
      title: '标签',
      dataIndex: 'tags',
      key: 'tags',
      width: 200,
      render: (tags: string[]) => (
        <Space wrap>
          {tags?.map(tag => (
            <Tag key={tag} size="small">{tag}</Tag>
          )) || '-'}
        </Space>
      ),
    },
    {
      title: '作者',
      dataIndex: 'author',
      key: 'author',
      width: 120,
      ellipsis: true,
    },
    {
      title: '浏览次数',
      dataIndex: 'view_count',
      key: 'view_count',
      width: 100,
      render: (count: number) => count || 0,
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
      render: (_, record: KnowledgeBase) => (
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
            title="确定要删除这个文档吗？"
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

  const treeData = buildCategoryTree(categories);

  return (
    <Row gutter={16}>
      {/* 左侧分类树 */}
      <Col span={6}>
        <Card
          title="文档分类"
          size="small"
          extra={
            <Button
              type="text"
              size="small"
              onClick={handleManageCategories}
            >
              管理
            </Button>
          }
        >
          <div style={{ marginBottom: 8 }}>
            <Button
              type="text"
              size="small"
              onClick={() => handleCategoryChange(undefined)}
              style={{
                color: selectedCategory === undefined ? '#1890ff' : undefined,
                fontWeight: selectedCategory === undefined ? 'bold' : undefined,
              }}
            >
              全部文档
            </Button>
          </div>
          <Tree
            treeData={treeData}
            selectedKeys={selectedCategory ? [selectedCategory] : []}
            onSelect={(keys) => {
              const key = keys[0] as string;
              handleCategoryChange(key);
            }}
            showIcon
            blockNode
          />
        </Card>
      </Col>

      {/* 右侧文档列表 */}
      <Col span={18}>
        <Card>
          {/* 操作栏 */}
          <div style={{ marginBottom: 16 }}>
            <Space wrap>
              <Button
                type="primary"
                icon={<PlusOutlined />}
                onClick={handleCreate}
              >
                创建文档
              </Button>
              <Search
                placeholder="搜索文档标题或内容"
                value={searchText}
                onChange={(e) => setSearchText(e.target.value)}
                onSearch={handleSearch}
                style={{ width: 250 }}
                allowClear
              />
              <Button onClick={handleClearFilters}>清除筛选</Button>
              <Button icon={<ReloadOutlined />} onClick={handleRefresh}>
                刷新
              </Button>
            </Space>
          </div>

          {/* 表格 */}
          <Table
            columns={columns}
            dataSource={knowledgeList}
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
                fetchKnowledgeList({ ...filters, page: newPage, limit: newPageSize });
              },
            }}
            scroll={{ x: 1200 }}
          />
        </Card>
      </Col>
    </Row>
  );
};

export default KnowledgeList;