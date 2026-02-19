package copilot

import (
	"testing"
	"time"

	sdk "github.com/github/copilot-sdk/go"

	"github.com/e-9/copilot-icq/internal/domain"
)

func TestMetadataToSession(t *testing.T) {
	summary := "Fix the login bug"
	m := sdk.SessionMetadata{
		SessionID:    "abc-123",
		StartTime:    "2026-01-15T10:30:00Z",
		ModifiedTime: "2026-01-15T11:00:00Z",
		Summary:      &summary,
		Context: &sdk.SessionContext{
			Cwd: "/home/user/project",
		},
	}

	s := metadataToSession(m)

	if s.ID != "abc-123" {
		t.Errorf("ID = %q, want %q", s.ID, "abc-123")
	}
	if s.CWD != "/home/user/project" {
		t.Errorf("CWD = %q, want %q", s.CWD, "/home/user/project")
	}
	if s.Summary != "Fix the login bug" {
		t.Errorf("Summary = %q, want %q", s.Summary, "Fix the login bug")
	}
	if s.CreatedAt.IsZero() {
		t.Error("CreatedAt should not be zero")
	}
	if s.UpdatedAt.IsZero() {
		t.Error("UpdatedAt should not be zero")
	}
}

func TestMetadataToSessionNoContext(t *testing.T) {
	m := sdk.SessionMetadata{
		SessionID:    "def-456",
		StartTime:    "2026-02-01T09:00:00Z",
		ModifiedTime: "2026-02-01T09:30:00Z",
	}

	s := metadataToSession(m)

	if s.CWD != "" {
		t.Errorf("CWD = %q, want empty", s.CWD)
	}
	if s.Summary != "" {
		t.Errorf("Summary = %q, want empty", s.Summary)
	}
}

func TestSessionEventToMessage(t *testing.T) {
	content := "Hello, world!"

	tests := []struct {
		name     string
		event    sdk.SessionEvent
		wantOK   bool
		wantRole domain.MessageRole
	}{
		{
			name: "user message",
			event: sdk.SessionEvent{
				Type:      sdk.UserMessage,
				ID:        "1",
				Timestamp: time.Now(),
				Data:      sdk.Data{Content: &content},
			},
			wantOK:   true,
			wantRole: domain.RoleUser,
		},
		{
			name: "assistant message",
			event: sdk.SessionEvent{
				Type:      sdk.AssistantMessage,
				ID:        "2",
				Timestamp: time.Now(),
				Data:      sdk.Data{Content: &content},
			},
			wantOK:   true,
			wantRole: domain.RoleAssistant,
		},
		{
			name: "session idle (not a message)",
			event: sdk.SessionEvent{
				Type:      sdk.SessionIdle,
				ID:        "3",
				Timestamp: time.Now(),
			},
			wantOK: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg, ok := sessionEventToMessage(tt.event)
			if ok != tt.wantOK {
				t.Fatalf("ok = %v, want %v", ok, tt.wantOK)
			}
			if ok && msg.Role != tt.wantRole {
				t.Errorf("role = %v, want %v", msg.Role, tt.wantRole)
			}
		})
	}
}

func TestEventsToMessages(t *testing.T) {
	userContent := "What is Go?"
	assistantContent := "Go is a programming language."

	events := []sdk.SessionEvent{
		{Type: sdk.SessionStart, ID: "0", Timestamp: time.Now()},
		{Type: sdk.UserMessage, ID: "1", Timestamp: time.Now(), Data: sdk.Data{Content: &userContent}},
		{Type: sdk.AssistantMessage, ID: "2", Timestamp: time.Now(), Data: sdk.Data{Content: &assistantContent}},
		{Type: sdk.SessionIdle, ID: "3", Timestamp: time.Now()},
	}

	msgs := eventsToMessages(events)

	if len(msgs) != 2 {
		t.Fatalf("got %d messages, want 2", len(msgs))
	}
	if msgs[0].Role != domain.RoleUser {
		t.Errorf("first message role = %v, want User", msgs[0].Role)
	}
	if msgs[1].Role != domain.RoleAssistant {
		t.Errorf("second message role = %v, want Assistant", msgs[1].Role)
	}
}

func TestParseTime(t *testing.T) {
	// RFC3339
	ts := parseTime("2026-01-15T10:30:00Z")
	if ts.IsZero() {
		t.Error("should parse RFC3339")
	}

	// RFC3339Nano
	ts2 := parseTime("2026-01-15T10:30:00.123456789Z")
	if ts2.IsZero() {
		t.Error("should parse RFC3339Nano")
	}

	// Invalid
	ts3 := parseTime("not-a-time")
	if !ts3.IsZero() {
		t.Error("invalid time should return zero")
	}
}
