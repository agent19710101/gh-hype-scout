# Roadmap

## Shipped (through v0.0.6)

- ✅ Age filters: `--min-age-days` / `--max-age-days`
- ✅ Rate-limit aware API hints (403/429)
- ✅ Table description truncation (`DESC`) + output golden tests
- ✅ Query merge/dedupe reliability tests
- ✅ Sorting modes: `hot`, `stars`, `stars-day`, `age`, `accel`
- ✅ Config file support (`~/.config/gh-hype-scout/config.yaml`)
- ✅ Release automation (GitHub Actions + goreleaser)
- ✅ Query presets (`--preset oss|agents|cli|tui|devtools`)
- ✅ Preset overrides via config (`preset_overrides`)
- ✅ Snapshot persistence (`~/.cache/gh-hype-scout/snapshots.json`) with retention
- ✅ Watch mode with periodic delta output (`--watch`, `--interval`)
- ✅ Watch JSONL sink (`--watch-jsonl`)
- ✅ Watch webhook sink (`--watch-webhook`) with retry/backoff
- ✅ Modular architecture (thin `main`, internal packages)

## Next (v0.7.0)

- [ ] Add snapshot diff subcommand for offline comparisons
- [ ] Add richer momentum metrics (velocity trend and decay)
- [ ] Add machine-readable watch summaries with schema versioning
- [ ] Add delivery signing/auth options for webhook sink

## Open ideas

- [ ] Terminal TUI mode for live watch sessions
- [ ] Export/import snapshot archives
- [ ] Per-preset scoring profiles and alert thresholds
