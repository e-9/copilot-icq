package app

import (
	"github.com/e-9/copilot-icq/internal/domain"
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
