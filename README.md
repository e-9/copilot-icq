# 🟢 Copilot ICQ

A terminal user interface (TUI) for managing multiple [GitHub Copilot CLI](https://docs.github.com/en/copilot/concepts/agents/copilot-cli/about-copilot-cli) sessions from a single window. Think of it as Slack or Teams — but for your AI coding agents instead of people.

> Inspired by [ICQ](https://en.wikipedia.org/wiki/ICQ), the classic instant messenger. Instead of chatting with friends, you're chatting with your Copilot sessions.

![Go](https://img.shields.io/badge/Go-1.25+-00ADD8?logo=go&logoColor=white)
![Platform](https://img.shields.io/badge/Platform-macOS%20%7C%20Linux%20%7C%20Windows-blue)
![License](https://img.shields.io/badge/License-MIT-green)

---

## Why?

If you use GitHub Copilot CLI heavily, you've probably run into this:

- 🪟 **Too many terminal tabs** — each Copilot session lives in its own tab
- 🔄 **Constant tab-switching** — you're bouncing between sessions to check progress
- 🔕 **No notifications** — you don't know when a session needs your attention
- 🏷️ **No labels** — sessions are identified by UUIDs, not meaningful names

**Copilot ICQ** solves all of this by providing a unified interface where you can:

- **See all sessions** in a sidebar with smart sorting (active → notifications → idle)
- **Read conversations** with full markdown rendering, tool call previews, and diffs
- **Send messages** to any session directly from the TUI
- **Get real-time updates** via hooks and file system watching
- **Approve or deny tools** via a security policy before they execute
- **Open sessions** in a native terminal for interactive work
- **Rename sessions** with meaningful names instead of UUIDs
- **Export conversations** to markdown for documentation

---

## Quick Start

### Prerequisites

- **Go 1.25+** — [Install Go](https://go.dev/dl/)
- **GitHub Copilot CLI** — [Install Copilot CLI](https://docs.github.com/en/copilot/how-tos/copilot-cli/using-copilot-cli)
- Active Copilot CLI sessions (at least one session in `~/.copilot/session-state/`)

### Install

```bash
# Clone the repository
git clone https://github.com/e-9/copilot-icq.git
cd copilot-icq

# Build both binaries
make all

# Install the hook companion binary to your PATH
cp bin/copilot-icq-hook /usr/local/bin/

# Verify your setup
./bin/copilot-icq doctor

# Run the TUI
./bin/copilot-icq
```

### Install Hooks (Optional but Recommended)

Hooks enable **instant** real-time notifications when Copilot sessions fire events. Without hooks, the TUI still works — it falls back to file system watching (fsnotify), which is slightly delayed but fully functional.

To install hooks in a project you use with Copilot CLI:

```bash
# Make sure copilot-icq-hook is in your PATH first
which copilot-icq-hook

# Install hooks in a project
./bin/copilot-icq install-hooks /path/to/your/project
```

This creates `.github/hooks/copilot-icq.json` in the project with handlers for 6 lifecycle events.

---

## Platform Setup

### macOS

```bash
make all
cp bin/copilot-icq-hook /usr/local/bin/
./bin/copilot-icq
```

The `a` keybinding opens sessions in **Terminal.app**. The hook companion binary communicates via a Unix socket at `~/.copilot/copilot-icq.sock`.

### Linux

```bash
make all
cp bin/copilot-icq-hook /usr/local/bin/
# Or: cp bin/copilot-icq-hook ~/go/bin/  (if ~/go/bin is in PATH)
./bin/copilot-icq
```

OS notifications use `notify-send` (install `libnotify-bin` on Debian/Ubuntu if needed). The `a` keybinding to open sessions in a terminal window is **macOS-only** currently (uses AppleScript). Linux support for this feature is planned.

### Windows

```powershell
go build -o bin\copilot-icq.exe .\cmd\copilot-icq
go build -o bin\copilot-icq-hook.exe .\cmd\copilot-icq-hook

# Add bin\ to your PATH, or copy to a directory already in PATH
copy bin\copilot-icq-hook.exe C:\Users\%USERNAME%\go\bin\

.\bin\copilot-icq.exe
```

OS notifications use PowerShell toast notifications. The hook companion binary communicates via a Unix socket (requires Windows 10 1803+ with AF_UNIX support).

---

## The Interface

```
┌──────────────────────────────────────────────────────────────────────────────┐
│ 🟢 Copilot ICQ  3 sessions  🔒 scoped              ? help  e export  q quit │
├─ Sessions ──────────────╥─ Chat · my-project (d8f6) ────────────────────────┤
│                         ║                                                    │
│ Active ╶╶╶╶╶╶╶╶╶╶╶╶╶╶╶ ║  Copilot                               14:23     │
│ ◉ my-project            ║  I'll set up the map component. Let me             │
│   d8f6 · …/my-project   ║  check the existing files...                      │
│                         ║                                                    │
│ Notifications ╶╶╶╶╶╶╶╶ ║  ⚙ bash  ✓                                        │
│ ○ 🔔 api-server  (2)   ║    ls src/components/                              │
│   a1b2 · …/api-server   ║                                                    │
│                         ║  ⚙ edit  ✓                                        │
│ Idle ╶╶╶╶╶╶╶╶╶╶╶╶╶╶╶╶╶ ║    📝 src/components/Map.tsx                      │
│ ○ old-experiment         ║                                                    │
│   c3d4 · …/experiment   ║  You                                     14:25    │
│                         ║  Can you add error handling?                       │
│                         ║                                                    │
├─────────────────────────╠─ Input ───────────────────────────────────────────┤
│                         ║ ❯ Type a message...                                │
├─────────────────────────╨────────────────────────────────────────────────────┤
│ [chat] · 💬 my-project (d8f6)                                                │
└──────────────────────────────────────────────────────────────────────────────┘
```

### Panels

| Panel | Purpose |
|-------|---------|
| **Sessions** (left) | Lists all Copilot CLI sessions. Smart-sorted: active session first, then sessions with unread messages, then idle. Shows activity icons (◉ active / ○ idle), status indicators (⏳ waiting / 🔔 has response), and unread badges. |
| **Chat** (top right) | Displays the full conversation for the selected session. Renders markdown via [glamour](https://github.com/charmbracelet/glamour), shows tool calls with expand/collapse, file diffs with color coding, `ask_user` prompts, and approval status. |
| **Input** (bottom right) | Type and send messages to the selected session. Shows sending state per-session — you can send to multiple sessions concurrently. |
| **Status Bar** (bottom) | Shows current focus panel, selected session info, and transient status messages. |

### Sidebar Sections

| Section | Description |
|---------|-------------|
| **Active** | The session you're currently viewing. Marked with ◉. |
| **Notifications** | Sessions with unread messages or pending activity. Shows 🔔 (response ready) or ⏳ (waiting for Copilot). Unread count in orange badge. |
| **Idle** | Sessions with no recent activity. |

---

## Keyboard Shortcuts

Press `?` at any time to see the full shortcut overlay.

### Navigation

| Key | Action |
|-----|--------|
| `Tab` | Switch panel forward (sidebar → chat → input) |
| `Shift+Tab` | Switch panel backward (input → chat → sidebar) |
| `Enter` | Open session (sidebar) / Send message (input) |
| `Esc` | Go back (input → chat, cancel rename) |
| `↑` `↓` | Navigate sessions (sidebar) / Scroll chat |
| `/` | Filter sessions by name (sidebar) |

### Actions

| Key | Context | Action |
|-----|---------|--------|
| `a` | Chat | Open session in Terminal.app (macOS only) |
| `t` | Chat | Toggle all tool call details (expand/collapse) |
| `r` | Any | Refresh session list |
| `R` | Sidebar | Rename selected session |
| `e` | Any | Export conversation to markdown file |
| `?` | Any | Toggle keyboard shortcuts overlay |
| `q` | Any (except input) | Quit |
| `Ctrl+C` | Any | Force quit |

---

## Commands

### `copilot-icq`

Launches the TUI. Discovers sessions from `~/.copilot/session-state/` and starts watching for changes.

```bash
# Run with default config
./bin/copilot-icq

# Run with custom config
./bin/copilot-icq --config /path/to/config.yaml
```

### `copilot-icq install-hooks [directory]`

Installs Copilot CLI hook configuration in the specified project directory (defaults to current directory). Creates `.github/hooks/copilot-icq.json` with handlers for 6 lifecycle events.

```bash
# Install hooks in a project
./bin/copilot-icq install-hooks /path/to/project

# Install hooks in current directory
./bin/copilot-icq install-hooks
```

**Hook events installed:**

| Event | When | Purpose |
|-------|------|---------|
| `sessionStart` | Session begins | TUI updates sidebar, shows new session |
| `sessionEnd` | Session ends | TUI updates session status |
| `preToolUse` | Before tool executes | Deny policy enforcement, TUI shows pending tool |
| `postToolUse` | After tool executes | TUI reloads conversation, clears pending state |
| `userPromptSubmitted` | User sends prompt | TUI tracks activity |
| `errorOccurred` | Error in session | TUI shows error notification |

### `copilot-icq doctor`

Runs system diagnostics — checks PTY device usage, detects orphaned shell processes, and verifies the hook server socket.

```bash
./bin/copilot-icq doctor
```

---

## Configuration

Config file location: `~/.copilot-icq/config.yaml`

```yaml
# Security mode for TUI-sent messages
# "scoped"    — only allow specific tools (default, recommended)
# "full-auto" — allow all tools (use with caution)
security_mode: scoped

# Tools allowed in scoped mode (used with --allow-tool flags)
allowed_tools:
  - view
  - glob
  - grep
  - bash

# Tools to block via preToolUse hook (applies to ALL sessions)
denied_tools: []
  # - rm
  # - sudo

# Argument patterns to block (substring match against tool args)
denied_patterns: []
  # - "rm -rf /"
  # - "DROP TABLE"

# Directory for conversation exports (e key)
export_dir: "."

# Notification settings
notifications:
  os: false     # OS desktop notifications (macOS/Linux/Windows)
  push: false   # ntfy.sh push notifications
  topic: ""     # ntfy.sh topic name
```

### Security Modes

| Mode | Behavior | When to use |
|------|----------|-------------|
| **scoped** (default) | Messages sent from TUI use `--allow-tool` flags for specific tools only | Day-to-day use. Copilot can read and edit, but you control what's allowed. |
| **full-auto** | Messages sent from TUI use `--allow-all-tools` | When you trust the session fully and want zero interruptions. |

### Deny Policy

The deny policy is enforced via the `preToolUse` hook and applies to **all sessions** — both TUI-sent and sessions running in other terminals. When a tool matches `denied_tools` or `denied_patterns`, Copilot CLI receives a deny decision and skips the tool.

---

## How It Works

### Architecture

```
┌─────────────┐     ┌──────────────────┐     ┌────────────────────┐
│ Copilot CLI  │────→│ events.jsonl     │────→│ Copilot ICQ TUI    │
│ (sessions)   │     │ (append-only log)│     │                    │
│              │     └──────────────────┘     │  ┌──────┐ ┌──────┐ │
│              │            ↑ fsnotify        │  │Sidebar│ │ Chat │ │
│              │     ┌──────────────────┐     │  └──────┘ └──────┘ │
│              │────→│ copilot-icq-hook │────→│  ┌──────┐          │
│              │     │ (companion bin)  │  ⚡  │  │Input │          │
│              │     └──────────────────┘ sock │  └──────┘          │
└─────────────┘                               └────────────────────┘
```

1. **Session discovery** — Scans `~/.copilot/session-state/` for session directories with `workspace.yaml`
2. **Conversation reading** — Parses `events.jsonl` (append-only event log) into messages, tool calls, and metadata
3. **Real-time updates** — Two complementary channels:
   - **fsnotify** watches `events.jsonl` for file changes (reliable, slightly delayed)
   - **Hooks** fire immediately when Copilot CLI events occur (instant, requires setup)
4. **Sending messages** — Spawns `copilot -p "message" --resume <session-id>` subprocesses with scoped tool permissions
5. **Security policy** — `preToolUse` hook checks deny lists before tools execute

### Data Access

Copilot ICQ **reads** from `~/.copilot/session-state/` (owned by Copilot CLI):

- `workspace.yaml` — session metadata (ID, CWD, summary, timestamps)
- `events.jsonl` — conversation events (prompts, responses, tool calls)

The only writes to that directory are:
- Updating `workspace.yaml` summary field when renaming sessions via the TUI

Other write operations:
- Spawning `copilot -p --resume` subprocesses (which Copilot CLI manages)
- Writing to `~/.copilot-icq/config.yaml` for user configuration
- Exporting conversations to markdown files (in `export_dir`)

---

## Chat Panel Features

### Markdown Rendering

Assistant messages are rendered with full markdown support via [glamour](https://github.com/charmbracelet/glamour) — code blocks with syntax highlighting, headers, lists, bold/italic, and more.

### Tool Call Display

Tool calls show with expand/collapse (press `t`):

```
⚙ bash  ✓                          ← collapsed (default)
  ls src/components/

⚙ edit  ✓                          ← shows file path and diff preview
  📝 src/components/Map.tsx
  + import { MapView } from './MapView'
  - import { OldMap } from './OldMap'
```

### Ask User Prompts

When Copilot asks a question (`ask_user` tool), the TUI renders it with choices:

```
❓ What testing framework should I use?
  1. Jest (Recommended)
  2. Vitest
  3. Mocha
→ Jest
```

### Approval Status

For tools pending approval in external sessions:
```
⚡ Waiting for approval in terminal
```

---

## Build Targets

```bash
make build       # Build TUI binary → bin/copilot-icq
make build-hook  # Build hook binary → bin/copilot-icq-hook
make all         # Build both
make run         # Build and run TUI
make test        # Run all tests (verbose)
make lint        # Run golangci-lint
make fmt         # Format code with gofmt
make tidy        # Tidy Go dependencies
make clean       # Remove bin/ directory
```

---

## Project Structure

```
copilot-icq/
├── cmd/
│   ├── copilot-icq/          # Main TUI entry point + subcommands
│   └── copilot-icq-hook/     # Hook companion binary
├── internal/
│   ├── app/                   # Root Bubble Tea model (Elm Architecture)
│   │   ├── model.go           # Application state
│   │   ├── update.go          # Message handling
│   │   ├── view.go            # UI rendering
│   │   ├── commands.go        # Async commands (load, send, watch)
│   │   └── messages.go        # Message types
│   ├── domain/                # Core types: Session, Message, ToolCall
│   ├── config/                # App config + doctor diagnostics
│   ├── infra/
│   │   ├── eventparser/       # events.jsonl streaming parser
│   │   ├── hookserver/        # Unix socket server for hook events
│   │   ├── notifier/          # Notification router (TUI, OS, push)
│   │   ├── runner/            # copilot subprocess management
│   │   ├── sessionrepo/       # Session discovery from disk
│   │   └── watcher/           # fsnotify file watching
│   └── ui/
│       ├── chat/              # Chat viewport with markdown rendering
│       ├── input/             # Text input with send/rename modes
│       ├── sidebar/           # Session list with smart sorting
│       └── theme/             # Colors, styles, titled borders
├── Makefile
├── go.mod
└── go.sum
```

---

## Status

🟡 **Early Alpha** — Copilot ICQ is functional and actively developed. The core features (session viewing, messaging, hooks, deny policy) work well. Some features are platform-specific (Terminal.app handoff is macOS-only). Expect rough edges and breaking changes.

**What works today:**
- ✅ Multi-session monitoring with real-time updates
- ✅ Send messages to sessions from the TUI
- ✅ Deny policy via `preToolUse` hooks
- ✅ Conversation export, session renaming, markdown rendering
- ✅ macOS, Linux (Windows untested but should work)

**What's coming:**
- ⬜ In-TUI tool approval (blocked by Copilot CLI not supporting `"allow"` in `preToolUse` hooks yet)
- ⬜ Multi-agent support (Claude Code, Gemini CLI)
- ⬜ `brew install` / `go install` distribution
- ⬜ Screenshot/GIF demo with [vhs](https://github.com/charmbracelet/vhs)

---

## Built With

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) — TUI framework (Elm Architecture)
- [Lip Gloss](https://github.com/charmbracelet/lipgloss) — Terminal styling and layout
- [Glamour](https://github.com/charmbracelet/glamour) — Markdown rendering
- [Bubbles](https://github.com/charmbracelet/bubbles) — TUI components (viewport, list, text input)
- [fsnotify](https://github.com/fsnotify/fsnotify) — Cross-platform file system notifications

---

## License

MIT — see [LICENSE](LICENSE).

---

## Advanced

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `COPILOT_ICQ_SOCKET` | `~/.copilot/copilot-icq.sock` | Override the Unix socket path for hook communication |
| `COPILOT_ICQ_NTFY_TOPIC` | *(none)* | Set to enable push notifications via [ntfy.sh](https://ntfy.sh) |

### Troubleshooting

- **TUI shows no sessions** — Make sure you have at least one Copilot CLI session. Run `ls ~/.copilot/session-state/` to check.
- **Hooks not firing** — Verify `copilot-icq-hook` is in your PATH (`which copilot-icq-hook`) and that you ran `install-hooks` in the project directory.
- **`a` key doesn't work** — This feature currently only works on macOS (uses Terminal.app via AppleScript).
- **PTY exhaustion** — Run `./bin/copilot-icq doctor` to check PTY device usage. Copilot CLI has a known PTY fd leak bug on macOS.

---

## Contributing

```bash
# Clone and build
git clone https://github.com/e-9/copilot-icq.git
cd copilot-icq
make all

# Run tests (required before submitting PRs)
make test

# Optional: run linter (requires golangci-lint)
make lint

# Format code
make fmt
```
