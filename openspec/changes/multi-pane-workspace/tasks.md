## 1. Pane framework

- [ ] 1.1 Define a `pane` interface in `internal/model` (`Update(m, msg, focused) tea.Cmd`, `View(m, w, h, focused) string`, `HelpBindings() []key.Binding`, `Title() string`)
- [ ] 1.2 Add focused/unfocused pane box styles (border + title) in `internal/ui`
- [ ] 1.3 Add a `workspaceScreen` skeleton implementing `Screen`, owning a slice of panes and a `focus` index, rendering placeholder panes in a 3-pane grid
- [ ] 1.4 Implement layout math: top method+URL+Send bar, left tree column (~30% clamped to [24,45] cols), right column split editor/response ≈45/55 with ≥6 rows each; compute each pane's `w,h` from terminal size and recompute on `WindowSizeMsg`
- [ ] 1.7 Implement the binary fallback: when panes can't meet their minimums (roughly width<100 or height<24), render only the focused pane full-area and make the `ctrl+w` chord switch which single pane is shown; restore the full layout (same pane focused) when space returns
- [ ] 1.5 Implement pane focus: `ctrl+w` window chord + `h/j/k/l` directional move (no-op at edges); route key msgs only to the focused pane; broadcast lifecycle/async msgs (spinner ticks, send results) to all panes; suspend the window chord + quit keys while a pane is in text-edit mode
- [ ] 1.6 Add the `ctrl+w` window chord, `[`/`]` tab-switch, and send bindings to `keys.go`; keep `Tab`/`j`/`k` and field letters (n/m/u/h/p/b/a) as in-pane keys with no clash

## 2. Response pane

- [ ] 2.1 Port `responseScreen` (viewport, `prettyBody`, `renderHeaders`, `looksLikeJSON`) into a response pane rendering at the given `w,h`
- [ ] 2.2 Add Body/Headers tabs switched with `[`/`]` and an active-tab indicator
- [ ] 2.3 Show a status badge and an empty state when the selected request has no last response

## 3. Request editor pane

- [ ] 3.1 Port `requestEditScreen` field model (`reqField`, nav, edit mode, `syncFromInputs`, `persist`) into an editor pane rendering at the given `w,h`
- [ ] 3.2 Group fields into Headers/Query/Body (and Info as available) tabs switched with `[`/`]`, with an active-tab indicator
- [ ] 3.3 Inline the KV editor: move `screen_kv.go` row/add/remove logic into a reusable component embedded under the Headers and Query tabs (no pushed screen); keep Assertions as a pushed overlay
- [ ] 3.4 Track a dirty flag (working request vs persisted); reflect the selected request; persist edits via the existing save path

## 4. Tree pane and top bar

- [ ] 4.1 Build a collapsible 2-level tree pane: collections as expandable nodes (▸/▾) with their requests as children, per-collection expand state in pane memory; Enter/Space toggles a collection, Enter on a request loads it, up/down skips collapsed children; reuse collection/request item + CRUD logic
- [ ] 4.2 Guard request switching: when the editor is dirty, push the existing `confirmScreen` ("Discard unsaved changes?") — confirm loads the new request, cancel keeps the current one; clean editor switches with no prompt
- [ ] 4.3 Wire collection/request new/rename/delete to launch the existing prompt/confirm overlays from the pane
- [ ] 4.4 Render the top method+URL+Send bar bound to the selected request; trigger send from the workspace send key

## 5. Wiring, navigation, chrome

- [ ] 5.1 Switch `New` to push `workspaceScreen` as the initial screen
- [ ] 5.2 Make the action bar reflect the focused pane's bindings plus global bindings, updating on focus change
- [ ] 5.3 Launch secondary flows (environments, secrets, assertions editor, run-all results) as screens/overlays from the workspace; Esc returns to the workspace
- [ ] 5.4 Update the breadcrumb to anchor on the workspace and reflect any open secondary screen (overlays excluded)
- [ ] 5.5 Guard soft-quit: when the editor is dirty, `q` pushes the discard-changes `confirmScreen` before quitting; keep `ctrl+c` an unconditional hard-quit
- [ ] 5.6 Remove standalone screen wrappers that are no longer reachable once their logic lives in panes

## 6. Tests and verification

- [ ] 6.1 Update/extend render + smoke tests in `internal/model` for the composite layout (all panes visible, sized on first render)
- [ ] 6.2 Test focus transitions: `ctrl+w`+`h/j/k/l` moves between panes, no-op at edges, key msgs only hit the focused pane while async msgs reach all panes
- [ ] 6.3 Test reflow on resize, the single-pane fallback below threshold (ctrl+w switches the shown pane), and restoration of the full layout when space returns
- [ ] 6.8 Test the soft-quit guard: `q` on a dirty editor prompts before exit; `q` on a clean editor quits; `ctrl+c` always quits without a prompt
- [ ] 6.4 Test in-pane keys (`Tab`/`j`/`k`, field letters, `[`/`]`) do not trigger pane focus changes
- [ ] 6.5 Test the unsaved-edit guard: dirty switch prompts, confirm discards + loads, cancel keeps; clean switch is silent
- [ ] 6.6 Test inline KV editing (add/edit/remove header rows) persists with the request
- [ ] 6.9 Test the tree pane: toggle expand/collapse, up/down skips collapsed requests, Enter toggles a collection vs loads a request
- [ ] 6.7 Run `make fmt`, `make vet`, `make test`
