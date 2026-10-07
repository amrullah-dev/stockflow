import { api } from './api'

const TOKEN_KEY = 'stockflow_token'
const USER_KEY = 'stockflow_user'
const LOGIN_PATH = import.meta.env.VITE_AUTH_LOGIN_PATH || '/login'

export async function login(credentials) {
  const payload = await api.post(LOGIN_PATH, credentials, { auth: false })
  const token = payload?.data?.token || payload?.token || payload?.data?.access_token || payload?.access_token
  if (!token) throw new Error('Login response did not contain a JWT token')
  localStorage.setItem(TOKEN_KEY, token)
  if (payload?.data?.user || payload?.user) localStorage.setItem(USER_KEY, JSON.stringify(payload.data?.user || payload.user))
  return payload
}

export function logout() {
  localStorage.removeItem(TOKEN_KEY)
  localStorage.removeItem(USER_KEY)
}

export function isAuthenticated() {
  return Boolean(localStorage.getItem(TOKEN_KEY))
}

export function currentUser() {
  try {
    return JSON.parse(localStorage.getItem(USER_KEY) || 'null')
  } catch {
    return null
  }
}
