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

	// Sidebar
	sidebarView := m.sidebar.View()

	// Chat panel (placeholder — Phase 2)
	chatWidth := m.width - theme.SidebarWidth - 4
	chatHeight := m.height - 2

	var chatContent string
	if m.selected != nil {
		chatContent = lipgloss.NewStyle().
			Bold(true).
			Foreground(theme.Accent).
			Render(m.selected.DisplayName()) +
			"\n" +
			lipgloss.NewStyle().
				Foreground(theme.Subtle).
				Render(fmt.Sprintf("ID: %s\nCWD: %s\nCreated: %s\nUpdated: %s",
					m.selected.ID,
					m.selected.CWD,
					m.selected.CreatedAt.Format("2006-01-02 15:04"),
					m.selected.UpdatedAt.Format("2006-01-02 15:04"),
				))
	} else {
		chatContent = lipgloss.NewStyle().
			Foreground(theme.Subtle).
			Render("Select a session to view")
	}

	chatPanel := theme.ChatPanelStyle.
		Width(chatWidth).
		Height(chatHeight).
		Render(chatContent)

	// Layout: sidebar | chat
	content := lipgloss.JoinHorizontal(lipgloss.Top, sidebarView, chatPanel)

	// Status bar
	statusBar := theme.StatusBarStyle.
		Width(m.width).
		Render(fmt.Sprintf("  %d sessions · ↑↓ navigate · r refresh · q quit", len(m.sessions)))

	return lipgloss.JoinVertical(lipgloss.Left, content, statusBar)
}
