# Changelog

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
