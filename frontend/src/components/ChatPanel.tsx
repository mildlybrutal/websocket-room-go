import type { FormEvent } from 'react'
import type { ChatMessage } from '../types/chat'

type ChatPanelProps = {
  roomId: string
  onlineCount: number
  connected: boolean
  messages: ChatMessage[]
  typingUsers: string[]
  messageText: string
  onInputChange: (value: string) => void
  onSendMessage: (event: FormEvent) => void
}

export function ChatPanel({
  roomId,
  onlineCount,
  connected,
  messages,
  typingUsers,
  messageText,
  onInputChange,
  onSendMessage,
}: ChatPanelProps) {
  return (
    <main className="chat-panel">
      <header className="chat-header">
        <div>
          <h2>Room {roomId}</h2>
          <p>{onlineCount} active member{onlineCount > 1 ? 's' : ''}</p>
        </div>
        <div className={`status-pill ${connected ? 'on' : ''}`}>{connected ? 'Connected' : 'Disconnected'}</div>
      </header>

      <section className="messages">
        {messages.length === 0 && (
          <p className="empty-state">Join a room and start chatting. Messages update in real-time over WebSocket.</p>
        )}
        {messages.map((message) => (
          <article
            key={message.id}
            className={`message ${message.mine ? 'mine' : ''} ${message.kind === 'system' ? 'system' : ''}`}
          >
            <strong>{message.senderName}</strong>
            <p>{message.content}</p>
            <time>{new Date(message.timestamp * 1000).toLocaleTimeString()}</time>
          </article>
        ))}
      </section>

      <footer className="composer-wrap">
        {typingUsers.length > 0 && <div className="typing-indicator">{typingUsers.join(', ')} typing…</div>}
        <form onSubmit={onSendMessage} className="composer">
          <input
            value={messageText}
            onChange={(event) => onInputChange(event.target.value)}
            placeholder="Type your message…"
          />
          <button type="submit">Send</button>
        </form>
      </footer>
    </main>
  )
}
