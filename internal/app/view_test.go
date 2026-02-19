package app

import (
"fmt"
"testing"
"time"

tea "github.com/charmbracelet/bubbletea"
"github.com/charmbracelet/lipgloss"
"github.com/e-9/copilot-icq/internal/domain"
)

func TestViewFitsTerminalHeight(t *testing.T) {
m := NewModel("", nil, nil)

sizes := []struct {
w, h int
}{
{80, 24},
{120, 40},
{200, 50},
{60, 15},
}

for _, sz := range sizes {
msg := tea.WindowSizeMsg{Width: sz.w, Height: sz.h}
model, _ := m.Update(msg)
m = model.(Model)

output := m.View()
renderedH := lipgloss.Height(output)

if renderedH > sz.h {
t.Errorf("at %dx%d: rendered height %d exceeds terminal height %d (overflow by %d rows)",
sz.w, sz.h, renderedH, sz.h, renderedH-sz.h)
}
}
}

func TestViewFitsWithManySessions(t *testing.T) {
m := NewModel("", nil, nil)

model, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
m = model.(Model)

var sessions []domain.Session
for i := 0; i < 10; i++ {
sessions = append(sessions, domain.Session{
ID:        fmt.Sprintf("session-%d", i),
CWD:       fmt.Sprintf("/tmp/project-%d", i),
Summary:   fmt.Sprintf("Session number %d with a long name", i),
CreatedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
UpdatedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
})
}

model, _ = m.Update(SessionsLoadedMsg{Sessions: sessions})
m = model.(Model)

output := m.View()
renderedH := lipgloss.Height(output)

if renderedH > 24 {
t.Errorf("with 10 sessions at 80x24: rendered height %d exceeds terminal height 24 (overflow by %d rows)",
renderedH, renderedH-24)
}
}

func TestViewFitsWithSessions(t *testing.T) {
m := NewModel("", nil, nil)

model, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
m = model.(Model)

sessions := []domain.Session{{
ID:        "test-session",
CWD:       "/tmp",
Summary:   "Test",
CreatedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
UpdatedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
}}

model, _ = m.Update(SessionsLoadedMsg{Sessions: sessions})
m = model.(Model)

output := m.View()
renderedH := lipgloss.Height(output)

if renderedH > 40 {
t.Errorf("with sessions: rendered height %d exceeds terminal height 40 (overflow by %d rows)",
renderedH, renderedH-40)
}
}
