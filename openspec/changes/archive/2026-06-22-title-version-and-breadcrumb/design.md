## Context

The root model (`internal/model/model.go`) renders the top chrome in `View`:

```go
title := ui.Title.Render("termipost") + "  " + ui.Subtle.Render(s.Title())
// + env indicator …
```

It only shows the active screen's `Title()`. The model already holds `stack []Screen` (each `Screen` has `Title()`), so the breadcrumb is derivable from existing state. The build version lives in `main.version` (package `main`), which `internal/model` cannot import.

## Goals / Non-Goals

**Goals:**
- Title shows `termipost vX.Y.Z`, with a sensible `(dev)` form for local builds.
- A breadcrumb of the open screens, omitting transient overlays.
- Everything rendered in one place (the root `View`); screens stay untouched except an optional short-label hook.

**Non-Goals:**
- No per-screen layout changes, no new navigation behavior, no config for the separator.

## Decisions

### Thread the version into the model
Add a `version` field to `Model` and a parameter to `New`:

```go
func New(app *App, version string, loadErrs []error) *Model
```

`main.go` passes its `version` var. Rendering:

```go
name := "termipost"
if version == "" || version == "dev" {
    name += " (dev)"
} else {
    name += " v" + strings.TrimPrefix(version, "v")
}
```

- *Why a parameter, not a package var:* keeps `internal/model` free of globals and makes it testable with an explicit version. `TrimPrefix` avoids `vv1.1.0` if a tag-like value is passed.

### Build the breadcrumb from the stack
Add an optional crumb hook so screens can offer a short, stable label (or opt out):

```go
type crumber interface{ Crumb() string }

func (m *Model) breadcrumb() string {
    var parts []string
    for _, s := range m.stack {
        label := s.Title()
        if c, ok := s.(crumber); ok {
            label = c.Crumb() // "" means omit
        }
        if label != "" {
            parts = append(parts, label)
        }
    }
    return strings.Join(parts, sep)
}
```

- Screens with no `Crumb()` fall back to `Title()` — so the breadcrumb works for every existing screen with zero changes.
- `promptScreen` and `confirmScreen` implement `Crumb() string { return "" }` so overlays don't appear in the trail.
- The request editor (whose `Title()` changes while editing) implements `Crumb() { return "Edit request" }` for a stable label.

- *Why an optional interface:* avoids adding a method to every screen and to the `Screen` interface; only the few screens that need special behavior implement it.

### Separator
Use a chevron with subtle styling: `sep := ui.Subtle.Render(" › ")` (U+203A). A `/` is a drop-in alternative if preferred. Defining it as a single `const`/var keeps it consistent and easy to change.

### Overflow handling
The title line (name + breadcrumb + env) can exceed narrow widths. Render the breadcrumb into the remaining width after the name and env, and if it still overflows, keep the **last** crumbs (the deepest, most relevant) and elide the front with `…`. Implemented with a small width-aware trim using `lipgloss.Width`.

## Risks / Trade-offs

- **Long breadcrumbs on narrow terminals** → mitigated by eliding from the front and keeping the active/deepest crumbs visible.
- **Version not injected in some build paths** (e.g. `go run`) → handled by the `dev` fallback, so it never shows a broken `v`.
- **A new screen forgetting to opt out of the breadcrumb** → harmless: it simply shows its `Title()`; only overlays need the empty `Crumb()`.
