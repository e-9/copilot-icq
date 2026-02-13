package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"time"
)

// hookPayload is the envelope sent to the TUI via Unix socket.
type hookPayload struct {
	Event     string          `json:"event"`
	SessionID string          `json:"sessionId"`
	CWD       string          `json:"cwd"`
	Timestamp string          `json:"timestamp"`
	Data      json.RawMessage `json:"data"`
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "usage: copilot-icq-hook <event-name>\n")
		os.Exit(1)
	}

	eventName := os.Args[1]

	// Read stdin payload from Copilot CLI
	stdinData, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "copilot-icq-hook: failed to read stdin: %v\n", err)
		os.Exit(1)
	}

	// Extract sessionId and cwd from the stdin JSON if present
	var stdinJSON struct {
		SessionID string `json:"sessionId"`
		CWD       string `json:"cwd"`
	}
	json.Unmarshal(stdinData, &stdinJSON)

	sessionID := stdinJSON.SessionID
	if sessionID == "" {
		sessionID = os.Getenv("COPILOT_SESSION_ID")
	}
	cwd := stdinJSON.CWD
	if cwd == "" {
		cwd, _ = os.Getwd()
	}

	payload := hookPayload{
		Event:     eventName,
		SessionID: sessionID,
		CWD:       cwd,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Data:      stdinData,
	}

	// Connect to TUI socket
	sockPath := socketPath()
	conn, err := net.DialTimeout("unix", sockPath, 2*time.Second)
	if err != nil {
		// TUI not running — silently exit (hook should not block Copilot CLI)
		os.Exit(0)
	}
	defer conn.Close()

	data, err := json.Marshal(payload)
	if err != nil {
		fmt.Fprintf(os.Stderr, "copilot-icq-hook: marshal error: %v\n", err)
		os.Exit(1)
	}

	data = append(data, '\n')
	conn.Write(data)
}

func socketPath() string {
	// Allow override via env var
	if p := os.Getenv("COPILOT_ICQ_SOCKET"); p != "" {
		return p
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".copilot", "copilot-icq.sock")
}
