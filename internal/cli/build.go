package cli

import (
	"context"
	"os"
	"path/filepath"

	"github.com/FabioSol/fuego-adr/adr"
	adrconfig "github.com/FabioSol/fuego-adr/internal/config"
	"github.com/FabioSol/fuego/engine"
	"github.com/spf13/cobra"
)

func newBuildCmd() *cobra.Command {
	var outputDir string
	var baseURL string
	var incremental bool

	cmd := &cobra.Command{
		Use:   "build [adr-path]",
		Short: "Build ADR documentation site",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			adrPath := "."
			if len(args) > 0 {
				adrPath = args[0]
			}

			cfg, err := loadConfig(adrPath)
			if err != nil {
				return err
			}
			if outputDir != "" {
				cfg.OutputDir = outputDir
			}
			if baseURL != "" {
				cfg.BaseURL = baseURL
			}

			return buildSite(adrPath, cfg, incremental)
		},
	}

	cmd.Flags().StringVarP(&outputDir, "output", "o", "", "output directory (default: build)")
	cmd.Flags().StringVar(&baseURL, "base-url", "", "base URL for deployment (e.g. /my-repo)")
	cmd.Flags().BoolVar(&incremental, "incremental", false, "reuse cached parses for unchanged ADRs")

	return cmd
}

// buildSite assembles the engine with the ADR pack and runs a build. The pack
// supplies the parser, theme, routes, taxonomies, and hooks; the CLI supplies
// only the site-specific dirs and metadata.
func buildSite(adrPath string, cfg *adrconfig.Config, incremental bool) error {
	eng := engine.New()
	eng.Use(adr.Pack())
	return eng.Build(context.Background(), engine.BuildOptions{
		ContentDir:  adrPath,
		OutputDir:   cfg.OutputDir,
		SiteName:    cfg.SiteName,
		BaseURL:     cfg.BaseURL,
		Incremental: incremental,
	})
}

// loadConfig finds and loads fuego-adr.yaml from the ADR directory or its parent.
func loadConfig(adrPath string) (*adrconfig.Config, error) {
	candidates := []string{
		filepath.Join(adrPath, "fuego-adr.yaml"),
		filepath.Join(adrPath, "..", "fuego-adr.yaml"),
		"fuego-adr.yaml",
	}

	for _, path := range candidates {
		if _, err := os.Stat(path); err == nil {
			return adrconfig.Load(path)
		}
	}

	return adrconfig.Defaults(), nil
}
