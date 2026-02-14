package chat

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/e-9/copilot-icq/internal/domain"
	"github.com/e-9/copilot-icq/internal/ui/theme"
)

// Model represents the chat panel showing conversation history.
type Model struct {
	viewport  viewport.Model
	messages  []domain.Message
	width     int
	height    int
	ready     bool
	mdRender  *glamour.TermRenderer
	collapsed map[int]bool // tool call indices collapsed state
}

// New creates a new chat panel.
func New(width, height int) Model {
	vp := viewport.New(width, height)
	vp.SetContent("")

	r, _ := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(width-4),
	)

	return Model{
		viewport:  vp,
		width:     width,
		height:    height,
		ready:     true,
		mdRender:  r,
		collapsed: make(map[int]bool),
	}
}

// SetSize updates the chat panel dimensions.
func (m *Model) SetSize(w, h int) {
	m.width = w
	m.height = h
	m.viewport.Width = w
	m.viewport.Height = h
	m.mdRender, _ = glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(w-4),
	)
}

// SetMessages replaces the displayed messages and re-renders.
func (m *Model) SetMessages(msgs []domain.Message) {
	m.messages = msgs
	m.collapsed = make(map[int]bool) // reset collapsed state
	m.viewport.SetContent(m.renderMessages())
	m.viewport.GotoBottom()
}

// Messages returns the current messages for export.
func (m Model) Messages() []domain.Message {
	return m.messages
}

// AppendMessages adds new messages and scrolls to bottom.
func (m *Model) AppendMessages(msgs []domain.Message) {
	m.messages = append(m.messages, msgs...)
	m.viewport.SetContent(m.renderMessages())
	m.viewport.GotoBottom()
}

// ToggleAllToolCalls toggles collapse state for all tool calls.
func (m *Model) ToggleAllToolCalls() {
	// Count total tool calls
	total := 0
	for _, msg := range m.messages {
		total += len(msg.ToolCalls)
	}
	if total == 0 {
		return
	}

	// If any are expanded, collapse all; otherwise expand all
	anyExpanded := false
	for i := 0; i < total; i++ {
		if !m.collapsed[i] {
			anyExpanded = true
			break
		}
	}

	for i := 0; i < total; i++ {
		m.collapsed[i] = anyExpanded
	}

	m.viewport.SetContent(m.renderMessages())
}

// Update handles viewport messages (scrolling).
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

// View renders the chat panel.
func (m Model) View() string {
	return m.viewport.View()
}

func (m Model) renderMessages() string {
	if len(m.messages) == 0 {
		return lipgloss.NewStyle().
			Foreground(theme.Subtle).
			Render("  No messages yet")
	}

	var sb strings.Builder
	toolIdx := 0
	for _, msg := range m.messages {
		sb.WriteString(m.renderMessage(msg, m.width-2, &toolIdx))
		sb.WriteString("\n")
	}
	return sb.String()
}

var (
	userLabelStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("86"))

	assistantLabelStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("205"))

	systemLabelStyle = lipgloss.NewStyle().
				Foreground(theme.Subtle).
				Italic(true)

	timestampStyle = lipgloss.NewStyle().
			Foreground(theme.Subtle)

	toolCallStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("214"))

	contentStyle = lipgloss.NewStyle()
)

func (m Model) renderMessage(msg domain.Message, maxWidth int, toolIdx *int) string {
	var sb strings.Builder
	ts := timestampStyle.Render(msg.Timestamp.Format("15:04"))

	switch msg.Role {
	case domain.RoleUser:
		label := userLabelStyle.Render("You")
		sb.WriteString(fmt.Sprintf("%s %s\n", label, ts))
		sb.WriteString(contentStyle.Width(maxWidth).Render(msg.Content))

	case domain.RoleAssistant:
		label := assistantLabelStyle.Render("Copilot")
		sb.WriteString(fmt.Sprintf("%s %s\n", label, ts))
		if msg.Content != "" {
			content := strings.TrimSpace(msg.Content)
			if content != "" {
				rendered := m.renderMarkdown(content)
				if rendered != "" {
					sb.WriteString(rendered)
				} else {
					sb.WriteString(contentStyle.Width(maxWidth).Render(content))
					sb.WriteString("\n")
				}
			}
		}
		for _, tc := range msg.ToolCalls {
			sb.WriteString(m.renderToolCall(tc, *toolIdx))
			sb.WriteString("\n")
			*toolIdx++
		}

	case domain.RoleSystem:
		sb.WriteString(fmt.Sprintf("%s %s\n", systemLabelStyle.Render("ℹ "+msg.Content), ts))

	case domain.RoleTool:
		sb.WriteString(fmt.Sprintf("  %s\n", toolCallStyle.Render(msg.Content)))
	}

	return sb.String()
}

// renderMarkdown renders markdown content using Glamour.
func (m Model) renderMarkdown(content string) string {
	if m.mdRender == nil {
		return ""
	}
	out, err := m.mdRender.Render(content)
	if err != nil {
		return ""
	}
	return strings.TrimRight(out, "\n")
}

func (m Model) renderToolCall(tc domain.ToolCall, idx int) string {
	var icon string
	switch tc.Status {
	case domain.ToolCallComplete:
		icon = "✓"
	case domain.ToolCallFailed:
		icon = "✗"
	case domain.ToolCallRunning:
		icon = "⟳ awaiting approval"
	case domain.ToolCallPending:
		icon = "⏳ pending"
	default:
		icon = "…"
	}

	collapsed := m.collapsed[idx]
	chevron := "▸" // collapsed
	if !collapsed {
		chevron = "▾" // expanded
	}

	header := toolCallStyle.Render(fmt.Sprintf("  %s 🔧 %s %s", chevron, tc.Name, icon))

	// For pending/running tools, always show the command if available
	isPending := tc.Status == domain.ToolCallPending || tc.Status == domain.ToolCallRunning
	if isPending && tc.Command != "" {
		cmdStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("251")).
			Background(lipgloss.Color("237")).
			Padding(0, 1)
		cmdBlock := cmdStyle.Render("$ " + tc.Command)
		waitingHint := lipgloss.NewStyle().
			Foreground(theme.Warning).
			Bold(true).
			Render("    ⚡ Waiting for approval in terminal")
		return header + "\n" + "    " + cmdBlock + "\n" + waitingHint
	}

	if collapsed {
		return header
	}

	if tc.Summary == "" {
		return header
	}

	// Expanded: show summary (truncated to 3 lines max)
	summary := tc.Summary
	lines := strings.Split(summary, "\n")
	if len(lines) > 3 {
		summary = strings.Join(lines[:3], "\n") + "\n    ..."
	}
	detail := lipgloss.NewStyle().Foreground(theme.Subtle).Render("    " + strings.ReplaceAll(summary, "\n", "\n    "))
	return header + "\n" + detail
}
