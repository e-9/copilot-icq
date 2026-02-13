package domain

import (
	"encoding/json"
	"time"
)

// EventType represents the type of a Copilot CLI event.
type EventType string

const (
	EventSessionStart        EventType = "session.start"
	EventSessionInfo         EventType = "session.info"
	EventUserMessage         EventType = "user.message"
	EventAssistantMessage    EventType = "assistant.message"
	EventAssistantTurnStart  EventType = "assistant.turn_start"
	EventAssistantTurnEnd    EventType = "assistant.turn_end"
	EventToolExecutionStart  EventType = "tool.execution_start"
	EventToolExecutionComplete EventType = "tool.execution_complete"
)

// Event is a raw event from events.jsonl.
type Event struct {
	Type      EventType       `json:"type"`
	Data      json.RawMessage `json:"data"`
	ID        string          `json:"id"`
	Timestamp time.Time       `json:"timestamp"`
	ParentID  *string         `json:"parentId"`
}

// UserMessageData holds the payload for user.message events.
type UserMessageData struct {
	Content            string `json:"content"`
	TransformedContent string `json:"transformedContent"`
}

// ToolRequest represents a tool call requested by the assistant.
type ToolRequest struct {
	ToolCallID string          `json:"toolCallId"`
	Name       string          `json:"name"`
	Arguments  json.RawMessage `json:"arguments"`
	Type       string          `json:"type"`
}

// AssistantMessageData holds the payload for assistant.message events.
type AssistantMessageData struct {
	MessageID    string        `json:"messageId"`
	Content      string        `json:"content"`
	ToolRequests []ToolRequest `json:"toolRequests"`
}

// ToolExecutionData holds the payload for tool execution events.
type ToolExecutionData struct {
	ToolCallID string `json:"toolCallId"`
	ToolName   string `json:"toolName"`
	Success    bool   `json:"success"`
	Result     *struct {
		Content         string `json:"content"`
		DetailedContent string `json:"detailedContent"`
	} `json:"result"`
}

// ParseUserMessage parses the data field of a user.message event.
func (e Event) ParseUserMessage() (UserMessageData, error) {
	var d UserMessageData
	err := json.Unmarshal(e.Data, &d)
	return d, err
}

// ParseAssistantMessage parses the data field of an assistant.message event.
func (e Event) ParseAssistantMessage() (AssistantMessageData, error) {
	var d AssistantMessageData
	err := json.Unmarshal(e.Data, &d)
	return d, err
}

// ParseToolExecution parses the data field of tool execution events.
func (e Event) ParseToolExecution() (ToolExecutionData, error) {
	var d ToolExecutionData
	err := json.Unmarshal(e.Data, &d)
	return d, err
}
