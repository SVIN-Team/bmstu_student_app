import { useEffect, useMemo, useState } from 'react'
import { useAuth } from '../../providers/auth-context.js'
import {
  LESSON_TYPE_LABELS,
  formatDate,
  formatDateTime,
  getWeekRange,
  groupLessonsByDate,
} from '../../utils/format.js'
import {
  Button,
  EmptyState,
  Field,
  InlineMessage,
  LoaderBlock,
  PageHeader,
  Panel,
  TextInput,
} from '../../components/ui/ui.jsx'

export function SchedulePage() {
  const { api, user } = useAuth()
  const [weekOffset, setWeekOffset] = useState(0)
  const [groupId, setGroupId] = useState('')
  const [state, setState] = useState({ loading: true, error: '', lessons: [] })
  const range = useMemo(() => getWeekRange(weekOffset), [weekOffset])
  const dateFrom = useMemo(() => range.from.toISOString().slice(0, 10), [range])
  const dateTo = useMemo(() => range.to.toISOString().slice(0, 10), [range])
  const requestedGroupId = user?.role === 'admin' ? groupId.trim() : user?.group?.id || ''

  useEffect(() => {
    let cancelled = false

    async function loadLessons() {
      if (!requestedGroupId) {
        setState({
          loading: false,
          error:
            user?.role === 'admin'
              ? 'Укажите UUID группы, чтобы загрузить расписание.'
              : 'Для пользователя не определена учебная группа.',
          lessons: [],
        })
        return
      }

      setState((current) => ({ ...current, loading: true, error: '' }))

      try {
        const lessons = await api.getLessons({
          group_id: requestedGroupId,
          date_from: dateFrom,
          date_to: dateTo,
        })

        if (!cancelled) {
          setState({ loading: false, error: '', lessons: lessons || [] })
        }
      } catch (error) {
        if (!cancelled) {
          setState({ loading: false, error: error.message, lessons: [] })
        }
      }
    }

    loadLessons()

    return () => {
      cancelled = true
    }
  }, [api, dateFrom, dateTo, requestedGroupId, user?.role])

  const groupedLessons = groupLessonsByDate(state.lessons)

  return (
    <div className="page-stack">
      <PageHeader
        eyebrow="Расписание"
        title="Недельный обзор занятий"
        description="Занятия группы на выбранную учебную неделю."
        actions={
          <div className="button-group">
            <Button variant="ghost" onClick={() => setWeekOffset((current) => current - 1)}>
              Прошлая неделя
            </Button>
            <Button variant="secondary" onClick={() => setWeekOffset(0)}>
              Текущая
            </Button>
            <Button variant="ghost" onClick={() => setWeekOffset((current) => current + 1)}>
              Следующая неделя
            </Button>
          </div>
        }
      />

      <Panel title="Фильтры" description={`${formatDate(range.from)} - ${formatDate(range.to)}`}>
        <div className="form-row">
          {user?.role === 'admin' ? (
            <Field label="UUID группы">
              <TextInput
                placeholder="UUID группы"
                value={groupId}
                onChange={(event) => setGroupId(event.target.value)}
              />
            </Field>
          ) : null}
        </div>
      </Panel>

      {state.loading ? <LoaderBlock label="Загружаем расписание..." /> : null}
      {state.error ? <InlineMessage tone="danger">{state.error}</InlineMessage> : null}

      {!state.loading && !state.error && !state.lessons.length ? (
        <EmptyState title="Расписание пустое" description="На выбранный диапазон занятий ничего не найдено." />
      ) : null}

      {!state.loading &&
        !state.error &&
        Object.entries(groupedLessons).map(([dateKey, lessons]) => (
          <Panel key={dateKey} title={formatDate(dateKey)}>
            <div className="list-stack">
              {lessons.map((lesson) => (
                <article className="lesson-card" key={lesson.id}>
                  <div>
                    <h3>{lesson.subject?.name}</h3>
                    <p>
                      {LESSON_TYPE_LABELS[lesson.type] || lesson.type} ·{' '}
                      {lesson.teacher?.full_name || 'Преподаватель не указан'}
                    </p>
                    <p>{lesson.room?.name || 'Аудитория не указана'}</p>
                  </div>
                  <div className="lesson-side">
                    <strong>{formatDateTime(lesson.starts_at)}</strong>
                    {lesson.queue_id ? <span className="badge badge-success">Есть очередь</span> : null}
                  </div>
                </article>
              ))}
            </div>
          </Panel>
        ))}
    </div>
  )
}
