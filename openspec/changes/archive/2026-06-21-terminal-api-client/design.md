## Context

`termipost` is a greenfield Go terminal application that brings a Postman-like workflow (collections, requests, response inspection, environment switching, and assertion-based tests) into the terminal. The UI must be fully keyboard-driven and always show the available actions for the current screen, and all data must live in human-readable JSON files.

The Charm ecosystem is the de-facto standard for Go TUIs and is specified by the request: Bubble Tea (the Elm-style runtime), Bubbles (reusable components: list, textinput, textarea, viewport, spinner, help, key), and Lip Gloss (styling/layout). This document records the architectural choices made before implementation; motivation lives in `proposal.md` and the testable requirements live in `specs/`.

## Goals / Non-Goals

**Goals:**
- A single-binary Go TUI built on Bubble Tea, Bubbles, and Lip Gloss.
- Collections and the requests inside them are CRUD-manageable and persisted as JSON.
- Requests support method, URL, headers, query params, and body; sending is non-blocking and shows a formatted response.
- Assertions (status code, headers, body, latency) attach directly to requests; running a request evaluates its assertions; running a collection runs all requests that have assertions.
- Multiple environments with one active at a time; `{{name}}` references resolve against the active environment then the global secrets store.
- Secrets kept in a single gitignored file and masked everywhere in the UI.
- 100% keyboard operability and a persistent, context-aware action/help bar on every screen.
- Resilient file persistence: human-readable, hand-editable, and tolerant of missing/malformed files.

**Non-Goals:**
- No mouse interaction, GUI, or web server.
- No separate test or test-suite entities — a "test" is a request with assertions, and the collection is the batch-run unit.
- No layered variable resolution (no collection-level or global variable scopes beyond the active environment + secrets).
- No per-environment secrets in v1 — secrets are a single global store.
- No request scripting/pre-request scripts beyond `{{variable}}` substitution.
- No cloud sync, sharing, or team workspaces; no OAuth helper flows (auth is just headers); no Postman/OpenAPI import/export (future work).

## Decisions

### Architecture: Elm-style model with a screen/state machine
Use one root Bubble Tea `Model` that holds shared state (loaded data, active environment, terminal size, status/error line) and a `screen` enum identifying the active view. Each screen is its own sub-model implementing `Update`/`View`, and the root delegates to the active screen. A small navigation stack supports "back" (Esc) behavior.

- *Why:* Bubble Tea is single-threaded Elm-architecture; a screen state machine with a nav stack is the idiomatic way to model multi-view flows and back navigation.
- *Alternative considered:* A flat model with conditional rendering — rejected because it becomes unmaintainable as screens grow and tangles the action-bar logic.

### Screens
- Collection list → Request list (within a collection) → Request editor / Response view, with an Assertions panel on the request.
- Environment list / Environment editor and a Secrets editor, reachable from a top-level menu.
- The action/help bar is rendered by the root using `bubbles/help` + `bubbles/key`, driven by the active screen's declared key map, so help text and actual bindings never drift apart.

### Data model: two entities, not four
The graph collapses to collections and requests; assertions are nested on requests, and environments/secrets are a separate concern.

```
Collection
  └─ Request (method, url, headers, query params, body)
       └─ Assertion[]  (status_code | header | body | latency)
Environment (id, name, vars: map[string]string)   // multiple, one active
Secrets (map[string]string)                        // single global, gitignored
```

- *Why:* A "test" is just a request carrying assertions, so it needs no separate identity, and there is no dangling-reference problem between tests and requests. The collection replaces the suite as the batch-run unit — which is also how Postman's collection runner works.

### HTTP execution: stdlib `net/http` inside a `tea.Cmd`
Requests run in a `tea.Cmd` (goroutine) and return a result message; the UI shows a `bubbles/spinner` while in flight and stays responsive.

- *Why:* Keeps the Update loop non-blocking (required by the non-blocking-send scenario). `net/http` covers all needs without extra dependencies.
- *Alternative considered:* A third-party HTTP client (resty) — rejected to keep dependencies minimal.

### Variable & secret resolution
Before sending, substitute `{{name}}` references in URL, headers, query params, and body. Resolution is single-layer: look up the active environment's vars first, then the global secrets store. Unresolved references are left intact and the user is warned. Values sourced from secrets are tracked so any "resolved request" preview masks them; the real value is still placed in the outgoing request.

- *Why single-layer:* avoids precedence complexity in both resolution and UI; the active-environment + secrets split covers the local/staging/prod use case.
- *Known limitation:* a single global secrets file cannot hold two different values for the same key across environments (e.g. distinct prod vs local tokens). Workaround is name-prefixing (`prod_token`, `local_token`). Per-environment secrets are a deliberate future evolution; because secrets are referenced by name, that change is backward-compatible.

### Run vs. test as separate actions
The request editor exposes two distinct actions: **run** (`R`) sends the request and shows the response without regard to assertions, and **test** (`T`) sends the request and evaluates its assertions. Run works on any request; test requires at least one assertion and otherwise prompts the user to add one. This keeps "just hit the endpoint and look at the response" a first-class, friction-free path separate from validation.

### Copying the response body
The response view can copy the raw (unformatted) response body to the system clipboard via `github.com/atotto/clipboard`. The copy binding is `c` (with `y` as a vim-style alias) rather than Ctrl+C, because Ctrl+C is globally reserved for quitting the program. `atotto/clipboard` is already pulled in transitively by Bubbles, so this adds no new external surface.

### Body & response formatting
Request body is free-text via `bubbles/textarea`. Response body is shown in a `bubbles/viewport`; if `Content-Type` is JSON (or the body parses as JSON), pretty-print with `encoding/json` indentation and cache the formatted string.

### Assertions
Assertion types: `status_code` (equals), `header` (name + equals/contains/regex), `body` (contains / equals / json-path-equals), `latency` (max ms). Each evaluates against the captured response and yields pass/fail with expected-vs-actual detail. JSON field assertions use a minimal dotted-path lookup over the decoded body. A collection run evaluates each request that has assertions, skips those without, and aggregates a pass/fail/skip summary.

### Data model & storage layout
JSON files under `os.UserConfigDir()/termipost/` (fallback `~/.termipost/`):
- `config.json` — app config + active environment id.
- `collections/<id>.json` — one file per collection containing its requests and their nested assertions.
- `environments/<id>.json` — one file per environment containing its variable map.
- `secrets.json` — single global secret key→value map (gitignored).
- `.gitignore` — generated to exclude `secrets.json`.

Each entity has a stable `id` (slug + short random suffix) plus a display `name`. Writes use atomic write-to-temp-then-rename and `json.MarshalIndent` for readability.

- *Why per-collection / per-environment files:* avoids one giant file, reduces noisy diffs and write contention when hand-editing or version-controlling, and makes deletion a single file removal.
- *Alternative considered:* Single `data.json` — rejected for poor scalability and noisy diffs. YAML/TOML considered but JSON chosen for zero-dependency stdlib support.

### Project layout
`main.go` plus `internal/` packages: `model` (root + screens), `httpclient` (send), `store` (load/save JSON + gitignore), `domain` (Collection/Request/Assertion/Environment types), `runner` (assertion evaluation + collection runs), `vars` (resolution + masking), `ui` (styles/keymaps). Keeps concerns separated and unit-testable independent of the TUI.

## Risks / Trade-offs

- **Bubble Tea key conflicts between text entry and global shortcuts** → Track an explicit edit/navigation mode; suspend global single-key shortcuts (q, j/k) while a text field is focused.
- **Secret leakage into previews/screenshots** → Track secret-sourced values through resolution and mask them in any rendered "resolved request" view; only the outgoing HTTP request carries the real value.
- **Global secrets collision across environments** → Documented limitation; mitigate with name-prefixing now, evolve to per-environment secrets later without breaking name-based references.
- **Hand-edited JSON can become malformed** → Resilient loader reports the error and continues without overwriting the bad file; never silently clobber user files.
- **Large response bodies could degrade rendering** → Use `viewport` for scrolling and cache the formatted body instead of re-pretty-printing per frame.
- **Concurrent in-flight request + quit** → Use Bubble Tea `Cmd` cancellation / `context` so quitting does not hang on a slow request.

## Open Questions

- Do we need request history/response saving across sessions, or is in-session only acceptable for v1? (Assuming in-session only.)
- Should a collection run support ordering control / stop-on-first-failure, or always run all and summarize? (Assuming run-all + summary for v1.)
