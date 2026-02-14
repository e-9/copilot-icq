package ptyproxy

import (
	"os"
	"strings"
	"testing"
	"time"
)

// skipIfNoPTY skips the test if PTY devices are unavailable (e.g., CI, sandbox).
func skipIfNoPTY(t *testing.T) {
	t.Helper()
	// Check if /dev/ptmx exists (macOS/Linux PTY master)
	if _, err := os.Stat("/dev/ptmx"); err != nil {
		t.Skip("PTY not available: /dev/ptmx missing")
	}
	// Quick test: try to spawn a trivial command
	s, err := SpawnRaw("true")
	if err != nil {
		t.Skipf("PTY not available: %v", err)
	}
	s.Close()
}

func TestSpawnRaw_Echo(t *testing.T) {
	skipIfNoPTY(t)
	s, err := SpawnRaw("echo", "hello from pty")
	if err != nil {
		t.Fatalf("SpawnRaw failed: %v", err)
	}
	defer s.Close()

	var output strings.Builder
	timeout := time.After(5 * time.Second)

	for {
		select {
		case chunk, ok := <-s.Output():
			if !ok {
				goto done
			}
			output.WriteString(chunk.Cleaned)
		case <-timeout:
			t.Fatal("timeout waiting for output")
		}
	}

done:
	if !strings.Contains(output.String(), "hello from pty") {
		t.Errorf("expected output to contain 'hello from pty', got: %q", output.String())
	}
}

func TestSpawnRaw_WriteInput(t *testing.T) {
	skipIfNoPTY(t)
	s, err := SpawnRaw("cat")
	if err != nil {
		t.Fatalf("SpawnRaw failed: %v", err)
	}
	defer s.Close()

	if err := s.Write("test input\n"); err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	var output strings.Builder
	timeout := time.After(5 * time.Second)

	for {
		select {
		case chunk, ok := <-s.Output():
			if !ok {
				goto done
			}
			output.WriteString(chunk.Cleaned)
			if strings.Contains(output.String(), "test input") {
				goto done
			}
		case <-timeout:
			t.Fatal("timeout waiting for echo")
		}
	}

done:
	if !strings.Contains(output.String(), "test input") {
		t.Errorf("expected echo of 'test input', got: %q", output.String())
	}
}

func TestSession_IsRunning(t *testing.T) {
	skipIfNoPTY(t)
	s, err := SpawnRaw("echo", "done")
	if err != nil {
		t.Fatalf("SpawnRaw failed: %v", err)
	}
	defer s.Close()

	<-s.Done()

	if s.IsRunning() {
		t.Error("expected IsRunning() to be false after process exits")
	}
}

func TestSession_Close(t *testing.T) {
	skipIfNoPTY(t)
	s, err := SpawnRaw("cat")
	if err != nil {
		t.Fatalf("SpawnRaw failed: %v", err)
	}

	if err := s.Close(); err != nil {
		t.Errorf("Close() returned error: %v", err)
	}

	if err := s.Close(); err != nil {
		t.Errorf("double Close() returned error: %v", err)
	}
}
