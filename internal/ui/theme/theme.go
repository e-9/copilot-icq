package theme

import "github.com/charmbracelet/lipgloss"

var (
	// Colors
	Accent     = lipgloss.Color("205") // Pink
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
)
