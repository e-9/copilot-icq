package notifier

import (
	tea "github.com/charmbracelet/bubbletea"
)

// HookEventMsg is sent to the Bubble Tea program when a hook event arrives.
type HookEventMsg struct {
	SessionID string
	Event     string
	Title     string
	Body      string
}

// TUINotifier sends notifications into the Bubble Tea event loop.
type TUINotifier struct {
	program *tea.Program
}

// NewTUI creates a TUI notifier that sends messages to the given program.
func NewTUI(p *tea.Program) *TUINotifier {
	return &TUINotifier{program: p}
}

// Notify sends a HookEventMsg into the Bubble Tea event loop.
func (t *TUINotifier) Notify(n Notification) error {
	if t.program != nil {
		t.program.Send(HookEventMsg{
			SessionID: n.SessionID,
			Event:     n.Event,
			Title:     n.Title,
			Body:      n.Body,
		})
	}
	return nil
}
