package config

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// CheckResult represents a single prerequisite check.
type CheckResult struct {
	Name    string
	OK      bool
	Detail  string
}

// Doctor runs prerequisite checks and returns results.
func Doctor() []CheckResult {
	var results []CheckResult

	// 1. Check Go runtime
	if goVer, err := exec.Command("go", "version").Output(); err == nil {
		results = append(results, CheckResult{
			Name:   "Go runtime",
			OK:     true,
			Detail: strings.TrimSpace(string(goVer)),
		})
	} else {
		results = append(results, CheckResult{
			Name:   "Go runtime",
			OK:     false,
			Detail: "go not found in PATH",
		})
	}

	// 2. Check copilot CLI
	copilotBin, err := exec.LookPath("copilot")
	if err == nil {
		results = append(results, CheckResult{
			Name:   "Copilot CLI",
			OK:     true,
			Detail: copilotBin,
		})
	} else {
		results = append(results, CheckResult{
			Name:   "Copilot CLI",
			OK:     false,
			Detail: "copilot not found in PATH — install with: gh extension install github/gh-copilot",
		})
	}

	// 3. Check gh auth status
	if out, err := exec.Command("gh", "auth", "status").CombinedOutput(); err == nil {
		// Extract the account line
		detail := "authenticated"
		for _, line := range strings.Split(string(out), "\n") {
			if strings.Contains(line, "Logged in") || strings.Contains(line, "account") {
				detail = strings.TrimSpace(line)
				break
			}
		}
		results = append(results, CheckResult{
			Name:   "GitHub auth",
			OK:     true,
			Detail: detail,
		})
	} else {
		results = append(results, CheckResult{
			Name:   "GitHub auth",
			OK:     false,
			Detail: "not authenticated — run: gh auth login",
		})
	}

	// 4. Check session-state directory
	home, err := os.UserHomeDir()
	if err != nil {
		results = append(results, CheckResult{
			Name:   "Session state dir",
			OK:     false,
			Detail: "cannot determine home directory",
		})
	} else {
		sessionPath := home + "/.copilot/session-state"
		if info, err := os.Stat(sessionPath); err == nil && info.IsDir() {
			entries, _ := os.ReadDir(sessionPath)
			sessionCount := 0
			for _, e := range entries {
				if e.IsDir() {
					sessionCount++
				}
			}
			results = append(results, CheckResult{
				Name:   "Session state dir",
				OK:     true,
				Detail: fmt.Sprintf("%s (%d sessions)", sessionPath, sessionCount),
			})
		} else {
			results = append(results, CheckResult{
				Name:   "Session state dir",
				OK:     false,
				Detail: fmt.Sprintf("%s not found — start a Copilot CLI session first", sessionPath),
			})
		}
	}

	return results
}
