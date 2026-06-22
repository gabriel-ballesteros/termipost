# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [1.2.1](https://github.com/gabriel-ballesteros/termipost/releases/tag/v1.2.1) - 2026-06-22

Internal only — no user-facing changes to the `termipost` binary.

### Changed

- Added a unit and TUI test suite across the `domain`, `httpclient`, `model`,
  `runner`, `store`, and `vars` packages.
- Added a CI workflow that builds, vets, and tests on every pull request to
  `main` and `develop`, and uploads coverage to Codecov.
- Added Codecov project and patch coverage gates (80% targets) so pull requests
  must keep both whole-repo and changed-line coverage at or above the threshold.
- Added build, coverage, code-quality, release, and license badges to the README.

## [1.2.0](https://github.com/gabriel-ballesteros/termipost/releases/tag/v1.2.0) - 2026-06-21

### Added

- Status-code assertions support a not-equals operator that passes when the
  response status differs from the expected value (summarized as `status code != N`).
- The title bar shows the running version (e.g. `termipost v1.2.0`) and a
  breadcrumb of the open screens (e.g. `Collections › Requests › Edit request`)
  so it is clear where Esc leads.

### Changed

- The action bar shows the run and test shortcuts as `shift+r` and `shift+t`
  (and `shift+r` for "run all") to make the required Shift explicit.

### Fixed

- The assertion editor no longer takes an extra ↓ to move past a hidden field
  (e.g. reaching "Max ms" for a latency assertion); navigation now skips hidden
  fields.

## [1.1.0](https://github.com/gabriel-ballesteros/termipost/releases/tag/v1.1.0) - 2026-06-21

### Added

- Request editor: every field, including Assertions, is reachable with Tab/arrow
  keys and openable with Enter; first-letter shortcuts jump straight to a field
  (`n` name, `m` method, `u` url, `h` headers, `p` params, `b` body,
  `a` assertions).

### Changed

- The request method now cycles with the arrow keys; the `h`/`l` bindings were
  removed so `h` can jump to the Headers field.

### Fixed

- Screens opened after startup (a collection's request list, the environments
  list) now render their items immediately instead of showing only a pagination
  indicator.

## [1.0.4](https://github.com/gabriel-ballesteros/termipost/releases/tag/v1.0.4) - 2026-06-21

### Added

- Windows builds for amd64 and arm64, published as `.zip` archives.

### Changed

- Homebrew distribution switched from a formula to a cask (the modern vehicle for
  pre-built binaries); the cask is published under `Casks/` in the tap. Install is
  unchanged: `brew install gabriel-ballesteros/tap/termipost`.

## [1.0.3](https://github.com/gabriel-ballesteros/termipost/releases/tag/v1.0.3) - 2026-06-21

### Changed

- Homebrew formula is now published under `Formula/` in the tap (was the repo root).

## [1.0.2](https://github.com/gabriel-ballesteros/termipost/releases/tag/v1.0.2) - 2026-06-21

### Added

- MIT license.
- Homebrew installation via tap: `brew install gabriel-ballesteros/tap/termipost`.

## [1.0.1](https://github.com/gabriel-ballesteros/termipost/releases/tag/v1.0.1) - 2026-06-21

### Added

- `termipost --version` prints the build version (injected at release time).

## [1.0.0](https://github.com/gabriel-ballesteros/termipost/releases/tag/v1.0.0) - 2026-06-21

### Added

- Terminal UI application (`termipost`) built with Bubble Tea, Bubbles, and Lip Gloss.
- Collection management: create, rename, delete, and browse collections.
- Request management: create, edit, and delete requests with method, URL, headers,
  query parameters, and body.
- Run action that sends a request and shows the response (status, headers,
  pretty-printed JSON body, and elapsed time), independent of assertions.
- Copy the raw response body to the system clipboard from the response view
  (`c` / `y`; `Ctrl+C` is reserved for quit).
- API testing: attach assertions (status code, header, body, latency) to a
  request, test a single request, and run a whole collection with an aggregate
  pass/fail/skip summary.
- Environment management: multiple named environments with one active at a time,
  and `{{name}}` variable substitution resolved against the active environment
  then the global secrets store.
- Global secrets store kept in a gitignored file, with values masked everywhere
  in the UI.
- Human-readable JSON persistence with atomic writes, resilient loading of
  missing/malformed files, and an auto-generated `.gitignore` for secrets.
- Keyboard-only navigation with a contextual action/help bar on every screen.

[unreleased]: https://github.com/gabriel-ballesteros/termipost/commits/develop
