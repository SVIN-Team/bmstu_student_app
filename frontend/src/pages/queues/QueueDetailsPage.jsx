import { useMemo, useCallback, useEffect, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { useAuth } from '../../providers/auth-context.js'
import {
  QUEUE_STATUS_LABELS,
  SLOT_STATUS_LABELS,
  canManageQueues,
  formatDateTime,
  fromDatetimeLocalValue,
  fullName,
  toDatetimeLocalValue,
} from '../../utils/format.js'
import {
  Badge,
  Button,
  EmptyState,
  Field,
  InlineMessage,
  KeyValueList,
  LoaderBlock,
  PageHeader,
  Panel,
  SelectInput,
  TextInput,
} from '../../components/ui/ui.jsx'

export function QueueDetailsPage() {
  const { queueId } = useParams()
  const navigate = useNavigate()
  const { api, user } = useAuth()
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [queue, setQueue] = useState(null)
  const [slots, setSlots] = useState([])
  const [relatedQueues, setRelatedQueues] = useState([])
  const [message, setMessage] = useState('')
  const [editForm, setEditForm] = useState({
    status: '',
    closes_at: '',
    max_size: '',
  })
  const [transferSourceQueue, setTransferSourceQueue] = useState('')

  const loadQueueData = useCallback(async () => {
    setLoading(true)
    setError('')

    try {
      const [queueData, slotsData, queuesData] = await Promise.all([
        api.getQueue(queueId),
        api.getQueueSlots(queueId),
        api.getQueues({ per_page: 100 }),
      ])

      setQueue(queueData)
      setSlots(slotsData || [])
      setRelatedQueues(queuesData || [])
      setEditForm({
        status: queueData.status || '',
        closes_at: toDatetimeLocalValue(queueData.closes_at),
        max_size: queueData.max_size || '',
      })
    } catch (requestError) {
      setError(requestError.message)
    } finally {
      setLoading(false)
    }
  }, [api, queueId])

  useEffect(() => {
    loadQueueData()
  }, [loadQueueData])

  async function handleJoinQueue() {
    setMessage('')
    try {
      await api.createQueueSlot(queueId)
      setMessage('Вы успешно записались в очередь.')
      await loadQueueData()
    } catch (requestError) {
      setMessage(requestError.message)
    }
  }

  async function handleUpdateQueue(event) {
    event.preventDefault()
    setMessage('')
    try {
      await api.updateQueue(queueId, {
        status: editForm.status || undefined,
        closes_at: fromDatetimeLocalValue(editForm.closes_at),
        max_size: editForm.max_size ? Number(editForm.max_size) : undefined,
      })
      setMessage('Очередь обновлена.')
      await loadQueueData()
    } catch (requestError) {
      setMessage(requestError.message)
    }
  }

  async function handleDeleteQueue() {
    if (!window.confirm('Удалить очередь? Это действие нельзя быстро откатить.')) {
      return
    }

    try {
      await api.deleteQueue(queueId)
      navigate('/queues')
    } catch (requestError) {
      setMessage(requestError.message)
    }
  }

  async function handleUpdateSlot(slotId, status) {
    try {
      await api.updateQueueSlot(queueId, slotId, { status })
      await loadQueueData()
    } catch (requestError) {
      setMessage(requestError.message)
    }
  }

  async function handleLeaveQueue() {
    try {
      await api.deleteQueueSlot(queueId)
      setMessage('Запись в очередь отменена.')
      await loadQueueData()
    } catch (requestError) {
      setMessage(requestError.message)
    }
  }

  async function handleTransferFailed() {
    try {
      await api.transferQueueSlots(queueId, { source_queue_id: transferSourceQueue })
      setMessage('Студенты со статусами failed и no_show перенесены в текущую очередь.')
      await loadQueueData()
    } catch (requestError) {
      setMessage(requestError.message)
    }
  }

  const mySlot = useMemo(
    () => slots.find((slot) => slot.student?.id === user?.id),
    [slots, user?.id],
  )
  const manager = canManageQueues(user?.role)
  const transferableQueues = useMemo(
    () =>
      relatedQueues.filter(
        (item) =>
          item.id !== queue?.id &&
          item.subject?.id === queue?.subject?.id &&
          item.group?.id === queue?.group?.id &&
          ['closed', 'archived'].includes(item.status),
      ),
    [queue?.group?.id, queue?.id, queue?.subject?.id, relatedQueues],
  )

  if (loading) {
    return <LoaderBlock label="Собираем детали очереди и список слотов..." />
  }

  if (error) {
    return <InlineMessage tone="danger">{error}</InlineMessage>
  }

  if (!queue) {
    return <EmptyState title="Очередь не найдена" description="Запись удалена или у вас нет доступа." />
  }

  return (
    <div className="page-stack">
      <PageHeader
        eyebrow="Детали очереди"
        title={queue.subject?.name || 'Очередь'}
        description={`Группа ${queue.group?.name || 'не указана'} · статус ${QUEUE_STATUS_LABELS[queue.status] || queue.status}`}
        actions={<Badge tone="lilac">{QUEUE_STATUS_LABELS[queue.status] || queue.status}</Badge>}
      />

      {message ? <InlineMessage tone="info">{message}</InlineMessage> : null}

      <div className="content-grid">
        <Panel title="Паспорт очереди">
          <KeyValueList
            items={[
              { label: 'Предмет', value: queue.subject?.name || '—' },
              { label: 'Группа', value: queue.group?.name || '—' },
              { label: 'Открытие', value: formatDateTime(queue.opens_at) },
              { label: 'Закрытие', value: formatDateTime(queue.closes_at) },
              { label: 'Лимит', value: queue.max_size || 'Без ограничений' },
              { label: 'Создал', value: fullName(queue.created_by) },
            ]}
          />

          {!manager && queue.status === 'open' && !mySlot ? (
            <Button onClick={handleJoinQueue}>Записаться в очередь</Button>
          ) : null}

          {!manager && mySlot?.status === 'waiting' ? (
            <Button variant="ghost" onClick={handleLeaveQueue}>
              Отменить свою запись
            </Button>
          ) : null}
        </Panel>

        {manager ? (
          <Panel title="Управление очередью">
            <form className="form-grid" onSubmit={handleUpdateQueue}>
              <Field label="Статус">
                <SelectInput
                  value={editForm.status}
                  onChange={(event) => setEditForm({ ...editForm, status: event.target.value })}
                >
                  {Object.entries(QUEUE_STATUS_LABELS).map(([value, label]) => (
                    <option key={value} value={value}>
                      {label}
                    </option>
                  ))}
                </SelectInput>
              </Field>
              <Field label="Закрытие">
                <TextInput
                  type="datetime-local"
                  value={editForm.closes_at}
                  onChange={(event) => setEditForm({ ...editForm, closes_at: event.target.value })}
                />
              </Field>
              <Field label="Лимит мест">
                <TextInput
                  type="number"
                  min="1"
                  value={editForm.max_size}
                  onChange={(event) => setEditForm({ ...editForm, max_size: event.target.value })}
                />
              </Field>
              <div className="button-group">
                <Button type="submit">Сохранить очередь</Button>
                <Button variant="danger" onClick={handleDeleteQueue}>
                  Удалить очередь
                </Button>
              </div>
            </form>

            {queue.status === 'draft' ? (
              <div className="transfer-box">
                <Field label="Источник для переноса неуспевших">
                  <SelectInput
                    value={transferSourceQueue}
                    onChange={(event) => setTransferSourceQueue(event.target.value)}
                  >
                    <option value="">Выберите очередь</option>
                    {transferableQueues.map((item) => (
                      <option key={item.id} value={item.id}>
                        {item.subject?.name} · {QUEUE_STATUS_LABELS[item.status]}
                      </option>
                    ))}
                  </SelectInput>
                </Field>
                <Button
                  variant="secondary"
                  onClick={handleTransferFailed}
                  disabled={!transferSourceQueue}
                >
                  Перенести failed / no_show
                </Button>
              </div>
            ) : null}
          </Panel>
        ) : null}
      </div>

      <Panel title="Слоты" description="Список отсортирован по времени записи.">
        {slots.length ? (
          <div className="list-stack">
            {slots.map((slot, index) => (
              <div className="slot-row" key={slot.id}>
                <div>
                  <strong>
                    {index + 1}. {fullName(slot.student)}
                  </strong>
                  <p>{formatDateTime(slot.signed_up_at)}</p>
                </div>
                <div className="slot-actions">
                  <Badge tone={slot.status === 'waiting' ? 'default' : 'success'}>
                    {SLOT_STATUS_LABELS[slot.status] || slot.status}
                  </Badge>
                  {manager ? (
                    <SelectInput
                      value={slot.status}
                      onChange={(event) => handleUpdateSlot(slot.id, event.target.value)}
                    >
                      {Object.entries(SLOT_STATUS_LABELS).map(([value, label]) => (
                        <option key={value} value={value}>
                          {label}
                        </option>
                      ))}
                    </SelectInput>
                  ) : null}
                  {!manager && slot.student?.id === user?.id && slot.status === 'waiting' ? (
                    <Button variant="ghost" onClick={handleLeaveQueue}>
                      Отменить
                    </Button>
                  ) : null}
                </div>
              </div>
            ))}
          </div>
        ) : (
          <EmptyState title="Слотов пока нет" description="Эта очередь еще пустая." />
        )}
      </Panel>
    </div>
  )
}
