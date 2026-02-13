package app

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
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

	// Chat panel
	chatView := m.chat.View()
	if m.selected == nil {
		chatWidth := m.width - theme.SidebarWidth - 4
		chatView = lipgloss.NewStyle().
			Width(chatWidth).
			Height(m.height - 2).
			Foreground(theme.Subtle).
			Render("  Select a session and press Enter to view conversation")
	}
	if m.focus == FocusChat {
		chatView = lipgloss.NewStyle().
			BorderForeground(theme.Accent).
			Render(chatView)
	}

	// Layout: sidebar | chat
	content := lipgloss.JoinHorizontal(lipgloss.Top, sidebarView, chatView)

	// Status bar
	focusLabel := "sidebar"
	if m.focus == FocusChat {
		focusLabel = "chat"
	}
	sessionInfo := ""
	if m.selected != nil {
		sessionInfo = fmt.Sprintf(" · %s", m.selected.DisplayName())
	}
	statusBar := theme.StatusBarStyle.
		Width(m.width).
		Render(fmt.Sprintf("  %d sessions%s · [%s] · Tab switch · ↑↓ navigate · Enter open · r refresh · q quit",
			len(m.sessions), sessionInfo, focusLabel))

	return lipgloss.JoinVertical(lipgloss.Left, content, statusBar)
}
