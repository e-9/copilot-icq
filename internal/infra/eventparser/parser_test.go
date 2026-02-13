package eventparser

import (
	"os"
	"path/filepath"
	"testing"
)

const sampleEvents = `{"type":"session.start","data":{"sessionId":"test-123"},"id":"e1","timestamp":"2026-01-01T00:00:00Z","parentId":null}
{"type":"user.message","data":{"content":"Hello copilot","transformedContent":"Hello copilot"},"id":"e2","timestamp":"2026-01-01T00:01:00Z","parentId":"e1"}
{"type":"assistant.message","data":{"messageId":"m1","content":"Hi there!","toolRequests":[]},"id":"e3","timestamp":"2026-01-01T00:01:05Z","parentId":"e2"}
`

func TestReadAll(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "events.jsonl")
	if err := os.WriteFile(path, []byte(sampleEvents), 0o644); err != nil {
		t.Fatal(err)
	}

	p := New(path)
	events, err := p.ReadAll()
	if err != nil {
		t.Fatalf("ReadAll() error: %v", err)
	}

	if len(events) != 3 {
		t.Fatalf("expected 3 events, got %d", len(events))
	}

	if events[0].Type != "session.start" {
		t.Errorf("event[0].Type = %q, want session.start", events[0].Type)
	}
	if events[1].Type != "user.message" {
		t.Errorf("event[1].Type = %q, want user.message", events[1].Type)
	}
	if events[2].Type != "assistant.message" {
		t.Errorf("event[2].Type = %q, want assistant.message", events[2].Type)
	}
}

func TestReadNewIncremental(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "events.jsonl")
	if err := os.WriteFile(path, []byte(sampleEvents), 0o644); err != nil {
		t.Fatal(err)
	}

	p := New(path)
	events, _ := p.ReadAll()
	if len(events) != 3 {
		t.Fatalf("initial read: expected 3 events, got %d", len(events))
	}

	// Append a new event
	f, _ := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0o644)
	f.WriteString(`{"type":"user.message","data":{"content":"follow up"},"id":"e4","timestamp":"2026-01-01T00:02:00Z","parentId":"e3"}` + "\n")
	f.Close()

	newEvents, err := p.ReadNew()
	if err != nil {
		t.Fatalf("ReadNew() error: %v", err)
	}
	if len(newEvents) != 1 {
		t.Fatalf("incremental read: expected 1 event, got %d", len(newEvents))
	}
	if newEvents[0].Type != "user.message" {
		t.Errorf("new event type = %q, want user.message", newEvents[0].Type)
	}
}
