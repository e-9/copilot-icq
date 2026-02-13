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
	headerH := 1
	statusBarH := 1

	// Header bar
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.Accent).
		Render("🟢 Copilot ICQ")

	sessionCount := lipgloss.NewStyle().
		Foreground(theme.Subtle).
		Render(fmt.Sprintf(" %d sessions", len(m.sessions)))

	sendingInfo := ""
	if m.input.IsSending() {
		sendingInfo = lipgloss.NewStyle().
			Foreground(theme.Warning).
			Bold(true).
			Render("  ⏳ sending...")
	}

	securityIcon := ""
	if m.runner != nil {
		if m.runner.Mode() == runner.ModeScoped {
			securityIcon = lipgloss.NewStyle().Foreground(theme.Highlight).Render("  🔒 scoped")
		} else {
			securityIcon = lipgloss.NewStyle().Foreground(theme.Warning).Render("  ⚠️ full-auto")
		}
	}

	shortcuts := lipgloss.NewStyle().
		Foreground(theme.Subtle).
		Render("  Tab·Click switch  Enter open/send  t tools  Esc back  / filter  q quit")

	headerLeft := title + sessionCount + securityIcon + sendingInfo
	headerRight := shortcuts
	headerGap := m.width - lipgloss.Width(headerLeft) - lipgloss.Width(headerRight) - 2
	if headerGap < 0 {
		headerGap = 0
		headerRight = ""
	}
	headerBar := lipgloss.NewStyle().
		Width(m.width).
		MaxWidth(m.width).
		Background(lipgloss.Color("236")).
		Padding(0, 1).
		Render(headerLeft + lipgloss.NewStyle().Width(headerGap).Render("") + headerRight)

	panelHeight := m.height - headerH - statusBarH - borderH
	if panelHeight < 1 {
		panelHeight = 1
	}
	sidebarInnerW := theme.SidebarWidth
	chatInnerW := m.width - sidebarInnerW - borderW*2
	if chatInnerW < 1 {
		chatInnerW = 1
	}

	// Sidebar panel — MaxHeight clips overflow from list component
	sidebarRenderedH := panelHeight + borderH
	sidebarContent := m.sidebar.View()
	sidebarBorder := unfocusedBorder.Width(sidebarInnerW).Height(panelHeight).MaxHeight(sidebarRenderedH)
	if m.focus == FocusSidebar {
		sidebarBorder = focusedBorder.Width(sidebarInnerW).Height(panelHeight).MaxHeight(sidebarRenderedH)
	}
	sidebarView := sidebarBorder.Render(sidebarContent)

	// Right panel (chat + input)
	var rightPanel string
	if m.selected == nil {
		placeholder := lipgloss.NewStyle().
			Foreground(theme.Subtle).
			Render("  Select a session and press Enter to view conversation")
		rightBorder := unfocusedBorder.Width(chatInnerW).Height(panelHeight).MaxHeight(sidebarRenderedH)
		rightPanel = rightBorder.Render(placeholder)
	} else {
		inputInnerH := 1
		chatInnerH := panelHeight - inputInnerH - borderH
		chatRenderedH := chatInnerH + borderH
		inputRenderedH := inputInnerH + borderH

		// Chat viewport
		chatContent := m.chat.View()
		chatBorder := unfocusedBorder.Width(chatInnerW).Height(chatInnerH).MaxHeight(chatRenderedH)
		if m.focus == FocusChat {
			chatBorder = focusedBorder.Width(chatInnerW).Height(chatInnerH).MaxHeight(chatRenderedH)
		}
		chatView := chatBorder.Render(chatContent)

		// Input area
		inputContent := m.input.View()
		inputBorder := unfocusedBorder.Width(chatInnerW).Height(inputInnerH).MaxHeight(inputRenderedH)
		if m.focus == FocusInput {
			inputBorder = focusedBorder.Width(chatInnerW).Height(inputInnerH).MaxHeight(inputRenderedH)
		}
		inputView := inputBorder.Render(inputContent)

		rightPanel = lipgloss.JoinVertical(lipgloss.Left, chatView, inputView)
	}

	// Layout: sidebar | right
	content := lipgloss.JoinHorizontal(lipgloss.Top, sidebarView, rightPanel)

	// Status bar — contextual info
	focusLabel := "sidebar"
	switch m.focus {
	case FocusChat:
		focusLabel = "chat"
	case FocusInput:
		focusLabel = "input"
	}

	sessionInfo := ""
	if m.selected != nil {
		sessionInfo = fmt.Sprintf(" · 💬 %s (%s)", m.selected.DisplayName(), m.selected.ShortID())
	}

	statusBar := theme.StatusBarStyle.
		Width(m.width).
		Render(fmt.Sprintf(" [%s]%s", focusLabel, sessionInfo))

	return lipgloss.JoinVertical(lipgloss.Left, headerBar, content, statusBar)
}
