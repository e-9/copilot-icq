package app

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "r":
			cmds = append(cmds, loadSessions(m.repo))
		case "tab":
			if m.focus == FocusSidebar {
				m.focus = FocusChat
			} else {
				m.focus = FocusSidebar
			}
			return m, nil
		case "enter":
			if m.focus == FocusSidebar {
				if s := m.sidebar.SelectedSession(); s != nil {
					m.selected = s
					m.focus = FocusChat
					m.unread[s.ID] = 0
					m.sidebar.SetUnread(m.unread)
					// Watch this session for live updates
					m.watcher.WatchSession(s.ID)
					cmds = append(cmds, loadEvents(m.repo.BasePath(), *s))
				}
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true
		sidebarHeight := m.height - 2
		m.sidebar.SetSize(m.sidebar.Width, sidebarHeight)
		chatWidth := m.width - m.sidebar.Width - 4
		m.chat.SetSize(chatWidth, sidebarHeight)

	case SessionsLoadedMsg:
		if msg.Err != nil {
			m.err = msg.Err
			return m, nil
		}
		m.sessions = msg.Sessions
		m.sidebar.SetItems(msg.Sessions)
		// Watch all sessions for updates
		for _, s := range msg.Sessions {
			m.watcher.WatchSession(s.ID)
		}

	case EventsLoadedMsg:
		if msg.Err != nil {
			m.err = msg.Err
			return m, nil
		}
		if m.selected != nil && m.selected.ID == msg.SessionID {
			m.chat.SetMessages(msg.Messages)
		}

	case FileChangedMsg:
		m.lastSeen[msg.SessionID] = time.Now()
		m.sidebar.SetLastSeen(m.lastSeen)
		if m.selected != nil && m.selected.ID == msg.SessionID {
			// Reload events for the active session
			cmds = append(cmds, loadEvents(m.repo.BasePath(), *m.selected))
		} else {
			// Increment unread count for inactive sessions
			m.unread[msg.SessionID]++
			m.sidebar.SetUnread(m.unread)
		}
		// Re-subscribe to watcher
		cmds = append(cmds, watchFiles(m.watcher))

	case SessionDirChangedMsg:
		// New session appeared, refresh list
		cmds = append(cmds, loadSessions(m.repo))
		cmds = append(cmds, watchFiles(m.watcher))
	}

	// Route input to focused panel
	if m.focus == FocusSidebar {
		var cmd tea.Cmd
		m.sidebar, cmd = m.sidebar.Update(msg)
		cmds = append(cmds, cmd)
	} else {
		var cmd tea.Cmd
		m.chat, cmd = m.chat.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}
