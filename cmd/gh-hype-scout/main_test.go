package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/agent19710101/gh-hype-scout/internal/githubapi"
	"github.com/agent19710101/gh-hype-scout/internal/output"
	"github.com/agent19710101/gh-hype-scout/internal/query"
	"github.com/agent19710101/gh-hype-scout/internal/rank"
	"github.com/agent19710101/gh-hype-scout/internal/snapshot"
)

func TestResolveQueriesDefault(t *testing.T) {
	now := time.Date(2026, 3, 9, 1, 30, 0, 0, time.UTC)
	qs, err := query.Resolve(nil, nil, 30, now)
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if len(qs) != 4 {
		t.Fatalf("expected 4 default queries, got %d", len(qs))
	}
}

func TestResolveUnknownPreset(t *testing.T) {
	_, err := query.Resolve(nil, []string{"nope"}, 30, time.Now().UTC())
	if err == nil || !strings.Contains(err.Error(), "unknown -preset") {
		t.Fatalf("expected unknown preset error, got %v", err)
	}
}

func TestSnapshotRetention(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "snapshots.json")
	for i := 0; i < snapshot.MaxRuns+5; i++ {
		in := []rank.Repo{{Repo: githubapi.Repo{FullName: "org/a", StargazersCount: i + 1}, AgeDays: 1}}
		if err := snapshot.Append(path, in, time.Unix(int64(i), 0)); err != nil {
			t.Fatalf("append: %v", err)
		}
	}
	s, err := snapshot.Load(path)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if len(s.Runs) != snapshot.MaxRuns {
		t.Fatalf("expected %d runs, got %d", snapshot.MaxRuns, len(s.Runs))
	}
}

func TestAppendDeltaJSONL(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "watch.jsonl")
	report := output.DeltaReport{Moves: []output.RankMove{{FullName: "org/a", FromRank: 3, ToRank: 1, DeltaRank: 2, DeltaStars: 10}}}
	if err := output.AppendDeltaJSONL(path, time.Now().UTC(), report); err != nil {
		t.Fatalf("append jsonl: %v", err)
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read jsonl: %v", err)
	}
	if !strings.Contains(string(b), "org/a") {
		t.Fatalf("expected org/a in jsonl, got %s", string(b))
	}
}
