import { useLocation } from 'react-router-dom'
import { currentUser, logout } from '../../services/auth'

export default function Topbar({ onMenu }) {
  const location = useLocation()
  const user = currentUser()
  const title = location.pathname === '/dashboard' ? 'Dashboard' : location.pathname.slice(1).split('/')[0].replaceAll('-', ' ')
  const displayName = user?.name || user?.email || 'Administrator'
  return <header className="topbar">
    <button className="mobile-menu" onClick={onMenu} aria-label="Open navigation">☰</button>
    <div><div className="eyebrow">Workspace</div><h1>{title.replace(/\b\w/g, (c) => c.toUpperCase())}</h1></div>
    <div className="topbar-actions"><div className="user-chip"><div className="avatar">{displayName.charAt(0).toUpperCase()}</div><span>{displayName}</span></div><button className="logout-button" onClick={() => { logout(); window.location.assign('/login') }}>Logout</button></div>
  </header>
}
