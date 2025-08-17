import React, { useEffect, useState } from 'react';
import {
  Card,
  Button,
  Space,
  Tag,
  Typography,
  Divider,
  Row,
  Col,
  Avatar,
  Tooltip,
  message,
  Modal,
  Rate,
  List,
  Comment,
} from 'antd';
import {
  ArrowLeftOutlined,
  EditOutlined,
  DeleteOutlined,
  LikeOutlined,
  LikeFilled,
  EyeOutlined,
  ShareAltOutlined,
  PrinterOutlined,
  DownloadOutlined,
  StarOutlined,
  StarFilled,
} from '@ant-design/icons';
import { useNavigate, useParams } from 'react-router-dom';
import { useKnowledge, useUI } from '../../hooks';
import { KnowledgeBase } from '../../types';
import { formatDate } from '../../utils';

const { Title, Paragraph, Text } = Typography;
const { confirm } = Modal;

const KnowledgeDetail: React.FC = () => {
  const navigate = useNavigate();
  const { id } = useParams<{ id: string }>();
  const { setBreadcrumb } = useUI();
  const {
    currentKnowledge,
    loading,
    fetchKnowledge,
    deleteKnowledge,
    clearCurrentKnowledge,
  } = useKnowledge();

  const [liked, setLiked] = useState(false);
  const [starred, setStarred] = useState(false);
  const [rating, setRating] = useState(0);
  const [relatedDocs, setRelatedDocs] = useState<KnowledgeBase[]>([]);

  useEffect(() => {
    if (id) {
      fetchKnowledge(id);
    }

    return () => {
      clearCurrentKnowledge();
    };
  }, [id, fetchKnowledge, clearCurrentKnowledge]);

  useEffect(() => {
    if (currentKnowledge) {
      setBreadcrumb([
        { title: '知识库管理' },
        { title: '文档列表', path: '/knowledge' },
        { title: currentKnowledge.title },
      ]);
    }
  }, [currentKnowledge, setBreadcrumb]);

  // 返回列表
  const handleBack = () => {
    navigate('/knowledge');
  };

  // 编辑文档
  const handleEdit = () => {
    if (currentKnowledge) {
      navigate(`/knowledge/${currentKnowledge.id}/edit`);
    }
  };

  // 删除文档
  const handleDelete = () => {
    if (!currentKnowledge) return;

    confirm({
      title: '确认删除',
      content: `确定要删除文档 "${currentKnowledge.title}" 吗？此操作不可恢复。`,
      okText: '删除',
      okType: 'danger',
      cancelText: '取消',
      onOk: async () => {
        try {
          await deleteKnowledge(currentKnowledge.id);
          message.success('删除成功');
          navigate('/knowledge');
        } catch (error) {
          message.error('删除失败');
        }
      },
    });
  };

  // 点赞
  const handleLike = () => {
    setLiked(!liked);
    message.success(liked ? '取消点赞' : '点赞成功');
  };

  // 收藏
  const handleStar = () => {
    setStarred(!starred);
    message.success(starred ? '取消收藏' : '收藏成功');
  };

  // 分享
  const handleShare = () => {
    const url = window.location.href;
    navigator.clipboard.writeText(url).then(() => {
      message.success('链接已复制到剪贴板');
    }).catch(() => {
      message.error('复制失败');
    });
  };

  // 打印
  const handlePrint = () => {
    window.print();
  };

  // 下载
  const handleDownload = () => {
    if (currentKnowledge) {
      const element = document.createElement('a');
      const file = new Blob([currentKnowledge.content], { type: 'text/markdown' });
      element.href = URL.createObjectURL(file);
      element.download = `${currentKnowledge.title}.md`;
      document.body.appendChild(element);
      element.click();
      document.body.removeChild(element);
    }
  };

  // 评分
  const handleRate = (value: number) => {
    setRating(value);
    message.success(`评分：${value}星`);
  };

  if (loading || !currentKnowledge) {
    return (
      <Card loading={loading}>
        <div style={{ height: 400 }} />
      </Card>
    );
  }

  return (
    <div className="knowledge-detail">
      <Card
        title={
          <Space>
            <Button
              icon={<ArrowLeftOutlined />}
              onClick={handleBack}
            >
              返回
            </Button>
            <span>文档详情</span>
          </Space>
        }
        extra={
          <Space>
            <Button
              icon={<ShareAltOutlined />}
              onClick={handleShare}
            >
              分享
            </Button>
            <Button
              icon={<PrinterOutlined />}
              onClick={handlePrint}
            >
              打印
            </Button>
            <Button
              icon={<DownloadOutlined />}
              onClick={handleDownload}
            >
              下载
            </Button>
            <Button
              icon={<EditOutlined />}
              type="primary"
              onClick={handleEdit}
            >
              编辑
            </Button>
            <Button
              icon={<DeleteOutlined />}
              danger
              onClick={handleDelete}
            >
              删除
            </Button>
          </Space>
        }
      >
        <Row gutter={24}>
          <Col span={18}>
            {/* 文档标题 */}
            <Title level={1}>{currentKnowledge.title}</Title>

            {/* 文档元信息 */}
            <Space size={16} style={{ marginBottom: 24 }}>
              <Space>
                <Avatar src={currentKnowledge.author?.avatar} size="small">
                  {currentKnowledge.author?.username?.[0]?.toUpperCase()}
                </Avatar>
                <Text>{currentKnowledge.author?.username}</Text>
              </Space>
              <Text type="secondary">
                创建时间：{formatDate(currentKnowledge.created_at)}
              </Text>
              <Text type="secondary">
                更新时间：{formatDate(currentKnowledge.updated_at)}
              </Text>
              <Space>
                <EyeOutlined />
                <Text type="secondary">{currentKnowledge.views} 次浏览</Text>
              </Space>
              <Space>
                <Text type="secondary">评分：</Text>
                <Rate disabled value={currentKnowledge.score} allowHalf />
                <Text type="secondary">({currentKnowledge.score})</Text>
              </Space>
            </Space>

            {/* 文档标签 */}
            {currentKnowledge.tags && currentKnowledge.tags.length > 0 && (
              <div style={{ marginBottom: 24 }}>
                <Space size={[0, 8]} wrap>
                  {currentKnowledge.tags.map(tag => (
                    <Tag key={tag} color="blue">{tag}</Tag>
                  ))}
                </Space>
              </div>
            )}

            {/* 文档摘要 */}
            {currentKnowledge.summary && (
              <Card size="small" style={{ marginBottom: 24, backgroundColor: '#f6f8fa' }}>
                <Text strong>摘要：</Text>
                <Paragraph style={{ marginBottom: 0, marginTop: 8 }}>
                  {currentKnowledge.summary}
                </Paragraph>
              </Card>
            )}

            <Divider />

            {/* 文档内容 */}
            <div className="knowledge-content">
              <Paragraph>
                <pre style={{ whiteSpace: 'pre-wrap', fontFamily: 'inherit' }}>
                  {currentKnowledge.content}
                </pre>
              </Paragraph>
            </div>

            <Divider />

            {/* 操作按钮 */}
            <Space size={16}>
              <Button
                type={liked ? 'primary' : 'default'}
                icon={liked ? <LikeFilled /> : <LikeOutlined />}
                onClick={handleLike}
              >
                {liked ? '已点赞' : '点赞'} ({currentKnowledge.likes})
              </Button>
              <Button
                type={starred ? 'primary' : 'default'}
                icon={starred ? <StarFilled /> : <StarOutlined />}
                onClick={handleStar}
              >
                {starred ? '已收藏' : '收藏'}
              </Button>
              <Space>
                <Text>为这篇文档评分：</Text>
                <Rate value={rating} onChange={handleRate} />
              </Space>
            </Space>
          </Col>

          <Col span={6}>
            {/* 文档信息卡片 */}
            <Card title="文档信息" size="small" style={{ marginBottom: 16 }}>
              <Space direction="vertical" style={{ width: '100%' }}>
                <div>
                  <Text type="secondary">分类：</Text>
                  <Tag color="green">
                    {currentKnowledge.category?.name || '未分类'}
                  </Tag>
                </div>
                <div>
                  <Text type="secondary">状态：</Text>
                  <Tag color={
                    currentKnowledge.status === 'published' ? 'green' :
                    currentKnowledge.status === 'draft' ? 'orange' : 'default'
                  }>
                    {currentKnowledge.status === 'published' ? '已发布' :
                     currentKnowledge.status === 'draft' ? '草稿' : '已归档'}
                  </Tag>
                </div>
                <div>
                  <Text type="secondary">浏览次数：</Text>
                  <Text>{currentKnowledge.views}</Text>
                </div>
                <div>
                  <Text type="secondary">点赞次数：</Text>
                  <Text>{currentKnowledge.likes}</Text>
                </div>
              </Space>
            </Card>

            {/* 相关文档 */}
            {relatedDocs.length > 0 && (
              <Card title="相关文档" size="small">
                <List
                  size="small"
                  dataSource={relatedDocs}
                  renderItem={item => (
                    <List.Item>
                      <Button
                        type="link"
                        onClick={() => navigate(`/knowledge/${item.id}`)}
                        style={{ padding: 0, height: 'auto', textAlign: 'left' }}
                      >
                        {item.title}
                      </Button>
                    </List.Item>
                  )}
                />
              </Card>
            )}
          </Col>
        </Row>
      </Card>
    </div>
  );
};

export default KnowledgeDetail;