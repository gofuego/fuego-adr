// Package parser provides the compiled Fuego parser for .adr.md files.
//
// It splits YAML frontmatter from the body, normalizes list fields
// (author, approvers, tags, supersedes, superseded_by, affects),
// splits the body on ## headings into Context/Decision/Consequences
// sections, and renders each section's Markdown to HTML via goldmark.
package adr

import (
	"bytes"
	"fmt"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gofuego/fuego/core"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
)

var md = goldmark.New(goldmark.WithExtensions(extension.GFM))

// adrNumberRe matches the leading numeric prefix in filenames like "0012-use-postgres.adr.md".
var adrNumberRe = regexp.MustCompile(`^(\d+)-`)

// requiredSections are the section headings enforced for accepted ADRs.
var requiredSections = []string{"context", "decision", "consequences"}

// ADRParser implements core.Parser and core.FilenameParser for .adr.md files.
type ADRParser struct{}

// New returns a new ADRParser.
func New() *ADRParser { return &ADRParser{} }

func (p *ADRParser) Type() string        { return "adr" }
func (p *ADRParser) Filenames() []string { return []string{"*.adr.md"} }

// Parse extracts frontmatter, normalizes metadata, splits the body into
// sections, and renders each section's Markdown to HTML.
func (p *ADRParser) Parse(raw []byte) (core.Envelope, []core.Node, error) {
	env, payload, err := core.SplitFrontmatter(raw)
	if err != nil {
		return nil, nil, err
	}
	if env == nil {
		env = make(core.Envelope)
	}

	normalizeEnvelope(env)

	sections := splitSections(payload)
	var nodes []core.Node
	for _, sec := range sections {
		html, err := renderMarkdown(sec.content)
		if err != nil {
			return nil, nil, fmt.Errorf("rendering section %q: %w", sec.heading, err)
		}
		nodes = append(nodes, core.Node{
			Type:    sec.heading,
			Content: html,
			Raw:     true,
		})
	}

	return env, nodes, nil
}

// section represents a parsed body section.
type section struct {
	heading string // lowercase, e.g. "context", "decision"
	content []byte // raw Markdown content under the heading
}

// splitSections splits the Markdown body on ## headings.
// Content before the first heading is emitted as type "preamble".
func splitSections(body []byte) []section {
	lines := bytes.Split(body, []byte("\n"))
	var sections []section
	var current *section

	for _, line := range lines {
		trimmed := bytes.TrimSpace(line)
		if bytes.HasPrefix(trimmed, []byte("## ")) {
			if current != nil && len(bytes.TrimSpace(current.content)) > 0 {
				sections = append(sections, *current)
			}
			heading := strings.ToLower(strings.TrimSpace(string(trimmed[3:])))
			current = &section{heading: heading}
			continue
		}
		if current != nil {
			current.content = append(current.content, line...)
			current.content = append(current.content, '\n')
		} else if len(trimmed) > 0 {
			// Content before first heading
			if current == nil {
				current = &section{heading: "preamble"}
			}
			current.content = append(current.content, line...)
			current.content = append(current.content, '\n')
		}
	}
	if current != nil && len(bytes.TrimSpace(current.content)) > 0 {
		sections = append(sections, *current)
	}
	return sections
}

// renderMarkdown converts Markdown bytes to HTML.
func renderMarkdown(src []byte) (string, error) {
	var buf bytes.Buffer
	if err := md.Convert(src, &buf); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// normalizeEnvelope converts scalar-or-list fields to consistent list form
// and normalizes status to lowercase.
func normalizeEnvelope(env core.Envelope) {
	// Normalize list fields
	for _, key := range []string{"author", "approvers", "tags", "affects"} {
		if v, ok := env[key]; ok {
			env[key] = toStringSlice(v)
		}
	}
	for _, key := range []string{"supersedes", "superseded_by"} {
		if v, ok := env[key]; ok {
			env[key] = toIntSlice(v)
		}
	}

	// Normalize status to lowercase
	if s, ok := env["status"].(string); ok {
		env["status"] = strings.ToLower(strings.TrimSpace(s))
	}

	// Normalize date fields: time.Time → "YYYY-MM-DD" string
	for _, key := range []string{"date_proposed", "date_accepted", "date_deprecated", "date_superseded", "deadline"} {
		if v, ok := env[key]; ok {
			env[key] = formatDate(v)
		}
	}
}

// formatDate converts a value to a date string.
// YAML parses bare dates like 2026-01-15 as time.Time.
func formatDate(v any) string {
	switch d := v.(type) {
	case time.Time:
		return d.Format("2006-01-02")
	case string:
		return d
	default:
		return fmt.Sprintf("%v", v)
	}
}


// toStringSlice normalizes a value to []string.
// Accepts: string, []any (of strings), []string.
func toStringSlice(v any) []string {
	switch val := v.(type) {
	case string:
		return []string{val}
	case []any:
		out := make([]string, 0, len(val))
		for _, item := range val {
			if s, ok := item.(string); ok {
				out = append(out, s)
			}
		}
		return out
	case []string:
		return val
	default:
		return nil
	}
}

// toIntSlice normalizes a value to []int.
// Accepts: int, float64 (YAML numbers), []any (of numbers), []int.
func toIntSlice(v any) []int {
	switch val := v.(type) {
	case int:
		return []int{val}
	case float64:
		return []int{int(val)}
	case []any:
		out := make([]int, 0, len(val))
		for _, item := range val {
			switch n := item.(type) {
			case int:
				out = append(out, n)
			case float64:
				out = append(out, int(n))
			}
		}
		return out
	case []int:
		return val
	default:
		return nil
	}
}

// ExtractADRNumber parses the leading number from a filename like "0012-use-postgres.adr.md".
// Returns -1 if no number is found.
func ExtractADRNumber(filename string) int {
	base := filepath.Base(filename)
	m := adrNumberRe.FindStringSubmatch(base)
	if m == nil {
		return -1
	}
	n, err := strconv.Atoi(m[1])
	if err != nil {
		return -1
	}
	return n
}

// ValidSections checks whether a page with the given status has all required sections.
// Returns a list of missing section names. Only enforced for "accepted" status.
func ValidateSections(status string, nodes []core.Node) []string {
	if status != "accepted" {
		return nil
	}

	present := make(map[string]bool)
	for _, n := range nodes {
		present[n.Type] = true
	}

	var missing []string
	for _, req := range requiredSections {
		if !present[req] {
			missing = append(missing, req)
		}
	}
	return missing
}

// ValidStatuses is the set of allowed status values.
var ValidStatuses = map[string]bool{
	"tbd":        true,
	"proposed":   true,
	"accepted":   true,
	"deprecated": true,
	"superseded": true,
}
