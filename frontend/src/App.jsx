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
import StockOuts from './pages/StockOuts'
import StockOutForm from './pages/StockOutForm'
import StockOutDetail from './pages/StockOutDetail'
import Transfers from './pages/Transfers'
import TransferForm from './pages/TransferForm'
import TransferDetail from './pages/TransferDetail'

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
          <Route path="/transfers" element={<Transfers />} />
          <Route path="/transfers/new" element={<TransferForm />} />
          <Route path="/transfers/:id/edit" element={<TransferForm />} />
          <Route path="/transfers/:id" element={<TransferDetail />} />
          <Route path="/stock-outs" element={<StockOuts />} />
          <Route path="/stock-outs/new" element={<StockOutForm />} />
          <Route path="/stock-outs/:id/edit" element={<StockOutForm />} />
          <Route path="/stock-outs/:id" element={<StockOutDetail />} />
          <Route path="/stock-ins/new" element={<StockInForm />} />
          <Route path="/stock-ins/:id/edit" element={<StockInForm />} />
          <Route path="/stock-ins/:id" element={<StockInDetail />} />
          {placeholderRoutes.filter(([path]) => !['products', 'warehouses', 'stocks', 'stock-ins', 'stock-outs', 'transfers', 'audit-logs'].includes(path)).map(([path, title]) => (
            <Route key={path} path={`/${path}`} element={<Placeholder title={title} />} />
          ))}
        </Route>
      </Route>
      <Route path="/" element={<Navigate to="/dashboard" replace />} />
      <Route path="*" element={<Navigate to="/dashboard" replace />} />
    </Routes>
  )
}
