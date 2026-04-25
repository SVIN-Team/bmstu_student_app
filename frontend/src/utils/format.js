export const ROLE_LABELS = {
  student: 'Студент',
  headman: 'Староста',
  admin: 'Администратор',
}

export const QUEUE_STATUS_LABELS = {
  draft: 'Черновик',
  open: 'Открыта',
  closed: 'Закрыта',
  archived: 'Архив',
}

export const SLOT_STATUS_LABELS = {
  waiting: 'Ожидание',
  passed: 'Сдал',
  failed: 'Не сдал',
  no_show: 'Не явился',
}

export const LESSON_TYPE_LABELS = {
  lecture: 'Лекция',
  lab: 'Лабораторная',
  seminar: 'Семинар',
}

export function formatDateTime(value) {
  if (!value) {
    return '—'
  }

  return new Intl.DateTimeFormat('ru-RU', {
    dateStyle: 'medium',
    timeStyle: 'short',
  }).format(new Date(value))
}

export function formatDate(value) {
  if (!value) {
    return '—'
  }

  return new Intl.DateTimeFormat('ru-RU', {
    weekday: 'long',
    day: 'numeric',
    month: 'long',
  }).format(new Date(value))
}

export function formatShortDate(value) {
  if (!value) {
    return ''
  }

  return new Intl.DateTimeFormat('ru-RU', {
    day: '2-digit',
    month: '2-digit',
    year: 'numeric',
  }).format(new Date(value))
}

export function toDateInputValue(date) {
  return new Date(date).toISOString().slice(0, 10)
}

export function toDatetimeLocalValue(value) {
  if (!value) {
    return ''
  }

  const date = new Date(value)
  const local = new Date(date.getTime() - date.getTimezoneOffset() * 60000)
  return local.toISOString().slice(0, 16)
}

export function fromDatetimeLocalValue(value) {
  if (!value) {
    return undefined
  }

  return new Date(value).toISOString()
}

export function getWeekRange(offset = 0) {
  const now = new Date()
  const currentDay = now.getDay() || 7
  const monday = new Date(now)
  monday.setHours(0, 0, 0, 0)
  monday.setDate(now.getDate() - currentDay + 1 + offset * 7)

  const sunday = new Date(monday)
  sunday.setDate(monday.getDate() + 6)
  sunday.setHours(23, 59, 59, 999)

  return { from: monday, to: sunday }
}

export function fullName(user) {
  if (!user) {
    return '—'
  }

  return [user.last_name, user.first_name, user.patronymic].filter(Boolean).join(' ')
}

export function groupLessonsByDate(lessons) {
  return lessons.reduce((acc, lesson) => {
    const key = lesson.starts_at.slice(0, 10)
    if (!acc[key]) {
      acc[key] = []
    }
    acc[key].push(lesson)
    return acc
  }, {})
}

export function canManageQueues(role) {
  return role === 'headman' || role === 'admin'
}
