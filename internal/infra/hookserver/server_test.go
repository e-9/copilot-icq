package hookserver

import (
	"encoding/json"
	"net"
	"testing"
	"time"
)

func TestServerReceivesEvent(t *testing.T) {
	sock := t.TempDir() + "/test.sock"
	srv, err := New(sock)
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}
	defer srv.Close()
	go srv.Start()

	// Give server a moment to start
	time.Sleep(50 * time.Millisecond)

	// Connect and send a hook event
	conn, err := net.Dial("unix", sock)
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}

	evt := HookEvent{
		Event:     "postToolUse",
		SessionID: "abc-123",
		CWD:       "/tmp/test",
		Timestamp: "2026-01-01T00:00:00Z",
	}
	data, _ := json.Marshal(evt)
	data = append(data, '\n')
	conn.Write(data)
	conn.Close()

	select {
	case received := <-srv.Events():
		if received.Event != "postToolUse" {
			t.Errorf("expected event 'postToolUse', got %q", received.Event)
		}
		if received.SessionID != "abc-123" {
			t.Errorf("expected sessionId 'abc-123', got %q", received.SessionID)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for event")
	}
}

func TestServerMultipleEvents(t *testing.T) {
	sock := t.TempDir() + "/test.sock"
	srv, err := New(sock)
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}
	defer srv.Close()
	go srv.Start()

	time.Sleep(50 * time.Millisecond)

	conn, err := net.Dial("unix", sock)
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}

	events := []string{"sessionStart", "postToolUse", "sessionEnd"}
	for _, e := range events {
		evt := HookEvent{Event: e, SessionID: "test-session"}
		data, _ := json.Marshal(evt)
		data = append(data, '\n')
		conn.Write(data)
	}
	conn.Close()

	for _, expected := range events {
		select {
		case received := <-srv.Events():
			if received.Event != expected {
				t.Errorf("expected %q, got %q", expected, received.Event)
			}
		case <-time.After(2 * time.Second):
			t.Fatalf("timed out waiting for event %q", expected)
		}
	}
}
