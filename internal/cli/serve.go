package cli

import (
	"context"

	"github.com/FabioSol/fuego-adr/adr"
	"github.com/FabioSol/fuego/engine"
	"github.com/spf13/cobra"
)

func newServeCmd() *cobra.Command {
	var baseURL string
	var port int

	cmd := &cobra.Command{
		Use:   "serve [adr-path]",
		Short: "Start development server with live reload",
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
			if baseURL != "" {
				cfg.BaseURL = baseURL
			}

			eng := engine.New()
			eng.Use(adr.Pack())
			return eng.Serve(context.Background(), engine.BuildOptions{
				ContentDir: adrPath,
				OutputDir:  cfg.OutputDir,
				SiteName:   cfg.SiteName,
				BaseURL:    cfg.BaseURL,
				DevPort:    port,
			})
		},
	}

	cmd.Flags().StringVar(&baseURL, "base-url", "", "base URL for deployment")
	cmd.Flags().IntVar(&port, "port", 8080, "dev server port")

	return cmd
}
