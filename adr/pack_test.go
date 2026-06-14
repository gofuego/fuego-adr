package adr_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gofuego/fuego-adr/adr"
	"github.com/gofuego/fuego/engine"
)

// TestPackOnVanillaEngine is the #15 acceptance: a plain Fuego engine with only
// eng.Use(adr.Pack()) and some .adr.md content renders a complete ADR site —
// no fuego-adr CLI, no extra registration.
func TestPackOnVanillaEngine(t *testing.T) {
	dir := t.TempDir()
	content := filepath.Join(dir, "adr")
	out := filepath.Join(dir, "out")

	write := func(name, body string) {
		p := filepath.Join(content, name)
		if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(body), 0644); err != nil {
			t.Fatal(err)
		}
	}
	write("001-use-postgres.adr.md", `---
title: Use PostgreSQL
status: accepted
date_accepted: 2026-01-15
tags: [database]
affects:
  - src/db/**
---

## Context
Need a database.

## Decision
PostgreSQL.

## Consequences
Strong SQL.
`)

	eng := engine.New()
	eng.Use(adr.Pack())
	if err := eng.Build(context.Background(), engine.BuildOptions{
		ContentDir: content,
		OutputDir:  out,
		SiteName:   "Decisions",
	}); err != nil {
		t.Fatalf("vanilla pack build failed: %v", err)
	}

	want := []string{
		"decisions/001-use-postgres.adr/index.html", // routed via pack ConfigDefaults
		"index.html",          // homepage virtual page (Index hook)
		"timeline/index.html", // timeline virtual page
		"affects/index.html",  // affects-index virtual page
		"tags/database/index.html", // pack taxonomy
		"style.css",           // pack static asset
	}
	for _, rel := range want {
		if _, err := os.Stat(filepath.Join(out, rel)); err != nil {
			t.Errorf("expected output %s: %v", rel, err)
		}
	}

	home, err := os.ReadFile(filepath.Join(out, "index.html"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(home), "/decisions/001-use-postgres.adr/") {
		t.Error("homepage should link the ADR decision page")
	}
}
