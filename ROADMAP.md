# Roadmap

## Shipped (through v0.5.0)

- ✅ Age filters: `--min-age-days` / `--max-age-days`
- ✅ Rate-limit aware API hints (403/429)
- ✅ Table description truncation (`DESC`) + output golden tests
- ✅ Query merge/dedupe reliability tests
- ✅ Sorting modes: `hot`, `stars`, `stars-day`, `age`
- ✅ Config file support (`~/.config/gh-hype-scout/config.yaml`)
- ✅ Release automation (GitHub Actions + goreleaser)
- ✅ Query presets (`--preset oss|agents|cli|tui|devtools`)
- ✅ Snapshot persistence (`~/.cache/gh-hype-scout/snapshots.json`) with retention
- ✅ Watch mode with periodic delta output (`--watch`, `--interval`)

## Next (v0.6.0)

- [x] Add watchlist output sink: JSONL (`--watch-jsonl`)
- [ ] Add acceleration scoring from snapshot history (week-over-week trend signal)
- [ ] Add preset customization via config file overrides
- [ ] Optional webhook sink for watch delta events

## Open ideas (post-v0.6.0)

- [ ] Add snapshot diff subcommand for offline comparisons
- [ ] Add richer momentum metrics (star velocity trend, decay curves)
- [ ] Add machine-readable watch summaries with schema versioning
