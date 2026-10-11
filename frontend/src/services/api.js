const API_BASE_URL = (import.meta.env.VITE_API_BASE_URL || 'http://localhost:3000/api').replace(/\/$/, '')

export class ApiError extends Error {
  constructor(message, status, payload) {
    super(message)
    this.name = 'ApiError'
    this.status = status
    this.payload = payload
  }
}

async function request(path, options = {}) {
  const headers = new Headers(options.headers || {})
  if (options.body && !(options.body instanceof FormData)) headers.set('Content-Type', 'application/json')

  const token = localStorage.getItem('stockflow_token')
  if (options.auth !== false && token) headers.set('Authorization', `Bearer ${token}`)

  const url = `${API_BASE_URL}${path}`
  const response = await fetch(url, { ...options, headers })
  const responseText = await response.text()
  const bodyPreview = responseText.trim().slice(0, 500)

  let payload = null
  if (responseText.trim()) {
    try { payload = JSON.parse(responseText) } catch { payload = null }
  }

  if (response.status === 401) {
    localStorage.removeItem('stockflow_token')
    localStorage.removeItem('stockflow_user')
    if (window.location.pathname !== '/login') window.location.assign('/login')
  }
  if (!response.ok) {
    const detail = payload?.message || bodyPreview
    throw new ApiError(detail || `Request failed (${response.status})`, response.status, payload)
  }
  if (!payload) {
    throw new ApiError(`API returned non-JSON response (${response.status})${bodyPreview ? `: ${bodyPreview}` : ''}`, response.status, null)
  }
  return payload
}

export const api = {
  get: (path, options) => request(path, { ...options, method: 'GET' }),
  post: (path, body, options) => request(path, { ...options, method: 'POST', body: JSON.stringify(body) }),
  put: (path, body, options) => request(path, { ...options, method: 'PUT', body: JSON.stringify(body) }),
  delete: (path, options) => request(path, { ...options, method: 'DELETE' }),
}

export { API_BASE_URL }
