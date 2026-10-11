import { useCallback, useEffect, useState } from 'react'
import { ApiError } from '../services/api'
import { getDashboardSummary } from '../services/dashboardService'

const emptySummary = { total_products: 0, total_warehouses: 0, total_suppliers: 0, total_stock_quantity: 0, low_stock_count: 0, recent_transactions: [] }
const numberFormatter = new Intl.NumberFormat(undefined)
const dateFormatter = new Intl.DateTimeFormat(undefined, { month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' })

function formatDate(value) {
  if (!value) return 'Date unavailable'
  const date = new Date(value)
  return Number.isNaN(date.getTime()) ? 'Date unavailable' : dateFormatter.format(date)
}

function transactionLabel(transaction) {
  const type = String(transaction?.type || transaction?.reference_type || 'MOVEMENT').replaceAll('_', ' ')
  return type.toLowerCase().replace(/\b\w/g, (letter) => letter.toUpperCase())
}

export default function Dashboard() {
  const [summary, setSummary] = useState(emptySummary)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  const loadSummary = useCallback(async () => {
    setLoading(true)
    setError('')
    try {
      setSummary(await getDashboardSummary())
    } catch (requestError) {
      if (!(requestError instanceof ApiError && requestError.status === 401)) {
        setError(requestError?.message || 'Dashboard data could not be loaded.')
      }
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => { loadSummary() }, [loadSummary])

  const cards = [
    ['Total Products', summary.total_products, 'Active products'],
    ['Total Warehouses', summary.total_warehouses, 'Active warehouses'],
    ['Total Suppliers', summary.total_suppliers, 'Active suppliers'],
    ['Total Stock', summary.total_stock_quantity, 'Units across warehouses'],
    ['Low Stock', summary.low_stock_count, 'Warehouse-product rows at minimum'],
  ]

  return <div className="dashboard">
    <div className="page-heading">
      <div><div className="eyebrow">Overview</div><h2>Good to see you</h2><p className="muted">Monitor your inventory workspace from one place.</p></div>
      <div className="date-badge">{new Date().toLocaleDateString(undefined, { month: 'short', day: 'numeric', year: 'numeric' })}</div>
    </div>
    {loading ? <div className="state-card"><div className="spinner" /><strong>Loading dashboard</strong><span>Reading current inventory summary.</span></div> : error ? <div className="state-card"><div className="state-icon error">!</div><strong>Dashboard unavailable</strong><span>{error}</span><button className="secondary-button" type="button" onClick={loadSummary}>Try again</button></div> : <>
      <div className="metric-grid">{cards.map(([label, value, note]) => <article className="metric-card" key={label}><div className="metric-icon">✦</div><span>{label}</span><strong>{numberFormatter.format(Number(value) || 0)}</strong><small>{note}</small></article>)}</div>
      <div className="dashboard-grid">
        <section className="empty-card dashboard-transactions"><div className="section-heading"><div><h3>Recent Transactions</h3><p className="muted">Latest posted inventory movements.</p></div><span className="status-dot">Live</span></div>
          {summary.recent_transactions.length === 0 ? <div className="empty-state"><div className="empty-icon">↗</div><strong>No transactions yet</strong><span>Recent inventory activity will appear here.</span></div> : <div className="transaction-list">{summary.recent_transactions.map((transaction) => <div className="transaction-row" key={transaction.id}><div className="transaction-main"><span className={`movement-type movement-${String(transaction.type || '').toLowerCase()}`}>{transactionLabel(transaction)}</span><strong>{transaction.product?.name || 'Product unavailable'}</strong><small>{transaction.warehouse?.name || 'Warehouse unavailable'}</small></div><div className="transaction-meta"><strong>{numberFormatter.format(Number(transaction.quantity) || 0)} {transaction.product?.unit || ''}</strong><small>{formatDate(transaction.created_at)}</small></div></div>)}</div>}
        </section>
        <section className="empty-card dashboard-inventory"><div className="section-heading"><div><h3>Inventory Overview</h3><p className="muted">Stock health at a glance.</p></div><span className="status-dot">Live</span></div><div className="inventory-highlight"><div className="inventory-ring">{numberFormatter.format(Number(summary.total_stock_quantity) || 0)}</div><strong>Total units in stock</strong><span>{numberFormatter.format(Number(summary.low_stock_count) || 0)} warehouse-product rows need attention.</span></div></section>
      </div>
    </>}
  </div>
}
