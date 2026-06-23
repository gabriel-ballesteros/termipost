## Context

termipost's TUI (`internal/model`) is a root `Model` that owns shared `*App`
state and a `[]Screen` navigation stack (`model.go`). Each `Screen` is a
full-screen view (`collectionListScreen`, `requestListScreen`,
`requestEditScreen`, `responseScreen`, plus `prompt`/`confirm` overlays) that
implements `Init/Update/View/Title/HelpBindings`. The root `View` draws chrome
(title+breadcrumb, body, status, help bar) and delegates the body to the top
screen; the root `Update` forwards every message to the top screen.

This change introduces a composite **workspace** screen that hosts several panes
at once for the core request/response loop, while keeping the existing
`Screen`/overlay machinery for secondary flows. Constraint: keyboard-only,
Bubble Tea + Bubbles (`list`, `textinput`, `viewport`) + Lip Gloss, no new
external deps.

## Goals / Non-Goals

**Goals:**
- One screen showing tree + request editor + response + method/URL/Send bar.
- Clear single-pane focus with `Tab`/`Shift+Tab` cycling and directional jumps.
- Reuse the existing list/textinput/viewport-backed logic that already lives in
  the per-screen files rather than rewriting request/response rendering.
- Preserve the contextual action bar (now keyed off the focused pane) and a
  sensible breadcrumb.

**Non-Goals:**
- Inlining environments, secrets, the assertions editor, or run-all results —
  they stay as launched screens/overlays.
- Changing storage, request semantics, or assertion behavior (no delta to
  `data-persistence`, `request-management`, `api-testing`,
  `collection-management`, `environment-management`).
- Mouse support.

## Decisions

### 1. Introduce a `workspaceScreen` that composes panes, not a new framework
Add one `Screen` implementation (`workspaceScreen`) that becomes the initial
screen pushed in `New`. It owns a small set of **pane** values and a `focus`
index. This keeps the root `Model`/stack untouched: the workspace is just the
bottom-of-stack screen, and secondary flows still `push`/`pop` on top of it.

*Alternative considered:* make the root `Model` itself multi-pane (drop the
stack). Rejected — it would force every overlay (prompts, confirm, env, secrets)
to be reworked at once, exploding scope and breaking the BREAKING surface beyond
the core loop.

### 2. A `pane` interface mirroring `Screen`, minus stack concerns
Define `pane { Update(m, msg, focused) tea.Cmd; View(m, w, h, focused) string;
HelpBindings() []key.Binding; Title() string }`. The workspace routes content
keys only to the focused pane and asks each pane to render at an explicit width
and height with a focused flag (so it can draw a highlighted border/title).

*Rationale:* explicit `w,h` per pane is required because Bubble Tea only emits
`WindowSizeMsg` globally; the workspace computes each pane's box from the current
terminal size and passes it down (same problem `push` already works around in
`model.go`).

### 3. Reuse existing screen logic as pane bodies
- Tree pane: a `list.Model` showing collections and their requests (start with a
  flat/grouped list reusing `collectionListScreen`/`requestListScreen` item and
  CRUD logic; nested-tree rendering is a presentation refinement, not a blocker).
- Request editor pane: lift the field model, `reqField` navigation, edit mode,
  `syncFromInputs`/`persist` from `requestEditScreen`; group fields into tabs
  (Headers/Query/Body/…) for the tabbed-section requirement.
- Response pane: reuse `responseScreen`'s `viewport`, `prettyBody`,
  `renderHeaders`, with tabs Body/Headers.
Keep these as the same code paths so existing unit/smoke tests largely hold.

### 4. Focus model and key map
Global (workspace-level) keys, active whenever no pane is in text-edit mode:
`Tab` next pane, `Shift+Tab` prev pane (cyclic, fixed order tree → editor →
response), and directional jumps `ctrl+h/j/k/l` mapped to the pane grid. A
send key (e.g. `ctrl+j`/`R`) triggers send from anywhere. When the focused pane
enters text-edit mode, the workspace suspends focus/quit keys (existing
`updateEditing` pattern) so typing is safe.

*Clash avoidance:* the request editor already uses single-letter field jumps
(n/m/u/h/p/b/a). Pane focus uses `Tab`/`Shift+Tab` + `ctrl`-modified keys, which
do not overlap those letters — satisfying the "no clash" scenario.

### 5. Layout math
Top bar (method+URL+Send) = 1–2 rows; bottom action+status = existing chrome.
Remaining body splits left tree pane (fixed/ratio width, e.g. ~30%) and a right
column split vertically into editor (top) and response (bottom). The workspace
clamps to minimum sizes and recomputes on every `WindowSizeMsg`, matching the
existing responsive-layout requirement.

## Risks / Trade-offs

- **[Scope creep into overlays]** → Hard line: only tree/editor/response move
  inline; env/secrets/assertions/results stay screens. Spec deltas reflect this.
- **[Key binding clashes between panes and field shortcuts]** → Reserve
  `Tab`/`ctrl`-modified keys for focus; covered by an explicit spec scenario and
  tests.
- **[Small terminals can't fit three panes]** → Define minimum sizes; below them,
  fall back to focusing one pane full-area (degraded but usable). Validate with
  render tests at small sizes.
- **[Test churn]** existing smoke/render tests assume full-screen navigation →
  Update them alongside; reusing pane logic from current screens limits the blast
  radius.
- **[Muscle-memory break for the core loop]** (BREAKING) → Action bar always
  shows the new focus keys; breadcrumb still anchors the user.

## Migration Plan

1. Add `pane` interface + `workspaceScreen` skeleton with layout math and focus,
   rendering placeholder panes.
2. Port response pane (simplest, read-only viewport), then request editor pane,
   then tree pane — each reusing existing screen logic.
3. Switch `New` to push `workspaceScreen`; keep secondary flows launching as
   screens/overlays from it.
4. Update render/smoke tests; keep the old screen files only as far as their
   logic is reused, removing dead standalone wrappers.

Rollback: revert to pushing `newCollectionListScreen` in `New`; the stack and
secondary screens are untouched, so reverting the initial screen restores the
old flow.

## Open Questions

- Exact left-pane width ratio and minimum terminal size thresholds — settle
  during implementation via render tests.
- Whether the tree pane renders true nested collapsing (Posting-style) now or in
  a follow-up; baseline is a grouped list.
