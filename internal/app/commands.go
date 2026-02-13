package app

import (
	"context"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
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
