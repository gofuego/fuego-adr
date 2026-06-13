package cli

import (
	"github.com/spf13/cobra"
)

// Version is set at build time via -ldflags.
var Version = "dev"

// Execute runs the fuego-adr CLI.
func Execute() error {
	root := newRootCmd()
	return root.Execute()
}

func newRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "fuego-adr",
		Short:         "Architecture Decision Record documentation generator",
		Version:       Version,
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	cmd.AddCommand(
		newBuildCmd(),
		newServeCmd(),
		newNewCmd(),
		newValidateCmd(),
		newListCmd(),
		newAffectedCmd(),
	)

	return cmd
}
