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
	queries, err := query.Resolve(cfg.Queries, cfg.Presets, cfg.SinceDays, time.Now().UTC())
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
		}
		_ = snapshot.Append(cfg.SnapshotPath, scored, now)
		prev = snapshot.FromRanked(scored)
		time.Sleep(cfg.Interval)
	}
}
