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
- Nested sub-collections / folders (true Posting-style depth) — the data model
  stays two-level; deep nesting is a separate future proposal.
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

### 4. Focus model and key map (vim window chord)
Pane focus moves with a **vim-style window chord**: `ctrl+w` puts the workspace
into a one-shot "window" mode where the next `h/j/k/l` moves focus to the
adjacent pane (no-op at edges). Everything else stays *in-pane*: `Tab`/`Shift+Tab`
and `j/k`/arrows navigate fields/rows inside the focused pane (keeping the
editor's existing bindings untouched), and `[`/`]` switch the active tab within
the focused pane. A send key (`R` run / `T` test, as today) acts on the selected
request. When the focused pane enters text-edit mode, the workspace suspends the
window chord and quit keys (existing `updateEditing` pattern) so typing is safe.

*Why `ctrl+w` prefix over `ctrl+h/j/k/l` direct:* `ctrl+h` is ASCII backspace and
is often indistinguishable from it; the two-key `ctrl+w`-then-`h` avoids that
trap and frees `Tab` to keep its current in-pane meaning.

*Clash avoidance:* the request editor keeps single-letter field jumps
(n/m/u/h/p/b/a) and `Tab` field nav unchanged; the only new global keys are the
`ctrl+w` chord and `[`/`]`, none of which overlap existing pane content keys.

### 5. Inline key/value editor in the request pane
The KV editing logic in `screen_kv.go` (rows + name/value inputs + add/remove)
moves into a reusable component embedded in the request editor pane, rendered
under the Headers and Query tabs. Selecting the Headers/Query tab no longer
pushes a screen; rows are edited in place. This adds a nested focus level inside
the editor pane (pane → field list → KV table → row), all governed by the same
text-edit suspension rule. The assertions editor stays a pushed overlay
(scoped out), so the editor pane mixes inline tabs (Headers/Query/Body) with one
overlay-launching field (Assertions) — an accepted, documented inconsistency.

*Alternative considered:* keep KV as a pushed screen and only re-skin it.
Rejected — the header table is the most prominent part of the target layout;
leaving it as a full-screen push would make the "single screen" claim hollow.

### 6. Unsaved-edit guard on request switch
The editor pane tracks a **dirty** flag by comparing its working request against
the persisted one (it already holds a copy in `s.req` and persists on save/run).
When the tree pane requests a selection change and the editor is dirty, the
workspace pushes the existing `confirmScreen` ("Discard unsaved changes?"):
confirm loads the new request, cancel keeps the current one. A clean editor
switches with no prompt.

*Alternatives considered:* silent discard (data-loss footgun) and auto-save
(surprising writes of half-finished edits). Confirm is the least surprising.

### 7. Layout math
Top bar (method+URL+Send) = 1–2 rows; bottom action+status = existing chrome.
Remaining body splits left tree pane (fixed/ratio width, e.g. ~30%) and a right
column split vertically into editor (top) and response (bottom). The workspace
clamps to minimum sizes and recomputes on every `WindowSizeMsg`, matching the
existing responsive-layout requirement.

### 8. Quit guard (soft `q` only, never `ctrl+c`)
`ctrl+c` stays the unconditional hard-quit in `model.go` — never guarded, so it is
always a reliable escape hatch. The workspace owns `q` (replacing
`collectionListScreen` as stack bottom): when the editor is dirty, `q` pushes the
same `confirmScreen` as the request-switch guard before quitting; a clean editor
quits immediately.

*Rationale:* symmetric with the switch guard and closes the dirty-on-quit
data-loss window for free, without ever compromising the `ctrl+c` exit.

### 9. Responsive thresholds and single-pane fallback
Two layout modes only. Above the minimum multi-pane size the full three-pane grid
renders; below it the workspace shows just the focused pane full-area and the
`ctrl+w` chord becomes a pane *switcher* (same focus graph, rendered solo).
The fallback threshold is **derived from the pane minimums**, not hardcoded, so it
stays correct if the minimums change:
- Pane minimums: tree `24w`, right `40w`; editor `6h`, response `6h`.
- Multi-pane requires `width ≥ tree_min + right_min + borders ≈ 72` and
  `height ≥ chrome(4) + topbar(1) + editor_min + response_min ≈ 18`; below either,
  fall back to single-pane.
- When multi-pane: tree width = 30% clamped to `[24, 45]`; right column split
  editor/response ≈ 45/55.
These exact constants are tunable via render tests.

*Alternative considered:* an intermediate two-pane tier (tree + active right).
Rejected for now — a third layout mode triples the layout test surface for
marginal benefit; can be added later without reworking the focus model.

### 10. Collapsible 2-level tree (no new folder model)
The data model is strictly two levels (`Collection.Requests`, `domain.go`); there
are no sub-collections. The tree pane renders collections as expandable nodes and
their requests as children, with per-collection expand/collapse held in pane
state (in-memory; not persisted). The toggle key is Enter/Space on a collection
node; Enter on a request loads it. Up/down movement skips the children of
collapsed collections.

*Alternative considered:* true Posting-style arbitrary nesting. Rejected here —
it requires expanding the storage format to nested sub-collections plus changes
to collection CRUD, run-all semantics, and a migration of existing files. That is
a separate, larger proposal (a deliberate Non-Goal of this change), not a tree
rendering detail.

## Risks / Trade-offs

- **[Scope creep into overlays]** → Hard line: only tree/editor/response move
  inline; env/secrets/assertions/results stay screens. Spec deltas reflect this.
- **[Key binding clashes between panes and field shortcuts]** → Pane focus uses
  only the `ctrl+w` window chord; `Tab`/`j`/`k` and field letters stay in-pane.
  Covered by an explicit spec scenario and tests.
- **[`ctrl+w` interacting with terminal/shell defaults]** → Capture `ctrl+w`
  before pane content; verify it is not swallowed across target terminals.
- **[Nested focus in the inline KV editor]** → Multiple edit levels (pane → field
  → KV row → text input) must share one text-edit suspension flag so the window
  chord never fires mid-typing; test entering/leaving each level.
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

- Final numeric tuning of the pane minimums, width clamp, and 45/55 split —
  derived defaults set (Decision 9); confirm via render tests. (Tuning only, not
  a design fork.)

*Resolved during exploration:* pane movement = `ctrl+w` then `h/j/k/l`; tab
switching = `[`/`]`; `Tab`/`j`/`k` stay in-pane; KV editor is inlined; switching
a dirty request confirms first; `q` confirms on a dirty editor while `ctrl+c`
always hard-quits; layout is binary (full three-pane ↔ single-pane fallback);
tree is a collapsible two-level tree (folders deferred to a separate proposal).
