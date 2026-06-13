package cli

import (
	"context"
	"fmt"

	"github.com/FabioSol/fuego-adr/adr"
	"github.com/FabioSol/fuego/engine"
	"github.com/spf13/cobra"
)

func newValidateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "validate [adr-path]",
		Short: "Validate ADRs without building (CI gate)",
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

			eng := engine.New()
			eng.Use(adr.Pack())
			n, err := eng.Validate(context.Background(), engine.BuildOptions{
				ContentDir: adrPath,
				OutputDir:  cfg.OutputDir,
				SiteName:   cfg.SiteName,
				BaseURL:    cfg.BaseURL,
			})
			if err != nil {
				return err
			}
			fmt.Printf("fuego-adr: %d pages validated successfully\n", n)
			return nil
		},
	}
}
