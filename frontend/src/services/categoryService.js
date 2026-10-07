import { api } from './api'

const data = (response) => response?.data
export const getCategories = async () => data(await api.get('/categories')) || []
export const getCategory = async (id) => data(await api.get(`/categories/${id}`))
export const createCategory = async (payload) => data(await api.post('/categories', payload))
export const updateCategory = async (id, payload) => data(await api.put(`/categories/${id}`, payload))
export const deleteCategory = async (id) => api.delete(`/categories/${id}`)
