# KMS Messenger - Corporate Messaging System

Закрытая корпоративная мессенджер-система для внутреннего использования компании.

## 🚀 Быстрый старт

### Требования
- Docker и Docker Compose
- 2GB+ свободной памяти
- Порты 3000, 8080, 5432, 6379 должны быть свободны

### Запуск

```bash
cd kms-messenger

# Запустить все сервисы
docker-compose -f infra/docker-compose.yml up --build

# Или в фоновом режиме
docker-compose -f infra/docker-compose.yml up -d --build
```

### Доступ к приложению
- **Frontend:** http://localhost:3000
- **Backend API:** http://localhost:8080
- **Health Check:** http://localhost:8080/health

## 📋 API Endpoints

### Аутентификация
| Метод | Endpoint | Описание |
|-------|----------|----------|
| POST | `/api/auth/register` | Регистрация пользователя |
| POST | `/api/auth/login` | Вход в систему |
| POST | `/api/auth/refresh` | Обновление токена |
| POST | `/api/auth/logout` | Выход из системы |

### Чаты
| Метод | Endpoint | Описание |
|-------|----------|----------|
| GET | `/api/chats` | Получить список чатов |
| POST | `/api/chats` | Создать чат |
| GET | `/api/chats/:id/messages` | Получить сообщения чата |
| POST | `/api/chats/:id/messages` | Отправить сообщение |
| GET | `/api/chats/ws` | WebSocket подключение |

## 🧪 Тестирование

### Регистрация пользователя
```bash
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "password123",
    "first_name": "John",
    "last_name": "Doe"
  }'
```

### Вход
```bash
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "password123"
  }'
```

### Создание чата
```bash
curl -X POST http://localhost:8080/api/chats \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -d '{
    "type": "direct"
  }'
```

## 🛠 Технологии

### Backend
- Go 1.19+
- Gin (HTTP framework)
- GORM (ORM)
- PostgreSQL (БД)
- Redis (кэш, Pub/Sub)
- gorilla/websocket (WebSocket)
- JWT (аутентификация)

### Frontend
- React 18 + TypeScript
- Vite (сборка)
- Zustand (state management)
- Tailwind CSS (стилизация)
- WebSocket (real-time)

### Инфраструктура
- Docker & Docker Compose
- PostgreSQL 16
- Redis 7

## 📁 Структура проекта

```
kms-messenger/
├── backend/
│   ├── cmd/server/main.go
│   ├── internal/
│   │   ├── config/
│   │   ├── database/
│   │   ├── handlers/
│   │   ├── models/
│   │   └── websocket/
│   ├── go.mod
│   └── Dockerfile
├── frontend/
│   ├── src/
│   │   ├── components/
│   │   ├── store/
│   │   ├── types/
│   │   └── App.tsx
│   ├── package.json
│   └── Dockerfile
├── infra/
│   └── docker-compose.yml
├── .env
└── README.md
```

## 🔐 Безопасность

- JWT аутентификация с Access и Refresh токенами
- Хеширование паролей (bcrypt)
- HTTPS готовность (через reverse proxy)
- Изолированная сеть Docker
- Health checks для всех сервисов

## 📝 Следующие шаги (MVP → Этап 2)

1. ✅ JWT аутентификация
2. ✅ Личные и групповые чаты
3. ✅ WebSocket real-time обновления
4. ✅ Базовая админка (через UI)
5. ⏳ Загрузка файлов (MinIO)
6. ⏳ Антивирус проверка (ClamAV)
7. ⏳ Версионирование файлов

## 🤝 Поддержка

Для вопросов и предложений обратитесь к команде разработки.
