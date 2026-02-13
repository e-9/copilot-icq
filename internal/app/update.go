package app

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/e-9/copilot-icq/internal/ui/theme"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "q":
			if m.focus != FocusInput {
				return m, tea.Quit
			}
		case "esc":
			if m.focus == FocusInput {
				m.focus = FocusChat
				m.input.Blur()
				return m, nil
			}
		case "r":
			if m.focus != FocusInput {
				cmds = append(cmds, loadSessions(m.repo))
			}
		case "tab":
			switch m.focus {
			case FocusSidebar:
				m.focus = FocusChat
				m.input.Blur()
			case FocusChat:
				if m.selected != nil && m.runner != nil {
					m.focus = FocusInput
					m.input.Focus()
				} else {
					m.focus = FocusSidebar
					m.input.Blur()
				}
			case FocusInput:
				m.focus = FocusSidebar
				m.input.Blur()
			}
			return m, nil
		case "enter":
			if m.focus == FocusSidebar {
				if s := m.sidebar.SelectedSession(); s != nil {
					m.selected = s
					m.focus = FocusChat
					m.unread[s.ID] = 0
					m.sidebar.SetUnread(m.unread)
					m.watcher.WatchSession(s.ID)
					cmds = append(cmds, loadEvents(m.repo.BasePath(), *s))
				}
			} else if m.focus == FocusInput {
				text := m.input.Value()
				if text != "" && m.selected != nil && m.runner != nil && !m.input.IsSending() {
					m.input.SetSending(true)
					cmds = append(cmds, sendMessage(m.runner, m.selected.ID, text))
				}
			}
		}

	case tea.MouseMsg:
		// Determine which panel was clicked based on X coordinate
		if msg.Action == tea.MouseActionPress || msg.Action == tea.MouseActionMotion {
			if msg.Button == tea.MouseButtonLeft {
				if msg.X < m.sidebar.Width+2 {
					// Clicked on sidebar
					if m.focus != FocusSidebar {
						m.focus = FocusSidebar
						m.input.Blur()
					}
				} else if m.height > 0 && msg.Y >= m.height-5 && m.selected != nil {
					// Clicked on input area (bottom ~3 rows)
					if m.focus != FocusInput && m.runner != nil {
						m.focus = FocusInput
						m.input.Focus()
					}
				} else {
					// Clicked on chat area
					if m.focus != FocusChat {
						m.focus = FocusChat
						m.input.Blur()
					}
				}
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true
		// Border takes 2 rows (top+bottom) and 2 cols (left+right)
		borderH := 2
		borderW := 2
		headerH := 1
		statusBarH := 1

		panelHeight := m.height - headerH - statusBarH - borderH
		if panelHeight < 1 {
			panelHeight = 1
		}
		sidebarInnerW := theme.SidebarWidth
		chatInnerW := m.width - sidebarInnerW - borderW*2
		if chatInnerW < 1 {
			chatInnerW = 1
		}

		m.sidebar.SetSize(sidebarInnerW, panelHeight)

		inputHeight := 1
		inputBorderH := inputHeight + borderH
		chatInnerH := panelHeight - inputBorderH
		m.chat.SetSize(chatInnerW, chatInnerH)
		m.input.SetWidth(chatInnerW)

	case SessionsLoadedMsg:
		if msg.Err != nil {
			m.err = msg.Err
			return m, nil
		}
		m.sessions = msg.Sessions
		m.sidebar.SetItems(msg.Sessions)
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
			cmds = append(cmds, loadEvents(m.repo.BasePath(), *m.selected))
		} else {
			m.unread[msg.SessionID]++
			m.sidebar.SetUnread(m.unread)
		}
		cmds = append(cmds, watchFiles(m.watcher))

	case SessionDirChangedMsg:
		cmds = append(cmds, loadSessions(m.repo))
		cmds = append(cmds, watchFiles(m.watcher))

	case TickMsg:
		// Periodic rescan for new sessions
		cmds = append(cmds, loadSessions(m.repo))
		cmds = append(cmds, tickEvery(5*time.Second))

	case MessageSentMsg:
		m.input.SetSending(false)
		if msg.Err != nil {
			m.err = msg.Err
		} else {
			m.input.Reset()
		}
	}

	// Route input to focused panel
	switch m.focus {
	case FocusSidebar:
		var cmd tea.Cmd
		m.sidebar, cmd = m.sidebar.Update(msg)
		cmds = append(cmds, cmd)
	case FocusChat:
		var cmd tea.Cmd
		m.chat, cmd = m.chat.Update(msg)
		cmds = append(cmds, cmd)
	case FocusInput:
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}
