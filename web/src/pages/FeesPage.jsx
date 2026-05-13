import { Card, Col, DatePicker, Row, Space, Statistic, Table, Typography, message } from 'antd'
import dayjs from 'dayjs'
import { useEffect, useState } from 'react'
import { api } from '../api.js'

export default function FeesPage() {
  const [year, setYear] = useState(dayjs().year())
  const [summary, setSummary] = useState(null)
  const [byCat, setByCat] = useState([])
  const [fees, setFees] = useState([])
  const [loading, setLoading] = useState(false)

  const load = async () => {
    setLoading(true)
    try {
      const [s, b, f] = await Promise.all([
        api.feesSummary(year),
        api.feesByCategory(year),
        api.fees(''),
      ])
      setSummary(s)
      setByCat(b.items || [])
      setFees(f)
    } catch (e) {
      message.error(e.message)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    load()
  }, [year])

  const fmt = (cents) => (Number(cents || 0) / 100).toLocaleString('zh-CN', { style: 'currency', currency: 'CNY' })

  return (
    <Space direction="vertical" size="large" style={{ width: '100%' }}>
      <Card title="费用分析" loading={loading}>
        <Row gutter={16} align="middle">
          <Col>
            <Typography.Text>年度：</Typography.Text>
            <DatePicker picker="year" value={dayjs(String(year), 'YYYY')} onChange={(d) => d && setYear(d.year())} />
          </Col>
          <Col>
            <Statistic title="维护总费用" value={fmt(summary?.total_cents)} />
          </Col>
        </Row>
      </Card>

      <Card title="按证书类别对比" loading={loading}>
        <Table
          pagination={false}
          rowKey="category"
          dataSource={byCat.map((r) => ({ key: r.category, category: r.category, sum: r.sum }))}
          columns={[
            { title: '类别', dataIndex: 'category' },
            { title: '费用合计', dataIndex: 'sum', render: (v) => fmt(v) },
          ]}
        />
      </Card>

      <Card title="费用明细（近记录）" loading={loading}>
        <Table
          rowKey="id"
          dataSource={fees.slice(0, 40)}
          pagination={false}
          columns={[
            { title: '日期', dataIndex: 'fee_date', render: (v) => dayjs(v).format('YYYY-MM-DD') },
            { title: '证书', dataIndex: ['certificate', 'name'], render: (_, r) => r.certificate?.name },
            { title: '行政', dataIndex: 'admin_fee_cents', render: fmt },
            { title: '代理', dataIndex: 'agency_fee_cents', render: fmt },
            { title: '备注', dataIndex: 'description' },
          ]}
        />
      </Card>
    </Space>
  )
}
