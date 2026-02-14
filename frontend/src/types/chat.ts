export type ChatMessage = {
  id: string
  content: string
  senderName: string
  senderId: string
  timestamp: number
  mine: boolean
  kind: 'message' | 'system'
}

export type Presence = {
  roomMembers: string[]
  typingUsers: string[]
}

export type Contact = {
  id: string
  name: string
  status: string
}
