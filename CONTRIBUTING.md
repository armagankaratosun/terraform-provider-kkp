# Contributing

Thanks for your interest in contributing to terraform-provider-kkp!

This document describes how to set up your environment, the branching model, coding standards, and how releases are cut.

## Development Setup

- Go: use the version declared in `go.mod`.
- Tools: `make dev-deps` installs linters.
- Build: `make build` produces `./bin/terraform-provider-kkp_v<version>`.
- Format/Lint/Test:
  - `make fmt`
  - `make lint`
  - `make test`

## Changelog

We keep a human-readable Keep a Changelog file (`CHANGELOG.md`). Entries are grouped under `[Unreleased]` until a release.

- For each PR, add concise notes under `[Unreleased]` using sections: Added, Changed, Deprecated, Removed, Fixed, Security.
- CI validates the changelog format and fails PRs with an empty `[Unreleased]`.
- You can draft/update the changelog using git-cliff:
  - Locally (requires `git-cliff`): `make changelog` to update `[Unreleased]`, or `make changelog-release VERSION=X.Y.Z` to prepare a release section.
  - GitHub Actions: run the manual workflow `changelog` and optionally provide `version` (blank = Unreleased).

## Branching Model

- `main`: active development; keep green. All changes should land via pull requests.
- `release/v*` (e.g., `release/v0.0.1`): maintenance branch for a specific release line; cherry‑pick fixes as needed.
- Tags (`vX.Y.Z`): immutable, created from `main` or a `release/v*` branch to cut a release.

Suggested flow:
1. Branch from `main` using a descriptive name (e.g., `feat/x`, `fix/y`).
2. Open a PR; ensure CI, lint, and tests pass.
3. Squash‑merge or rebase‑merge after review.

## Coding Standards

- Keep changes small and focused; write clear commit messages.
- Run `make fmt tidy lint test` before pushing.
- Follow existing code style and module layout; avoid large refactors in feature PRs.
- Add/update docs under `docs/` when you add resources or data sources.

## Testing

- Unit tests: `go test ./...`.
- Acceptance/e2e tests: not yet automated; prefer adding small, deterministic unit tests.

## Release Process

Releases are built from tags using GitHub Actions and GoReleaser.

1. Ensure `CHANGELOG.md` has entries under `[Unreleased]`; when preparing a release, add or generate a new `## [X.Y.Z] - YYYY-MM-DD` section per Keep a Changelog (use git-cliff locally or the `changelog` workflow).
2. Ensure `main` (or your target `release/v*` branch) is green.
3. Create and push a semver tag:
   ```bash
   git tag vX.Y.Z
   git push origin vX.Y.Z
   ```
4. GitHub Actions will build artifacts for multiple OS/ARCH and publish a GitHub Release.

Notes:
- Publishing to the Terraform/OpenTofu registries is not wired up yet; current automation publishes GitHub releases only.
- The provider binary embeds the version via `-ldflags -X main.version=<tag>`.
 - The release workflow validates that the tag version exists in `CHANGELOG.md`.

## Branch Protection

The `main` branch is protected. Changes land via PRs with passing CI checks.

Notes:
- The CI job `ci / lint-test` must pass, which includes changelog validation on PRs.
- Release tags are validated against `CHANGELOG.md`. Before tagging, ensure a matching `## [X.Y.Z] - YYYY-MM-DD` section exists, or the release workflow will fail.

## Issue Triage

- Use labels `bug`, `enhancement`, `docs`, `good first issue`.
- Link PRs to issues when applicable.
- Prefer minimal repros and logs in bug reports.

## Code of Conduct

Be respectful and constructive. The maintainers may guide scope to keep the project stable and focused.
