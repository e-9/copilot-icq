package sidebar

import (
	"testing"
	"time"

	"github.com/e-9/copilot-icq/internal/domain"
)

func TestSortSessions(t *testing.T) {
	now := time.Now()

	sessions := []domain.Session{
		{ID: "idle-old", Summary: "Idle Old", UpdatedAt: now.Add(-1 * time.Hour)},
		{ID: "unread", Summary: "Has Unread", UpdatedAt: now.Add(-30 * time.Minute)},
		{ID: "active", Summary: "Active Session", UpdatedAt: now.Add(-2 * time.Hour)},
		{ID: "idle-new", Summary: "Idle New", UpdatedAt: now.Add(-5 * time.Minute)},
	}

	m := New(nil, 30, 20)
	m.SetActiveID("active")
	m.delegate.Unread = map[string]int{"unread": 3}
	m.delegate.LastSeen = map[string]time.Time{
		"unread": now.Add(-1 * time.Minute),
	}

	sorted := m.sortSessions(sessions)

	// Active session should be first
	if sorted[0].ID != "active" {
		t.Errorf("expected active session first, got %q", sorted[0].ID)
	}

	// Unread session should be second
	if sorted[1].ID != "unread" {
		t.Errorf("expected unread session second, got %q", sorted[1].ID)
	}

	// Idle sessions sorted by most recent
	if sorted[2].ID != "idle-new" {
		t.Errorf("expected idle-new third, got %q", sorted[2].ID)
	}
	if sorted[3].ID != "idle-old" {
		t.Errorf("expected idle-old fourth, got %q", sorted[3].ID)
	}
}

func TestSortPreservesCursor(t *testing.T) {
	sessions := []domain.Session{
		{ID: "a", Summary: "A", UpdatedAt: time.Now().Add(-2 * time.Hour)},
		{ID: "b", Summary: "B", UpdatedAt: time.Now().Add(-1 * time.Hour)},
		{ID: "c", Summary: "C", UpdatedAt: time.Now()},
	}

	m := New(sessions, 30, 20)

	// Select session "b" (index 1)
	m.List.Select(1)
	sel := m.SelectedSession()
	if sel == nil || sel.ID != "b" {
		t.Fatalf("expected selection on 'b', got %v", sel)
	}

	// Now set "c" as active — this will re-sort, moving "c" to top
	m.SetActiveID("c")
	m.SetItems(sessions)

	// Cursor should still be on "b"
	sel = m.SelectedSession()
	if sel == nil || sel.ID != "b" {
		t.Errorf("cursor should stay on 'b' after re-sort, got %q", sel.ID)
	}
}
