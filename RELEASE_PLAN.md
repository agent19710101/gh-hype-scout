# Release Plan (v0.x)

## v0.2.0 — Output Reliability + UX

Goal: make terminal output stable and easier to scan.

- Add description truncation column in table output.
- Add golden tests for table output stability.
- Add tests for query merge/dedupe behavior.

## v0.3.0 — Signal Controls

Goal: improve ranking control without adding heavy complexity.

- Add `--sort age` mode.
- Add optional weighted scoring toggle (hot vs. freshness-biased preset).
- Tighten flag validation and error messages for scoring/sorting combinations.

## v0.4.0 — Operability + Distribution

Goal: make the CLI easier to adopt and run continuously.

- Add config file support (`~/.config/gh-hype-scout/config.yaml`).
- Add GitHub Actions release pipeline (tag build + checks).
- Add goreleaser for multi-platform binaries.

## Release quality gate (applies to each v0.x)

- `go test ./...` green.
- README updated for all user-facing flags.
- No tracked build artifacts in Git.
