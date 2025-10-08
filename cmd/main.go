package main

import (
	"fmt"
	"log"
	"metrics-scrapper/internal/analyzer"
	"metrics-scrapper/internal/config"
	"metrics-scrapper/internal/github"
)

func main() {
	fmt.Println("=== Gofish PR project analysis ===\n")

	cfg := config.LoadConfig()

	client := github.NewClient(cfg)

	prs, err := client.GetAllPullRequests()
	if err != nil {
		log.Fatalf("Error when receiving PR: %v", err)
	}

	fmt.Printf("Found %d pull requests\n", len(prs))

	if len(prs) == 0 {
		fmt.Println("No PR was found for analysis.")
		return
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
}
