package app

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/e-9/copilot-icq/internal/copilot"
	"github.com/e-9/copilot-icq/internal/domain"
	"github.com/e-9/copilot-icq/internal/infra/eventparser"
	"github.com/e-9/copilot-icq/internal/infra/runner"
	"github.com/e-9/copilot-icq/internal/infra/sessionrepo"
	"github.com/e-9/copilot-icq/internal/infra/watcher"
)

// loadSessions returns a Cmd that discovers sessions from disk.
func loadSessions(repo *sessionrepo.Repo) tea.Cmd {
	return func() tea.Msg {
		sessions, err := repo.List()
		return SessionsLoadedMsg{Sessions: sessions, Err: err}
	}
}

// loadEvents returns a Cmd that reads events for a specific session.
func loadEvents(basePath string, session domain.Session) tea.Cmd {
	return func() tea.Msg {
		eventsPath := filepath.Join(basePath, session.ID, "events.jsonl")
		parser := eventparser.New(eventsPath)
		events, err := parser.ReadAll()
		if err != nil {
			return EventsLoadedMsg{SessionID: session.ID, Err: err}
		}
		messages := domain.EventsToMessages(events)
		return EventsLoadedMsg{SessionID: session.ID, Messages: messages}
	}
}

// watchFiles returns a Cmd that listens for file system changes.
func watchFiles(w *watcher.Watcher) tea.Cmd {
	return func() tea.Msg {
		evt := <-w.Events()
		switch e := evt.(type) {
		case watcher.EventFileChanged:
			return FileChangedMsg{e}
		case watcher.SessionDirChanged:
			return SessionDirChangedMsg{}
		}
		return nil
	}
}

// sendMessage dispatches a message to a copilot session via the runner.
func sendMessage(r *runner.Runner, sessionID, message, cwd string) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		result := r.Send(ctx, sessionID, message, cwd)
		return MessageSentMsg{
			SessionID: result.SessionID,
			Success:   result.Success,
			Output:    result.Output,
			Err:       result.Err,
		}
	}
}

// tickEvery returns a Cmd that sends a TickMsg after the given duration.
func tickEvery(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(_ time.Time) tea.Msg {
		return TickMsg{}
	})
}

// approveSession opens a new Terminal.app window with copilot --resume for the session.
// The TUI stays running while the user interacts in the native terminal.
func approveSession(copilotBin, sessionID, cwd string) tea.Cmd {
	return func() tea.Msg {
		script := fmt.Sprintf(
			`tell application "Terminal"
				activate
				do script "cd %s && %s --resume %s"
			end tell`, cwd, copilotBin, sessionID)
		cmd := exec.Command("osascript", "-e", script)
		err := cmd.Run()
		return ApprovalFinishedMsg{SessionID: sessionID, Err: err}
	}
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
		// The SDK will send SessionIdle via Events channel
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
