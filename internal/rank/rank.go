package rank

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/agent19710101/gh-hype-scout/internal/githubapi"
)

type Config struct {
	Sort        string
	ScorePreset string
}

type Repo struct {
	githubapi.Repo
	AgeDays     float64 `json:"age_days"`
	StarsPerDay float64 `json:"stars_per_day"`
	HotScore    float64 `json:"hot_score"`
	Category    string  `json:"category"`
}

func Validate(c Config) error {
	switch c.Sort {
	case "hot", "stars-day", "stars", "age":
	default:
		return fmt.Errorf("invalid -sort %q (expected: hot, stars-day, stars, age)", c.Sort)
	}
	switch c.ScorePreset {
	case "hot", "fresh":
	default:
		return fmt.Errorf("invalid -score-preset %q (expected: hot, fresh)", c.ScorePreset)
	}
	if c.Sort != "hot" && c.ScorePreset != "hot" {
		return fmt.Errorf("-score-preset=%q only applies with -sort hot", c.ScorePreset)
	}
	return nil
}

func Score(in []githubapi.Repo, cfg Config) []Repo {
	now := time.Now().UTC()
	out := make([]Repo, 0, len(in))
	for _, r := range in {
		age := now.Sub(r.CreatedAt).Hours() / 24
		if age < 1 {
			age = 1
		}
		spd := float64(r.StargazersCount) / age
		hot := spd * math.Log10(float64(r.StargazersCount)+1)
		if cfg.ScorePreset == "fresh" {
			hot *= 1 + (1 / math.Sqrt(age))
		}
		out = append(out, Repo{Repo: r, AgeDays: age, StarsPerDay: spd, HotScore: hot, Category: categorize(r)})
	}
	sortScored(out, cfg.Sort)
	return out
}

func Filter(in []Repo, minStars, minAge, maxAge int, limit int) ([]Repo, error) {
	if minAge < 0 || maxAge < 0 {
		return nil, fmt.Errorf("age filters must be >= 0")
	}
	if maxAge > 0 && minAge > maxAge {
		return nil, fmt.Errorf("min-age-days (%d) cannot be greater than max-age-days (%d)", minAge, maxAge)
	}
	out := in[:0]
	for _, r := range in {
		if minStars > 0 && r.StargazersCount < minStars {
			continue
		}
		if minAge > 0 && r.AgeDays < float64(minAge) {
			continue
		}
		if maxAge > 0 && r.AgeDays > float64(maxAge) {
			continue
		}
		out = append(out, r)
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("no repositories matched filters")
	}
	if limit > len(out) {
		limit = len(out)
	}
	return out[:limit], nil
}

func categorize(r githubapi.Repo) string {
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

func sortScored(in []Repo, sortBy string) {
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
