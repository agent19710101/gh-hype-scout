package config

import (
	"errors"
	"flag"
	"fmt"
	"net/url"
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

type PresetProfile struct {
	Sort        string  `yaml:"sort"`
	ScorePreset string  `yaml:"score_preset"`
	MinStars    int     `yaml:"min_stars"`
	AlertAccel  float64 `yaml:"alert_accel"`
}

type RoutingProfile struct {
	WatchJSONL   string `yaml:"watch_jsonl"`
	WatchWebhook string `yaml:"watch_webhook"`
}

type File struct {
	Queries         []string                  `yaml:"queries"`
	Presets         []string                  `yaml:"presets"`
	PresetOverrides map[string][]string       `yaml:"preset_overrides"`
	PresetProfiles  map[string]PresetProfile  `yaml:"preset_profiles"`
	RoutingProfiles map[string]RoutingProfile `yaml:"routing_profiles"`
	Limit           int                       `yaml:"limit"`
	JSON            bool                      `yaml:"json"`
	Themes          bool                      `yaml:"themes"`
	MinStars        int                       `yaml:"min_stars"`
	SinceDays       int                       `yaml:"since_days"`
	MinAgeDays      int                       `yaml:"min_age_days"`
	MaxAgeDays      int                       `yaml:"max_age_days"`
	Sort            string                    `yaml:"sort"`
	ScorePreset     string                    `yaml:"score_preset"`
	DescWidth       int                       `yaml:"desc_width"`
	Watch           bool                      `yaml:"watch"`
	IntervalSeconds int                       `yaml:"interval_seconds"`
	SnapshotPath    string                    `yaml:"snapshot_path"`
	WatchJSONL      string                    `yaml:"watch_jsonl"`
	WatchWebhook    string                    `yaml:"watch_webhook"`
	WatchAuthToken  string                    `yaml:"watch_auth_token"`
	WatchSignSecret string                    `yaml:"watch_sign_secret"`
	UI              string                    `yaml:"ui"`
	SnapshotExport  string                    `yaml:"snapshot_export"`
	SnapshotImport  string                    `yaml:"snapshot_import"`
	SnapshotDiff    string                    `yaml:"snapshot_diff"`
	MomentumModel   string                    `yaml:"momentum_model"`
	RoutingProfile  string                    `yaml:"routing_profile"`
	PluginCmd       string                    `yaml:"plugin_cmd"`
}

type Run struct {
	Queries         []string
	Presets         []string
	PresetOverrides map[string][]string
	Limit           int
	JSON            bool
	Themes          bool
	MinStars        int
	SinceDays       int
	MinAgeDays      int
	MaxAgeDays      int
	Sort            string
	ScorePreset     string
	DescWidth       int
	Watch           bool
	Interval        time.Duration
	SnapshotPath    string
	WatchJSONL      string
	WatchWebhook    string
	WatchAuthToken  string
	WatchSignSecret string
	UIMode          string
	SnapshotExport  string
	SnapshotImport  string
	SnapshotDiff    string
	AlertAccel      float64
	MomentumModel   string
	RoutingProfile  string
	PluginCmd       string
	Explicit        map[string]bool
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
	flag.StringVar(&cfg.WatchWebhook, "watch-webhook", "", "POST watch delta events to this webhook URL")
	flag.StringVar(&cfg.WatchAuthToken, "watch-auth-token", "", "Bearer token for watch webhook authorization")
	flag.StringVar(&cfg.WatchSignSecret, "watch-sign-secret", "", "HMAC secret for watch webhook payload signing")
	flag.StringVar(&cfg.UIMode, "ui", "stdout", "UI mode: stdout or tui")
	flag.StringVar(&cfg.SnapshotExport, "snapshot-export", "", "Export snapshots to a JSON file and exit")
	flag.StringVar(&cfg.SnapshotImport, "snapshot-import", "", "Import snapshots from a JSON file and exit")
	flag.StringVar(&cfg.SnapshotDiff, "snapshot-diff", "", "Compare two snapshot files: pathA:pathB")
	flag.StringVar(&cfg.MomentumModel, "momentum-model", "baseline", "Momentum model: baseline, decay, trend")
	flag.StringVar(&cfg.RoutingProfile, "routing-profile", "", "Routing profile to apply from config")
	flag.StringVar(&cfg.PluginCmd, "plugin-cmd", "", "External plugin command to process watch events")
	flag.Parse()

	set := map[string]bool{}
	flag.Visit(func(f *flag.Flag) { set[f.Name] = true })
	cfg.Explicit = set
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
	if len(fc.PresetOverrides) > 0 {
		cfg.PresetOverrides = make(map[string][]string, len(fc.PresetOverrides))
		for k, v := range fc.PresetOverrides {
			cfg.PresetOverrides[k] = append([]string{}, v...)
		}
	}
	if len(fc.PresetProfiles) > 0 {
		for _, p := range cfg.Presets {
			prof, ok := fc.PresetProfiles[strings.ToLower(strings.TrimSpace(p))]
			if !ok {
				continue
			}
			if !cfg.Explicit["sort"] && strings.TrimSpace(prof.Sort) != "" {
				cfg.Sort = prof.Sort
			}
			if !cfg.Explicit["score-preset"] && strings.TrimSpace(prof.ScorePreset) != "" {
				cfg.ScorePreset = prof.ScorePreset
			}
			if !cfg.Explicit["min-stars"] && prof.MinStars > 0 {
				cfg.MinStars = prof.MinStars
			}
			if prof.AlertAccel > 0 {
				cfg.AlertAccel = prof.AlertAccel
			}
			break
		}
	}
	if !set["routing-profile"] && strings.TrimSpace(cfg.RoutingProfile) == "" && len(fc.RoutingProfiles) > 0 {
		for name := range fc.RoutingProfiles {
			cfg.RoutingProfile = name
			break
		}
	}
	if strings.TrimSpace(cfg.RoutingProfile) != "" {
		if rp, ok := fc.RoutingProfiles[cfg.RoutingProfile]; ok {
			if !set["watch-jsonl"] && strings.TrimSpace(rp.WatchJSONL) != "" {
				cfg.WatchJSONL = rp.WatchJSONL
			}
			if !set["watch-webhook"] && strings.TrimSpace(rp.WatchWebhook) != "" {
				cfg.WatchWebhook = rp.WatchWebhook
			}
		}
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
	if !set["watch-webhook"] && strings.TrimSpace(fc.WatchWebhook) != "" {
		cfg.WatchWebhook = fc.WatchWebhook
	}
	if !set["watch-auth-token"] && strings.TrimSpace(fc.WatchAuthToken) != "" {
		cfg.WatchAuthToken = fc.WatchAuthToken
	}
	if !set["watch-sign-secret"] && strings.TrimSpace(fc.WatchSignSecret) != "" {
		cfg.WatchSignSecret = fc.WatchSignSecret
	}
	if !set["ui"] && strings.TrimSpace(fc.UI) != "" {
		cfg.UIMode = fc.UI
	}
	if !set["snapshot-export"] && strings.TrimSpace(fc.SnapshotExport) != "" {
		cfg.SnapshotExport = fc.SnapshotExport
	}
	if !set["snapshot-import"] && strings.TrimSpace(fc.SnapshotImport) != "" {
		cfg.SnapshotImport = fc.SnapshotImport
	}
	if !set["snapshot-diff"] && strings.TrimSpace(fc.SnapshotDiff) != "" {
		cfg.SnapshotDiff = fc.SnapshotDiff
	}
	if !set["momentum-model"] && strings.TrimSpace(fc.MomentumModel) != "" {
		cfg.MomentumModel = fc.MomentumModel
	}
	if !set["plugin-cmd"] && strings.TrimSpace(fc.PluginCmd) != "" {
		cfg.PluginCmd = fc.PluginCmd
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
	if strings.TrimSpace(cfg.WatchWebhook) != "" {
		u, err := url.Parse(cfg.WatchWebhook)
		if err != nil || u.Scheme == "" || u.Host == "" {
			return fmt.Errorf("invalid -watch-webhook URL")
		}
	}
	switch strings.ToLower(strings.TrimSpace(cfg.UIMode)) {
	case "", "stdout", "tui":
	default:
		return fmt.Errorf("invalid -ui %q (expected: stdout, tui)", cfg.UIMode)
	}
	switch strings.ToLower(strings.TrimSpace(cfg.MomentumModel)) {
	case "", "baseline", "decay", "trend":
	default:
		return fmt.Errorf("invalid -momentum-model %q (expected: baseline, decay, trend)", cfg.MomentumModel)
	}
	return nil
}
