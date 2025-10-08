package github

import (
	"encoding/json"
	"fmt"
	"time"
)

type GitHubService interface {
	GetAllPullRequests() ([]PullRequest, error)
	GetReviews(prNumber int) ([]Review, error)
	GetComments(prNumber int) ([]IssueComment, error)
}

func (c *Client) GetAllPullRequests() ([]PullRequest, error) {
	var allPRs []PullRequest
	page := 1

	for {
		url := fmt.Sprintf("https://api.github.com/repos/%s/%s/pulls?state=all&page=%d&per_page=%d&sort=created&direction=desc",
			c.config.Owner, c.config.Repo, page, c.config.PerPage)

		fmt.Printf("Page Request %d...\n", page)

		req, err := c.createRequest(url)
		if err != nil {
			return nil, fmt.Errorf("request creation error: %v", err)
		}

		resp, err := c.doRequest(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		var prs []PullRequest
		if err := json.NewDecoder(resp.Body).Decode(&prs); err != nil {
			return nil, fmt.Errorf("parsing JSON error: %v", err)
		}

		if len(prs) == 0 {
			break
		}

		allPRs = append(allPRs, prs...)
		fmt.Printf("Received %d PR from the page %d\n", len(prs), page)

		page++
		if page > c.config.MaxPages {
			fmt.Printf("The page limit has been reached (%d)\n", c.config.MaxPages)
			break
		}

		time.Sleep(time.Duration(c.config.DelayMS) * time.Millisecond)
	}

	return allPRs, nil
}

func (c *Client) GetReviews(prNumber int) ([]Review, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/pulls/%d/reviews",
		c.config.Owner, c.config.Repo, prNumber)

	req, err := c.createRequest(url)
	if err != nil {
		return nil, err
	}

	resp, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var reviews []Review
	if err := json.NewDecoder(resp.Body).Decode(&reviews); err != nil {
		return nil, err
	}

	return reviews, nil
}

func (c *Client) GetComments(prNumber int) ([]IssueComment, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/issues/%d/comments",
		c.config.Owner, c.config.Repo, prNumber)

	req, err := c.createRequest(url)
	if err != nil {
		return nil, err
	}

	resp, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var comments []IssueComment
	if err := json.NewDecoder(resp.Body).Decode(&comments); err != nil {
		return nil, err
	}

	return comments, nil
}
