package input

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/e-9/copilot-icq/internal/ui/theme"
)

// Model represents the message input area at the bottom of the chat panel.
type Model struct {
	textInput textinput.Model
	width     int
	sending   bool
}

// New creates a new input model.
func New(width int) Model {
	ti := textinput.New()
	ti.Placeholder = "Type a message... (Enter to send)"
	ti.CharLimit = 4096
	ti.Width = width - 4
	ti.Prompt = "❯ "
	ti.PromptStyle = lipgloss.NewStyle().Foreground(theme.Accent)

	return Model{
		textInput: ti,
		width:     width,
	}
}

// Focus gives keyboard focus to the input.
func (m *Model) Focus() {
	m.textInput.Focus()
}

// Blur removes keyboard focus.
func (m *Model) Blur() {
	m.textInput.Blur()
}

// SetWidth updates the input width.
func (m *Model) SetWidth(w int) {
	m.width = w
	m.textInput.Width = w - 4
}

// Value returns the current input text.
func (m Model) Value() string {
	return m.textInput.Value()
}

// Reset clears the input.
func (m *Model) Reset() {
	m.textInput.Reset()
}

// SetSending toggles the sending state.
func (m *Model) SetSending(sending bool) {
	m.sending = sending
	if sending {
		m.textInput.Placeholder = "Sending..."
		m.Blur()
	} else {
		m.textInput.Placeholder = "Type a message... (Enter to send)"
		m.Focus()
	}
}

// SetRenaming puts the input into rename mode with prefilled text.
func (m *Model) SetRenaming(currentName string) {
	m.textInput.Placeholder = "Enter new session name..."
	m.textInput.SetValue(currentName)
	m.textInput.CursorEnd()
	m.Focus()
}

// ClearRenaming restores the input to normal message mode.
func (m *Model) ClearRenaming() {
	m.textInput.Placeholder = "Type a message... (Enter to send)"
}

// IsSending returns whether a message is being sent.
func (m Model) IsSending() bool {
	return m.sending
}

// Update handles textinput messages.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

// View renders the input area.
func (m Model) View() string {
	if m.sending {
		spinner := lipgloss.NewStyle().
			Foreground(theme.Warning).
			Bold(true).
			Render("⏳ Sending message to Copilot...")
		return spinner
	}
	return m.textInput.View()
}
