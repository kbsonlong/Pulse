import React, { useEffect } from 'react';
import { Row, Col, Card, Statistic, Table, Tag, Progress, List, Avatar } from 'antd';
import {
  AlertOutlined,
  CheckCircleOutlined,
  ClockCircleOutlined,
  ExclamationCircleOutlined,
  UserOutlined,
} from '@ant-design/icons';
import { useAlert, useTicket, useUI } from '../hooks';
import { formatDate, getPriorityColor } from '../utils';
import { AlertSeverity, TicketStatus } from '../types';

const Dashboard: React.FC = () => {
  const { alerts, statistics: alertStats, fetchAlerts, fetchStatistics } = useAlert();
  const { tickets, fetchTickets } = useTicket();
  const { setBreadcrumb } = useUI();

  useEffect(() => {
    setBreadcrumb([{ title: '仪表盘' }]);
    fetchStatistics();
    fetchAlerts({ limit: 5, sort: 'created_at', order: 'desc' });
    fetchTickets({ limit: 5, status: 'open' });
  }, [setBreadcrumb, fetchStatistics, fetchAlerts, fetchTickets]);

  // 告警统计数据
  const alertStatsData = [
    {
      title: '总告警数',
      value: alertStats?.total || 0,
      icon: <AlertOutlined style={{ color: '#1890ff' }} />,
    },
    {
      title: '严重告警',
      value: alertStats?.critical || 0,
      icon: <ExclamationCircleOutlined style={{ color: '#ff4d4f' }} />,
    },
    {
      title: '已处理',
      value: alertStats?.resolved || 0,
      icon: <CheckCircleOutlined style={{ color: '#52c41a' }} />,
    },
    {
      title: '待处理',
      value: alertStats?.pending || 0,
      icon: <ClockCircleOutlined style={{ color: '#faad14' }} />,
    },
  ];

  // 告警趋势表格列
  const alertColumns = [
    {
      title: '告警名称',
      dataIndex: 'name',
      key: 'name',
      ellipsis: true,
    },
    {
      title: '严重程度',
      dataIndex: 'severity',
      key: 'severity',
      render: (severity: AlertSeverity) => (
        <Tag color={getPriorityColor(severity)}>
          {severity.toUpperCase()}
        </Tag>
      ),
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (status: string) => (
        <Tag color={status === 'resolved' ? 'green' : 'red'}>
          {status === 'resolved' ? '已解决' : '未解决'}
        </Tag>
      ),
    },
    {
      title: '创建时间',
      dataIndex: 'created_at',
      key: 'created_at',
      render: (date: string) => formatDate(date, 'MM-dd HH:mm'),
    },
  ];

  // 工单列表数据
  const ticketListData = tickets.map((ticket) => ({
    title: ticket.title,
    description: `优先级: ${ticket.priority} | 状态: ${ticket.status}`,
    avatar: <Avatar icon={<UserOutlined />} />,
    content: `创建时间: ${formatDate(ticket.created_at, 'MM-dd HH:mm')}`,
  }));

  return (
    <div>
      {/* 统计卡片 */}
      <Row gutter={[16, 16]} style={{ marginBottom: 24 }}>
        {alertStatsData.map((item, index) => (
          <Col xs={24} sm={12} lg={6} key={index}>
            <Card>
              <Statistic
                title={item.title}
                value={item.value}
                prefix={item.icon}
                valueStyle={{ fontSize: 24 }}
              />
            </Card>
          </Col>
        ))}
      </Row>

      <Row gutter={[16, 16]}>
        {/* 告警处理率 */}
        <Col xs={24} lg={8}>
          <Card title="告警处理率" style={{ height: 300 }}>
            <div style={{ textAlign: 'center' }}>
              <Progress
                type="circle"
                percent={
                  alertStats?.total
                    ? Math.round(
                        ((alertStats.resolved || 0) / alertStats.total) * 100
                      )
                    : 0
                }
                format={(percent) => `${percent}%`}
                size={120}
              />
              <div style={{ marginTop: 16, color: '#666' }}>
                已处理 {alertStats?.resolved || 0} / 总计 {alertStats?.total || 0}
              </div>
            </div>
          </Card>
        </Col>

        {/* 最近告警 */}
        <Col xs={24} lg={16}>
          <Card title="最近告警" style={{ height: 300 }}>
            <Table
              columns={alertColumns}
              dataSource={alerts.slice(0, 5).map((alert, index) => ({
                ...alert,
                key: alert.id || index,
              }))}
              pagination={false}
              size="small"
              scroll={{ y: 200 }}
              locale={{ emptyText: '暂无数据' }}
            />
          </Card>
        </Col>
      </Row>

      <Row gutter={[16, 16]} style={{ marginTop: 16 }}>
        {/* 待处理工单 */}
        <Col xs={24} lg={12}>
          <Card title="待处理工单" style={{ height: 300 }}>
            <List
              itemLayout="horizontal"
              dataSource={ticketListData}
              renderItem={(item) => (
                <List.Item>
                  <List.Item.Meta
                    avatar={item.avatar}
                    title={item.title}
                    description={item.description}
                  />
                  <div>{item.content}</div>
                </List.Item>
              )}
              locale={{ emptyText: '暂无待处理工单' }}
            />
          </Card>
        </Col>

        {/* 系统状态 */}
        <Col xs={24} lg={12}>
          <Card title="系统状态" style={{ height: 300 }}>
            <div style={{ padding: '16px 0' }}>
              <Row gutter={[0, 16]}>
                <Col span={24}>
                  <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                    <span>CPU 使用率</span>
                    <span>65%</span>
                  </div>
                  <Progress percent={65} size="small" />
                </Col>
                <Col span={24}>
                  <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                    <span>内存使用率</span>
                    <span>78%</span>
                  </div>
                  <Progress percent={78} size="small" status="active" />
                </Col>
                <Col span={24}>
                  <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                    <span>磁盘使用率</span>
                    <span>45%</span>
                  </div>
                  <Progress percent={45} size="small" />
                </Col>
                <Col span={24}>
                  <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                    <span>网络延迟</span>
                    <span>12ms</span>
                  </div>
                  <Progress percent={20} size="small" strokeColor="#52c41a" />
                </Col>
              </Row>
            </div>
          </Card>
        </Col>
      </Row>
    </div>
  );
};

export default Dashboard;