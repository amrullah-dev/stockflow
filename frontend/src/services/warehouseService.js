import { api } from './api'

export async function getWarehouses() {
  const response = await api.get('/warehouses')
  return Array.isArray(response?.data) ? response.data : []
}
