package app

import (
	"github.com/e-9/copilot-icq/internal/domain"
	"github.com/e-9/copilot-icq/internal/infra/ptyproxy"
	"github.com/e-9/copilot-icq/internal/infra/watcher"
)

// SessionsLoadedMsg is sent when sessions are discovered from disk.
type SessionsLoadedMsg struct {
	Sessions []domain.Session
	Err      error
}

// EventsLoadedMsg is sent when a session's events are parsed.
type EventsLoadedMsg struct {
	SessionID string
	Messages  []domain.Message
	Err       error
}

// FileChangedMsg wraps a watcher event for Bubble Tea.
type FileChangedMsg struct {
	watcher.EventFileChanged
}

// SessionDirChangedMsg wraps a session directory change.
type SessionDirChangedMsg struct{}

// MessageSentMsg is returned when a message has been dispatched to copilot.
type MessageSentMsg struct {
	SessionID string
	Success   bool
	Output    string
	Err       error
}

// TickMsg is sent periodically to trigger session rescans.
type TickMsg struct{}

// SessionRenamedMsg is sent when a session has been renamed.
type SessionRenamedMsg struct {
	SessionID string
	Err       error
}

// ExportCompleteMsg is sent when a conversation export finishes.
type ExportCompleteMsg struct {
	Path string
	Err  error
}

// PTYOutputMsg is sent when the PTY session produces output.
type PTYOutputMsg struct {
	SessionID string
	Chunk     ptyproxy.OutputChunk
}

// PTYPromptMsg is sent when an approval prompt is detected in PTY output.
type PTYPromptMsg struct {
	SessionID string
	Prompt    *ptyproxy.ApprovalPrompt
}

// PTYClosedMsg is sent when the PTY session process exits.
type PTYClosedMsg struct {
	SessionID string
	Err       error
}

// ApprovalSelectedMsg is sent when the user selects an approval option.
type ApprovalSelectedMsg struct {
	SessionID string
	Shortcut  string // the number key to send (e.g., "1", "2", "3")
}

// ptyStartedMsg is sent when a PTY session has been successfully spawned.
type ptyStartedMsg struct {
	SessionID string
	Session   *ptyproxy.Session
}
