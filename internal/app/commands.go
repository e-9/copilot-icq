package app

import (
	"context"
	"path/filepath"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/e-9/copilot-icq/internal/domain"
	"github.com/e-9/copilot-icq/internal/infra/eventparser"
	"github.com/e-9/copilot-icq/internal/infra/ptyproxy"
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
func sendMessage(r *runner.Runner, sessionID, message string) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		result := r.Send(ctx, sessionID, message)
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

// streamPTY reads the next chunk from a PTY session and returns it as a Msg.
func streamPTY(session *ptyproxy.Session, sessionID string) tea.Cmd {
	return func() tea.Msg {
		select {
		case chunk, ok := <-session.Output():
			if !ok {
				return PTYClosedMsg{SessionID: sessionID}
			}
			if chunk.IsPrompt {
				return PTYPromptMsg{SessionID: sessionID, Prompt: chunk.Prompt}
			}
			return PTYOutputMsg{SessionID: sessionID, Chunk: chunk}
		case <-session.Done():
			return PTYClosedMsg{SessionID: sessionID}
		}
	}
}

// sendApproval writes the user's selection to the PTY stdin.
func sendApproval(session *ptyproxy.Session, sessionID, shortcut string) tea.Cmd {
	return func() tea.Msg {
		err := session.Write(shortcut + "\n")
		if err != nil {
			return PTYClosedMsg{SessionID: sessionID, Err: err}
		}
		return ApprovalSelectedMsg{SessionID: sessionID, Shortcut: shortcut}
	}
}
