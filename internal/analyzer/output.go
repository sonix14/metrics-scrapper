package analyzer

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"
)

func PrintAnalysisResults(owner, repo string, result AnalysisResult) {
	fmt.Printf("\n--- Results for %s/%s ---\n", owner, repo)
	fmt.Printf("Total number of PR: %d\n", result.TotalPRs)
	fmt.Printf("Successfully merged: %d (%.1f%%)\n", result.MergedPRs, result.MergeRate)
	fmt.Printf("Closed without a merge: %d\n", result.ClosedPRs)
	fmt.Printf("Average PR lifetime: %v\n", result.AverageLifetime.Round(time.Hour))
	fmt.Printf("Median lifetime of a PR: %v\n", result.MedianLifetime.Round(time.Hour))
	fmt.Printf("Average time to the first response: %v\n", result.AverageTimeToFirstReview.Round(time.Hour))
	fmt.Printf("Median time to the first response: %v\n", result.MedianTimeToFirstReview.Round(time.Hour))

	authorStats := make(map[string]int)
	reviewerStats := make(map[string]int)
	authorReviewerPairs := make(map[string]map[string]int)

	for _, m := range result.PRMetrics {
		authorStats[m.Author]++
		for _, reviewer := range m.Reviewers {
			reviewerStats[reviewer]++
			if authorReviewerPairs[m.Author] == nil {
				authorReviewerPairs[m.Author] = make(map[string]int)
			}
			authorReviewerPairs[m.Author][reviewer]++
		}
	}

	printAuthorStats(authorStats)
	printReviewerStats(reviewerStats)
	printAuthorReviewerPairs(authorReviewerPairs)
	printPredictions(result)
	printRecommendations(result)
}

func printAuthorStats(stats map[string]int) {
	fmt.Printf("\n=== AUTHOR STATISTICS ===\n")
	for author, count := range stats {
		fmt.Printf("  %s: %d PR\n", author, count)
	}
}

func printReviewerStats(stats map[string]int) {
	fmt.Printf("\n=== REVIEWERS' STATISTICS ===\n")
	for reviewer, count := range stats {
		fmt.Printf("  %s: %d review\n", reviewer, count)
	}
}

func printAuthorReviewerPairs(pairs map[string]map[string]int) {
	fmt.Printf("\n=== EFFECTIVE PAIRS OF AUTHOR-REVIEWER ===\n")
	foundPairs := false
	for author, reviewers := range pairs {
		for reviewer, count := range reviewers {
			if count >= 2 {
				fmt.Printf("  %s ↔ %s: %d collaboration\n", author, reviewer, count)
				foundPairs = true
			}
		}
	}
	if !foundPairs {
		fmt.Printf("  No pairs with 2+ interactions found\n")
	}
}

func printPredictions(result AnalysisResult) {
	fmt.Printf("\n=== PROGNOSIS FOR THE NEW PR ===\n")
	fmt.Printf("Expected time before merge: %v\n", result.MedianLifetime.Round(time.Hour*24))
	fmt.Printf("Expected time until the first response: %v\n", result.MedianTimeToFirstReview.Round(time.Hour))
}

func printRecommendations(result AnalysisResult) {
	fmt.Printf("\n=== RECOMMENDATIONS ===\n")
	if result.MergeRate < 50 {
		fmt.Printf("⚠️  Low percentage of merge (%0.1f%%) - there may be problems with code quality or planning\n", result.MergeRate)
	} else {
		fmt.Printf("✅ A good percentage of merge (%0.1f%%)\n", result.MergeRate)
	}

	if result.AverageTimeToFirstReview > 7*24*time.Hour {
		fmt.Printf("⚠️  Long time to first response (%v) - possible lack of reviewers\n", result.AverageTimeToFirstReview.Round(time.Hour))
	} else {
		fmt.Printf("✅ Good team response time\n")
	}
}

func SaveRawData(metrics []PRMetrics) error {
	file, err := os.Create("gofish_pr_data.json")
	if err != nil {
		return fmt.Errorf("error when creating the data file: %v", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(metrics); err != nil {
		return fmt.Errorf("error saving data: %v", err)
	}

	fmt.Printf("\nThe raw data is saved to a file: gofish_pr_data.json\n")
	return nil
}

func PrintComparativeAnalysis(comparative ComparativeAnalyser) {
	fmt.Printf("\n" + strings.Repeat("=", 60) + "\n")
	fmt.Printf("COMPARATIVE ANALYSIS OF REPOSITORIES\n")
	fmt.Printf(strings.Repeat("=", 60) + "\n")

	fmt.Printf("\n📊 GENERAL STATISTICS:\n")
	fmt.Printf("   Total repositoriesв: %d\n", comparative.Summary.TotalRepositories)
	fmt.Printf("   Total PR: %d\n", comparative.Summary.TotalPRs)
	fmt.Printf("   Everything is confused: %d\n", comparative.Summary.TotalMergedPRs)
	fmt.Printf("   The average percentage of merge: %.1f%%\n", comparative.Summary.AvgMergeRate)

	if len(comparative.Summary.BestPerforming) > 0 {
		fmt.Printf("\n🏆 TOP-3 REPOSITORIES IN TERMS OF EFFECTIVENESS:\n")
		for i, repo := range comparative.Summary.BestPerforming {
			fmt.Printf("   %d. %s - %.1f%% merge, average time: %v\n",
				i+1, repo.Repository, repo.MergeRate, repo.AvgTime.Round(time.Hour*24))
		}
	}

	fmt.Printf("\n📈 DETAILED STATISTICS ON REPOSITORIES:\n")
	fmt.Printf("   %-30s %-8s %-8s %-12s %-15s\n",
		"Repository", "PR", "Merge%", "Wed. time", "Response")
	fmt.Printf("   %s\n", strings.Repeat("-", 80))

	var repoKeys []string
	for key := range comparative.RepositoryResults {
		repoKeys = append(repoKeys, key)
	}
	sort.Strings(repoKeys)

	for _, key := range repoKeys {
		result := comparative.RepositoryResults[key]
		analysis := result.Analysis
		fmt.Printf("   %-30s %-8d %-8.1f %-12v %-15v\n",
			key,
			analysis.TotalPRs,
			analysis.MergeRate,
			analysis.AverageLifetime.Round(time.Hour*24),
			analysis.AverageTimeToFirstReview.Round(time.Hour),
		)
	}
}

func SaveAllData(results map[string]RepositoryResult) error {
	for repoKey, result := range results {
		filename := fmt.Sprintf("metrics_%s.json", sanitizeFilename(repoKey))
		file, err := os.Create(filename)
		if err != nil {
			return fmt.Errorf("error when creating the file %s: %v", filename, err)
		}
		defer file.Close()

		encoder := json.NewEncoder(file)
		encoder.SetIndent("", "  ")

		if err := encoder.Encode(result); err != nil {
			return fmt.Errorf("error when saving data in %s: %v", filename, err)
		}
		fmt.Printf("The data is saved to a file: %s\n", filename)
	}

	comparativeFile, err := os.Create("comparative_analysis.json")
	if err != nil {
		return fmt.Errorf("error when creating the file comparative_analysis.json: %v", err)
	}
	defer comparativeFile.Close()

	comparative := ComparativeAnalysis(results)
	encoder := json.NewEncoder(comparativeFile)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(comparative); err != nil {
		return fmt.Errorf("error saving the comparative analysis: %v", err)
	}

	fmt.Printf("The comparative analysis is saved to a file: comparative_analysis.json\n")
	return nil
}

func sanitizeFilename(name string) string {
	return strings.ReplaceAll(name, "/", "_")
}
