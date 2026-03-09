package app

import (
	"fmt"
	"io"
	"time"

	"github.com/agent19710101/gh-hype-scout/internal/config"
	"github.com/agent19710101/gh-hype-scout/internal/githubapi"
	"github.com/agent19710101/gh-hype-scout/internal/output"
	"github.com/agent19710101/gh-hype-scout/internal/query"
	"github.com/agent19710101/gh-hype-scout/internal/rank"
	"github.com/agent19710101/gh-hype-scout/internal/snapshot"
)

type Runner struct {
	Out io.Writer
	Err io.Writer
}

func (r Runner) Run(cfg config.Run) error {
	if err := rank.Validate(rank.Config{Sort: cfg.Sort, ScorePreset: cfg.ScorePreset}); err != nil {
		return err
	}
	queries, err := query.ResolveWithOverrides(cfg.Queries, cfg.Presets, cfg.PresetOverrides, cfg.SinceDays, time.Now().UTC())
	if err != nil {
		return err
	}
	client := githubapi.New()
	if cfg.Watch {
		return r.runWatch(cfg, queries, client)
	}
	scored, err := runOnce(cfg, queries, client)
	if err != nil {
		return err
	}
	if err := render(r.Out, cfg, scored); err != nil {
		return err
	}
	_ = snapshot.Append(cfg.SnapshotPath, scored, time.Now().UTC())
	return nil
}

func runOnce(cfg config.Run, queries []string, client *githubapi.Client) ([]rank.Repo, error) {
	all, err := githubapi.MergeSearch(queries, client.Search)
	if err != nil {
		return nil, fmt.Errorf("fetch trends: %w", err)
	}
	if len(all) == 0 {
		return nil, fmt.Errorf("no repositories returned")
	}
	scored := rank.Score(all, rank.Config{Sort: cfg.Sort, ScorePreset: cfg.ScorePreset})
	store, _ := snapshot.Load(cfg.SnapshotPath)
	rank.ApplyAcceleration(scored, deriveAcceleration(store, scored, time.Now().UTC()))
	return rank.Filter(scored, cfg.MinStars, cfg.MinAgeDays, cfg.MaxAgeDays, cfg.Limit)
}

func render(w io.Writer, cfg config.Run, scored []rank.Repo) error {
	if cfg.JSON {
		return output.PrintJSON(w, scored)
	}
	output.PrintTable(w, scored, cfg.DescWidth)
	if cfg.Themes {
		fmt.Fprintln(w)
		output.PrintThemeSummary(w, scored)
	}
	return nil
}

func deriveAcceleration(store snapshot.Store, current []rank.Repo, now time.Time) map[string]float64 {
	if len(store.Runs) == 0 {
		return map[string]float64{}
	}
	latest := store.Runs[len(store.Runs)-1]
	latestByName := map[string]snapshot.Item{}
	for _, it := range latest.Items {
		latestByName[it.FullName] = it
	}

	var prevByName map[string]snapshot.Item
	if len(store.Runs) > 1 {
		prev := store.Runs[len(store.Runs)-2]
		prevByName = map[string]snapshot.Item{}
		for _, it := range prev.Items {
			prevByName[it.FullName] = it
		}
	}

	hoursSinceLatest := now.Sub(latest.CapturedAt).Hours()
	if hoursSinceLatest <= 0 {
		hoursSinceLatest = 1
	}

	accel := make(map[string]float64, len(current))
	for _, r := range current {
		l, ok := latestByName[r.FullName]
		if !ok {
			continue
		}
		recentRate := float64(r.StargazersCount-l.Stars) / hoursSinceLatest
		baselineRate := 0.0
		if prevByName != nil {
			if p, ok := prevByName[r.FullName]; ok {
				window := latest.CapturedAt.Sub(store.Runs[len(store.Runs)-2].CapturedAt).Hours()
				if window > 0 {
					baselineRate = float64(l.Stars-p.Stars) / window
				}
			}
		}
		accel[r.FullName] = recentRate - baselineRate
	}
	return accel
}

func (r Runner) runWatch(cfg config.Run, queries []string, client *githubapi.Client) error {
	store, _ := snapshot.Load(cfg.SnapshotPath)
	var prev []snapshot.Item
	if len(store.Runs) > 0 {
		prev = store.Runs[len(store.Runs)-1].Items
	}
	for {
		now := time.Now().UTC()
		scored, err := runOnce(cfg, queries, client)
		if err != nil {
			return err
		}
		if err := render(r.Out, cfg, scored); err != nil {
			return err
		}
		report := output.BuildDelta(prev, scored)
		if len(prev) > 0 {
			output.PrintDelta(r.Out, report)
			_ = output.AppendDeltaJSONL(cfg.WatchJSONL, now, report)
			reportCopy := report
			webhook := cfg.WatchWebhook
			go func(ts time.Time) {
				_ = output.SendDeltaWebhook(webhook, ts, reportCopy)
			}(now)
		}
		_ = snapshot.Append(cfg.SnapshotPath, scored, now)
		prev = snapshot.FromRanked(scored)
		time.Sleep(cfg.Interval)
	}
}
