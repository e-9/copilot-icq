package hookserver

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"sync"
)

// HookEvent represents an event received from the companion binary.
type HookEvent struct {
	Event     string          `json:"event"`     // e.g. sessionStart, sessionEnd, postToolUse
	SessionID string          `json:"sessionId"` // copilot session ID
	CWD       string          `json:"cwd"`       // working directory
	Timestamp string          `json:"timestamp"` // ISO 8601
	Data      json.RawMessage `json:"data"`      // raw event payload from Copilot CLI
}

// Server listens on a Unix socket for hook events from companion binaries.
type Server struct {
	socketPath string
	listener   net.Listener
	events     chan HookEvent
	mu         sync.Mutex
	closed     bool
}

// SocketPath returns the default socket path for the hook server.
func SocketPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".copilot", "copilot-icq.sock")
}

// New creates a new hook server on the given socket path.
func New(socketPath string) (*Server, error) {
	// Remove stale socket file
	os.Remove(socketPath)

	// Ensure parent directory exists
	if err := os.MkdirAll(filepath.Dir(socketPath), 0700); err != nil {
		return nil, fmt.Errorf("cannot create socket directory: %w", err)
	}

	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		return nil, fmt.Errorf("cannot listen on %s: %w", socketPath, err)
	}

	return &Server{
		socketPath: socketPath,
		listener:   listener,
		events:     make(chan HookEvent, 64),
	}, nil
}

// Events returns a channel that receives hook events.
func (s *Server) Events() <-chan HookEvent {
	return s.events
}

// Start begins accepting connections. Call in a goroutine.
func (s *Server) Start() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			s.mu.Lock()
			closed := s.closed
			s.mu.Unlock()
			if closed {
				return
			}
			continue
		}
		go s.handleConn(conn)
	}
}

func (s *Server) handleConn(conn net.Conn) {
	defer conn.Close()
	scanner := bufio.NewScanner(conn)
	// Allow large payloads (1MB)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		var evt HookEvent
		if err := json.Unmarshal(scanner.Bytes(), &evt); err != nil {
			continue
		}
		select {
		case s.events <- evt:
		default:
			// Drop if channel full (non-blocking)
		}
	}
}

// Close shuts down the server and removes the socket file.
func (s *Server) Close() error {
	s.mu.Lock()
	s.closed = true
	s.mu.Unlock()
	err := s.listener.Close()
	os.Remove(s.socketPath)
	return err
}

// SocketPathValue returns the socket path this server listens on.
func (s *Server) SocketPathValue() string {
	return s.socketPath
}
