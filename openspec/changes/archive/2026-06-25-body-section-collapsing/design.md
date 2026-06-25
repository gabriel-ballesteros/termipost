## Context

termipost is a Bubble Tea TUI. Bodies are rendered in two read-only places:

- **Request body** ([pane_editor.go](../../../internal/model/pane_editor.go), `etBody` tab): when not editing, `viewBodyTab` shows `bodyPreview` ([panes_shared.go](../../../internal/model/panes_shared.go)), which highlights JSON and hard-truncates to 6 lines inside a `ui.Box`. Editing swaps to a `textarea.Model`.
- **Response body** ([pane_response.go](../../../internal/model/pane_response.go)): `prettyBody` indents + highlights JSON into a scrollable `viewport.Model`. There is no line cursor today; arrow keys scroll the viewport.

JSON structure is available from the hand-written scanner in [syntax.go](../../../internal/syntax/syntax.go), which already emits a flat token stream with `tkPunct` for brackets. Highlighting and prettify reuse this scanner.

Folding is additive: highlighting/prettify/validation (the `payload-formatting` capability) are reused unchanged. Per the proposal, folding is read-only and all nesting depths are foldable; the toggle key is `space`.

## Goals / Non-Goals

**Goals:**
- A reusable fold component that, given formatted JSON text, computes foldable regions, tracks per-section collapsed state and a line cursor, and renders the gutter + visible lines with existing highlighting.
- Integrate the component into both the request body preview (read-only) and the response body view.
- `up`/`down` move the cursor over visible lines; `space` toggles the section at/around the cursor.

**Non-Goals:**
- Folding inside the live request `textarea` (editing stays plain text).
- Folding non-JSON bodies.
- Persisting fold state across reloads, requests, or app restarts.
- Horizontal scrolling or per-token cursoring (cursor is line-granular).

## Decisions

### Decision: A line-based fold model built from the formatted JSON, not the token stream directly
Prettify already produces canonical one-structural-element-per-line JSON (`json.Indent` with 2-space indent). Rather than fold arbitrary source, the fold component operates on the **prettified** text so each foldable section maps cleanly to a contiguous line range whose header line ends in `{` or `[` and whose closing line is the matching `}` or `]` at the same indent.

- Build the line list by prettifying the JSON (reusing `syntax.Prettify`; fall back to plain non-foldable display if it fails).
- Compute fold regions with a single pass using a bracket-depth stack over line-final `{`/`[` and line-initial `}`/`]`. Each region records `headerLine`, `closeLine`, and `kind` (object/array).
- This avoids re-implementing structural matching against raw tokens and keeps the gutter aligned with what the user sees.

*Alternative considered:* fold directly over the `syntax` token positions. Rejected — it requires mapping byte offsets to wrapped display lines and duplicates what prettify already normalizes.

### Decision: A standalone `foldView` value type in `internal/model`
Add `internal/model/foldview.go` defining a `foldView` struct holding the prettified lines, the fold regions, a `collapsed map[int]bool` keyed by header line, and a `cursor` line index. Methods:

- `setJSON(s string) bool` — prettify + recompute regions; returns false (caller renders plainly) when not JSON.
- `visibleLines() []int` — header lines plus lines not hidden under any collapsed region.
- `moveUp()/moveDown()` — move `cursor` within visible lines.
- `toggle()` — find the region whose header is the cursor line, else the nearest enclosing region, and flip its `collapsed` state; keep cursor on a still-visible line.
- `render(width int, focused bool) string` — emit gutter (`-`/`+`/space) + highlighted line; collapsed headers render `{ … }` / `[ … ]`; cursor line styled via a `ui` highlight.

### Decision: Region detection and which sections are foldable
After prettify, structure is canonical — an opener (`{`/`[`) is the last non-space char on its line and its closer (`}`/`]`) is the first non-space char on the matching line. Detection is a single depth-stack pass: a line ending in an opener pushes `{headerLine, kind}`; a line starting with a closer pops, yielding a region `{header, close, kind}`.

A region is foldable only when it **spans more than one line** (decision #2/#5): `close > header + 1`, i.e. it has at least one body line. Empty `{}`/`[]` and any object/array that prettify keeps on a single line are therefore not foldable and get a blank gutter. Every multi-line section at every depth is an independent region (decision #3), including the root object/array and each element of an array of objects.

### Decision: Collapsed rendering preserves the close-line tail (comma trap)
A collapsed section renders as `trim(headerLine) + " … " + trim(closeLine)`. Using the trimmed close line (not a bare bracket) carries any trailing comma that followed the closer, so `"b": {\n …\n },` collapses to `"b": { … },` and the surrounding JSON stays comma-correct on screen.

Keeping it in `internal/model` (not `internal/syntax`) lets it depend on `ui` styles and stay UI-only, matching how the panes already mix rendering and state.

*Alternative considered:* put fold logic in `internal/syntax`. Rejected — `syntax` is intentionally a pure scanner with no UI/style coupling.

### Decision: Response pane drives the viewport from the fold view
`responsePane` gains a `fold foldView`. `setResponse` calls `fold.setJSON` on JSON bodies. `content()`/`View` feed `fold.render(...)` into the existing `viewport` (so long bodies still scroll, with the cursor line kept in view). Key handling: `up`/`down` move the fold cursor (instead of raw viewport scroll) when foldable; `space` toggles; copy/tab unchanged. When the body is non-JSON, behavior is exactly as today.

### Decision: Cursor keeps itself visible with soft-wrap intact (Option A)
Two positions coexist: the fold `cursor` (logical visible-line index) and the viewport `YOffset` (first screen row shown). Soft-wrap stays on, so one logical line can span several screen rows and `cursor != screen row`. After each cursor move, compute the cursor line's **screen-row range** by summing the wrapped height (`ceil(displayWidth(line)/wrapWidth)`, min 1) of the visible lines above it, then scroll the viewport so that range is fully in view (`SetYOffset` up if above the window, down if below). `PageUp`/`PageDown`/mouse keep driving raw viewport scroll; only `up`/`down` move the cursor. This is the "scrolloff=0" editor pattern.

*Alternative considered:* disable soft-wrap in fold mode (truncate long values with `…`) so cursor == row trivially. Rejected — full values stay readable matters more than the row math, and wrap height is a small helper.

### Decision: Request body preview becomes a foldable read-only view
`editorPane` gains a `bodyFold foldView`. In `viewBodyTab` (not editing), when the body is JSON **and has foldable (multi-line) sections**, render `bodyFold.render(...)` instead of the 6-line `bodyPreview` truncation; a JSON body with no multi-line section (e.g. a one-line `{"a":1}`) and any non-JSON body keep the current `bodyPreview`. Navigation only when the Body tab is focused and not editing: `up`/`down` move the cursor, `space` toggles. Entering edit (`enter`/`i`) is unchanged and disables folding; on load/value change the fold view is re-seeded from the current body. `space` must not be swallowed by other Body-tab handlers when not editing.

### Decision: `space` binding added centrally
Add `keys.Fold = key.NewBinding(key.WithKeys(" "), key.WithHelp("space", "fold/expand"))` in [keys.go](../../../internal/model/keys.go) and advertise it in both panes' `HelpBindings` only when a foldable body is shown and not editing.

## Risks / Trade-offs

- **Prettify changes whitespace of the displayed request body** → The read-only preview already highlights and the editor stores raw text; folding view is display-only and does not write back, so the stored body is untouched. Editing still shows the raw textarea content.
- **Response pane currently uses arrow keys for viewport scroll; remapping to cursor moves changes feel** → `up`/`down` move the cursor and the viewport auto-follows (Option A); PageUp/Down/mouse keep raw scroll, so large bodies stay navigable.
- **Soft-wrap breaks cursor↔row alignment** → A wrapped logical line spans multiple screen rows, so naive `YOffset = cursor` would mis-scroll. Mitigation: compute the cursor's screen-row range from summed wrapped line heights before scrolling (Option A). If this proves fiddly, fallback is truncating long lines in fold mode.
- **Fold state lost on reload/re-run** → Accepted for v1 (decision #4): `setJSON` resets `collapsed` and `cursor` each time the body changes. Persisting fold state per request is a future enhancement.
- **`space` collisions** → `space` is currently unbound in both panes; gate it to read-only foldable state to avoid interfering with the textarea (which never reaches fold handling while editing).
- **Very large JSON re-prettified each render** → Cache prettified lines + regions; recompute only when the source body changes, not every frame.
- **Malformed/streamed JSON** → `setJSON` returns false and the view degrades to the existing plain/highlighted display with no gutter.
