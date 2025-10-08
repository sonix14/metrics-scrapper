package analyzer

import (
	"time"
)

type PRMetrics struct {
	PRNumber          int
	Author            string
	State             string
	CreatedAt         time.Time
	ClosedAt          time.Time
	MergedAt          time.Time
	FirstReviewTime   time.Time
	TotalLifetime     time.Duration
	TimeToFirstReview time.Duration
	Reviewers         []string
	CommentsCount     int
	IsMerged          bool
}

type AnalysisResult struct {
	TotalPRs                 int
	MergedPRs                int
	ClosedPRs                int
	MergeRate                float64
	AverageLifetime          time.Duration
	AverageTimeToFirstReview time.Duration
	MedianLifetime           time.Duration
	MedianTimeToFirstReview  time.Duration
	PRMetrics                []PRMetrics
}

type AuthorStats struct {
	Author        string
	PRCount       int
	MergedCount   int
	AvgLifetime   time.Duration
	AvgReviewTime time.Duration
}

type ReviewerStats struct {
	Reviewer      string
	ReviewCount   int
	AvgReviewTime time.Duration
}

type AuthorReviewerPair struct {
	Author       string
	Reviewer     string
	Interactions int
	AvgLifetime  time.Duration
}
