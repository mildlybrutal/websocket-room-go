import { CONTACTS } from '../constants'

export function LeftPanel() {
  return (
    <aside className="left-panel">
      <div className="brand">Messages</div>
      <div className="search">Search chats…</div>
      <div className="contact-list">
        {CONTACTS.map((contact) => (
          <article key={contact.id} className={`contact ${contact.id === '1' ? 'active' : ''}`}>
            <div className="avatar">{contact.name[0]}</div>
            <div>
              <h4>{contact.name}</h4>
              <p>{contact.status}</p>
            </div>
          </article>
        ))}
      </div>
    </aside>
  )
}
