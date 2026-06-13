// Package adr is the fuego-adr format pack: it turns *.adr.md files into an
// Architecture Decision Record documentation site. Register it on any Fuego
// engine with eng.Use(adr.Pack()) — it brings its own parser, theme, routes,
// taxonomies, and the enrichment/virtual-page hooks, so a vanilla Fuego
// project needs only content and one line of wiring.
package adr

import (
	"embed"
	"io/fs"

	"github.com/FabioSol/fuego/core"
)

//go:embed theme
var themeFS embed.FS

//go:embed config-defaults.yaml
var configDefaults []byte

// Pack returns the fuego-adr format pack.
func Pack() core.Pack {
	theme, _ := fs.Sub(themeFS, "theme")
	return core.Pack{
		Name:           "adr",
		Parsers:        []core.Parser{New()},
		Theme:          theme,
		ConfigDefaults: configDefaults,
		Hooks: core.Hooks{
			AfterParse: []core.AfterParseHook{AfterParseHook()},
			Index:      []core.IndexHook{IndexHook()},
		},
	}
}
