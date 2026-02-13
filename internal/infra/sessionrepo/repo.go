package sessionrepo

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/e-9/copilot-icq/internal/domain"
	"gopkg.in/yaml.v3"
)

// Repo discovers and reads Copilot CLI sessions from the file system.
type Repo struct {
	basePath string
}

// New creates a session repository rooted at the given session-state path.
func New(basePath string) *Repo {
	return &Repo{basePath: basePath}
}

// BasePath returns the session-state directory path.
func (r *Repo) BasePath() string {
	return r.basePath
}

// List returns all discovered sessions, sorted by UpdatedAt descending (most recent first).
func (r *Repo) List() ([]domain.Session, error) {
	entries, err := os.ReadDir(r.basePath)
	if err != nil {
		return nil, fmt.Errorf("reading session-state dir: %w", err)
	}

	var sessions []domain.Session
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		wsPath := filepath.Join(r.basePath, entry.Name(), "workspace.yaml")
		s, err := r.parseWorkspace(wsPath)
		if err != nil {
			continue // skip unparseable sessions
		}
		sessions = append(sessions, s)
	}

	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].UpdatedAt.After(sessions[j].UpdatedAt)
	})

	return sessions, nil
}

func (r *Repo) parseWorkspace(path string) (domain.Session, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return domain.Session{}, fmt.Errorf("reading workspace.yaml: %w", err)
	}

	var s domain.Session
	if err := yaml.Unmarshal(data, &s); err != nil {
		return domain.Session{}, fmt.Errorf("parsing workspace.yaml: %w", err)
	}
	return s, nil
}
