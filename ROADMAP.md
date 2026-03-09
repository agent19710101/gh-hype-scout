# Roadmap

## Shipped (through v0.8.0)

- ✅ Age filters: `--min-age-days` / `--max-age-days`
- ✅ Rate-limit aware API hints (403/429)
- ✅ Table description truncation (`DESC`) + output golden tests
- ✅ Query merge/dedupe reliability tests
- ✅ Sorting modes: `hot`, `stars`, `stars-day`, `age`, `accel`
- ✅ Config file support (`~/.config/gh-hype-scout/config.yaml`)
- ✅ Release automation (GitHub Actions + goreleaser)
- ✅ Query presets (`--preset oss|agents|cli|tui|devtools`)
- ✅ Preset overrides via config (`preset_overrides`)
- ✅ Preset scoring profiles + accel thresholds (`preset_profiles`)
- ✅ Snapshot persistence (`~/.cache/gh-hype-scout/snapshots.json`) with retention
- ✅ Watch mode with periodic delta output (`--watch`, `--interval`)
- ✅ Watch JSONL sink (`--watch-jsonl`) with schema versioning
- ✅ Watch webhook sink (`--watch-webhook`) with retry/backoff and signing/auth
- ✅ Modular architecture (thin `main`, internal packages)
- ✅ Optional Bubble Tea TUI mode (`--ui tui`) while preserving stdout default
- ✅ Snapshot export/import/diff workflows

## Next (v0.9.0)

- [ ] Multi-panel live TUI with interactive filters
- [ ] Snapshot compression + archive rotation policies
- [ ] Advanced momentum models (time-weighted regressions)
- [ ] Team-level notification routing profiles

## Open ideas

- [ ] Plugin SDK for third-party output/analysis modules
- [ ] Historical replay mode for trend strategy backtesting
- [ ] Built-in anomaly detector for sudden repo spikes
