package adr

import (
	"strings"
	"testing"

	"github.com/FabioSol/fuego/core"
)

func TestParse_BasicADR(t *testing.T) {
	raw := []byte(`---
title: Use PostgreSQL
status: Accepted
author: fabio
tags: [database, infrastructure]
supersedes: 3
---
## Context

We need a relational database for our application.

## Decision

We will use **PostgreSQL** for all persistent storage.

## Consequences

- Strong ecosystem
- Operational overhead
`)

	p := New()
	env, nodes, err := p.Parse(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Status normalized to lowercase
	if env["status"] != "accepted" {
		t.Errorf("status = %q, want %q", env["status"], "accepted")
	}

	// Author normalized to []string
	authors, ok := env["author"].([]string)
	if !ok {
		t.Fatalf("author type = %T, want []string", env["author"])
	}
	if len(authors) != 1 || authors[0] != "fabio" {
		t.Errorf("author = %v, want [fabio]", authors)
	}

	// Tags normalized to []string
	tags, ok := env["tags"].([]string)
	if !ok {
		t.Fatalf("tags type = %T, want []string", env["tags"])
	}
	if len(tags) != 2 || tags[0] != "database" {
		t.Errorf("tags = %v, want [database, infrastructure]", tags)
	}

	// Supersedes normalized to []int
	supersedes, ok := env["supersedes"].([]int)
	if !ok {
		t.Fatalf("supersedes type = %T, want []int", env["supersedes"])
	}
	if len(supersedes) != 1 || supersedes[0] != 3 {
		t.Errorf("supersedes = %v, want [3]", supersedes)
	}

	// Three sections parsed
	if len(nodes) != 3 {
		t.Fatalf("len(nodes) = %d, want 3", len(nodes))
	}

	wantTypes := []string{"context", "decision", "consequences"}
	for i, wt := range wantTypes {
		if nodes[i].Type != wt {
			t.Errorf("nodes[%d].Type = %q, want %q", i, nodes[i].Type, wt)
		}
		if !nodes[i].Raw {
			t.Errorf("nodes[%d].Raw = false, want true", i)
		}
	}

	// Check HTML rendering
	if !strings.Contains(nodes[1].Content, "<strong>PostgreSQL</strong>") {
		t.Errorf("decision content missing bold: %s", nodes[1].Content)
	}
}

func TestParse_MultipleAuthors(t *testing.T) {
	raw := []byte(`---
title: Test
status: tbd
author: [alice, bob]
date_proposed: 2026-06-11
---
## Context

Some context.
`)

	p := New()
	env, _, err := p.Parse(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	authors, ok := env["author"].([]string)
	if !ok {
		t.Fatalf("author type = %T, want []string", env["author"])
	}
	if len(authors) != 2 || authors[0] != "alice" || authors[1] != "bob" {
		t.Errorf("author = %v, want [alice, bob]", authors)
	}
}

func TestParse_SupersededBy(t *testing.T) {
	raw := []byte(`---
title: Old Decision
status: superseded
superseded_by: 12
---
## Context

Old context.
`)

	p := New()
	env, _, err := p.Parse(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	supersededBy, ok := env["superseded_by"].([]int)
	if !ok {
		t.Fatalf("superseded_by type = %T, want []int", env["superseded_by"])
	}
	if len(supersededBy) != 1 || supersededBy[0] != 12 {
		t.Errorf("superseded_by = %v, want [12]", supersededBy)
	}
}

func TestExtractADRNumber(t *testing.T) {
	tests := []struct {
		filename string
		want     int
	}{
		{"0012-use-postgres.adr.md", 12},
		{"001-initial.adr.md", 1},
		{"42-something.adr.md", 42},
		{"0001-first.adr.md", 1},
		{"no-number.adr.md", -1},
		{"path/to/0005-nested.adr.md", 5},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			got := ExtractADRNumber(tt.filename)
			if got != tt.want {
				t.Errorf("ExtractADRNumber(%q) = %d, want %d", tt.filename, got, tt.want)
			}
		})
	}
}

func TestValidateSections_Accepted(t *testing.T) {
	tests := []struct {
		name    string
		status  string
		types   []string
		wantLen int
	}{
		{"accepted with all sections", "accepted", []string{"context", "decision", "consequences"}, 0},
		{"accepted missing consequences", "accepted", []string{"context", "decision"}, 1},
		{"accepted missing all", "accepted", []string{}, 3},
		{"tbd missing sections is OK", "tbd", []string{}, 0},
		{"proposed missing sections is OK", "proposed", []string{"context"}, 0},
		{"deprecated missing sections is OK", "deprecated", []string{}, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var nodes []core.Node
			for _, typ := range tt.types {
				nodes = append(nodes, core.Node{Type: typ})
			}
			missing := ValidateSections(tt.status, nodes)
			if len(missing) != tt.wantLen {
				t.Errorf("missing = %v (len %d), want len %d", missing, len(missing), tt.wantLen)
			}
		})
	}
}

func TestSplitSections(t *testing.T) {
	body := []byte(`## Context

First section content.

## Decision

Second section with **bold**.

## Consequences

- Item 1
- Item 2

## Extra Section

Bonus content.
`)

	sections := splitSections(body)
	if len(sections) != 4 {
		t.Fatalf("len(sections) = %d, want 4", len(sections))
	}

	wantHeadings := []string{"context", "decision", "consequences", "extra section"}
	for i, wh := range wantHeadings {
		if sections[i].heading != wh {
			t.Errorf("sections[%d].heading = %q, want %q", i, sections[i].heading, wh)
		}
	}
}

func TestType(t *testing.T) {
	p := New()
	if p.Type() != "adr" {
		t.Errorf("Type() = %q, want %q", p.Type(), "adr")
	}
}

func TestFilenames(t *testing.T) {
	p := New()
	fns := p.Filenames()
	if len(fns) != 1 || fns[0] != "*.adr.md" {
		t.Errorf("Filenames() = %v, want [*.adr.md]", fns)
	}
}
