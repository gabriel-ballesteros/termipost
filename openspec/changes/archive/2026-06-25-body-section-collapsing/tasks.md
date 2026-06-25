## 1. Fold component

- [x] 1.1 Add `internal/model/foldview.go` with a `foldView` struct holding prettified lines, fold regions, `collapsed map[int]bool`, and a `cursor` index.
- [x] 1.2 Implement `setJSON(s string) bool`: prettify via `syntax.Prettify`; on non-JSON / parse failure return false and leave the view non-foldable.
- [x] 1.3 Implement fold-region detection: single depth-stack pass over prettified lines, recording header line, close line, and kind for every `{ }`/`[ ]` at all depths; keep a region only when `close > header + 1` (multi-line), discarding empty/single-line containers.
- [x] 1.4 Implement `visibleLines()` skipping lines hidden under any collapsed region (collapsed parent hides nested content).
- [x] 1.5 Implement `moveUp()`/`moveDown()` constrained to visible lines.
- [x] 1.6 Implement `toggle()`: act on the region whose header is the cursor line, else the nearest enclosing region; keep cursor on a still-visible line.
- [x] 1.7 Implement `render(width int, focused bool)`: gutter (`-`/`+`/blank, aligned) + highlighted line via `syntax.HighlightJSON`; collapsed headers render `trim(header) + " … " + trim(closeLine)` so trailing commas survive; highlight the cursor line.
- [x] 1.8 Add a wrap-aware helper that returns the cursor line's screen-row range (summing `ceil(displayWidth/wrapWidth)` over visible lines) for viewport scroll-to-cursor.
- [x] 1.9 Cache prettified lines + regions; recompute (and reset `collapsed`/`cursor`) only when the source body changes.

## 2. Keybinding

- [x] 2.1 Add `keys.Fold` (`space`, help "fold/expand") in `internal/model/keys.go`.

## 3. Response pane integration

- [x] 3.1 Add a `fold foldView` field to `responsePane`; seed it in `setResponse` for JSON bodies.
- [x] 3.2 In `View`/`content`, feed `fold.render(...)` into the viewport for JSON bodies; keep current plain/highlighted display for non-JSON.
- [x] 3.3 In `Update`, map `up`/`down` to fold cursor moves and `space` to `toggle()` when foldable; after each move, `SetYOffset` from the wrap-aware row range (1.8) to keep the cursor visible; keep PageUp/Down/mouse on raw viewport scroll; leave copy/tab and non-JSON behavior as-is.
- [x] 3.4 Advertise `keys.Fold` in `HelpBindings` only when a foldable body is shown.

## 4. Request body preview integration

- [x] 4.1 Add a `bodyFold foldView` field to `editorPane`; re-seed it from the body in `load` and whenever the body value changes (e.g. after prettify / leaving edit).
- [x] 4.2 In `viewBodyTab` (not editing), render `bodyFold.render(...)` only when the JSON body has foldable (multi-line) sections; keep `bodyPreview` for non-JSON and single-line JSON.
- [x] 4.3 In the Body-tab key path (not editing, focused), handle `up`/`down` for cursor moves and `space` for `toggle()`; ensure `space` is not swallowed by other handlers; entering edit (`enter`/`i`) still disables folding.
- [x] 4.4 Advertise `keys.Fold` in `editorPane.HelpBindings` only on the Body tab when not editing and the body is foldable.

## 5. Tests & verification

- [x] 5.1 Unit-test `foldView`: region detection over nested JSON, `visibleLines` under collapsed parents, cursor skipping hidden lines, `toggle` from header and non-header lines, non-JSON returns non-foldable.
- [x] 5.2 Test gutter rendering: `-` on expanded headers, `+` on collapsed, blank on non-headers/single-line/empty containers, and collapsed placeholder preserving a trailing comma.
- [x] 5.5 Test wrap-aware scroll: cursor moving past the window scrolls the viewport, including a soft-wrapped long line spanning multiple rows.
- [x] 5.3 Add/extend pane render tests confirming folding shows only in read-only views and that the stored request body is unchanged after folding.
- [x] 5.4 Run `make test` (or `go test ./...`) and confirm green.
