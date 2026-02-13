package app

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/e-9/copilot-icq/internal/domain"
	"github.com/e-9/copilot-icq/internal/infra/sessionrepo"
	"github.com/e-9/copilot-icq/internal/ui/chat"
	"github.com/e-9/copilot-icq/internal/ui/sidebar"
	"github.com/e-9/copilot-icq/internal/ui/theme"
)

// Focus tracks which panel has keyboard focus.
type Focus int

const (
	FocusSidebar Focus = iota
	FocusChat
)

// Model is the root application state following the Elm Architecture.
type Model struct {
	width    int
	height   int
	ready    bool
	focus    Focus
	sidebar  sidebar.Model
	chat     chat.Model
	repo     *sessionrepo.Repo
	sessions []domain.Session
	selected *domain.Session
	err      error
}

// NewModel creates the initial application model.
func NewModel(repo *sessionrepo.Repo) Model {
	return Model{
		repo:    repo,
		sidebar: sidebar.New(nil, theme.SidebarWidth, 20),
		chat:    chat.New(80, 20),
	}
}

func (m Model) Init() tea.Cmd {
	return loadSessions(m.repo)
}
