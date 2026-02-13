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

	borderH := 2
	borderW := 2
	statusBarH := 1

	panelHeight := m.height - statusBarH - borderH
	sidebarInnerW := theme.SidebarWidth
	chatInnerW := m.width - sidebarInnerW - borderW*2

	// Sidebar panel
	sidebarContent := m.sidebar.View()
	sidebarBorder := unfocusedBorder.Width(sidebarInnerW).Height(panelHeight)
	if m.focus == FocusSidebar {
		sidebarBorder = focusedBorder.Width(sidebarInnerW).Height(panelHeight)
	}
	sidebarView := sidebarBorder.Render(sidebarContent)

	// Right panel (chat + input)
	var rightPanel string
	if m.selected == nil {
		placeholder := lipgloss.NewStyle().
			Foreground(theme.Subtle).
			Render("  Select a session and press Enter to view conversation")
		rightBorder := unfocusedBorder.Width(chatInnerW).Height(panelHeight)
		rightPanel = rightBorder.Render(placeholder)
	} else {
		inputInnerH := 1
		chatInnerH := panelHeight - inputInnerH - borderH

		// Chat viewport
		chatContent := m.chat.View()
		chatBorder := unfocusedBorder.Width(chatInnerW).Height(chatInnerH)
		if m.focus == FocusChat {
			chatBorder = focusedBorder.Width(chatInnerW).Height(chatInnerH)
		}
		chatView := chatBorder.Render(chatContent)

		// Input area
		inputContent := m.input.View()
		inputBorder := unfocusedBorder.Width(chatInnerW).Height(inputInnerH)
		if m.focus == FocusInput {
			inputBorder = focusedBorder.Width(chatInnerW).Height(inputInnerH)
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
