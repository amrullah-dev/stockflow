import { api } from './api'
export async function getWarehouses() { const response = await api.get('/warehouses'); return Array.isArray(response?.data) ? response.data : [] }
const data = (response) => response?.data
export const createWarehouse = async (payload) => data(await api.post('/warehouses', payload))
export const updateWarehouse = async (id, payload) => data(await api.put(`/warehouses/${id}`, payload))
export const deleteWarehouse = async (id) => api.delete(`/warehouses/${id}`)
