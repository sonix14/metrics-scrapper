package analyzer

import (
	"fmt"
	"log"
	"metrics-scrapper/internal/github"
	"time"
)

func CollectPRMetrics(client *github.Client, prs []github.PullRequest) ([]PRMetrics, error) {
	var metrics []PRMetrics

	for i, pr := range prs {
		fmt.Printf("PR processing #%d (%d/%d)\n", pr.Number, i+1, len(prs))

		prMetrics, err := collectMetricsForPR(client, pr)
		if err != nil {
			log.Printf("Error when getting metrics for PR #%d: %v", pr.Number, err)
			continue
		}

		metrics = append(metrics, prMetrics)
		time.Sleep(500 * time.Millisecond) // delay between requests
	}

	return metrics, nil
}

func collectMetricsForPR(client *github.Client, pr github.PullRequest) (PRMetrics, error) {
	metrics := PRMetrics{
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

	reviews, err := client.GetReviews(pr.Number)
	if err != nil {
		return metrics, fmt.Errorf("reviews: %v", err)
	}

	comments, err := client.GetComments(pr.Number)
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
