package eventparser

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/e-9/copilot-icq/internal/domain"
)

// Parser reads events from an events.jsonl file, supporting incremental reads.
type Parser struct {
	path   string
	offset int64
}

// New creates a parser for the given events.jsonl path.
func New(path string) *Parser {
	return &Parser{path: path}
}

// ReadAll reads all events from the file (resets offset to 0 first).
func (p *Parser) ReadAll() ([]domain.Event, error) {
	p.offset = 0
	return p.ReadNew()
}

// ReadNew reads only events added since the last read.
func (p *Parser) ReadNew() ([]domain.Event, error) {
	f, err := os.Open(p.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // file doesn't exist yet — empty conversation
		}
		return nil, fmt.Errorf("opening events file: %w", err)
	}
	defer f.Close()

	if p.offset > 0 {
		if _, err := f.Seek(p.offset, io.SeekStart); err != nil {
			return nil, fmt.Errorf("seeking to offset %d: %w", p.offset, err)
		}
	}

	var events []domain.Event
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 1024*1024), 10*1024*1024) // support up to 10MB lines

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var evt domain.Event
		if err := json.Unmarshal(line, &evt); err != nil {
			continue // skip malformed lines
		}
		events = append(events, evt)
	}

	if err := scanner.Err(); err != nil {
		return events, fmt.Errorf("scanning events: %w", err)
	}

	// Update offset to current position
	newOffset, err := f.Seek(0, io.SeekCurrent)
	if err == nil {
		p.offset = newOffset
	}

	return events, nil
}

// Offset returns the current file offset.
func (p *Parser) Offset() int64 {
	return p.offset
}
