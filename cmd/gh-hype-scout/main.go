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

const (
	defaultDescWidth   = 56
	defaultIntervalSec = 300
	maxSnapshotRuns    = 96
)

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
	Queries         []string `yaml:"queries"`
	Presets         []string `yaml:"presets"`
	Limit           int      `yaml:"limit"`
	JSON            bool     `yaml:"json"`
	Themes          bool     `yaml:"themes"`
	MinStars        int      `yaml:"min_stars"`
	SinceDays       int      `yaml:"since_days"`
	MinAgeDays      int      `yaml:"min_age_days"`
	MaxAgeDays      int      `yaml:"max_age_days"`
	Sort            string   `yaml:"sort"`
	ScorePreset     string   `yaml:"score_preset"`
	DescWidth       int      `yaml:"desc_width"`
	Watch           bool     `yaml:"watch"`
	IntervalSeconds int      `yaml:"interval_seconds"`
	SnapshotPath    string   `yaml:"snapshot_path"`
	WatchJSONL      string   `yaml:"watch_jsonl"`
}

type scoreConfig struct {
	Sort        string
	ScorePreset string
}

type runConfig struct {
	Queries      []string
	Limit        int
	JSON         bool
	Themes       bool
	MinStars     int
	MinAgeDays   int
	MaxAgeDays   int
	DescWidth    int
	Score        scoreConfig
	Watch        bool
	Interval     time.Duration
	SnapshotPath string
	WatchJSONL   string
}

type snapshotItem struct {
	FullName string `json:"full_name"`
	Stars    int    `json:"stars"`
	Rank     int    `json:"rank"`
}

type snapshotRun struct {
	CapturedAt time.Time      `json:"captured_at"`
	Items      []snapshotItem `json:"items"`
}

type snapshotStore struct {
	Runs []snapshotRun `json:"runs"`
}

type rankMove struct {
	FullName   string
	FromRank   int
	ToRank     int
	DeltaRank  int
	DeltaStars int
}

type deltaReport struct {
	NewRepos []scoredRepo
	Moves    []rankMove
}

func main() {
	cfg := parseAndBuildConfig()

	if cfg.Watch {
		runWatch(cfg)
		return
	}
	scored := runOnce(cfg)
	_ = appendSnapshot(cfg.SnapshotPath, scored, time.Now().UTC())
}

func parseAndBuildConfig() runConfig {
	var queries multiFlag
	var presets multiFlag
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
	var watch bool
	var intervalSeconds int
	var snapshotPath string
	var watchJSONL string

	flag.Var(&queries, "q", "GitHub search query (repeatable)")
	flag.Var(&presets, "preset", "Built-in query preset (repeatable): oss, agents, cli, tui, devtools")
	flag.IntVar(&limit, "n", 15, "Top results to print")
	flag.BoolVar(&jsonOut, "json", false, "Print JSON output")
	flag.BoolVar(&showThemes, "themes", false, "Print theme distribution summary")
	flag.IntVar(&minStars, "min-stars", 0, "Hide repos with stars below this threshold")
	flag.IntVar(&sinceDays, "since-days", 60, "Default query window in days (only used without -q/-preset)")
	flag.IntVar(&minAgeDays, "min-age-days", 0, "Hide repos younger than this age in days")
	flag.IntVar(&maxAgeDays, "max-age-days", 0, "Hide repos older than this age in days")
	flag.StringVar(&sortBy, "sort", "hot", "Sort results by: hot, stars-day, stars, age")
	flag.StringVar(&scorePreset, "score-preset", "hot", "Score preset for -sort hot: hot, fresh")
	flag.StringVar(&configPath, "config", defaultConfigPath(), "Config file path")
	flag.IntVar(&descWidth, "desc-width", defaultDescWidth, "Description column max width for table output")
	flag.BoolVar(&watch, "watch", false, "Run continuously and show deltas between scans")
	flag.IntVar(&intervalSeconds, "interval", defaultIntervalSec, "Watch interval in seconds")
	flag.StringVar(&snapshotPath, "snapshot-path", defaultSnapshotPath(), "Snapshot store path")
	flag.StringVar(&watchJSONL, "watch-jsonl", "", "Append watch delta events as JSONL to this path")
	flag.Parse()

	setFlags := map[string]bool{}
	flag.Visit(func(f *flag.Flag) { setFlags[f.Name] = true })

	fileCfg, err := loadConfig(configPath)
	if err != nil {
		log.Fatal(err)
	}
	applyConfigIfUnset(fileCfg, setFlags, &queries, &presets, &limit, &jsonOut, &showThemes, &minStars, &sinceDays, &minAgeDays, &maxAgeDays, &sortBy, &scorePreset, &descWidth, &watch, &intervalSeconds, &snapshotPath, &watchJSONL)

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
	if intervalSeconds < 15 {
		log.Fatal("-interval must be >= 15 seconds")
	}
	if watch && jsonOut {
		log.Fatal("-watch cannot be combined with -json")
	}

	resolvedQueries, err := resolveQueries([]string(queries), []string(presets), sinceDays, time.Now().UTC())
	if err != nil {
		log.Fatal(err)
	}

	return runConfig{
		Queries:      resolvedQueries,
		Limit:        limit,
		JSON:         jsonOut,
		Themes:       showThemes,
		MinStars:     minStars,
		MinAgeDays:   minAgeDays,
		MaxAgeDays:   maxAgeDays,
		DescWidth:    descWidth,
		Score:        scfg,
		Watch:        watch,
		Interval:     time.Duration(intervalSeconds) * time.Second,
		SnapshotPath: snapshotPath,
		WatchJSONL:   watchJSONL,
	}
}

func runWatch(cfg runConfig) {
	store, _ := loadSnapshotStore(cfg.SnapshotPath)
	var prev []snapshotItem
	if len(store.Runs) > 0 {
		prev = store.Runs[len(store.Runs)-1].Items
	}

	for {
		now := time.Now().UTC()
		scored := runOnce(cfg)
		report := buildDelta(prev, scored)
		if len(prev) > 0 {
			printDelta(os.Stdout, report)
			if err := appendDeltaJSONL(cfg.WatchJSONL, now, report); err != nil {
				log.Printf("watch jsonl warning: %v", err)
			}
		}
		if err := appendSnapshot(cfg.SnapshotPath, scored, now); err != nil {
			log.Printf("snapshot warning: %v", err)
		}
		prev = toSnapshotItems(scored)
		time.Sleep(cfg.Interval)
	}
}

func runOnce(cfg runConfig) []scoredRepo {
	client := &http.Client{Timeout: 20 * time.Second}
	all, err := fetchAndMerge(client, cfg.Queries)
	if err != nil {
		log.Fatalf("fetch trends: %v", err)
	}
	if len(all) == 0 {
		log.Fatal("no repositories returned")
	}

	scored := score(all, cfg.Score)
	if cfg.MinStars > 0 {
		filtered := scored[:0]
		for _, r := range scored {
			if r.StargazersCount >= cfg.MinStars {
				filtered = append(filtered, r)
			}
		}
		scored = filtered
	}
	scored = filterByAge(scored, cfg.MinAgeDays, cfg.MaxAgeDays)
	if len(scored) == 0 {
		log.Fatal("no repositories matched filters")
	}
	if cfg.Limit > len(scored) {
		cfg.Limit = len(scored)
	}
	scored = scored[:cfg.Limit]

	if cfg.JSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(scored); err != nil {
			log.Fatal(err)
		}
		return scored
	}

	printTable(os.Stdout, scored, cfg.DescWidth)
	if cfg.Themes {
		fmt.Println()
		printThemeSummary(scored)
	}
	return scored
}

func resolveQueries(custom []string, presets []string, sinceDays int, now time.Time) ([]string, error) {
	presetQueries, err := expandPresets(presets, sinceDays, now)
	if err != nil {
		return nil, err
	}
	if len(custom) == 0 && len(presetQueries) == 0 {
		return defaultQueries(sinceDays, now), nil
	}
	out := make([]string, 0, len(presetQueries)+len(custom))
	out = append(out, presetQueries...)
	out = append(out, custom...)
	return out, nil
}

func expandPresets(presets []string, sinceDays int, now time.Time) ([]string, error) {
	if len(presets) == 0 {
		return nil, nil
	}
	catalog := presetCatalog(sinceDays, now)
	out := make([]string, 0, len(presets)*2)
	seen := map[string]struct{}{}
	for _, p := range presets {
		name := strings.ToLower(strings.TrimSpace(p))
		queries, ok := catalog[name]
		if !ok {
			return nil, fmt.Errorf("unknown -preset %q (available: %s)", p, strings.Join(listPresetNames(catalog), ", "))
		}
		for _, q := range queries {
			if _, ok := seen[q]; ok {
				continue
			}
			seen[q] = struct{}{}
			out = append(out, q)
		}
	}
	return out, nil
}

func presetCatalog(sinceDays int, now time.Time) map[string][]string {
	if sinceDays < 1 {
		sinceDays = 1
	}
	since := now.AddDate(0, 0, -sinceDays).Format("2006-01-02")
	return map[string][]string{
		"oss":      {"stars:>50 created:>" + since},
		"agents":   {"(agent OR mcp) created:>" + since + " stars:>80"},
		"cli":      {"topic:cli created:>" + since + " stars:>40"},
		"tui":      {"topic:tui created:>" + since + " stars:>20"},
		"devtools": {"(developer tools) created:>" + since + " stars:>50"},
	}
}

func listPresetNames(catalog map[string][]string) []string {
	keys := make([]string, 0, len(catalog))
	for k := range catalog {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func defaultConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil || strings.TrimSpace(home) == "" {
		return ""
	}
	return filepath.Join(home, ".config", "gh-hype-scout", "config.yaml")
}

func defaultSnapshotPath() string {
	cache, err := os.UserCacheDir()
	if err != nil || strings.TrimSpace(cache) == "" {
		return ""
	}
	return filepath.Join(cache, "gh-hype-scout", "snapshots.json")
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

func applyConfigIfUnset(cfg appConfig, setFlags map[string]bool, queries *multiFlag, presets *multiFlag, limit *int, jsonOut *bool, showThemes *bool, minStars *int, sinceDays *int, minAgeDays *int, maxAgeDays *int, sortBy *string, scorePreset *string, descWidth *int, watch *bool, intervalSeconds *int, snapshotPath *string, watchJSONL *string) {
	if !setFlags["q"] && len(cfg.Queries) > 0 {
		*queries = append((*queries)[:0], cfg.Queries...)
	}
	if !setFlags["preset"] && len(cfg.Presets) > 0 {
		*presets = append((*presets)[:0], cfg.Presets...)
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
	if !setFlags["watch"] && cfg.Watch {
		*watch = cfg.Watch
	}
	if !setFlags["interval"] && cfg.IntervalSeconds > 0 {
		*intervalSeconds = cfg.IntervalSeconds
	}
	if !setFlags["snapshot-path"] && strings.TrimSpace(cfg.SnapshotPath) != "" {
		*snapshotPath = cfg.SnapshotPath
	}
	if !setFlags["watch-jsonl"] && strings.TrimSpace(cfg.WatchJSONL) != "" {
		*watchJSONL = cfg.WatchJSONL
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
		out = append(out, scoredRepo{repo: r, AgeDays: age, StarsPerDay: spd, HotScore: hot, Category: categorize(r)})
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

type deltaEvent struct {
	CapturedAt string        `json:"captured_at"`
	NewRepos   []string      `json:"new_repos"`
	RankMoves  []rankMoveLog `json:"rank_moves"`
}

type rankMoveLog struct {
	Repo       string `json:"repo"`
	FromRank   int    `json:"from_rank"`
	ToRank     int    `json:"to_rank"`
	DeltaRank  int    `json:"delta_rank"`
	DeltaStars int    `json:"delta_stars"`
}

func appendDeltaJSONL(path string, now time.Time, report deltaReport) error {
	if strings.TrimSpace(path) == "" {
		return nil
	}
	e := deltaEvent{CapturedAt: now.Format(time.RFC3339)}
	for _, r := range report.NewRepos {
		e.NewRepos = append(e.NewRepos, r.FullName)
	}
	for _, m := range report.Moves {
		e.RankMoves = append(e.RankMoves, rankMoveLog{Repo: m.FullName, FromRank: m.FromRank, ToRank: m.ToRank, DeltaRank: m.DeltaRank, DeltaStars: m.DeltaStars})
	}
	b, err := json.Marshal(e)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err := f.Write(append(b, '\n')); err != nil {
		return err
	}
	return nil
}

func toSnapshotItems(in []scoredRepo) []snapshotItem {
	items := make([]snapshotItem, 0, len(in))
	for i, r := range in {
		items = append(items, snapshotItem{FullName: r.FullName, Stars: r.StargazersCount, Rank: i + 1})
	}
	return items
}

func loadSnapshotStore(path string) (snapshotStore, error) {
	if strings.TrimSpace(path) == "" {
		return snapshotStore{}, nil
	}
	b, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return snapshotStore{}, nil
		}
		return snapshotStore{}, err
	}
	var store snapshotStore
	if err := json.Unmarshal(b, &store); err != nil {
		return snapshotStore{}, nil
	}
	return store, nil
}

func appendSnapshot(path string, current []scoredRepo, now time.Time) error {
	if strings.TrimSpace(path) == "" {
		return nil
	}
	store, err := loadSnapshotStore(path)
	if err != nil {
		return err
	}
	store.Runs = append(store.Runs, snapshotRun{CapturedAt: now, Items: toSnapshotItems(current)})
	if len(store.Runs) > maxSnapshotRuns {
		store.Runs = store.Runs[len(store.Runs)-maxSnapshotRuns:]
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(store, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o644)
}

func buildDelta(prev []snapshotItem, current []scoredRepo) deltaReport {
	report := deltaReport{}
	if len(prev) == 0 || len(current) == 0 {
		return report
	}
	prevByName := map[string]snapshotItem{}
	for _, p := range prev {
		prevByName[p.FullName] = p
	}
	for i, c := range current {
		old, ok := prevByName[c.FullName]
		if !ok {
			report.NewRepos = append(report.NewRepos, c)
			continue
		}
		newRank := i + 1
		if old.Rank != newRank {
			report.Moves = append(report.Moves, rankMove{
				FullName:   c.FullName,
				FromRank:   old.Rank,
				ToRank:     newRank,
				DeltaRank:  old.Rank - newRank,
				DeltaStars: c.StargazersCount - old.Stars,
			})
		}
	}
	sort.Slice(report.Moves, func(i, j int) bool {
		if report.Moves[i].DeltaRank == report.Moves[j].DeltaRank {
			return report.Moves[i].ToRank < report.Moves[j].ToRank
		}
		return report.Moves[i].DeltaRank > report.Moves[j].DeltaRank
	})
	return report
}

func printDelta(w io.Writer, report deltaReport) {
	if len(report.NewRepos) == 0 && len(report.Moves) == 0 {
		fmt.Fprintln(w, "\nΔ No rank changes since previous scan.")
		return
	}
	fmt.Fprintln(w, "\nΔ Changes since previous scan:")
	if len(report.NewRepos) > 0 {
		fmt.Fprintln(w, "  New repos:")
		for _, r := range report.NewRepos {
			fmt.Fprintf(w, "  + %s (%d★)\n", r.FullName, r.StargazersCount)
		}
	}
	if len(report.Moves) > 0 {
		fmt.Fprintln(w, "  Rank movers:")
		for _, m := range report.Moves {
			fmt.Fprintf(w, "  • %s %d→%d (%+d rank, %+d★)\n", m.FullName, m.FromRank, m.ToRank, m.DeltaRank, m.DeltaStars)
		}
	}
}
