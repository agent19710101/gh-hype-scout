# Release Plan (v0.x)

Status: ✅ Completed through **v0.4.0**

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

## Release quality gate (applies to each v0.x)

- [x] `go test ./...` green.
- [x] README updated for all user-facing flags.
- [x] No tracked build artifacts in Git.
