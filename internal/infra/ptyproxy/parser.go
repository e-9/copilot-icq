package ptyproxy

import (
	"regexp"
	"strings"

	"github.com/acarl005/stripansi"
)

// ApprovalOption represents a selectable choice in a tool approval prompt.
type ApprovalOption struct {
	Label    string // e.g., "Yes", "No, and tell Copilot what to do differently"
	Shortcut string // e.g., "1", "2", "3"
	Index    int    // 0-based index
}

// ApprovalPrompt represents a detected tool approval prompt from Copilot CLI.
type ApprovalPrompt struct {
	Question string
	Options  []ApprovalOption
	Raw      string // raw text that triggered the detection
}

// OutputChunk represents a parsed chunk of PTY output.
type OutputChunk struct {
	Raw       []byte
	Cleaned   string          // ANSI-stripped text
	IsPrompt  bool            // true if this chunk contains an approval prompt
	Prompt    *ApprovalPrompt // non-nil if IsPrompt
	IsPartial bool            // true if this is a partial/streaming token
}

// Copilot CLI approval prompt patterns
var (
	promptQuestionRe = regexp.MustCompile(`(?im)((?:do you want|allow|approve|confirm|proceed|permission|accept|execute|apply|run this)\b.*)$`)
	numberedOptionRe = regexp.MustCompile(`(?m)^\s*[❯►>]?\s*(\d+)\.\s+(.+)$`)
)

// StripANSI removes ANSI escape sequences from raw bytes.
func StripANSI(raw []byte) string {
	return stripansi.Strip(string(raw))
}

// DetectApprovalPrompt checks cleaned text for a Copilot CLI approval prompt.
// Returns nil if no prompt is detected. Requires non-whitespace text after the
// last numbered option (e.g., "Confirm with..." instruction line) to confirm
// the prompt is complete and all options have been received.
func DetectApprovalPrompt(cleaned string) *ApprovalPrompt {
	qMatch := promptQuestionRe.FindString(cleaned)
	if qMatch == "" {
		return nil
	}

	optMatches := numberedOptionRe.FindAllStringSubmatch(cleaned, -1)
	if len(optMatches) < 2 {
		return nil
	}

	// Require non-whitespace text after the last option to confirm completeness.
	// This prevents triggering while options are still arriving in chunks.
	lastOptLoc := numberedOptionRe.FindAllStringIndex(cleaned, -1)
	afterOpts := cleaned[lastOptLoc[len(lastOptLoc)-1][1]:]
	if strings.TrimSpace(afterOpts) == "" {
		return nil
	}

	prompt := &ApprovalPrompt{
		Question: strings.TrimSpace(qMatch),
		Raw:      cleaned,
	}

	for i, m := range optMatches {
		prompt.Options = append(prompt.Options, ApprovalOption{
			Label:    strings.TrimSpace(m[2]),
			Shortcut: m[1],
			Index:    i,
		})
	}

	return prompt
}

// Parser is a stateful PTY output parser that accumulates chunks
// and detects approval prompts, handling split ANSI sequences.
type Parser struct {
	buffer strings.Builder
}

// NewParser creates a new stateful PTY output parser.
func NewParser() *Parser {
	return &Parser{}
}

// Feed processes a raw chunk of PTY output and returns parsed results.
func (p *Parser) Feed(raw []byte) OutputChunk {
	cleaned := StripANSI(raw)
	p.buffer.WriteString(cleaned)

	// Cap buffer size to prevent regex performance issues
	if p.buffer.Len() > 2048 {
		content := p.buffer.String()
		keep := content[len(content)-1024:]
		p.buffer.Reset()
		p.buffer.WriteString(keep)
	}

	chunk := OutputChunk{
		Raw:     raw,
		Cleaned: cleaned,
	}

	buffered := p.buffer.String()
	if prompt := DetectApprovalPrompt(buffered); prompt != nil {
		chunk.IsPrompt = true
		chunk.Prompt = prompt
		p.buffer.Reset()
	}

	return chunk
}

// Reset clears the parser's internal buffer.
func (p *Parser) Reset() {
	p.buffer.Reset()
}

// Buffer returns the current accumulated text.
func (p *Parser) Buffer() string {
	return p.buffer.String()
}
