package main

import (
	"bytes"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestDefaultQueriesUsesSinceDays(t *testing.T) {
	now := time.Date(2026, 3, 9, 1, 30, 0, 0, time.UTC)
	queries := defaultQueries(30, now)

	if len(queries) != 4 {
		t.Fatalf("expected 4 default queries, got %d", len(queries))
	}

	for _, q := range queries {
		if !strings.Contains(q, "created:>2026-02-07") {
			t.Fatalf("query does not include expected since date: %q", q)
		}
	}
}

func TestDefaultQueriesClampsSinceDays(t *testing.T) {
	now := time.Date(2026, 3, 9, 1, 30, 0, 0, time.UTC)
	queries := defaultQueries(0, now)

	for _, q := range queries {
		if !strings.Contains(q, "created:>2026-03-08") {
			t.Fatalf("expected clamped one-day window in query: %q", q)
		}
	}
}

func TestScoreSortsDescendingByHotScore(t *testing.T) {
	now := time.Now().UTC()
	in := []repo{
		{FullName: "a/slow", StargazersCount: 100, CreatedAt: now.AddDate(0, 0, -100)},
		{FullName: "b/fast", StargazersCount: 100, CreatedAt: now.AddDate(0, 0, -10)},
	}
	out := score(in, scoreConfig{Sort: "hot", ScorePreset: "hot"})
	if len(out) != 2 {
		t.Fatalf("expected 2 scored repos, got %d", len(out))
	}
	if out[0].FullName != "b/fast" {
		t.Fatalf("expected fastest repo first, got %s", out[0].FullName)
	}
}

func TestScoreSortAge(t *testing.T) {
	now := time.Now().UTC()
	in := []repo{
		{FullName: "a/older", StargazersCount: 100, CreatedAt: now.AddDate(0, 0, -20)},
		{FullName: "b/newer", StargazersCount: 100, CreatedAt: now.AddDate(0, 0, -5)},
	}
	out := score(in, scoreConfig{Sort: "age", ScorePreset: "hot"})
	if out[0].FullName != "b/newer" {
		t.Fatalf("expected youngest repo first, got %s", out[0].FullName)
	}
}

func TestFreshPresetChangesScore(t *testing.T) {
	now := time.Now().UTC()
	in := []repo{{FullName: "a/repo", StargazersCount: 100, CreatedAt: now.AddDate(0, 0, -30)}}
	hot := score(in, scoreConfig{Sort: "hot", ScorePreset: "hot"})[0].HotScore
	fresh := score(in, scoreConfig{Sort: "hot", ScorePreset: "fresh"})[0].HotScore
	if fresh <= hot {
		t.Fatalf("expected fresh preset to increase score, hot=%.4f fresh=%.4f", hot, fresh)
	}
}

func TestValidateAgeFlags(t *testing.T) {
	tests := []struct {
		name    string
		minAge  int
		maxAge  int
		wantErr bool
	}{
		{name: "valid disabled", minAge: 0, maxAge: 0, wantErr: false},
		{name: "valid range", minAge: 7, maxAge: 30, wantErr: false},
		{name: "negative min", minAge: -1, maxAge: 0, wantErr: true},
		{name: "min greater than max", minAge: 31, maxAge: 30, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateAgeFlags(tt.minAge, tt.maxAge)
			if (err != nil) != tt.wantErr {
				t.Fatalf("validateAgeFlags() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFilterByAge(t *testing.T) {
	in := []scoredRepo{
		{repo: repo{FullName: "a/new"}, AgeDays: 2},
		{repo: repo{FullName: "b/mid"}, AgeDays: 10},
		{repo: repo{FullName: "c/old"}, AgeDays: 40},
	}

	out := filterByAge(append([]scoredRepo(nil), in...), 7, 30)
	if len(out) != 1 || out[0].FullName != "b/mid" {
		t.Fatalf("unexpected filtered result: %#v", out)
	}
}

func TestGithubRateLimitHint(t *testing.T) {
	resp := &http.Response{StatusCode: http.StatusTooManyRequests, Header: http.Header{"Retry-After": []string{"42"}}}
	hint := githubRateLimitHint(resp)
	if !strings.Contains(hint, "retry after 42s") {
		t.Fatalf("expected retry-after hint, got %q", hint)
	}
}

func TestCategorize(t *testing.T) {
	tests := []struct {
		name string
		repo repo
		want string
	}{
		{name: "agent", repo: repo{FullName: "org/tool", Description: "MCP agent runtime"}, want: "agent"},
		{name: "cli", repo: repo{FullName: "org/term", Description: "terminal cli helper"}, want: "cli"},
		{name: "tui", repo: repo{FullName: "org/tuiapp", Description: "beautiful TUI"}, want: "tui"},
		{name: "devtool", repo: repo{FullName: "org/repo", Description: "github dev workflow"}, want: "devtool"},
		{name: "general", repo: repo{FullName: "org/other", Description: "random"}, want: "general"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := categorize(tt.repo)
			if got != tt.want {
				t.Fatalf("categorize() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestValidateScoreConfig(t *testing.T) {
	if err := validateScoreConfig(scoreConfig{Sort: "hot", ScorePreset: "hot"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := validateScoreConfig(scoreConfig{Sort: "age", ScorePreset: "hot"}); err != nil {
		t.Fatalf("unexpected age+hot error: %v", err)
	}
	if err := validateScoreConfig(scoreConfig{Sort: "stars", ScorePreset: "fresh"}); err == nil {
		t.Fatal("expected combination error for stars+fresh")
	}
}

func TestSortScoredByStars(t *testing.T) {
	in := []scoredRepo{
		{repo: repo{FullName: "a/one", StargazersCount: 50}, StarsPerDay: 10, HotScore: 20},
		{repo: repo{FullName: "b/two", StargazersCount: 200}, StarsPerDay: 4, HotScore: 8},
	}
	sortScored(in, "stars")
	if in[0].FullName != "b/two" {
		t.Fatalf("expected highest stars first, got %s", in[0].FullName)
	}
}

func TestSortScoredByStarsDay(t *testing.T) {
	in := []scoredRepo{
		{repo: repo{FullName: "a/one", StargazersCount: 200}, StarsPerDay: 4, HotScore: 8},
		{repo: repo{FullName: "b/two", StargazersCount: 50}, StarsPerDay: 10, HotScore: 20},
	}
	sortScored(in, "stars-day")
	if in[0].FullName != "b/two" {
		t.Fatalf("expected highest stars/day first, got %s", in[0].FullName)
	}
}

func TestFetchAndMergeDedupe(t *testing.T) {
	searcher := func(_ *http.Client, q string) ([]repo, error) {
		switch q {
		case "q1":
			return []repo{{FullName: "org/a", StargazersCount: 10}, {FullName: "org/b", StargazersCount: 20}}, nil
		case "q2":
			return []repo{{FullName: "org/a", StargazersCount: 30}, {FullName: "org/c", StargazersCount: 5}}, nil
		default:
			return nil, nil
		}
	}
	out, err := fetchAndMergeWithSearcher(&http.Client{}, []string{"q1", "q2"}, searcher)
	if err != nil {
		t.Fatalf("fetchAndMergeWithSearcher error: %v", err)
	}
	if len(out) != 3 {
		t.Fatalf("expected 3 repos after dedupe, got %d", len(out))
	}

	m := map[string]int{}
	for _, r := range out {
		m[r.FullName] = r.StargazersCount
	}
	if m["org/a"] != 30 {
		t.Fatalf("expected dedupe to keep highest stars for org/a, got %d", m["org/a"])
	}
}

func TestExpandPresets(t *testing.T) {
	now := time.Date(2026, 3, 9, 1, 30, 0, 0, time.UTC)
	qs, err := expandPresets([]string{"cli", "agents"}, 30, now)
	if err != nil {
		t.Fatalf("expandPresets error: %v", err)
	}
	if len(qs) != 2 {
		t.Fatalf("expected 2 preset queries, got %d", len(qs))
	}
	if !strings.Contains(qs[0], "2026-02-07") || !strings.Contains(qs[1], "2026-02-07") {
		t.Fatalf("expected since date in preset queries, got %#v", qs)
	}
}

func TestExpandPresetsUnknown(t *testing.T) {
	_, err := expandPresets([]string{"nope"}, 30, time.Now().UTC())
	if err == nil {
		t.Fatal("expected error for unknown preset")
	}
	if !strings.Contains(err.Error(), "unknown -preset") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestResolveQueriesMergesPresetAndCustom(t *testing.T) {
	now := time.Date(2026, 3, 9, 1, 30, 0, 0, time.UTC)
	out, err := resolveQueries([]string{"custom:query"}, []string{"cli"}, 30, now)
	if err != nil {
		t.Fatalf("resolveQueries error: %v", err)
	}
	if len(out) != 2 {
		t.Fatalf("expected 2 queries, got %d", len(out))
	}
	if out[1] != "custom:query" {
		t.Fatalf("expected custom query after presets, got %#v", out)
	}
}

func TestResolveQueriesFallsBackToDefault(t *testing.T) {
	now := time.Date(2026, 3, 9, 1, 30, 0, 0, time.UTC)
	out, err := resolveQueries(nil, nil, 30, now)
	if err != nil {
		t.Fatalf("resolveQueries error: %v", err)
	}
	if len(out) != 4 {
		t.Fatalf("expected default 4 queries, got %d", len(out))
	}
}

func TestPrintTableGolden(t *testing.T) {
	in := []scoredRepo{
		{repo: repo{FullName: "org/alpha", Description: "A very long description that should be truncated to keep terminal output compact.", StargazersCount: 1200, Language: "Go"}, AgeDays: 12.3, StarsPerDay: 97.6, HotScore: 300.2, Category: "cli"},
		{repo: repo{FullName: "org/beta", Description: "", StargazersCount: 500, Language: "Rust"}, AgeDays: 25.0, StarsPerDay: 20.0, HotScore: 54.2, Category: "tui"},
	}

	var b bytes.Buffer
	printTable(&b, in, 28)

	goldenPath := filepath.Join("testdata", "table.golden")
	want, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("read golden: %v", err)
	}
	if strings.TrimSpace(b.String()) != strings.TrimSpace(string(want)) {
		t.Fatalf("table output mismatch\n--- got ---\n%s\n--- want ---\n%s", b.String(), string(want))
	}
}

func TestSnapshotStoreCorruptIsNonFatal(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "snapshots.json")
	if err := os.WriteFile(path, []byte("{not-json"), 0o644); err != nil {
		t.Fatalf("write corrupt snapshot: %v", err)
	}
	store, err := loadSnapshotStore(path)
	if err != nil {
		t.Fatalf("loadSnapshotStore returned error for corrupt file: %v", err)
	}
	if len(store.Runs) != 0 {
		t.Fatalf("expected empty store for corrupt snapshot, got %d runs", len(store.Runs))
	}
}

func TestSnapshotStoreRetention(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "snapshots.json")
	for i := 0; i < maxSnapshotRuns+7; i++ {
		in := []scoredRepo{{repo: repo{FullName: "org/a", StargazersCount: i + 1}}}
		if err := appendSnapshot(path, in, time.Unix(int64(i), 0).UTC()); err != nil {
			t.Fatalf("appendSnapshot: %v", err)
		}
	}
	store, err := loadSnapshotStore(path)
	if err != nil {
		t.Fatalf("loadSnapshotStore: %v", err)
	}
	if len(store.Runs) != maxSnapshotRuns {
		t.Fatalf("expected %d retained runs, got %d", maxSnapshotRuns, len(store.Runs))
	}
}

func TestBuildDelta(t *testing.T) {
	prev := []snapshotItem{{FullName: "org/a", Stars: 100, Rank: 2}, {FullName: "org/b", Stars: 90, Rank: 1}}
	cur := []scoredRepo{
		{repo: repo{FullName: "org/a", StargazersCount: 110}},
		{repo: repo{FullName: "org/c", StargazersCount: 50}},
	}
	report := buildDelta(prev, cur)
	if len(report.NewRepos) != 1 || report.NewRepos[0].FullName != "org/c" {
		t.Fatalf("expected org/c as new repo, got %#v", report.NewRepos)
	}
	if len(report.Moves) != 1 || report.Moves[0].FullName != "org/a" {
		t.Fatalf("expected one move for org/a, got %#v", report.Moves)
	}
	if report.Moves[0].DeltaStars != 10 {
		t.Fatalf("expected delta stars 10, got %d", report.Moves[0].DeltaStars)
	}
}
