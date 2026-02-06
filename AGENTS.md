# AGENTS.md

## Repo overview
- Go 1.25 Telegram bot; entrypoints live in cmd/.
- Business logic is under internal/feat; data access in internal/repo; shared types in models/.
- Notifications run in a background worker under internal/feat/notify using NextNotification.
- If NextNotification is overdue beyond the max interval (or zero), reschedule from now without sending.
- Item box close message ID is stored on the user to delete stale "Шкатулка закрыта" after restarts.

## Coding guidelines
- Run gofmt on any modified Go files.
- Prefer the standard library; avoid new deps unless necessary.
- Telegram user IDs are int64; keep method signatures consistent.
- Use the existing logger interface; avoid fmt.Println/print.

## Data and storage
- Use GORM for persistence and keep DB queries in internal/repo.
- Add gorm tags for new models and keep defaults/checks explicit.
- Item identity is per-user and keyed by normalized name (max 40 chars, 200 items/user).

## Testing
- Use package-local *_test.go files; prefer table tests when practical.
- Run `go test ./...` (or targeted packages) for behavior changes.

## Docs and ops
- Update `README.md` and `PROMT_INSTRUCTIONS` when behavior changes.
- Environment variables live in env.example; deploy assets are under deploy/.
