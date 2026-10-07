import { useCallback, useEffect, useState } from 'react'
import { Link, useNavigate, useParams } from 'react-router-dom'
import { deleteStockIn, getStockIn, postStockIn } from '../services/stockInService'

function formatDate(value, withTime = false) {
  if (!value) return '—'
  return new Intl.DateTimeFormat('id-ID', withTime ? { dateStyle: 'medium', timeStyle: 'short' } : { dateStyle: 'medium' }).format(new Date(value))
}

function errorMessage(error, fallback) { return error?.payload?.message || error?.message || fallback }

export default function StockInDetail() {
  const { id } = useParams()
  const navigate = useNavigate()
  const [stockIn, setStockIn] = useState(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [action, setAction] = useState('')

  const load = useCallback(async () => {
    setLoading(true); setError('')
    try { setStockIn(await getStockIn(id)) } catch (err) { setError(err.status === 404 ? 'Stock IN tidak ditemukan.' : errorMessage(err, 'Gagal memuat detail Stock IN.')) } finally { setLoading(false) }
  }, [id])

  useEffect(() => { load() }, [load])

  async function remove() {
    if (!window.confirm('Apakah Anda yakin ingin menghapus Stock IN ini?')) return
    setAction('delete')
    try { await deleteStockIn(id); navigate('/stock-ins') } catch (err) { setError(errorMessage(err, 'Gagal menghapus Stock IN.')); setAction('') }
  }

  async function post() {
    if (!window.confirm('Post Stock IN? Setelah diposting, transaksi tidak dapat diedit atau dihapus lagi.')) return
    setAction('post')
    try { await postStockIn(id); await load() } catch (err) { setError(errorMessage(err, 'Gagal memproses posting Stock IN.')) } finally { setAction('') }
  }

  if (loading) return <div className="table-card state-card"><div className="spinner" /><strong>Memuat detail...</strong><span>Mengambil data Stock IN.</span></div>
  if (error && !stockIn) return <div className="table-card state-card"><div className="state-icon error">!</div><strong>{error}</strong><button className="secondary-button" onClick={load}>Coba Lagi</button><Link className="secondary-button" to="/stock-ins">Kembali</Link></div>

  const total = stockIn.items.reduce((sum, item) => sum + Number(item.subtotal || 0), 0)
  const draft = stockIn.status === 'DRAFT'
  return <div className="data-page"><div className="page-heading data-page-heading"><div><div className="eyebrow">Transaksi Stock IN</div><h2>{stockIn.reference_number}</h2><p className="muted">Detail penerimaan stok dan status transaksi.</p></div><div className="detail-actions"><Link className="secondary-button" to="/stock-ins">Kembali</Link>{draft && <><Link className="secondary-button" to={`/stock-ins/${id}/edit`}>Edit</Link><button className="secondary-button danger-button" disabled={Boolean(action)} onClick={remove}>{action === 'delete' ? 'Menghapus...' : 'Hapus'}</button><button className="primary-small" disabled={Boolean(action)} onClick={post}>{action === 'post' ? 'Posting...' : 'Post'}</button></>}</div></div>{error && <div className="form-error">{error}</div>}<section className="detail-grid"><div className="form-card"><div className="table-title"><strong>Informasi transaksi</strong><span className={`transaction-status ${draft ? 'draft' : 'posted'}`}>{stockIn.status}</span></div><div className="detail-fields"><div><span>Nomor Referensi</span><strong>{stockIn.reference_number}</strong></div><div><span>Tanggal Transaksi</span><strong>{formatDate(stockIn.transaction_date)}</strong></div><div><span>Supplier</span><strong>{stockIn.supplier?.name || '—'}</strong><small>{stockIn.supplier?.code || ''}</small></div><div><span>Gudang</span><strong>{stockIn.warehouse?.name || '—'}</strong><small>{stockIn.warehouse?.code || ''}</small></div><div><span>Dibuat Oleh</span><strong>{stockIn.created_by_user?.name || '—'}</strong><small>{stockIn.created_by_user?.email || ''}</small></div><div><span>Diposting Pada</span><strong>{formatDate(stockIn.posted_at, true)}</strong></div><div className="full-field"><span>Catatan</span><strong>{stockIn.note || '—'}</strong></div></div></div><div className="form-card detail-summary"><div className="table-title"><strong>Ringkasan</strong><span>{stockIn.items.length} produk</span></div><div className="summary-value"><span>Total</span><strong>{total.toLocaleString('id-ID', { minimumFractionDigits: 2 })}</strong></div></div></section><section className="form-card stock-in-detail-items"><div className="table-title"><strong>Produk</strong><span>{stockIn.items.length} item</span></div><div className="table-scroll"><table><thead><tr><th>Produk</th><th>Jumlah</th><th>Harga Satuan</th><th>Subtotal</th></tr></thead><tbody>{stockIn.items.map((item) => <tr key={item.id}><td><strong className="table-name">{item.product?.name || '—'}</strong><span className="table-description">{item.product?.code || '—'}{item.product?.unit ? ` · ${item.product.unit}` : ''}</span></td><td>{item.quantity}</td><td>{Number(item.unit_price || 0).toLocaleString('id-ID', { minimumFractionDigits: 2 })}</td><td>{Number(item.subtotal || 0).toLocaleString('id-ID', { minimumFractionDigits: 2 })}</td></tr>)}</tbody></table></div></section></div>
}
