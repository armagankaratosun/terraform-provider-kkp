# Contributing

Thanks for your interest in contributing to terraform-provider-kkp!

This document describes how to set up your environment, the branching model, coding standards, and how releases are cut.

## Development Setup

- Go: use the version declared in `go.mod`.
- Tools: `make dev-deps` installs linters/tools (golangci-lint, goimports). Add `$GOPATH/bin` to your `PATH`.
- Build: `make build` produces `./bin/terraform-provider-kkp_v<version>`.
  - If `VERSION` is not set, builds use `dev` as the version for the binary name and embedded version info.
- Format/Lint/Test:
  - `make fmt` formats Go code and runs `terraform fmt` to write changes in `examples/` if present.
  - `make fmt-check` checks formatting (Go via gofmt -s, Terraform via `terraform fmt -check` in `examples/`). Install Terraform locally to run this; CI enforces it.
  - `make lint`
  - `make test`

### Documentation Generation (tfplugindocs)

This provider uses HashiCorp's terraform-plugin-docs (tfplugindocs) to generate provider, resource, and data source documentation under `docs/` as required by Terraform Registry publishing guidelines.

- Install: `go install github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs@latest`
- Generate docs: `make docs`
- Verify docs are up-to-date: `make docs-check`

Docs auto-commit (default on)
- `make docs` will auto-commit doc/snippet changes by default with a Conventional Commit message:
  - `docs: bump provider snippets to ~> <version>; regenerate` (when VERSION is resolved)
  - `docs: regenerate with tfplugindocs` (when VERSION resolves to `dev`)
- To disable auto-commit for a run:
  ```bash
  AUTO_COMMIT=0 make docs
  ```

Please do not edit files under `docs/` manually; update schemas/descriptions in code instead and regenerate.

### Conventional Commits (recommended)

Use Conventional Commits for clear history and better automated changelogs. Format:

- `type(scope): short summary`
- Types: `feat`, `fix`, `docs`, `refactor`, `perf`, `test`, `build`, `ci`, `chore`.
- Breaking changes: add `!` after type/scope (e.g., `feat!: ...`) or include a `BREAKING CHANGE:` footer.
- Examples:
  - `feat(addon): support v2 API`
  - `fix(auth): refresh token before expiry`
  - `refactor(client): simplify retry logic`

Notes:
- Squash merges are fine—set the PR title to a Conventional Commit for cleaner release notes.

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

Releases are built from tags using GitHub Actions and GoReleaser. Release notes are generated from git history on GitHub Releases.

1. Ensure `main` (or your target `release/v*` branch) is green.
2. Create and push a semver tag (v‑prefixed) using Makefile helpers:
   ```bash
   make tag VERSION=vX.Y.Z
   # This will verify docs/snippets/examples first and fail on differences
   make tag-push VERSION=vX.Y.Z
   ```
   If you need to fix a tag, use:
   ```bash
   make tag-repush VERSION=vX.Y.Z
   ```
   Alternatively, you can do it manually:
   ```bash
   git tag vX.Y.Z && git push origin vX.Y.Z
   ```
3. GitHub Actions (GoReleaser) builds artifacts for multiple OS/ARCH and publishes the GitHub Release. Changelog sections are grouped by Conventional Commit types.

Notes:
- Publishing to the Terraform/OpenTofu registries is not wired up yet; current automation publishes GitHub releases only.
- The provider binary embeds the version via `-ldflags -X main.version=<tag>`.
 - The provider binary embeds the registry address at build time; we build artifacts for both Terraform and OpenTofu registries.

If `make tag-push` fails on docs:
- Run `make docs` (auto-commit on) or `AUTO_COMMIT=0 make docs` to inspect diffs.
- Commit the changes if needed, then re-run `make tag-push VERSION=vX.Y.Z`.


## Branch Protection

The `main` branch is protected. Changes land via PRs with passing CI checks.

Notes:
- The CI job `ci / lint-test` must pass.

## Issue Triage

- Use labels `bug`, `enhancement`, `docs`, `good first issue`.
- Link PRs to issues when applicable.
- Prefer minimal repros and logs in bug reports.

## Code of Conduct

Be respectful and constructive. The maintainers may guide scope to keep the project stable and focused.
- Local install: for `make install`, set `VERSION=vX.Y.Z` to install under the correct Terraform plugin path.
