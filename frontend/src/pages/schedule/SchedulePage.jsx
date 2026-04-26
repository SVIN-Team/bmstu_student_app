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

function normalizeName(value) {
  return value?.trim().toLowerCase() || ''
}

function resolveGroupId(groups, query) {
  const normalizedQuery = normalizeName(query)
  if (!normalizedQuery) {
    return ''
  }

  const exactMatch = groups.find((group) => normalizeName(group.name) === normalizedQuery)
  if (exactMatch) {
    return exactMatch.id
  }

  const containsMatches = groups.filter((group) => normalizeName(group.name).includes(normalizedQuery))
  if (containsMatches.length === 1) {
    return containsMatches[0].id
  }

  return ''
}

export function SchedulePage() {
  const { api, user } = useAuth()
  const [weekOffset, setWeekOffset] = useState(0)
  const [groupQuery, setGroupQuery] = useState('')
  const [groups, setGroups] = useState([])
  const [state, setState] = useState({ loading: true, error: '', lessons: [], queueLessonIds: new Set() })
  const range = useMemo(() => getWeekRange(weekOffset), [weekOffset])
  const dateFrom = useMemo(() => range.from.toISOString().slice(0, 10), [range])
  const dateTo = useMemo(() => range.to.toISOString().slice(0, 10), [range])
  const resolvedAdminGroupId = useMemo(() => resolveGroupId(groups, groupQuery), [groupQuery, groups])
  const requestedGroupId = user?.role === 'admin' ? resolvedAdminGroupId : user?.group?.id || ''

  useEffect(() => {
    let cancelled = false

    async function loadGroups() {
      if (user?.role !== 'admin') {
        setGroups([])
        return
      }

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
  }, [api, user?.role])

  useEffect(() => {
    let cancelled = false

    async function loadLessons() {
      if (!requestedGroupId) {
        setState({
          loading: false,
          error:
            user?.role === 'admin'
              ? groupQuery.trim()
                ? 'Группа не найдена. Уточните название.'
                : 'Укажите название группы, чтобы загрузить расписание.'
              : 'Для пользователя не определена учебная группа.',
          lessons: [],
          queueLessonIds: new Set(),
        })
        return
      }

      setState((current) => ({ ...current, loading: true, error: '' }))

      try {
        const [lessons, queues] = await Promise.all([
          api.getLessons({
            group_id: requestedGroupId,
            date_from: dateFrom,
            date_to: dateTo,
          }),
          api.getQueues({
            group_id: requestedGroupId,
            per_page: 100,
          }),
        ])

        const queueLessonIds = new Set(
          (queues || []).map((queue) => queue.lesson_id).filter(Boolean),
        )

        if (!cancelled) {
          setState({
            loading: false,
            error: '',
            lessons: lessons || [],
            queueLessonIds,
          })
        }
      } catch (error) {
        if (!cancelled) {
          setState({ loading: false, error: error.message, lessons: [], queueLessonIds: new Set() })
        }
      }
    }

    loadLessons()

    return () => {
      cancelled = true
    }
  }, [api, dateFrom, dateTo, groupQuery, requestedGroupId, user?.role])

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
            <Field label="Название группы">
              <TextInput
                placeholder="Например: ИУ7-81Б"
                value={groupQuery}
                onChange={(event) => setGroupQuery(event.target.value)}
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
              {lessons.map((lesson) => {
                const hasQueue = lesson.type === 'lab' && state.queueLessonIds.has(lesson.id)

                return (
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
                      {hasQueue ? <span className="badge badge-success">Есть очередь</span> : null}
                    </div>
                  </article>
                )
              })}
            </div>
          </Panel>
        ))}
    </div>
  )
}
