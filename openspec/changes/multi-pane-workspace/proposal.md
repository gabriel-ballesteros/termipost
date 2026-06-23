## Why

Today every view in termipost is a full-screen step on a navigation stack: you
see the collection list, *then* a request, *then* its response — each replacing
the last. This hides context the user needs at a glance (what request am I
editing? what did it just return?) and forces constant Esc/Enter churn. A
single-screen, multi-pane workspace (in the spirit of Posting) keeps the
collection tree, the request editor, and the response visible together, so the
common loop — pick a request, tweak it, send, read the response — happens
without leaving the screen.

## What Changes

- Introduce a **composite workspace screen** as the primary view: a top
  method+URL+Send bar, a left **collection/request tree** pane, a top-right
  **request editor** pane (tabbed: Headers, Query, Body, Auth/Info as available),
  and a bottom-right **response** pane (tabbed: Body, Headers, Cookies/Trace as
  available), all on one screen and reflowing to terminal size.
- Add a **pane focus model**: exactly one pane is focused at a time; `Tab` /
  `Shift+Tab` cycle panes and direct directional jump keys move focus between
  panes. Only the focused pane consumes its content keys; the action bar reflects
  the focused pane.
- Replace the collection-list → requests → request-editor → response **stack
  navigation** for this core loop with in-place pane focus. **BREAKING** for the
  navigation model: the primary loop no longer pushes/pops full screens.
- Keep **secondary flows** (environments, secrets, assertions editor, run-all
  results, new/rename/delete prompts) as overlays/screens launched from the
  workspace; they are out of scope for inlining here.
- Preserve **keyboard-only** operation and the persistent contextual action bar
  end to end.

## Capabilities

### New Capabilities
- `workspace-layout`: the composite single-screen layout — pane composition,
  pane focus and movement, per-pane tabs, and reflow/sizing of the multi-pane
  grid to the terminal.

### Modified Capabilities
- `tui-navigation`: the core request/response loop moves from full-screen stack
  navigation to in-place pane focus; back/quit, breadcrumb, and the contextual
  action bar adapt to the workspace and its focused pane.

## Impact

- **Code**: `internal/model` — new workspace screen and pane components;
  refactor of `model.go` View/Update to host panes; `screen_collections.go`,
  `screen_requests.go`, `screen_request_edit.go`, `screen_response.go` become
  pane content rather than standalone screens; `keys.go` gains pane-focus
  bindings; `internal/ui` styling for focused/unfocused pane borders.
- **Behavior**: navigation muscle memory changes for the core loop; existing
  per-field jump keys must coexist with pane-focus keys (no clashes).
- **Specs**: new `workspace-layout`, delta on `tui-navigation`. No change to
  `data-persistence`, `collection-management`, `environment-management`,
  `api-testing`, or `request-management` requirements (storage and request
  semantics unchanged).
- **Tests**: render/smoke tests in `internal/model` updated for the composite
  layout and focus transitions.
