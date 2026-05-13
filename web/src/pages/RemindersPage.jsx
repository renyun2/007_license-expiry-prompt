import { Card, Form, Input, InputNumber, message, Modal, Table, Typography } from 'antd'
import { useEffect, useState } from 'react'
import { api } from '../api.js'

export default function RemindersPage() {
  const [rows, setRows] = useState([])
  const [loading, setLoading] = useState(false)
  const [edit, setEdit] = useState(null)
  const [form] = Form.useForm()

  const load = async () => {
    setLoading(true)
    try {
      setRows(await api.reminders())
    } catch (e) {
      message.error(e.message)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    load()
  }, [])

  const save = async () => {
    const v = await form.validateFields()
    try {
      await api.updateReminder(edit.id, { ...edit, ...v })
      message.success('已更新')
      setEdit(null)
      load()
    } catch (e) {
      message.error(e.message)
    }
  }

  return (
    <Card title="提醒配置" loading={loading}>
      <Typography.Paragraph type="secondary">
        为每类证书配置 180 / 90 / 30 天三级提醒窗口；台账中“即将到期”按一级窗口（默认 180 天）判定。可设置默认负责人用于生成待办。
      </Typography.Paragraph>
      <Table
        rowKey="id"
        dataSource={rows}
        pagination={false}
        columns={[
          { title: '类别', dataIndex: 'category', width: 160 },
          { title: '一级(天)', dataIndex: 'days_tier1', width: 100 },
          { title: '二级(天)', dataIndex: 'days_tier2', width: 100 },
          { title: '三级(天)', dataIndex: 'days_tier3', width: 100 },
          { title: '默认负责人', dataIndex: 'default_responsible' },
          {
            title: '操作',
            width: 100,
            render: (_, r) => <a onClick={() => { setEdit(r); form.setFieldsValue(r) }}>编辑</a>,
          },
        ]}
      />

      <Modal title="编辑提醒阈值" open={!!edit} onCancel={() => setEdit(null)} onOk={save} destroyOnClose>
        <Form form={form} layout="vertical">
          <Form.Item name="days_tier1" label="一级提醒（天）" rules={[{ required: true }]}><InputNumber min={1} style={{ width: '100%' }} /></Form.Item>
          <Form.Item name="days_tier2" label="二级提醒（天）" rules={[{ required: true }]}><InputNumber min={1} style={{ width: '100%' }} /></Form.Item>
          <Form.Item name="days_tier3" label="三级提醒（天）" rules={[{ required: true }]}><InputNumber min={1} style={{ width: '100%' }} /></Form.Item>
          <Form.Item name="default_responsible" label="默认负责人"><Input /></Form.Item>
        </Form>
      </Modal>
    </Card>
  )
}
