import {
  AppstoreOutlined,
  CalendarOutlined,
  DollarOutlined,
  FileProtectOutlined,
  HomeOutlined,
  ProfileOutlined,
  SettingOutlined,
} from '@ant-design/icons'
import { Layout, Menu, Typography } from 'antd'
import { useMemo, useState } from 'react'
import { BrowserRouter, Navigate, Route, Routes, useLocation, useNavigate } from 'react-router-dom'
import CalendarPage from './pages/CalendarPage.jsx'
import CertificatesPage from './pages/CertificatesPage.jsx'
import DashboardPage from './pages/DashboardPage.jsx'
import FeesPage from './pages/FeesPage.jsx'
import RemindersPage from './pages/RemindersPage.jsx'
import RenewalsPage from './pages/RenewalsPage.jsx'

const { Header, Sider, Content } = Layout

function Shell() {
  const nav = useNavigate()
  const loc = useLocation()
  const selected = useMemo(() => {
    if (loc.pathname.startsWith('/certs')) return ['certs']
    if (loc.pathname.startsWith('/reminders')) return ['reminders']
    if (loc.pathname.startsWith('/renewals')) return ['renewals']
    if (loc.pathname.startsWith('/fees')) return ['fees']
    if (loc.pathname.startsWith('/calendar')) return ['calendar']
    return ['home']
  }, [loc.pathname])

  const [collapsed, setCollapsed] = useState(false)

  return (
    <Layout style={{ minHeight: '100vh' }}>
      <Sider collapsible collapsed={collapsed} onCollapse={setCollapsed} theme="light" width={220}>
        <div style={{ padding: '16px 16px 8px' }}>
          <Typography.Title level={5} style={{ margin: 0 }}>资质到期管理</Typography.Title>
          <Typography.Text type="secondary" style={{ fontSize: 12 }}>企业证书 · 续期 · 费用</Typography.Text>
        </div>
        <Menu
          mode="inline"
          selectedKeys={selected}
          items={[
            { key: 'home', icon: <HomeOutlined />, label: '合规总览', onClick: () => nav('/') },
            { key: 'certs', icon: <FileProtectOutlined />, label: '证书台账', onClick: () => nav('/certs') },
            { key: 'reminders', icon: <SettingOutlined />, label: '提醒配置', onClick: () => nav('/reminders') },
            { key: 'renewals', icon: <ProfileOutlined />, label: '续期申请', onClick: () => nav('/renewals') },
            { key: 'fees', icon: <DollarOutlined />, label: '费用分析', onClick: () => nav('/fees') },
            { key: 'calendar', icon: <CalendarOutlined />, label: '到期日历', onClick: () => nav('/calendar') },
          ]}
        />
      </Sider>
      <Layout>
        <Header
          style={{
            background: '#fff',
            padding: '0 24px',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between',
            borderBottom: '1px solid #f0f0f0',
          }}
        >
          <Typography.Text strong>企业资质证书与许可证到期管理系统</Typography.Text>
          <Typography.Text type="secondary"><AppstoreOutlined /> 单容器 · SQLite</Typography.Text>
        </Header>
        <Content style={{ margin: 24 }}>
          <Routes>
            <Route path="/" element={<DashboardPage />} />
            <Route path="/certs" element={<CertificatesPage />} />
            <Route path="/reminders" element={<RemindersPage />} />
            <Route path="/renewals" element={<RenewalsPage />} />
            <Route path="/fees" element={<FeesPage />} />
            <Route path="/calendar" element={<CalendarPage />} />
            <Route path="*" element={<Navigate to="/" replace />} />
          </Routes>
        </Content>
      </Layout>
    </Layout>
  )
}

export default function App() {
  return (
    <BrowserRouter>
      <Shell />
    </BrowserRouter>
  )
}
