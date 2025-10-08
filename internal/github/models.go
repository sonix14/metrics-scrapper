package github

import "time"

type PullRequest struct {
	ID        int        `json:"id"`
	Number    int        `json:"number"`
	State     string     `json:"state"`
	Title     string     `json:"title"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	ClosedAt  *time.Time `json:"closed_at"`
	MergedAt  *time.Time `json:"merged_at"`
	User      User       `json:"user"`
	URL       string     `json:"url"`
	HTMLURL   string     `json:"html_url"`
}

type User struct {
	ID    int    `json:"id"`
	Login string `json:"login"`
}

type Review struct {
	ID          int        `json:"id"`
	User        User       `json:"user"`
	State       string     `json:"state"`
	SubmittedAt *time.Time `json:"submitted_at"`
}

type IssueComment struct {
	ID        int       `json:"id"`
	User      User      `json:"user"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type RateLimitInfo struct {
	Remaining string
	Reset     string
	Limit     string
}
