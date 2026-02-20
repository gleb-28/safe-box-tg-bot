# TODO

## How we work with this list
- Keep items small and outcome-focused; define what “done” means.
- Priority:
    - **P1** — critical
    - **P2** — soon
    - **P3** — nice-to-have
- Use checkboxes:
    - `[ ]` — todo / in progress
    - `[x]` — done

## Backlog

### P1 — critical


### P2 — soon


### P3 — nice-to-have

- [ ] **Add CI to run tests and formatting + pre-commit hooks**
  _Notes_: Run `go test ./...` and `gofmt` on pushes/PRs (GitHub Actions)

## Done

- [x] **Add status to open box**
  _Notes_: Режим: cozy, Вещей: 5, Окно: 12:00–22:00
- [x] **Reminders feature**
  _Notes_: Allow users to set a reminder time/window for an item so the bot nudges them when the reminder is due.
- [x] **Handle no user in db**
    _Notes_: `func (r *UserRepo) TryGet(id int64) (*models.User, bool, error)`