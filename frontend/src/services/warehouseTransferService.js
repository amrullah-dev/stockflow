import { api } from './api'

const data = (response) => response?.data

export const getTransfers = async () => data(await api.get('/warehouse-transfers')) || []
export const getTransfer = async (id) => data(await api.get(`/warehouse-transfers/${id}`))
export const createTransfer = async (payload) => data(await api.post('/warehouse-transfers', payload))
export const updateTransfer = async (id, payload) => data(await api.put(`/warehouse-transfers/${id}`, payload))
export const deleteTransfer = async (id) => api.delete(`/warehouse-transfers/${id}`)
export const postTransfer = async (id) => api.post(`/warehouse-transfers/${id}/post`, {})
