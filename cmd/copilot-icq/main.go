package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/e-9/copilot-icq/internal/app"
	"github.com/e-9/copilot-icq/internal/config"
	"github.com/e-9/copilot-icq/internal/infra/sessionrepo"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	repo := sessionrepo.New(cfg.SessionStatePath)
	model := app.NewModel(repo)

	p := tea.NewProgram(model, tea.WithAltScreen(), tea.WithMouseCellMotion())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
