# DEVELOPMENT

This file covers local development for Safe Box TG Bot.

## Prerequisites
- Go 1.25
- SQLite (sqlite3 CLI is helpful)

## Setup
1) Copy `.env` from `.env.example`:
   ```bash
   cp .env.example .env
   ```
2) Fill in required values in `.env`:
   - `TG_BOT_TOKEN`
   - `ADMIN_ID`
   - `ACTIVATION_KEY`
   - `DB_FILE_NAME` (default `./data/bot.db`)
   - `MODEL_API_KEY`
   - `MODEL_NAME` (optional, default `openrouter/auto`)
3) Ensure the DB directory exists:
   ```bash
   mkdir -p data
   ```

## Run locally
```bash
make run
```
Or:
```bash
go run cmd/bot/main.go
```

## Tests
```bash
make test
```
Or:
```bash
go test ./... -v
```

## Formatting
Run gofmt on modified Go files:
```bash
gofmt -w <file.go>
```

## Common tasks
- Tidy modules: `make tidy`
- Docker deploy (VPS): `make deploy` (see `deploy/`)

## Debugging
Set `IS_DEBUG=true` in `.env` to enable debug logs.

## LLM testing
- Use `notify.NewLLMPreviewService(userService, itemsService, messageGenerator, bot, logger)` to preview texts for a specific user.
- `SendPreviews(ctx, userID)` generates messages for all user items in the current mode/time-of-day and sends them to the user as `<item>: <message>`.
- Handy for iterating on the prompt without pushing real production notifications.
- Admin-only command: `/preview_llm` triggers the preview service for the admin chat.

## Notifications
- The notification worker starts automatically with the bot, runs once immediately, and uses `NextNotification` in UTC.
- Messages are sent only within DayStart/DayEnd in the user's timezone.
- DayStart/DayEnd are minutes in 24-hour format; DayStart != DayEnd is enforced by validation.
- Notifications use LLM generation (prompt from PROMPT_PATH, default `data/prompt`) via a message generator (prompt builder + LLM client); if it fails, the item name plus an emoji is sent as a fallback.
- On success the worker logs info with userID, itemID, item name, and the sent text to aid ops investigations.
- If `NextNotification` is overdue beyond the max interval (or zero), recalculate it from now without sending.
- Randomized interval is 40–150 minutes (40min–2.5 hours), stored/treated in minutes across the system.

## Item box UI
- On close, the bot sends "Шкатулка закрыта" with the main menu keyboard.
- The message ID is stored on the user and deleted on next open (so restarts can clean up the old message).

## Item constraints
- Names are normalized (trim, lowercase, collapse spaces) and limited to 40 characters.
- Max 200 items per user; identity is per-user by name.

## Architecture pointers
- Models: `models/`
- Repos (GORM only): `internal/repo/`
- Services/logic: `internal/feat/`
- Session cache: `internal/session/`
- Telegram handlers: `internal/handler/{commands,keyboard,message}`, middleware in `internal/middleware/`

See `PROMT_INSTRUCTIONS` for behavior rules and messaging constraints.
