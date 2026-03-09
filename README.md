# gh-hype-scout

Find fast-rising GitHub repositories before they become obvious.

`gh-hype-scout` queries GitHub Search, merges results across themes, and ranks repos by a simple velocity score (stars/day + recency weighting).

## Why

Trending pages are great, but noisy. This tool gives developers and OSS maintainers a practical terminal-first view of **emerging repos** with explainable scoring.

## Install

```bash
go install github.com/agent19710101/gh-hype-scout@latest
```

Or build locally:

```bash
go build -o gh-hype-scout ./cmd/gh-hype-scout
```

## Usage

```bash
# default multi-theme scan
gh-hype-scout

# custom query (repeatable)
gh-hype-scout -q 'topic:cli created:>2026-01-01 stars:>30' -q 'topic:tui created:>2026-01-01 stars:>20'

# JSON output
gh-hype-scout --json
```

## Authentication

Unauthenticated mode works (lower rate limits). For higher limits:

- `GITHUB_TOKEN=... gh-hype-scout`

## Scoring model (v0)

- `ageDays = max(1, now - createdAt)`
- `starsPerDay = stars / ageDays`
- `hotScore = starsPerDay * log10(stars+1)`

This is intentionally simple and explainable; future versions can add contributor growth and release cadence.

## License

MIT
