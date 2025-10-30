package analyzer

import (
	"fmt"
	"log"
	"metrics-scrapper/internal/github"
	"sort"
	"time"
)

func CollectPRMetrics(client *github.Client, owner, repo string, prs []github.PullRequest) ([]PRMetrics, error) {
	var metrics []PRMetrics

	for i, pr := range prs {
		fmt.Printf("PR processing #%d (%d/%d)\n", pr.Number, i+1, len(prs))

		prMetrics, err := collectMetricsForPR(client, owner, repo, pr)
		if err != nil {
			log.Printf("Error when getting metrics for PR #%d: %v", pr.Number, err)
			continue
		}

		metrics = append(metrics, prMetrics)
		time.Sleep(500 * time.Millisecond) // delay between requests
	}

	return metrics, nil
}

func collectMetricsForPR(client *github.Client, owner, repo string, pr github.PullRequest) (PRMetrics, error) {
	metrics := PRMetrics{
		Repository: fmt.Sprintf("%s/%s", owner, repo),

		PRNumber:  pr.Number,
		Author:    pr.User.Login,
		State:     pr.State,
		CreatedAt: pr.CreatedAt,
		IsMerged:  pr.MergedAt != nil,
	}

	if pr.ClosedAt != nil {
		metrics.ClosedAt = *pr.ClosedAt
	}
	if pr.MergedAt != nil {
		metrics.MergedAt = *pr.MergedAt
	}

	reviews, err := client.GetReviews(owner, repo, pr.Number)
	if err != nil {
		return metrics, fmt.Errorf("reviews: %v", err)
	}

	comments, err := client.GetComments(owner, repo, pr.Number)
	if err != nil {
		return metrics, fmt.Errorf("comments: %v", err)
	}

	processReviews(&metrics, reviews, pr.User.Login)

	calculateLifetime(&metrics, pr)

	metrics.CommentsCount = len(comments)

	return metrics, nil
}

func processReviews(metrics *PRMetrics, reviews []github.Review, author string) {
	if len(reviews) > 0 {
		firstReview := findFirstReview(reviews)
		if firstReview.SubmittedAt != nil {
			metrics.FirstReviewTime = *firstReview.SubmittedAt
			metrics.TimeToFirstReview = firstReview.SubmittedAt.Sub(metrics.CreatedAt)

			reviewerSet := make(map[string]bool)
			for _, review := range reviews {
				if review.User.Login != author && review.SubmittedAt != nil {
					reviewerSet[review.User.Login] = true
				}
			}
			for reviewer := range reviewerSet {
				metrics.Reviewers = append(metrics.Reviewers, reviewer)
			}
		}
	}
}

func calculateLifetime(metrics *PRMetrics, pr github.PullRequest) {
	if metrics.IsMerged && pr.MergedAt != nil {
		metrics.TotalLifetime = pr.MergedAt.Sub(pr.CreatedAt)
	} else if pr.ClosedAt != nil {
		metrics.TotalLifetime = pr.ClosedAt.Sub(pr.CreatedAt)
	} else {
		metrics.TotalLifetime = time.Since(pr.CreatedAt)
	}
}

func findFirstReview(reviews []github.Review) github.Review {
	var first github.Review
	firstFound := false

	for _, review := range reviews {
		if review.SubmittedAt != nil {
			if !firstFound || review.SubmittedAt.Before(*first.SubmittedAt) {
				first = review
				firstFound = true
			}
		}
	}
	return first
}

func ComparativeAnalysis(results map[string]RepositoryResult) ComparativeAnalyser {
	comparative := ComparativeAnalyser{
		RepositoryResults: results,
		Summary: SummaryStats{
			TotalRepositories: len(results),
		},
	}

	// Calculation of total statistics
	totalPRs := 0
	totalMerged := 0
	var mergeRates []float64
	var performances []RepoPerformance

	for repoKey, result := range results {
		totalPRs += result.Analysis.TotalPRs
		totalMerged += result.Analysis.MergedPRs
		mergeRates = append(mergeRates, result.Analysis.MergeRate)

		performances = append(performances, RepoPerformance{
			Repository: repoKey,
			MergeRate:  result.Analysis.MergeRate,
			AvgTime:    result.Analysis.AverageLifetime,
		})
	}

	// Sorting by efficiency
	sort.Slice(performances, func(i, j int) bool {
		return performances[i].MergeRate > performances[j].MergeRate
	})

	comparative.Summary.TotalPRs = totalPRs
	comparative.Summary.TotalMergedPRs = totalMerged
	if len(results) > 0 {
		comparative.Summary.AvgMergeRate = float64(totalMerged) / float64(totalPRs) * 100
	}

	// Top 3 best and worst
	if len(performances) > 3 {
		comparative.Summary.BestPerforming = performances[:3]
		comparative.Summary.WorstPerforming = performances[len(performances)-3:]
	} else {
		comparative.Summary.BestPerforming = performances
	}

	return comparative
}
