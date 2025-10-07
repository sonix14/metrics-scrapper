package analyzer

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

func PrintAnalysisResults(result AnalysisResult) {
	fmt.Printf("\n=== ANALYSIS RESULTS ===\n")
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
