package app

import tea "github.com/charmbracelet/bubbletea"

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "r":
			// Refresh sessions
			cmds = append(cmds, loadSessions(m.repo))
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true
		sidebarHeight := m.height - 2 // leave room for status bar
		m.sidebar.SetSize(m.sidebar.Width, sidebarHeight)

	case SessionsLoadedMsg:
		if msg.Err != nil {
			m.err = msg.Err
			return m, nil
		}
		m.sessions = msg.Sessions
		m.sidebar.SetItems(msg.Sessions)
	}

	// Update sidebar
	var cmd tea.Cmd
	m.sidebar, cmd = m.sidebar.Update(msg)
	cmds = append(cmds, cmd)

	// Track selected session
	m.selected = m.sidebar.SelectedSession()

	return m, tea.Batch(cmds...)
}
