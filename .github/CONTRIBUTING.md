# Contributing to Kage

## Reporting Issues

Did you find a bug or have a suggestion for improvement?

1. Check if the issue already exists by searching the [Issues](https://github.com/nanoninja/kage/issues) section.
2. If not, [open a new one](https://github.com/nanoninja/kage/issues/new). Include a clear title, a minimal reproduction case, and your Go version.

## Before You Start

Open an issue to discuss your idea before submitting a pull request. This avoids duplicate work and ensures the change aligns with the project direction.

## Ground Rules

- No external dependencies. Kage is stdlib-only by design.
- No breaking changes without a major version bump.
- All exported symbols must have godoc comments.
- All new code must be covered by tests. Run with `-race` to check for data races.

## Commit Messages

Follow [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>(<scope>): <short description>

Types: feat, fix, refactor, test, docs, ci, chore
Scope: middleware, router, or omit for top-level changes

Examples:
  feat(middleware): add RateLimit middleware
  fix(router): normalize trailing slash in wrapPath
  docs: update README with Mount example
```

## Pull Request Process

1. Fork the repository and create a branch: `feature/<name>` or `fix/<name>`.
2. Make your changes with tests.
3. Run `go test -v -race ./...` and `golangci-lint run ./...` — both must pass.
4. Update `CHANGELOG.md` under `[Unreleased]`.
5. Open a pull request against `main`.

## Release Process (maintainers)

1. Move `[Unreleased]` entries to a new versioned section in `CHANGELOG.md`.
2. Tag the commit: `git tag vX.Y.Z && git push origin vX.Y.Z`.
3. Create a GitHub release using the CHANGELOG section as release notes.
