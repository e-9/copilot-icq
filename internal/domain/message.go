package domain

import "time"

// MessageRole indicates who sent the message.
type MessageRole string

const (
RoleUser      MessageRole = "user"
RoleAssistant MessageRole = "assistant"
RoleTool      MessageRole = "tool"
RoleSystem    MessageRole = "system"
)

// Message is a display-ready chat message.
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
Command  string   // for bash tools, the command being run
Question string   // for ask_user tools, the question text
Choices  []string // for ask_user tools, the selectable options
FilePath string   // for edit/create tools, the target file
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
