package domain

import (
	"encoding/json"
	"time"
)

// MessageRole indicates who sent the message.
type MessageRole string

const (
	RoleUser      MessageRole = "user"
	RoleAssistant MessageRole = "assistant"
	RoleTool      MessageRole = "tool"
	RoleSystem    MessageRole = "system"
)

// Message is a display-ready chat message derived from events.
type Message struct {
	Role      MessageRole
	Content   string
	Timestamp time.Time
	ToolCalls []ToolCall
}

// ToolCall represents a tool invocation shown in the chat.
type ToolCall struct {
	Name    string
	Status  ToolCallStatus
	Summary string
	Command string // for bash tools, the command being run/requested
}

// ToolCallStatus tracks the state of a tool invocation.
type ToolCallStatus string

const (
	ToolCallPending  ToolCallStatus = "pending"
	ToolCallRunning  ToolCallStatus = "running"
	ToolCallComplete ToolCallStatus = "complete"
	ToolCallFailed   ToolCallStatus = "failed"
)

// EventsToMessages converts a slice of raw events into display-ready messages.
// It merges related events (e.g., tool starts/completes) into coherent messages.
func EventsToMessages(events []Event) []Message {
	var messages []Message
	// Track tool calls by pointer so status updates propagate to messages
	toolStatus := make(map[string]*ToolCall)    // toolCallId → ToolCall ptr
	toolMsgIdx := make(map[string]int)          // toolCallId → message index
	toolCallIdx := make(map[string]int)         // toolCallId → index within message's ToolCalls

	for _, evt := range events {
		switch evt.Type {
		case EventUserMessage:
			d, err := evt.ParseUserMessage()
			if err != nil {
				continue
			}
			messages = append(messages, Message{
				Role:      RoleUser,
				Content:   d.Content,
				Timestamp: evt.Timestamp,
			})

		case EventAssistantMessage:
			d, err := evt.ParseAssistantMessage()
			if err != nil {
				continue
			}
			var toolCalls []ToolCall
			for _, tr := range d.ToolRequests {
				tc := ToolCall{
					Name:   tr.Name,
					Status: ToolCallPending,
				}
				// Extract command for bash tools
				if tr.Name == "bash" {
					var args struct {
						Command string `json:"command"`
					}
					if err := parseJSON(tr.Arguments, &args); err == nil {
						tc.Command = args.Command
					}
				}
				toolCalls = append(toolCalls, tc)
				msgIdx := len(messages) // will be this message's index
				tcIdx := len(toolCalls) - 1
				toolStatus[tr.ToolCallID] = &toolCalls[tcIdx]
				toolMsgIdx[tr.ToolCallID] = msgIdx
				toolCallIdx[tr.ToolCallID] = tcIdx
			}
			// Only add if there's content or tool calls
			if d.Content != "" || len(toolCalls) > 0 {
				messages = append(messages, Message{
					Role:      RoleAssistant,
					Content:   d.Content,
					Timestamp: evt.Timestamp,
					ToolCalls: toolCalls,
				})
				// Re-point toolStatus to the actual slice elements in messages
				for _, tr := range d.ToolRequests {
					mi := toolMsgIdx[tr.ToolCallID]
					ci := toolCallIdx[tr.ToolCallID]
					toolStatus[tr.ToolCallID] = &messages[mi].ToolCalls[ci]
				}
			}

		case EventToolExecutionStart:
			d, err := evt.ParseToolExecution()
			if err != nil {
				continue
			}
			if tc, ok := toolStatus[d.ToolCallID]; ok {
				tc.Status = ToolCallRunning
			}

		case EventToolExecutionComplete:
			d, err := evt.ParseToolExecution()
			if err != nil {
				continue
			}
			if tc, ok := toolStatus[d.ToolCallID]; ok {
				if d.Success {
					tc.Status = ToolCallComplete
				} else {
					tc.Status = ToolCallFailed
				}
				if d.Result != nil {
					tc.Summary = d.Result.Content
				}
			}

		case EventSessionInfo:
			// Show session info as system messages
			var info struct {
				Message string `json:"message"`
			}
			if err := parseJSON(evt.Data, &info); err == nil && info.Message != "" {
				messages = append(messages, Message{
					Role:      RoleSystem,
					Content:   info.Message,
					Timestamp: evt.Timestamp,
				})
			}
		}
	}

	return messages
}

func parseJSON(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}
