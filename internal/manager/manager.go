package manager

import (
	"errors"
	"fmt"
	"log"
	"metrics-scrapper/internal/analyzer"
	"metrics-scrapper/internal/config"
	"metrics-scrapper/internal/github"
	"metrics-scrapper/internal/vmdb"
	"time"
)

const Name = "METRIC_ANALYSIS_RESULT"

type MetricManager struct {
	GithubClient *github.Client
	VMDBExporter VMDBExporter
}

func NewMetricManager(
	vmdbExporter VMDBExporter,
	githubClient *github.Client,
) *MetricManager {
	newManager := &MetricManager{ //nolint:exhaustruct
		VMDBExporter: vmdbExporter,
		GithubClient: githubClient,
	}

	return newManager
}

func (m *MetricManager) ScrapeAndPush(cfg *config.Config) error {
	fmt.Println("=== Gofish PR project analysis ===")

	prs, err := m.GithubClient.GetAllPullRequests()
	if err != nil {
		log.Fatalf("Error when receiving PR: %v", err)
	}

	fmt.Printf("Found %d pull requests\n", len(prs))

	if len(prs) == 0 {
		fmt.Println("No PR was found for analysis.")
		return errors.New("no PR found for analysis.")
	}

	metrics, err := analyzer.CollectPRMetrics(m.GithubClient, prs)
	if err != nil {
		log.Fatalf("Error when collecting metrics: %v", err)
	}

	result := analyzer.AnalyzeData(metrics)

	vmMetrics := &vmdb.Metrics{}

	vmMetrics.AddPRMetric("MergeRate", "gofish", result.MergeRate, uint64(time.Now().UnixMilli()))
	vmMetrics.AddPRMetric("AverageTimeToFirstReview", "gofish", result.AverageTimeToFirstReview/time.Second, uint64(time.Now().UnixMilli()))
	vmMetrics.AddPRMetric("MedianLifeTime", "gofish", result.MedianLifetime/time.Second, uint64(time.Now().UnixMilli()))

	vmMetrics.AddPRMetric("MergeRate", "project1", result.MergeRate, uint64(time.Now().UnixMilli()))
	vmMetrics.AddPRMetric("AverageTimeToFirstReview", "project1", result.AverageTimeToFirstReview/time.Second, uint64(time.Now().UnixMilli()))
	vmMetrics.AddPRMetric("MedianLifeTime", "project1", result.MedianLifetime/time.Second, uint64(time.Now().UnixMilli()))

	err = m.VMDBExporter.PushMetrics(vmMetrics)
	if err != nil {
		return fmt.Errorf("%w: %w", "xueta", err)
	}

	return nil
}

//func (m *MetricManager) runScraper(
//	repo string,
//	projectKey string,
//	scrapeFrom time.Time,
//) error {
//	collection := &collection.ScrapedCollection{RepoName: repo} //nolint:exhaustruct
//
//	for _, scraper := range m.Scrapers {
//		m.Logger.Info(fmt.Sprintf("%q started", scraper.GetName()), "repo", repo)
//
//		err := scraper.Scrape(projectKey, collection, scrapeFrom)
//		if err != nil {
//			return fmt.Errorf("%q, %q, %w: %w", scraper.GetName(), repo, ErrScraperFailed, err)
//		}
//
//		m.Logger.Info(fmt.Sprintf("%q finished", scraper.GetName()), "repo", repo)
//	}
//
//	err := m.Push(collection)
//	if err != nil {
//		return fmt.Errorf("%q, %w: %w", repo, ErrPushingMetrics, err)
//	}
//
//	return nil
//}
//
//func (m *MetricManager) Push(collection *collection.ScrapedCollection) error {
//	m.Logger.Info("exporters started", "repo", collection.RepoName)
//
//	metrics := &vmdb.Metrics{} //nolint:exhaustruct
//
//	for _, exporter := range m.Exporters {
//		exporter.Export(collection, metrics)
//	}
//
//	m.Logger.Info("push metrics", "repo", collection.RepoName)
//
//	err := m.VMDBExporter.PushMetrics(metrics)
//	if err != nil {
//		return fmt.Errorf("%w: %w", ErrPushingMetrics, err)
//	}
//
//	return nil
//}
