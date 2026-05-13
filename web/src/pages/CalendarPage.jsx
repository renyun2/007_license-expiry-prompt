import { Badge, Card, Col, List, Row, Select, Space, Typography, message } from 'antd'
import dayjs from 'dayjs'
import { useEffect, useMemo, useState } from 'react'
import { api } from '../api.js'

export default function CalendarPage() {
  const [year, setYear] = useState(dayjs().year())
  const [data, setData] = useState({ by_month: {} })
  const [loading, setLoading] = useState(false)

  useEffect(() => {
    setLoading(true)
    api.calendar(year)
      .then(setData)
      .catch((e) => message.error(e.message))
      .finally(() => setLoading(false))
  }, [year])

  const monthCell = (m) => {
    const items = data.by_month?.[m] || []
    return (
      <List
        size="small"
        dataSource={items.slice(0, 4)}
        renderItem={(it) => (
          <List.Item style={{ padding: '2px 0', border: 'none' }}>
            <Typography.Text ellipsis style={{ fontSize: 12 }}>{it.name}</Typography.Text>
            <Badge
              status={it.status === '已过期' ? 'error' : it.status === '即将到期' ? 'warning' : 'processing'}
              style={{ marginLeft: 8 }}
            />
          </List.Item>
        )}
      />
    )
  }

  const months = useMemo(() => [1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12], [])

  return (
    <Card
      title="年度证书到期日历视图"
      loading={loading}
      extra={(
        <Space>
          <Typography.Text>年份</Typography.Text>
          <Select
            style={{ width: 120 }}
            value={year}
            onChange={setYear}
            options={[2025, 2026, 2027].map((y) => ({ label: y, value: y }))}
          />
        </Space>
      )}
    >
      <Typography.Paragraph type="secondary">
        展示选定自然年内到期的证书（按月份归档）。颜色徽标指示计算状态。
      </Typography.Paragraph>
      <Row gutter={[12, 12]}>
        {months.map((m) => (
          <Col xs={24} sm={12} md={8} lg={6} key={m}>
            <Card size="small" title={`${year}年${m}月`}>
              {monthCell(m)}
            </Card>
          </Col>
        ))}
      </Row>
    </Card>
  )
}
