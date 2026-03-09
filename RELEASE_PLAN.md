# Release Plan (v0.x)

Status: ✅ Completed through **v0.0.6**

## v0.2.0 — Output Reliability + UX

Goal: make terminal output stable and easier to scan.

- [x] Add description truncation column in table output.
- [x] Add golden tests for table output stability.
- [x] Add tests for query merge/dedupe behavior.

## v0.3.0 — Signal Controls

Goal: improve ranking control without adding heavy complexity.

- [x] Add `--sort age` mode.
- [x] Add optional weighted scoring toggle (hot vs. freshness-biased preset).
- [x] Tighten flag validation and error messages for scoring/sorting combinations.

## v0.4.0 — Operability + Distribution

Goal: make the CLI easier to adopt and run continuously.

- [x] Add config file support (`~/.config/gh-hype-scout/config.yaml`).
- [x] Add GitHub Actions release pipeline (tag build + checks).
- [x] Add goreleaser for multi-platform binaries.

## v0.5.0 — Continuous Trend Tracking

Goal: move from one-shot scouting to iterative tracking.

- [x] Add query presets (`--preset`) and config support.
- [x] Add snapshot persistence and retention policy.
- [x] Add watch mode (`--watch`, `--interval`) with delta output.

## v0.6.0 — Open Ideas to Productized Features

Goal: turn open ideas into automation-ready workflows.

- [x] Add watch JSONL sink (`--watch-jsonl`).
- [x] Add acceleration scoring based on snapshot history (`--sort accel`).
- [x] Add preset customization/overrides from config (`preset_overrides`).
- [x] Add optional webhook sink for watch delta events (`--watch-webhook`).
- [x] Refactor to modular internal package structure (thin main).

## v0.7.0 — Analysis + Automation Hardening

Goal: strengthen trend analysis depth and downstream integrations.

- [ ] Add snapshot diff subcommand for offline comparisons.
- [ ] Add richer momentum metrics and decay-aware ranking.
- [ ] Add schema-versioned machine-readable watch summaries.
- [ ] Add webhook delivery auth/signing options.

## Release quality gate (applies to each v0.x)

- [x] `go test ./...` green.
- [x] README updated for all user-facing flags.
- [x] No tracked build artifacts in Git.
