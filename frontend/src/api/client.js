import { API_BASE_URL, getApiBaseUrl } from './config.js'

export class ApiError extends Error {
  constructor(message, options = {}) {
    super(message)
    this.name = 'ApiError'
    this.code = options.code
    this.status = options.status
    this.payload = options.payload
  }
}

function withPagination(items, pagination) {
  const normalized = Array.isArray(items) ? items : []
  Object.defineProperty(normalized, 'pagination', {
    value: pagination || null,
    enumerable: false,
  })
  return normalized
}

export class ApiClient {
  constructor({ getAccessToken, setAccessToken, clearSession }) {
    this.getAccessToken = getAccessToken
    this.setAccessToken = setAccessToken
    this.clearSession = clearSession
  }

  async request(path, options = {}) {
    const {
      method = 'GET',
      body,
      params,
      auth = true,
      formData = false,
      retry = true,
    } = options

    const url = new URL(`${API_BASE_URL}${path}`, window.location.origin)

    if (params) {
      Object.entries(params).forEach(([key, value]) => {
        if (value !== undefined && value !== null && value !== '') {
          url.searchParams.set(key, String(value))
        }
      })
    }

    const headers = new Headers()
    if (!formData) {
      headers.set('Content-Type', 'application/json')
    }

    if (auth) {
      const token = this.getAccessToken()
      if (token) {
        headers.set('Authorization', `Bearer ${token}`)
      }
    }

    const response = await fetch(url.toString(), {
      method,
      headers,
      credentials: 'include',
      body: body
        ? formData
          ? body
          : JSON.stringify(body)
        : undefined,
    })

    if (response.status === 401 && auth && retry && path !== '/auth/refresh') {
      try {
        await this.refresh()
        return this.request(path, { ...options, retry: false })
      } catch {
        this.clearSession()
      }
    }

    const isJson = response.headers.get('content-type')?.includes('application/json')
    const payload = isJson ? await response.json() : null

    if (!response.ok) {
      throw new ApiError(
        payload?.error?.message || payload?.message || 'Не удалось выполнить запрос',
        {
          code: payload?.error?.code,
          status: response.status,
          payload,
        },
      )
    }

    return payload?.data ?? null
  }

  refresh() {
    return this.request('/auth/refresh', { method: 'POST', auth: false, retry: false }).then(
      (data) => {
        if (data?.access_token) {
          this.setAccessToken(data.access_token)
        }
        return data
      },
    )
  }

  login(payload) {
    return this.request('/auth/login', { method: 'POST', body: payload, auth: false }).then(
      (data) => {
        if (data?.access_token) {
          this.setAccessToken(data.access_token)
        }
        return data
      },
    )
  }

  register(payload) {
    return this.request('/auth/register', { method: 'POST', body: payload, auth: false })
  }

  logout() {
    return this.request('/auth/logout', { method: 'POST' })
  }

  logoutAll() {
    return this.request('/auth/logout-all', { method: 'POST' })
  }

  getMe() {
    return this.request('/users/me')
  }

  updateMe(payload) {
    return this.request('/users/me', { method: 'PATCH', body: payload })
  }

  getMySlots(params) {
    return this.request('/users/me/slots', { params }).then(async (slots) => {
      if (!Array.isArray(slots)) {
        return []
      }

      return Promise.all(
        slots.map(async (slot) => {
          if (!slot.queue?.id || (slot.queue.subject?.name && slot.queue.status)) {
            return slot
          }

          try {
            const queue = await this.getQueue(slot.queue.id)
            return { ...slot, queue }
          } catch {
            return slot
          }
        }),
      )
    })
  }

  transferHeadmanRole(payload) {
    return this.request('/users/me/headman-role', { method: 'PUT', body: payload })
  }

  getLessons(params) {
    return this.request('/lessons', { params })
  }

  getLesson(id) {
    return this.request(`/lessons/${id}`)
  }

  importLessons(file) {
    const formData = new FormData()
    formData.append('file', file)
    return this.request('/lessons/imports', {
      method: 'POST',
      body: formData,
      formData: true,
    })
  }

  getQueues(params) {
    return this.request('/queues', { params }).then((data) =>
      Array.isArray(data) ? data : withPagination(data?.items, data?.pagination),
    )
  }

  getQueue(id) {
    return this.request(`/queues/${id}`)
  }

  createQueue(payload) {
    return this.request('/queues', {
      method: 'POST',
      body: {
        lesson_id: payload.lesson_id,
        opens_at: payload.opens_at,
        closes_at: payload.closes_at,
        max_size: payload.max_size,
        transfer_failed_from: payload.transfer_failed_from,
      },
    })
  }

  updateQueue(id, payload) {
    return this.request(`/queues/${id}`, { method: 'PATCH', body: payload })
  }

  deleteQueue(id) {
    return this.request(`/queues/${id}`, { method: 'DELETE' })
  }

  getQueueSlots(queueId) {
    return this.request(`/queues/${queueId}/slots`)
  }

  createQueueSlot(queueId) {
    return this.request(`/queues/${queueId}/slots`, { method: 'POST', body: {} })
  }

  updateQueueSlot(queueId, slotId, payload) {
    return this.request(`/queues/${queueId}/slots/${slotId}`, {
      method: 'PATCH',
      body: payload,
    })
  }

  deleteQueueSlot(queueId) {
    return this.request(`/queues/${queueId}/slots/me`, { method: 'DELETE' })
  }

  transferQueueSlots(queueId, payload) {
    return this.request(`/queues/${queueId}/transfers`, { method: 'POST', body: payload })
  }

  getAdminUsers(params) {
    return this.request('/admin/users', { params }).then((data) =>
      Array.isArray(data) ? data : withPagination(data?.items, data?.pagination),
    )
  }

  updateAdminUser(id, payload) {
    return this.request(`/admin/users/${id}`, { method: 'PATCH', body: payload })
  }

  deleteAdminUser(id) {
    return this.request(`/admin/users/${id}`, { method: 'DELETE' })
  }

  getResource(resource) {
    return this.request(`/admin/${resource}`)
  }

  createResource(resource, payload) {
    return this.request(`/admin/${resource}`, { method: 'POST', body: payload })
  }

  updateResource(resource, id, payload) {
    return this.request(`/admin/${resource}/${id}`, { method: 'PATCH', body: payload })
  }

  deleteResource(resource, id) {
    return this.request(`/admin/${resource}/${id}`, { method: 'DELETE' })
  }

  getAdminLogs(params) {
    return this.request('/admin/logs', { params })
  }
}

export { getApiBaseUrl }

