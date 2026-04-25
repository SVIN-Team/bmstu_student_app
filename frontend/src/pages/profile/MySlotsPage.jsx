import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { useAuth } from '../../providers/auth-context.js'
import { QUEUE_STATUS_LABELS, SLOT_STATUS_LABELS, formatDateTime } from '../../utils/format.js'
import {
  EmptyState,
  Field,
  LoaderBlock,
  PageHeader,
  Panel,
  SelectInput,
} from '../../components/ui/ui.jsx'

export function MySlotsPage() {
  const { api } = useAuth()
  const [filters, setFilters] = useState({ status: '', queue_status: '' })
  const [state, setState] = useState({ loading: true, error: '', slots: [] })

  useEffect(() => {
    async function load() {
      setState((current) => ({ ...current, loading: true, error: '' }))

      try {
        const slots = await api.getMySlots(filters)
        setState({ loading: false, error: '', slots: slots || [] })
      } catch (error) {
        setState({ loading: false, error: error.message, slots: [] })
      }
    }

    load()
  }, [api, filters])

  return (
    <div className="page-stack">
      <PageHeader
        eyebrow="Мои записи"
        title="История слотов"
        description="Здесь можно отслеживать статусы своих записей и быстро открывать нужную очередь."
      />

      <Panel title="Фильтры">
        <div className="form-row">
          <Field label="Статус слота">
            <SelectInput
              value={filters.status}
              onChange={(event) => setFilters({ ...filters, status: event.target.value })}
            >
              <option value="">Все</option>
              {Object.entries(SLOT_STATUS_LABELS).map(([value, label]) => (
                <option key={value} value={value}>
                  {label}
                </option>
              ))}
            </SelectInput>
          </Field>
          <Field label="Статус очереди">
            <SelectInput
              value={filters.queue_status}
              onChange={(event) => setFilters({ ...filters, queue_status: event.target.value })}
            >
              <option value="">Все</option>
              <option value="open">Открыта</option>
              <option value="closed">Закрыта</option>
            </SelectInput>
          </Field>
        </div>
      </Panel>

      {state.loading ? <LoaderBlock label="Загружаем ваши слоты..." /> : null}
      {state.error ? <div className="inline-message inline-message-danger">{state.error}</div> : null}

      {!state.loading && !state.slots.length ? (
        <EmptyState title="Записей нет" description="Под выбранные фильтры ничего не найдено." />
      ) : null}

      {state.slots.map((slot) => (
        <Panel key={slot.id} title={slot.queue?.subject?.name || 'Очередь'}>
          <div className="slot-row">
            <div>
              <p>Слот: {SLOT_STATUS_LABELS[slot.status] || slot.status}</p>
              <p>Очередь: {QUEUE_STATUS_LABELS[slot.queue?.status] || slot.queue?.status}</p>
              <p>Записан: {formatDateTime(slot.signed_up_at)}</p>
            </div>
            <Link className="button button-secondary" to={`/queues/${slot.queue?.id}`}>
              Открыть очередь
            </Link>
          </div>
        </Panel>
      ))}
    </div>
  )
}
