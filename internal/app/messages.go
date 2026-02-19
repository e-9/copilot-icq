package app

import (
"github.com/e-9/copilot-icq/internal/copilot"
"github.com/e-9/copilot-icq/internal/domain"
)

// SessionsLoadedMsg is sent when sessions are discovered.
type SessionsLoadedMsg struct {
Sessions []domain.Session
Err      error
}

// EventsLoadedMsg is sent when a session's conversation history is loaded.
type EventsLoadedMsg struct {
SessionID string
Messages  []domain.Message
Err       error
}

// MessageSentMsg is returned when a message has been dispatched.
type MessageSentMsg struct {
SessionID string
Success   bool
Output    string
Err       error
}

// TickMsg is sent periodically to trigger session rescans.
type TickMsg struct{}

// SessionRenamedMsg is sent when a session has been renamed.
type SessionRenamedMsg struct {
SessionID string
Err       error
}

// ExportCompleteMsg is sent when a conversation export finishes.
type ExportCompleteMsg struct {
Path string
Err  error
}

// ClearFlashMsg clears the transient status bar message.
type ClearFlashMsg struct{}

// --- SDK messages ---

// SDKConnectedMsg is sent when the SDK adapter connects to Copilot CLI.
type SDKConnectedMsg struct {
Err error
}

// SDKSessionResumedMsg is sent when a session has been resumed via the SDK.
type SDKSessionResumedMsg struct {
SessionID string
Err       error
}

// SDKEventMsg wraps an event from the SDK adapter's Events channel.
type SDKEventMsg struct {
Event copilot.Event
}

// SDKDisconnectedMsg is sent when the SDK adapter's Events channel closes.
type SDKDisconnectedMsg struct{}
