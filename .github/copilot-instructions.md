# Copilot ICQ — Copilot Instructions

## Project Overview

Go + Bubble Tea TUI messenger for managing multiple GitHub Copilot CLI sessions.
Repository: https://github.com/e-9/copilot-icq

## Critical: PTY Resource Management

macOS has a hard limit of ~511 pseudo-terminal (PTY) devices (`kern.tty.ptmx_max`).
**Every bash tool invocation allocates a PTY that may not be released until the
copilot process exits.** Over long sessions this exhausts all PTY slots, locking
out ALL terminal tools system-wide.

### Rules for bash tool usage

1. **Prefer non-bash tools whenever possible:**
   - Use `grep` / `glob` instead of `grep`, `find`, `ls` in bash
   - Use `view` / `edit` / `create` instead of `cat`, `sed`, `echo >>`
   - Use the `task` tool for complex multi-step operations

2. **Chain commands with `&&` or `;`** to minimize the number of bash invocations:
   ```bash
   # GOOD — one PTY
   go build ./... && go test ./... 2>&1 | tail -20

   # BAD — three PTYs
   go build ./...
   go test ./...
   tail -20 output.log
   ```

3. **Periodically check for leaked sessions:**
   - Run `list_bash` to see active shell sessions
   - Use `stop_bash` to clean up any lingering sessions
   - Run `go run ./cmd/copilot-icq doctor` to check PTY health

4. **Monitor PTY usage during long sessions:**
   - If PTY usage exceeds 70%, proactively clean up
   - If bash starts failing with `pty_posix_spawn`, orphaned shells need killing
   - Run `go run ./cmd/copilot-icq cleanup` to reclaim PTY slots

5. **Never leave interactive/async bash sessions running** unless explicitly needed
   (e.g., dev servers). Always `stop_bash` when done.

## Architecture

- **Language:** Go 1.26+ with Bubble Tea (Elm Architecture)
- **Pattern:** Layered — `cmd/` → `internal/app` → `internal/ui` → `internal/domain` → `internal/infra`
- **Testing:** `go test ./...` — all tests must pass before committing
- **Build:** `go build ./...` — must compile cleanly

### Key packages

| Package | Purpose |
|---------|---------|
| `cmd/copilot-icq` | Main entry point, subcommands: `doctor`, `cleanup`, `install-hooks` |
| `cmd/copilot-icq-hook` | Companion binary for Copilot CLI hook events |
| `internal/app` | Root Elm Architecture model (model.go, update.go, view.go, commands.go, messages.go) |
| `internal/domain` | Core types: Session, Message, Event, ToolCall |
| `internal/infra/ptyproxy` | PTY session management + ANSI parser + approval prompt detection |
| `internal/infra/runner` | Fire-and-forget message dispatch via `copilot -p --resume` |
| `internal/infra/sessionrepo` | Session discovery from `~/.copilot/session-state/` |
| `internal/infra/watcher` | fsnotify file watching for real-time updates |
| `internal/infra/hookserver` | Unix socket server for hook event ingestion |
| `internal/infra/notifier` | Notification router: TUI, OS, push (ntfy.sh) |
| `internal/config` | App config (`~/.copilot-icq/config.yaml`) + doctor + cleanup |
| `internal/ui/sidebar` | Session list with smart sorting, filtering, unread badges |
| `internal/ui/chat` | Chat viewport with glamour markdown, collapsible tool calls |
| `internal/ui/input` | Text input with rename/sending modes |
| `internal/ui/theme` | Colors, dimensions, shared styles |

### lipgloss gotcha

`Height(N)` does NOT clip content — use `MaxHeight(N)` to constrain rendered height.

### bubbles/list filter gotcha

`SetItems()` while filter is active clears the visible list. Check `FilterState()`
before calling `SetItems()` and use `ClearFilterAndSetItems()` for explicit selection.

## Coding Standards

- Minimal changes — change as few lines as possible
- No comments except where clarification is genuinely needed
- Run `go build ./...` and `go test ./...` before every commit
- Commit messages: `"Phase N: description"` for features, `"fix: description"` for bugs, `"feat: description"` for enhancements

## Session Data (read-only)

- `~/.copilot/session-state/<uuid>/workspace.yaml` — session metadata
- `~/.copilot/session-state/<uuid>/events.jsonl` — append-only event log
- **Never write** to `~/.copilot/session-state/` — it's owned by the CLI
- Config lives in `~/.copilot-icq/config.yaml`
