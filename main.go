package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

// Конфигурация
type Config struct {
	GitHubToken string
	Owner       string
	Repo        string
	MaxPages    int
	DelayMS     int
}

// Структуры для парсинга JSON ответов от GitHub API
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

func main() {
	fmt.Println("=== Анализ PR проекта gofish ===\n")

	// Конфигурация
	config := Config{
		GitHubToken: getGitHubToken(),
		Owner:       "stmcginnis",
		Repo:        "gofish",
		MaxPages:    3,   // Ограничение для демо
		DelayMS:     500, // Задержка между запросами
	}

	if config.GitHubToken == "" {
		fmt.Println("⚠️  Работаем без токена (ограниченное кол-во запросов)")
		fmt.Println("   Для увеличения лимитов создайте GITHUB_TOKEN\n")
	}

	// Получаем все PR
	prs, err := getAllPullRequests(config)
	if err != nil {
		log.Fatalf("Ошибка при получении PR: %v", err)
	}

	fmt.Printf("Найдено %d pull request'ов\n", len(prs))

	if len(prs) == 0 {
		fmt.Println("Не найдено PR для анализа")
		return
	}

	// Собираем метрики для каждого PR
	var prMetrics []PRMetrics
	for i, pr := range prs {
		fmt.Printf("Обработка PR #%d (%d/%d)\n", pr.Number, i+1, len(prs))

		metrics, err := getPRMetrics(config, pr)
		if err != nil {
			log.Printf("Ошибка при получении метрик для PR #%d: %v", pr.Number, err)
			continue
		}
		prMetrics = append(prMetrics, metrics)

		// Задержка чтобы не превысить лимиты GitHub API
		time.Sleep(time.Duration(config.DelayMS) * time.Millisecond)
	}

	// Сохраняем сырые данные в JSON
	saveRawData(prMetrics)
}

func getGitHubToken() string {
	// 1. Проверяем переменную окружения
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		return token
	}

	// 2. Можно добавить чтение из файла или ввод пользователя
	fmt.Println("Подсказка: установите переменную окружения GITHUB_TOKEN для увеличения лимитов")
	fmt.Println("export GITHUB_TOKEN=your_token_here")

	return ""
}

func getAllPullRequests(config Config) ([]PullRequest, error) {
	var allPRs []PullRequest
	page := 1
	perPage := 5 // Уменьшаем для анонимных запросов

	for {
		url := fmt.Sprintf("https://api.github.com/repos/%s/%s/pulls?state=all&page=%d&per_page=%d&sort=created&direction=desc",
			config.Owner, config.Repo, page, perPage)

		fmt.Printf("Запрос страницы %d...\n", page)

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, fmt.Errorf("ошибка создания запроса: %v", err)
		}

		// Устанавливаем заголовки
		req.Header.Set("User-Agent", "GoFish-Analyzer-1.0")
		req.Header.Set("Accept", "application/vnd.github.v3+json")
		if config.GitHubToken != "" {
			req.Header.Set("Authorization", "token "+config.GitHubToken)
		}

		client := &http.Client{Timeout: 30 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("ошибка выполнения запроса: %v", err)
		}
		defer resp.Body.Close()

		// Проверяем лимиты API
		checkRateLimit(resp)

		if resp.StatusCode == 403 {
			return nil, fmt.Errorf("доступ запрещен. Проверьте токен или подождите сброса лимитов")
		}

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("HTTP ошибка %d: %s", resp.StatusCode, resp.Status)
		}

		var prs []PullRequest
		if err := json.NewDecoder(resp.Body).Decode(&prs); err != nil {
			return nil, fmt.Errorf("ошибка парсинга JSON: %v", err)
		}

		if len(prs) == 0 {
			break // Больше нет данных
		}

		allPRs = append(allPRs, prs...)
		fmt.Printf("Получено %d PR со страницы %d\n", len(prs), page)

		page++

		// Ограничим количество страниц для демонстрации
		if page > config.MaxPages {
			fmt.Printf("Достигнут лимит страниц (%d)\n", config.MaxPages)
			break
		}

		time.Sleep(time.Duration(config.DelayMS) * time.Millisecond)
	}

	return allPRs, nil
}

func checkRateLimit(resp *http.Response) {
	remaining := resp.Header.Get("X-RateLimit-Remaining")
	reset := resp.Header.Get("X-RateLimit-Reset")

	if remaining != "" {
		fmt.Printf("Лимит API: осталось %s запросов\n", remaining)

		if remaining == "0" {
			fmt.Printf("Лимит исчерпан, сброс в: %s\n", reset)
		}
	}
}

func getPRMetrics(config Config, pr PullRequest) (PRMetrics, error) {
	metrics := PRMetrics{
		PRNumber:  pr.Number,
		Author:    pr.User.Login,
		State:     pr.State,
		CreatedAt: pr.CreatedAt,
		IsMerged:  pr.MergedAt != nil,
	}

	// Устанавливаем время закрытия/мерджа
	if pr.ClosedAt != nil {
		metrics.ClosedAt = *pr.ClosedAt
	}
	if pr.MergedAt != nil {
		metrics.MergedAt = *pr.MergedAt
	}

	// Получаем reviews
	reviews, err := getReviews(config, pr.Number)
	if err != nil {
		fmt.Printf("  Предупреждение: не удалось получить reviews для PR #%d: %v\n", pr.Number, err)
	}

	// Получаем комментарии
	comments, err := getComments(config, pr.Number)
	if err != nil {
		fmt.Printf("  Предупреждение: не удалось получить комментарии для PR #%d: %v\n", pr.Number, err)
	}

	// Находим время первого ревью
	if len(reviews) > 0 {
		firstReview := findFirstReview(reviews)
		if firstReview.SubmittedAt != nil {
			metrics.FirstReviewTime = *firstReview.SubmittedAt
			metrics.TimeToFirstReview = firstReview.SubmittedAt.Sub(pr.CreatedAt)

			// Собираем уникальных ревьюверов
			reviewerSet := make(map[string]bool)
			for _, review := range reviews {
				if review.User.Login != pr.User.Login && review.SubmittedAt != nil {
					reviewerSet[review.User.Login] = true
				}
			}
			for reviewer := range reviewerSet {
				metrics.Reviewers = append(metrics.Reviewers, reviewer)
			}
		}
	}

	// Считаем общее время жизни PR
	if metrics.IsMerged && pr.MergedAt != nil {
		metrics.TotalLifetime = pr.MergedAt.Sub(pr.CreatedAt)
	} else if pr.ClosedAt != nil {
		metrics.TotalLifetime = pr.ClosedAt.Sub(pr.CreatedAt)
	} else {
		metrics.TotalLifetime = time.Since(pr.CreatedAt)
	}

	metrics.CommentsCount = len(comments)

	return metrics, nil
}

func getReviews(config Config, prNumber int) ([]Review, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/pulls/%d/reviews",
		config.Owner, config.Repo, prNumber)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "GoFish-Analyzer-1.0")
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	if config.GitHubToken != "" {
		req.Header.Set("Authorization", "token "+config.GitHubToken)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP ошибка при получении reviews: %s", resp.Status)
	}

	var reviews []Review
	if err := json.NewDecoder(resp.Body).Decode(&reviews); err != nil {
		return nil, err
	}

	return reviews, nil
}

func getComments(config Config, prNumber int) ([]IssueComment, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/issues/%d/comments",
		config.Owner, config.Repo, prNumber)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "GoFish-Analyzer-1.0")
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	if config.GitHubToken != "" {
		req.Header.Set("Authorization", "token "+config.GitHubToken)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP ошибка при получении комментариев: %s", resp.Status)
	}

	var comments []IssueComment
	if err := json.NewDecoder(resp.Body).Decode(&comments); err != nil {
		return nil, err
	}

	return comments, nil
}

func findFirstReview(reviews []Review) Review {
	var first Review
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

func calculateMedianDuration(durations []time.Duration) time.Duration {
	if len(durations) == 0 {
		return 0
	}

	// Создаем копию для сортировки
	sorted := make([]time.Duration, len(durations))
	copy(sorted, durations)

	// Сортировка
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

func saveRawData(metrics []PRMetrics) {
	file, err := os.Create("gofish_pr_data.json")
	if err != nil {
		log.Printf("Ошибка при создании файла данных: %v", err)
		return
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(metrics); err != nil {
		log.Printf("Ошибка при сохранении данных: %v", err)
		return
	}

	fmt.Printf("\nСырые данные сохранены в файл: gofish_pr_data.json\n")
}
