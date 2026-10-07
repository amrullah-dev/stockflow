import { api } from './api'

const data = (response) => response?.data

export const getStockIns = async () => data(await api.get('/stock-ins')) || []
export const getStockIn = async (id) => data(await api.get(`/stock-ins/${id}`))
export const createStockIn = async (payload) => data(await api.post('/stock-ins', payload))
export const updateStockIn = async (id, payload) => data(await api.put(`/stock-ins/${id}`, payload))
export const deleteStockIn = async (id) => api.delete(`/stock-ins/${id}`)
export const postStockIn = async (id) => api.post(`/stock-ins/${id}/post`, {})
