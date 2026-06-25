## Context

termipost is a Bubble Tea TUI. Body content appears in two places:

- The request editor `etBody` tab (`internal/model/pane_editor.go`): a
  `textarea.Model` while editing, and `bodyPreview()` (first 6 lines) when not.
- The response pane (`internal/model/pane_response.go`): rendered through
  `prettyBody()` in `internal/model/panes_shared.go`, which already runs
  `json.Indent` for JSON content and otherwise returns the raw bytes.

Styling is centralised in `internal/ui/ui.go` as `lipgloss` styles. The project
keeps its dependency set deliberately small (`bubbletea`, `bubbles`, `lipgloss`,
`clipboard`). There is no existing syntax-highlighting machinery.

## Goals / Non-Goals

**Goals:**
- Colourise JSON body content in both the request editor preview and the
  response view.
- Add a prettify action to the request body that re-indents JSON in place.
- Validate JSON on prettify, leaving the body untouched and surfacing a parse
  error with line/column on failure.
- Show live JSON validity feedback while editing the body (valid/invalid
  indicator + inline error), gated so it only appears for JSON-looking content.
- Fail safe: non-JSON and malformed content render as plain text.

**Non-Goals:**
- Per-token colouring *inside* the `textarea` while typing — the textarea widget
  renders flat runes and cannot style per-token without a fork. Highlighting
  applies to the non-editing preview and the response view; while editing we
  provide live *validity* feedback only, not colour.
- Languages other than JSON (XML/HTML/YAML) — structure the code so another
  lexer can be added later, but only JSON ships now.
- Prettify of non-JSON bodies.
- Any third-party JSON parser or syntax-highlighting library.

## Decisions

### One in-repo, position-tracking JSON scanner serves all three jobs
Add `internal/syntax` with a single hand-written JSON scanner that walks the
input byte-by-byte tracking line/column. Because it already knows position and
token boundaries, it powers all three needs from one code path:

- `HighlightJSON(s string) string` — wraps tokens with `lipgloss` styles (key,
  string, number, keyword, punctuation). On any unexpected byte it bails and
  returns the input unstyled (graceful degradation).
- `ValidateJSON(s string) (ok bool, err *SyntaxErr)` — full parse; on failure
  returns a `SyntaxErr{Line, Col, Msg}`.

A full syntax library (e.g. `chroma`) would pull a large dependency tree for one
language, and a third-party JSON parser would duplicate position tracking the
scanner already does. The JSON grammar is tiny; one hand-written scanner keeps
the dep set at zero and gives precise, controllable error locations — better than
`encoding/json`, whose `*json.SyntaxError.Offset` points *after* the bad token
with a terse message.

Alternatives considered: `json.Decoder.Token()` — discards whitespace, can't
reproduce the indented layout. Third-party parser (`goccy/go-json`,
`encoding/json/v2`) — rejected; redundant with the scanner we write anyway, and
adds a dep / GOEXPERIMENT gate for no error-quality win over our own scanner.

**Correctness burden**: because the scanner now *validates* (not just colours), a
false "invalid" on good JSON is the trust-killing failure mode. It must be a
correct parser, covered by a stress corpus: escapes, `\uXXXX` unicode, nested
objects/arrays, deep nesting, and edge numbers (`1e10`, `-0`, `1.5E-3`, leading
zeros rejected).

### Prettify: format with `json.Indent`, errors from the scanner
The prettify action runs `ValidateJSON` first; on success it formats with
`json.Indent(..., "", "  ")` (same indent already used by `prettyBody` and the
store) and writes the result back into the body textarea, marking dirty if
changed. On failure it reports the scanner's `SyntaxErr` (line/column + message)
and leaves the body unchanged. `encoding/json` is used only for formatting, never
for error reporting.

### Live validation while editing the body
On each keystroke while editing the `etBody` tab, if the trimmed body looks like
JSON (`looksLikeJSON`: starts with `{` or `[`), run `ValidateJSON` and render a
valid/invalid indicator plus the inline error line beneath the textarea. When the
body is empty or does not look like JSON, show no indicator (avoids nagging on
plain-text / form-encoded bodies). Validation is linear over the body and runs
only while the body field is focused, so cost is bounded by typing cadence.

### New styles in `internal/ui`
Add token styles (`JSONKey`, `JSONString`, `JSONNumber`, `JSONKeyword`,
`JSONPunct`) reusing the existing palette colours so the theme stays coherent.

### Key binding for prettify
Add a binding in `internal/model/keys.go` (e.g. `f` "format" or `ctrl+f`) wired
in `editorPane.key` under the `etBody` tab, active in navigation mode and while
editing the body. It calls the prettify helper and routes success/failure to
`m.setStatus` / `m.setError`. Final key choice confirmed during implementation
against existing bindings to avoid collisions.

### Wiring points
- `prettyBody()` returns highlighted output: after `json.Indent` succeeds, pass
  the indented string through `syntax.HighlightJSON`.
- `bodyPreview()` (and/or the body preview path in `viewBodyTab`) highlights the
  JSON preview lines.

## Risks / Trade-offs

- [Highlighting cost on large bodies re-running every render] → operate on the
  already-indented string and keep the scanner linear; the response viewport
  already re-renders content each frame, so cost is comparable to existing wrap.
- [lipgloss colour codes inflate string width / interact with viewport wrap] →
  wrap with `lipgloss.NewStyle().Width(...)` as today; verify ANSI-aware
  wrapping still behaves on the highlighted output.
- [Scanner now validates, so a false "invalid" on good JSON kills trust — worse
  than highlighting bugs, which only degrade to plain text] → treat the scanner
  as a correctness-critical parser; gate merge on the stress corpus (escapes,
  `\uXXXX`, nesting depth, edge numbers) passing.
- [Live validation on every keystroke could feel laggy on huge bodies] →
  validation is linear and only runs while the body is focused and looks like
  JSON; revisit with debouncing only if a real body triggers lag.
- [Live red indicator nagging on non-JSON bodies] → gated on `looksLikeJSON`, so
  plain-text / form-encoded bodies show no indicator.
- [Prettify error line/column off-by-one] → unit-test the scanner's reported
  positions against known malformed inputs.

## Open Questions

- Exact prettify key binding (pending collision check in implementation).
- Visual treatment of the live validity indicator (gutter dot vs status text) —
  settle during implementation against available space in the body tab.
