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

## Item constraints
- Names are normalized (trim, lowercase, collapse spaces) and limited to 40 characters.
- Max 200 items per user; identity is per-user by name.

## Architecture pointers
- Models: `models/`
- Repos (GORM only): `internal/repo/`
- Services/logic: `internal/feat/`
- Session cache: `internal/session/`

See `PROMT_INSTRUCTIONS` for behavior rules and messaging constraints.
