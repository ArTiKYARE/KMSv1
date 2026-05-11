export interface User {
  id: string
  email: string
  first_name: string
  last_name: string
  avatar_url?: string
}

export interface AuthTokens {
  access_token: string
  refresh_token: string
  expires_in: number
}

export interface Message {
  id: string
  chat_id: string
  sender_id: string
  content: string
  message_type: 'text' | 'file' | 'system'
  created_at: string
  sender?: User
}

export interface Chat {
  id: string
  name?: string
  description?: string
  chat_type: 'direct' | 'group' | 'channel'
  created_by: string
  created_at: string
  members?: ChatMember[]
}

export interface ChatMember {
  id: string
  chat_id: string
  user_id: string
  role: 'admin' | 'member'
  joined_at: string
  user?: User
}

export interface WSMessage {
  type: 'new_message' | 'user_typing' | 'chat_message' | 'typing'
  payload: any
}
