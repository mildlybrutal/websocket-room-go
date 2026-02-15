export type ChatMessage = {
  id: string
  messageId?: number
  content: string
  senderName: string
  senderId: string
  senderUserId?: number
  timestamp: number
  mine: boolean
  kind: 'message' | 'system'
  isEdited?: boolean
  isDeleted?: boolean
  editedAt?: number
}

export type Presence = {
  roomMembers: string[]
  typingUsers: string[]
}

export type OnlineUser = {
  id: number
  username: string
}

export type Contact = {
  id: string
  name: string
  status: string
}
