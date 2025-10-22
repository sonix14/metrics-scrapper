package cli

import (
	_ "embed"
	"log/slog"

	"github.com/spf13/cobra"

	"metrics-scrapper/internal/config"
)

const AppTitle = "Metrics Scraper"

//go:embed data/root_desc.md
var rootCmdDesc string

var (
	logger *slog.Logger
	cfg    *config.Config
)

func setupConfig(_ *cobra.Command, _ []string) error {
	cfg = config.LoadConfig()

	return nil
}

func NewRootCmd() (*cobra.Command, error) {
	//nolint:exhaustruct
	rootCmd := &cobra.Command{
		Use:               "metrics-scraper",
		Short:             AppTitle,
		Long:              rootCmdDesc,
		Version:           "v1",
		CompletionOptions: cobra.CompletionOptions{DisableDefaultCmd: true}, //nolint:exhaustruct
		DisableAutoGenTag: true,

		PersistentPreRunE: setupConfig,
	}

	rootCmd.AddCommand(
		newRunCmd(),
	)

	return rootCmd, nil
}
