import { api } from './api'

const data = (response) => response?.data

export const getStockOuts = async () => data(await api.get('/stock-outs')) || []
export const getStockOut = async (id) => data(await api.get(`/stock-outs/${id}`))
export const createStockOut = async (payload) => data(await api.post('/stock-outs', payload))
export const updateStockOut = async (id, payload) => data(await api.put(`/stock-outs/${id}`, payload))
export const deleteStockOut = async (id) => api.delete(`/stock-outs/${id}`)
export const postStockOut = async (id) => api.post(`/stock-outs/${id}/post`, {})
