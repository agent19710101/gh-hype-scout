package query

import (
	"strings"
	"testing"
	"time"
)

func TestCatalogWithOverrides_Table(t *testing.T) {
	now := time.Date(2026, 3, 9, 0, 0, 0, 0, time.UTC)
	tests := []struct {
		name      string
		overrides map[string][]string
		wantErr   string
		check     func(t *testing.T, c map[string][]string)
	}{
		{
			name:      "override existing preset",
			overrides: map[string][]string{"cli": {"language:go stars:>100"}},
			check: func(t *testing.T, c map[string][]string) {
				if got := c["cli"][0]; got != "language:go stars:>100" {
					t.Fatalf("unexpected cli override: %q", got)
				}
			},
		},
		{
			name:      "add custom preset",
			overrides: map[string][]string{"security": {"topic:security stars:>40"}},
			check: func(t *testing.T, c map[string][]string) {
				if _, ok := c["security"]; !ok {
					t.Fatal("security preset missing")
				}
				if _, ok := c["cli"]; !ok {
					t.Fatal("built-in preset missing")
				}
			},
		},
		{name: "empty key", overrides: map[string][]string{"": {"x"}}, wantErr: "key cannot be empty"},
		{name: "empty list", overrides: map[string][]string{"cli": {}}, wantErr: "at least one query"},
		{name: "empty query", overrides: map[string][]string{"cli": {" "}}, wantErr: "contains empty query"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := CatalogWithOverrides(30, now, tt.overrides)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("expected error containing %q, got %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.check != nil {
				tt.check(t, c)
			}
		})
	}
}

func TestResolveWithOverrides_Table(t *testing.T) {
	now := time.Date(2026, 3, 9, 0, 0, 0, 0, time.UTC)
	tests := []struct {
		name    string
		custom  []string
		presets []string
		over    map[string][]string
		wantN   int
	}{
		{name: "default fallback", wantN: 4},
		{name: "preset + custom", presets: []string{"cli"}, custom: []string{"language:go stars:>50"}, wantN: 2},
		{name: "custom preset override", presets: []string{"my"}, over: map[string][]string{"my": {"topic:go stars:>10"}}, wantN: 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			qs, err := ResolveWithOverrides(tt.custom, tt.presets, tt.over, 30, now)
			if err != nil {
				t.Fatalf("resolve error: %v", err)
			}
			if len(qs) != tt.wantN {
				t.Fatalf("expected %d queries, got %d (%v)", tt.wantN, len(qs), qs)
			}
		})
	}
}
