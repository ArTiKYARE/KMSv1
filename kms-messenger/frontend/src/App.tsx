import { useEffect } from 'react';
import { useAuthStore } from './store/authStore';
import { AuthPage } from './components/AuthPage';
import { ChatPage } from './components/ChatPage';

function App() {
  const { isAuthenticated, loadFromStorage } = useAuthStore();

  useEffect(() => {
    loadFromStorage();
  }, []);

  return isAuthenticated ? <ChatPage /> : <AuthPage />;
}

export default App;
