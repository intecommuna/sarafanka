# Sarafanka

Мини-проект хаба мини-приложений на стеке Go + Vite + TypeScript.

## Структура

- `backend/` — HTTP-сервер Go
- `frontend/` — интерфейс Vite + TypeScript

## Запуск backend

```bash
cd backend
go run .
```

Сервер будет доступен по адресу: http://localhost:8080/api/health

## Запуск frontend

```bash
cd frontend
npm install
npm run dev -- --host 0.0.0.0
```

После запуска интерфейс будет доступен по адресу: http://localhost:5173

## Сборка frontend

```bash
cd frontend
npm run build
```

## CI/CD

Схема: `git push` → GitLab CI → SSH на VPS → `git pull` + сборка + restart сервиса.

### Переменные в GitLab (Settings → CI/CD → Variables)

| Имя | Тип | Значение |
|---|---|---|
| SSH_PRIVATE_KEY | File, masked | Приватный SSH-ключ для доступа с GitLab-раннера на сервер |
| SERVER_HOST | Variable | 72.56.22.16 |
| SERVER_USER | Variable | Юзер на VPS (с правами sudo) |

### Первичная настройка сервера

```bash
bash deploy/setup-server.sh
```

### Серверный SSH-ключ (для `git pull`)

На VPS:

```bash
ssh-keygen -t ed25519 -C "sarafanka-server"
cat ~/.ssh/id_ed25519.pub   # → добавить в GitLab: Settings → SSH Keys
```

### Локальный деплой-ключ (для CI)

```bash
ssh-keygen -t ed25519 -C "sarafanka-deploy" -f ~/.ssh/sarafanka_deploy
cat ~/.ssh/sarafanka_deploy.pub        # → в ~/.ssh/authorized_keys на VPS
cat ~/.ssh/sarafanka_deploy            # → в GitLab: SSH_PRIVATE_KEY (File)
```

### Запуск локально

```bash
# терминал 1 — бэкенд
cd backend && go run main.go

# терминал 2 — фронтенд
cd frontend && npm install && npm run dev
# открыть http://localhost:5173
```

### Проверка после деплоя

```bash
curl http://sarafanka.su/api/health
```

