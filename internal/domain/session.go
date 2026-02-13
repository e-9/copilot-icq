package domain

import (
	"path/filepath"
	"time"
)

// Session represents a Copilot CLI session with its metadata.
type Session struct {
	ID           string    `yaml:"id"`
	CWD          string    `yaml:"cwd"`
	Summary      string    `yaml:"summary"`
	SummaryCount int       `yaml:"summary_count"`
	CreatedAt    time.Time `yaml:"created_at"`
	UpdatedAt    time.Time `yaml:"updated_at"`
}

// DisplayName returns a human-readable name for the session.
func (s Session) DisplayName() string {
	if s.Summary != "" {
		return s.Summary
	}
	// Fall back to last directory component of CWD
	if s.CWD != "" {
		return filepath.Base(s.CWD)
	}
	return s.ShortID()
}

// ShortID returns the first 8 characters of the session ID.
func (s Session) ShortID() string {
	if len(s.ID) >= 8 {
		return s.ID[:8]
	}
	return s.ID
}
