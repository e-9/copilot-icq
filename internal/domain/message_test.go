package domain

import (
	"encoding/json"
	"testing"
	"time"
)

func TestEventsToMessages(t *testing.T) {
	events := []Event{
		{
			Type:      EventSessionInfo,
			Data:      json.RawMessage(`{"infoType":"mcp","message":"GitHub MCP: Connected"}`),
			ID:        "e0",
			Timestamp: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			Type:      EventUserMessage,
			Data:      json.RawMessage(`{"content":"Hello copilot","transformedContent":"Hello copilot"}`),
			ID:        "e1",
			Timestamp: time.Date(2026, 1, 1, 0, 1, 0, 0, time.UTC),
		},
		{
			Type:      EventAssistantMessage,
			Data:      json.RawMessage(`{"messageId":"m1","content":"Hi there!","toolRequests":[{"toolCallId":"tc1","name":"bash","arguments":{},"type":"function"}]}`),
			ID:        "e2",
			Timestamp: time.Date(2026, 1, 1, 0, 1, 5, 0, time.UTC),
		},
		{
			Type:      EventToolExecutionComplete,
			Data:      json.RawMessage(`{"toolCallId":"tc1","success":true,"result":{"content":"done","detailedContent":"full output"}}`),
			ID:        "e3",
			Timestamp: time.Date(2026, 1, 1, 0, 1, 10, 0, time.UTC),
		},
	}

	messages := EventsToMessages(events)

	if len(messages) != 3 {
		t.Fatalf("expected 3 messages, got %d", len(messages))
	}

	// System message from session.info
	if messages[0].Role != RoleSystem {
		t.Errorf("msg[0].Role = %q, want system", messages[0].Role)
	}
	if messages[0].Content != "GitHub MCP: Connected" {
		t.Errorf("msg[0].Content = %q", messages[0].Content)
	}

	// User message
	if messages[1].Role != RoleUser {
		t.Errorf("msg[1].Role = %q, want user", messages[1].Role)
	}
	if messages[1].Content != "Hello copilot" {
		t.Errorf("msg[1].Content = %q", messages[1].Content)
	}

	// Assistant message with tool calls
	if messages[2].Role != RoleAssistant {
		t.Errorf("msg[2].Role = %q, want assistant", messages[2].Role)
	}
	if len(messages[2].ToolCalls) != 1 {
		t.Fatalf("msg[2] expected 1 tool call, got %d", len(messages[2].ToolCalls))
	}
	if messages[2].ToolCalls[0].Name != "bash" {
		t.Errorf("tool call name = %q, want bash", messages[2].ToolCalls[0].Name)
	}
}
