package github

import (
	"fmt"
	"metrics-scrapper/internal/config"
	"net/http"
	"time"
)

type Client struct {
	config     *config.Config
	httpClient *http.Client
}

func NewClient(cfg *config.Config) *Client {
	return &Client{
		config:     cfg,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *Client) createRequest(url string) (*http.Request, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "GoFish-Analyzer-1.0")
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	if c.config.GitHubToken != "" {
		req.Header.Set("Authorization", "token "+c.config.GitHubToken)
	}

	return req, nil
}

func (c *Client) doRequest(req *http.Request) (*http.Response, error) {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	c.checkRateLimit(resp)

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("HTTP ошибка %d: %s", resp.StatusCode, resp.Status)
	}

	return resp, nil
}

func (c *Client) checkRateLimit(resp *http.Response) {
	remaining := resp.Header.Get("X-RateLimit-Remaining")
	reset := resp.Header.Get("X-RateLimit-Reset")
	limit := resp.Header.Get("X-RateLimit-Limit")

	if remaining != "" {
		fmt.Printf("Limit API: %s/%s requests, reset: %s\n", remaining, limit, reset)
	}
}
