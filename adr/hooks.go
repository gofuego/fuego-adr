package adr

import (
	"fmt"
	"sort"
	"strings"

	"github.com/gofuego/fuego/core"
)

// adrSummary is a lightweight representation of an ADR for virtual pages.
type adrSummary struct {
	Number     int
	Title      string
	Status     string
	URL        string
	Date       string
	Deadline   string
	Tags       []string
	Supersedes []int
}

// AfterParseHook enriches ADR pages with extracted numbers, validates
// supersession consistency, and enforces section completeness.
// Virtual page generation happens in IndexHook.
func AfterParseHook() core.AfterParseHook {
	return func(pages []*core.Page) ([]*core.Page, error) {
		adrPages := enrichADRNumbers(pages)

		if err := validateSupersession(adrPages); err != nil {
			return nil, err
		}

		validateSectionCompleteness(adrPages)

		for _, p := range adrPages {
			if p.Layout == "" {
				p.Layout = "adr"
			}
		}

		return pages, nil
	}
}

// IndexHook generates the virtual pages (homepage, timeline, affects-index)
// during INDEX, where ROUTE has resolved real-page URLs and the added pages
// are collision-checked alongside taxonomy pages.
func IndexHook() core.IndexHook {
	return func(pages []*core.Page) ([]*core.Page, error) {
		var adrPages []*core.Page
		for _, p := range pages {
			if p.Type == "adr" {
				adrPages = append(adrPages, p)
			}
		}

		summaries := buildSummaries(adrPages)
		pages = append(pages, generateVirtualPages(summaries, adrPages)...)

		return pages, nil
	}
}

// enrichADRNumbers extracts ADR numbers from filenames and stores them
// in the envelope. Returns only ADR-type pages.
func enrichADRNumbers(pages []*core.Page) []*core.Page {
	var adrPages []*core.Page
	for _, p := range pages {
		if p.Type != "adr" {
			continue
		}
		adrPages = append(adrPages, p)

		num := ExtractADRNumber(p.RelPath)
		if num >= 0 {
			p.Envelope["adr_number"] = num
		}
	}
	return adrPages
}

// validateSupersession checks bidirectional consistency of supersedes/superseded_by.
func validateSupersession(pages []*core.Page) error {
	byNumber := make(map[int]*core.Page)
	for _, p := range pages {
		if num, ok := p.Envelope["adr_number"].(int); ok {
			byNumber[num] = p
		}
	}

	var errors []string
	for _, p := range pages {
		num, ok := p.Envelope["adr_number"].(int)
		if !ok {
			continue
		}

		if supersedes, ok := p.Envelope["supersedes"].([]int); ok {
			for _, targetNum := range supersedes {
				target, exists := byNumber[targetNum]
				if !exists {
					errors = append(errors, fmt.Sprintf("ADR-%d supersedes ADR-%d, but ADR-%d does not exist", num, targetNum, targetNum))
					continue
				}
				if !intSliceContains(getIntSlice(target.Envelope, "superseded_by"), num) {
					errors = append(errors, fmt.Sprintf("ADR-%d supersedes ADR-%d, but ADR-%d does not declare superseded_by: %d", num, targetNum, targetNum, num))
				}
			}
		}

		if supersededBy, ok := p.Envelope["superseded_by"].([]int); ok {
			for _, targetNum := range supersededBy {
				target, exists := byNumber[targetNum]
				if !exists {
					errors = append(errors, fmt.Sprintf("ADR-%d is superseded_by ADR-%d, but ADR-%d does not exist", num, targetNum, targetNum))
					continue
				}
				if !intSliceContains(getIntSlice(target.Envelope, "supersedes"), num) {
					errors = append(errors, fmt.Sprintf("ADR-%d is superseded_by ADR-%d, but ADR-%d does not declare supersedes: %d", num, targetNum, targetNum, num))
				}
			}
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("supersession validation failed:\n  %s", strings.Join(errors, "\n  "))
	}
	return nil
}

func validateSectionCompleteness(pages []*core.Page) {
	for _, p := range pages {
		status, _ := p.Envelope["status"].(string)
		missing := ValidateSections(status, p.Nodes)
		if len(missing) > 0 {
			num := getInt(p.Envelope, "adr_number", -1)
			fmt.Printf("fuego-adr: warning: ADR-%d (%s) missing required sections: %s\n", num, status, strings.Join(missing, ", "))
		}
	}
}

func buildSummaries(pages []*core.Page) []adrSummary {
	var summaries []adrSummary
	for _, p := range pages {
		s := adrSummary{
			Number:     getInt(p.Envelope, "adr_number", -1),
			Title:      getString(p.Envelope, "title"),
			Status:     getString(p.Envelope, "status"),
			URL:        p.URL,
			Tags:       getStringSlice(p.Envelope, "tags"),
			Supersedes: getIntSlice(p.Envelope, "supersedes"),
		}

		if d := getString(p.Envelope, "date_accepted"); d != "" {
			s.Date = d
		} else if d := getString(p.Envelope, "date_proposed"); d != "" {
			s.Date = d
		}

		if d := getString(p.Envelope, "deadline"); d != "" {
			s.Deadline = d
		}

		summaries = append(summaries, s)
	}

	sort.Slice(summaries, func(i, j int) bool {
		return summaries[i].Number < summaries[j].Number
	})

	return summaries
}

func generateVirtualPages(summaries []adrSummary, adrPages []*core.Page) []*core.Page {
	var pages []*core.Page
	pages = append(pages, generateHomepage(summaries))
	pages = append(pages, generateTimeline(summaries))
	pages = append(pages, generateAffectsIndex(adrPages))
	return pages
}

func generateHomepage(summaries []adrSummary) *core.Page {
	var tbd, proposed, accepted, inactive []map[string]any

	for _, s := range summaries {
		entry := map[string]any{
			"number": s.Number,
			"title":  s.Title,
			"status": s.Status,
			"url":    s.URL,
			"tags":   s.Tags,
		}
		switch s.Status {
		case "tbd":
			if s.Deadline != "" {
				entry["deadline"] = s.Deadline
			}
			tbd = append(tbd, entry)
		case "proposed":
			proposed = append(proposed, entry)
		case "accepted":
			accepted = append(accepted, entry)
		case "deprecated", "superseded":
			inactive = append(inactive, entry)
		}
	}

	return &core.Page{
		RelPath: "_virtual/homepage",
		Type:    "virtual",
		URL:     "/",
		Layout:  "homepage",
		Envelope: core.Envelope{
			"title":         "Home",
			"tbd_adrs":      tbd,
			"proposed_adrs": proposed,
			"accepted_adrs": accepted,
			"inactive_adrs": inactive,
		},
	}
}

func generateTimeline(summaries []adrSummary) *core.Page {
	sorted := make([]adrSummary, len(summaries))
	copy(sorted, summaries)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Date > sorted[j].Date
	})

	var entries []map[string]any
	for _, s := range sorted {
		entries = append(entries, map[string]any{
			"number":     s.Number,
			"title":      s.Title,
			"status":     s.Status,
			"url":        s.URL,
			"date":       s.Date,
			"supersedes": s.Supersedes,
		})
	}

	return &core.Page{
		RelPath: "_virtual/timeline",
		Type:    "virtual",
		URL:     "/timeline/",
		Layout:  "timeline",
		Envelope: core.Envelope{
			"title":            "Timeline",
			"timeline_entries": entries,
		},
	}
}

func generateAffectsIndex(adrPages []*core.Page) *core.Page {
	patternMap := make(map[string][]map[string]any)
	for _, p := range adrPages {
		affects := getStringSlice(p.Envelope, "affects")
		num := getInt(p.Envelope, "adr_number", -1)
		title := getString(p.Envelope, "title")

		for _, pattern := range affects {
			patternMap[pattern] = append(patternMap[pattern], map[string]any{
				"number": num,
				"title":  title,
				"url":    p.URL,
			})
		}
	}

	var patterns []string
	for p := range patternMap {
		patterns = append(patterns, p)
	}
	sort.Strings(patterns)

	var entries []map[string]any
	for _, pattern := range patterns {
		entries = append(entries, map[string]any{
			"pattern": pattern,
			"adrs":    patternMap[pattern],
			"exists":  true,
		})
	}

	return &core.Page{
		RelPath: "_virtual/affects-index",
		Type:    "virtual",
		URL:     "/affects/",
		Layout:  "affects-index",
		Envelope: core.Envelope{
			"title":           "Affected Files",
			"affects_entries": entries,
		},
	}
}

func getString(env core.Envelope, key string) string {
	if v, ok := env[key].(string); ok {
		return v
	}
	return ""
}

func getInt(env core.Envelope, key string, def int) int {
	switch v := env[key].(type) {
	case int:
		return v
	case float64:
		return int(v)
	default:
		return def
	}
}

func getStringSlice(env core.Envelope, key string) []string {
	switch v := env[key].(type) {
	case []string:
		return v
	case []any:
		var out []string
		for _, item := range v {
			if s, ok := item.(string); ok {
				out = append(out, s)
			}
		}
		return out
	default:
		return nil
	}
}

func getIntSlice(env core.Envelope, key string) []int {
	switch v := env[key].(type) {
	case []int:
		return v
	case []any:
		var out []int
		for _, item := range v {
			switch n := item.(type) {
			case int:
				out = append(out, n)
			case float64:
				out = append(out, int(n))
			}
		}
		return out
	default:
		return nil
	}
}

func intSliceContains(slice []int, val int) bool {
	for _, v := range slice {
		if v == val {
			return true
		}
	}
	return false
}
