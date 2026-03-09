package main

import (
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
