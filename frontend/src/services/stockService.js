import { api } from './api'

export async function getWarehouseStocks(warehouseID) {
  const response = await api.get(`/warehouses/${warehouseID}/stocks`)
  return Array.isArray(response?.data) ? response.data : []
}
