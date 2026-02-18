package copilot

import sdk "github.com/github/copilot-sdk/go"

// EventType classifies the kind of event emitted by the adapter.
type EventType int

const (
	// EventSession wraps a SDK SessionEvent (message, tool, idle, error, etc.)
	EventSession EventType = iota
	// EventLifecycle wraps a SDK SessionLifecycleEvent (created, deleted, updated)
	EventLifecycle
)

// Event is emitted by the Adapter to the app layer via the Events channel.
// Exactly one of SessionEvent or Lifecycle will be non-nil.
type Event struct {
	Type         EventType
	SessionID    string
	SessionEvent *sdk.SessionEvent
	Lifecycle    *sdk.SessionLifecycleEvent
}
