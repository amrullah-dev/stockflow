import { useCallback, useEffect, useState } from 'react'
import { getAuditLog, getAuditLogs } from '../services/auditLogService'

function errorMessage(error, fallback) {
  return error?.payload?.message || error?.message || fallback
}

function formatDate(value) {
  if (!value) return '-'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return '-'
  return new Intl.DateTimeFormat('id-ID', { dateStyle: 'short', timeStyle: 'short' }).format(date)
}

function displayValue(value) {
  if (value === null || value === undefined || value === '') return '-'
  if (typeof value === 'string') return value
  return JSON.stringify(value)
}

function UserSummary({ user }) {
  return user?.name || '-'
}

export default function AuditLogs() {
  const [logs, setLogs] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [detail, setDetail] = useState(null)
  const [detailLoading, setDetailLoading] = useState(false)
  const [detailError, setDetailError] = useState('')

  const loadLogs = useCallback(async () => {
    setLoading(true)
    setError('')
    try {
      setLogs(await getAuditLogs())
    } catch (err) {
      setLogs([])
      setError(errorMessage(err, 'Gagal memuat audit log.'))
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => { loadLogs() }, [loadLogs])

  async function showDetail(id) {
    setDetail(null)
    setDetailError('')
    setDetailLoading(true)
    try {
      setDetail(await getAuditLog(id))
    } catch (err) {
      setDetailError(errorMessage(err, 'Gagal memuat detail audit log.'))
    } finally {
      setDetailLoading(false)
    }
  }

  return <div className="data-page">
    <div className="page-heading data-page-heading">
      <div><div className="eyebrow">System</div><h2>Audit Logs</h2><p className="muted">Riwayat aktivitas penting pada StockFlow.</p></div>
      <button className="secondary-button audit-refresh-button" disabled={loading} onClick={loadLogs}>{loading ? 'Memuat...' : 'Refresh'}</button>
    </div>

    {loading && <div className="table-card state-card"><div className="spinner" /><strong>Memuat audit logs...</strong><span>Mengambil riwayat aktivitas dari backend.</span></div>}
    {!loading && error && <div className="table-card state-card"><div className="state-icon error">!</div><strong>{error}</strong><button className="secondary-button" onClick={loadLogs}>Coba Lagi</button></div>}
    {!loading && !error && logs.length === 0 && <div className="table-card state-card"><div className="state-icon">◷</div><strong>Belum ada audit log.</strong><span>Aktivitas sistem dari backend akan tampil di sini.</span></div>}
    {!loading && !error && logs.length > 0 && <div className="table-card table-scroll"><table className="audit-table"><thead><tr><th>Waktu</th><th>Action</th><th>Subject Type</th><th>Subject ID</th><th>User</th><th>Old Values</th><th>New Values</th><th>Aksi</th></tr></thead><tbody>{logs.map((log) => <tr key={log.id}><td>{formatDate(log.created_at)}</td><td className="code-cell">{log.action || '-'}</td><td>{log.subject_type || '-'}</td><td>{log.subject_id ?? '-'}</td><td><UserSummary user={log.user} /></td><td className="audit-value-cell" title={displayValue(log.old_values)}>{displayValue(log.old_values)}</td><td className="audit-value-cell" title={displayValue(log.new_values)}>{displayValue(log.new_values)}</td><td><button className="table-action-button" onClick={() => showDetail(log.id)}>Detail</button></td></tr>)}</tbody></table></div>}

    {(detailLoading || detailError || detail) && <div className="audit-detail-backdrop" role="presentation" onClick={() => { if (!detailLoading) { setDetail(null); setDetailError('') } }}>
      <section className="audit-detail-panel" role="dialog" aria-modal="true" aria-label="Detail audit log" onClick={(event) => event.stopPropagation()}>
        <div className="audit-detail-heading"><div><div className="eyebrow">Audit Detail</div><h3>Detail Audit Log</h3></div><button className="audit-close-button" onClick={() => { setDetail(null); setDetailError('') }} aria-label="Tutup">×</button></div>
        {detailLoading && <div className="mini-state"><div className="spinner" /><span>Memuat detail audit log...</span></div>}
        {!detailLoading && detailError && <div className="form-error">{detailError}</div>}
        {!detailLoading && !detailError && detail && <div className="audit-detail-content"><div className="audit-detail-grid"><div><span>ID</span><strong>{detail.id}</strong></div><div><span>Waktu</span><strong>{formatDate(detail.created_at)}</strong></div><div><span>Action</span><strong>{detail.action || '-'}</strong></div><div><span>Subject</span><strong>{detail.subject_type || '-'} #{detail.subject_id ?? '-'}</strong></div><div><span>User</span><strong><UserSummary user={detail.user} /></strong></div><div><span>User ID</span><strong>{detail.user_id ?? '-'}</strong></div><div><span>IP Address</span><strong>{detail.ip_address || '-'}</strong></div><div><span>User Agent</span><strong>{detail.user_agent || '-'}</strong></div></div><div className="audit-detail-values"><div><span>Old Values</span><pre>{displayValue(detail.old_values)}</pre></div><div><span>New Values</span><pre>{displayValue(detail.new_values)}</pre></div></div></div>}
      </section>
    </div>}
  </div>
}
