package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"text/tabwriter"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/gofuego/fuego-adr/adr"
	"github.com/spf13/cobra"
)

type affectedResult struct {
	ADR            int    `json:"adr"`
	Status         string `json:"status"`
	Title          string `json:"title"`
	MatchedPattern string `json:"matched_pattern"`
}

func newAffectedCmd() *cobra.Command {
	var adrPath string
	var jsonOutput bool
	var quiet bool

	cmd := &cobra.Command{
		Use:   "affected <file> [file...]",
		Short: "Find ADRs affecting given files",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAffected(adrPath, args, jsonOutput, quiet)
		},
	}

	cmd.Flags().StringVarP(&adrPath, "path", "p", ".", "directory containing ADR files")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "output as JSON")
	cmd.Flags().BoolVarP(&quiet, "quiet", "q", false, "output ADR numbers only")

	return cmd
}

func runAffected(adrPath string, files []string, jsonOutput, quiet bool) error {
	adrFiles, err := os.ReadDir(adrPath)
	if err != nil {
		return fmt.Errorf("reading ADR directory: %w", err)
	}

	p := adr.New()
	var results []affectedResult
	seen := make(map[int]bool)

	for _, f := range adrFiles {
		if f.IsDir() || !adrFileRe.MatchString(f.Name()) {
			continue
		}

		raw, err := os.ReadFile(filepath.Join(adrPath, f.Name()))
		if err != nil {
			continue
		}

		env, _, err := p.Parse(raw)
		if err != nil {
			continue
		}

		num := adr.ExtractADRNumber(f.Name())
		title, _ := env["title"].(string)
		status, _ := env["status"].(string)

		affects, _ := env["affects"].([]string)
		for _, pattern := range affects {
			for _, file := range files {
				matched, _ := doublestar.Match(pattern, file)
				if matched && !seen[num] {
					seen[num] = true
					results = append(results, affectedResult{
						ADR:            num,
						Status:         status,
						Title:          title,
						MatchedPattern: pattern,
					})
				}
			}
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].ADR < results[j].ADR
	})

	if len(results) == 0 {
		if !quiet {
			fmt.Fprintln(os.Stderr, "No ADRs affect the given files.")
		}
		return nil
	}

	if quiet {
		for _, r := range results {
			fmt.Println(r.ADR)
		}
		return nil
	}

	if jsonOutput {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(results)
	}

	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "ADR\tSTATUS\tTITLE\tPATTERN")
	for _, r := range results {
		fmt.Fprintf(tw, "%d\t%s\t%s\t%s\n", r.ADR, r.Status, r.Title, r.MatchedPattern)
	}
	tw.Flush()

	return nil
}
