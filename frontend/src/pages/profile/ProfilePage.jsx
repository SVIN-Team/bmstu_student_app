import { useEffect, useMemo, useState } from 'react'
import { useAuth } from '../../providers/auth-context.js'
import { ROLE_LABELS, fullName, formatDateTime } from '../../utils/format.js'
import {
  Button,
  Field,
  InlineMessage,
  KeyValueList,
  PageHeader,
  Panel,
  TextInput,
} from '../../components/ui/ui.jsx'

function normalizeName(value) {
  return value?.trim().toLowerCase() || ''
}

function resolveGroup(groups, query) {
  const normalizedQuery = normalizeName(query)
  if (!normalizedQuery) {
    return null
  }

  const exactMatch = groups.find((group) => normalizeName(group.name) === normalizedQuery)
  if (exactMatch) {
    return exactMatch
  }

  const containsMatches = groups.filter((group) => normalizeName(group.name).includes(normalizedQuery))
  if (containsMatches.length === 1) {
    return containsMatches[0]
  }

  return null
}

export function ProfilePage() {
  const { api, user, refreshProfile } = useAuth()
  const [groups, setGroups] = useState([])
  const [form, setForm] = useState({
    first_name: user?.first_name || '',
    last_name: user?.last_name || '',
    group_name: user?.group?.name || '',
  })
  const [headmanTargetId, setHeadmanTargetId] = useState('')
  const [message, setMessage] = useState('')
  const [error, setError] = useState('')
  const resolvedGroup = useMemo(() => resolveGroup(groups, form.group_name), [form.group_name, groups])
  const currentGroupName = user?.group?.name || ''
  const currentGroupId = user?.group?.id || ''

  useEffect(() => {
    let cancelled = false

    async function loadGroups() {
      try {
        const result = await api.getResource('groups')
        if (!cancelled) {
          setGroups(Array.isArray(result) ? result : [])
        }
      } catch {
        if (!cancelled) {
          setGroups([])
        }
      }
    }

    loadGroups()

    return () => {
      cancelled = true
    }
  }, [api])

  async function handleSave(event) {
    event.preventDefault()
    setMessage('')
    setError('')

    const trimmedGroupName = form.group_name.trim()
    const groupChanged = normalizeName(trimmedGroupName) !== normalizeName(currentGroupName)

    if (groupChanged && trimmedGroupName && !resolvedGroup) {
      setError('Группа не найдена. Уточните название.')
      return
    }

    try {
      await api.updateMe({
        first_name: form.first_name,
        last_name: form.last_name,
        group_id: groupChanged ? resolvedGroup?.id || null : currentGroupId || null,
      })
      await refreshProfile()
      setMessage('Профиль обновлен.')
    } catch (requestError) {
      setError(requestError.message)
    }
  }

  async function handleTransferRole() {
    setMessage('')
    setError('')

    try {
      await api.transferHeadmanRole({ to_user_id: headmanTargetId })
      await refreshProfile()
      setMessage('Роль старосты передана.')
      setHeadmanTargetId('')
    } catch (requestError) {
      setError(requestError.message)
    }
  }

  return (
    <div className="page-stack">
      <PageHeader
        eyebrow="Профиль"
        title={fullName(user)}
        description="Личные данные, группа и передача роли старосты."
      />

      {message ? <InlineMessage tone="success">{message}</InlineMessage> : null}
      {error ? <InlineMessage tone="danger">{error}</InlineMessage> : null}

      <div className="content-grid">
        <Panel title="Текущие данные">
          <KeyValueList
            items={[
              { label: 'Email', value: user?.email },
              { label: 'Роль', value: ROLE_LABELS[user?.role] || user?.role },
              { label: 'Группа', value: user?.group?.name || user?.group?.id || '—' },
              { label: 'Создан', value: formatDateTime(user?.created_at) },
            ]}
          />
        </Panel>

        <Panel title="Редактирование профиля">
          <form className="form-grid" onSubmit={handleSave}>
            <Field label="Имя">
              <TextInput
                value={form.first_name}
                onChange={(event) => setForm({ ...form, first_name: event.target.value })}
                required
              />
            </Field>
            <Field label="Фамилия">
              <TextInput
                value={form.last_name}
                onChange={(event) => setForm({ ...form, last_name: event.target.value })}
                required
              />
            </Field>
            <Field
              label="Группа"
              hint="Указывайте название группы. Например: ИУ7-81Б. Староста не сможет сменить группу, пока не передаст роль другому студенту."
            >
              <TextInput
                value={form.group_name}
                onChange={(event) => setForm({ ...form, group_name: event.target.value })}
                placeholder="ИУ7-81Б"
              />
            </Field>
            <Button type="submit">Сохранить</Button>
          </form>
        </Panel>
      </div>

      {user?.role === 'headman' ? (
        <Panel title="Передача роли старосты" description="Используется реальная ручка смены старосты.">
          <div className="form-row">
            <Field label="ID нового старосты">
              <TextInput
                value={headmanTargetId}
                onChange={(event) => setHeadmanTargetId(event.target.value)}
                placeholder="UUID пользователя"
              />
            </Field>
            <Button onClick={handleTransferRole} disabled={!headmanTargetId}>
              Передать роль
            </Button>
          </div>
        </Panel>
      ) : null}
    </div>
  )
}
