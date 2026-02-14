package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/e-9/copilot-icq/internal/app"
	"github.com/e-9/copilot-icq/internal/config"
	"github.com/e-9/copilot-icq/internal/infra/hookserver"
	"github.com/e-9/copilot-icq/internal/infra/notifier"
	"github.com/e-9/copilot-icq/internal/infra/runner"
	"github.com/e-9/copilot-icq/internal/infra/sessionrepo"
	"github.com/e-9/copilot-icq/internal/infra/watcher"
)

func main() {
	// Handle subcommands
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "doctor":
			runDoctor()
			return
		case "install-hooks":
			runInstallHooks()
			return
		}
	}

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	// Load user config
	configPath := ""
	for i, arg := range os.Args {
		if arg == "--config" && i+1 < len(os.Args) {
			configPath = os.Args[i+1]
		}
	}
	appCfg := config.LoadAppConfig(configPath)

	repo := sessionrepo.New(cfg.SessionStatePath)

	w, err := watcher.New(cfg.SessionStatePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error creating watcher: %v\n", err)
		os.Exit(1)
	}
	defer w.Close()
	go w.Start()

	// Create runner for sending messages (nil if copilot not found)
	mode := runner.ModeScoped
	if appCfg.SecurityMode == "full-auto" {
		mode = runner.ModeFullAuto
	}
	var r *runner.Runner
	if cfg.CopilotBinPath != "" {
		r = runner.New(cfg.CopilotBinPath, mode)
	}

	// Start hook server
	hookSrv, err := hookserver.New(hookserver.SocketPath())
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: hook server failed to start: %v\n", err)
	}
	if hookSrv != nil {
		defer hookSrv.Close()
		go hookSrv.Start()
	}

	model := app.NewModel(repo, w, r, appCfg)

	p := tea.NewProgram(model, tea.WithAltScreen(), tea.WithMouseCellMotion())

	// Set up notification router with TUI backend
	tuiNotifier := notifier.NewTUI(p)
	router := notifier.NewRouter(tuiNotifier)

	// Add OS notifications
	router.Add(notifier.NewOS())

	// Add push notifications if configured
	if push := notifier.NewPush(); push != nil {
		router.Add(push)
	}

	// Bridge hook events to notification router
	if hookSrv != nil {
		go func() {
			for evt := range hookSrv.Events() {
				router.Notify(notifier.Notification{
					SessionID: evt.SessionID,
					Event:     evt.Event,
					Title:     fmt.Sprintf("Copilot: %s", evt.Event),
					Body:      fmt.Sprintf("Session %s in %s", shortID(evt.SessionID), evt.CWD),
				})
			}
		}()
	}

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func runDoctor() {
	fmt.Println("🟢 Copilot ICQ — Doctor")
	fmt.Println()

	results := config.Doctor()
	allOK := true
	for _, r := range results {
		icon := "✅"
		if !r.OK {
			icon = "❌"
			allOK = false
		}
		fmt.Printf("  %s %s: %s\n", icon, r.Name, r.Detail)
	}

	fmt.Println()
	if allOK {
		fmt.Println("  All checks passed! Run copilot-icq to start.")
	} else {
		fmt.Println("  Some checks failed. Fix the issues above and try again.")
		os.Exit(1)
	}
}

func runInstallHooks() {
	// Determine target directory
	dir := "."
	if len(os.Args) > 2 {
		dir = os.Args[2]
	}

	hooksDir := filepath.Join(dir, ".github", "hooks")
	if err := os.MkdirAll(hooksDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "error creating hooks directory: %v\n", err)
		os.Exit(1)
	}

	// Find the hook binary path
	hookBin, err := exec.LookPath("copilot-icq-hook")
	if err != nil {
		// Fall back to relative path
		hookBin = "copilot-icq-hook"
	}

	// Generate hook config
	hookConfig := map[string]interface{}{
		"version": 1,
		"hooks": map[string]interface{}{
			"sessionStart": []map[string]interface{}{
				{
					"type":       "command",
					"bash":       fmt.Sprintf("%s sessionStart", hookBin),
					"powershell": fmt.Sprintf("%s sessionStart", hookBin),
					"timeoutSec": 5,
				},
			},
			"sessionEnd": []map[string]interface{}{
				{
					"type":       "command",
					"bash":       fmt.Sprintf("%s sessionEnd", hookBin),
					"powershell": fmt.Sprintf("%s sessionEnd", hookBin),
					"timeoutSec": 5,
				},
			},
			"postToolUse": []map[string]interface{}{
				{
					"type":       "command",
					"bash":       fmt.Sprintf("%s postToolUse", hookBin),
					"powershell": fmt.Sprintf("%s postToolUse", hookBin),
					"timeoutSec": 5,
				},
			},
			"errorOccurred": []map[string]interface{}{
				{
					"type":       "command",
					"bash":       fmt.Sprintf("%s errorOccurred", hookBin),
					"powershell": fmt.Sprintf("%s errorOccurred", hookBin),
					"timeoutSec": 5,
				},
			},
		},
	}

	data, err := json.MarshalIndent(hookConfig, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error marshaling hook config: %v\n", err)
		os.Exit(1)
	}

	hookFile := filepath.Join(hooksDir, "copilot-icq.json")
	if err := os.WriteFile(hookFile, data, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "error writing hook config: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("🟢 Copilot ICQ — Hooks Installed")
	fmt.Println()
	fmt.Printf("  ✅ Created %s\n", hookFile)
	fmt.Println()
	fmt.Println("  Hook events: sessionStart, sessionEnd, postToolUse, errorOccurred")
	fmt.Println("  The TUI will receive real-time notifications when Copilot CLI fires these hooks.")
	fmt.Println()
	fmt.Printf("  Make sure '%s' is in your PATH.\n", hookBin)
}

func shortID(id string) string {
	if len(id) > 8 {
		return id[:8]
	}
	return id
}
