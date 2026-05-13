import { Button, Card, Form, Input, Modal, Select, Space, Table, Tag, Typography, message } from 'antd'
import dayjs from 'dayjs'
import { useEffect, useState } from 'react'
import { api } from '../api.js'

const progressOptions = ['材料准备中', '已提交', '审核中', '已通过']

export default function RenewalsPage() {
  const [rows, setRows] = useState([])
  const [loading, setLoading] = useState(false)
  const [open, setOpen] = useState(false)
  const [form] = Form.useForm()
  const [certs, setCerts] = useState([])

  const load = async () => {
    setLoading(true)
    try {
      const [r, c] = await Promise.all([api.renewals(), api.certificates('')])
      setRows(r)
      setCerts(c)
    } catch (e) {
      message.error(e.message)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    load()
  }, [])

  const create = async () => {
    const v = await form.validateFields()
    const raw = v.materials_checklist || ''
    const mats = raw.split(/[,，]/).map((s) => s.trim()).filter(Boolean)
    const body = {
      certificate_id: v.certificate_id,
      applicant: v.applicant,
      materials_checklist: JSON.stringify(mats),
      progress: v.progress || '材料准备中',
      expected_submit_date: v.expected_submit_date ? dayjs(v.expected_submit_date).toISOString() : null,
    }
    try {
      await api.createRenewal(body)
      message.success('已创建申请')
      setOpen(false)
      form.resetFields()
      load()
    } catch (e) {
      message.error(e.message)
    }
  }

  const patch = async (id, progress) => {
    try {
      await api.patchRenewal(id, { progress })
      message.success('已更新状态')
      load()
    } catch (e) {
      message.error(e.message)
    }
  }

  return (
    <Card
      title="续期 / 年检申请"
      extra={<Button type="primary" onClick={() => setOpen(true)}>新建申请</Button>}
      loading={loading}
    >
      <Typography.Paragraph type="secondary">
        跟踪材料清单与审批进展；当状态置为「已通过」时，可在 PATCH 中附带 new_cert_notes 写入台账备注（演示：请在网络面板手动调用或后续扩展表单）。
      </Typography.Paragraph>
      <Table
        rowKey="id"
        dataSource={rows}
        pagination={{ pageSize: 8 }}
        columns={[
          { title: '证书', dataIndex: ['certificate', 'name'], render: (_, r) => r.certificate?.name },
          { title: '申请人', dataIndex: 'applicant' },
          { title: '申请时间', dataIndex: 'apply_time', render: (v) => dayjs(v).format('YYYY-MM-DD HH:mm') },
          {
            title: '进展',
            dataIndex: 'progress',
            render: (p, r) => (
              <Select
                value={p}
                style={{ width: 140 }}
                options={progressOptions.map((x) => ({ label: x, value: x }))}
                onChange={(v) => patch(r.id, v)}
              />
            ),
          },
        ]}
      />

      <Modal title="新建续期申请" open={open} onCancel={() => setOpen(false)} onOk={create} destroyOnClose width={640}>
        <Form form={form} layout="vertical" initialValues={{ progress: '材料准备中', applicant: '负责人' }}>
          <Form.Item name="certificate_id" label="证书" rules={[{ required: true }]}>
            <Select
              showSearch
              optionFilterProp="label"
              options={certs.map((c) => ({ label: `${c.name} (${c.cert_no})`, value: c.id }))}
            />
          </Form.Item>
          <Form.Item name="applicant" label="申请人/负责人" rules={[{ required: true }]}><Input /></Form.Item>
          <Form.Item name="materials_checklist" label="材料清单（逗号分隔）">
            <Input placeholder="如：营业执照副本,申请表" />
          </Form.Item>
          <Form.Item name="expected_submit_date" label="预计提交日期"><Input placeholder="YYYY-MM-DD" /></Form.Item>
          <Form.Item name="progress" label="当前进展">
            <Select options={progressOptions.map((x) => ({ label: x, value: x }))} />
          </Form.Item>
        </Form>
      </Modal>
    </Card>
  )
}
