package copilot

import sdk "github.com/github/copilot-sdk/go"

// EventType classifies the kind of event emitted by the adapter.
type EventType int

const (
	// EventSession wraps a SDK SessionEvent (message, tool, idle, error, etc.)
	EventSession EventType = iota
	// EventLifecycle wraps a SDK SessionLifecycleEvent (created, deleted, updated)
	EventLifecycle
	// EventPermission is a permission request from the agent
	EventPermission
	// EventUserInput is a user input request from the agent (ask_user)
	EventUserInput
)

// Event is emitted by the Adapter to the app layer via the Events channel.
// Exactly one of the pointer fields will be non-nil, depending on Type.
type Event struct {
	Type         EventType
	SessionID    string
	SessionEvent *sdk.SessionEvent
	Lifecycle    *sdk.SessionLifecycleEvent
	Permission   *PermissionEvent
	UserInput    *UserInputEvent
}

// PermissionEvent wraps a tool permission request with a response channel.
type PermissionEvent struct {
	ToolName string
	Action   string
	Response chan<- PermissionResponse
}

// PermissionResponse is the user's decision on a permission request.
type PermissionResponse struct {
	Allow bool
}

// UserInputEvent wraps an ask_user request with a response channel.
type UserInputEvent struct {
	Question      string
	Choices       []string
	AllowFreeform bool
	Response      chan<- UserInputResponse
}

// UserInputResponse is the user's answer to an ask_user question.
type UserInputResponse struct {
	Answer      string
	WasFreeform bool
}
