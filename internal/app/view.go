package app

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/e-9/copilot-icq/internal/infra/runner"
	"github.com/e-9/copilot-icq/internal/ui/theme"
)

func (m Model) View() string {
	if !m.ready {
		return "Loading..."
	}

	if m.err != nil {
		return fmt.Sprintf("Error: %v\nPress q to quit.", m.err)
	}

	// Sidebar with focus indicator
	sidebarView := m.sidebar.View()
	if m.focus == FocusSidebar {
		sidebarView = lipgloss.NewStyle().
			BorderForeground(theme.Accent).
			Render(sidebarView)
	}

	// Chat + input panel
	var rightPanel string
	if m.selected == nil {
		chatWidth := m.width - theme.SidebarWidth - 4
		rightPanel = lipgloss.NewStyle().
			Width(chatWidth).
			Height(m.height - 2).
			Foreground(theme.Subtle).
			Render("  Select a session and press Enter to view conversation")
	} else {
		chatView := m.chat.View()
		if m.focus == FocusChat {
			chatView = lipgloss.NewStyle().
				BorderForeground(theme.Accent).
				Render(chatView)
		}

		inputView := m.input.View()
		if m.focus == FocusInput {
			inputView = lipgloss.NewStyle().
				BorderForeground(theme.Accent).
				Render(inputView)
		}

		rightPanel = lipgloss.JoinVertical(lipgloss.Left, chatView, inputView)
	}

	// Layout: sidebar | chat+input
	content := lipgloss.JoinHorizontal(lipgloss.Top, sidebarView, rightPanel)

	// Status bar
	focusLabel := "sidebar"
	switch m.focus {
	case FocusChat:
		focusLabel = "chat"
	case FocusInput:
		focusLabel = "input"
	}

	sessionInfo := ""
	if m.selected != nil {
		sessionInfo = fmt.Sprintf(" · %s", m.selected.DisplayName())
	}

	securityIcon := ""
	if m.runner != nil {
		if m.runner.Mode() == runner.ModeScoped {
			securityIcon = " · 🔒 scoped"
		} else {
			securityIcon = " · ⚠️ full-auto"
		}
	}

	statusBar := theme.StatusBarStyle.
		Width(m.width).
		Render(fmt.Sprintf("  %d sessions%s%s · [%s] · Tab switch · Enter open/send · q quit",
			len(m.sessions), sessionInfo, securityIcon, focusLabel))

	return lipgloss.JoinVertical(lipgloss.Left, content, statusBar)
}
