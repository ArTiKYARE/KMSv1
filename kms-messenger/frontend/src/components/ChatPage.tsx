import { useState, useEffect } from 'react';
import { useAuthStore } from '../store/authStore';
import { useChatStore } from '../store/chatStore';
import { Chat, Message } from '../types';

export function ChatPage() {
  const { user, logout, accessToken } = useAuthStore();
  const { 
    chats, setChats, currentChat, setCurrentChat, 
    messages, setMessages, addMessage, wsConnected, setWsConnected, setWs 
  } = useChatStore();
  
  const [newMessage, setNewMessage] = useState('');
  const [showNewChatModal, setShowNewChatModal] = useState(false);
  const [newChatType, setNewChatType] = useState<'direct' | 'group'>('direct');
  const [newChatName, setNewChatName] = useState('');

  // Fetch chats on mount
  useEffect(() => {
    fetchChats();
  }, []);

  // WebSocket connection
  useEffect(() => {
    if (!accessToken) return;

    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const wsUrl = `${protocol}//${window.location.host}/api/chats/ws`;
    
    const ws = new WebSocket(wsUrl);
    
    ws.onopen = () => {
      console.log('WebSocket connected');
      setWsConnected(true);
    };
    
    ws.onclose = () => {
      console.log('WebSocket disconnected');
      setWsConnected(false);
      setWs(null);
    };
    
    ws.onerror = (error) => {
      console.error('WebSocket error:', error);
    };
    
    ws.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data);
        if (data.type === 'chat_message' && data.content) {
          const message = JSON.parse(data.content);
          addMessage(message);
        }
      } catch (e) {
        console.error('Failed to parse WebSocket message:', e);
      }
    };
    
    setWs(ws);
    
    return () => {
      if (ws.readyState === WebSocket.OPEN) {
        ws.close();
      }
    };
  }, [accessToken]);

  const fetchChats = async () => {
    try {
      const response = await fetch('/api/chats', {
        headers: { 'Authorization': `Bearer ${accessToken}` },
      });
      if (response.ok) {
        const data = await response.json();
        setChats(data);
      }
    } catch (e) {
      console.error('Failed to fetch chats:', e);
    }
  };

  const selectChat = async (chat: Chat) => {
    setCurrentChat(chat);
    try {
      const response = await fetch(`/api/chats/${chat.id}/messages`, {
        headers: { 'Authorization': `Bearer ${accessToken}` },
      });
      if (response.ok) {
        const data = await response.json();
        setMessages(data);
      }
    } catch (e) {
      console.error('Failed to fetch messages:', e);
    }
  };

  const sendMessage = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!newMessage.trim() || !currentChat) return;

    try {
      const response = await fetch(`/api/chats/${currentChat.id}/messages`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${accessToken}`,
        },
        body: JSON.stringify({ content: newMessage }),
      });

      if (response.ok) {
        const message = await response.json();
        addMessage(message);
        setNewMessage('');
        
        // Send via WebSocket for real-time delivery
        const ws = useChatStore.getState().ws;
        if (ws && ws.readyState === WebSocket.OPEN) {
          ws.send(JSON.stringify({
            type: 'chat_message',
            chat_id: currentChat.id,
            sender_id: user?.id,
            content: JSON.stringify(message),
          }));
        }
      }
    } catch (e) {
      console.error('Failed to send message:', e);
    }
  };

  const createChat = async (e: React.FormEvent) => {
    e.preventDefault();
    
    try {
      const response = await fetch('/api/chats', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${accessToken}`,
        },
        body: JSON.stringify({
          type: newChatType,
          name: newChatType === 'group' ? newChatName : undefined,
        }),
      });

      if (response.ok) {
        setShowNewChatModal(false);
        setNewChatName('');
        fetchChats();
      }
    } catch (e) {
      console.error('Failed to create chat:', e);
    }
  };

  const handleLogout = async () => {
    try {
      await fetch('/api/auth/logout', {
        method: 'POST',
        headers: { 'Authorization': `Bearer ${accessToken}` },
      });
    } catch (e) {
      console.error('Logout failed:', e);
    }
    logout();
  };

  return (
    <div className="h-screen flex">
      {/* Sidebar */}
      <div className="w-64 bg-gray-800 text-white flex flex-col">
        <div className="p-4 border-b border-gray-700">
          <h2 className="text-xl font-bold">KMS Messenger</h2>
          <p className="text-sm text-gray-400">{user?.first_name} {user?.last_name}</p>
        </div>
        
        <div className="flex-1 overflow-y-auto p-2">
          <div className="flex justify-between items-center mb-2">
            <h3 className="font-semibold text-gray-300">Chats</h3>
            <button
              onClick={() => setShowNewChatModal(true)}
              className="text-blue-400 hover:text-blue-300 text-sm"
            >
              + New
            </button>
          </div>
          
          {chats.map((chat) => (
            <button
              key={chat.id}
              onClick={() => selectChat(chat)}
              className={`w-full text-left p-2 rounded mb-1 ${
                currentChat?.id === chat.id ? 'bg-blue-600' : 'hover:bg-gray-700'
              }`}
            >
              <div className="font-medium truncate">
                {chat.type === 'direct' 
                  ? chat.members.find(m => m.id !== user?.id)?.first_name || 'Chat'
                  : chat.name || 'Group Chat'}
              </div>
              <div className="text-xs text-gray-400">
                {chat.members.length} members
              </div>
            </button>
          ))}
        </div>
        
        <div className="p-4 border-t border-gray-700">
          <button
            onClick={handleLogout}
            className="w-full bg-red-600 hover:bg-red-700 text-white py-2 px-4 rounded"
          >
            Logout
          </button>
        </div>
      </div>

      {/* Main Chat Area */}
      <div className="flex-1 flex flex-col bg-gray-100">
        {currentChat ? (
          <>
            {/* Chat Header */}
            <div className="bg-white border-b border-gray-200 p-4">
              <h2 className="text-lg font-bold">
                {currentChat.type === 'direct'
                  ? currentChat.members.find(m => m.id !== user?.id)?.first_name || 'Chat'
                  : currentChat.name || 'Group Chat'}
              </h2>
              <div className="flex items-center gap-2 mt-1">
                <span className={`w-2 h-2 rounded-full ${wsConnected ? 'bg-green-500' : 'bg-red-500'}`} />
                <span className="text-sm text-gray-500">
                  {wsConnected ? 'Connected' : 'Disconnected'}
                </span>
              </div>
            </div>

            {/* Messages */}
            <div className="flex-1 overflow-y-auto p-4 space-y-3">
              {messages.map((msg) => (
                <div
                  key={msg.id}
                  className={`flex ${msg.sender_id === user?.id ? 'justify-end' : 'justify-start'}`}
                >
                  <div
                    className={`max-w-md px-4 py-2 rounded-lg ${
                      msg.sender_id === user?.id
                        ? 'bg-blue-600 text-white'
                        : 'bg-white text-gray-800'
                    }`}
                  >
                    <div className="text-sm">{msg.content}</div>
                    <div className={`text-xs mt-1 ${
                      msg.sender_id === user?.id ? 'text-blue-200' : 'text-gray-500'
                    }`}>
                      {new Date(msg.created_at).toLocaleTimeString()}
                    </div>
                  </div>
                </div>
              ))}
            </div>

            {/* Message Input */}
            <form onSubmit={sendMessage} className="bg-white border-t border-gray-200 p-4">
              <div className="flex gap-2">
                <input
                  type="text"
                  value={newMessage}
                  onChange={(e) => setNewMessage(e.target.value)}
                  placeholder="Type a message..."
                  className="flex-1 px-4 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                />
                <button
                  type="submit"
                  disabled={!newMessage.trim()}
                  className="bg-blue-600 text-white px-6 py-2 rounded-md hover:bg-blue-700 disabled:opacity-50"
                >
                  Send
                </button>
              </div>
            </form>
          </>
        ) : (
          <div className="flex-1 flex items-center justify-center text-gray-500">
            Select a chat or create a new one
          </div>
        )}
      </div>

      {/* New Chat Modal */}
      {showNewChatModal && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center">
          <div className="bg-white rounded-lg p-6 w-full max-w-md">
            <h3 className="text-lg font-bold mb-4">Create New Chat</h3>
            <form onSubmit={createChat} className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Type
                </label>
                <select
                  value={newChatType}
                  onChange={(e) => setNewChatType(e.target.value as 'direct' | 'group')}
                  className="w-full px-3 py-2 border border-gray-300 rounded-md"
                >
                  <option value="direct">Direct Message</option>
                  <option value="group">Group Chat</option>
                </select>
              </div>
              
              {newChatType === 'group' && (
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">
                    Group Name
                  </label>
                  <input
                    type="text"
                    value={newChatName}
                    onChange={(e) => setNewChatName(e.target.value)}
                    required={newChatType === 'group'}
                    className="w-full px-3 py-2 border border-gray-300 rounded-md"
                  />
                </div>
              )}
              
              <div className="flex gap-2 justify-end">
                <button
                  type="button"
                  onClick={() => setShowNewChatModal(false)}
                  className="px-4 py-2 text-gray-600 hover:text-gray-800"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700"
                >
                  Create
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}
