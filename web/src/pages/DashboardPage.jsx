import { Card, Col, List, Row, Space, Statistic, Table, Tag, Typography, message } from 'antd'
import dayjs from 'dayjs'
import { useEffect, useState } from 'react'
import { api } from '../api.js'

const statusColor = (s) => {
  if (s === '已过期') return 'error'
  if (s === '即将到期') return 'warning'
  if (s === '已注销') return 'default'
  return 'success'
}

export default function DashboardPage() {
  const [dash, setDash] = useState(null)
  const [urgent, setUrgent] = useState([])
  const [todos, setTodos] = useState([])
  const [loading, setLoading] = useState(true)

  const load = async () => {
    setLoading(true)
    try {
      const [d, u, t] = await Promise.all([api.dashboard(), api.urgent(), api.todos()])
      setDash(d)
      setUrgent(u)
      setTodos(t)
    } catch (e) {
      message.error(e.message || '加载失败')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    load()
  }, [])

  const cats = dash?.by_category ? Object.keys(dash.by_category).sort() : []

  return (
    <Space direction="vertical" size="large" style={{ width: '100%' }}>
      <Typography.Title level={4} style={{ margin: 0 }}>合规视图</Typography.Title>
      <Row gutter={[16, 16]}>
        {['有效', '即将到期', '已过期', '已注销'].map((k) => (
          <Col xs={24} sm={12} md={6} key={k}>
            <Card loading={loading}>
              <Statistic title={k} value={dash?.counts?.[k] ?? 0} />
            </Card>
          </Col>
        ))}
      </Row>

      <Card title="按类别分布" loading={loading}>
        <Table
          pagination={false}
          size="small"
          dataSource={cats.map((c) => ({ key: c, category: c, ...(dash.by_category[c] || {}) }))}
          columns={[
            { title: '类别', dataIndex: 'category' },
            { title: '有效', dataIndex: '有效' },
            { title: '即将到期', dataIndex: '即将到期' },
            { title: '已过期', dataIndex: '已过期' },
            { title: '已注销', dataIndex: '已注销' },
          ]}
        />
      </Card>

      <Card
        title="今日起 90 天内到期（紧迫度排序）"
        extra={<Typography.Text type="secondary">越靠前越紧迫</Typography.Text>}
        loading={loading}
      >
        <Table
          size="small"
          rowKey="id"
          dataSource={urgent}
          pagination={{ pageSize: 8 }}
          columns={[
            { title: '证书名称', dataIndex: 'name' },
            { title: '类别', dataIndex: 'category' },
            { title: '编号', dataIndex: 'cert_no' },
            { title: '到期日', dataIndex: 'expiry_date', render: (v) => dayjs(v).format('YYYY-MM-DD') },
            {
              title: '剩余(天)',
              dataIndex: 'urgency_days',
              render: (v) => <Tag color={v <= 30 ? 'red' : v <= 90 ? 'orange' : 'blue'}>{v}</Tag>,
            },
            { title: '状态', dataIndex: 'computed_status', render: (s) => <Tag color={statusColor(s)}>{s}</Tag> },
            { title: '负责人', dataIndex: 'responsible_person' },
          ]}
        />
      </Card>

      <Card title="到期待办" loading={loading}>
        <List
          dataSource={todos}
          locale={{ emptyText: '暂无待办' }}
          renderItem={(item) => (
            <List.Item
              actions={[
                <a key="done" onClick={async () => {
                  try {
                    await api.patchTodo(item.id, true)
                    message.success('已标记完成')
                    load()
                  } catch (e) {
                    message.error(e.message)
                  }
                }}>完成</a>,
              ]}
            >
              <List.Item.Meta
                title={item.title}
                description={<Space>负责人：{item.assignee} · 截止 {dayjs(item.due_date).format('YYYY-MM-DD')}</Space>}
              />
            </List.Item>
          )}
        />
      </Card>
    </Space>
  )
}
