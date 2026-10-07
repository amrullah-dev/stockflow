import { api } from './api'

const data = (response) => response?.data
export const getSuppliers = async () => data(await api.get('/suppliers')) || []
export const getSupplier = async (id) => data(await api.get(`/suppliers/${id}`))
export const createSupplier = async (payload) => data(await api.post('/suppliers', payload))
export const updateSupplier = async (id, payload) => data(await api.put(`/suppliers/${id}`, payload))
export const deleteSupplier = async (id) => api.delete(`/suppliers/${id}`)
