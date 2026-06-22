## Why

The title bar shows only `termipost`, so users can't tell which version they're running, and there's no indication of where they are in the navigation stack — when several screens are nested (collection → request → assertions), pressing Esc is a guess about where "back" leads. Showing the version and a breadcrumb of the open screens makes the app more legible at a glance.

## What Changes

- Show the **app version** in the title: `termipost v1.1.0`. For local builds where the version is `dev` (or unset), show `termipost (dev)` instead of an ugly `vdev`.
- Add a **breadcrumb** to the top chrome that reflects the current navigation stack, e.g. `Collections › Requests › Edit request › Assertions`, so users can see the path they'll unwind with Esc.
- Use a chevron separator (`›`) between crumbs (a `/` is the simpler alternative). Transient overlays (the create/rename prompt, the confirm dialog) are omitted from the trail so it reflects real navigation.

## Capabilities

### New Capabilities
<!-- None. -->

### Modified Capabilities
- `tui-navigation`: the top chrome shows the app version and a breadcrumb of the open screens.

## Impact

- Code: `main.go` (pass the build `version` into the model) and `internal/model/model.go` (root `New` signature + title/breadcrumb rendering in `View`). Screens may optionally provide a short crumb label.
- No data, dependency, or behavior changes beyond the title row; purely additive chrome.
