package ptyproxy

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/creack/pty"
)

// Session manages an interactive Copilot CLI process running in a pseudo-terminal.
type Session struct {
	cmd    *exec.Cmd
	ptmx   *os.File // PTY master
	output chan OutputChunk
	done   chan struct{}
	parser *Parser
	mu     sync.Mutex
	closed bool
}

// Spawn starts an interactive copilot session in a PTY.
// It runs: copilot -i "message" --resume <sessionID>
// The cwd parameter sets the working directory (should match the session's CWD).
func Spawn(copilotBin, sessionID, message, cwd string, extraArgs ...string) (*Session, error) {
	args := []string{"-i", message, "--resume", sessionID}
	args = append(args, extraArgs...)

	cmd := exec.Command(copilotBin, args...)
	cmd.Env = append(os.Environ(), "TERM=xterm-256color")
	if cwd != "" {
		cmd.Dir = cwd
	}

	ptmx, err := pty.Start(cmd)
	if err != nil {
		return nil, fmt.Errorf("pty.Start: %w", err)
	}

	s := &Session{
		cmd:    cmd,
		ptmx:   ptmx,
		output: make(chan OutputChunk, 64),
		done:   make(chan struct{}),
		parser: NewParser(),
	}

	go s.readLoop()

	return s, nil
}

// SpawnRaw starts an arbitrary command in a PTY (useful for testing).
func SpawnRaw(name string, args ...string) (*Session, error) {
	cmd := exec.Command(name, args...)
	cmd.Env = append(os.Environ(), "TERM=xterm-256color")

	ptmx, err := pty.Start(cmd)
	if err != nil {
		return nil, fmt.Errorf("pty.Start: %w", err)
	}

	s := &Session{
		cmd:    cmd,
		ptmx:   ptmx,
		output: make(chan OutputChunk, 64),
		done:   make(chan struct{}),
		parser: NewParser(),
	}

	go s.readLoop()

	return s, nil
}

// readLoop continuously reads from the PTY and sends parsed chunks.
func (s *Session) readLoop() {
	defer close(s.done)
	defer close(s.output)

	buf := make([]byte, 4096)
	consecutiveEmpty := 0
	for {
		n, err := s.ptmx.Read(buf)
		if n == 0 && err == nil {
			time.Sleep(20 * time.Millisecond)
			continue
		}
		if n > 0 {
			raw := make([]byte, n)
			copy(raw, buf[:n])
			chunk := s.parser.Feed(raw)
			// Only send chunks with meaningful content or detected prompts
			if chunk.Cleaned != "" || chunk.IsPrompt {
				s.output <- chunk
				consecutiveEmpty = 0
			} else {
				consecutiveEmpty++
				if consecutiveEmpty > 10 {
					time.Sleep(10 * time.Millisecond)
				}
			}
		}
		if err != nil {
			if err != io.EOF {
				s.output <- OutputChunk{
					Cleaned: fmt.Sprintf("[PTY error: %v]", err),
				}
			}
			return
		}
	}
}

// Output returns the channel of parsed output chunks.
func (s *Session) Output() <-chan OutputChunk {
	return s.output
}

// Done returns a channel that closes when the PTY process exits.
func (s *Session) Done() <-chan struct{} {
	return s.done
}

// Write sends input to the PTY stdin (e.g., "1\n" to select option 1).
func (s *Session) Write(input string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return fmt.Errorf("session closed")
	}
	_, err := s.ptmx.WriteString(input)
	return err
}

// Close terminates the PTY session.
func (s *Session) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return nil
	}
	s.closed = true

	if err := s.ptmx.Close(); err != nil {
		return err
	}

	_ = s.cmd.Wait()
	return nil
}

// IsRunning returns true if the PTY process is still active.
func (s *Session) IsRunning() bool {
	select {
	case <-s.done:
		return false
	default:
		return true
	}
}
