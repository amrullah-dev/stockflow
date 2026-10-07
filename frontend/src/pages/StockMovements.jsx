import { useCallback, useEffect, useState } from 'react'
import { getStockMovements } from '../services/stockMovementService'

function errorMessage(error, fallback) {
  return error?.payload?.message || error?.message || fallback
}

function formatDate(value) {
  if (!value) return '-'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return '-'
  return new Intl.DateTimeFormat('id-ID', { dateStyle: 'short', timeStyle: 'short' }).format(date)
}

function formatInteger(value) {
  if (value === null || value === undefined) return '0'
  return new Intl.NumberFormat('id-ID', { maximumFractionDigits: 0 }).format(value)
}

function referenceLabel(movement) {
  if (!movement.reference_type) return '-'
  if (movement.reference_id === null || movement.reference_id === undefined) return movement.reference_type
  return `${movement.reference_type} #${movement.reference_id}`
}

function typeClass(type) {
  return type === 'OUT' ? 'movement-out' : 'movement-in'
}

export default function StockMovements() {
  const [movements, setMovements] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  const loadMovements = useCallback(async () => {
    setLoading(true)
    setError('')
    try {
      setMovements(await getStockMovements())
    } catch (err) {
      setMovements([])
      setError(errorMessage(err, 'Gagal memuat stock movement.'))
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => { loadMovements() }, [loadMovements])

  return <div className="data-page">
    <div className="page-heading data-page-heading">
      <div><div className="eyebrow">Inventory</div><h2>Stock Movement</h2><p className="muted">Riwayat perubahan stok pada seluruh gudang.</p></div>
      <button className="secondary-button movement-refresh-button" disabled={loading} onClick={loadMovements}>{loading ? 'Memuat...' : 'Refresh'}</button>
    </div>
    {loading && <div className="table-card state-card"><div className="spinner" /><strong>Memuat stock movement...</strong><span>Mengambil histori perubahan stok dari backend.</span></div>}
    {!loading && error && <div className="table-card state-card"><div className="state-icon error">!</div><strong>{error}</strong><button className="secondary-button" onClick={loadMovements}>Coba Lagi</button></div>}
    {!loading && !error && movements.length === 0 && <div className="table-card state-card"><div className="state-icon">↕</div><strong>Belum ada stock movement.</strong><span>Histori perubahan stok dari backend akan tampil di sini.</span></div>}
    {!loading && !error && movements.length > 0 && <div className="table-card table-scroll"><table className="movement-table"><thead><tr><th>Waktu</th><th>Warehouse</th><th>Product</th><th>Type</th><th>Quantity</th><th>Balance Before</th><th>Balance After</th><th>Reference</th><th>User</th></tr></thead><tbody>{movements.map((movement) => <tr key={movement.id}><td>{formatDate(movement.created_at)}</td><td><strong className="table-name">{movement.warehouse?.name || '-'}</strong><span className="table-description">{movement.warehouse?.code || '-'}</span></td><td><strong className="table-name">{movement.product?.name || '-'}</strong><span className="table-description">{movement.product?.code || '-'}</span></td><td><span className={`movement-type ${typeClass(movement.type)}`}>{movement.type || '-'}</span></td><td className="quantity-cell">{formatInteger(movement.quantity)}</td><td>{formatInteger(movement.balance_before)}</td><td className="quantity-cell">{formatInteger(movement.balance_after)}</td><td>{referenceLabel(movement)}</td><td>{movement.created_by_user?.name || '-'}</td></tr>)}</tbody></table></div>}
  </div>
}
