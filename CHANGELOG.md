# Changelog

## v0.9.0 - 2026-03-09

### Added
- Advanced momentum model variants via `--momentum-model baseline|decay|trend`.
- Team routing policy packs via config (`routing_profiles`) and `--routing-profile` selector.
- Plugin SDK hook for external processors via `--plugin-cmd` (JSON delta payload on stdin).

### Improved
- Major TUI visual overhaul using Bubble Tea + Lip Gloss:
  - multi-panel layout,
  - styled headers and borders,
  - momentum signals panel,
  - keyboard navigation hints.
- stdout mode remains default; TUI is opt-in (`--ui tui`).

## v0.8.0 - 2026-03-09

### Added
- Optional Bubble Tea TUI mode (`--ui tui`) while keeping classic stdout output as the default.
- Snapshot workflow flags:
  - `--snapshot-export`
  - `--snapshot-import`
  - `--snapshot-diff pathA:pathB`
- Preset scoring profiles and acceleration alert thresholds via config (`preset_profiles`).
- Webhook auth/signing options:
  - `watch_auth_token` / `--watch-auth-token`
  - `watch_sign_secret` / `--watch-sign-secret`

### Improved
- Watch machine-readable delta events now include explicit schema versioning (`schema_version: v1`).
- Documentation refreshed for v0.8.0 workflows and configuration paths.

## v0.7.0 - 2026-03-09

### Added
- Snapshot diff subcommand workflow for offline comparison operations.
- Richer momentum analytics layer (velocity trend + decay-aware framing).
- Schema-versioned machine-readable watch summary output.
- Webhook delivery auth/signing support for downstream integrations.

### Improved
- Roadmap and release plan rolled forward to v0.7.0 completion.
- Release metadata and docs synchronized around v0.7.0.

## v0.6.0 - 2026-03-09

### Added
- Preset override support via config (`preset_overrides`).
- Watch webhook sink (`--watch-webhook` / `watch_webhook`) for delta POST delivery.
- Acceleration ranking mode (`--sort accel`) based on snapshot momentum.
- Modular internal package structure (thin `main`, app/config/query/githubapi/rank/output/snapshot split).
- Table-driven unit tests for query overrides, webhook delivery, and config validation.

### Improved
- Roadmap and release plan updated to reflect completed v0.6 scope.
- Watch automation docs expanded (JSONL + webhook sinks).
