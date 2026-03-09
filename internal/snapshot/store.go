package snapshot

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/agent19710101/gh-hype-scout/internal/rank"
)

const MaxRuns = 96

type Item struct {
	FullName string `json:"full_name"`
	Stars    int    `json:"stars"`
	Rank     int    `json:"rank"`
}

type Run struct {
	CapturedAt time.Time `json:"captured_at"`
	Items      []Item    `json:"items"`
}

type Store struct {
	Runs []Run `json:"runs"`
}

func Load(path string) (Store, error) {
	if strings.TrimSpace(path) == "" {
		return Store{}, nil
	}
	b, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Store{}, nil
		}
		return Store{}, err
	}
	var s Store
	if err := json.Unmarshal(b, &s); err != nil {
		return Store{}, nil
	}
	return s, nil
}

func Append(path string, scored []rank.Repo, now time.Time) error {
	if strings.TrimSpace(path) == "" {
		return nil
	}
	s, err := Load(path)
	if err != nil {
		return err
	}
	s.Runs = append(s.Runs, Run{CapturedAt: now, Items: FromRanked(scored)})
	if len(s.Runs) > MaxRuns {
		s.Runs = s.Runs[len(s.Runs)-MaxRuns:]
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o644)
}

func FromRanked(in []rank.Repo) []Item {
	items := make([]Item, 0, len(in))
	for i, r := range in {
		items = append(items, Item{FullName: r.FullName, Stars: r.StargazersCount, Rank: i + 1})
	}
	return items
}
