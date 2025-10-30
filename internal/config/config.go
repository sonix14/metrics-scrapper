package config

import (
	"fmt"
	"os"
	"strconv"
)

type RepoConfig struct {
	Owner string
	Repo  string
}

type Config struct {
	GitHubToken  string
	Repositories []RepoConfig
	MaxPages     int
	DelayMS      int
	PerPage      int
}

func LoadConfig() *Config {
	repos := []RepoConfig{
		{Owner: "stmcginnis", Repo: "gofish"},
		{Owner: "golang", Repo: "go"},
		{Owner: "ipmitool", Repo: "ipmitool"},
		{Owner: "docker", Repo: "compose"},
		{Owner: "VictoriaMetrics", Repo: "VictoriaMetrics"},
		{Owner: "prometheus", Repo: "prometheus"},
	}

	cfg := &Config{
		GitHubToken:  getGitHubToken(),
		Repositories: repos,
		MaxPages:     getEnvAsInt("MAX_PAGES", 3),
		DelayMS:      getEnvAsInt("DELAY_MS", 1000),
		PerPage:      getEnvAsInt("PER_PAGE", 5),
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
