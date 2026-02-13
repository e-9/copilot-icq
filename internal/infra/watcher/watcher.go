package watcher

import (
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
)

// EventFileChanged is sent when an events.jsonl file is modified.
type EventFileChanged struct {
	SessionID string
	Path      string
}

// SessionDirChanged is sent when sessions are added or removed.
type SessionDirChanged struct{}

// Watcher monitors the session-state directory for changes.
type Watcher struct {
	basePath string
	fsw      *fsnotify.Watcher
	events   chan interface{}
	done     chan struct{}
}

// New creates a watcher for the given session-state directory.
func New(basePath string) (*Watcher, error) {
	fsw, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	return &Watcher{
		basePath: basePath,
		fsw:      fsw,
		events:   make(chan interface{}, 64),
		done:     make(chan struct{}),
	}, nil
}

// Events returns the channel of file change events.
func (w *Watcher) Events() <-chan interface{} {
	return w.events
}

// WatchSession starts watching a specific session's events.jsonl.
func (w *Watcher) WatchSession(sessionID string) error {
	sessionDir := filepath.Join(w.basePath, sessionID)
	return w.fsw.Add(sessionDir)
}

// UnwatchSession stops watching a session directory.
func (w *Watcher) UnwatchSession(sessionID string) {
	sessionDir := filepath.Join(w.basePath, sessionID)
	w.fsw.Remove(sessionDir)
}

// Start begins processing file system events. Call in a goroutine.
func (w *Watcher) Start() {
	// Watch the base session-state directory for new/removed sessions
	w.fsw.Add(w.basePath)

	// Debounce timer to coalesce rapid writes
	var debounceTimer *time.Timer
	pending := make(map[string]bool)

	for {
		select {
		case event, ok := <-w.fsw.Events:
			if !ok {
				return
			}

			name := filepath.Base(event.Name)

			if name == "events.jsonl" && (event.Has(fsnotify.Write) || event.Has(fsnotify.Create)) {
				// Extract session ID from path: .../session-state/<id>/events.jsonl
				sessionDir := filepath.Dir(event.Name)
				sessionID := filepath.Base(sessionDir)
				pending[sessionID] = true

				// Debounce: wait 100ms for rapid writes to settle
				if debounceTimer != nil {
					debounceTimer.Stop()
				}
				debounceTimer = time.AfterFunc(100*time.Millisecond, func() {
					for sid := range pending {
						select {
						case w.events <- EventFileChanged{SessionID: sid, Path: filepath.Join(w.basePath, sid, "events.jsonl")}:
						default:
						}
					}
					pending = make(map[string]bool)
				})
			}

			// Detect new session directories
			if event.Has(fsnotify.Create) && filepath.Dir(event.Name) == w.basePath {
				select {
				case w.events <- SessionDirChanged{}:
				default:
				}
			}

		case _, ok := <-w.fsw.Errors:
			if !ok {
				return
			}

		case <-w.done:
			return
		}
	}
}

// Close stops the watcher and releases resources.
func (w *Watcher) Close() error {
	close(w.done)
	return w.fsw.Close()
}
