import { api } from './api'

export async function getProducts() {
  const response = await api.get('/products')
  return Array.isArray(response?.data) ? response.data : []
}
