# gh-hype-scout

Terminal-first scout for discovering fast-rising GitHub repositories before they become obvious.

## Problem

GitHub Trending is useful but noisy and less configurable for focused discovery. If you care about early OSS signals (new CLIs, agent tooling, dev infra), you need a repeatable way to query, rank, and filter emerging repos from the terminal.

`gh-hype-scout` solves this by combining multiple GitHub Search queries, deduplicating results, and ranking repos by an explainable hotness model.

## Status

- Current version: v0 (actively iterating)
- Works today for terminal scanning + JSON export
- Priorities: rate-limit UX, richer ranking signals, stronger output ergonomics

## Install

### Option A: go install

```bash
go install github.com/agent19710101/gh-hype-scout@latest
```

### Option B: build from source

```bash
git clone https://github.com/agent19710101/gh-hype-scout.git
cd gh-hype-scout
go build -o gh-hype-scout ./cmd/gh-hype-scout
```

## Usage

```bash
# default multi-theme scan (last 60 days by default)
gh-hype-scout

# use config defaults (auto-loads ~/.config/gh-hype-scout/config.yaml)
gh-hype-scout

# adjust default time window
gh-hype-scout --since-days 30

# built-in presets (repeatable): oss, agents, cli, tui, devtools
gh-hype-scout --preset cli --preset agents

# custom query (repeatable)
gh-hype-scout \
  -q 'topic:cli created:>2026-01-01 stars:>30' \
  -q 'topic:tui created:>2026-01-01 stars:>20'

# combine presets + custom queries
gh-hype-scout --preset oss -q 'language:go stars:>80 created:>2026-01-01'

# limit output rows
gh-hype-scout -n 25

# filter out smaller repos
gh-hype-scout --min-stars 500

# focus on mature-but-still-fresh repos
gh-hype-scout --min-age-days 7 --max-age-days 45

# include category/theme summary
gh-hype-scout --themes

# alternative ranking views
gh-hype-scout --sort stars-day
gh-hype-scout --sort stars
gh-hype-scout --sort age

# freshness-biased scoring (only valid with --sort hot)
gh-hype-scout --sort hot --score-preset fresh

# tune table description width
gh-hype-scout --desc-width 80

# watch mode with periodic delta output (every 2 minutes)
gh-hype-scout --watch --interval 120

# JSON output for automation
gh-hype-scout --json
```

## Configuration file

`gh-hype-scout` auto-loads config from:

- `~/.config/gh-hype-scout/config.yaml`

Override with `--config /path/to/config.yaml`.

Snapshot history is stored by default at:

- `~/.cache/gh-hype-scout/snapshots.json`

Override with `--snapshot-path /path/to/snapshots.json`.

Use [`config.example.yaml`](./config.example.yaml) as a starting point.

## Authentication

Unauthenticated mode works, but rate limits are lower.

For better limits, provide a token:

```bash
GITHUB_TOKEN=... gh-hype-scout
```

## Scoring model (v0)

For each repo:

- `ageDays = max(1, now - createdAt)`
- `starsPerDay = stars / ageDays`
- `hotScore = starsPerDay * log10(stars+1)`

This is intentionally simple and explainable; future versions can include contributor growth, commit activity, and release cadence.

## Roadmap

See [ROADMAP.md](./ROADMAP.md) and the scoped [RELEASE_PLAN.md](./RELEASE_PLAN.md) for v0.x milestones.

## Releases

- CI runs `go test ./...` on PRs and pushes.
- Tagging `v*` triggers goreleaser via GitHub Actions.
- Artifacts include Linux/macOS/Windows binaries and checksums.

## License

MIT — see [LICENSE](./LICENSE).
