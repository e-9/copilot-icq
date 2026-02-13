package chat

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/e-9/copilot-icq/internal/domain"
	"github.com/e-9/copilot-icq/internal/ui/theme"
)

// Model represents the chat panel showing conversation history.
type Model struct {
	viewport viewport.Model
	messages []domain.Message
	width    int
	height   int
	ready    bool
}

// New creates a new chat panel.
func New(width, height int) Model {
	vp := viewport.New(width, height)
	vp.SetContent("")
	return Model{
		viewport: vp,
		width:    width,
		height:   height,
		ready:    true,
	}
}

// SetSize updates the chat panel dimensions.
func (m *Model) SetSize(w, h int) {
	m.width = w
	m.height = h
	m.viewport.Width = w
	m.viewport.Height = h
}

// SetMessages replaces the displayed messages and re-renders.
func (m *Model) SetMessages(msgs []domain.Message) {
	m.messages = msgs
	m.viewport.SetContent(m.renderMessages())
	m.viewport.GotoBottom()
}

// AppendMessages adds new messages and scrolls to bottom.
func (m *Model) AppendMessages(msgs []domain.Message) {
	m.messages = append(m.messages, msgs...)
	m.viewport.SetContent(m.renderMessages())
	m.viewport.GotoBottom()
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
	for _, msg := range m.messages {
		sb.WriteString(renderMessage(msg, m.width-2))
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

func renderMessage(msg domain.Message, maxWidth int) string {
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
				sb.WriteString(contentStyle.Width(maxWidth).Render(content))
				sb.WriteString("\n")
			}
		}
		for _, tc := range msg.ToolCalls {
			sb.WriteString(renderToolCall(tc))
			sb.WriteString("\n")
		}

	case domain.RoleSystem:
		sb.WriteString(fmt.Sprintf("%s %s\n", systemLabelStyle.Render("ℹ "+msg.Content), ts))

	case domain.RoleTool:
		sb.WriteString(fmt.Sprintf("  %s\n", toolCallStyle.Render(msg.Content)))
	}

	return sb.String()
}

func renderToolCall(tc domain.ToolCall) string {
	var icon string
	switch tc.Status {
	case domain.ToolCallComplete:
		icon = "✓"
	case domain.ToolCallFailed:
		icon = "✗"
	case domain.ToolCallRunning:
		icon = "⟳"
	default:
		icon = "…"
	}

	result := toolCallStyle.Render(fmt.Sprintf("  🔧 %s %s", tc.Name, icon))
	if tc.Summary != "" {
		summary := tc.Summary
		if len(summary) > 60 {
			summary = summary[:57] + "..."
		}
		result += " " + lipgloss.NewStyle().Foreground(theme.Subtle).Render(summary)
	}
	return result
}
