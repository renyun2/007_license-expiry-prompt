import { Button, Card, Descriptions, Drawer, Form, Input, Modal, Select, Space, Switch, Table, Typography, message } from 'antd'
import dayjs from 'dayjs'
import { useEffect, useState } from 'react'
import { api } from '../api.js'

const categories = [
  '营业执照', '行业许可证', '认证证书', '安全生产许可', '特种经营许可', '资质等级证书',
]

export default function CertificatesPage() {
  const [rows, setRows] = useState([])
  const [loading, setLoading] = useState(false)
  const [cat, setCat] = useState('')
  const [open, setOpen] = useState(false)
  const [editing, setEditing] = useState(null)
  const [form] = Form.useForm()
  const [detail, setDetail] = useState(null)
  const [inspOpen, setInspOpen] = useState(false)
  const [inspections, setInspections] = useState([])

  const load = async () => {
    setLoading(true)
    try {
      const q = cat ? `category=${encodeURIComponent(cat)}` : ''
      setRows(await api.certificates(q))
    } catch (e) {
      message.error(e.message)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    load()
  }, [cat])

  const startCreate = () => {
    setEditing(null)
    form.resetFields()
    setOpen(true)
  }

  const startEdit = (r) => {
    setEditing(r)
    form.setFieldsValue({
      ...r,
      issue_date: dayjs(r.issue_date).format('YYYY-MM-DD'),
      expiry_date: dayjs(r.expiry_date).format('YYYY-MM-DD'),
    })
    setOpen(true)
  }

  const save = async () => {
    const v = await form.validateFields()
    const body = {
      ...v,
      issue_date: dayjs(v.issue_date).startOf('day').toISOString(),
      expiry_date: dayjs(v.expiry_date).startOf('day').toISOString(),
    }
    try {
      if (editing) await api.updateCert(editing.id, { ...editing, ...body })
      else await api.createCert(body)
      message.success('已保存')
      setOpen(false)
      load()
    } catch (e) {
      message.error(e.message)
    }
  }

  const remove = async (r) => {
    Modal.confirm({
      title: '确认删除该证书？',
      onOk: async () => {
        await api.deleteCert(r.id)
        message.success('已删除')
        load()
      },
    })
  }

  const openDetail = async (r) => {
    setDetail(r)
    try {
      setInspections(await api.inspections(r.id))
    } catch {
      setInspections([])
    }
    setInspOpen(true)
  }

  return (
    <Card
      title="证书台账"
      extra={
        <Space>
          <Select
            allowClear
            placeholder="按类别筛选"
            style={{ width: 200 }}
            value={cat || undefined}
            onChange={setCat}
            options={categories.map((c) => ({ label: c, value: c }))}
          />
          <Button onClick={() => api.exportCsv(cat)}>导出 CSV</Button>
          <Button type="primary" onClick={startCreate}>新建证书</Button>
        </Space>
      }
    >
      <Table
        loading={loading}
        rowKey="id"
        dataSource={rows}
        pagination={{ pageSize: 10 }}
        rowClassName={(r) => (r.computed_status === '已过期' ? 'row-expired' : '')}
        columns={[
          { title: '名称', dataIndex: 'name' },
          { title: '类别', dataIndex: 'category', width: 120 },
          { title: '编号', dataIndex: 'cert_no', width: 140 },
          { title: '到期日', dataIndex: 'expiry_date', width: 120, render: (v) => dayjs(v).format('YYYY-MM-DD') },
          { title: '部门', dataIndex: 'department', width: 120 },
          { title: '负责人', dataIndex: 'responsible_person', width: 100 },
          {
            title: '状态',
            dataIndex: 'computed_status',
            width: 100,
            render: (s) => (
              <span style={{ color: s === '已过期' ? '#cf1322' : undefined }}>{s}</span>
            ),
          },
          {
            title: '操作',
            key: 'op',
            width: 200,
            render: (_, r) => (
              <Space>
                <Button type="link" onClick={() => openDetail(r)}>附件/年检</Button>
                <Button type="link" onClick={() => startEdit(r)}>编辑</Button>
                <Button type="link" danger onClick={() => remove(r)}>删除</Button>
              </Space>
            ),
          },
        ]}
      />
      <Modal
        title={editing ? '编辑证书' : '新建证书'}
        open={open}
        onCancel={() => setOpen(false)}
        onOk={save}
        width={720}
        destroyOnClose
      >
        <Form form={form} layout="vertical" initialValues={{ is_cancelled: false }}>
          <Form.Item name="name" label="证书名称" rules={[{ required: true }]}><Input /></Form.Item>
          <Form.Item name="category" label="类别" rules={[{ required: true }]}>
            <Select options={categories.map((c) => ({ label: c, value: c }))} />
          </Form.Item>
          <Form.Item name="issuing_org" label="颁证机构"><Input /></Form.Item>
          <Form.Item name="cert_no" label="证书编号"><Input /></Form.Item>
          <Form.Item name="issue_date" label="发证日期" rules={[{ required: true }]}><Input placeholder="YYYY-MM-DD" /></Form.Item>
          <Form.Item name="expiry_date" label="有效期至" rules={[{ required: true }]}><Input placeholder="YYYY-MM-DD" /></Form.Item>
          <Form.Item name="department" label="责任部门"><Input /></Form.Item>
          <Form.Item name="responsible_person" label="负责人"><Input /></Form.Item>
          <Form.Item name="is_cancelled" label="已注销" valuePropName="checked"><Switch /></Form.Item>
          <Form.Item name="scan_url" label="扫描件 URL"><Input /></Form.Item>
          <Form.Item name="front_image_url" label="正面图 URL"><Input /></Form.Item>
          <Form.Item name="back_image_url" label="背面图 URL"><Input /></Form.Item>
          <Form.Item name="notes" label="备注"><Input.TextArea rows={3} /></Form.Item>
        </Form>
      </Modal>

      <Drawer title={detail ? `${detail.name} · 附件与年检` : ''} width={640} open={inspOpen} onClose={() => setInspOpen(false)}>
        {detail && (
          <Space direction="vertical" size="large" style={{ width: '100%' }}>
            <Descriptions bordered size="small" column={1}>
              <Descriptions.Item label="扫描件">{detail.scan_url || '-'}</Descriptions.Item>
              <Descriptions.Item label="正面">{detail.front_image_url || '-'}</Descriptions.Item>
              <Descriptions.Item label="背面">{detail.back_image_url || '-'}</Descriptions.Item>
            </Descriptions>
            <Typography.Title level={5}>历届年检</Typography.Title>
            <Table
              size="small"
              rowKey="id"
              dataSource={inspections}
              pagination={false}
              columns={[
                { title: '年度', dataIndex: 'year', width: 80 },
                { title: '记录链接', dataIndex: 'record_url' },
                { title: '备注', dataIndex: 'notes' },
              ]}
            />
          </Space>
        )}
      </Drawer>
      <style>{`.row-expired { background: #fff1f0 !important; }`}</style>
    </Card>
  )
}
