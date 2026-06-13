package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var adrFileRe = regexp.MustCompile(`^(\d+)-.+\.adr\.md$`)

func newNewCmd() *cobra.Command {
	var adrPath string

	cmd := &cobra.Command{
		Use:   "new [title]",
		Short: "Create a new ADR from template",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			title := strings.Join(args, " ")
			return runNewADR(adrPath, title)
		},
	}

	cmd.Flags().StringVarP(&adrPath, "path", "p", ".", "directory containing ADR files")

	return cmd
}

func runNewADR(adrPath, title string) error {
	// Scan existing ADRs to find max number and detect padding width
	maxNum, padWidth, err := scanExistingADRs(adrPath)
	if err != nil {
		return fmt.Errorf("scanning ADRs: %w", err)
	}

	nextNum := maxNum + 1

	// Detect author from git config
	author := detectGitAuthor()

	// Generate filename
	slug := slugify(title)
	numStr := fmt.Sprintf("%0*d", padWidth, nextNum)
	filename := fmt.Sprintf("%s-%s.adr.md", numStr, slug)
	filePath := filepath.Join(adrPath, filename)

	// Generate template
	content := generateTemplate(title, author)

	if err := os.MkdirAll(adrPath, 0755); err != nil {
		return fmt.Errorf("creating directory: %w", err)
	}

	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return fmt.Errorf("writing file: %w", err)
	}

	fmt.Printf("Created %s\n", filePath)
	return nil
}

func scanExistingADRs(dir string) (maxNum, padWidth int, err error) {
	padWidth = 3 // default

	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, padWidth, nil
		}
		return 0, 0, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		m := adrFileRe.FindStringSubmatch(entry.Name())
		if m == nil {
			continue
		}

		numStr := m[1]
		n, _ := strconv.Atoi(numStr)
		if n > maxNum {
			maxNum = n
		}

		// Detect padding width from existing files
		if len(numStr) > padWidth {
			padWidth = len(numStr)
		}
	}

	return maxNum, padWidth, nil
}

func slugify(title string) string {
	s := strings.ToLower(title)
	// Replace non-alphanumeric chars with hyphens
	var result []rune
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			result = append(result, r)
		} else if r == ' ' || r == '_' || r == '-' {
			if len(result) > 0 && result[len(result)-1] != '-' {
				result = append(result, '-')
			}
		}
	}
	return strings.Trim(string(result), "-")
}

func detectGitAuthor() string {
	// Best-effort from git config
	return os.Getenv("USER")
}

func generateTemplate(title, author string) string {
	today := time.Now().Format("2006-01-02")

	return fmt.Sprintf(`---
title: %s
status: tbd
date_proposed: %s
author: %s
# approvers: []
# deadline: YYYY-MM-DD
# supersedes: []
# superseded_by: []
# tags: []
# affects: []
---

## Context

## Decision

## Consequences
`, title, today, author)
}
