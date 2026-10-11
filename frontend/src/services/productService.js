import { api } from './api'
export async function getProducts() { const response = await api.get('/products'); return Array.isArray(response?.data) ? response.data : [] }
const data = (response) => response?.data
export const createProduct = async (payload) => data(await api.post('/products', payload))
export const updateProduct = async (id, payload) => data(await api.put(`/products/${id}`, payload))
export const deleteProduct = async (id) => api.delete(`/products/${id}`)
