import { api } from './api'

export async function getStockMovements() {
  const response = await api.get('/stock-movements')
  return Array.isArray(response?.data) ? response.data : []
}
