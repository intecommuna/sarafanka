#!/usr/bin/env bash
set -e

GITHUB_USER="intecommuna"
REPO="git@github.com:intecommuna/sarafanka.git"
PROJECT_DIR="/opt/sarafanka"

apt update -qq
apt install -y git curl ufw fail2ban certbot docker.io docker-compose-v2

systemctl enable --now docker

ufw allow 22/tcp
ufw allow 80/tcp
ufw allow 443/tcp
ufw allow 443/udp
ufw --force enable

mkdir -p /var/www/certbot

if [ ! -d "${PROJECT_DIR}/.git" ]; then
    git clone "${REPO}" "${PROJECT_DIR}"
fi

cd "${PROJECT_DIR}"

git remote set-url origin "${REPO}" || true

echo ">>> Выпуск SSL-сертификата (если доступен порт 80)"
certbot certonly --standalone --non-interactive --agree-tos \
  -m intecommuna@gmail.com \
  -d sarafanka.su -d www.sarafanka.su || echo "WARN: certificate generation failed; check DNS / ports"

echo ">>> Запуск Docker-сервисов"
docker compose up -d --build

echo ""
echo "========================================="
echo "✅ Docker-сервер готов"
echo ""
echo "Ручные шаги:"
echo "  1. Серверный SSH-ключ: добавить в GitHub (Settings → SSH and GPG keys)"
echo "  2. GitVerse зеркало: git remote add gitverse git@gitverse.ru:ITcommuna/sarafanka.git"
echo "  3. Проверка логов: docker compose logs -f"
echo "========================================="
