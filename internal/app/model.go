package app

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/e-9/copilot-icq/internal/domain"
	"github.com/e-9/copilot-icq/internal/infra/sessionrepo"
	"github.com/e-9/copilot-icq/internal/infra/watcher"
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
	watcher  *watcher.Watcher
	sessions []domain.Session
	selected *domain.Session
	unread   map[string]int // sessionID → unread count
	lastSeen map[string]time.Time // sessionID → last update time
	err      error
}

// NewModel creates the initial application model.
func NewModel(repo *sessionrepo.Repo, w *watcher.Watcher) Model {
	return Model{
		repo:     repo,
		watcher:  w,
		sidebar:  sidebar.New(nil, theme.SidebarWidth, 20),
		chat:     chat.New(80, 20),
		unread:   make(map[string]int),
		lastSeen: make(map[string]time.Time),
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		loadSessions(m.repo),
		watchFiles(m.watcher),
	)
}
