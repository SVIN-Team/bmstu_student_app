import { useEffect, useMemo, useState } from 'react'
import { useAuth } from '../../providers/auth-context.js'
import {
  QUEUE_STATUS_LABELS,
  canManageQueues,
  formatDateTime,
  fromDatetimeLocalValue,
  getWeekRange,
} from '../../utils/format.js'
import {
  Badge,
  Button,
  EmptyState,
  Field,
  InlineMessage,
  LoaderBlock,
  PageHeader,
  Panel,
  QueueCard,
  SelectInput,
  TextInput,
} from '../../components/ui/ui.jsx'

const defaultFilters = {
  group_id: '',
  status: '',
}

const defaultQueueForm = {
  lesson_id: '',
  opens_at: '',
  closes_at: '',
  max_size: '',
}

export function QueuesPage() {
  const { api, user } = useAuth()
  const [filters, setFilters] = useState(defaultFilters)
  const [queuesState, setQueuesState] = useState({ loading: true, error: '', queues: [] })
  const [formState, setFormState] = useState(defaultQueueForm)
  const [lessons, setLessons] = useState([])
  const [busy, setBusy] = useState(false)
  const [message, setMessage] = useState('')
  const range = useMemo(() => getWeekRange(0), [])

  useEffect(() => {
    let cancelled = false

    async function loadLessons() {
      if (user?.role === 'admin' && !user?.group?.id) {
        if (!cancelled) {
          setLessons([])
        }
        return
      }

      try {
        const result = await api.getLessons({
          date_from: range.from.toISOString().slice(0, 10),
          date_to: range.to.toISOString().slice(0, 10),
        })

        if (!cancelled) {
          setLessons(result || [])
        }
      } catch {
        if (!cancelled) {
          setLessons([])
        }
      }
    }

    loadLessons()

    return () => {
      cancelled = true
    }
  }, [api, range, user?.group?.id, user?.role])

  useEffect(() => {
    let cancelled = false

    async function loadQueues() {
      setQueuesState((current) => ({ ...current, loading: true, error: '' }))

      try {
        const result = await api.getQueues({
          ...filters,
          page: 1,
          per_page: 20,
        })

        if (!cancelled) {
          setQueuesState({
            loading: false,
            error: '',
            queues: result || [],
          })
        }
      } catch (error) {
        if (!cancelled) {
          setQueuesState({ loading: false, error: error.message, queues: [] })
        }
      }
    }

    loadQueues()

    return () => {
      cancelled = true
    }
  }, [api, filters])

  async function refreshQueues() {
    const refreshed = await api.getQueues({ page: 1, per_page: 20, ...filters })
    setQueuesState({ loading: false, error: '', queues: refreshed || [] })
  }

  async function handleCreateQueue(event) {
    event.preventDefault()
    setBusy(true)
    setMessage('')

    try {
      await api.createQueue({
        lesson_id: formState.lesson_id,
        opens_at: fromDatetimeLocalValue(formState.opens_at),
        closes_at: fromDatetimeLocalValue(formState.closes_at),
        max_size: formState.max_size ? Number(formState.max_size) : undefined,
      })

      setFormState(defaultQueueForm)
      setMessage('Очередь создана.')
      await refreshQueues()
    } catch (error) {
      setMessage(error.message)
    } finally {
      setBusy(false)
    }
  }

  const subjectOptions = []
  lessons.forEach((lesson) => {
    if (!subjectOptions.find((subject) => subject.id === lesson.subject?.id) && lesson.subject?.id) {
      subjectOptions.push(lesson.subject)
    }
  })

  return (
    <div className="page-stack">
      <PageHeader
        eyebrow="Очереди"
        title="Работа с очередями"
        description="Просмотр, фильтрация и создание очередей для вашей группы."
      />

      <div className="content-grid">
        <Panel title="Фильтры">
          <div className="form-row">
            <Field label="Статус">
              <SelectInput
                value={filters.status}
                onChange={(event) => setFilters({ ...filters, status: event.target.value })}
              >
                <option value="">Все</option>
                {Object.entries(QUEUE_STATUS_LABELS).map(([value, label]) => (
                  <option key={value} value={value}>
                    {label}
                  </option>
                ))}
              </SelectInput>
            </Field>
            <Field label="Group ID">
              <TextInput
                value={filters.group_id}
                onChange={(event) => setFilters({ ...filters, group_id: event.target.value })}
                placeholder={user?.group?.id || 'UUID группы'}
              />
            </Field>
          </div>
        </Panel>

        {canManageQueues(user?.role) ? (
          <Panel title="Создать очередь" description="Доступно старосте и администратору.">
            {message ? <InlineMessage tone="info">{message}</InlineMessage> : null}
            <form className="form-grid" onSubmit={handleCreateQueue}>
              <Field label="Предмет">
                <SelectInput
                  value={lessons.find((lesson) => lesson.id === formState.lesson_id)?.subject?.id || ''}
                  onChange={(event) =>
                    setFormState({
                      ...formState,
                      lesson_id:
                        lessons.find((lesson) => lesson.subject?.id === event.target.value)?.id || '',
                    })
                  }
                >
                  <option value="">Выберите предмет</option>
                  {subjectOptions.map((subject) => (
                    <option key={subject.id} value={subject.id}>
                      {subject.name}
                    </option>
                  ))}
                </SelectInput>
              </Field>
              <Field label="Занятие">
                <SelectInput
                  value={formState.lesson_id}
                  onChange={(event) => setFormState({ ...formState, lesson_id: event.target.value })}
                  required
                >
                  <option value="">Выберите занятие</option>
                  {lessons.map((lesson) => (
                    <option key={lesson.id} value={lesson.id}>
                      {lesson.subject?.name} · {formatDateTime(lesson.starts_at)}
                    </option>
                  ))}
                </SelectInput>
              </Field>
              <Field label="Открытие">
                <TextInput
                  type="datetime-local"
                  value={formState.opens_at}
                  onChange={(event) => setFormState({ ...formState, opens_at: event.target.value })}
                  required
                />
              </Field>
              <Field label="Закрытие">
                <TextInput
                  type="datetime-local"
                  value={formState.closes_at}
                  onChange={(event) => setFormState({ ...formState, closes_at: event.target.value })}
                />
              </Field>
              <Field label="Лимит слотов">
                <TextInput
                  type="number"
                  min="1"
                  value={formState.max_size}
                  onChange={(event) => setFormState({ ...formState, max_size: event.target.value })}
                />
              </Field>
              <Button type="submit" busy={busy}>
                Создать очередь
              </Button>
            </form>
          </Panel>
        ) : null}
      </div>

      {queuesState.loading ? <LoaderBlock label="Загружаем очереди..." /> : null}
      {queuesState.error ? <InlineMessage tone="danger">{queuesState.error}</InlineMessage> : null}

      {!queuesState.loading && !queuesState.queues.length ? (
        <EmptyState
          title="Очередей не найдено"
          description="Попробуйте ослабить фильтры или создать новую очередь."
        />
      ) : null}

      <div className="queue-grid">
        {queuesState.queues.map((queue) => (
          <QueueCard
            key={queue.id}
            queue={queue}
            extra={<Badge tone="lilac">{QUEUE_STATUS_LABELS[queue.status] || queue.status}</Badge>}
          >
            <span>Открытие: {formatDateTime(queue.opens_at)}</span>
            <span>Закрытие: {formatDateTime(queue.closes_at)}</span>
            <span>
              Слоты: {queue.slots_count}
              {queue.max_size ? ` / ${queue.max_size}` : ''}
            </span>
          </QueueCard>
        ))}
      </div>
    </div>
  )
}
