import { useEffect, useMemo, useRef, useState } from 'react'
import type { FormEvent } from 'react'
import type { ChatMessage, Presence } from '../types/chat'

const now = () => Math.floor(Date.now() / 1000)

export function useChatSocket(roomId: string, token: string | null) {
  const [wsUrl] = useState('ws://localhost:8080/ws')
  // clientId is now derived from token or random, but for now let's keep it random-ish but maybe stable if we had user info
  // simpler to just let server handle identity via token
  const [clientId] = useState(`ui_${Math.random().toString(36).slice(2, 8)}`)
  const [connected, setConnected] = useState(false)
  const [messageText, setMessageText] = useState('')
  const [messages, setMessages] = useState<ChatMessage[]>([])
  const [presence, setPresence] = useState<Presence>({ roomMembers: [], typingUsers: [] })

  const wsRef = useRef<WebSocket | null>(null)
  const typingTimeoutRef = useRef<number | null>(null)

  const onlineCount = useMemo(() => Math.max(1, new Set(presence.roomMembers).size), [presence.roomMembers])

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
      console.log('Joining room:', roomId);
      ws.send(JSON.stringify({ type: 'join_room', room: roomId }))
      // Clear previous messages when switching rooms
      setMessages([])
    }

    ws.onmessage = (event) => {
      try {
        console.log('WS Message:', event.data);
        const raw = JSON.parse(event.data)

        if (raw.type === 'room_joined') {
          console.log('Room Joined:', raw);
          setPresence((prev) => ({ ...prev, roomMembers: raw.members ?? [] }))
          return
        }

        if (raw.type === 'user_typing' && raw.userID && raw.userID !== clientId) {
          // ... existing logic
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
          console.log('User Joined:', raw);
          setPresence((prev) => ({
            ...prev,
            roomMembers: [...new Set([...prev.roomMembers, raw.userId])],
          }))
          return
        }

        if (raw.type === 'user_left_room' && raw.userID) {
          console.log('User Left:', raw);
          setPresence((prev) => ({
            ...prev,
            roomMembers: prev.roomMembers.filter((id) => id !== raw.userID),
          }))
          return
        }

        if (raw.type === 'history_message') {
          console.log('History Message:', raw);
          appendMessage({
            id: crypto.randomUUID(),
            content: raw.content,
            senderName: raw.sender_name ?? `User ${raw.sender}`,
            senderId: String(raw.sender),
            timestamp: Math.floor(new Date(raw.time).getTime() / 1000),
            mine: false, // History doesn't easily convert to "mine" without user ID check, improving later
            kind: 'message',
          })
          return
        }

        if (raw.type === 'room_message') {
          console.log('Room Message Received:', raw);
          // We need to know our own user ID accurately for "mine" to work.
          // For now relying on server echoing sender_id vs our clientId might be flaky if clientId isn't the auth ID.
          // Ideally server sends "sender_id" which matches the decoded JWT sub. 
          // We will assume `raw.sender_id === clientId` logic might need fix if clientId != auth user id.
          // BUT, since we are sending token, server knows us.

          // Temporary fix: check if we just sent this? No, async. 
          // Let's assume for now `mine` needs better check against Auth User ID.

          const mine = raw.sender_id === clientId // This might be wrong if clientId is random.
          // Better: use the token's user ID if available in context. 

          appendMessage({
            id: String(raw.message_id ?? crypto.randomUUID()),
            content: raw.content,
            senderName: raw.sender_name ?? `User ${raw.sender}`,
            senderId: raw.sender_id ?? String(raw.sender),
            timestamp: Number(raw.timestamp ?? now()),
            mine,
            kind: 'message',
          })
          return
        }

        if (raw.type === 'error') {
          console.error("WS Error:", raw.error);
        }
      } catch (e) {
        console.error("WS Parse Error", e);
      }
    }

    ws.onclose = () => {
      console.log('WS Closed');
      setConnected(false)
    }
    ws.onerror = (e) => {
      console.error('WS Error:', e);
      setConnected(false)
    }

    return () => {
      console.log('WS Cleanup');
      ws.close()
    }
  }, [roomId, token, wsUrl, clientId])


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
    typingTimeoutRef.current = window.setTimeout(() => {
      // Optimistic local clear? No, wait for server or just timeout.
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

  return {
    connected,
    messageText,
    messages,
    presence,
    onlineCount,
    handleInput,
    sendMessage,
  }
}
