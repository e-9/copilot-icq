package app

import (
"context"
"time"

tea "github.com/charmbracelet/bubbletea"
"github.com/e-9/copilot-icq/internal/copilot"
)

// tickEvery returns a Cmd that sends a TickMsg after the given duration.
func tickEvery(d time.Duration) tea.Cmd {
return tea.Tick(d, func(_ time.Time) tea.Msg {
return TickMsg{}
})
}

// --- SDK commands ---

// sdkStart initializes the SDK adapter connection.
func sdkStart(a *copilot.Adapter) tea.Cmd {
return func() tea.Msg {
err := a.Start(context.Background())
return SDKConnectedMsg{Err: err}
}
}

// sdkListSessions lists sessions via the SDK.
func sdkListSessions(a *copilot.Adapter) tea.Cmd {
return func() tea.Msg {
sessions, err := a.ListSessions(context.Background())
return SessionsLoadedMsg{Sessions: sessions, Err: err}
}
}

// sdkResumeSession resumes a session via the SDK.
func sdkResumeSession(a *copilot.Adapter, sessionID string) tea.Cmd {
return func() tea.Msg {
err := a.ResumeSession(context.Background(), sessionID)
return SDKSessionResumedMsg{SessionID: sessionID, Err: err}
}
}

// sdkLoadHistory loads conversation history via the SDK.
func sdkLoadHistory(a *copilot.Adapter, sessionID string) tea.Cmd {
return func() tea.Msg {
msgs, err := a.GetHistory(context.Background(), sessionID)
return EventsLoadedMsg{SessionID: sessionID, Messages: msgs, Err: err}
}
}

// sdkSendMessage sends a message via the SDK.
func sdkSendMessage(a *copilot.Adapter, sessionID, text string) tea.Cmd {
return func() tea.Msg {
_, err := a.Send(context.Background(), sessionID, text)
if err != nil {
return MessageSentMsg{SessionID: sessionID, Err: err}
}
// Don't return MessageSentMsg here — the SDK will send SessionIdle
// via the Events channel when the response is complete.
return nil
}
}

// sdkAbort cancels the current message processing via the SDK.
func sdkAbort(a *copilot.Adapter, sessionID string) tea.Cmd {
return func() tea.Msg {
_ = a.Abort(context.Background(), sessionID)
return nil
}
}

// listenSDKEvents reads from the SDK adapter's Events channel.
func listenSDKEvents(a *copilot.Adapter) tea.Cmd {
return func() tea.Msg {
evt, ok := <-a.Events
if !ok {
return SDKDisconnectedMsg{}
}
return SDKEventMsg{Event: evt}
}
}
