import { create } from 'zustand';
import { Chat, Message } from '../types';

interface ChatStore {
  chats: Chat[];
  currentChat: Chat | null;
  messages: Message[];
  wsConnected: boolean;
  ws: WebSocket | null;
  setChats: (chats: Chat[]) => void;
  setCurrentChat: (chat: Chat | null) => void;
  addMessage: (message: Message) => void;
  setMessages: (messages: Message[]) => void;
  setWsConnected: (connected: boolean) => void;
  setWs: (ws: WebSocket | null) => void;
}

export const useChatStore = create<ChatStore>((set) => ({
  chats: [],
  currentChat: null,
  messages: [],
  wsConnected: false,
  ws: null,

  setChats: (chats) => set({ chats }),
  
  setCurrentChat: (chat) => set({ currentChat: chat }),
  
  addMessage: (message) => set((state) => ({ 
    messages: [...state.messages, message] 
  })),
  
  setMessages: (messages) => set({ messages }),
  
  setWsConnected: (connected) => set({ wsConnected: connected }),
  
  setWs: (ws) => set({ ws }),
}));
