package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// hookPayload is the envelope sent to the TUI via Unix socket.
type hookPayload struct {
	Event     string          `json:"event"`
	SessionID string          `json:"sessionId"`
	CWD       string          `json:"cwd"`
	Timestamp string          `json:"timestamp"`
	Data      json.RawMessage `json:"data"`
}

// preToolUseInput is the JSON structure Copilot CLI sends for preToolUse events.
type preToolUseInput struct {
	SessionID string `json:"sessionId"`
	CWD       string `json:"cwd"`
	ToolName  string `json:"toolName"`
	ToolArgs  string `json:"toolArgs"`
}

// denyConfig mirrors the deny fields from ~/.copilot-icq/config.yaml.
type denyConfig struct {
	DeniedTools    []string `yaml:"denied_tools"`
	DeniedPatterns []string `yaml:"denied_patterns"`
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

	// Handle preToolUse deny policy
	if eventName == "preToolUse" {
		handlePreToolUse(stdinData)
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

// handlePreToolUse checks the tool against the deny policy and outputs
// a deny decision to stdout if matched. Copilot CLI reads this response.
func handlePreToolUse(stdinData []byte) {
	var input preToolUseInput
	if err := json.Unmarshal(stdinData, &input); err != nil {
		return
	}

	cfg := loadDenyConfig()
	if cfg == nil {
		return
	}

	// Check denied tools
	toolLower := strings.ToLower(input.ToolName)
	for _, denied := range cfg.DeniedTools {
		if strings.ToLower(denied) == toolLower {
			outputDeny(fmt.Sprintf("Tool '%s' is blocked by policy", input.ToolName))
			return
		}
	}

	// Check denied patterns against tool args
	argsLower := strings.ToLower(input.ToolArgs)
	for _, pattern := range cfg.DeniedPatterns {
		if strings.Contains(argsLower, strings.ToLower(pattern)) {
			outputDeny(fmt.Sprintf("Blocked by pattern: %s", pattern))
			return
		}
	}
}

func outputDeny(reason string) {
	resp := map[string]string{
		"permissionDecision":       "deny",
		"permissionDecisionReason": reason,
	}
	json.NewEncoder(os.Stdout).Encode(resp)
}

func loadDenyConfig() *denyConfig {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil
	}
	data, err := os.ReadFile(filepath.Join(home, ".copilot-icq", "config.yaml"))
	if err != nil {
		return nil
	}
	var cfg denyConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil
	}
	if len(cfg.DeniedTools) == 0 && len(cfg.DeniedPatterns) == 0 {
		return nil
	}
	return &cfg
}

func socketPath() string {
	if p := os.Getenv("COPILOT_ICQ_SOCKET"); p != "" {
		return p
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".copilot", "copilot-icq.sock")
}
