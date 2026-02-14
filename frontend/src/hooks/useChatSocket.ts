import { useEffect, useMemo, useRef, useState } from 'react'
import type { FormEvent } from 'react'
import type { ChatMessage, Presence } from '../types/chat'

const now = () => Math.floor(Date.now() / 1000)

export function useChatSocket() {
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

  const onlineCount = useMemo(() => Math.max(1, new Set(presence.roomMembers).size), [presence.roomMembers])

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

  const joinRoom = () => {
    wsRef.current?.send(JSON.stringify({ type: 'join_room', room: roomId }))
  }

  return {
    wsUrl,
    setWsUrl,
    token,
    setToken,
    clientId,
    setClientId,
    roomId,
    setRoomId,
    connected,
    messageText,
    messages,
    presence,
    onlineCount,
    connect,
    joinRoom,
    handleInput,
    sendMessage,
  }
}
