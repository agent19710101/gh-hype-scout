package query

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

func Resolve(custom, presets []string, sinceDays int, now time.Time) ([]string, error) {
	p, err := ExpandPresets(presets, sinceDays, now)
	if err != nil {
		return nil, err
	}
	if len(custom) == 0 && len(p) == 0 {
		return Default(sinceDays, now), nil
	}
	out := append([]string{}, p...)
	out = append(out, custom...)
	return out, nil
}

func ExpandPresets(presets []string, sinceDays int, now time.Time) ([]string, error) {
	if len(presets) == 0 {
		return nil, nil
	}
	catalog := Catalog(sinceDays, now)
	out := make([]string, 0)
	seen := map[string]struct{}{}
	for _, p := range presets {
		name := strings.ToLower(strings.TrimSpace(p))
		queries, ok := catalog[name]
		if !ok {
			return nil, fmt.Errorf("unknown -preset %q (available: %s)", p, strings.Join(Names(catalog), ", "))
		}
		for _, q := range queries {
			if _, ok := seen[q]; ok {
				continue
			}
			seen[q] = struct{}{}
			out = append(out, q)
		}
	}
	return out, nil
}

func Default(sinceDays int, now time.Time) []string {
	if sinceDays < 1 {
		sinceDays = 1
	}
	since := now.AddDate(0, 0, -sinceDays).Format("2006-01-02")
	return []string{
		"topic:cli created:>" + since + " stars:>40",
		"topic:tui created:>" + since + " stars:>20",
		"(agent OR mcp) created:>" + since + " stars:>80",
		"(developer tools) created:>" + since + " stars:>50",
	}
}

func Catalog(sinceDays int, now time.Time) map[string][]string {
	if sinceDays < 1 {
		sinceDays = 1
	}
	since := now.AddDate(0, 0, -sinceDays).Format("2006-01-02")
	return map[string][]string{
		"oss":      {"stars:>50 created:>" + since},
		"agents":   {"(agent OR mcp) created:>" + since + " stars:>80"},
		"cli":      {"topic:cli created:>" + since + " stars:>40"},
		"tui":      {"topic:tui created:>" + since + " stars:>20"},
		"devtools": {"(developer tools) created:>" + since + " stars:>50"},
	}
}

func Names(c map[string][]string) []string {
	keys := make([]string, 0, len(c))
	for k := range c {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
