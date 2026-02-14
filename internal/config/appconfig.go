package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// AppConfig holds user-configurable settings loaded from ~/.copilot-icq/config.yaml.
type AppConfig struct {
	SecurityMode  string   `yaml:"security_mode"`  // "scoped" (default), "full-auto", or "interactive"
	AllowedTools  []string `yaml:"allowed_tools"`   // tools allowed in scoped mode
	ExportDir     string   `yaml:"export_dir"`      // directory for conversation exports
	Notifications struct {
		OS   bool   `yaml:"os"`   // enable OS desktop notifications
		Push bool   `yaml:"push"` // enable ntfy.sh push notifications
		Topic string `yaml:"topic"` // ntfy.sh topic
	} `yaml:"notifications"`
}

// DefaultAppConfig returns the default configuration.
func DefaultAppConfig() *AppConfig {
	return &AppConfig{
		SecurityMode: "scoped",
		AllowedTools: []string{"view", "glob", "grep", "bash"},
		ExportDir:    ".",
	}
}

// AppConfigPath returns the default config file path.
func AppConfigPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".copilot-icq", "config.yaml")
}

// LoadAppConfig loads user configuration from file, falling back to defaults.
func LoadAppConfig(path string) *AppConfig {
	cfg := DefaultAppConfig()

	if path == "" {
		path = AppConfigPath()
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return cfg // file doesn't exist, use defaults
	}

	yaml.Unmarshal(data, cfg)
	return cfg
}

// SaveAppConfig writes configuration to the given path.
func SaveAppConfig(cfg *AppConfig, path string) error {
	if path == "" {
		path = AppConfigPath()
	}

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}
