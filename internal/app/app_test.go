package app

import (
	"testing"
	"time"

	"github.com/agent19710101/gh-hype-scout/internal/githubapi"
	"github.com/agent19710101/gh-hype-scout/internal/rank"
	"github.com/agent19710101/gh-hype-scout/internal/snapshot"
)

func TestDeriveAccelerationMissingHistory(t *testing.T) {
	cur := []rank.Repo{{Repo: githubapi.Repo{FullName: "org/a", StargazersCount: 120}}}
	acc := deriveAcceleration(snapshot.Store{}, cur, time.Now().UTC(), "baseline")
	if len(acc) != 0 {
		t.Fatalf("expected empty acceleration map without history, got %#v", acc)
	}
}

func TestDeriveAccelerationDeterministic(t *testing.T) {
	t0 := time.Date(2026, 3, 9, 10, 0, 0, 0, time.UTC)
	t1 := t0.Add(2 * time.Hour)
	now := t1.Add(2 * time.Hour)
	store := snapshot.Store{Runs: []snapshot.Run{
		{CapturedAt: t0, Items: []snapshot.Item{{FullName: "org/a", Stars: 100, Rank: 1}}},
		{CapturedAt: t1, Items: []snapshot.Item{{FullName: "org/a", Stars: 130, Rank: 1}}},
	}}
	cur := []rank.Repo{{Repo: githubapi.Repo{FullName: "org/a", StargazersCount: 180}}}
	acc := deriveAcceleration(store, cur, now, "baseline")
	got := acc["org/a"]
	if got != 10 {
		t.Fatalf("expected acceleration 10.0 stars/hour delta, got %.2f", got)
	}
}
