import { Link } from 'react-router-dom'

export function PageHeader({ eyebrow, title, description, actions }) {
  return (
    <div className="page-header">
      <div>
        {eyebrow ? <div className="page-eyebrow">{eyebrow}</div> : null}
        <h1>{title}</h1>
        {description ? <p>{description}</p> : null}
      </div>
      {actions ? <div className="page-actions">{actions}</div> : null}
    </div>
  )
}

export function Panel({ title, description, actions, children, className = '' }) {
  return (
    <section className={`panel ${className}`.trim()}>
      {(title || description || actions) && (
        <div className="panel-head">
          <div>
            {title ? <h2>{title}</h2> : null}
            {description ? <p>{description}</p> : null}
          </div>
          {actions ? <div className="panel-actions">{actions}</div> : null}
        </div>
      )}
      <div className="panel-body">{children}</div>
    </section>
  )
}

export function StatCard({ label, value, hint }) {
  return (
    <div className="stat-card">
      <span>{label}</span>
      <strong>{value}</strong>
      {hint ? <small>{hint}</small> : null}
    </div>
  )
}

export function Badge({ tone = 'default', children }) {
  return <span className={`badge badge-${tone}`}>{children}</span>
}

export function Button({
  children,
  variant = 'primary',
  onClick,
  type = 'button',
  disabled,
  busy = false,
}) {
  return (
    <button className={`button button-${variant}`} type={type} onClick={onClick} disabled={disabled || busy}>
      {busy ? 'Сохраняем...' : children}
    </button>
  )
}

export function Field({ label, hint, error, children }) {
  return (
    <label className="field">
      <span className="field-label">{label}</span>
      {children}
      {hint ? <small className="field-hint">{hint}</small> : null}
      {error ? <small className="field-error">{error}</small> : null}
    </label>
  )
}

export function TextInput(props) {
  return <input className="input" {...props} />
}

export function SelectInput({ children, ...props }) {
  return (
    <select className="input" {...props}>
      {children}
    </select>
  )
}

export function TextArea(props) {
  return <textarea className="input textarea" {...props} />
}

export function InlineMessage({ tone = 'info', children }) {
  return <div className={`inline-message inline-message-${tone}`}>{children}</div>
}

export function LoaderBlock({ label = 'Загружаем...' }) {
  return (
    <div className="loader-block">
      <div className="loader-dot" />
      <p>{label}</p>
    </div>
  )
}

export function EmptyState({ title, description, action }) {
  return (
    <div className="empty-state">
      <h3>{title}</h3>
      <p>{description}</p>
      {action ? <div>{action}</div> : null}
    </div>
  )
}

export function KeyValueList({ items }) {
  return (
    <dl className="kv-list">
      {items.map((item) => (
        <div key={item.label}>
          <dt>{item.label}</dt>
          <dd>{item.value}</dd>
        </div>
      ))}
    </dl>
  )
}

export function QueueCard({ queue, extra, children }) {
  return (
    <article className="queue-card">
      <div className="queue-card-head">
        <div>
          <h3>{queue.subject?.name || 'Очередь без предмета'}</h3>
          <p>{queue.group?.name || 'Группа не указана'}</p>
        </div>
        {extra}
      </div>
      <div className="queue-card-meta">{children}</div>
      <Link className="text-link" to={`/queues/${queue.id}`}>
        Открыть детали
      </Link>
    </article>
  )
}
