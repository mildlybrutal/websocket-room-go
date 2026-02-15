import { useEffect, useMemo, useRef, useState, useCallback } from 'react'
import type { FormEvent } from 'react'
import type { ChatMessage, OnlineUser } from '../types/chat'

const now = () => Math.floor(Date.now() / 1000)

export function useChatSocket(roomId: string, token: string | null, userId?: string) {
  const [wsUrl] = useState('ws://localhost:8080/ws')
  const [clientId] = useState(`ui_${Math.random().toString(36).slice(2, 8)}`)
  const [connected, setConnected] = useState(false)
  const [messageText, setMessageText] = useState('')
  const [messages, setMessages] = useState<ChatMessage[]>([])
  const [typingUsers, setTypingUsers] = useState<string[]>([])
  const [onlineUsers, setOnlineUsers] = useState<OnlineUser[]>([])

  const wsRef = useRef<WebSocket | null>(null)
  const typingTimeoutRef = useRef<number | null>(null)

  const onlineCount = useMemo(() => onlineUsers.length, [onlineUsers])

  // Auto-connect when roomId or token changes
  useEffect(() => {
    if (!token) return;

    // cleanup previous
    wsRef.current?.close()
    if (typingTimeoutRef.current) {
      window.clearTimeout(typingTimeoutRef.current)
    }

    const url = new URL(wsUrl)
    url.searchParams.set('id', clientId)
    url.searchParams.set('token', token)

    const ws = new WebSocket(url.toString())
    wsRef.current = ws

    ws.onopen = () => {
      console.log('WS Open');
      setConnected(true)
      // Join the specific room
      ws.send(JSON.stringify({ type: 'join_room', room: roomId }))
      // Clear previous messages when switching rooms
      setMessages([])
      setTypingUsers([])
      // Request online users
      ws.send(JSON.stringify({ type: 'get_online_users' }))
    }

    ws.onmessage = (event) => {
      try {
        const raw = JSON.parse(event.data)

        switch (raw.type) {
          case 'room_joined': {
            break
          }

          case 'user_typing': {
            if (raw.userID && raw.userID !== clientId) {
              const typingName = raw.username || raw.userID
              setTypingUsers((prev) => {
                if (prev.includes(typingName)) return prev
                return [...prev, typingName]
              })
              window.setTimeout(() => {
                setTypingUsers((prev) => prev.filter((name) => name !== typingName))
              }, 1800)
            }
            break
          }

          case 'user_joined_room': {
            break
          }

          case 'user_left_room': {
            break
          }

          case 'history_message': {
            const senderUserId = Number(raw.sender)
            const mine = userId ? String(senderUserId) === String(userId) : false
            appendMessage({
              id: String(raw.message_id ?? crypto.randomUUID()),
              messageId: raw.message_id ? Number(raw.message_id) : undefined,
              content: raw.content,
              senderName: raw.sender_name ?? `User ${raw.sender}`,
              senderId: String(raw.sender),
              senderUserId,
              timestamp: Math.floor(new Date(raw.time).getTime() / 1000),
              mine,
              kind: 'message',
              isEdited: raw.is_edited ?? false,
              isDeleted: raw.is_deleted ?? false,
              editedAt: raw.edited_at ? Math.floor(new Date(raw.edited_at).getTime() / 1000) : undefined,
            })
            break
          }

          case 'room_message': {
            const senderUserId = Number(raw.sender)
            const mine = userId ? String(senderUserId) === String(userId) : raw.sender_id === clientId
            appendMessage({
              id: String(raw.message_id ?? crypto.randomUUID()),
              messageId: raw.message_id ? Number(raw.message_id) : undefined,
              content: raw.content,
              senderName: raw.sender_name ?? `User ${raw.sender}`,
              senderId: raw.sender_id ?? String(raw.sender),
              senderUserId,
              timestamp: Number(raw.timestamp ?? now()),
              mine,
              kind: 'message',
            })
            break
          }

          case 'message_edited': {
            setMessages((prev) =>
              prev.map((msg) =>
                msg.messageId === Number(raw.message_id)
                  ? {
                    ...msg,
                    content: raw.content,
                    isEdited: true,
                    editedAt: Number(raw.edited_at ?? now()),
                  }
                  : msg
              )
            )
            break
          }

          case 'message_deleted': {
            setMessages((prev) =>
              prev.map((msg) =>
                msg.messageId === Number(raw.message_id)
                  ? {
                    ...msg,
                    content: '[Message deleted]',
                    isDeleted: true,
                  }
                  : msg
              )
            )
            break
          }

          case 'presence_update': {
            if (Array.isArray(raw.online_users)) {
              setOnlineUsers(raw.online_users.map((uid: number) => ({ id: uid, username: `User ${uid}` })))
            }
            break
          }

          case 'error': {
            console.error("WS Error:", raw.error)
            break
          }

          default:
            break
        }
      } catch (e) {
        console.error("WS Parse Error", e)
      }
    }

    ws.onclose = () => {
      setConnected(false)
    }
    ws.onerror = () => {
      setConnected(false)
    }

    return () => {
      ws.close()
    }
  }, [roomId, token, wsUrl, clientId, userId])

  const appendMessage = (message: ChatMessage) => {
    setMessages((prev) => [...prev.slice(-199), message])
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
    typingTimeoutRef.current = window.setTimeout(() => { }, 1200)
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

  const editMessage = useCallback((messageId: number, newContent: string) => {
    if (!wsRef.current || wsRef.current.readyState !== WebSocket.OPEN) return
    wsRef.current.send(
      JSON.stringify({
        type: 'edit_message',
        room: roomId,
        message_id: messageId,
        content: newContent,
      }),
    )
  }, [roomId])

  const deleteMessage = useCallback((messageId: number) => {
    if (!wsRef.current || wsRef.current.readyState !== WebSocket.OPEN) return
    wsRef.current.send(
      JSON.stringify({
        type: 'delete_message',
        room: roomId,
        message_id: messageId,
      }),
    )
  }, [roomId])

  const requestOnlineUsers = useCallback(() => {
    if (!wsRef.current || wsRef.current.readyState !== WebSocket.OPEN) return
    wsRef.current.send(JSON.stringify({ type: 'get_online_users' }))
  }, [])

  return {
    connected,
    messageText,
    messages,
    typingUsers,
    onlineCount,
    onlineUsers,
    handleInput,
    sendMessage,
    editMessage,
    deleteMessage,
    requestOnlineUsers,
  }
}
