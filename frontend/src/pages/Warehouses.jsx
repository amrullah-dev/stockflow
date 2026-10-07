import { useCallback, useEffect, useState } from 'react'
import { getWarehouses } from '../services/warehouseService'

function StatusBadge({ active }) { return <span className={active ? 'status-badge active' : 'status-badge inactive'}>{active ? 'Active' : 'Inactive'}</span> }

export default function Warehouses() {
  const [warehouses, setWarehouses] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(false)
  const loadWarehouses = useCallback(async () => { setLoading(true); setError(false); try { setWarehouses(await getWarehouses()) } catch { setError(true) } finally { setLoading(false) } }, [])
  useEffect(() => { loadWarehouses() }, [loadWarehouses])

  return <div className="data-page">
    <div className="page-heading data-page-heading"><div><div className="eyebrow">Master Data</div><h2>Warehouses</h2><p className="muted">Manage warehouse locations</p></div></div>
    {loading && <div className="table-card state-card"><div className="spinner" /><strong>Loading warehouses…</strong><span>Fetching the latest warehouse information.</span></div>}
    {!loading && error && <div className="table-card state-card"><div className="state-icon error">!</div><strong>Failed to load warehouses.</strong><span>There was a problem connecting to the API.</span><button className="secondary-button" onClick={loadWarehouses}>Retry</button></div>}
    {!loading && !error && warehouses.length === 0 && <div className="table-card state-card"><div className="state-icon">⌂</div><strong>No warehouses found.</strong><span>Warehouses from the backend will appear here.</span></div>}
    {!loading && !error && warehouses.length > 0 && <div className="table-card table-scroll"><table><thead><tr><th>Code</th><th>Warehouse Name</th><th>Location</th><th>Description</th><th>Status</th></tr></thead><tbody>{warehouses.map((warehouse) => <tr key={warehouse.id}><td className="code-cell">{warehouse.code || '—'}</td><td><strong className="table-name">{warehouse.name || '—'}</strong></td><td>{warehouse.location || '—'}</td><td className="description-cell">{warehouse.description || '—'}</td><td><StatusBadge active={warehouse.is_active} /></td></tr>)}</tbody></table></div>}
  </div>
}
