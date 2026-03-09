package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"text/tabwriter"
	"time"
)

type searchResponse struct {
	Items []repo `json:"items"`
}

type repo struct {
	FullName        string    `json:"full_name"`
	Description     string    `json:"description"`
	StargazersCount int       `json:"stargazers_count"`
	HTMLURL         string    `json:"html_url"`
	Language        string    `json:"language"`
	CreatedAt       time.Time `json:"created_at"`
}

type scoredRepo struct {
	repo
	AgeDays     float64 `json:"age_days"`
	StarsPerDay float64 `json:"stars_per_day"`
	HotScore    float64 `json:"hot_score"`
	Category    string  `json:"category"`
}

func main() {
	var queries multiFlag
	var limit int
	var jsonOut bool
	var showThemes bool
	flag.Var(&queries, "q", "GitHub search query (repeatable)")
	flag.IntVar(&limit, "n", 15, "Top results to print")
	flag.BoolVar(&jsonOut, "json", false, "Print JSON output")
	flag.BoolVar(&showThemes, "themes", false, "Print theme distribution summary")
	flag.Parse()

	if len(queries) == 0 {
		queries = []string{
			"topic:cli created:>2026-01-01 stars:>40",
			"topic:tui created:>2026-01-01 stars:>20",
			"(agent OR mcp) created:>2026-01-01 stars:>80",
			"(developer tools) created:>2026-01-01 stars:>50",
		}
	}

	client := &http.Client{Timeout: 20 * time.Second}
	all, err := fetchAndMerge(client, queries)
	if err != nil {
		log.Fatalf("fetch trends: %v", err)
	}
	if len(all) == 0 {
		log.Fatal("no repositories returned")
	}

	scored := score(all)
	if limit > len(scored) {
		limit = len(scored)
	}
	scored = scored[:limit]

	if jsonOut {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(scored); err != nil {
			log.Fatal(err)
		}
		return
	}

	printTable(scored)
	if showThemes {
		fmt.Println()
		printThemeSummary(scored)
	}
}

type multiFlag []string

func (m *multiFlag) String() string { return strings.Join(*m, ",") }
func (m *multiFlag) Set(v string) error {
	*m = append(*m, v)
	return nil
}

func fetchAndMerge(client *http.Client, queries []string) ([]repo, error) {
	byName := make(map[string]repo)
	for _, q := range queries {
		repos, err := search(client, q)
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
	out := make([]repo, 0, len(byName))
	for _, r := range byName {
		out = append(out, r)
	}
	return out, nil
}

func search(client *http.Client, q string) ([]repo, error) {
	u := "https://api.github.com/search/repositories?sort=stars&order=desc&per_page=30&q=" + url.QueryEscape(q)
	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	if token := strings.TrimSpace(os.Getenv("GITHUB_TOKEN")); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return nil, fmt.Errorf("github api status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var sr searchResponse
	if err := json.NewDecoder(resp.Body).Decode(&sr); err != nil {
		return nil, err
	}
	return sr.Items, nil
}

func score(in []repo) []scoredRepo {
	now := time.Now().UTC()
	out := make([]scoredRepo, 0, len(in))
	for _, r := range in {
		age := now.Sub(r.CreatedAt).Hours() / 24
		if age < 1 {
			age = 1
		}
		spd := float64(r.StargazersCount) / age
		hot := spd * math.Log10(float64(r.StargazersCount)+1)
		out = append(out, scoredRepo{
			repo:        r,
			AgeDays:     age,
			StarsPerDay: spd,
			HotScore:    hot,
			Category:    categorize(r),
		})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].HotScore == out[j].HotScore {
			return out[i].StargazersCount > out[j].StargazersCount
		}
		return out[i].HotScore > out[j].HotScore
	})
	return out
}

func categorize(r repo) string {
	text := strings.ToLower(r.FullName + " " + r.Description)
	switch {
	case strings.Contains(text, "agent") || strings.Contains(text, "mcp"):
		return "agent"
	case strings.Contains(text, "cli") || strings.Contains(text, "terminal"):
		return "cli"
	case strings.Contains(text, "tui"):
		return "tui"
	case strings.Contains(text, "github") || strings.Contains(text, "dev"):
		return "devtool"
	default:
		return "general"
	}
}

func printTable(in []scoredRepo) {
	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "RANK\tREPO\tSTARS\tAGE(d)\tSTARS/DAY\tSCORE\tCATEGORY\tLANG")
	for i, r := range in {
		fmt.Fprintf(tw, "%d\t%s\t%d\t%.1f\t%.1f\t%.1f\t%s\t%s\n",
			i+1, r.FullName, r.StargazersCount, r.AgeDays, r.StarsPerDay, r.HotScore, r.Category, r.Language)
	}
	tw.Flush()
}

func printThemeSummary(in []scoredRepo) {
	type themeStats struct {
		Count    int
		AvgScore float64
	}
	m := map[string]themeStats{}
	for _, r := range in {
		s := m[r.Category]
		s.Count++
		s.AvgScore += r.HotScore
		m[r.Category] = s
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "THEME\tCOUNT\tAVG_SCORE")
	for _, k := range keys {
		s := m[k]
		avg := s.AvgScore / float64(s.Count)
		fmt.Fprintf(tw, "%s\t%d\t%.1f\n", k, s.Count, avg)
	}
	tw.Flush()
}
