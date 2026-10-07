import { useCallback, useEffect, useState } from 'react'
import { getProducts } from '../services/productService'

function StatusBadge({ active }) {
  return <span className={active ? 'status-badge active' : 'status-badge inactive'}>{active ? 'Active' : 'Inactive'}</span>
}

export default function Products() {
  const [products, setProducts] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(false)

  const loadProducts = useCallback(async () => {
    setLoading(true)
    setError(false)
    try { setProducts(await getProducts()) } catch { setError(true) } finally { setLoading(false) }
  }, [])

  useEffect(() => { loadProducts() }, [loadProducts])

  return <div className="data-page">
    <div className="page-heading data-page-heading"><div><div className="eyebrow">Master Data</div><h2>Products</h2><p className="muted">Manage products and inventory item information</p></div></div>
    {loading && <div className="table-card state-card"><div className="spinner" /><strong>Loading products…</strong><span>Fetching the latest product information.</span></div>}
    {!loading && error && <div className="table-card state-card"><div className="state-icon error">!</div><strong>Failed to load products.</strong><span>There was a problem connecting to the API.</span><button className="secondary-button" onClick={loadProducts}>Retry</button></div>}
    {!loading && !error && products.length === 0 && <div className="table-card state-card"><div className="state-icon">◈</div><strong>No products found.</strong><span>Products from the backend will appear here.</span></div>}
    {!loading && !error && products.length > 0 && <div className="table-card table-scroll"><table><thead><tr><th>Code</th><th>Product Name</th><th>Category</th><th>Unit</th><th>Minimum Stock</th><th>Status</th></tr></thead><tbody>{products.map((product) => <tr key={product.id}><td className="code-cell">{product.code || '—'}</td><td><strong className="table-name">{product.name || '—'}</strong><span className="table-description">{product.description || 'No description'}</span></td><td>{product.category?.name || '—'}</td><td>{product.unit || '—'}</td><td>{product.minimum_stock ?? '—'}</td><td><StatusBadge active={product.is_active} /></td></tr>)}</tbody></table></div>}
  </div>
}
