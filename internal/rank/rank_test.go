package rank

import (
	"testing"

	"github.com/agent19710101/gh-hype-scout/internal/githubapi"
)

func TestValidateSupportsAccel(t *testing.T) {
	if err := Validate(Config{Sort: "accel", ScorePreset: "hot"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSortByAcceleration(t *testing.T) {
	in := []Repo{
		{Repo: githubapi.Repo{FullName: "org/a", StargazersCount: 100}, HotScore: 10, Acceleration: 1},
		{Repo: githubapi.Repo{FullName: "org/b", StargazersCount: 80}, HotScore: 9, Acceleration: 5},
	}
	sortScored(in, "accel")
	if in[0].FullName != "org/b" {
		t.Fatalf("expected org/b first by accel, got %s", in[0].FullName)
	}
}
