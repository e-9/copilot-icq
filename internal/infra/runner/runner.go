package runner

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// SecurityMode controls which tools the Copilot session can use.
type SecurityMode string

const (
	// ModeScoped allows only read-only tools (safe default).
	ModeScoped SecurityMode = "scoped"
	// ModeFullAuto allows all tools (user must opt in).
	ModeFullAuto SecurityMode = "full-auto"
)

// ScopedTools is the default whitelist for scoped mode.
var ScopedTools = []string{
	"view",
	"glob",
	"grep",
	"bash",
}

// Result holds the outcome of a runner execution.
type Result struct {
	SessionID string
	Success   bool
	Output    string
	Err       error
}

// Runner sends messages to Copilot CLI sessions via subprocess.
type Runner struct {
	copilotBin   string
	securityMode SecurityMode
	allowedTools []string
}

// New creates a runner with the given copilot binary path and security mode.
func New(copilotBin string, mode SecurityMode) *Runner {
	tools := ScopedTools
	if mode == ModeFullAuto {
		tools = nil
	}
	return &Runner{
		copilotBin:   copilotBin,
		securityMode: mode,
		allowedTools: tools,
	}
}

// Send dispatches a message to a Copilot CLI session and waits for completion.
func (r *Runner) Send(ctx context.Context, sessionID, message string) Result {
	args := []string{
		"-p", message,
		"--resume", sessionID,
	}

	if r.securityMode == ModeFullAuto {
		args = append(args, "--allow-all-tools")
	} else {
		for _, tool := range r.allowedTools {
			args = append(args, "--allow-tool", tool)
		}
	}

	cmd := exec.CommandContext(ctx, r.copilotBin, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return Result{
			SessionID: sessionID,
			Success:   false,
			Output:    strings.TrimSpace(stderr.String()),
			Err:       fmt.Errorf("copilot subprocess: %w", err),
		}
	}

	return Result{
		SessionID: sessionID,
		Success:   true,
		Output:    strings.TrimSpace(stdout.String()),
	}
}

// Mode returns the current security mode.
func (r *Runner) Mode() SecurityMode {
	return r.securityMode
}
