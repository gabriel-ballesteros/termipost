## 1. Pane framework

- [ ] 1.1 Define a `pane` interface in `internal/model` (`Update(m, msg, focused) tea.Cmd`, `View(m, w, h, focused) string`, `HelpBindings() []key.Binding`, `Title() string`)
- [ ] 1.2 Add focused/unfocused pane box styles (border + title) in `internal/ui`
- [ ] 1.3 Add a `workspaceScreen` skeleton implementing `Screen`, owning a slice of panes and a `focus` index, rendering placeholder panes in a 3-pane grid
- [ ] 1.4 Implement layout math: top method+URL+Send bar, left tree column, right column split into editor (top) + response (bottom); compute each pane's `w,h` from terminal size and recompute on `WindowSizeMsg`
- [ ] 1.5 Implement focus: `Tab`/`Shift+Tab` cyclic order (tree → editor → response) and `ctrl+h/j/k/l` directional jumps; route content keys only to the focused pane; suspend focus/quit keys while a pane is in text-edit mode
- [ ] 1.6 Add pane-focus + send bindings to `keys.go`; ensure no clash with request-editor field shortcuts (n/m/u/h/p/b/a)

## 2. Response pane

- [ ] 2.1 Port `responseScreen` (viewport, `prettyBody`, `renderHeaders`, `looksLikeJSON`) into a response pane rendering at the given `w,h`
- [ ] 2.2 Add Body/Headers tabs with a focused-pane tab-switch key and an active-tab indicator
- [ ] 2.3 Show a status badge and an empty state when the selected request has no last response

## 3. Request editor pane

- [ ] 3.1 Port `requestEditScreen` field model (`reqField`, nav, edit mode, `syncFromInputs`, `persist`) into an editor pane rendering at the given `w,h`
- [ ] 3.2 Group fields into Headers/Query/Body (and Info as available) tabs with a tab-switch key and active-tab indicator
- [ ] 3.3 Reflect the currently selected request; persist edits via the existing save path

## 4. Tree pane and top bar

- [ ] 4.1 Build a tree pane `list.Model` showing collections and their requests, reusing collection/request item + CRUD logic; selecting a request updates the editor and response panes
- [ ] 4.2 Wire collection/request new/rename/delete to launch the existing prompt/confirm overlays from the pane
- [ ] 4.3 Render the top method+URL+Send bar bound to the selected request; trigger send from the workspace send key

## 5. Wiring, navigation, chrome

- [ ] 5.1 Switch `New` to push `workspaceScreen` as the initial screen
- [ ] 5.2 Make the action bar reflect the focused pane's bindings plus global bindings, updating on focus change
- [ ] 5.3 Launch secondary flows (environments, secrets, assertions editor, run-all results) as screens/overlays from the workspace; Esc returns to the workspace
- [ ] 5.4 Update the breadcrumb to anchor on the workspace and reflect any open secondary screen (overlays excluded)
- [ ] 5.5 Remove standalone screen wrappers that are no longer reachable once their logic lives in panes

## 6. Tests and verification

- [ ] 6.1 Update/extend render + smoke tests in `internal/model` for the composite layout (all panes visible, sized on first render)
- [ ] 6.2 Test focus transitions: Tab/Shift+Tab cycle, directional jumps, no-op at edges, content keys only hit the focused pane
- [ ] 6.3 Test reflow on resize and degraded fallback at small terminal sizes
- [ ] 6.4 Test pane-focus keys do not clash with request-editor field shortcuts
- [ ] 6.5 Run `make fmt`, `make vet`, `make test`
