import React, { useState, useEffect } from 'react';
import {
  Card,
  Form,
  Input,
  Button,
  Select,
  Upload,
  Space,
  Row,
  Col,
  message,
  Spin,
  Tag,
  Divider,
  Switch,
  Alert
} from 'antd';
import {
  SaveOutlined,
  ArrowLeftOutlined,
  UploadOutlined,
  PlusOutlined,
  DeleteOutlined,
  EyeOutlined
} from '@ant-design/icons';
import { useNavigate, useParams } from 'react-router-dom';
import { useKnowledge } from '../../hooks/useKnowledge';
import { useUI } from '../../hooks/useUI';
import { knowledgeService } from '../../services/knowledge';
import type { KnowledgeDocument, KnowledgeCategory } from '../../types';
import type { UploadFile } from 'antd/es/upload/interface';

const { TextArea } = Input;
const { Option } = Select;

interface KnowledgeFormData {
  title: string;
  content: string;
  category_id?: string;
  tags: string[];
  is_public: boolean;
  summary?: string;
  attachments?: UploadFile[];
}

const KnowledgeForm: React.FC = () => {
  const navigate = useNavigate();
  const { id } = useParams<{ id: string }>();
  const { loading, setLoading } = useUI();
  const [form] = Form.useForm<KnowledgeFormData>();
  
  const [categories, setCategories] = useState<KnowledgeCategory[]>([]);
  const [tags, setTags] = useState<string[]>([]);
  const [newTag, setNewTag] = useState('');
  const [attachments, setAttachments] = useState<UploadFile[]>([]);
  const [isPreview, setIsPreview] = useState(false);
  const [autoSummary, setAutoSummary] = useState(false);
  
  const isEdit = Boolean(id);

  // 加载分类列表
  const loadCategories = async () => {
    try {
      const response = await knowledgeService.getCategories();
      setCategories(response);
    } catch (error) {
      message.error('加载分类失败');
    }
  };

  // 加载标签列表
  const loadTags = async () => {
    try {
      const response = await knowledgeService.getTags();
      setTags(response);
    } catch (error) {
      message.error('加载标签失败');
    }
  };

  // 加载文档详情（编辑模式）
  const loadDocument = async () => {
    if (!id) return;
    
    try {
      setLoading(true);
      const document = await knowledgeService.getDocument(id);
      
      form.setFieldsValue({
        title: document.title,
        content: document.content,
        category_id: document.category_id,
        tags: document.tags || [],
        is_public: document.is_public,
        summary: document.summary
      });
      
      if (document.attachments) {
        setAttachments(document.attachments.map((att, index) => ({
          uid: `${index}`,
          name: att.name,
          status: 'done',
          url: att.url
        })));
      }
    } catch (error) {
      message.error('加载文档失败');
      navigate('/knowledge');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadCategories();
    loadTags();
    if (isEdit) {
      loadDocument();
    }
  }, [id]);

  // 处理表单提交
  const handleSubmit = async (values: KnowledgeFormData) => {
    try {
      setLoading(true);
      
      const formData = {
        ...values,
        attachments: attachments.filter(file => file.originFileObj)
      };

      if (isEdit) {
        await knowledgeService.updateDocument(id!, formData);
        message.success('文档更新成功');
      } else {
        await knowledgeService.createDocument(formData);
        message.success('文档创建成功');
      }
      
      navigate('/knowledge');
    } catch (error) {
      message.error(isEdit ? '文档更新失败' : '文档创建失败');
    } finally {
      setLoading(false);
    }
  };

  // 生成摘要
  const handleGenerateSummary = async () => {
    const content = form.getFieldValue('content');
    if (!content) {
      message.warning('请先输入文档内容');
      return;
    }

    try {
      setAutoSummary(true);
      const summary = await knowledgeService.generateSummary(content);
      form.setFieldValue('summary', summary);
      message.success('摘要生成成功');
    } catch (error) {
      message.error('摘要生成失败');
    } finally {
      setAutoSummary(false);
    }
  };

  // 自动标记标签
  const handleAutoTag = async () => {
    const content = form.getFieldValue('content');
    if (!content) {
      message.warning('请先输入文档内容');
      return;
    }

    try {
      const autoTags = await knowledgeService.autoTag(content);
      const currentTags = form.getFieldValue('tags') || [];
      const newTags = [...new Set([...currentTags, ...autoTags])];
      form.setFieldValue('tags', newTags);
      message.success('标签自动标记成功');
    } catch (error) {
      message.error('自动标记失败');
    }
  };

  // 添加新标签
  const handleAddTag = () => {
    if (newTag && !tags.includes(newTag)) {
      setTags([...tags, newTag]);
      const currentTags = form.getFieldValue('tags') || [];
      form.setFieldValue('tags', [...currentTags, newTag]);
      setNewTag('');
    }
  };

  // 处理文件上传
  const handleUploadChange = ({ fileList }: { fileList: UploadFile[] }) => {
    setAttachments(fileList);
  };

  // 预览内容
  const renderPreview = () => {
    const values = form.getFieldsValue();
    return (
      <Card title="预览" size="small">
        <div style={{ marginBottom: '16px' }}>
          <h2>{values.title || '未命名文档'}</h2>
          {values.category_id && (
            <Tag color="blue">
              {categories.find(c => c.id === values.category_id)?.name}
            </Tag>
          )}
          {values.tags?.map(tag => (
            <Tag key={tag} color="green">{tag}</Tag>
          ))}
          {values.is_public && <Tag color="orange">公开</Tag>}
        </div>
        
        {values.summary && (
          <Alert
            message="摘要"
            description={values.summary}
            type="info"
            style={{ marginBottom: '16px' }}
          />
        )}
        
        <div
          style={{
            border: '1px solid #d9d9d9',
            borderRadius: '6px',
            padding: '16px',
            minHeight: '200px',
            whiteSpace: 'pre-wrap'
          }}
        >
          {values.content || '暂无内容'}
        </div>
      </Card>
    );
  };

  return (
    <div style={{ padding: '24px' }}>
      <Card
        title={
          <Space>
            <Button
              icon={<ArrowLeftOutlined />}
              onClick={() => navigate('/knowledge')}
            >
              返回
            </Button>
            <span>{isEdit ? '编辑文档' : '创建文档'}</span>
          </Space>
        }
        extra={
          <Space>
            <Button
              icon={<EyeOutlined />}
              onClick={() => setIsPreview(!isPreview)}
            >
              {isPreview ? '编辑' : '预览'}
            </Button>
            <Button
              type="primary"
              icon={<SaveOutlined />}
              loading={loading}
              onClick={() => form.submit()}
            >
              {isEdit ? '更新' : '创建'}
            </Button>
          </Space>
        }
      >
        <Spin spinning={loading}>
          {isPreview ? (
            renderPreview()
          ) : (
            <Form
              form={form}
              layout="vertical"
              onFinish={handleSubmit}
              initialValues={{
                is_public: false,
                tags: []
              }}
            >
              <Row gutter={24}>
                <Col span={16}>
                  <Form.Item
                    name="title"
                    label="文档标题"
                    rules={[
                      { required: true, message: '请输入文档标题' },
                      { max: 200, message: '标题不能超过200个字符' }
                    ]}
                  >
                    <Input placeholder="请输入文档标题" />
                  </Form.Item>

                  <Form.Item
                    name="content"
                    label="文档内容"
                    rules={[
                      { required: true, message: '请输入文档内容' }
                    ]}
                  >
                    <TextArea
                      rows={20}
                      placeholder="请输入文档内容，支持Markdown格式"
                    />
                  </Form.Item>

                  <Form.Item
                    name="summary"
                    label={
                      <Space>
                        <span>文档摘要</span>
                        <Button
                          type="link"
                          size="small"
                          loading={autoSummary}
                          onClick={handleGenerateSummary}
                        >
                          自动生成
                        </Button>
                      </Space>
                    }
                  >
                    <TextArea
                      rows={3}
                      placeholder="请输入文档摘要，或点击自动生成"
                    />
                  </Form.Item>
                </Col>

                <Col span={8}>
                  <Form.Item
                    name="category_id"
                    label="文档分类"
                  >
                    <Select
                      placeholder="请选择文档分类"
                      allowClear
                    >
                      {categories.map(category => (
                        <Option key={category.id} value={category.id}>
                          {category.name}
                        </Option>
                      ))}
                    </Select>
                  </Form.Item>

                  <Form.Item
                    name="tags"
                    label={
                      <Space>
                        <span>标签</span>
                        <Button
                          type="link"
                          size="small"
                          onClick={handleAutoTag}
                        >
                          自动标记
                        </Button>
                      </Space>
                    }
                  >
                    <Select
                      mode="multiple"
                      placeholder="请选择或输入标签"
                      dropdownRender={menu => (
                        <div>
                          {menu}
                          <Divider style={{ margin: '4px 0' }} />
                          <Space style={{ padding: '0 8px 4px' }}>
                            <Input
                              placeholder="新标签"
                              value={newTag}
                              onChange={e => setNewTag(e.target.value)}
                              onPressEnter={handleAddTag}
                            />
                            <Button
                              type="text"
                              icon={<PlusOutlined />}
                              onClick={handleAddTag}
                            >
                              添加
                            </Button>
                          </Space>
                        </div>
                      )}
                    >
                      {tags.map(tag => (
                        <Option key={tag} value={tag}>
                          {tag}
                        </Option>
                      ))}
                    </Select>
                  </Form.Item>

                  <Form.Item
                    name="is_public"
                    label="公开文档"
                    valuePropName="checked"
                  >
                    <Switch />
                  </Form.Item>

                  <Form.Item label="附件">
                    <Upload
                      fileList={attachments}
                      onChange={handleUploadChange}
                      beforeUpload={() => false}
                      multiple
                    >
                      <Button icon={<UploadOutlined />}>
                        选择文件
                      </Button>
                    </Upload>
                  </Form.Item>
                </Col>
              </Row>
            </Form>
          )}
        </Spin>
      </Card>
    </div>
  );
};

export default KnowledgeForm;