package app

import "github.com/charmbracelet/lipgloss"

func (m Model) View() string {
	if !m.ready {
		return "Loading..."
	}

	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		Render("🌸 Copilot ICQ")

	subtitle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Render("Your AI sessions, one terminal away")

	help := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Render("Press q to quit")

	return lipgloss.Place(
		m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		lipgloss.JoinVertical(lipgloss.Center, title, subtitle, "", help),
	)
}
