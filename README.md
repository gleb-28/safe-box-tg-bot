# 🤖 Safe Box TG Bot

## ✨ Description

Safe Box is a Telegram bot that sends short, human-like nudges during the day based on user items
(e.g., "tea", "look out the window"). Messages are generated via an LLM in different styles
(rofl/cozy/care) and are delivered only within the user's day window, with randomized intervals driven by presets:
- Редко: 2–4 часа.
- Иногда (по умолчанию): 60–120 минут.
- Часто: 40–90 минут.
- Хаос: 30–180 минут.
Access is gated by an activation key.
Users can switch the message style at any time with `/change_mode` (buttons: rofl/cozy/care + close), adjust the frequency with `/change_interval`, and quickly mute/unmute with `/toggle_notifications`.

Active hours are stored as DayStart/DayEnd minutes in 24-hour format; the notification worker runs once on startup
and then periodically to process due users. If `NextNotification` is overdue beyond the max interval, it is recalculated
from "now" without sending. Each successful notification is logged at info level with user ID, item ID, name, and the
sent text for easier ops tracing.

LLM requests go through OpenRouter using the prompt in `data/prompt`; replies are trimmed and unwrapped from
`json`/`text` code fences before sending. If generation fails, the item name plus an emoji (palette in
`internal/core/constants`) is sent as a fallback.

## 🔒 Limits

- Max items per user: 200.
- Item names are normalized (trimmed, lowercased, collapsed spaces) and limited to 40 characters.

## 🧱 Tech stack

- **Go 1.25** – primary language.
- **Telebot v4** – Telegram framework.
- **Looplab FSM** – FSM.
- **GORM + sqlite driver** – persistence layer for chats and forward-mode settings (`gorm.io/gorm`, `gorm.io/driver/sqlite`).
- **SQLite** – lightweight storage for bot data.
- **cleanenv** – loads `.env` file.

## 📦 Requirements

Before running the bot make sure you have installed:
- Go 1.25
- SQLite

Check installed versions:
```bash
go version
sqlite3 --version
````

## ⚙️ Environment variables

Create .env file based on env.example:
```env
TG_BOT_TOKEN=              # REQUIRED - Telegram bot token
LOGGER_BOT_TOKEN=          # OPTIONAL (if used for logging bot)
ADMIN_ID=                  # REQUIRED - Telegram admin user ID
ACTIVATION_KEY=            # REQUIRED - password to use the bot
DB_FILE_NAME=./data/bot.db # REQUIRED - SQLite db file (*.db)
MODEL_API_KEY=             # REQUIRED - OpenRouter Model API key 
MODEL_NAME=openrouter/auto # OPTIONAL - OpenRouter model name
PROMPT_PATH=./data/prompt  # OPTIONAL - LLM prompt file path
IS_DEBUG=false             # OPTIONAL - print logs for debugging
```
## 📁 Project commands
Makefile included.

### Build:
```bash
make build
```
### Run locally:
```bash
make run
```
### Tidy dependencies:
```bash
make tidy
```

## 🚀 VPS Deployment

This guide shows how to deploy the bot on a fresh Ubuntu VPS using Docker.
All deployment assets (compose file, helper script, Dockerfile, and env templates) live under `deploy/`.

1. Create prod.env with and other constants:
```env
DB_FILE_NAME=/app/data/bot.db
```

2. Create `deploy/prod.env` (if you need to override the defaults above) and `deploy/deploy.env`, then run the deploy helper from the repo root:
```
sudo chmod +x deploy/deploy.sh && make deploy
```
