## Why

Request and response bodies are often large, deeply-nested JSON documents. The current read-only views either truncate the request body to its first lines or render the full response with no way to focus on a region. Users need to collapse uninteresting objects/arrays so they can navigate large payloads and keep the structure they care about on screen.

## What Changes

- Add **section collapsing** to the read-only JSON views of both the request body (editor "Body" tab preview) and the response body.
- Render a **left gutter** on every line of a foldable JSON view: `-` on a line that opens an expanded object/array, `+` when that section is collapsed, and a blank space on lines that are not fold headers.
- Add a **line cursor** to these read-only views so the user can move up/down across visible lines; the cursor highlights the line whose section will be toggled.
- Add a **toggle binding** (`space`) that collapses the section under the cursor when expanded and expands it when collapsed. Collapsed sections render as a single header line with a placeholder (e.g. `{ … }` / `[ … ]`).
- Support **all nesting depths**: every `{ }` and `[ ]` block is independently foldable.
- Collapsing is a **view-only** concern: it never alters the stored body, and it does not apply while the request body textarea is being edited.

## Capabilities

### New Capabilities
- `body-section-collapsing`: navigable, gutter-marked folding of nested JSON objects/arrays in the read-only request and response body views, including the line cursor, fold-state model, and the collapse/expand toggle.

### Modified Capabilities
<!-- None: payload-formatting (highlighting/prettify/validation) is reused unchanged; folding is an additive view layer. -->

## Impact

- **Code**: `internal/model/pane_response.go` (response body view + key handling), `internal/model/pane_editor.go` (Body tab read-only view + key handling), `internal/model/panes_shared.go` (`bodyPreview`/preview rendering), `internal/model/keys.go` (new toggle binding + help). A new fold component (likely `internal/model/foldview.go` or `internal/syntax`) computes fold regions over formatted JSON and renders the gutter.
- **Dependencies**: none new; reuses the existing `internal/syntax` JSON scanner and `internal/ui` styles.
- **Behavior**: read-only body views gain a cursor and respond to `up`/`down`/`space`; the request body preview is no longer hard-truncated to a fixed line count when foldable.
