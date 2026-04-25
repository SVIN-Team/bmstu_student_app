import { useState } from 'react'
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

export function ProfilePage() {
  const { api, user, refreshProfile } = useAuth()
  const [form, setForm] = useState({
    first_name: user?.first_name || '',
    last_name: user?.last_name || '',
    group_id: user?.group?.id || '',
  })
  const [headmanTargetId, setHeadmanTargetId] = useState('')
  const [message, setMessage] = useState('')
  const [error, setError] = useState('')

  async function handleSave(event) {
    event.preventDefault()
    setMessage('')
    setError('')

    try {
      await api.updateMe(form)
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
              label="ID группы"
              hint="Староста не сможет сменить группу, пока не передаст роль другому студенту."
            >
              <TextInput
                value={form.group_id}
                onChange={(event) => setForm({ ...form, group_id: event.target.value })}
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
