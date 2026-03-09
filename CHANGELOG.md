# Changelog

## v1.0.1 - 2026-03-09

### Fixed
- Added root `main` package entrypoint so this now works as expected:
  - `go install github.com/agent19710101/gh-hype-scout@latest`
- Preserved existing command layout (`cmd/gh-hype-scout`) for source builds.

## v1.0.0 - 2026-03-09

### Added
- Webhook auth/signing runtime flags:
  - `--watch-auth-token`
  - `--watch-sign-secret`
- Momentum model selector: `--momentum-model baseline|decay|trend`.
- Team routing profile selector: `--routing-profile` from config policy packs.
- Plugin processor hook: `--plugin-cmd` (JSON delta payload via stdin).

### Improved
- TUI polish pass with responsive multi-panel layout and visual styling.
- TUI now safely handles very long repository names with adaptive width truncation.
- stdout behavior remains unchanged default; TUI stays explicit opt-in (`--ui tui`).

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
