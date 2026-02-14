import { useEffect, useMemo, useRef, useState } from 'react'
import type { FormEvent } from 'react'
import './App.css'

type ChatMessage = {
  id: string
  content: string
  senderName: string
  senderId: string
  timestamp: number
  mine: boolean
  kind: 'message' | 'system'
}

type Presence = {
  roomMembers: string[]
  typingUsers: string[]
}

const CONTACTS = [
  { id: '1', name: 'Product Team', status: 'Active now' },
  { id: '2', name: 'Marketing', status: 'Offline' },
  { id: '3', name: 'Engineers', status: '12 members online' },
]

const now = () => Math.floor(Date.now() / 1000)

function App() {
  const [wsUrl, setWsUrl] = useState('ws://localhost:8080/ws')
  const [token, setToken] = useState('')
  const [clientId, setClientId] = useState(`ui_${Math.random().toString(36).slice(2, 8)}`)
  const [roomId, setRoomId] = useState('1')
  const [connected, setConnected] = useState(false)
  const [messageText, setMessageText] = useState('')
  const [messages, setMessages] = useState<ChatMessage[]>([])
  const [presence, setPresence] = useState<Presence>({ roomMembers: [], typingUsers: [] })

  const wsRef = useRef<WebSocket | null>(null)
  const typingTimeoutRef = useRef<number | null>(null)

  const onlineCount = useMemo(
    () => Math.max(1, new Set(presence.roomMembers).size),
    [presence.roomMembers],
  )

  useEffect(() => {
    return () => {
      wsRef.current?.close()
      if (typingTimeoutRef.current) {
        window.clearTimeout(typingTimeoutRef.current)
      }
    }
  }, [])

  const appendMessage = (message: ChatMessage) => {
    setMessages((prev) => [...prev.slice(-199), message])
  }

  const connect = () => {
    wsRef.current?.close()

    const url = new URL(wsUrl)
    url.searchParams.set('id', clientId)
    if (token.trim()) {
      url.searchParams.set('token', token.trim())
    }

    const ws = new WebSocket(url.toString())
    wsRef.current = ws

    ws.onopen = () => {
      setConnected(true)
      ws.send(JSON.stringify({ type: 'join_room', room: roomId }))
      appendMessage({
        id: crypto.randomUUID(),
        content: `Connected and joined room ${roomId}`,
        senderName: 'System',
        senderId: 'system',
        timestamp: now(),
        mine: false,
        kind: 'system',
      })
    }

    ws.onmessage = (event) => {
      const raw = JSON.parse(event.data)

      if (raw.type === 'room_joined') {
        setPresence((prev) => ({ ...prev, roomMembers: raw.members ?? [] }))
        return
      }

      if (raw.type === 'user_typing' && raw.userID && raw.userID !== clientId) {
        setPresence((prev) => {
          if (prev.typingUsers.includes(raw.userID)) return prev
          return { ...prev, typingUsers: [...prev.typingUsers, raw.userID] }
        })
        window.setTimeout(() => {
          setPresence((prev) => ({
            ...prev,
            typingUsers: prev.typingUsers.filter((id) => id !== raw.userID),
          }))
        }, 1800)
        return
      }

      if (raw.type === 'user_joined_room' && raw.userId) {
        setPresence((prev) => ({
          ...prev,
          roomMembers: [...new Set([...prev.roomMembers, raw.userId])],
        }))
        return
      }

      if (raw.type === 'user_left_room' && raw.userID) {
        setPresence((prev) => ({
          ...prev,
          roomMembers: prev.roomMembers.filter((id) => id !== raw.userID),
        }))
        return
      }

      if (raw.type === 'history_message') {
        appendMessage({
          id: crypto.randomUUID(),
          content: raw.content,
          senderName: `User ${raw.sender}`,
          senderId: String(raw.sender),
          timestamp: Math.floor(new Date(raw.time).getTime() / 1000),
          mine: false,
          kind: 'message',
        })
        return
      }

      if (raw.type === 'room_message') {
        const mine = raw.sender_id === clientId
        appendMessage({
          id: String(raw.message_id ?? crypto.randomUUID()),
          content: raw.content,
          senderName: mine ? 'You' : raw.sender_id ?? `User ${raw.sender}`,
          senderId: raw.sender_id ?? String(raw.sender),
          timestamp: Number(raw.timestamp ?? now()),
          mine,
          kind: 'message',
        })
        return
      }

      if (raw.type === 'error') {
        appendMessage({
          id: crypto.randomUUID(),
          content: raw.error || 'Unknown error',
          senderName: 'System',
          senderId: 'system',
          timestamp: now(),
          mine: false,
          kind: 'system',
        })
      }
    }

    ws.onclose = () => setConnected(false)
    ws.onerror = () => setConnected(false)
  }

  const sendTyping = () => {
    if (!wsRef.current || wsRef.current.readyState !== WebSocket.OPEN) return
    wsRef.current.send(JSON.stringify({ type: 'typing', room: roomId }))
  }

  const handleInput = (value: string) => {
    setMessageText(value)
    sendTyping()

    if (typingTimeoutRef.current) {
      window.clearTimeout(typingTimeoutRef.current)
    }

    typingTimeoutRef.current = window.setTimeout(() => {
      setPresence((prev) => ({ ...prev, typingUsers: prev.typingUsers.filter((id) => id !== clientId) }))
    }, 1200)
  }

  const sendMessage = (event: FormEvent) => {
    event.preventDefault()
    if (!messageText.trim()) return
    if (!wsRef.current || wsRef.current.readyState !== WebSocket.OPEN) return

    wsRef.current.send(
      JSON.stringify({
        type: 'room_message',
        room: roomId,
        content: messageText.trim(),
      }),
    )

    setMessageText('')
  }

  return (
    <div className="app-shell">
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
          {presence.typingUsers.length > 0 && (
            <div className="typing-indicator">{presence.typingUsers.join(', ')} typing…</div>
          )}
          <form onSubmit={sendMessage} className="composer">
            <input
              value={messageText}
              onChange={(event) => handleInput(event.target.value)}
              placeholder="Type your message…"
            />
            <button type="submit">Send</button>
          </form>
        </footer>
      </main>

      <aside className="right-panel">
        <h3>Connection</h3>
        <label>
          WebSocket URL
          <input value={wsUrl} onChange={(event) => setWsUrl(event.target.value)} />
        </label>
        <label>
          Client ID
          <input value={clientId} onChange={(event) => setClientId(event.target.value)} />
        </label>
        <label>
          JWT token (optional)
          <input value={token} onChange={(event) => setToken(event.target.value)} placeholder="Only needed if require_auth=true" />
        </label>
        <label>
          Room ID
          <input value={roomId} onChange={(event) => setRoomId(event.target.value)} />
        </label>
        <div className="button-row">
          <button className="secondary" onClick={connect}>
            {connected ? 'Reconnect' : 'Connect'}
          </button>
          <button
            className="secondary"
            onClick={() => wsRef.current?.send(JSON.stringify({ type: 'join_room', room: roomId }))}
            disabled={!connected}
          >
            Join room
          </button>
        </div>
      </aside>
    </div>
  )
}

export default App
