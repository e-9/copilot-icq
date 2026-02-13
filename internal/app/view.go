package app

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/e-9/copilot-icq/internal/infra/runner"
	"github.com/e-9/copilot-icq/internal/ui/theme"
)

var (
	focusedBorder = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(theme.Accent)

	unfocusedBorder = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("238"))
)

func (m Model) View() string {
	if !m.ready {
		return "Loading..."
	}

	if m.err != nil {
		return fmt.Sprintf("Error: %v\nPress q to quit.", m.err)
	}

	sidebarHeight := m.height - 3
	chatWidth := m.width - theme.SidebarWidth - 6

	// Sidebar panel
	sidebarContent := m.sidebar.View()
	sidebarBorder := unfocusedBorder.Width(theme.SidebarWidth).Height(sidebarHeight)
	if m.focus == FocusSidebar {
		sidebarBorder = focusedBorder.Width(theme.SidebarWidth).Height(sidebarHeight)
	}
	sidebarView := sidebarBorder.Render(sidebarContent)

	// Right panel (chat + input)
	var rightPanel string
	if m.selected == nil {
		placeholder := lipgloss.NewStyle().
			Foreground(theme.Subtle).
			Render("  Select a session and press Enter to view conversation")
		rightBorder := unfocusedBorder.Width(chatWidth).Height(sidebarHeight)
		rightPanel = rightBorder.Render(placeholder)
	} else {
		inputHeight := 3
		chatHeight := sidebarHeight - inputHeight - 2

		// Chat viewport
		chatContent := m.chat.View()
		chatBorder := unfocusedBorder.Width(chatWidth).Height(chatHeight)
		if m.focus == FocusChat {
			chatBorder = focusedBorder.Width(chatWidth).Height(chatHeight)
		}
		chatView := chatBorder.Render(chatContent)

		// Input area
		inputContent := m.input.View()
		inputBorder := unfocusedBorder.Width(chatWidth).Height(inputHeight)
		if m.focus == FocusInput {
			inputBorder = focusedBorder.Width(chatWidth).Height(inputHeight)
		}
		inputView := inputBorder.Render(inputContent)

		rightPanel = lipgloss.JoinVertical(lipgloss.Left, chatView, inputView)
	}

	// Layout: sidebar | right
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

	sendingInfo := ""
	if m.input.IsSending() {
		sendingInfo = " · ⏳ sending..."
	}

	statusBar := theme.StatusBarStyle.
		Width(m.width).
		Render(fmt.Sprintf("  %d sessions%s%s%s · [%s] · Tab/Click switch · Esc back · q quit",
			len(m.sessions), sessionInfo, securityIcon, sendingInfo, focusLabel))

	return lipgloss.JoinVertical(lipgloss.Left, content, statusBar)
}
