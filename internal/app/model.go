package app

import tea "github.com/charmbracelet/bubbletea"

// Model is the root application state following the Elm Architecture.
type Model struct {
	width  int
	height int
	ready  bool
}

// NewModel creates the initial application model.
func NewModel() Model {
	return Model{}
}

func (m Model) Init() tea.Cmd {
	return nil
}
