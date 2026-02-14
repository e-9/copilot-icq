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
	viewport      viewport.Model
	messages      []domain.Message
	width         int
	height        int
	ready         bool
	mdRender      *glamour.TermRenderer
	collapsed     map[int]bool   // tool call indices collapsed state
	renderCache   map[int]string // message index → pre-rendered string
	streamingText string         // live PTY streaming output (shown at bottom)
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
		viewport:    vp,
		width:       width,
		height:      height,
		ready:       true,
		mdRender:    r,
		collapsed:   make(map[int]bool),
		renderCache: make(map[int]string),
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
	m.renderCache = make(map[int]string) // invalidate on resize
}

// SetMessages replaces the displayed messages and re-renders.
func (m *Model) SetMessages(msgs []domain.Message) {
	m.messages = msgs
	m.collapsed = make(map[int]bool)
	// Keep cache for unchanged messages; invalidate last 3 (may have status changes)
	cutoff := len(msgs) - 3
	for k := range m.renderCache {
		if k >= cutoff {
			delete(m.renderCache, k)
		}
	}
	// Remove cache entries beyond current message count
	for k := range m.renderCache {
		if k >= len(msgs) {
			delete(m.renderCache, k)
		}
	}
	m.refreshContent()
}

// SetStreamingText sets live PTY output shown at the bottom of the chat.
func (m *Model) SetStreamingText(text string) {
	m.streamingText = text
	m.refreshContent()
}

// sanitizeStreamText removes control characters that corrupt TUI layout.
func sanitizeStreamText(s string) string {
	var sb strings.Builder
	for _, r := range s {
		if r == '\n' || r == '\t' || (r >= 32 && r != 127) {
			sb.WriteRune(r)
		}
	}
	return sb.String()
}

func (m *Model) refreshContent() {
	content := m.renderMessages()
	if m.streamingText != "" {
		cleaned := sanitizeStreamText(m.streamingText)
		lines := strings.Split(strings.TrimSpace(cleaned), "\n")
		// Filter out empty/whitespace-only lines
		var nonEmpty []string
		for _, l := range lines {
			if strings.TrimSpace(l) != "" {
				nonEmpty = append(nonEmpty, l)
			}
		}
		if len(nonEmpty) > 6 {
			nonEmpty = nonEmpty[len(nonEmpty)-6:]
		}
		if len(nonEmpty) > 0 {
			// Truncate long lines to prevent horizontal overflow
			maxW := m.width - 8
			if maxW < 20 {
				maxW = 20
			}
			for i, l := range nonEmpty {
				if len(l) > maxW {
					nonEmpty[i] = l[:maxW] + "…"
				}
			}
			tail := strings.Join(nonEmpty, "\n")
			streamStyle := lipgloss.NewStyle().Foreground(theme.Accent)
			content += "\n" + streamStyle.Render("⟳ "+tail)
		}
	}
	m.viewport.SetContent(content)
	m.viewport.GotoBottom()
}

// Messages returns the current messages for export.
func (m Model) Messages() []domain.Message {
	return m.messages
}

// AppendMessages adds new messages and scrolls to bottom.
func (m *Model) AppendMessages(msgs []domain.Message) {
	m.messages = append(m.messages, msgs...)
	m.refreshContent()
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
	toolIdx := 0
	for i, msg := range m.messages {
		if cached, ok := m.renderCache[i]; ok {
			sb.WriteString(cached)
			// Advance toolIdx past this message's tool calls
			toolIdx += len(msg.ToolCalls)
		} else {
			rendered := m.renderMessage(msg, m.width-2, &toolIdx)
			m.renderCache[i] = rendered
			sb.WriteString(rendered)
		}
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
	if tc.Name == "ask_user" {
		header = toolCallStyle.Render(fmt.Sprintf("  %s ❓ %s %s", chevron, tc.Name, icon))
	}
	if tc.Name == "edit" || tc.Name == "create" || tc.Name == "apply_patch" {
		header = toolCallStyle.Render(fmt.Sprintf("  %s 📝 %s %s", chevron, tc.Name, icon))
	}

	// For pending/running tools, always show the command if available
	isPending := tc.Status == domain.ToolCallPending || tc.Status == domain.ToolCallRunning
	if isPending && tc.Command != "" {
		cmdStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("251")).
			Background(lipgloss.Color("237")).
			Padding(0, 1)
		cmdBlock := cmdStyle.Render("$ " + tc.Command)
		approvalStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("251")).Background(lipgloss.Color("237")).Padding(0, 1)
		approvalLine := "    " + approvalStyle.Render("❯ Allow   Deny   Always allow   Don't allow")
		waitingHint := lipgloss.NewStyle().
			Foreground(theme.Warning).
			Bold(true).
			Render("    ⚡ Respond in terminal")
		return header + "\n" + "    " + cmdBlock + "\n" + approvalLine + "\n" + waitingHint
	}

	// For ask_user tools, show question and choices
	if tc.Question != "" {
		qStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("229")).Bold(true)
		choiceStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("251")).
			Background(lipgloss.Color("237")).
			Padding(0, 1)
		var detail strings.Builder
		detail.WriteString("    " + qStyle.Render(tc.Question) + "\n")
		if len(tc.Choices) > 0 {
			for i, c := range tc.Choices {
				detail.WriteString(fmt.Sprintf("    %s\n", choiceStyle.Render(fmt.Sprintf("[%d] %s", i+1, c))))
			}
		}
		if isPending {
			waitingHint := lipgloss.NewStyle().
				Foreground(theme.Warning).
				Bold(true).
				Render("    ⚡ Respond in terminal")
			return header + "\n" + detail.String() + waitingHint
		}
		if tc.Summary != "" {
			responseStyle := lipgloss.NewStyle().Foreground(theme.Accent).Bold(true)
			detail.WriteString("    " + responseStyle.Render("→ "+tc.Summary) + "\n")
		}
		return header + "\n" + detail.String()
	}

	// For edit/create/apply_patch tools, show file path and diff
	if tc.FilePath != "" {
		pathStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("117")).Bold(true)
		var detail strings.Builder
		detail.WriteString("    " + pathStyle.Render(tc.FilePath) + "\n")
		if !collapsed && tc.Patch != "" {
			addStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("114"))
			delStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("210"))
			patchLines := strings.Split(tc.Patch, "\n")
			maxLines := 12
			shown := 0
			for _, line := range patchLines {
				if shown >= maxLines {
					detail.WriteString("    " + lipgloss.NewStyle().Foreground(theme.Subtle).Render(fmt.Sprintf("... +%d more lines", len(patchLines)-shown)) + "\n")
					break
				}
				if strings.HasPrefix(line, "+") {
					detail.WriteString("    " + addStyle.Render(line) + "\n")
				} else if strings.HasPrefix(line, "-") {
					detail.WriteString("    " + delStyle.Render(line) + "\n")
				} else if strings.HasPrefix(line, "***") || strings.HasPrefix(line, "@@") {
					detail.WriteString("    " + lipgloss.NewStyle().Foreground(theme.Subtle).Render(line) + "\n")
				} else if strings.TrimSpace(line) != "" {
					detail.WriteString("    " + line + "\n")
				}
				shown++
			}
		}
		if isPending {
			approvalStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("251")).Background(lipgloss.Color("237")).Padding(0, 1)
			detail.WriteString("    " + approvalStyle.Render("❯ Allow   Deny   Always allow   Don't allow") + "\n")
			waitingHint := lipgloss.NewStyle().
				Foreground(theme.Warning).
				Bold(true).
				Render("    ⚡ Respond in terminal")
			return header + "\n" + detail.String() + waitingHint
		}
		return header + "\n" + detail.String()
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
