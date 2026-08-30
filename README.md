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
