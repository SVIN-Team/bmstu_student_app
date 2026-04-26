import { Navigate, Route, Routes } from 'react-router-dom'
import { AuthProvider } from './providers/AuthProvider.jsx'
import { useAuth } from './providers/auth-context.js'
import { AppShell } from './components/layout/AppShell.jsx'
import { AuthPage } from './pages/auth/AuthPage.jsx'
import { DashboardPage } from './pages/dashboard/DashboardPage.jsx'
import { SchedulePage } from './pages/schedule/SchedulePage.jsx'
import { QueuesPage } from './pages/queues/QueuesPage.jsx'
import { QueueDetailsPage } from './pages/queues/QueueDetailsPage.jsx'
import { ProfilePage } from './pages/profile/ProfilePage.jsx'
import { MySlotsPage } from './pages/profile/MySlotsPage.jsx'
import { AdminPage } from './pages/admin/AdminPage.jsx'
import { LoaderBlock } from './components/ui/ui.jsx'

function ProtectedRoute({ children, roles }) {
  const { isReady, isAuthenticated, user } = useAuth()

  if (!isReady) {
    return <LoaderBlock label="Поднимаем приложение и проверяем сессию..." />
  }

  if (!isAuthenticated) {
    return <Navigate to="/auth" replace />
  }

  if (roles?.length && !roles.includes(user?.role)) {
    return <Navigate to="/dashboard" replace />
  }

  return children
}

function PublicRoute({ children }) {
  const { isReady, isAuthenticated } = useAuth()

  if (!isReady) {
    return <LoaderBlock label="Готовим интерфейс..." />
  }

  if (isAuthenticated) {
    return <Navigate to="/dashboard" replace />
  }

  return children
}

function AppRoutes() {
  return (
    <Routes>
      <Route
        path="/auth"
        element={
          <PublicRoute>
            <AuthPage />
          </PublicRoute>
        }
      />
      <Route
        path="/"
        element={
          <ProtectedRoute>
            <AppShell />
          </ProtectedRoute>
        }
      >
        <Route index element={<Navigate to="/dashboard" replace />} />
        <Route path="dashboard" element={<DashboardPage />} />
        <Route path="schedule" element={<SchedulePage />} />
        <Route path="queues" element={<QueuesPage />} />
        <Route path="queues/:queueId" element={<QueueDetailsPage />} />
        <Route path="my-slots" element={<MySlotsPage />} />
        <Route path="profile" element={<ProfilePage />} />
        <Route
          path="admin"
          element={
            <ProtectedRoute roles={['admin']}>
              <AdminPage />
            </ProtectedRoute>
          }
        />
      </Route>
      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  )
}

export default function App() {
  return (
    <AuthProvider>
      <AppRoutes />
    </AuthProvider>
  )
}
