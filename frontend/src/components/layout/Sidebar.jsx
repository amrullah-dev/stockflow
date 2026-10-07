import { NavLink } from 'react-router-dom'

const groups = [
  { label: 'Main', items: [['/dashboard', 'Dashboard', '▦']] },
  {
    label: 'Master Data',
    items: [['/products', 'Products', '◈'], ['/categories', 'Categories', '⌘'], ['/suppliers', 'Suppliers', '♙'], ['/warehouses', 'Warehouses', '⌂']],
  },
  { label: 'Inventory', items: [['/stocks', 'Stock Overview', '▤'], ['/stocks/movements', 'Stock Movement', '↕']] },
  { label: 'Transactions', items: [['/stock-ins', 'Stock IN', '↓'], ['/stock-outs', 'Stock OUT', '↑'], ['/transfers', 'Transfer', '⇄']] },
  { label: 'System', items: [['/audit-logs', 'Audit Logs', '◷']] },
]

export default function Sidebar({ onNavigate }) {
  return (
    <aside className="sidebar">
      <div className="brand">
        <div className="brand-mark">S</div>
        <div><strong>StockFlow</strong><span>Inventory Management System</span></div>
      </div>
      <nav>
        {groups.map((group) => <div className="nav-group" key={group.label}>
          <span className="nav-label">{group.label}</span>
          {group.items.map(([to, label, icon]) => <NavLink key={to} to={to} onClick={onNavigate} className={({ isActive }) => isActive ? 'nav-item active' : 'nav-item'}><i>{icon}</i><span>{label}</span></NavLink>)}
        </div>)}
      </nav>
      <div className="sidebar-footer">© {new Date().getFullYear()} StockFlow</div>
    </aside>
  )
}
