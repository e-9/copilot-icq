package app

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/e-9/copilot-icq/internal/infra/sessionrepo"
)

// loadSessions returns a Cmd that discovers sessions from disk.
func loadSessions(repo *sessionrepo.Repo) tea.Cmd {
	return func() tea.Msg {
		sessions, err := repo.List()
		return SessionsLoadedMsg{Sessions: sessions, Err: err}
	}
}
