# Roadmap

## Near-term (v0.x)

- [x] Add `--min-age-days` / `--max-age-days` filters to tune freshness.
- [ ] Add optional description column truncation for richer table UX.
- [x] Add rate-limit aware messaging when GitHub API returns 403/429.
- [ ] Add tests for query merging behavior (`fetchAndMerge` dedupe semantics).
- [ ] Add golden test for table output format stability.

## Mid-term (v1)

- [ ] Add saved query presets (`--preset oss`, `--preset agents`, etc.).
- [ ] Add optional output sorting modes (`score`, `stars`, `age`, `stars/day`).
- [ ] Add config file support (`~/.config/gh-hype-scout/config.yaml`).
- [ ] Add release automation (GitHub Actions + goreleaser).

## Open ideas

- [ ] Integrate lightweight historical snapshots to detect week-over-week acceleration.
- [ ] Add "watchlist" mode for periodic scans and delta output.
