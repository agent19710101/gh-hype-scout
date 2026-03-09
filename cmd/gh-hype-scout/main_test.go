package main

import (
	"net/http"
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
	out := score(in)
	if len(out) != 2 {
		t.Fatalf("expected 2 scored repos, got %d", len(out))
	}
	if out[0].FullName != "b/fast" {
		t.Fatalf("expected fastest repo first, got %s", out[0].FullName)
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
