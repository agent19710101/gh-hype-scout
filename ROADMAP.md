# Roadmap

## Shipped (through v1.0.0)

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
- ✅ Advanced momentum model variants (`--momentum-model baseline|decay|trend`)
- ✅ Team routing policy packs (`routing_profiles`, `--routing-profile`)
- ✅ Plugin SDK hook for external processors (`--plugin-cmd`)
- ✅ TUI long-repository-name handling with adaptive truncation and responsive layout

## Next (v1.1.0)

- [ ] Interactive sortable/filterable TUI controls with persisted layouts
- [ ] Webhook key rotation and delivery audit export
- [ ] Snapshot compaction and profile-based retention policies
- [ ] Typed plugin contracts and plugin registry metadata

## Open ideas

- [ ] Historical replay mode for strategy backtesting
- [ ] Built-in anomaly detection for sudden momentum shifts
- [ ] Team workspace templates for multi-repo scouting
