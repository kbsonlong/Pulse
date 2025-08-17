import React, { useEffect, useState } from 'react';
import {
  Card,
  Descriptions,
  Button,
  Space,
  Tag,
  Typography,
  Divider,
  Table,
  Modal,
  message,
  Switch,
  Popconfirm,
} from 'antd';
import {
  ArrowLeftOutlined,
  EditOutlined,
  DeleteOutlined,
  PlayCircleOutlined,
  PauseCircleOutlined,
  CopyOutlined,
} from '@ant-design/icons';
import { useNavigate, useParams } from 'react-router-dom';
import { useRule, useUI } from '../../hooks';
import { formatDate } from '../../utils';
import type { ColumnsType } from 'antd/es/table';

const { Title, Text } = Typography;

const RuleDetail: React.FC = () => {
  const navigate = useNavigate();
  const { id } = useParams<{ id: string }>();
  const { setBreadcrumbs } = useUI();
  const {
    currentRule,
    loading,
    fetchRule,
    deleteRule,
    toggleRule,
    clearCurrentRule,
  } = useRule();

  const [deleteModalVisible, setDeleteModalVisible] = useState(false);

  useEffect(() => {
    setBreadcrumbs([
      { title: '规则管理' },
      { title: '规则列表', path: '/rules' },
      { title: '规则详情' },
    ]);

    if (id) {
      fetchRule(id);
    }

    return () => {
      clearCurrentRule();
    };
  }, [setBreadcrumbs, id, fetchRule, clearCurrentRule]);

  const handleBack = () => {
    navigate('/rules');
  };

  const handleEdit = () => {
    if (currentRule) {
      navigate(`/rules/${currentRule.id}/edit`);
    }
  };

  const handleDelete = async () => {
    if (!currentRule) return;
    
    try {
      await deleteRule(currentRule.id);
      message.success('规则删除成功');
      navigate('/rules');
    } catch (error) {
      message.error('规则删除失败');
    }
    setDeleteModalVisible(false);
  };

  const handleToggleStatus = async () => {
    if (!currentRule) return;
    
    try {
      await toggleRule(currentRule.id, !currentRule.enabled);
      message.success('规则状态更新成功');
    } catch (error) {
      message.error('规则状态更新失败');
    }
  };

  const handleDuplicate = () => {
    if (currentRule) {
      navigate('/rules/create', {
        state: { duplicateFrom: currentRule }
      });
    }
  };

  if (!currentRule && !loading) {
    return (
      <div style={{ padding: '24px' }}>
        <Card>
          <div style={{ textAlign: 'center', padding: '48px' }}>
            <Text type="secondary">规则不存在或已被删除</Text>
            <br />
            <Button type="primary" onClick={handleBack} style={{ marginTop: '16px' }}>
              返回规则列表
            </Button>
          </div>
        </Card>
      </div>
    );
  }

  const conditionsColumns: ColumnsType<any> = [
    {
      title: '字段',
      dataIndex: 'field',
      key: 'field',
    },
    {
      title: '操作符',
      dataIndex: 'operator',
      key: 'operator',
      render: (operator: string) => {
        const operatorMap: Record<string, string> = {
          eq: '等于',
          ne: '不等于',
          gt: '大于',
          gte: '大于等于',
          lt: '小于',
          lte: '小于等于',
          contains: '包含',
          not_contains: '不包含',
          regex: '正则匹配',
        };
        return operatorMap[operator] || operator;
      },
    },
    {
      title: '值',
      dataIndex: 'value',
      key: 'value',
    },
  ];

  const actionsColumns: ColumnsType<any> = [
    {
      title: '动作类型',
      dataIndex: 'type',
      key: 'type',
      render: (type: string) => {
        const typeMap: Record<string, string> = {
          email: '发送邮件',
          sms: '发送短信',
          webhook: 'Webhook',
          dingtalk: '钉钉通知',
          wechat: '企业微信',
        };
        return typeMap[type] || type;
      },
    },
    {
      title: '配置',
      dataIndex: 'config',
      key: 'config',
      render: (config: Record<string, any>) => (
        <Text code style={{ fontSize: '12px' }}>
          {JSON.stringify(config, null, 2)}
        </Text>
      ),
    },
  ];

  return (
    <div style={{ padding: '24px' }}>
      <Card loading={loading}>
        <div style={{ marginBottom: '24px' }}>
          <Space>
            <Button icon={<ArrowLeftOutlined />} onClick={handleBack}>
              返回
            </Button>
            <Title level={4} style={{ margin: 0 }}>
              规则详情
            </Title>
          </Space>
        </div>

        {currentRule && (
          <>
            <div style={{ marginBottom: '24px' }}>
              <Space>
                <Button
                  type="primary"
                  icon={<EditOutlined />}
                  onClick={handleEdit}
                >
                  编辑
                </Button>
                <Button
                  icon={currentRule.enabled ? <PauseCircleOutlined /> : <PlayCircleOutlined />}
                  onClick={handleToggleStatus}
                >
                  {currentRule.enabled ? '禁用' : '启用'}
                </Button>
                <Button
                  icon={<CopyOutlined />}
                  onClick={handleDuplicate}
                >
                  复制
                </Button>
                <Button
                  danger
                  icon={<DeleteOutlined />}
                  onClick={() => setDeleteModalVisible(true)}
                >
                  删除
                </Button>
              </Space>
            </div>

            <Descriptions
              title="基本信息"
              bordered
              column={2}
              style={{ marginBottom: '24px' }}
            >
              <Descriptions.Item label="规则名称" span={2}>
                {currentRule.name}
              </Descriptions.Item>
              <Descriptions.Item label="描述" span={2}>
                {currentRule.description || '-'}
              </Descriptions.Item>
              <Descriptions.Item label="数据源ID">
                {currentRule.data_source_id}
              </Descriptions.Item>
              <Descriptions.Item label="状态">
                <Switch
                  checked={currentRule.enabled}
                  onChange={handleToggleStatus}
                  checkedChildren="启用"
                  unCheckedChildren="禁用"
                />
              </Descriptions.Item>
              <Descriptions.Item label="创建时间">
                {formatDate(currentRule.created_at)}
              </Descriptions.Item>
              <Descriptions.Item label="更新时间">
                {formatDate(currentRule.updated_at)}
              </Descriptions.Item>
            </Descriptions>

            <Card title="查询语句" style={{ marginBottom: '24px' }}>
              <Text code style={{ whiteSpace: 'pre-wrap' }}>
                {currentRule.query}
              </Text>
            </Card>

            <Card title="触发条件" style={{ marginBottom: '24px' }}>
              <Table
                columns={conditionsColumns}
                dataSource={currentRule.conditions}
                pagination={false}
                size="small"
                rowKey={(record, index) => index?.toString() || '0'}
              />
            </Card>

            <Card title="执行动作">
              <Table
                columns={actionsColumns}
                dataSource={currentRule.actions}
                pagination={false}
                size="small"
                rowKey={(record, index) => index?.toString() || '0'}
              />
            </Card>
          </>
        )}
      </Card>

      <Modal
        title="确认删除"
        open={deleteModalVisible}
        onOk={handleDelete}
        onCancel={() => setDeleteModalVisible(false)}
        okText="确认删除"
        cancelText="取消"
        okButtonProps={{ danger: true }}
      >
        <p>确定要删除规则 "{currentRule?.name}" 吗？</p>
        <p style={{ color: '#ff4d4f' }}>此操作不可撤销，请谨慎操作。</p>
      </Modal>
    </div>
  );
};

export default RuleDetail;