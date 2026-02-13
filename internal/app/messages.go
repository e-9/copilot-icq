package app

import (
	"github.com/e-9/copilot-icq/internal/domain"
)

// SessionsLoadedMsg is sent when sessions are discovered from disk.
type SessionsLoadedMsg struct {
	Sessions []domain.Session
	Err      error
}

// SessionSelectedMsg is sent when a session is selected in the sidebar.
type SessionSelectedMsg struct {
	Session domain.Session
}
