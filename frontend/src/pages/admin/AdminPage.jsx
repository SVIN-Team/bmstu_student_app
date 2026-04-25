import { useCallback, useEffect, useState } from 'react'
import { useAuth } from '../../providers/auth-context.js'
import { ROLE_LABELS } from '../../utils/format.js'
import {
  Button,
  EmptyState,
  Field,
  InlineMessage,
  LoaderBlock,
  PageHeader,
  Panel,
  SelectInput,
  TextInput,
} from '../../components/ui/ui.jsx'

function GroupManager({ api }) {
  const [groups, setGroups] = useState([])
  const [loading, setLoading] = useState(true)
  const [message, setMessage] = useState('')
  const [draft, setDraft] = useState({ name: '' })

  const loadGroups = useCallback(async () => {
    setLoading(true)

    try {
      const result = await api.getResource('groups')
      setGroups(result || [])
      setMessage('')
    } catch (error) {
      setMessage(error.message)
    } finally {
      setLoading(false)
    }
  }, [api])

  useEffect(() => {
    loadGroups()
  }, [loadGroups])

  async function createGroup(event) {
    event.preventDefault()

    try {
      await api.createResource('groups', draft)
      setDraft({ name: '' })
      await loadGroups()
    } catch (error) {
      setMessage(error.message)
    }
  }

  async function updateGroup(group, value) {
    try {
      await api.updateResource('groups', group.id, { name: value })
      await loadGroups()
    } catch (error) {
      setMessage(error.message)
    }
  }

  async function deleteGroup(id) {
    try {
      await api.deleteResource('groups', id)
      await loadGroups()
    } catch (error) {
      setMessage(error.message)
    }
  }

  return (
    <Panel title="Группы">
      {message ? <InlineMessage tone="info">{message}</InlineMessage> : null}
      <form className="form-row" onSubmit={createGroup}>
        <Field label="Название группы">
          <TextInput
            value={draft.name}
            onChange={(event) => setDraft({ name: event.target.value })}
            required
          />
        </Field>
        <Button type="submit">Создать</Button>
      </form>

      {loading ? <LoaderBlock label="Загружаем группы..." /> : null}

      {!loading && !groups.length ? (
        <EmptyState title="Групп пока нет" description="Создайте первую учебную группу." />
      ) : null}

      <div className="list-stack">
        {groups.map((group) => (
          <div className="admin-row" key={group.id}>
            <div className="admin-row-fields">
              <Field label="Название группы">
                <TextInput
                  value={group.name}
                  onChange={(event) => updateGroup(group, event.target.value)}
                />
              </Field>
            </div>
            <Button variant="danger" onClick={() => deleteGroup(group.id)}>
              Удалить
            </Button>
          </div>
        ))}
      </div>
    </Panel>
  )
}

export function AdminPage() {
  const { api } = useAuth()
  const [tab, setTab] = useState('users')
  const [users, setUsers] = useState([])
  const [loadingUsers, setLoadingUsers] = useState(true)
  const [userMessage, setUserMessage] = useState('')
  const [importMessage, setImportMessage] = useState('')

  const loadUsers = useCallback(async () => {
    setLoadingUsers(true)

    try {
      const result = await api.getAdminUsers({})
      setUsers(result || [])
      setUserMessage('')
    } catch (error) {
      setUserMessage(error.message)
    } finally {
      setLoadingUsers(false)
    }
  }, [api])

  useEffect(() => {
    if (tab === 'users') {
      loadUsers()
    }
  }, [loadUsers, tab])

  async function updateUser(userId, payload) {
    try {
      await api.updateAdminUser(userId, payload)
      await loadUsers()
    } catch (error) {
      setUserMessage(error.message)
    }
  }

  async function deleteUser(userId) {
    try {
      await api.deleteAdminUser(userId)
      await loadUsers()
    } catch (error) {
      setUserMessage(error.message)
    }
  }

  async function handleImport(event) {
    const file = event.target.files?.[0]

    if (!file) {
      return
    }

    try {
      const result = await api.importLessons(file)
      setImportMessage(
        `Импорт завершен. Добавлено: ${result.imported_count}. Ошибок: ${result.errors?.length || 0}.`,
      )
    } catch (error) {
      setImportMessage(error.message)
    }
  }

  const tabs = [
    { id: 'users', label: 'Пользователи' },
    { id: 'groups', label: 'Группы' },
    { id: 'imports', label: 'Импорт' },
  ]

  return (
    <div className="page-stack">
      <PageHeader
        eyebrow="Администрирование"
        title="Управление системой"
        description="Пользователи, учебные группы и импорт расписания."
      />

      <div className="tab-switch">
        {tabs.map((item) => (
          <button
            key={item.id}
            className={tab === item.id ? 'tab-active' : ''}
            onClick={() => setTab(item.id)}
          >
            {item.label}
          </button>
        ))}
      </div>

      {tab === 'users' ? (
        <Panel title="Пользователи">
          {userMessage ? <InlineMessage tone="info">{userMessage}</InlineMessage> : null}
          {loadingUsers ? <LoaderBlock label="Загружаем пользователей..." /> : null}
          <div className="list-stack">
            {users.map((user) => (
              <div className="admin-row" key={user.id}>
                <div>
                  <strong>
                    {user.first_name} {user.last_name}
                  </strong>
                  <p>{user.email}</p>
                  <p>
                    {ROLE_LABELS[user.role] || user.role} · {user.group?.name || 'Без группы'}
                  </p>
                </div>
                <div className="admin-row-actions">
                  <SelectInput
                    value={user.role}
                    onChange={(event) => updateUser(user.id, { role: event.target.value })}
                  >
                    {Object.entries(ROLE_LABELS).map(([value, label]) => (
                      <option key={value} value={value}>
                        {label}
                      </option>
                    ))}
                  </SelectInput>
                  <Button
                    variant={user.is_blocked ? 'secondary' : 'ghost'}
                    onClick={() => updateUser(user.id, { is_blocked: !user.is_blocked })}
                  >
                    {user.is_blocked ? 'Разблокировать' : 'Заблокировать'}
                  </Button>
                  <Button variant="danger" onClick={() => deleteUser(user.id)}>
                    Удалить
                  </Button>
                </div>
              </div>
            ))}
          </div>
        </Panel>
      ) : null}

      {tab === 'groups' ? <GroupManager api={api} /> : null}

      {tab === 'imports' ? (
        <Panel title="Импорт расписания" description="Загрузите JSON-файл расписания.">
          {importMessage ? <InlineMessage tone="info">{importMessage}</InlineMessage> : null}
          <Field label="Файл расписания">
            <TextInput type="file" accept=".json,application/json" onChange={handleImport} />
          </Field>
        </Panel>
      ) : null}
    </div>
  )
}
