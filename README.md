# gh-hype-scout

Terminal-first scout for discovering fast-rising GitHub repositories before they become obvious.

## Problem

GitHub Trending is useful but noisy and less configurable for focused discovery. If you care about early OSS signals (new CLIs, agent tooling, dev infra), you need a repeatable way to query, rank, and filter emerging repos from the terminal.

`gh-hype-scout` solves this by combining multiple GitHub Search queries, deduplicating results, and ranking repos by an explainable hotness model.

## Status

- Current version: v0 (actively iterating)
- Works today for terminal scanning + JSON export
- Priorities: better signal quality, richer filters, stronger tests

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

# adjust default time window
gh-hype-scout --since-days 30

# custom query (repeatable)
gh-hype-scout \
  -q 'topic:cli created:>2026-01-01 stars:>30' \
  -q 'topic:tui created:>2026-01-01 stars:>20'

# limit output rows
gh-hype-scout -n 25

# filter out smaller repos
gh-hype-scout --min-stars 500

# include category/theme summary
gh-hype-scout --themes

# JSON output for automation
gh-hype-scout --json
```

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

See [ROADMAP.md](./ROADMAP.md).

## License

MIT — see [LICENSE](./LICENSE).
