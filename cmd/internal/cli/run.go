package cli

import (
	_ "embed"
	"errors"
	"fmt"
	"log"

	"github.com/spf13/cobra"

	"metrics-scrapper/internal/analyzer"
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
	fmt.Println("=== Gofish PR project analysis ===")

	client := github.NewClient(cfg)

	prs, err := client.GetAllPullRequests()
	if err != nil {
		log.Fatalf("Error when receiving PR: %v", err)
	}

	fmt.Printf("Found %d pull requests\n", len(prs))

	if len(prs) == 0 {
		fmt.Println("No PR was found for analysis.")
		return errors.New("no PR found for analysis.")
	}

	metrics, err := analyzer.CollectPRMetrics(client, prs)
	if err != nil {
		log.Fatalf("Error when collecting metrics: %v", err)
	}

	result := analyzer.AnalyzeData(metrics)

	analyzer.PrintAnalysisResults(result)

	if err := analyzer.SaveRawData(metrics); err != nil {
		log.Printf("Error saving data: %v", err)
	}

	return nil
}
