import { NavLink, Outlet, useNavigate } from 'react-router-dom'
import { useAuth } from '../../providers/auth-context.js'
import { ROLE_LABELS, canManageQueues } from '../../utils/format.js'
import { Button } from '../ui/ui.jsx'

export function AppShell() {
  const { user, logout } = useAuth()
  const navigate = useNavigate()

  const items = [
    { to: '/dashboard', label: 'Обзор' },
    { to: '/schedule', label: 'Расписание' },
    { to: '/queues', label: 'Очереди' },
    { to: '/my-slots', label: 'Мои записи' },
    { to: '/profile', label: 'Профиль' },
  ]

  if (user?.role === 'admin') {
    items.push({ to: '/admin', label: 'Админка' })
  }

  return (
    <div className="app-shell">
      <aside className="sidebar">
        <div className="brand-card">
          <span className="brand-chip">StudHub</span>
          <h2>Расписание, очереди и учебные процессы</h2>
          <p>Один рабочий интерфейс для студента, старосты и администратора.</p>
        </div>

        <nav className="nav-list">
          {items.map((item) => (
            <NavLink
              key={item.to}
              to={item.to}
              className={({ isActive }) => (isActive ? 'nav-item nav-item-active' : 'nav-item')}
            >
              {item.label}
            </NavLink>
          ))}
        </nav>

        <div className="sidebar-foot">
          <div className="mini-user-card">
            <strong className="mini-user-name">
              {[user?.first_name, user?.last_name].filter(Boolean).join(' ') || 'Пользователь'}
            </strong>
            <div className="mini-user-meta">
              <span>{ROLE_LABELS[user?.role] || 'Пользователь'}</span>
              <small>{user?.group?.name || 'Без группы'}</small>
            </div>
          </div>
          <Button
            variant="ghost"
            onClick={async () => {
              await logout()
              navigate('/auth')
            }}
          >
            Выйти
          </Button>
        </div>
      </aside>

      <main className="content">
        <header className="topbar">
          <div>
            <p className="topbar-kicker">Рабочее пространство</p>
            <h1>
              {canManageQueues(user?.role)
                ? 'Панель управления учебным потоком'
                : 'Личный кабинет студента'}
            </h1>
          </div>
        </header>

        <Outlet />
      </main>
    </div>
  )
}
