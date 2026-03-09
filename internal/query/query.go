package query

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

func Resolve(custom, presets []string, sinceDays int, now time.Time) ([]string, error) {
	return ResolveWithOverrides(custom, presets, nil, sinceDays, now)
}

func ResolveWithOverrides(custom, presets []string, overrides map[string][]string, sinceDays int, now time.Time) ([]string, error) {
	p, err := ExpandPresetsWithOverrides(presets, overrides, sinceDays, now)
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
	return ExpandPresetsWithOverrides(presets, nil, sinceDays, now)
}

func ExpandPresetsWithOverrides(presets []string, overrides map[string][]string, sinceDays int, now time.Time) ([]string, error) {
	if len(presets) == 0 {
		return nil, nil
	}
	catalog, err := CatalogWithOverrides(sinceDays, now, overrides)
	if err != nil {
		return nil, err
	}
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
	c, _ := CatalogWithOverrides(sinceDays, now, nil)
	return c
}

func CatalogWithOverrides(sinceDays int, now time.Time, overrides map[string][]string) (map[string][]string, error) {
	if sinceDays < 1 {
		sinceDays = 1
	}
	since := now.AddDate(0, 0, -sinceDays).Format("2006-01-02")
	catalog := map[string][]string{
		"oss":      {"stars:>50 created:>" + since},
		"agents":   {"(agent OR mcp) created:>" + since + " stars:>80"},
		"cli":      {"topic:cli created:>" + since + " stars:>40"},
		"tui":      {"topic:tui created:>" + since + " stars:>20"},
		"devtools": {"(developer tools) created:>" + since + " stars:>50"},
	}
	for k, v := range overrides {
		name := strings.ToLower(strings.TrimSpace(k))
		if name == "" {
			return nil, fmt.Errorf("preset override key cannot be empty")
		}
		if len(v) == 0 {
			return nil, fmt.Errorf("preset override %q must contain at least one query", k)
		}
		clean := make([]string, 0, len(v))
		for _, q := range v {
			q = strings.TrimSpace(q)
			if q == "" {
				return nil, fmt.Errorf("preset override %q contains empty query", k)
			}
			clean = append(clean, q)
		}
		catalog[name] = clean
	}
	return catalog, nil
}

func Names(c map[string][]string) []string {
	keys := make([]string, 0, len(c))
	for k := range c {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
