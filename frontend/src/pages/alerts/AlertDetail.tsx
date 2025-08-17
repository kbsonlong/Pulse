import React, { useEffect } from 'react';
import {
  Card,
  Descriptions,
  Tag,
  Button,
  Space,
  Modal,
  message,
  Badge,
  Divider,
  Timeline,
  Typography,
} from 'antd';
import {
  ArrowLeftOutlined,
  CheckOutlined,
  CloseOutlined,
  ExclamationCircleOutlined,
  ReloadOutlined,
} from '@ant-design/icons';
import { useParams, useNavigate } from 'react-router-dom';
import { useAlert, useUI } from '../../hooks';
import { AlertStatus } from '../../types';
import { formatDate, getPriorityColor } from '../../utils';

const { confirm } = Modal;
const { Title, Paragraph } = Typography;

const AlertDetail: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { setBreadcrumb } = useUI();
  const {
    currentAlert: alert,
    loading,
    fetchAlert,
    updateStatus,
    clearCurrentAlert,
  } = useAlert();

  useEffect(() => {
    if (id) {
      fetchAlert(id);
    }
    return () => {
      clearCurrentAlert();
    };
  }, [id, fetchAlert, clearCurrentAlert]);

  useEffect(() => {
    setBreadcrumb([
      { title: '告警管理' },
      { title: '告警列表', path: '/alerts' },
      { title: '告警详情' },
    ]);
  }, [setBreadcrumb]);

  // 返回列表
  const handleBack = () => {
    navigate('/alerts');
  };

  // 刷新数据
  const handleRefresh = () => {
    if (id) {
      fetchAlert(id);
    }
  };

  // 更新告警状态
  const handleUpdateStatus = (status: AlertStatus) => {
    if (!alert) return;

    confirm({
      title: '确认操作',
      icon: <ExclamationCircleOutlined />,
      content: `确定要将告警状态更新为"${status === 'resolved' ? '已解决' : '未解决'}"吗？`,
      onOk: async () => {
        try {
          await updateStatus(alert.id, status);
          message.success('状态更新成功');
          if (id) {
            fetchAlert(id);
          }
        } catch (error) {
          message.error('状态更新失败');
        }
      },
    });
  };

  if (!alert && !loading) {
    return (
      <Card>
        <div style={{ textAlign: 'center', padding: '50px 0' }}>
          <Title level={4}>告警不存在</Title>
          <Button type="primary" onClick={handleBack}>
            返回列表
          </Button>
        </div>
      </Card>
    );
  }

  return (
    <div>
      {/* 操作栏 */}
      <Card style={{ marginBottom: 16 }}>
        <Space>
          <Button icon={<ArrowLeftOutlined />} onClick={handleBack}>
            返回列表
          </Button>
          <Button icon={<ReloadOutlined />} onClick={handleRefresh}>
            刷新
          </Button>
          {alert && alert.status !== 'resolved' && (
            <Button
              type="primary"
              icon={<CheckOutlined />}
              onClick={() => handleUpdateStatus('resolved')}
            >
              标记为已解决
            </Button>
          )}
          {alert && alert.status === 'resolved' && (
            <Button
              icon={<CloseOutlined />}
              onClick={() => handleUpdateStatus('active')}
            >
              标记为未解决
            </Button>
          )}
        </Space>
      </Card>

      {/* 基本信息 */}
      <Card title="基本信息" loading={loading} style={{ marginBottom: 16 }}>
        {alert && (
          <Descriptions column={2} bordered>
            <Descriptions.Item label="告警名称" span={2}>
              <Title level={4} style={{ margin: 0 }}>
                {alert.name}
              </Title>
            </Descriptions.Item>
            <Descriptions.Item label="严重程度">
              <Tag color={getPriorityColor(alert.severity)} size="large">
                {alert.severity.toUpperCase()}
              </Tag>
            </Descriptions.Item>
            <Descriptions.Item label="状态">
              <Badge
                status={alert.status === 'resolved' ? 'success' : 'error'}
                text={alert.status === 'resolved' ? '已解决' : '未解决'}
              />
            </Descriptions.Item>
            <Descriptions.Item label="数据源">
              {alert.source}
            </Descriptions.Item>
            <Descriptions.Item label="告警类型">
              {alert.type || '-'}
            </Descriptions.Item>
            <Descriptions.Item label="创建时间">
              {formatDate(alert.created_at)}
            </Descriptions.Item>
            <Descriptions.Item label="更新时间">
              {formatDate(alert.updated_at)}
            </Descriptions.Item>
            {alert.resolved_at && (
              <Descriptions.Item label="解决时间" span={2}>
                {formatDate(alert.resolved_at)}
              </Descriptions.Item>
            )}
            <Descriptions.Item label="描述" span={2}>
              <Paragraph>
                {alert.description || '暂无描述'}
              </Paragraph>
            </Descriptions.Item>
          </Descriptions>
        )}
      </Card>

      {/* 告警详情 */}
      {alert && alert.details && (
        <Card title="告警详情" style={{ marginBottom: 16 }}>
          <pre style={{ 
            background: '#f5f5f5', 
            padding: 16, 
            borderRadius: 4,
            overflow: 'auto',
            maxHeight: 400
          }}>
            {JSON.stringify(alert.details, null, 2)}
          </pre>
        </Card>
      )}

      {/* 标签信息 */}
      {alert && alert.labels && Object.keys(alert.labels).length > 0 && (
        <Card title="标签信息" style={{ marginBottom: 16 }}>
          <Space wrap>
            {Object.entries(alert.labels).map(([key, value]) => (
              <Tag key={key} color="blue">
                {key}: {value}
              </Tag>
            ))}
          </Space>
        </Card>
      )}

      {/* 操作历史 */}
      <Card title="操作历史">
        <Timeline>
          {alert && (
            <>
              <Timeline.Item color="blue">
                <div>
                  <strong>告警创建</strong>
                  <div style={{ color: '#666', fontSize: '12px' }}>
                    {formatDate(alert.created_at)}
                  </div>
                </div>
              </Timeline.Item>
              {alert.updated_at !== alert.created_at && (
                <Timeline.Item color="orange">
                  <div>
                    <strong>告警更新</strong>
                    <div style={{ color: '#666', fontSize: '12px' }}>
                      {formatDate(alert.updated_at)}
                    </div>
                  </div>
                </Timeline.Item>
              )}
              {alert.resolved_at && (
                <Timeline.Item color="green">
                  <div>
                    <strong>告警解决</strong>
                    <div style={{ color: '#666', fontSize: '12px' }}>
                      {formatDate(alert.resolved_at)}
                    </div>
                  </div>
                </Timeline.Item>
              )}
            </>
          )}
        </Timeline>
      </Card>
    </div>
  );
};

export default AlertDetail;