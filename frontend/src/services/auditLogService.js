import { api } from './api'

export async function getAuditLogs() {
  const response = await api.get('/audit-logs')
  return Array.isArray(response?.data) ? response.data : []
}

export async function getAuditLog(id) {
  const response = await api.get(`/audit-logs/${id}`)
  return response?.data || null
}
