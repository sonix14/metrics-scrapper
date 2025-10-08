package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	GitHubToken string
	Owner       string
	Repo        string
	MaxPages    int
	DelayMS     int
	PerPage     int
}

func LoadConfig() *Config {
	cfg := &Config{
		GitHubToken: getGitHubToken(),
		Owner:       getEnv("GITHUB_OWNER", "stmcginnis"),
		Repo:        getEnv("GITHUB_REPO", "gofish"),
		MaxPages:    getEnvAsInt("MAX_PAGES", 3),
		DelayMS:     getEnvAsInt("DELAY_MS", 500),
		PerPage:     getEnvAsInt("PER_PAGE", 5),
	}

	if cfg.GitHubToken == "" {
		fmt.Println("⚠️  Working without a token (limited number of requests)")
		fmt.Println("   To increase the limits, create a GITHUB_TOKEN")
	}

	return cfg
}

func getGitHubToken() string {
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		return token
	}
	return ""
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
