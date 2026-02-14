package app

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/e-9/copilot-icq/internal/config"
	"github.com/e-9/copilot-icq/internal/domain"
	"github.com/e-9/copilot-icq/internal/infra/runner"
	"github.com/e-9/copilot-icq/internal/infra/sessionrepo"
	"github.com/e-9/copilot-icq/internal/infra/watcher"
	"github.com/e-9/copilot-icq/internal/ui/chat"
	"github.com/e-9/copilot-icq/internal/ui/input"
	"github.com/e-9/copilot-icq/internal/ui/sidebar"
	"github.com/e-9/copilot-icq/internal/ui/theme"
)

// Focus tracks which panel has keyboard focus.
type Focus int

const (
	FocusSidebar Focus = iota
	FocusChat
	FocusInput
)

// Model is the root application state following the Elm Architecture.
type Model struct {
	width    int
	height   int
	ready    bool
	focus    Focus
	sidebar  sidebar.Model
	chat     chat.Model
	input    input.Model
	repo     *sessionrepo.Repo
	watcher  *watcher.Watcher
	runner   *runner.Runner
	sessions []domain.Session
	selected *domain.Session
	unread   map[string]int // sessionID → unread count
	lastSeen map[string]time.Time // sessionID → last update time
	err      error
	showHelp bool   // keyboard shortcuts overlay
	renaming bool   // inline session rename mode
	cfg      *config.AppConfig // user configuration
}

// NewModel creates the initial application model.
func NewModel(repo *sessionrepo.Repo, w *watcher.Watcher, r *runner.Runner, cfg *config.AppConfig) Model {
	return Model{
		repo:     repo,
		watcher:  w,
		runner:   r,
		sidebar:  sidebar.New(nil, theme.SidebarWidth, 20),
		chat:     chat.New(80, 20),
		input:    input.New(80),
		unread:   make(map[string]int),
		lastSeen: make(map[string]time.Time),
		cfg:      cfg,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		loadSessions(m.repo),
		watchFiles(m.watcher),
		tickEvery(5*time.Second),
	)
}
