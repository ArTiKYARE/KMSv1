export interface User {
  id: number;
  email: string;
  first_name: string;
  last_name: string;
  avatar_url?: string;
}

export interface Chat {
  id: number;
  name?: string;
  type: 'direct' | 'group';
  created_by: number;
  created_at: string;
  members: Member[];
}

export interface Member {
  id: number;
  email: string;
  first_name: string;
  last_name: string;
  avatar_url?: string;
  role: 'admin' | 'member';
}

export interface Message {
  id: number;
  chat_id: number;
  sender_id: number;
  content: string;
  message_type: 'text' | 'file' | 'system';
  file_url?: string;
  created_at: string;
  sender: UserSummary;
}

export interface UserSummary {
  id: number;
  first_name: string;
  last_name: string;
  avatar_url?: string;
}

export interface AuthState {
  user: User | null;
  accessToken: string | null;
  refreshToken: string | null;
  isAuthenticated: boolean;
}

export interface ChatState {
  chats: Chat[];
  currentChat: Chat | null;
  messages: Message[];
  wsConnected: boolean;
}
