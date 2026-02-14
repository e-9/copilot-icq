package ptyproxy

import (
	"testing"
)

func TestStripANSI(t *testing.T) {
	tests := []struct {
		name string
		raw  []byte
		want string
	}{
		{
			name: "plain text",
			raw:  []byte("Hello, World!"),
			want: "Hello, World!",
		},
		{
			name: "color codes",
			raw:  []byte("\x1b[31mError\x1b[0m: something failed"),
			want: "Error: something failed",
		},
		{
			name: "bold and color",
			raw:  []byte("\x1b[1m\x1b[32mSuccess\x1b[0m"),
			want: "Success",
		},
		{
			name: "cursor movement",
			raw:  []byte("\x1b[2K\x1b[1G> prompt"),
			want: "> prompt",
		},
		{
			name: "empty input",
			raw:  []byte(""),
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := StripANSI(tt.raw)
			if got != tt.want {
				t.Errorf("StripANSI() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDetectApprovalPrompt(t *testing.T) {
	tests := []struct {
		name      string
		cleaned   string
		wantNil   bool
		wantQ     string
		wantCount int
	}{
		{
			name: "copilot tool approval",
			cleaned: `Do you want to run this command?

❯ 1. Yes
  2. Yes, and approve sed for the rest of the running session
  3. No, and tell Copilot what to do differently (Esc to stop)

Confirm with number keys or ↑↓ keys and Enter, Cancel with Esc`,
			wantNil:   false,
			wantQ:     "Do you want to run this command?",
			wantCount: 3,
		},
		{
			name: "simple yes/no",
			cleaned: `Do you want to proceed?

  1. Yes
  2. No

Confirm with Enter`,
			wantNil:   false,
			wantQ:     "Do you want to proceed?",
			wantCount: 2,
		},
		{
			name:    "no prompt - plain text",
			cleaned: "I'll implement dark mode for you.",
			wantNil: true,
		},
		{
			name:    "question but no options",
			cleaned: "Do you want to proceed?",
			wantNil: true,
		},
		{
			name: "allow tool prompt",
			cleaned: `Allow bash tool?

  1. Yes
  2. Always allow
  3. Deny

Choose an option`,
			wantNil:   false,
			wantQ:     "Allow bash tool?",
			wantCount: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DetectApprovalPrompt(tt.cleaned)
			if tt.wantNil {
				if got != nil {
					t.Errorf("expected nil, got %+v", got)
				}
				return
			}
			if got == nil {
				t.Fatal("expected non-nil prompt, got nil")
			}
			if got.Question != tt.wantQ {
				t.Errorf("Question = %q, want %q", got.Question, tt.wantQ)
			}
			if len(got.Options) != tt.wantCount {
				t.Errorf("got %d options, want %d", len(got.Options), tt.wantCount)
			}
		})
	}
}

func TestParser_SplitChunks(t *testing.T) {
	p := NewParser()

	// Simulate prompt arriving in multiple chunks
	chunks := []string{
		"Do you want to ",
		"run this command?\n\n",
		"❯ 1. Yes\n",
		"  2. No\n",
		"  3. Cancel\n",
		"\nConfirm with number keys\n",
	}

	var detected bool
	for _, chunk := range chunks {
		result := p.Feed([]byte(chunk))
		if result.IsPrompt {
			detected = true
			if len(result.Prompt.Options) != 3 {
				t.Errorf("expected 3 options, got %d", len(result.Prompt.Options))
			}
		}
	}

	if !detected {
		t.Error("failed to detect prompt across split chunks")
	}
}

func TestParser_Reset(t *testing.T) {
	p := NewParser()
	p.Feed([]byte("some partial text"))
	if p.Buffer() == "" {
		t.Error("buffer should not be empty after Feed")
	}
	p.Reset()
	if p.Buffer() != "" {
		t.Error("buffer should be empty after Reset")
	}
}

func TestApprovalOption_Fields(t *testing.T) {
	prompt := DetectApprovalPrompt(`Do you want to proceed?

  1. Yes
  2. No, cancel it

Choose`)

	if prompt == nil {
		t.Fatal("expected prompt")
	}

	opt := prompt.Options[0]
	if opt.Label != "Yes" {
		t.Errorf("Label = %q, want %q", opt.Label, "Yes")
	}
	if opt.Shortcut != "1" {
		t.Errorf("Shortcut = %q, want %q", opt.Shortcut, "1")
	}
	if opt.Index != 0 {
		t.Errorf("Index = %d, want %d", opt.Index, 0)
	}
}
