package githubapi

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

type Repo struct {
	FullName        string    `json:"full_name"`
	Description     string    `json:"description"`
	StargazersCount int       `json:"stargazers_count"`
	HTMLURL         string    `json:"html_url"`
	Language        string    `json:"language"`
	CreatedAt       time.Time `json:"created_at"`
}

type searchResponse struct {
	Items []Repo `json:"items"`
}

type Client struct{ HTTP *http.Client }

func New() *Client { return &Client{HTTP: &http.Client{Timeout: 20 * time.Second}} }

func (c *Client) Search(q string) ([]Repo, error) {
	u := "https://api.github.com/search/repositories?sort=stars&order=desc&per_page=30&q=" + url.QueryEscape(q)
	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	if token := strings.TrimSpace(os.Getenv("GITHUB_TOKEN")); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return nil, fmt.Errorf("github api status %d: %s%s", resp.StatusCode, strings.TrimSpace(string(body)), rateHint(resp))
	}
	var sr searchResponse
	if err := json.NewDecoder(resp.Body).Decode(&sr); err != nil {
		return nil, err
	}
	return sr.Items, nil
}

func MergeSearch(queries []string, search func(string) ([]Repo, error)) ([]Repo, error) {
	byName := map[string]Repo{}
	for _, q := range queries {
		repos, err := search(q)
		if err != nil {
			return nil, fmt.Errorf("query %q: %w", q, err)
		}
		for _, r := range repos {
			if r.FullName == "" {
				continue
			}
			if old, ok := byName[r.FullName]; !ok || r.StargazersCount > old.StargazersCount {
				byName[r.FullName] = r
			}
		}
	}
	out := make([]Repo, 0, len(byName))
	for _, r := range byName {
		out = append(out, r)
	}
	return out, nil
}

func rateHint(resp *http.Response) string {
	if resp == nil || (resp.StatusCode != http.StatusForbidden && resp.StatusCode != http.StatusTooManyRequests) {
		return ""
	}
	if v := strings.TrimSpace(resp.Header.Get("Retry-After")); v != "" {
		if s, err := strconv.Atoi(v); err == nil && s > 0 {
			return fmt.Sprintf(" (rate limit hit; retry after %ds)", s)
		}
	}
	if v := strings.TrimSpace(resp.Header.Get("X-RateLimit-Reset")); v != "" {
		if ts, err := strconv.ParseInt(v, 10, 64); err == nil {
			wait := time.Until(time.Unix(ts, 0)).Round(time.Second)
			if wait > 0 {
				return fmt.Sprintf(" (rate limit hit; retry in %s)", wait)
			}
		}
	}
	return " (possible rate limit hit; provide GITHUB_TOKEN for higher limits)"
}
