package config

import (
"fmt"
"os"
"path/filepath"
)

// Config holds application configuration.
type Config struct {
SessionStatePath string
}

// Load discovers paths and returns the app configuration.
func Load() (*Config, error) {
home, err := os.UserHomeDir()
if err != nil {
return nil, fmt.Errorf("cannot determine home directory: %w", err)
}

sessionPath := filepath.Join(home, ".copilot", "session-state")
if _, err := os.Stat(sessionPath); os.IsNotExist(err) {
return nil, fmt.Errorf("session-state directory not found: %s", sessionPath)
}

return &Config{
SessionStatePath: sessionPath,
}, nil
}
