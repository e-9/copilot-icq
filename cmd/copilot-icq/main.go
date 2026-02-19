package main

import (
"fmt"
"os"

tea "github.com/charmbracelet/bubbletea"
"github.com/e-9/copilot-icq/internal/app"
"github.com/e-9/copilot-icq/internal/config"
"github.com/e-9/copilot-icq/internal/copilot"
)

func main() {
// Handle subcommands
if len(os.Args) > 1 {
switch os.Args[1] {
case "doctor":
runDoctor()
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

adapter := copilot.New()
model := app.NewModel(cfg.SessionStatePath, appCfg, adapter)

p := tea.NewProgram(model, tea.WithAltScreen(), tea.WithMouseCellMotion())
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
