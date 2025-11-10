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
	fmt.Println("=== PR analysis for multiple repositories ===")

	allResults := make(map[string]analyzer.RepositoryResult)

	for i, repo := range cfg.Repositories {
		fmt.Printf("\n=== Repository %d/%d: %s/%s ===\n",
			i+1, len(cfg.Repositories), repo.Owner, repo.Repo)

		prs, err := m.GithubClient.GetAllPullRequests(repo.Owner, repo.Repo)
		if err != nil {
			log.Fatalf("Error when receiving PR: %v", err)
		}

		fmt.Printf("Found %d pull requests\n", len(prs))

		if len(prs) == 0 {
			fmt.Println("No PR was found for analysis.")
			return errors.New("no PR found for analysis.")
		}

		metrics, err := analyzer.CollectPRMetrics(m.GithubClient, repo.Owner, repo.Repo, prs)
		if err != nil {
			log.Fatalf("Error when collecting metrics: %v", err)
		}

		result := analyzer.AnalyzeData(metrics)

		repoKey := fmt.Sprintf("%s/%s", repo.Owner, repo.Repo)
		allResults[repoKey] = analyzer.RepositoryResult{
			Owner:    repo.Owner,
			Repo:     repo.Repo,
			PRCount:  len(prs),
			Metrics:  metrics,
			Analysis: result,
		}

		// -------------------------------------------------

		vmMetrics := &vmdb.Metrics{}

		// 1. Общее время жизни PR
		vmMetrics.AddPRMetric("PRLifetime", repoKey, result.MedianLifetime/time.Second, uint64(time.Now().UnixMilli()))

		// 2. Время до первого ответа
		vmMetrics.AddPRMetric("TimeToFirstReview", repoKey, result.AverageTimeToFirstReview/time.Second, uint64(time.Now().UnixMilli()))

		// 3. Процент успешных мержей
		vmMetrics.AddPRMetric("MergeSuccessRate", repoKey, result.MergeRate, uint64(time.Now().UnixMilli()))

		// 4. Прогнозное время до мержа нового PR
		vmMetrics.AddPRMetric("PredictedMergeTime", repoKey, result.PredictedTimeToMerge/time.Second, uint64(time.Now().UnixMilli()))

		err = m.VMDBExporter.PushMetrics(vmMetrics)
		if err != nil {
			return fmt.Errorf("%w: %w", "xueta", err)
		}

		// -------------------------------------------------

		// analyzer.PrintAnalysisResults(repo.Owner, repo.Repo, result)

		if i < len(cfg.Repositories)-1 {
			fmt.Printf("Waiting for the next repository...\n")
			time.Sleep(2 * time.Second)
		}
	}

	// comparative := analyzer.ComparativeAnalysis(allResults)

	// analyzer.PrintComparativeAnalysis(comparative)

	// if err := analyzer.SaveAllData(allResults); err != nil {
	// 	log.Printf("Error saving data: %v", err)
	// }

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
