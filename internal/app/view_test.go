package app

import (
	"fmt"
	"os"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/e-9/copilot-icq/internal/infra/runner"
	"github.com/e-9/copilot-icq/internal/infra/sessionrepo"
	"github.com/e-9/copilot-icq/internal/infra/watcher"
)

func TestViewFitsTerminalHeight(t *testing.T) {
	tmp := t.TempDir()
	repo := sessionrepo.New(tmp)
	w, err := watcher.New(tmp)
	if err != nil {
		t.Fatal(err)
	}
	defer w.Close()
	r := runner.New("copilot", runner.ModeScoped)

	m := NewModel(repo, w, r, nil, nil)

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
	tmp := t.TempDir()
	// Create 10 sessions to ensure sidebar list content would overflow
	for i := 0; i < 10; i++ {
		sessionDir := fmt.Sprintf("%s/session-%d", tmp, i)
		os.MkdirAll(sessionDir, 0o755)
		yaml := fmt.Sprintf("id: session-%d\ncwd: /tmp/project-%d\nsummary: Session number %d with a long name\nsummary_count: 1\ncreated_at: 2026-01-01T00:00:00Z\nupdated_at: 2026-01-01T00:00:00Z\n", i, i, i)
		os.WriteFile(sessionDir+"/workspace.yaml", []byte(yaml), 0o644)
	}

	repo := sessionrepo.New(tmp)
	w, err := watcher.New(tmp)
	if err != nil {
		t.Fatal(err)
	}
	defer w.Close()
	r := runner.New("copilot", runner.ModeScoped)

	m := NewModel(repo, w, r, nil, nil)

	// Small terminal — 10 sessions * 2 lines each = 20 lines, way more than 24-row terminal
	model, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m = model.(Model)

	sessions, _ := repo.List()
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
	tmp := t.TempDir()
	sessionDir := tmp + "/test-session"
	os.MkdirAll(sessionDir, 0o755)
	os.WriteFile(sessionDir+"/workspace.yaml", []byte("id: test-session\ncwd: /tmp\nsummary: Test\nsummary_count: 1\ncreated_at: 2026-01-01T00:00:00Z\nupdated_at: 2026-01-01T00:00:00Z\n"), 0o644)

	repo := sessionrepo.New(tmp)
	w, err := watcher.New(tmp)
	if err != nil {
		t.Fatal(err)
	}
	defer w.Close()
	r := runner.New("copilot", runner.ModeScoped)

	m := NewModel(repo, w, r, nil, nil)

	model, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	m = model.(Model)

	sessions, _ := repo.List()
	model, _ = m.Update(SessionsLoadedMsg{Sessions: sessions})
	m = model.(Model)

	output := m.View()
	renderedH := lipgloss.Height(output)

	if renderedH > 40 {
		t.Errorf("with sessions: rendered height %d exceeds terminal height 40 (overflow by %d rows)",
			renderedH, renderedH-40)
	}
}
