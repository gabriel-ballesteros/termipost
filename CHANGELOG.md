# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

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
