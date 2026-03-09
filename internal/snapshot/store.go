package snapshot

import (
	"encoding/json"
	"errors"
	"io"
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

func Export(srcPath, dstPath string) error {
	if strings.TrimSpace(srcPath) == "" || strings.TrimSpace(dstPath) == "" {
		return nil
	}
	in, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer in.Close()
	if err := os.MkdirAll(filepath.Dir(dstPath), 0o755); err != nil {
		return err
	}
	out, err := os.Create(dstPath)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	return err
}

type DiffResult struct {
	Added   int
	Removed int
	Changed int
}

func Diff(pathA, pathB string) (DiffResult, error) {
	a, err := Load(pathA)
	if err != nil {
		return DiffResult{}, err
	}
	b, err := Load(pathB)
	if err != nil {
		return DiffResult{}, err
	}
	if len(a.Runs) == 0 || len(b.Runs) == 0 {
		return DiffResult{}, nil
	}
	am := map[string]int{}
	for _, it := range a.Runs[len(a.Runs)-1].Items {
		am[it.FullName] = it.Stars
	}
	bm := map[string]int{}
	for _, it := range b.Runs[len(b.Runs)-1].Items {
		bm[it.FullName] = it.Stars
	}
	var out DiffResult
	for k, av := range am {
		bv, ok := bm[k]
		if !ok {
			out.Removed++
			continue
		}
		if av != bv {
			out.Changed++
		}
	}
	for k := range bm {
		if _, ok := am[k]; !ok {
			out.Added++
		}
	}
	return out, nil
}

func Import(srcPath, dstPath string) error {
	if strings.TrimSpace(srcPath) == "" || strings.TrimSpace(dstPath) == "" {
		return nil
	}
	in, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer in.Close()
	b, err := io.ReadAll(in)
	if err != nil {
		return err
	}
	var s Store
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(dstPath), 0o755); err != nil {
		return err
	}
	return os.WriteFile(dstPath, b, 0o644)
}

func FromRanked(in []rank.Repo) []Item {
	items := make([]Item, 0, len(in))
	for i, r := range in {
		items = append(items, Item{FullName: r.FullName, Stars: r.StargazersCount, Rank: i + 1})
	}
	return items
}
