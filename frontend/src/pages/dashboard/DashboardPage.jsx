import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { useAuth } from '../../providers/auth-context.js'
import {
  LESSON_TYPE_LABELS,
  QUEUE_STATUS_LABELS,
  formatDateTime,
  formatQueueSlots,
  getWeekRange,
} from '../../utils/format.js'
import { EmptyState, LoaderBlock, PageHeader, Panel, QueueCard, StatCard } from '../../components/ui/ui.jsx'

function nearestOpenQueues(queues) {
  const now = Date.now()

  return [...queues]
    .filter((queue) => queue.status === 'open')
    .sort((left, right) => {
      const leftTime = new Date(left.opens_at || left.closes_at || 0).getTime()
      const rightTime = new Date(right.opens_at || right.closes_at || 0).getTime()

      return Math.abs(leftTime - now) - Math.abs(rightTime - now)
    })
    .slice(0, 4)
}

export function DashboardPage() {
  const { api, user } = useAuth()
  const [state, setState] = useState({
    loading: true,
    error: '',
    lessons: [],
    queues: [],
    slots: [],
  })

  useEffect(() => {
    const range = getWeekRange(0)

    async function load() {
      try {
        const lessonsPromise =
          user?.role === 'admin' && !user?.group?.id
            ? Promise.resolve([])
            : api.getLessons({
                date_from: range.from.toISOString().slice(0, 10),
                date_to: range.to.toISOString().slice(0, 10),
              })

        const [lessons, queues, slots] = await Promise.all([
          lessonsPromise,
          api.getQueues({ status: 'open', per_page: 20 }),
          api.getMySlots({}),
        ])

        setState({
          loading: false,
          error: '',
          lessons: lessons || [],
          queues: (queues || []).filter((queue) => queue.status === 'open'),
          slots: slots || [],
        })
      } catch (error) {
        setState((current) => ({ ...current, loading: false, error: error.message }))
      }
    }

    load()
  }, [api, user?.group?.id, user?.role])

  if (state.loading) {
    return <LoaderBlock label="Собираем обзор по расписанию, очередям и вашим записям..." />
  }

  const dashboardQueues = nearestOpenQueues(state.queues)

  return (
    <div className="page-stack">
      <PageHeader
        eyebrow="Обзор"
        title={`Привет, ${user?.first_name || 'пользователь'}`}
        description="Здесь собраны ближайшие занятия, активные очереди и ваши текущие записи."
        actions={
          <Link className="button button-secondary" to="/queues">
            Перейти к очередям
          </Link>
        }
      />

      {state.error ? <div className="inline-message inline-message-danger">{state.error}</div> : null}

      <div className="stats-grid">
        <StatCard label="Занятий на неделе" value={state.lessons.length} />
        <StatCard label="Доступных очередей" value={state.queues.length} />
        <StatCard
          label="Моих активных записей"
          value={state.slots.filter((slot) => slot.status === 'waiting').length}
        />
      </div>

      <div className="content-grid">
        <Panel title="Ближайшие занятия" description="Показываем занятия на текущую учебную неделю.">
          {state.lessons.length ? (
            <div className="list-stack">
              {state.lessons.slice(0, 5).map((lesson) => (
                <div className="list-row" key={lesson.id}>
                  <div>
                    <strong>{lesson.subject?.name}</strong>
                    <p>
                      {LESSON_TYPE_LABELS[lesson.type] || lesson.type} · {lesson.teacher?.full_name} ·{' '}
                      {lesson.room?.name || 'Без аудитории'}
                    </p>
                  </div>
                  <span>{formatDateTime(lesson.starts_at)}</span>
                </div>
              ))}
            </div>
          ) : (
            <EmptyState title="Занятий нет" description="На выбранную неделю расписание пока пустое." />
          )}
        </Panel>

        <Panel title="Мои записи" description="Здесь видны только ваши слоты в очередях.">
          {state.slots.length ? (
            <div className="list-stack">
              {state.slots.slice(0, 5).map((slot) => (
                <div className="list-row" key={slot.id}>
                  <div>
                    <strong>{slot.queue?.subject?.name}</strong>
                    <p>
                      Статус слота: {slot.status} · Статус очереди:{' '}
                      {QUEUE_STATUS_LABELS[slot.queue?.status] || slot.queue?.status}
                    </p>
                  </div>
                  <span>{formatDateTime(slot.signed_up_at)}</span>
                </div>
              ))}
            </div>
          ) : (
            <EmptyState
              title="Записей пока нет"
              description="После записи в очередь здесь появится краткая сводка."
              action={
                <Link className="button button-primary" to="/queues">
                  Найти очередь
                </Link>
              }
            />
          )}
        </Panel>
      </div>

      <Panel title="Актуальные очереди" description="Ближайшие открытые очереди вашей группы.">
        {dashboardQueues.length ? (
          <div className="queue-grid">
            {dashboardQueues.map((queue) => (
              <QueueCard
                key={queue.id}
                queue={queue}
                extra={<span className="badge badge-lilac">{QUEUE_STATUS_LABELS[queue.status] || queue.status}</span>}
              >
                <span>Открытие: {formatDateTime(queue.opens_at)}</span>
                <span>Закрытие: {formatDateTime(queue.closes_at)}</span>
                <span>{formatQueueSlots(queue)}</span>
              </QueueCard>
            ))}
          </div>
        ) : (
          <EmptyState title="Открытых очередей нет" description="Сейчас нет открытых очередей для записи." />
        )}
      </Panel>
    </div>
  )
}
