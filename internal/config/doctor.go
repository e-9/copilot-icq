package config

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
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

	// 5. Check PTY health (macOS/Linux only)
	if runtime.GOOS != "windows" {
		results = append(results, checkPTYHealth()...)
	}

	return results
}

// checkPTYHealth checks PTY device usage and orphaned copilot processes.
func checkPTYHealth() []CheckResult {
	var results []CheckResult

	// Count allocated PTY devices
	ptys, _ := filepath.Glob("/dev/ttys*")
	allocated := len(ptys)

	// Get kernel limit (macOS)
	limit := 0
	if out, err := exec.Command("sysctl", "-n", "kern.tty.ptmx_max").Output(); err == nil {
		limit, _ = strconv.Atoi(strings.TrimSpace(string(out)))
	}

	if limit > 0 {
		usage := float64(allocated) / float64(limit) * 100
		if usage > 80 {
			results = append(results, CheckResult{
				Name: "PTY devices",
				OK:   false,
				Detail: fmt.Sprintf("%d/%d allocated (%.0f%%) — risk of exhaustion! Run: copilot-icq cleanup",
					allocated, limit, usage),
			})
		} else {
			results = append(results, CheckResult{
				Name:   "PTY devices",
				OK:     true,
				Detail: fmt.Sprintf("%d/%d allocated (%.0f%%)", allocated, limit, usage),
			})
		}
	}

	// Count orphaned copilot bash subshells
	orphans := countOrphanBashShells()
	if orphans > 0 {
		results = append(results, CheckResult{
			Name: "Orphaned shells",
			OK:   false,
			Detail: fmt.Sprintf("%d orphaned bash shells from copilot sessions — Run: copilot-icq cleanup",
				orphans),
		})
	} else {
		results = append(results, CheckResult{
			Name:   "Orphaned shells",
			OK:     true,
			Detail: "no orphaned copilot bash shells detected",
		})
	}

	return results
}

// countOrphanBashShells finds bash --norc --noprofile processes whose parent
// is a copilot process, indicating leaked tool execution shells.
func countOrphanBashShells() int {
	out, err := exec.Command("ps", "-eo", "pid,ppid,command").Output()
	if err != nil {
		return 0
	}

	copilotPIDs := map[string]bool{}
	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) >= 3 && strings.HasSuffix(fields[2], "copilot") {
			copilotPIDs[fields[0]] = true
		}
	}

	count := 0
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) >= 4 &&
			copilotPIDs[fields[1]] &&
			strings.Contains(line, "bash") &&
			strings.Contains(line, "--norc") {
			count++
		}
	}

	return count
}

// CleanupOrphanShells finds and kills orphaned bash shells from copilot sessions.
// Returns the number of processes killed. Excludes shells that are running
// known services (dev servers, etc.) and the current process tree.
func CleanupOrphanShells() (int, error) {
	out, err := exec.Command("ps", "-eo", "pid,ppid,command").Output()
	if err != nil {
		return 0, fmt.Errorf("ps failed: %w", err)
	}

	myPID := fmt.Sprintf("%d", os.Getpid())
	myPPID := fmt.Sprintf("%d", os.Getppid())

	copilotPIDs := map[string]bool{}
	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) >= 3 && strings.HasSuffix(fields[2], "copilot") {
			copilotPIDs[fields[0]] = true
		}
	}

	var orphanPIDs []int
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) >= 4 &&
			copilotPIDs[fields[1]] &&
			strings.Contains(line, "bash") &&
			strings.Contains(line, "--norc") {
			pid := fields[0]
			ppid := fields[1]
			// Skip our own process tree
			if pid == myPID || pid == myPPID || ppid == myPID {
				continue
			}
			if pidInt, err := strconv.Atoi(pid); err == nil {
				// Don't kill shells running known services
				if !strings.Contains(line, "vite") &&
					!strings.Contains(line, "uvicorn") &&
					!strings.Contains(line, "npm") &&
					!strings.Contains(line, "node") &&
					!strings.Contains(line, "server") {
					orphanPIDs = append(orphanPIDs, pidInt)
				}
			}
		}
	}

	killed := 0
	for _, pid := range orphanPIDs {
		proc, err := os.FindProcess(pid)
		if err != nil {
			continue
		}
		if err := proc.Signal(os.Kill); err == nil {
			killed++
		}
	}

	return killed, nil
}
