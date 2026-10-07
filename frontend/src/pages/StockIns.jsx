import { useCallback, useEffect, useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { deleteStockIn, getStockIns, postStockIn } from '../services/stockInService'

function statusClass(status) {
  return status === 'POSTED' ? 'posted' : 'draft'
}

function formatDate(value) {
  if (!value) return '—'
  return new Intl.DateTimeFormat('id-ID', { dateStyle: 'medium' }).format(new Date(value))
}

function errorMessage(error, fallback) {
  return error?.payload?.message || error?.message || fallback
}

export default function StockIns() {
  const navigate = useNavigate()
  const [rows, setRows] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [actionID, setActionID] = useState(null)

  const load = useCallback(async () => {
    setLoading(true)
    setError('')
    try {
      setRows(await getStockIns())
    } catch (err) {
      setError(errorMessage(err, 'Gagal memuat Stock IN.'))
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => { load() }, [load])

  async function remove(row) {
    if (!window.confirm('Apakah Anda yakin ingin menghapus Stock IN ini?')) return
    setActionID(`delete-${row.id}`)
    setError('')
    try {
      await deleteStockIn(row.id)
      await load()
    } catch (err) {
      setError(errorMessage(err, 'Gagal menghapus Stock IN.'))
    } finally {
      setActionID(null)
    }
  }

  async function post(row) {
    if (!window.confirm('Post Stock IN? Setelah diposting, transaksi tidak dapat diedit atau dihapus lagi.')) return
    setActionID(`post-${row.id}`)
    setError('')
    try {
      await postStockIn(row.id)
      await load()
    } catch (err) {
      setError(errorMessage(err, 'Gagal memproses posting Stock IN.'))
    } finally {
      setActionID(null)
    }
  }

  return <div className="data-page">
    <div className="page-heading data-page-heading">
      <div><div className="eyebrow">Transaksi</div><h2>Stock IN</h2><p className="muted">Kelola penerimaan stok dari supplier.</p></div>
      <Link className="primary-small" to="/stock-ins/new">+ Tambah Stock IN</Link>
    </div>
    {error && <div className="form-error stock-in-page-error">{error}</div>}
    {loading && <div className="table-card state-card"><div className="spinner" /><strong>Memuat data Stock IN...</strong><span>Mengambil transaksi terbaru dari backend.</span></div>}
    {!loading && rows.length === 0 && <div className="table-card state-card"><div className="state-icon">↓</div><strong>Belum ada transaksi Stock IN.</strong><span>Tambahkan draft Stock IN untuk mulai mencatat penerimaan stok.</span><Link className="secondary-button" to="/stock-ins/new">+ Tambah Stock IN</Link></div>}
    {!loading && rows.length > 0 && <div className="table-card table-scroll"><table className="stock-in-table"><thead><tr><th>Nomor Referensi</th><th>Tanggal</th><th>Supplier</th><th>Gudang</th><th>Status</th><th>Aksi</th></tr></thead><tbody>{rows.map((row) => <tr key={row.id}><td className="code-cell">{row.reference_number || '—'}</td><td>{formatDate(row.transaction_date)}</td><td>{row.supplier?.name || '—'}</td><td>{row.warehouse?.name || '—'}</td><td><span className={`transaction-status ${statusClass(row.status)}`}>{row.status}</span></td><td><div className="row-actions"><Link to={`/stock-ins/${row.id}`}>Lihat</Link>{row.status === 'DRAFT' && <><Link to={`/stock-ins/${row.id}/edit`}>Edit</Link><button className="danger-link" disabled={actionID !== null} onClick={() => remove(row)}>{actionID === `delete-${row.id}` ? 'Menghapus...' : 'Hapus'}</button><button disabled={actionID !== null} onClick={() => post(row)}>{actionID === `post-${row.id}` ? 'Posting...' : 'Post'}</button></>}</div></td></tr>)}</tbody></table></div>}
  </div>
}
