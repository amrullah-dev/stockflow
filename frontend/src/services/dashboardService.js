import { api } from './api'

export async function getDashboardSummary() {
  const response = await api.get('/dashboard/summary')
  if (!response?.data || typeof response.data !== 'object') {
    throw new Error('Dashboard API returned an invalid response')
  }
  return {
    total_products: Number(response.data.total_products) || 0,
    total_warehouses: Number(response.data.total_warehouses) || 0,
    total_suppliers: Number(response.data.total_suppliers) || 0,
    total_stock_quantity: Number(response.data.total_stock_quantity) || 0,
    low_stock_count: Number(response.data.low_stock_count) || 0,
    recent_transactions: Array.isArray(response.data.recent_transactions) ? response.data.recent_transactions : [],
  }
}
