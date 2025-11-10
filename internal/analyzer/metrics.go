package analyzer

import (
	"time"
)

type PRMetrics struct {
	Repository        string
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

type RepositoryResult struct {
	Owner    string
	Repo     string
	PRCount  int
	Metrics  []PRMetrics
	Analysis AnalysisResult
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
	PredictedTimeToMerge     time.Duration
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

type ComparativeAnalyser struct {
	RepositoryResults map[string]RepositoryResult // key: "owner/repo"
	Summary           SummaryStats
}

type SummaryStats struct {
	TotalRepositories int
	TotalPRs          int
	TotalMergedPRs    int
	AvgMergeRate      float64
	BestPerforming    []RepoPerformance
	WorstPerforming   []RepoPerformance
}

type RepoPerformance struct {
	Repository string
	MergeRate  float64
	AvgTime    time.Duration
}
