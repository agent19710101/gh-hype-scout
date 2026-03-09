# Changelog

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
