import { useCallback, useEffect, useState } from 'react'
import { getWarehouses } from '../services/warehouseService'
import { getWarehouseStocks } from '../services/stockService'

function errorMessage(error, fallback) {
  return error?.payload?.message || error?.message || fallback
}

function stockStatus(quantity, minimumStock) {
  if (quantity <= 0) return { label: 'Habis', className: 'stock-empty' }
  if (quantity <= minimumStock) return { label: 'Minimum', className: 'stock-minimum' }
  return { label: 'Aman', className: 'stock-safe' }
}

function formatQuantity(value) {
  return new Intl.NumberFormat('id-ID').format(value)
}

export default function Stocks() {
  const [warehouses, setWarehouses] = useState([])
  const [selectedWarehouseID, setSelectedWarehouseID] = useState('')
  const [warehouseLoading, setWarehouseLoading] = useState(true)
  const [warehouseError, setWarehouseError] = useState('')
  const [stocks, setStocks] = useState([])
  const [stockLoading, setStockLoading] = useState(false)
  const [stockError, setStockError] = useState('')

  const loadWarehouses = useCallback(async () => {
    setWarehouseLoading(true)
    setWarehouseError('')
    try {
      const data = await getWarehouses()
      setWarehouses(data)
      setSelectedWarehouseID((current) => current || (data[0] ? String(data[0].id) : ''))
    } catch (error) {
      setWarehouses([])
      setSelectedWarehouseID('')
      setWarehouseError(errorMessage(error, 'Gagal memuat warehouse.'))
    } finally {
      setWarehouseLoading(false)
    }
  }, [])

  const loadStocks = useCallback(async () => {
    if (!selectedWarehouseID) {
      setStocks([])
      return
    }
    setStockLoading(true)
    setStockError('')
    try {
      setStocks(await getWarehouseStocks(selectedWarehouseID))
    } catch (error) {
      setStocks([])
      setStockError(errorMessage(error, 'Gagal memuat data stok.'))
    } finally {
      setStockLoading(false)
    }
  }, [selectedWarehouseID])

  useEffect(() => { loadWarehouses() }, [loadWarehouses])
  useEffect(() => { loadStocks() }, [loadStocks])

  return <div className="data-page">
    <div className="page-heading data-page-heading">
      <div><div className="eyebrow">Inventory</div><h2>Stocks</h2><p className="muted">Pantau stok produk berdasarkan warehouse.</p></div>
      <button className="secondary-button stock-refresh-button" disabled={!selectedWarehouseID || stockLoading} onClick={loadStocks}>{stockLoading ? 'Memuat...' : 'Refresh'}</button>
    </div>
    <div className="stock-toolbar">
      <label htmlFor="warehouse-select">Warehouse<select id="warehouse-select" value={selectedWarehouseID} disabled={warehouseLoading || !!warehouseError || warehouses.length === 0} onChange={(event) => setSelectedWarehouseID(event.target.value)}>
        <option value="">Pilih warehouse</option>
        {warehouses.map((warehouse) => <option key={warehouse.id} value={warehouse.id}>{warehouse.name} ({warehouse.code})</option>)}
      </select></label>
    </div>
    {warehouseLoading && <div className="table-card state-card"><div className="spinner" /><strong>Memuat warehouse...</strong><span>Mengambil daftar warehouse dari backend.</span></div>}
    {!warehouseLoading && warehouseError && <div className="table-card state-card"><div className="state-icon error">!</div><strong>{warehouseError}</strong><button className="secondary-button" onClick={loadWarehouses}>Coba Lagi</button></div>}
    {!warehouseLoading && !warehouseError && selectedWarehouseID && stockLoading && <div className="table-card state-card"><div className="spinner" /><strong>Memuat stok...</strong><span>Mengambil stok warehouse yang dipilih.</span></div>}
    {!warehouseLoading && !warehouseError && selectedWarehouseID && !stockLoading && stockError && <div className="table-card state-card"><div className="state-icon error">!</div><strong>{stockError}</strong><button className="secondary-button" onClick={loadStocks}>Coba Lagi</button></div>}
    {!warehouseLoading && !warehouseError && selectedWarehouseID && !stockLoading && !stockError && stocks.length === 0 && <div className="table-card state-card"><div className="state-icon">▦</div><strong>Belum ada data stok untuk warehouse ini.</strong><span>Stok produk dari backend akan tampil di sini.</span></div>}
    {!warehouseLoading && !warehouseError && selectedWarehouseID && !stockLoading && !stockError && stocks.length > 0 && <div className="table-card table-scroll"><table className="stock-table"><thead><tr><th>Product Code</th><th>Product Name</th><th>Unit</th><th>Quantity</th><th>Minimum Stock</th><th>Status</th></tr></thead><tbody>{stocks.map((stock) => {
      const product = stock.product || {}
      const status = stockStatus(stock.quantity, product.minimum_stock ?? 0)
      return <tr key={stock.id}><td className="code-cell">{product.code || '—'}</td><td><strong className="table-name">{product.name || '—'}</strong></td><td>{product.unit || '—'}</td><td className="quantity-cell">{formatQuantity(stock.quantity)}</td><td>{formatQuantity(product.minimum_stock ?? 0)}</td><td><span className={`stock-status ${status.className}`}>{status.label}</span></td></tr>
    })}</tbody></table></div>}
  </div>
}
