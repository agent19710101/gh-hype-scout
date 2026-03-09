# Release Plan (v0.x/v1.x)

Status: ✅ Completed through **v1.0.0**

## v0.2.0 — Output Reliability + UX
- [x] Description truncation + golden output tests.
- [x] Query merge/dedupe reliability tests.

## v0.3.0 — Signal Controls
- [x] `--sort age` and scoring validation hardening.

## v0.4.0 — Operability + Distribution
- [x] Config support, release workflow, goreleaser.

## v0.5.0 — Continuous Trend Tracking
- [x] Presets, snapshots, watch mode.

## v0.6.0 — Open Ideas to Productized Features
- [x] JSONL sink, acceleration sort, preset overrides, webhook sink, modular refactor.

## v0.7.0 — Analysis + Automation Hardening
- [x] Snapshot diff, richer momentum framing, schema summaries, webhook auth/signing.

## v0.8.0 — Operator Experience + Workflow Completion
- [x] Optional TUI mode with stdout default preserved.
- [x] Snapshot export/import workflows.
- [x] Preset scoring profiles + accel thresholds.
- [x] Output module selection path with backward compatibility.

## v0.9.0 — Team Ops + Extensibility
- [x] Advanced momentum model variants (`baseline`, `decay`, `trend`).
- [x] Team routing policy packs (`routing_profiles`, `--routing-profile`).
- [x] Plugin SDK hook (`--plugin-cmd` external processor contract).
- [x] Visual TUI overhaul (multi-panel styled dashboard).

## v1.0.0 — Stability + Platform Layer
- [x] Persistent interactive TUI behavior baseline with responsive panel layout.
- [x] Signed webhook delivery support with token + HMAC headers.
- [x] Snapshot diff/export/import workflows for operational lifecycle.
- [x] Plugin processing contract and execution hook.
- [x] Long repository names handled safely in TUI without layout breaks.

## v1.1.0 — Hardening + Ecosystem
- [ ] Interactive persisted TUI controls and saved layouts.
- [ ] Key rotation + delivery audit exports.
- [ ] Snapshot compaction and policy engine.
- [ ] Plugin registry and typed interface contracts.

## Release quality gate
- [x] `go test ./...` green.
- [x] README updated for user-facing flags.
- [x] No tracked build artifacts in Git.
