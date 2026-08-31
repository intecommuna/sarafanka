#!/bin/sh
cd /opt/sarafanka || exit 1
LOCAL=$(git rev-parse HEAD)
REMOTE=$(git ls-remote origin main | awk '{print $1}')
[ "$LOCAL" = "$REMOTE" ] && exit 0
echo "[$(date)] New commits detected, deploying..."
git pull origin main || exit 1
docker compose build && docker compose up -d
echo "[$(date)] Deploy done."
