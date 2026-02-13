package sessionrepo

import (
	"os"
	"path/filepath"
	"testing"
)

func TestListDiscoversSessions(t *testing.T) {
	// Create a temp session-state directory with a fake session
	tmp := t.TempDir()
	sessionDir := filepath.Join(tmp, "test-session-id")
	if err := os.MkdirAll(sessionDir, 0o755); err != nil {
		t.Fatal(err)
	}

	workspace := `id: test-session-id
cwd: /tmp/test
summary: Test Session
summary_count: 1
created_at: 2026-01-01T00:00:00Z
updated_at: 2026-01-02T00:00:00Z
`
	if err := os.WriteFile(filepath.Join(sessionDir, "workspace.yaml"), []byte(workspace), 0o644); err != nil {
		t.Fatal(err)
	}

	repo := New(tmp)
	sessions, err := repo.List()
	if err != nil {
		t.Fatalf("List() error: %v", err)
	}

	if len(sessions) != 1 {
		t.Fatalf("expected 1 session, got %d", len(sessions))
	}

	s := sessions[0]
	if s.ID != "test-session-id" {
		t.Errorf("ID = %q, want %q", s.ID, "test-session-id")
	}
	if s.Summary != "Test Session" {
		t.Errorf("Summary = %q, want %q", s.Summary, "Test Session")
	}
	if s.CWD != "/tmp/test" {
		t.Errorf("CWD = %q, want %q", s.CWD, "/tmp/test")
	}
}

func TestListSkipsInvalidSessions(t *testing.T) {
	tmp := t.TempDir()

	// Create a directory without workspace.yaml
	if err := os.MkdirAll(filepath.Join(tmp, "bad-session"), 0o755); err != nil {
		t.Fatal(err)
	}

	repo := New(tmp)
	sessions, err := repo.List()
	if err != nil {
		t.Fatalf("List() error: %v", err)
	}

	if len(sessions) != 0 {
		t.Fatalf("expected 0 sessions, got %d", len(sessions))
	}
}
