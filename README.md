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

## CI/CD и деплой

### Основной канал: GitHub Actions (.github/workflows/deploy.yml)

Секреты в GitHub: Settings → Secrets and variables → Actions:
- `SSH_PRIVATE_KEY` — приватный ключ целиком
- `SERVER_HOST=72.56.22.16`
- `SERVER_USER=root`

### Запасной канал: cron на сервере каждые 2 минуты

Скрипт: `deploy/auto-deploy.sh`

### Зеркалирование

- `origin` = GitHub
- `gitverse` = GitVerse
- `gitlab` = архив

Алиас:

```bash
git pushall
```

`git pushall` пушит в `origin` и `gitverse`.

### Первичная настройка сервера

```bash
bash deploy/setup-server.sh
```

или через cloud-init-скрипт при создании VPS на Timeweb.

### Серверный SSH-ключ (для `git pull`)

На VPS:

```bash
ssh-keygen -t ed25519 -C "sarafanka-server"
cat ~/.ssh/id_ed25519.pub   # → добавить в GitHub: Settings → SSH and GPG keys
```

### Локальный деплой-ключ (для CI)

```bash
ssh-keygen -t ed25519 -C "sarafanka-deploy" -f ~/.ssh/sarafanka_deploy
cat ~/.ssh/sarafanka_deploy.pub        # → в ~/.ssh/authorized_keys на VPS
cat ~/.ssh/sarafanka_deploy            # → в GitHub: SSH_PRIVATE_KEY
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

