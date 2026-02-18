// Package copilot provides a thin adapter around the official GitHub Copilot Go SDK.
// It wraps copilot.Client to expose session discovery, resume, messaging, and streaming
// in a way that integrates cleanly with Bubble Tea via channel-based event bridging.
package copilot

import (
	"context"
	"fmt"
	"sync"
	"time"

	sdk "github.com/github/copilot-sdk/go"

	"github.com/e-9/copilot-icq/internal/domain"
)

// Adapter wraps the official Copilot SDK client and provides a simplified API
// for the TUI app layer.
type Adapter struct {
	client   *sdk.Client
	sessions map[string]*sdk.Session // sessionID → active SDK session
	mu       sync.Mutex

	// Events is a buffered channel that receives session events.
	// The app layer reads from this channel to convert events into tea.Msg.
	Events chan Event
}

// New creates a new Adapter. Call Start() to connect to the Copilot CLI.
func New() *Adapter {
	client := sdk.NewClient(&sdk.ClientOptions{
		UseStdio:    true,
		AutoStart:   sdk.Bool(true),
		AutoRestart: sdk.Bool(true),
		LogLevel:    "error",
	})
	return &Adapter{
		client:   client,
		sessions: make(map[string]*sdk.Session),
		Events:   make(chan Event, 64),
	}
}

// Start connects to the Copilot CLI subprocess.
func (a *Adapter) Start(ctx context.Context) error {
	if err := a.client.Start(ctx); err != nil {
		return fmt.Errorf("copilot SDK start: %w", err)
	}

	// Subscribe to session lifecycle events (created, deleted, updated)
	a.client.On(func(event sdk.SessionLifecycleEvent) {
		a.Events <- Event{
			Type:      EventLifecycle,
			SessionID: event.SessionID,
			Lifecycle: &event,
		}
	})

	return nil
}

// Close stops the Copilot CLI subprocess and cleans up all sessions.
func (a *Adapter) Close() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	for _, s := range a.sessions {
		s.Destroy()
	}
	a.sessions = make(map[string]*sdk.Session)

	return a.client.Stop()
}

// ListSessions returns all known sessions, mapped to domain.Session.
func (a *Adapter) ListSessions(ctx context.Context) ([]domain.Session, error) {
	metas, err := a.client.ListSessions(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("list sessions: %w", err)
	}

	sessions := make([]domain.Session, 0, len(metas))
	for _, m := range metas {
		sessions = append(sessions, metadataToSession(m))
	}
	return sessions, nil
}

// ResumeSession resumes an existing session and subscribes to its events.
// Events are forwarded to the Events channel for the app layer.
func (a *Adapter) ResumeSession(ctx context.Context, sessionID string) error {
	a.mu.Lock()
	if _, ok := a.sessions[sessionID]; ok {
		a.mu.Unlock()
		return nil // already resumed
	}
	a.mu.Unlock()

	session, err := a.client.ResumeSessionWithOptions(ctx, sessionID, &sdk.ResumeSessionConfig{
		Streaming: true,
		OnPermissionRequest: a.makePermissionHandler(sessionID),
		OnUserInputRequest:  a.makeUserInputHandler(sessionID),
	})
	if err != nil {
		return fmt.Errorf("resume session %s: %w", sessionID, err)
	}

	a.mu.Lock()
	a.sessions[sessionID] = session
	a.mu.Unlock()

	// Subscribe to session events and forward to Events channel
	session.On(func(event sdk.SessionEvent) {
		a.Events <- Event{
			Type:         EventSession,
			SessionID:    sessionID,
			SessionEvent: &event,
		}
	})

	return nil
}

// GetHistory retrieves conversation history for a session, mapped to domain.Message.
func (a *Adapter) GetHistory(ctx context.Context, sessionID string) ([]domain.Message, error) {
	a.mu.Lock()
	session, ok := a.sessions[sessionID]
	a.mu.Unlock()

	if !ok {
		return nil, fmt.Errorf("session %s not resumed", sessionID)
	}

	events, err := session.GetMessages(ctx)
	if err != nil {
		return nil, fmt.Errorf("get messages for %s: %w", sessionID, err)
	}

	return eventsToMessages(events), nil
}

// Send sends a message to a resumed session.
func (a *Adapter) Send(ctx context.Context, sessionID, text string) (string, error) {
	a.mu.Lock()
	session, ok := a.sessions[sessionID]
	a.mu.Unlock()

	if !ok {
		return "", fmt.Errorf("session %s not resumed", sessionID)
	}

	msgID, err := session.Send(ctx, sdk.MessageOptions{
		Prompt: text,
	})
	if err != nil {
		return "", fmt.Errorf("send to %s: %w", sessionID, err)
	}
	return msgID, nil
}

// Abort cancels the currently processing message in a session.
func (a *Adapter) Abort(ctx context.Context, sessionID string) error {
	a.mu.Lock()
	session, ok := a.sessions[sessionID]
	a.mu.Unlock()

	if !ok {
		return fmt.Errorf("session %s not resumed", sessionID)
	}

	return session.Abort(ctx)
}

// IsResumed returns whether a session has been resumed.
func (a *Adapter) IsResumed(sessionID string) bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	_, ok := a.sessions[sessionID]
	return ok
}


// makePermissionHandler creates a permission handler that routes requests through the Events channel.
func (a *Adapter) makePermissionHandler(sessionID string) sdk.PermissionHandler {
	return func(req sdk.PermissionRequest, inv sdk.PermissionInvocation) (sdk.PermissionRequestResult, error) {
		respCh := make(chan PermissionResponse, 1)

		toolName := ""
		if name, ok := req.Extra["toolName"]; ok {
			toolName, _ = name.(string)
		}
		action := req.Kind

		a.Events <- Event{
			Type:      EventPermission,
			SessionID: sessionID,
			Permission: &PermissionEvent{
				ToolName: toolName,
				Action:   action,
				Response: respCh,
			},
		}

		// Block until the TUI user responds
		resp := <-respCh
		if resp.Allow {
			return sdk.PermissionRequestResult{Kind: "allow"}, nil
		}
		return sdk.PermissionRequestResult{Kind: "deny"}, nil
	}
}

// makeUserInputHandler creates a user input handler that routes requests through the Events channel.
func (a *Adapter) makeUserInputHandler(sessionID string) sdk.UserInputHandler {
	return func(req sdk.UserInputRequest, inv sdk.UserInputInvocation) (sdk.UserInputResponse, error) {
		respCh := make(chan UserInputResponse, 1)

		a.Events <- Event{
			Type:      EventUserInput,
			SessionID: sessionID,
			UserInput: &UserInputEvent{
				Question:      req.Question,
				Choices:       req.Choices,
				AllowFreeform: req.AllowFreeform,
				Response:      respCh,
			},
		}

		// Block until the TUI user responds
		resp := <-respCh
		return sdk.UserInputResponse{
			Answer:      resp.Answer,
			WasFreeform: resp.WasFreeform,
		}, nil
	}
}

// metadataToSession converts SDK SessionMetadata to our domain.Session.
func metadataToSession(m sdk.SessionMetadata) domain.Session {
	cwd := ""
	if m.Context != nil {
		cwd = m.Context.Cwd
	}

	summary := ""
	if m.Summary != nil {
		summary = *m.Summary
	}

	createdAt := parseTime(m.StartTime)
	updatedAt := parseTime(m.ModifiedTime)

	return domain.Session{
		ID:        m.SessionID,
		CWD:       cwd,
		Summary:   summary,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}
}

// eventsToMessages converts SDK SessionEvents to domain.Messages.
func eventsToMessages(events []sdk.SessionEvent) []domain.Message {
	var messages []domain.Message
	for _, e := range events {
		msg, ok := sessionEventToMessage(e)
		if ok {
			messages = append(messages, msg)
		}
	}
	return messages
}

// sessionEventToMessage converts a single SDK SessionEvent to a domain.Message.
// Returns false if the event type doesn't map to a displayable message.
func sessionEventToMessage(e sdk.SessionEvent) (domain.Message, bool) {
	switch e.Type {
	case sdk.UserMessage:
		content := ""
		if e.Data.Content != nil {
			content = *e.Data.Content
		}
		return domain.Message{
			Role:    domain.RoleUser,
			Content: content,
		}, true

	case sdk.AssistantMessage:
		content := ""
		if e.Data.Content != nil {
			content = *e.Data.Content
		}
		return domain.Message{
			Role:    domain.RoleAssistant,
			Content: content,
		}, true

	case sdk.ToolExecutionComplete:
		toolName := ""
		if e.Data.ToolName != nil {
			toolName = *e.Data.ToolName
		}
		resultContent := ""
		if e.Data.Result != nil {
			resultContent = e.Data.Result.Content
		}
		tc := domain.ToolCall{
			Name:    toolName,
			Summary: resultContent,
			Status:  domain.ToolCallComplete,
		}
		return domain.Message{
			Role:      domain.RoleAssistant,
			ToolCalls: []domain.ToolCall{tc},
		}, true

	default:
		return domain.Message{}, false
	}
}

func parseTime(s string) time.Time {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		t, err = time.Parse(time.RFC3339Nano, s)
		if err != nil {
			return time.Time{}
		}
	}
	return t
}
