#!/bin/bash

# ===== Load deploy config =====
set -e
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

if [ ! -f "$SCRIPT_DIR/deploy.env" ]; then
    echo "❌ deploy.env not found!"
    exit 1
fi
export $(grep -v '^#' "$SCRIPT_DIR/deploy.env" | xargs)

# ==================

echo "🚀 Deploying to VPS..."

rsync -avz --delete \
  --exclude '.git' \
  --exclude 'downloads' \
  --exclude '.idea/' \
  --exclude 'output' \
  --exclude 'tmp' \
  --exclude '.env' \
  --exclude '.vscode/' \
  --exclude '.ai/' \
  --exclude 'data/bot.db' \
  "$PROJECT_ROOT/" "$VPS_USER@$VPS_HOST:$VPS_PATH"

echo "📦 Files synced."

# ---------- Remote setup + run ----------
ssh $VPS_USER@$VPS_HOST <<'EOF'
set -e

echo "🔧 Installing dependencies..."

apt update -y && apt upgrade -y
apt install -y curl make

# install docker if missing
if ! command -v docker &> /dev/null; then
  curl -fsSL https://get.docker.com | sh
fi

# Install Docker Compose if missing
if ! command -v docker-compose &> /dev/null; then
    echo "🐳 Installing Docker Compose..."
    LATEST_COMPOSE=$(curl -s https://api.github.com/repos/docker/compose/releases/latest | grep -Po '"tag_name": "\K.*?(?=")')
    curl -L "https://github.com/docker/compose/releases/download/${LATEST_COMPOSE}/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
    chmod +x /usr/local/bin/docker-compose
fi

echo "📁 Preparing folders..."
mkdir -p /home/safe-box-bot/data

echo "🐳 Building & starting containers..."
cd /home/safe-box-bot/deploy

docker-compose down || true
docker-compose build
docker-compose up -d

echo "✅ VPS deployment finished!"
EOF

echo "🎉 Deploy completed successfully!"
