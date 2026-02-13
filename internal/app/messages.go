package app

import (
	"github.com/e-9/copilot-icq/internal/domain"
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
