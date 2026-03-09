# Roadmap

## Shipped (through v0.4.0)

- ✅ Age filters: `--min-age-days` / `--max-age-days`
- ✅ Rate-limit aware API hints (403/429)
- ✅ Table description truncation (`DESC`) + output golden tests
- ✅ Query merge/dedupe reliability tests
- ✅ Sorting modes: `hot`, `stars`, `stars-day`, `age`
- ✅ Config file support (`~/.config/gh-hype-scout/config.yaml`)
- ✅ Release automation (GitHub Actions + goreleaser)

## Next (v0.5.0)

- [ ] Add saved query presets (`--preset oss`, `--preset agents`, etc.)
- [ ] Persist lightweight snapshots for historical trend comparisons
- [ ] Add watchlist mode for periodic scans and delta output

## Open ideas (post-v0.5.0)

- [ ] Add acceleration scoring from snapshot history (week-over-week trend signal)
- [ ] Add preset customization via config file overrides
- [ ] Add watchlist output sinks (JSONL file and optional webhook)
