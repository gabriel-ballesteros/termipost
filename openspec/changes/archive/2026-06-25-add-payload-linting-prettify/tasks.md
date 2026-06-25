## 1. JSON scanner (single source: highlight + validate)

- [x] 1.1 Create `internal/syntax/syntax.go` with a position-tracking JSON scanner that walks input byte-by-byte tracking line/column.
- [x] 1.2 Expose `HighlightJSON(s string) string` that wraps tokens (key, string, number, keyword, punctuation); on any unexpected byte, return the input unstyled.
- [x] 1.3 Expose `ValidateJSON(s string) (bool, *SyntaxErr)` performing a full parse and returning `SyntaxErr{Line, Col, Msg}` on failure.
- [x] 1.4 Add JSON token styles (`JSONKey`, `JSONString`, `JSONNumber`, `JSONKeyword`, `JSONPunct`) to `internal/ui/ui.go` reusing the existing palette.
- [x] 1.5 Add `internal/syntax/syntax_test.go` with a correctness stress corpus: escapes, `\uXXXX` unicode, nested objects/arrays, deep nesting, edge numbers (`1e10`, `-0`, `1.5E-3`), rejected leading zeros, plus non-JSON and malformed passthrough for highlight. Assert zero false-invalids on valid JSON.

## 2. Prettify (format + validate)

- [x] 2.1 Add a prettify helper that calls `ValidateJSON`; on success formats with `json.Indent(..., "", "  ")` and returns `(formatted, nil)`, on failure returns the scanner's `SyntaxErr` and the original input unchanged.
- [x] 2.2 Treat empty/whitespace-only input as a no-op with no error.
- [x] 2.3 Unit-test prettify: valid JSON, already-formatted (idempotent), empty, and malformed-with-line/column cases.

## 3. Response view highlighting

- [x] 3.1 In `internal/model/panes_shared.go`, pass the `json.Indent` output of `prettyBody()` through `syntax.HighlightJSON`.
- [x] 3.2 Verify viewport soft-wrap in `pane_response.go` still behaves with ANSI-styled content.
- [x] 3.3 Update/add `prettyBody` tests in `screens_test.go` asserting highlighting for JSON and plain text for non-JSON.

## 4. Request body: highlight preview + prettify action

- [x] 4.1 Highlight the JSON body preview path in `pane_editor.go` (`bodyPreview` / `viewBodyTab`) via `syntax.HighlightJSON`.
- [x] 4.2 Add a prettify key binding in `internal/model/keys.go` (collision-checked against existing bindings).
- [x] 4.3 Wire the binding in `editorPane.key` under the `etBody` tab (navigation and body-edit modes): call the prettify helper, write the result back into the body textarea on success, mark dirty if changed.
- [x] 4.4 Route prettify failures to `m.setError` with the line/column message and successes to `m.setStatus`; leave the body unchanged on failure.
- [x] 4.5 Add the prettify binding to `editorPane.HelpBindings()` so the help bar shows it on the Body tab.

## 5. Live JSON validation while editing

- [x] 5.1 On body keystrokes (`etBody` editing), gate on `looksLikeJSON` (trimmed starts with `{` or `[`); when gated in, run `ValidateJSON`.
- [x] 5.2 Render a valid/invalid indicator plus inline error line beneath the textarea in `viewBodyTab`; show nothing for empty or non-JSON bodies.
- [x] 5.3 Reuse `ui` styles (e.g. `Good`/`Bad`) for the indicator; keep the body text itself uncoloured while editing.
- [x] 5.4 Add model tests: JSON-looking invalid → indicator+error; valid → indicator, no error; non-JSON/empty → no indicator.

## 6. Verification

- [x] 6.1 Add/extend model tests for the prettify key path (valid → formatted + dirty; invalid → unchanged + error).
- [x] 6.2 Run `make` / `go test ./...` and `gofmt`; fix any failures.
- [x] 6.3 Manually verify highlighting, prettify, and live validation in the running TUI for JSON, non-JSON, and malformed bodies.
- [x] 6.4 Update `README.md` Features to mention body syntax highlighting, prettify, and live JSON validation.
