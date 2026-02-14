package domain

import (
	"encoding/json"
	"strings"
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
	Name     string
	Status   ToolCallStatus
	Summary  string
	Command  string   // for bash tools, the command being run/requested
	Question string   // for ask_user tools, the question text
	Choices  []string // for ask_user tools, the selectable options
	FilePath string   // for edit/create/apply_patch tools, the target file
	Patch    string   // for edit/apply_patch tools, the diff content
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
				// Extract question/choices for ask_user tools
				if tr.Name == "ask_user" {
					var args struct {
						Question string   `json:"question"`
						Choices  []string `json:"choices"`
					}
					if err := parseJSON(tr.Arguments, &args); err == nil {
						tc.Question = args.Question
						tc.Choices = args.Choices
					}
				}
				// Extract file path for edit/create tools
				if tr.Name == "edit" || tr.Name == "create" {
					var args struct {
						Path   string `json:"path"`
						OldStr string `json:"old_str"`
						NewStr string `json:"new_str"`
					}
					if err := parseJSON(tr.Arguments, &args); err == nil {
						tc.FilePath = args.Path
						if args.OldStr != "" || args.NewStr != "" {
							tc.Patch = formatEditDiff(args.OldStr, args.NewStr)
						}
					}
				}
				// Extract file paths from apply_patch (Codex/GPT format)
				if tr.Name == "apply_patch" {
					patch := string(tr.Arguments)
					// Arguments is a raw string, not JSON object
					if len(patch) > 0 && patch[0] == '"' {
						var s string
						if err := parseJSON(tr.Arguments, &s); err == nil {
							patch = s
						}
					}
					tc.Patch = patch
					tc.FilePath = extractPatchFiles(patch)
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

// formatEditDiff creates a simple diff representation from old/new strings.
func formatEditDiff(oldStr, newStr string) string {
	var sb strings.Builder
	if oldStr != "" {
		for _, line := range strings.Split(oldStr, "\n") {
			sb.WriteString("- " + line + "\n")
		}
	}
	if newStr != "" {
		for _, line := range strings.Split(newStr, "\n") {
			sb.WriteString("+ " + line + "\n")
		}
	}
	return sb.String()
}

// extractPatchFiles extracts file paths from an apply_patch unified diff.
func extractPatchFiles(patch string) string {
	var files []string
	for _, line := range strings.Split(patch, "\n") {
		trimmed := strings.TrimSpace(line)
		for _, prefix := range []string{"*** Update File: ", "*** Add File: ", "*** Delete File: "} {
			if strings.HasPrefix(trimmed, prefix) {
				files = append(files, strings.TrimPrefix(trimmed, prefix))
			}
		}
	}
	return strings.Join(files, ", ")
}
