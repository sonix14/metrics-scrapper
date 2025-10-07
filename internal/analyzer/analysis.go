package analyzer

import (
	"time"
)

func AnalyzeData(metrics []PRMetrics) AnalysisResult {
	result := AnalysisResult{
		PRMetrics: metrics,
		TotalPRs:  len(metrics),
	}

	var totalLifetime time.Duration
	var totalTimeToFirstReview time.Duration
	var lifetimes []time.Duration
	var reviewTimes []time.Duration

	mergedCount := 0
	closedCount := 0

	for _, m := range metrics {
		if m.IsMerged {
			mergedCount++
		} else if m.State == "closed" {
			closedCount++
		}

		totalLifetime += m.TotalLifetime
		lifetimes = append(lifetimes, m.TotalLifetime)

		if m.TimeToFirstReview > 0 {
			totalTimeToFirstReview += m.TimeToFirstReview
			reviewTimes = append(reviewTimes, m.TimeToFirstReview)
		}
	}

	result.MergedPRs = mergedCount
	result.ClosedPRs = closedCount

	if result.TotalPRs > 0 {
		result.MergeRate = float64(result.MergedPRs) / float64(result.TotalPRs) * 100
		result.AverageLifetime = totalLifetime / time.Duration(result.TotalPRs)

		if len(reviewTimes) > 0 {
			result.AverageTimeToFirstReview = totalTimeToFirstReview / time.Duration(len(reviewTimes))
			result.MedianTimeToFirstReview = calculateMedianDuration(reviewTimes)
		}

		if len(lifetimes) > 0 {
			result.MedianLifetime = calculateMedianDuration(lifetimes)
		}
	}

	return result
}

func calculateMedianDuration(durations []time.Duration) time.Duration {
	if len(durations) == 0 {
		return 0
	}

	sorted := make([]time.Duration, len(durations))
	copy(sorted, durations)

	for i := 0; i < len(sorted)-1; i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[i] > sorted[j] {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	mid := len(sorted) / 2
	if len(sorted)%2 == 0 {
		return (sorted[mid-1] + sorted[mid]) / 2
	}
	return sorted[mid]
}
