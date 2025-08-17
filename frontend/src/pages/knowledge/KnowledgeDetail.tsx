import React, { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import {
  Card,
  Typography,
  Tag,
  Button,
  Space,
  Divider,
  Rate,
  message,
  Modal,
  Spin,
  Avatar,
  List,

  Form,
  Input,
  Row,
  Col
} from 'antd';
import {
  EditOutlined,
  DeleteOutlined,
  ShareAltOutlined,
  LikeOutlined,
  DislikeOutlined,
  EyeOutlined,
  CalendarOutlined,
  UserOutlined
} from '@ant-design/icons';
import { useKnowledge, useUI } from '../../hooks';
import { knowledgeService } from '../../services';
import { Knowledge } from '../../types';
import { formatDate } from '../../utils';

const { Title, Paragraph, Text } = Typography;
const { TextArea } = Input;

interface Comment {
  id: string;
  content: string;
  author: string;
  createdAt: string;
  rating?: number;
}

const KnowledgeDetail: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { loading, setLoading } = useUI();
  const [knowledge, setKnowledge] = useState<Knowledge | null>(null);
  const [comments, setComments] = useState<Comment[]>([]);
  const [relatedDocs, setRelatedDocs] = useState<Knowledge[]>([]);
  const [userRating, setUserRating] = useState<number>(0);
  const [form] = Form.useForm();

  useEffect(() => {
    if (id) {
      loadKnowledgeDetail();
      loadComments();
      loadRelatedDocs();
    }
  }, [id]);

  const loadKnowledgeDetail = async () => {
    try {
      setLoading(true);
      const response = await knowledgeService.getById(id!);
      setKnowledge(response.data);
    } catch (error) {
      message.error('加载知识库详情失败');
    } finally {
      setLoading(false);
    }
  };

  const loadComments = async () => {
    try {
      // 模拟评论数据
      setComments([
        {
          id: '1',
          content: '这篇文档很有帮助，解决了我的问题。',
          author: '张三',
          createdAt: '2024-01-15 10:30:00',
          rating: 5
        },
        {
          id: '2',
          content: '内容详细，但是示例可以更多一些。',
          author: '李四',
          createdAt: '2024-01-14 15:20:00',
          rating: 4
        }
      ]);
    } catch (error) {
      console.error('加载评论失败:', error);
    }
  };

  const loadRelatedDocs = async () => {
    try {
      const response = await knowledgeService.getRelated(id!);
      setRelatedDocs(response.data);
    } catch (error) {
      console.error('加载相关文档失败:', error);
    }
  };

  const handleEdit = () => {
    navigate(`/knowledge/edit/${id}`);
  };

  const handleDelete = () => {
    Modal.confirm({
      title: '确认删除',
      content: '确定要删除这篇知识库文档吗？此操作不可恢复。',
      okText: '确定',
      cancelText: '取消',
      onOk: async () => {
        try {
          await knowledgeService.delete(id!);
          message.success('删除成功');
          navigate('/knowledge');
        } catch (error) {
          message.error('删除失败');
        }
      }
    });
  };

  const handleRating = async (value: number) => {
    try {
      await knowledgeService.rate(id!, value);
      setUserRating(value);
      message.success('评分成功');
    } catch (error) {
      message.error('评分失败');
    }
  };

  const handleAddComment = async (values: { content: string; rating?: number }) => {
    try {
      // 这里应该调用实际的API
      const newComment: Comment = {
        id: Date.now().toString(),
        content: values.content,
        author: '当前用户',
        createdAt: new Date().toLocaleString(),
        rating: values.rating
      };
      setComments([newComment, ...comments]);
      form.resetFields();
      message.success('评论添加成功');
    } catch (error) {
      message.error('添加评论失败');
    }
  };

  const handleShare = () => {
    const url = window.location.href;
    navigator.clipboard.writeText(url).then(() => {
      message.success('链接已复制到剪贴板');
    }).catch(() => {
      message.error('复制失败');
    });
  };

  if (loading) {
    return (
      <div style={{ textAlign: 'center', padding: '50px' }}>
        <Spin size="large" />
      </div>
    );
  }

  if (!knowledge) {
    return (
      <div style={{ textAlign: 'center', padding: '50px' }}>
        <Text>知识库文档不存在</Text>
      </div>
    );
  }

  return (
    <div style={{ padding: '24px' }}>
      <Row gutter={24}>
        <Col span={18}>
          <Card>
            <div style={{ marginBottom: '24px' }}>
              <Space style={{ marginBottom: '16px' }}>
                <Button type="primary" icon={<EditOutlined />} onClick={handleEdit}>
                  编辑
                </Button>
                <Button danger icon={<DeleteOutlined />} onClick={handleDelete}>
                  删除
                </Button>
                <Button icon={<ShareAltOutlined />} onClick={handleShare}>
                  分享
                </Button>
              </Space>
            </div>

            <Title level={2}>{knowledge.title}</Title>
            
            <div style={{ marginBottom: '24px' }}>
              <Space wrap>
                <Text type="secondary">
                  <UserOutlined /> {knowledge.author}
                </Text>
                <Text type="secondary">
                  <CalendarOutlined /> {formatDate(knowledge.createdAt)}
                </Text>
                <Text type="secondary">
                  <EyeOutlined /> {knowledge.viewCount || 0} 次浏览
                </Text>
                <Rate disabled value={knowledge.rating || 0} />
                <Text type="secondary">({knowledge.ratingCount || 0} 人评分)</Text>
              </Space>
            </div>

            <div style={{ marginBottom: '24px' }}>
              <Space wrap>
                <Tag color="blue">{knowledge.category}</Tag>
                {knowledge.tags?.map(tag => (
                  <Tag key={tag}>{tag}</Tag>
                ))}
                <Tag color={knowledge.isPublic ? 'green' : 'orange'}>
                  {knowledge.isPublic ? '公开' : '私有'}
                </Tag>
              </Space>
            </div>

            {knowledge.summary && (
              <div style={{ marginBottom: '24px' }}>
                <Title level={4}>摘要</Title>
                <Paragraph>{knowledge.summary}</Paragraph>
              </div>
            )}

            <Divider />

            <div style={{ marginBottom: '24px' }}>
              <Title level={4}>内容</Title>
              <div 
                style={{ 
                  lineHeight: '1.8',
                  fontSize: '14px',
                  whiteSpace: 'pre-wrap'
                }}
                dangerouslySetInnerHTML={{ __html: knowledge.content }}
              />
            </div>

            <Divider />

            <div style={{ marginBottom: '24px' }}>
              <Title level={4}>评分</Title>
              <Space>
                <Text>为这篇文档评分：</Text>
                <Rate value={userRating} onChange={handleRating} />
              </Space>
            </div>

            <Divider />

            <div>
              <Title level={4}>评论 ({comments.length})</Title>
              
              <Form form={form} onFinish={handleAddComment} style={{ marginBottom: '24px' }}>
                <Form.Item name="content" rules={[{ required: true, message: '请输入评论内容' }]}>
                  <TextArea rows={4} placeholder="写下你的评论..." />
                </Form.Item>
                <Form.Item name="rating">
                  <Space>
                    <Text>评分：</Text>
                    <Rate />
                  </Space>
                </Form.Item>
                <Form.Item>
                  <Button type="primary" htmlType="submit">
                    发表评论
                  </Button>
                </Form.Item>
              </Form>

              <List
                dataSource={comments}
                renderItem={comment => (
                  <List.Item>
                    <List.Item.Meta
                      avatar={<Avatar icon={<UserOutlined />} />}
                      title={
                        <Space>
                          <Text strong>{comment.author}</Text>
                          <Text type="secondary" style={{ fontSize: '12px' }}>
                            {comment.createdAt}
                          </Text>
                        </Space>
                      }
                      description={
                        <div>
                          <Paragraph style={{ marginBottom: '8px' }}>
                            {comment.content}
                          </Paragraph>
                          {comment.rating && (
                            <Space>
                              <Text type="secondary">评分：</Text>
                              <Rate disabled value={comment.rating} size="small" />
                            </Space>
                          )}
                        </div>
                      }
                    />
                  </List.Item>
                )}
              />
            </div>
          </Card>
        </Col>
        
        <Col span={6}>
          <Card title="相关文档" size="small">
            <List
              size="small"
              dataSource={relatedDocs}
              renderItem={doc => (
                <List.Item>
                  <List.Item.Meta
                    title={
                      <a 
                        href={`/knowledge/detail/${doc.id}`}
                        style={{ fontSize: '14px' }}
                      >
                        {doc.title}
                      </a>
                    }
                    description={
                      <Space>
                        <Text type="secondary" style={{ fontSize: '12px' }}>
                          {formatDate(doc.createdAt)}
                        </Text>
                        <Rate disabled value={doc.rating || 0} size="small" />
                      </Space>
                    }
                  />
                </List.Item>
              )}
            />
          </Card>
        </Col>
      </Row>
    </div>
  );
};

export default KnowledgeDetail;