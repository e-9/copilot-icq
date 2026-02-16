package theme

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	// Colors
	Accent     = lipgloss.Color("34")  // Green
	Subtle     = lipgloss.Color("241") // Gray
	Highlight  = lipgloss.Color("86")  // Cyan-green
	Warning    = lipgloss.Color("214") // Orange
	Error      = lipgloss.Color("196") // Red
	Background = lipgloss.Color("235") // Dark gray

	// Sidebar styles
	SidebarWidth = 30

	SidebarStyle = lipgloss.NewStyle().
			Width(SidebarWidth).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderRight(true).
			BorderForeground(lipgloss.Color("238"))

	SelectedItemStyle = lipgloss.NewStyle().
				Foreground(Accent).
				Bold(true)

	NormalItemStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))

	DimItemStyle = lipgloss.NewStyle().
			Foreground(Subtle)

	UnreadBadgeStyle = lipgloss.NewStyle().
				Foreground(Warning).
				Bold(true)

	// Title
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(Accent).
			Padding(0, 1)

	// Status bar
	StatusBarStyle = lipgloss.NewStyle().
			Foreground(Subtle).
			Padding(0, 1)

	// Chat panel
	ChatPanelStyle = lipgloss.NewStyle().
			Padding(0, 1)

	// Section header style for sidebar groups
	SectionHeaderStyle = lipgloss.NewStyle().
				Foreground(Subtle).
				Bold(true)
)

// RenderTitledBorder renders content inside a rounded border with an inline
// title on the top edge. When focused is true the border uses Accent color.
func RenderTitledBorder(title, content string, width, height int, focused bool) string {
	border := lipgloss.RoundedBorder()
	borderColor := lipgloss.Color("238")
	titleColor := Subtle
	if focused {
		borderColor = Accent
		titleColor = Accent
	}

	bc := lipgloss.NewStyle().Foreground(borderColor)
	tc := lipgloss.NewStyle().Foreground(titleColor).Bold(true)

	// Build top border: ╭─ Title ─────────────────╮
	titleStr := ""
	if title != "" {
		titleStr = bc.Render("─ ") + tc.Render(title) + bc.Render(" ")
	}
	titleWidth := lipgloss.Width(titleStr)
	innerWidth := width
	topFill := innerWidth + 2 - titleWidth - 2 // +2 for corners, -2 for corners
	if topFill < 0 {
		topFill = 0
	}
	topBorder := bc.Render(border.TopLeft) + titleStr + bc.Render(strings.Repeat(border.Top, topFill)) + bc.Render(border.TopRight)

	// Render body with side + bottom borders only (no top)
	bodyStyle := lipgloss.NewStyle().
		Width(innerWidth).
		Height(height).
		MaxHeight(height + 1). // +1 for bottom border line
		BorderStyle(border).
		BorderTop(false).
		BorderBottom(true).
		BorderLeft(true).
		BorderRight(true).
		BorderForeground(borderColor)

	body := bodyStyle.Render(content)

	return topBorder + "\n" + body
}
