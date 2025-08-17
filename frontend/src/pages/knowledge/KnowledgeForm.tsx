import React, { useEffect, useState } from 'react';
import {
  Form,
  Input,
  Select,
  Button,
  Card,
  Space,
  message,
  Tag,
  TreeSelect,
  Switch,
  Row,
  Col,
} from 'antd';
import {
  SaveOutlined,
  ArrowLeftOutlined,
  EyeOutlined,
  TagsOutlined,
} from '@ant-design/icons';
import { useNavigate, useParams } from 'react-router-dom';
import { useKnowledge, useUI } from '../../hooks';
import { KnowledgeBase, KnowledgeCategory } from '../../types';
import type { DataNode } from 'antd/es/tree';

const { TextArea } = Input;
const { Option } = Select;

interface KnowledgeFormData {
  title: string;
  content: string;
  summary?: string;
  tags: string[];
  category_id?: string;
  status: 'draft' | 'published' | 'archived';
}

const KnowledgeForm: React.FC = () => {
  const navigate = useNavigate();
  const { id } = useParams<{ id: string }>();
  const { setBreadcrumb } = useUI();
  const {
    currentKnowledge,
    categories,
    tags,
    loading,
    fetchKnowledge,
    fetchCategories,
    fetchTags,
    createKnowledge,
    updateKnowledge,
    clearCurrentKnowledge,
  } = useKnowledge();

  const [form] = Form.useForm<KnowledgeFormData>();
  const [isEditing, setIsEditing] = useState(false);
  const [submitting, setSubmitting] = useState(false);
  const [tagInput, setTagInput] = useState('');
  const [customTags, setCustomTags] = useState<string[]>([]);

  const isEditMode = Boolean(id);

  useEffect(() => {
    setBreadcrumb([
      { title: '知识库管理' },
      { title: '文档列表', path: '/knowledge' },
      { title: isEditMode ? '编辑文档' : '创建文档' },
    ]);

    fetchCategories();
    fetchTags();

    if (isEditMode && id) {
      fetchKnowledge(id);
    }

    return () => {
      clearCurrentKnowledge();
    };
  }, [setBreadcrumb, isEditMode, id, fetchKnowledge, fetchCategories, fetchTags, clearCurrentKnowledge]);

  useEffect(() => {
    if (currentKnowledge && isEditMode) {
      form.setFieldsValue({
        title: currentKnowledge.title,
        content: currentKnowledge.content,
        summary: currentKnowledge.summary,
        tags: currentKnowledge.tags,
        category_id: currentKnowledge.category_id,
        status: currentKnowledge.status,
      });
      setCustomTags(currentKnowledge.tags);
    }
  }, [currentKnowledge, isEditMode, form]);

  // 构建分类树数据
  const buildCategoryTreeData = (categories: KnowledgeCategory[]): DataNode[] => {
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
    
    // 转换为 TreeSelect 组件需要的格式
    const convertToTreeData = (cats: (KnowledgeCategory & { children: KnowledgeCategory[] })[]): DataNode[] => {
      return cats.map(cat => ({
        key: cat.id,
        value: cat.id,
        title: cat.name,
        children: cat.children.length > 0 ? convertToTreeData(cat.children) : undefined,
      }));
    };
    
    return convertToTreeData(rootCategories);
  };

  // 处理标签输入
  const handleTagInputChange = (value: string) => {
    setTagInput(value);
  };

  // 添加自定义标签
  const handleAddTag = () => {
    if (tagInput && !customTags.includes(tagInput)) {
      const newTags = [...customTags, tagInput];
      setCustomTags(newTags);
      form.setFieldsValue({ tags: newTags });
      setTagInput('');
    }
  };

  // 移除标签
  const handleRemoveTag = (tagToRemove: string) => {
    const newTags = customTags.filter(tag => tag !== tagToRemove);
    setCustomTags(newTags);
    form.setFieldsValue({ tags: newTags });
  };

  // 处理表单提交
  const handleSubmit = async (values: KnowledgeFormData) => {
    setSubmitting(true);
    try {
      const submitData = {
        ...values,
        tags: customTags,
      };

      if (isEditMode && id) {
        await updateKnowledge(id, submitData);
        message.success('更新成功');
      } else {
        await createKnowledge(submitData);
        message.success('创建成功');
      }
      
      navigate('/knowledge');
    } catch (error) {
      message.error(isEditMode ? '更新失败' : '创建失败');
    } finally {
      setSubmitting(false);
    }
  };

  // 预览文档
  const handlePreview = () => {
    if (isEditMode && id) {
      navigate(`/knowledge/${id}`);
    }
  };

  // 返回列表
  const handleBack = () => {
    navigate('/knowledge');
  };

  return (
    <div className="knowledge-form">
      <Card
        title={
          <Space>
            <Button
              icon={<ArrowLeftOutlined />}
              onClick={handleBack}
            >
              返回
            </Button>
            <span>{isEditMode ? '编辑文档' : '创建文档'}</span>
          </Space>
        }
        extra={
          <Space>
            {isEditMode && (
              <Button
                icon={<EyeOutlined />}
                onClick={handlePreview}
              >
                预览
              </Button>
            )}
            <Button
              type="primary"
              icon={<SaveOutlined />}
              loading={submitting}
              onClick={() => form.submit()}
            >
              {isEditMode ? '更新' : '创建'}
            </Button>
          </Space>
        }
      >
        <Form
          form={form}
          layout="vertical"
          onFinish={handleSubmit}
          initialValues={{
            status: 'draft',
            tags: [],
          }}
        >
          <Row gutter={24}>
            <Col span={16}>
              <Form.Item
                name="title"
                label="文档标题"
                rules={[
                  { required: true, message: '请输入文档标题' },
                  { max: 200, message: '标题长度不能超过200个字符' },
                ]}
              >
                <Input
                  placeholder="请输入文档标题"
                  size="large"
                />
              </Form.Item>

              <Form.Item
                name="summary"
                label="文档摘要"
                rules={[
                  { max: 500, message: '摘要长度不能超过500个字符' },
                ]}
              >
                <TextArea
                  placeholder="请输入文档摘要（可选）"
                  rows={3}
                />
              </Form.Item>

              <Form.Item
                name="content"
                label="文档内容"
                rules={[
                  { required: true, message: '请输入文档内容' },
                ]}
              >
                <TextArea
                  placeholder="请输入文档内容，支持Markdown格式"
                  rows={20}
                />
              </Form.Item>
            </Col>

            <Col span={8}>
              <Form.Item
                name="category_id"
                label="文档分类"
              >
                <TreeSelect
                  placeholder="请选择文档分类"
                  treeData={buildCategoryTreeData(categories as KnowledgeCategory[])}
                  allowClear
                />
              </Form.Item>

              <Form.Item
                name="status"
                label="发布状态"
                rules={[{ required: true, message: '请选择发布状态' }]}
              >
                <Select placeholder="请选择发布状态">
                  <Option value="draft">草稿</Option>
                  <Option value="published">已发布</Option>
                  <Option value="archived">已归档</Option>
                </Select>
              </Form.Item>

              <Form.Item
                label="文档标签"
              >
                <Space direction="vertical" style={{ width: '100%' }}>
                  <Input.Group compact>
                    <Input
                      style={{ width: 'calc(100% - 80px)' }}
                      placeholder="输入标签名称"
                      value={tagInput}
                      onChange={(e) => handleTagInputChange(e.target.value)}
                      onPressEnter={handleAddTag}
                    />
                    <Button
                      type="primary"
                      icon={<TagsOutlined />}
                      onClick={handleAddTag}
                      disabled={!tagInput}
                    >
                      添加
                    </Button>
                  </Input.Group>
                  
                  <div style={{ minHeight: 60, border: '1px dashed #d9d9d9', padding: 8, borderRadius: 4 }}>
                    {customTags.length > 0 ? (
                      <Space size={[0, 8]} wrap>
                        {customTags.map(tag => (
                          <Tag
                            key={tag}
                            closable
                            onClose={() => handleRemoveTag(tag)}
                          >
                            {tag}
                          </Tag>
                        ))}
                      </Space>
                    ) : (
                      <div style={{ color: '#999', textAlign: 'center', padding: 16 }}>
                        暂无标签
                      </div>
                    )}
                  </div>
                </Space>
              </Form.Item>

              <Form.Item name="tags" hidden>
                <Input />
              </Form.Item>
            </Col>
          </Row>
        </Form>
      </Card>
    </div>
  );
};

export default KnowledgeForm;