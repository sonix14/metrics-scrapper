package cli

import (
	_ "embed"
	manager2 "metrics-scrapper/internal/manager"
	"metrics-scrapper/internal/vmdb"
	"net/http"

	"github.com/spf13/cobra"
	"golang.org/x/exp/slog"

	"metrics-scrapper/internal/github"
)

//go:embed data/run_desc.md
var runCmdDesc string

const scrapeThresholdFlag = "scrape-threshold"

func newRunCmd() *cobra.Command {
	runCmd := &cobra.Command{ //nolint:exhaustruct
		Use:   "run",
		Short: "Run metrics-scraper",
		Long:  runCmdDesc,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd)
		},
	}

	// var timestampFlag timestamp.Timestamp

	// runCmd.PersistentFlags().Var(
	// 	&timestampFlag,
	// 	scrapeThresholdFlag,
	// 	"timestamp in format \"YYYY-MM-DD HH:MM:SS\". PRs updated earlier then this timestamp will not be scraped",
	// )

	return runCmd
}

// Entry point of RunCmd (i.e. `metrics-scraper run`).
func run(_ *cobra.Command) error {
	var vmdbURLProvider vmdb.URLProvider

	vmdbURLProvider = &vmdb.SingleNodeURL{
		VMURL: "http://localhost:8428",
	}

	logger := slog.Default()

	client := github.NewClient(cfg)

	exporter := vmdb.NewVMDBExporter(
		&http.Client{}, //nolint:exhaustruct
		vmdbURLProvider,
		logger,
		"1y",
	)

	manager := manager2.NewMetricManager(exporter, client)

	err := manager.ScrapeAndPush(cfg)

	return err
}
