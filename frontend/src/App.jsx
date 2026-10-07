import { Navigate, Route, Routes } from 'react-router-dom'
import AppLayout from './components/layout/AppLayout'
import ProtectedRoute from './components/layout/ProtectedRoute'
import Login from './pages/Login'
import Dashboard from './pages/Dashboard'
import Placeholder from './pages/Placeholder'
import Products from './pages/Products'
import Warehouses from './pages/Warehouses'
import Categories from './pages/Categories'
import Suppliers from './pages/Suppliers'
import StockIns from './pages/StockIns'
import StockInForm from './pages/StockInForm'
import StockInDetail from './pages/StockInDetail'
import Stocks from './pages/Stocks'
import AuditLogs from './pages/AuditLogs'
import StockMovements from './pages/StockMovements'

const placeholderRoutes = [
  ['products', 'Products'],
  ['categories', 'Categories'],
  ['suppliers', 'Suppliers'],
  ['warehouses', 'Warehouses'],
  ['stocks', 'Stock Overview'],
  ['stock-ins', 'Stock IN'],
  ['stock-outs', 'Stock OUT'],
  ['transfers', 'Transfer'],
  ['audit-logs', 'Audit Logs'],
]

export default function App() {
  return (
    <Routes>
      <Route path="/login" element={<Login />} />
      <Route element={<ProtectedRoute />}>
        <Route element={<AppLayout />}>
          <Route path="/dashboard" element={<Dashboard />} />
          <Route path="/products" element={<Products />} />
          <Route path="/warehouses" element={<Warehouses />} />
          <Route path="/categories" element={<Categories />} />
          <Route path="/suppliers" element={<Suppliers />} />
          <Route path="/stocks" element={<Stocks />} />
          <Route path="/stocks/movements" element={<StockMovements />} />
          <Route path="/audit-logs" element={<AuditLogs />} />
          <Route path="/stock-ins" element={<StockIns />} />
          <Route path="/stock-ins/new" element={<StockInForm />} />
          <Route path="/stock-ins/:id/edit" element={<StockInForm />} />
          <Route path="/stock-ins/:id" element={<StockInDetail />} />
          {placeholderRoutes.filter(([path]) => !['products', 'warehouses', 'stocks', 'stock-ins', 'audit-logs'].includes(path)).map(([path, title]) => (
            <Route key={path} path={`/${path}`} element={<Placeholder title={title} />} />
          ))}
        </Route>
      </Route>
      <Route path="/" element={<Navigate to="/dashboard" replace />} />
      <Route path="*" element={<Navigate to="/dashboard" replace />} />
    </Routes>
  )
}
