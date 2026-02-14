package app

import (
	"fmt"
	"strings"

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

	// Help overlay
	if m.showHelp {
		return m.renderHelpOverlay()
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
		Render("  ? help  e export  R rename  q quit")

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

	modeLabel := ""
	if m.renaming {
		modeLabel = " · ✏️ renaming"
	}

	statusBar := theme.StatusBarStyle.
		Width(m.width).
		Render(fmt.Sprintf(" [%s]%s%s", focusLabel, sessionInfo, modeLabel))

	return lipgloss.JoinVertical(lipgloss.Left, headerBar, content, statusBar)
}

func (m Model) renderHelpOverlay() string {
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(theme.Accent)
	keyStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("252"))
	descStyle := lipgloss.NewStyle().Foreground(theme.Subtle)

	bindings := []struct{ key, desc string }{
		{"Tab", "Switch panel forward (sidebar → chat → input)"},
		{"Shift+Tab", "Switch panel backward (input → chat → sidebar)"},
		{"Enter", "Open session (sidebar) / Send message (input)"},
		{"Esc", "Go back (input → chat, cancel rename)"},
		{"↑ ↓", "Navigate sessions / scroll chat"},
		{"/ (sidebar)", "Filter sessions by name"},
		{"?", "Toggle this help overlay"},
		{"t", "Toggle tool call details (expand/collapse)"},
		{"r", "Refresh session list"},
		{"R (Shift+R)", "Rename selected session"},
		{"e", "Export conversation to markdown"},
		{"q", "Quit (not active in input mode)"},
		{"Ctrl+C", "Force quit"},
	}

	var sb strings.Builder
	sb.WriteString(titleStyle.Render("🟢 Copilot ICQ — Keyboard Shortcuts"))
	sb.WriteString("\n\n")

	for _, b := range bindings {
		sb.WriteString(fmt.Sprintf("  %s  %s\n",
			keyStyle.Width(18).Render(b.key),
			descStyle.Render(b.desc),
		))
	}

	sb.WriteString("\n")
	sb.WriteString(descStyle.Render("  Press any key to close this overlay"))

	overlay := lipgloss.NewStyle().
		Width(m.width - 4).
		Padding(2, 4).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.Accent).
		Render(sb.String())

	return lipgloss.Place(m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		overlay)
}
