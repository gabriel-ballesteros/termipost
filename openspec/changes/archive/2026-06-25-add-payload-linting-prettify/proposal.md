## Why

Request and response bodies are shown as flat, single-colour text, which makes
JSON payloads hard to scan and easy to get wrong. Users editing a request body
have no way to auto-format their JSON or to catch a syntax mistake before
sending — a malformed body is only discovered after a failed request.

## What Changes

- Add **syntax highlighting** for body content (JSON first) in both the request
  editor body field and the response body view: keys, strings, numbers,
  booleans/null, and punctuation get distinct colours.
- Add a **prettify** action in the request body editor that re-indents the JSON
  payload in place.
- The prettify action **validates JSON syntax**: on success it formats the body;
  on failure it leaves the body untouched and reports the parse error (with
  line/column location) to the user.
- Add **live JSON validity feedback** while editing the body: when the content
  looks like JSON, show a valid/invalid indicator and the parse error inline as
  the user types. This is feedback only — the body text itself is highlighted in
  the non-editing preview, not per-token while editing.
- Highlighting degrades gracefully: non-JSON or unparseable content is shown as
  plain text rather than garbled or erroring.

## Capabilities

### New Capabilities
- `payload-formatting`: syntax highlighting of request/response body content and
  the prettify (format + validate) action over editable request bodies.

### Modified Capabilities
- `request-management`: the request body editor gains a prettify action and
  highlighted display; the existing JSON response formatting requirement is
  extended to also highlight the formatted body.

## Impact

- Code: `internal/model/pane_editor.go` (body edit/preview, prettify key),
  `internal/model/pane_response.go` and `internal/model/panes_shared.go`
  (`prettyBody` highlighting), `internal/model/keys.go` (new binding), and a new
  highlighter/linter helper under `internal/` (e.g. `internal/syntax`).
- Dependencies: none. A single in-repo, position-tracking JSON scanner under
  `internal/syntax` powers highlighting, live validation, and prettify error
  locations; formatting reuses `encoding/json`'s `json.Indent`.
- No data model or persistence changes; bodies are still stored verbatim.
