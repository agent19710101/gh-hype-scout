package config

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

const (
	DefaultDescWidth   = 56
	DefaultIntervalSec = 300
)

type File struct {
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

type Run struct {
	Queries      []string
	Presets      []string
	Limit        int
	JSON         bool
	Themes       bool
	MinStars     int
	SinceDays    int
	MinAgeDays   int
	MaxAgeDays   int
	Sort         string
	ScorePreset  string
	DescWidth    int
	Watch        bool
	Interval     time.Duration
	SnapshotPath string
	WatchJSONL   string
}

func DefaultConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil || strings.TrimSpace(home) == "" {
		return ""
	}
	return filepath.Join(home, ".config", "gh-hype-scout", "config.yaml")
}

func DefaultSnapshotPath() string {
	cache, err := os.UserCacheDir()
	if err != nil || strings.TrimSpace(cache) == "" {
		return ""
	}
	return filepath.Join(cache, "gh-hype-scout", "snapshots.json")
}

func Parse() (Run, error) {
	cfg := Run{}
	var configPath string
	flag.Var((*multiFlag)(&cfg.Queries), "q", "GitHub search query (repeatable)")
	flag.Var((*multiFlag)(&cfg.Presets), "preset", "Built-in query preset (repeatable): oss, agents, cli, tui, devtools")
	flag.IntVar(&cfg.Limit, "n", 15, "Top results to print")
	flag.BoolVar(&cfg.JSON, "json", false, "Print JSON output")
	flag.BoolVar(&cfg.Themes, "themes", false, "Print theme distribution summary")
	flag.IntVar(&cfg.MinStars, "min-stars", 0, "Hide repos with stars below this threshold")
	flag.IntVar(&cfg.SinceDays, "since-days", 60, "Default query window in days (only used without -q/-preset)")
	flag.IntVar(&cfg.MinAgeDays, "min-age-days", 0, "Hide repos younger than this age in days")
	flag.IntVar(&cfg.MaxAgeDays, "max-age-days", 0, "Hide repos older than this age in days")
	flag.StringVar(&cfg.Sort, "sort", "hot", "Sort results by: hot, stars-day, stars, age, accel")
	flag.StringVar(&cfg.ScorePreset, "score-preset", "hot", "Score preset for -sort hot: hot, fresh")
	flag.StringVar(&configPath, "config", DefaultConfigPath(), "Config file path")
	flag.IntVar(&cfg.DescWidth, "desc-width", DefaultDescWidth, "Description column max width for table output")
	flag.BoolVar(&cfg.Watch, "watch", false, "Run continuously and show deltas between scans")
	intervalSeconds := DefaultIntervalSec
	flag.IntVar(&intervalSeconds, "interval", DefaultIntervalSec, "Watch interval in seconds")
	flag.StringVar(&cfg.SnapshotPath, "snapshot-path", DefaultSnapshotPath(), "Snapshot store path")
	flag.StringVar(&cfg.WatchJSONL, "watch-jsonl", "", "Append watch delta events as JSONL to this path")
	flag.Parse()

	set := map[string]bool{}
	flag.Visit(func(f *flag.Flag) { set[f.Name] = true })
	fc, err := loadFile(configPath)
	if err != nil {
		return Run{}, err
	}
	merge(&cfg, fc, set)
	cfg.Interval = time.Duration(intervalSeconds) * time.Second
	if err := validate(cfg, intervalSeconds); err != nil {
		return Run{}, err
	}
	return cfg, nil
}

type multiFlag []string

func (m *multiFlag) String() string { return strings.Join(*m, ",") }
func (m *multiFlag) Set(v string) error {
	*m = append(*m, v)
	return nil
}

func loadFile(path string) (File, error) {
	if strings.TrimSpace(path) == "" {
		return File{}, nil
	}
	b, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return File{}, nil
		}
		return File{}, fmt.Errorf("read config %q: %w", path, err)
	}
	var cfg File
	if err := yaml.Unmarshal(b, &cfg); err != nil {
		return File{}, fmt.Errorf("parse config %q: %w", path, err)
	}
	return cfg, nil
}

func merge(cfg *Run, fc File, set map[string]bool) {
	if !set["q"] && len(fc.Queries) > 0 {
		cfg.Queries = append([]string{}, fc.Queries...)
	}
	if !set["preset"] && len(fc.Presets) > 0 {
		cfg.Presets = append([]string{}, fc.Presets...)
	}
	if !set["n"] && fc.Limit > 0 {
		cfg.Limit = fc.Limit
	}
	if !set["json"] && fc.JSON {
		cfg.JSON = fc.JSON
	}
	if !set["themes"] && fc.Themes {
		cfg.Themes = fc.Themes
	}
	if !set["min-stars"] && fc.MinStars > 0 {
		cfg.MinStars = fc.MinStars
	}
	if !set["since-days"] && fc.SinceDays > 0 {
		cfg.SinceDays = fc.SinceDays
	}
	if !set["min-age-days"] && fc.MinAgeDays > 0 {
		cfg.MinAgeDays = fc.MinAgeDays
	}
	if !set["max-age-days"] && fc.MaxAgeDays > 0 {
		cfg.MaxAgeDays = fc.MaxAgeDays
	}
	if !set["sort"] && strings.TrimSpace(fc.Sort) != "" {
		cfg.Sort = fc.Sort
	}
	if !set["score-preset"] && strings.TrimSpace(fc.ScorePreset) != "" {
		cfg.ScorePreset = fc.ScorePreset
	}
	if !set["desc-width"] && fc.DescWidth > 0 {
		cfg.DescWidth = fc.DescWidth
	}
	if !set["watch"] && fc.Watch {
		cfg.Watch = fc.Watch
	}
	if !set["snapshot-path"] && strings.TrimSpace(fc.SnapshotPath) != "" {
		cfg.SnapshotPath = fc.SnapshotPath
	}
	if !set["watch-jsonl"] && strings.TrimSpace(fc.WatchJSONL) != "" {
		cfg.WatchJSONL = fc.WatchJSONL
	}
}

func validate(cfg Run, intervalSeconds int) error {
	if cfg.MinAgeDays < 0 || cfg.MaxAgeDays < 0 {
		return fmt.Errorf("age filters must be >= 0")
	}
	if cfg.MaxAgeDays > 0 && cfg.MinAgeDays > cfg.MaxAgeDays {
		return fmt.Errorf("min-age-days (%d) cannot be greater than max-age-days (%d)", cfg.MinAgeDays, cfg.MaxAgeDays)
	}
	if cfg.DescWidth < 8 {
		return fmt.Errorf("-desc-width must be >= 8")
	}
	if intervalSeconds < 15 {
		return fmt.Errorf("-interval must be >= 15 seconds")
	}
	if cfg.Watch && cfg.JSON {
		return fmt.Errorf("-watch cannot be combined with -json")
	}
	return nil
}
