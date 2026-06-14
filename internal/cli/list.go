package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"text/tabwriter"

	"github.com/gofuego/fuego-adr/adr"
	"github.com/spf13/cobra"
)

type adrEntry struct {
	Number int    `json:"adr"`
	Status string `json:"status"`
	Date   string `json:"date"`
	Title  string `json:"title"`
}

func newListCmd() *cobra.Command {
	var jsonOutput bool
	var quiet bool

	cmd := &cobra.Command{
		Use:   "list [adr-path]",
		Short: "List all ADRs",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			adrPath := "."
			if len(args) > 0 {
				adrPath = args[0]
			}
			return runList(adrPath, jsonOutput, quiet)
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "output as JSON")
	cmd.Flags().BoolVarP(&quiet, "quiet", "q", false, "output ADR numbers only")

	return cmd
}

func runList(adrPath string, jsonOutput, quiet bool) error {
	entries, err := parseAllADRs(adrPath)
	if err != nil {
		return err
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Number < entries[j].Number
	})

	if quiet {
		for _, e := range entries {
			fmt.Println(e.Number)
		}
		return nil
	}

	if jsonOutput {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(entries)
	}

	printADRTable(os.Stdout, entries)
	return nil
}

func printADRTable(w io.Writer, entries []adrEntry) {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "ADR\tSTATUS\tDATE\tTITLE")
	for _, e := range entries {
		fmt.Fprintf(tw, "%d\t%s\t%s\t%s\n", e.Number, e.Status, e.Date, e.Title)
	}
	tw.Flush()
}

// parseAllADRs reads and parses all .adr.md files in a directory.
func parseAllADRs(dir string) ([]adrEntry, error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("reading directory: %w", err)
	}

	p := adr.New()
	var entries []adrEntry

	for _, f := range files {
		if f.IsDir() || !adrFileRe.MatchString(f.Name()) {
			continue
		}

		raw, err := os.ReadFile(fmt.Sprintf("%s/%s", dir, f.Name()))
		if err != nil {
			fmt.Fprintf(os.Stderr, "fuego-adr: warning: could not read %s: %v\n", f.Name(), err)
			continue
		}

		env, _, err := p.Parse(raw)
		if err != nil {
			fmt.Fprintf(os.Stderr, "fuego-adr: warning: could not parse %s: %v\n", f.Name(), err)
			continue
		}

		num := adr.ExtractADRNumber(f.Name())
		title, _ := env["title"].(string)
		status, _ := env["status"].(string)

		date := ""
		if d, ok := env["date_accepted"].(string); ok && d != "" {
			date = d
		} else if d, ok := env["date_proposed"].(string); ok && d != "" {
			date = d
		}

		entries = append(entries, adrEntry{
			Number: num,
			Status: status,
			Date:   date,
			Title:  title,
		})
	}

	return entries, nil
}
