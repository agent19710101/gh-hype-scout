package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"gopkg.in/yaml.v3"
)

const defaultDescWidth = 56

type searchResponse struct {
	Items []repo `json:"items"`
}

type repo struct {
	FullName        string    `json:"full_name" yaml:"full_name"`
	Description     string    `json:"description" yaml:"description"`
	StargazersCount int       `json:"stargazers_count" yaml:"stargazers_count"`
	HTMLURL         string    `json:"html_url" yaml:"html_url"`
	Language        string    `json:"language" yaml:"language"`
	CreatedAt       time.Time `json:"created_at" yaml:"created_at"`
}

type scoredRepo struct {
	repo
	AgeDays     float64 `json:"age_days"`
	StarsPerDay float64 `json:"stars_per_day"`
	HotScore    float64 `json:"hot_score"`
	Category    string  `json:"category"`
}

type appConfig struct {
	Queries     []string `yaml:"queries"`
	Limit       int      `yaml:"limit"`
	JSON        bool     `yaml:"json"`
	Themes      bool     `yaml:"themes"`
	MinStars    int      `yaml:"min_stars"`
	SinceDays   int      `yaml:"since_days"`
	MinAgeDays  int      `yaml:"min_age_days"`
	MaxAgeDays  int      `yaml:"max_age_days"`
	Sort        string   `yaml:"sort"`
	ScorePreset string   `yaml:"score_preset"`
	DescWidth   int      `yaml:"desc_width"`
}

type scoreConfig struct {
	Sort        string
	ScorePreset string
}

func main() {
	var queries multiFlag
	var limit int
	var jsonOut bool
	var showThemes bool
	var minStars int
	var sinceDays int
	var minAgeDays int
	var maxAgeDays int
	var sortBy string
	var scorePreset string
	var configPath string
	var descWidth int

	flag.Var(&queries, "q", "GitHub search query (repeatable)")
	flag.IntVar(&limit, "n", 15, "Top results to print")
	flag.BoolVar(&jsonOut, "json", false, "Print JSON output")
	flag.BoolVar(&showThemes, "themes", false, "Print theme distribution summary")
	flag.IntVar(&minStars, "min-stars", 0, "Hide repos with stars below this threshold")
	flag.IntVar(&sinceDays, "since-days", 60, "Default query window in days (only used without -q)")
	flag.IntVar(&minAgeDays, "min-age-days", 0, "Hide repos younger than this age in days")
	flag.IntVar(&maxAgeDays, "max-age-days", 0, "Hide repos older than this age in days")
	flag.StringVar(&sortBy, "sort", "hot", "Sort results by: hot, stars-day, stars, age")
	flag.StringVar(&scorePreset, "score-preset", "hot", "Score preset for -sort hot: hot, fresh")
	flag.StringVar(&configPath, "config", defaultConfigPath(), "Config file path")
	flag.IntVar(&descWidth, "desc-width", defaultDescWidth, "Description column max width for table output")
	flag.Parse()

	setFlags := map[string]bool{}
	flag.Visit(func(f *flag.Flag) {
		setFlags[f.Name] = true
	})

	cfg, err := loadConfig(configPath)
	if err != nil {
		log.Fatal(err)
	}
	applyConfigIfUnset(cfg, setFlags, &queries, &limit, &jsonOut, &showThemes, &minStars, &sinceDays, &minAgeDays, &maxAgeDays, &sortBy, &scorePreset, &descWidth)

	if err := validateAgeFlags(minAgeDays, maxAgeDays); err != nil {
		log.Fatal(err)
	}
	scfg := scoreConfig{Sort: sortBy, ScorePreset: scorePreset}
	if err := validateScoreConfig(scfg); err != nil {
		log.Fatal(err)
	}
	if descWidth < 8 {
		log.Fatal("-desc-width must be >= 8")
	}

	if len(queries) == 0 {
		queries = defaultQueries(sinceDays, time.Now().UTC())
	}

	client := &http.Client{Timeout: 20 * time.Second}
	all, err := fetchAndMerge(client, queries)
	if err != nil {
		log.Fatalf("fetch trends: %v", err)
	}
	if len(all) == 0 {
		log.Fatal("no repositories returned")
	}

	scored := score(all, scfg)
	if minStars > 0 {
		filtered := scored[:0]
		for _, r := range scored {
			if r.StargazersCount >= minStars {
				filtered = append(filtered, r)
			}
		}
		scored = filtered
	}
	scored = filterByAge(scored, minAgeDays, maxAgeDays)
	if len(scored) == 0 {
		log.Fatal("no repositories matched filters")
	}
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

	printTable(os.Stdout, scored, descWidth)
	if showThemes {
		fmt.Println()
		printThemeSummary(scored)
	}
}

func defaultConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil || strings.TrimSpace(home) == "" {
		return ""
	}
	return filepath.Join(home, ".config", "gh-hype-scout", "config.yaml")
}

func loadConfig(path string) (appConfig, error) {
	if strings.TrimSpace(path) == "" {
		return appConfig{}, nil
	}
	b, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return appConfig{}, nil
		}
		return appConfig{}, fmt.Errorf("read config %q: %w", path, err)
	}
	var cfg appConfig
	if err := yaml.Unmarshal(b, &cfg); err != nil {
		return appConfig{}, fmt.Errorf("parse config %q: %w", path, err)
	}
	return cfg, nil
}

func applyConfigIfUnset(cfg appConfig, setFlags map[string]bool, queries *multiFlag, limit *int, jsonOut *bool, showThemes *bool, minStars *int, sinceDays *int, minAgeDays *int, maxAgeDays *int, sortBy *string, scorePreset *string, descWidth *int) {
	if !setFlags["q"] && len(cfg.Queries) > 0 {
		*queries = append((*queries)[:0], cfg.Queries...)
	}
	if !setFlags["n"] && cfg.Limit > 0 {
		*limit = cfg.Limit
	}
	if !setFlags["json"] && cfg.JSON {
		*jsonOut = cfg.JSON
	}
	if !setFlags["themes"] && cfg.Themes {
		*showThemes = cfg.Themes
	}
	if !setFlags["min-stars"] && cfg.MinStars > 0 {
		*minStars = cfg.MinStars
	}
	if !setFlags["since-days"] && cfg.SinceDays > 0 {
		*sinceDays = cfg.SinceDays
	}
	if !setFlags["min-age-days"] && cfg.MinAgeDays > 0 {
		*minAgeDays = cfg.MinAgeDays
	}
	if !setFlags["max-age-days"] && cfg.MaxAgeDays > 0 {
		*maxAgeDays = cfg.MaxAgeDays
	}
	if !setFlags["sort"] && strings.TrimSpace(cfg.Sort) != "" {
		*sortBy = cfg.Sort
	}
	if !setFlags["score-preset"] && strings.TrimSpace(cfg.ScorePreset) != "" {
		*scorePreset = cfg.ScorePreset
	}
	if !setFlags["desc-width"] && cfg.DescWidth > 0 {
		*descWidth = cfg.DescWidth
	}
}

func defaultQueries(sinceDays int, now time.Time) []string {
	if sinceDays < 1 {
		sinceDays = 1
	}
	since := now.AddDate(0, 0, -sinceDays).Format("2006-01-02")
	return []string{
		"topic:cli created:>" + since + " stars:>40",
		"topic:tui created:>" + since + " stars:>20",
		"(agent OR mcp) created:>" + since + " stars:>80",
		"(developer tools) created:>" + since + " stars:>50",
	}
}

type multiFlag []string

func (m *multiFlag) String() string { return strings.Join(*m, ",") }
func (m *multiFlag) Set(v string) error {
	*m = append(*m, v)
	return nil
}

type searchFunc func(client *http.Client, q string) ([]repo, error)

func fetchAndMerge(client *http.Client, queries []string) ([]repo, error) {
	return fetchAndMergeWithSearcher(client, queries, search)
}

func fetchAndMergeWithSearcher(client *http.Client, queries []string, searcher searchFunc) ([]repo, error) {
	byName := make(map[string]repo)
	for _, q := range queries {
		repos, err := searcher(client, q)
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
		rateHint := githubRateLimitHint(resp)
		return nil, fmt.Errorf("github api status %d: %s%s", resp.StatusCode, strings.TrimSpace(string(body)), rateHint)
	}

	var sr searchResponse
	if err := json.NewDecoder(resp.Body).Decode(&sr); err != nil {
		return nil, err
	}
	return sr.Items, nil
}

func githubRateLimitHint(resp *http.Response) string {
	if resp == nil {
		return ""
	}
	if resp.StatusCode != http.StatusForbidden && resp.StatusCode != http.StatusTooManyRequests {
		return ""
	}

	if v := strings.TrimSpace(resp.Header.Get("Retry-After")); v != "" {
		if seconds, err := strconv.Atoi(v); err == nil && seconds > 0 {
			return fmt.Sprintf(" (rate limit hit; retry after %ds)", seconds)
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

func score(in []repo, cfg scoreConfig) []scoredRepo {
	now := time.Now().UTC()
	out := make([]scoredRepo, 0, len(in))
	for _, r := range in {
		age := now.Sub(r.CreatedAt).Hours() / 24
		if age < 1 {
			age = 1
		}
		spd := float64(r.StargazersCount) / age
		hot := spd * math.Log10(float64(r.StargazersCount)+1)
		if cfg.ScorePreset == "fresh" {
			hot = hot * (1 + (1 / math.Sqrt(age)))
		}
		out = append(out, scoredRepo{
			repo:        r,
			AgeDays:     age,
			StarsPerDay: spd,
			HotScore:    hot,
			Category:    categorize(r),
		})
	}
	sortScored(out, cfg.Sort)
	return out
}

func validateScoreConfig(cfg scoreConfig) error {
	switch cfg.Sort {
	case "hot", "stars-day", "stars", "age":
	default:
		return fmt.Errorf("invalid -sort %q (expected: hot, stars-day, stars, age)", cfg.Sort)
	}
	switch cfg.ScorePreset {
	case "hot", "fresh":
	default:
		return fmt.Errorf("invalid -score-preset %q (expected: hot, fresh)", cfg.ScorePreset)
	}
	if cfg.Sort != "hot" && cfg.ScorePreset != "hot" {
		return fmt.Errorf("-score-preset=%q only applies with -sort hot", cfg.ScorePreset)
	}
	return nil
}

func sortScored(in []scoredRepo, sortBy string) {
	sort.Slice(in, func(i, j int) bool {
		switch sortBy {
		case "stars-day":
			if in[i].StarsPerDay == in[j].StarsPerDay {
				return in[i].StargazersCount > in[j].StargazersCount
			}
			return in[i].StarsPerDay > in[j].StarsPerDay
		case "stars":
			if in[i].StargazersCount == in[j].StargazersCount {
				return in[i].HotScore > in[j].HotScore
			}
			return in[i].StargazersCount > in[j].StargazersCount
		case "age":
			if in[i].AgeDays == in[j].AgeDays {
				return in[i].StarsPerDay > in[j].StarsPerDay
			}
			return in[i].AgeDays < in[j].AgeDays
		default:
			if in[i].HotScore == in[j].HotScore {
				return in[i].StargazersCount > in[j].StargazersCount
			}
			return in[i].HotScore > in[j].HotScore
		}
	})
}

func validateAgeFlags(minAgeDays, maxAgeDays int) error {
	if minAgeDays < 0 || maxAgeDays < 0 {
		return fmt.Errorf("age filters must be >= 0")
	}
	if maxAgeDays > 0 && minAgeDays > maxAgeDays {
		return fmt.Errorf("min-age-days (%d) cannot be greater than max-age-days (%d)", minAgeDays, maxAgeDays)
	}
	return nil
}

func filterByAge(in []scoredRepo, minAgeDays, maxAgeDays int) []scoredRepo {
	if minAgeDays == 0 && maxAgeDays == 0 {
		return in
	}
	out := in[:0]
	for _, r := range in {
		if minAgeDays > 0 && r.AgeDays < float64(minAgeDays) {
			continue
		}
		if maxAgeDays > 0 && r.AgeDays > float64(maxAgeDays) {
			continue
		}
		out = append(out, r)
	}
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

func printTable(w io.Writer, in []scoredRepo, descWidth int) {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "RANK\tREPO\tSTARS\tAGE(d)\tSTARS/DAY\tSCORE\tCATEGORY\tLANG\tDESC")
	for i, r := range in {
		desc := truncateDescription(r.Description, descWidth)
		fmt.Fprintf(tw, "%d\t%s\t%d\t%.1f\t%.1f\t%.1f\t%s\t%s\t%s\n",
			i+1, r.FullName, r.StargazersCount, r.AgeDays, r.StarsPerDay, r.HotScore, r.Category, r.Language, desc)
	}
	tw.Flush()
}

func truncateDescription(s string, max int) string {
	s = strings.Join(strings.Fields(strings.TrimSpace(s)), " ")
	if s == "" {
		return "-"
	}
	if max < 2 {
		return s
	}
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	return string(r[:max-1]) + "…"
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
