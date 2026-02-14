package app

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/e-9/copilot-icq/internal/domain"
	"github.com/e-9/copilot-icq/internal/infra/notifier"
	"github.com/e-9/copilot-icq/internal/ui/theme"
	"gopkg.in/yaml.v3"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Help overlay: any key dismisses
		if m.showHelp {
			m.showHelp = false
			return m, nil
		}

		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "q":
			if m.focus != FocusInput {
				return m, tea.Quit
			}
		case "esc":
			if m.renaming {
				m.renaming = false
				m.input.Reset()
				m.input.ClearRenaming()
				m.focus = FocusSidebar
				m.input.Blur()
				return m, nil
			}
			if m.focus == FocusInput {
				m.focus = FocusChat
				m.input.Blur()
				return m, nil
			}
		case "?":
			if m.focus != FocusInput {
				m.showHelp = !m.showHelp
				return m, nil
			}
		case "e":
			if m.focus != FocusInput && m.selected != nil {
				return m, m.exportConversation()
			}
		case "R":
			// Shift+R: rename session (sidebar only)
			if m.focus == FocusSidebar {
				if s := m.sidebar.SelectedSession(); s != nil {
					m.renaming = true
					m.focus = FocusInput
					m.input.SetRenaming(s.DisplayName())
				}
				return m, nil
			}
		case "r":
			if m.focus != FocusInput {
				cmds = append(cmds, loadSessions(m.repo))
			}
		case "t":
			if m.focus == FocusChat {
				m.chat.ToggleAllToolCalls()
				return m, nil
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
		case "shift+tab":
			switch m.focus {
			case FocusSidebar:
				if m.selected != nil && m.runner != nil {
					m.focus = FocusInput
					m.input.Focus()
				} else {
					m.focus = FocusChat
					m.input.Blur()
				}
			case FocusChat:
				m.focus = FocusSidebar
				m.input.Blur()
			case FocusInput:
				m.focus = FocusChat
				m.input.Blur()
			}
			return m, nil
		case "enter":
			if m.focus == FocusSidebar && !m.sidebar.IsFiltering() {
				if s := m.sidebar.SelectedSession(); s != nil {
					m.selected = s
					m.focus = FocusChat
					m.unread[s.ID] = 0
					m.sidebar.SetActiveID(s.ID)
					m.sidebar.SetUnread(m.unread)
					m.sidebar.SetItems(m.sessions) // re-sort with new active
					m.watcher.WatchSession(s.ID)
					cmds = append(cmds, loadEvents(m.repo.BasePath(), *s))
				}
			} else if m.focus == FocusInput {
				if m.renaming {
					// Complete rename
					newName := m.input.Value()
					if newName != "" && m.sidebar.SelectedSession() != nil {
						s := m.sidebar.SelectedSession()
						cmds = append(cmds, m.renameSession(s.ID, newName))
					}
					m.renaming = false
					m.input.Reset()
					m.input.ClearRenaming()
					m.focus = FocusSidebar
					m.input.Blur()
				} else {
					text := m.input.Value()
					if text != "" && m.selected != nil && m.runner != nil && !m.input.IsSending() {
						m.input.SetSending(true)
						cmds = append(cmds, sendMessage(m.runner, m.selected.ID, text))
					}
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
		m.sidebar.SetItems(m.sessions) // re-sort: unread sessions bubble up
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

	case notifier.HookEventMsg:
		// Hook event received from companion binary
		if msg.SessionID != "" {
			m.lastSeen[msg.SessionID] = time.Now()
			m.sidebar.SetLastSeen(m.lastSeen)
			if m.selected == nil || m.selected.ID != msg.SessionID {
				m.unread[msg.SessionID]++
				m.sidebar.SetUnread(m.unread)
			} else {
				cmds = append(cmds, loadEvents(m.repo.BasePath(), *m.selected))
			}
			m.sidebar.SetItems(m.sessions)
		}

	case SessionRenamedMsg:
		if msg.Err == nil {
			cmds = append(cmds, loadSessions(m.repo))
		}

	case ExportCompleteMsg:
		// Nothing to do in the model; status bar could show a flash
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

// renameSession updates the summary field in workspace.yaml.
func (m Model) renameSession(sessionID, newName string) tea.Cmd {
	return func() tea.Msg {
		wsPath := filepath.Join(m.repo.BasePath(), sessionID, "workspace.yaml")
		data, err := os.ReadFile(wsPath)
		if err != nil {
			return SessionRenamedMsg{SessionID: sessionID, Err: err}
		}

		var ws map[string]interface{}
		if err := yaml.Unmarshal(data, &ws); err != nil {
			return SessionRenamedMsg{SessionID: sessionID, Err: err}
		}

		ws["summary"] = newName
		out, err := yaml.Marshal(ws)
		if err != nil {
			return SessionRenamedMsg{SessionID: sessionID, Err: err}
		}

		if err := os.WriteFile(wsPath, out, 0600); err != nil {
			return SessionRenamedMsg{SessionID: sessionID, Err: err}
		}

		return SessionRenamedMsg{SessionID: sessionID}
	}
}

// exportConversation writes the current conversation to a markdown file.
func (m Model) exportConversation() tea.Cmd {
	return func() tea.Msg {
		if m.selected == nil {
			return ExportCompleteMsg{Err: fmt.Errorf("no session selected")}
		}

		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("# Copilot Session: %s\n\n", m.selected.DisplayName()))
		sb.WriteString(fmt.Sprintf("- **Session ID**: `%s`\n", m.selected.ID))
		sb.WriteString(fmt.Sprintf("- **CWD**: `%s`\n", m.selected.CWD))
		sb.WriteString(fmt.Sprintf("- **Created**: %s\n", m.selected.CreatedAt.Format(time.RFC3339)))
		sb.WriteString(fmt.Sprintf("- **Updated**: %s\n\n", m.selected.UpdatedAt.Format(time.RFC3339)))
		sb.WriteString("---\n\n")

		msgs := m.chat.Messages()
		for _, msg := range msgs {
			ts := msg.Timestamp.Format("15:04:05")
			switch msg.Role {
			case domain.RoleUser:
				sb.WriteString(fmt.Sprintf("### 🧑 You (%s)\n\n%s\n\n", ts, msg.Content))
			case domain.RoleAssistant:
				sb.WriteString(fmt.Sprintf("### 🤖 Copilot (%s)\n\n", ts))
				if msg.Content != "" {
					sb.WriteString(msg.Content + "\n\n")
				}
				for _, tc := range msg.ToolCalls {
					status := "⏳"
					if tc.Status == domain.ToolCallComplete {
						status = "✓"
					} else if tc.Status == domain.ToolCallFailed {
						status = "✗"
					}
					sb.WriteString(fmt.Sprintf("- 🔧 **%s** %s", tc.Name, status))
					if tc.Summary != "" {
						summary := tc.Summary
						if len(summary) > 100 {
							summary = summary[:97] + "..."
						}
						sb.WriteString(fmt.Sprintf(": `%s`", summary))
					}
					sb.WriteString("\n")
				}
				sb.WriteString("\n")
			case domain.RoleSystem:
				sb.WriteString(fmt.Sprintf("*ℹ %s* (%s)\n\n", msg.Content, ts))
			}
		}

		exportDir := "."
		if m.cfg != nil && m.cfg.ExportDir != "" {
			exportDir = m.cfg.ExportDir
		}

		filename := fmt.Sprintf("copilot-session-%s.md", m.selected.ShortID())
		outPath := filepath.Join(exportDir, filename)
		if err := os.WriteFile(outPath, []byte(sb.String()), 0644); err != nil {
			return ExportCompleteMsg{Err: err}
		}

		return ExportCompleteMsg{Path: outPath}
	}
}
