#!/usr/bin/env bash
set -e

GITHUB_USER="intecommuna"
REPO_SSH="git@github.com:${GITHUB_USER}/sarafanka.git"
PROJECT_DIR="/opt/sarafanka"

echo ">>> Обновление системы и установка зависимостей..."
sudo apt update -qq
sudo apt install -y git nginx nodejs npm golang

echo ">>> Клонирование репозитория в ${PROJECT_DIR}..."
if [ ! -d "${PROJECT_DIR}" ]; then
    sudo git clone ${REPO_SSH} ${PROJECT_DIR}
    sudo chown -R $USER:$USER ${PROJECT_DIR}
fi

echo ">>> Сборка frontend..."
cd ${PROJECT_DIR}/frontend
npm install
npm run build

echo ">>> Сборка backend..."
cd ${PROJECT_DIR}/backend
go build -o app .

echo ">>> Настройка systemd-сервиса sarafanka..."
sudo cp ${PROJECT_DIR}/deploy/sarafanka.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable --now sarafanka

echo ">>> Настройка Nginx..."
sudo cp ${PROJECT_DIR}/deploy/nginx-sarafanka.conf /etc/nginx/sites-available/sarafanka
sudo ln -sf /etc/nginx/sites-available/sarafanka /etc/nginx/sites-enabled/sarafanka
sudo rm -f /etc/nginx/sites-enabled/default
sudo nginx -t && sudo systemctl reload nginx

echo ""
echo "========================================="
echo "✅ Базовая настройка завершена"
echo ""
echo "Ручные шаги:"
echo "  1. DNS: A-запись sarafanka.su → 72.56.22.16 у регистратора"
echo "  2. На этом сервере:"
echo "     ssh-keygen -t ed25519 -C 'sarafanka-server' -f ~/.ssh/id_ed25519 -N ''"
echo "     cat ~/.ssh/id_ed25519.pub"
echo "     → добавь ключ в GitHub: Settings → SSH and GPG keys"
echo "  3. Проверь: cd /opt/sarafanka && git pull (должно быть без пароля)"
echo "  4. SSL: sudo certbot --nginx -d sarafanka.su"
echo "  5. На локальной машине: добавь публичный ключ локального SSH-ключа"
echo "     в ~/.ssh/authorized_keys на этом сервере (для GitHub Actions деплоя)"
echo "  6. Для зеркала GitVerse:"
echo "     git remote add gitverse git@gitverse.ru:ITcommuna/sarafanka.git"
echo "========================================="
