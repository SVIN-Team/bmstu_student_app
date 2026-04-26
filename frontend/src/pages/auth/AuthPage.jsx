import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useAuth } from '../../providers/auth-context.js'
import { Button, Field, InlineMessage, TextInput } from '../../components/ui/ui.jsx'

const initialLogin = { email: '', password: '' }
const initialRegister = {
  email: '',
  password: '',
  first_name: '',
  last_name: '',
  group_name: '',
}

const featureCards = [
  {
    title: 'Расписание',
    description: 'Смотрите занятия на неделю, переключайте даты и быстро находите нужную пару.',
  },
  {
    title: 'Очереди',
    description: 'Записывайтесь в очередь, следите за статусом и открывайте детали в пару кликов.',
  },
  {
    title: 'Профиль',
    description: 'Управляйте личными данными, группой и рабочими сценариями внутри одного кабинета.',
  },
]

export function AuthPage() {
  const [tab, setTab] = useState('login')
  const [loginData, setLoginData] = useState(initialLogin)
  const [registerData, setRegisterData] = useState(initialRegister)
  const [error, setError] = useState('')
  const [busy, setBusy] = useState(false)
  const navigate = useNavigate()
  const { login, register } = useAuth()

  async function handleLogin(event) {
    event.preventDefault()
    setBusy(true)
    setError('')

    try {
      await login(loginData)
      navigate('/dashboard')
    } catch (requestError) {
      setError(requestError.message)
    } finally {
      setBusy(false)
    }
  }

  async function handleRegister(event) {
    event.preventDefault()
    setBusy(true)
    setError('')

    try {
      await register(registerData)
      navigate('/dashboard')
    } catch (requestError) {
      setError(requestError.message)
    } finally {
      setBusy(false)
    }
  }

  const cardTitle = tab === 'login' ? 'Вход' : 'Регистрация'

  return (
    <div className="auth-page">
      <section className="auth-hero">
        <span className="brand-chip">StudHub</span>
        <h1>Вход в учебный сервис</h1>
        <p>Авторизуйтесь или зарегистрируйтесь, чтобы перейти к расписанию, очередям и профилю.</p>

        <div className="auth-facts">
          {featureCards.map((card) => (
            <div key={card.title}>
              <strong>{card.title}</strong>
              <span>{card.description}</span>
            </div>
          ))}
        </div>
      </section>

      <section className="auth-card">
        <h2 className="auth-card-title">{cardTitle}</h2>

        <div className="tab-switch">
          <button className={tab === 'login' ? 'tab-active' : ''} onClick={() => setTab('login')}>
            Вход
          </button>
          <button
            className={tab === 'register' ? 'tab-active' : ''}
            onClick={() => setTab('register')}
          >
            Регистрация
          </button>
        </div>

        {error ? <InlineMessage tone="danger">{error}</InlineMessage> : null}

        {tab === 'login' ? (
          <form className="form-grid" onSubmit={handleLogin}>
            <Field label="Email">
              <TextInput
                type="email"
                value={loginData.email}
                onChange={(event) => setLoginData({ ...loginData, email: event.target.value })}
                placeholder="student@bmstu.ru"
                required
              />
            </Field>
            <Field label="Пароль">
              <TextInput
                type="password"
                value={loginData.password}
                onChange={(event) => setLoginData({ ...loginData, password: event.target.value })}
                placeholder="SecurePass123!"
                required
              />
            </Field>
            <Button type="submit" busy={busy}>
              Войти
            </Button>
          </form>
        ) : (
          <form className="form-grid" onSubmit={handleRegister}>
            <Field label="Имя">
              <TextInput
                value={registerData.first_name}
                onChange={(event) =>
                  setRegisterData({ ...registerData, first_name: event.target.value })
                }
                required
              />
            </Field>
            <Field label="Фамилия">
              <TextInput
                value={registerData.last_name}
                onChange={(event) =>
                  setRegisterData({ ...registerData, last_name: event.target.value })
                }
                required
              />
            </Field>
            <Field label="Email">
              <TextInput
                type="email"
                value={registerData.email}
                onChange={(event) =>
                  setRegisterData({ ...registerData, email: event.target.value })
                }
                required
              />
            </Field>
            <Field label="Пароль" hint="Минимум 8 символов и хотя бы одна цифра.">
              <TextInput
                type="password"
                value={registerData.password}
                onChange={(event) =>
                  setRegisterData({ ...registerData, password: event.target.value })
                }
                required
              />
            </Field>
            <Field label="Группа" hint="Например: ИУ7-81Б.">
              <TextInput
                value={registerData.group_name}
                onChange={(event) =>
                  setRegisterData({ ...registerData, group_name: event.target.value })
                }
                placeholder="ИУ7-81Б"
                required
              />
            </Field>
            <Button type="submit" busy={busy}>
              Создать аккаунт
            </Button>
          </form>
        )}
      </section>
    </div>
  )
}
