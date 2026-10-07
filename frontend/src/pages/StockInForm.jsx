import { useEffect, useMemo, useState } from 'react'
import { Link, useNavigate, useParams } from 'react-router-dom'
import { getProducts } from '../services/productService'
import { getSuppliers } from '../services/supplierService'
import { getWarehouses } from '../services/warehouseService'
import { createStockIn, getStockIn, updateStockIn } from '../services/stockInService'

const newItem = () => ({ product_id: '', quantity: '', unit_price: '' })
const emptyForm = { transaction_date: '', supplier_id: '', warehouse_id: '', note: '', items: [newItem()] }

function dateValue(value) { return value ? String(value).slice(0, 10) : '' }
function errorMessage(error, fallback) { return error?.payload?.message || error?.message || fallback }
function itemSubtotal(item) { return (Number(item.quantity) || 0) * (Number(item.unit_price) || 0) }

export default function StockInForm() {
  const { id } = useParams()
  const editing = Boolean(id)
  const navigate = useNavigate()
  const [form, setForm] = useState(emptyForm)
  const [masters, setMasters] = useState({ products: [], suppliers: [], warehouses: [] })
  const [loading, setLoading] = useState(editing)
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState('')
  const [validation, setValidation] = useState('')
  const [referenceNumber, setReferenceNumber] = useState('')

  useEffect(() => {
    let active = true
    async function load() {
      setLoading(true)
      try {
        const requests = [getProducts(), getSuppliers(), getWarehouses()]
        if (editing) requests.push(getStockIn(id))
        const [products, suppliers, warehouses, detail] = await Promise.all(requests)
        if (!active) return
        setMasters({ products, suppliers, warehouses })
        if (detail) {
          setReferenceNumber(detail.reference_number || '')
          setForm({
            transaction_date: dateValue(detail.transaction_date),
            supplier_id: String(detail.supplier_id || ''),
            warehouse_id: String(detail.warehouse_id || ''),
            note: detail.note || '',
            items: (detail.items || []).map((item) => ({ product_id: String(item.product_id || ''), quantity: String(item.quantity ?? ''), unit_price: String(item.unit_price ?? '') })),
          })
          if (detail.status !== 'DRAFT') setError('Stock IN ini sudah POSTED dan tidak dapat diedit.')
        }
      } catch (err) {
        if (active) setError(errorMessage(err, 'Gagal memuat data form Stock IN.'))
      } finally {
        if (active) setLoading(false)
      }
    }
    load()
    return () => { active = false }
  }, [editing, id])

  const total = useMemo(() => form.items.reduce((sum, item) => sum + itemSubtotal(item), 0), [form.items])

  function setField(field, value) { setForm((current) => ({ ...current, [field]: value })) }
  function setItem(index, field, value) { setForm((current) => ({ ...current, items: current.items.map((item, itemIndex) => itemIndex === index ? { ...item, [field]: value } : item) })) }
  function addItem() { setForm((current) => ({ ...current, items: [...current.items, newItem()] })) }
  function removeItem(index) { setForm((current) => ({ ...current, items: current.items.filter((_, itemIndex) => itemIndex !== index) })) }

  function validate() {
    if (!form.transaction_date || !/^\d{4}-\d{2}-\d{2}$/.test(form.transaction_date)) return 'Tanggal transaksi wajib menggunakan format YYYY-MM-DD.'
    if (!form.supplier_id) return 'Supplier wajib dipilih.'
    if (!form.warehouse_id) return 'Gudang wajib dipilih.'
    if (form.items.length === 0) return 'Minimal satu produk harus ditambahkan.'
    const selected = new Set()
    for (const item of form.items) {
      if (!item.product_id) return 'Semua item harus memiliki produk.'
      if (selected.has(item.product_id)) return 'Produk tidak boleh dipilih lebih dari satu kali.'
      selected.add(item.product_id)
      if (!Number.isInteger(Number(item.quantity)) || Number(item.quantity) <= 0) return 'Jumlah setiap produk harus lebih besar dari nol.'
      if (item.unit_price === '' || Number(item.unit_price) < 0 || Number.isNaN(Number(item.unit_price))) return 'Harga satuan harus nol atau lebih.'
    }
    return ''
  }

  async function submit(event) {
    event.preventDefault()
    const message = validate()
    if (message) { setValidation(message); return }
    setSaving(true); setError(''); setValidation('')
    const payload = {
      transaction_date: form.transaction_date,
      supplier_id: Number(form.supplier_id),
      warehouse_id: Number(form.warehouse_id),
      note: form.note,
      items: form.items.map((item) => ({ product_id: Number(item.product_id), quantity: Number(item.quantity), unit_price: Number(item.unit_price) })),
    }
    try {
      const result = editing ? await updateStockIn(id, payload) : await createStockIn(payload)
      navigate(`/stock-ins/${result.id || id}`)
    } catch (err) {
      setError(errorMessage(err, 'Gagal menyimpan draft Stock IN.'))
    } finally { setSaving(false) }
  }

  if (loading) return <div className="table-card state-card"><div className="spinner" /><strong>Memuat form...</strong><span>Menyiapkan data master StockFlow.</span></div>
  return <div className="data-page"><div className="page-heading data-page-heading"><div><div className="eyebrow">Transaksi</div><h2>{editing ? 'Edit Stock IN' : 'Tambah Stock IN'}</h2><p className="muted">Simpan transaksi sebagai draft sebelum diposting.</p>{referenceNumber && <span className="form-reference">Nomor Referensi: {referenceNumber}</span>}</div></div>{error && <div className="form-error">{error}</div>}{!error || !editing ? <form onSubmit={submit} className="stock-in-form"><section className="form-card"><div className="table-title"><strong>Informasi transaksi</strong><span>Data penerimaan</span></div><div className="master-form form-grid"><label>Tanggal Transaksi<input type="date" value={form.transaction_date} onChange={(e) => setField('transaction_date', e.target.value)} /></label><label>Supplier<select value={form.supplier_id} onChange={(e) => setField('supplier_id', e.target.value)}><option value="">Pilih supplier</option>{masters.suppliers.map((supplier) => <option key={supplier.id} value={supplier.id}>{supplier.code} - {supplier.name}</option>)}</select></label><label>Gudang<select value={form.warehouse_id} onChange={(e) => setField('warehouse_id', e.target.value)}><option value="">Pilih gudang</option>{masters.warehouses.map((warehouse) => <option key={warehouse.id} value={warehouse.id}>{warehouse.code} - {warehouse.name}</option>)}</select></label><label className="full-field">Catatan<textarea rows="3" value={form.note} onChange={(e) => setField('note', e.target.value)} /></label></div></section><section className="form-card stock-in-items-card"><div className="table-title"><strong>Daftar produk</strong><button type="button" className="secondary-button compact-button" onClick={addItem}>+ Tambah Produk</button></div><div className="item-list">{form.items.map((item, index) => <div className="stock-in-item" key={index}><div className="item-number">{index + 1}</div><label>Produk<select value={item.product_id} onChange={(e) => setItem(index, 'product_id', e.target.value)}><option value="">Pilih produk</option>{masters.products.map((product) => <option key={product.id} value={product.id}>{product.code} - {product.name}</option>)}</select></label><label>Jumlah<input type="number" min="1" step="1" value={item.quantity} onChange={(e) => setItem(index, 'quantity', e.target.value)} /></label><label>Harga Satuan<input type="number" min="0" step="0.01" value={item.unit_price} onChange={(e) => setItem(index, 'unit_price', e.target.value)} /></label><div className="subtotal-preview"><span>Subtotal</span><strong>{itemSubtotal(item).toLocaleString('id-ID', { minimumFractionDigits: 2 })}</strong></div>{form.items.length > 1 && <button type="button" className="remove-item" onClick={() => removeItem(index)}>Hapus</button>}</div>)}</div><div className="stock-in-total"><span>Total</span><strong>{total.toLocaleString('id-ID', { minimumFractionDigits: 2 })}</strong></div></section>{validation && <div className="form-error">{validation}</div>}<div className="form-actions stock-in-form-actions"><button className="primary-small" disabled={saving}>{saving ? 'Menyimpan...' : 'Simpan Draft'}</button><Link className="secondary-button" to={editing ? `/stock-ins/${id}` : '/stock-ins'}>Batal</Link></div></form> : <Link className="secondary-button" to={`/stock-ins/${id}`}>Kembali ke Detail</Link>}</div>
}
