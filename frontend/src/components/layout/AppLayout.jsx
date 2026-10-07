import { useState } from 'react'
import { Outlet } from 'react-router-dom'
import Sidebar from './Sidebar'
import Topbar from './Topbar'

export default function AppLayout() {
  const [open, setOpen] = useState(false)
  return <div className="app-shell"><div className={open ? 'sidebar-wrap open' : 'sidebar-wrap'}><Sidebar onNavigate={() => setOpen(false)} /></div>{open && <div className="backdrop" onClick={() => setOpen(false)} /> }<main className="main-shell"><Topbar onMenu={() => setOpen(true)} /><section className="content"><Outlet /></section></main></div>
}
