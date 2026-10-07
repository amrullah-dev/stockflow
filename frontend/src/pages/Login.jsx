import { useState } from 'react'
import { useLocation, useNavigate } from 'react-router-dom'
import { login } from '../services/auth'

export default function Login() {
  const navigate = useNavigate()
  const location = useLocation()
  const [form, setForm] = useState({ email: '', password: '' })
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)
  async function submit(event) {
    event.preventDefault(); setError(''); setLoading(true)
    try { await login(form); navigate(location.state?.from?.pathname || '/dashboard', { replace: true }) } catch (err) { setError(err.message || 'Unable to sign in') } finally { setLoading(false) }
  }
  return <main className="login-page"><div className="login-panel"><div className="login-brand"><div className="brand-mark">S</div><span>StockFlow</span></div><p className="login-kicker">Inventory Management System</p><h1>Welcome back</h1><p className="muted">Sign in to continue to your workspace.</p><form onSubmit={submit}><label>Email<input type="email" value={form.email} onChange={(e) => setForm({ ...form, email: e.target.value })} required placeholder="you@example.com" /></label><label>Password<input type="password" value={form.password} onChange={(e) => setForm({ ...form, password: e.target.value })} required placeholder="••••••••" /></label>{error && <div className="form-error">{error}</div>}<button className="primary-button" disabled={loading}>{loading ? 'Signing in…' : 'Login'}</button></form><p className="login-note">Use your StockFlow account credentials.</p></div></main>
}
