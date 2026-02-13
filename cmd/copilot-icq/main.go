package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/e-9/copilot-icq/internal/app"
	"github.com/e-9/copilot-icq/internal/config"
	"github.com/e-9/copilot-icq/internal/infra/runner"
	"github.com/e-9/copilot-icq/internal/infra/sessionrepo"
	"github.com/e-9/copilot-icq/internal/infra/watcher"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	repo := sessionrepo.New(cfg.SessionStatePath)

	w, err := watcher.New(cfg.SessionStatePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error creating watcher: %v\n", err)
		os.Exit(1)
	}
	defer w.Close()
	go w.Start()

	// Create runner for sending messages (nil if copilot not found)
	var r *runner.Runner
	if cfg.CopilotBinPath != "" {
		r = runner.New(cfg.CopilotBinPath, runner.ModeScoped)
	}

	model := app.NewModel(repo, w, r)

	p := tea.NewProgram(model, tea.WithAltScreen(), tea.WithMouseCellMotion())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
