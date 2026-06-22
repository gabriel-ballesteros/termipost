## Context

The request editor (`internal/model/screen_request_edit.go`) models its fields with a `reqField` enum and a focus index:

```go
const ( fName reqField = iota; fMethod; fURL; fHeaders; fParams; fBody; fieldCount )
```

`updateNav` handles navigation-mode keys: Tab/↑/↓/`j`/`k` move focus, `h`/`left` and `l`/`right` cycle the method, and Enter/`i` "activates" the focused field (text → edit mode; Headers/Params → open a KV editor; Method → cycle). The Assertions row is rendered with a sentinel index (`-1`) so it is *not* in the focus cycle, and a separate `a` key pushes the assertions screen. Single-letter keys are only interpreted in navigation mode; `updateEditing` routes keystrokes to the focused input, so shortcuts are naturally inert while typing.

## Goals / Non-Goals

**Goals:**
- Assertions becomes a normal focusable field (Tab/arrows reach it, Enter opens it).
- One-keystroke first-letter shortcuts to jump to and activate each field.
- Keep edit-vs-navigation mode intact: shortcuts never fire while typing.

**Non-Goals:**
- No change to the field set, the assertions editor, or persistence.
- No new fuzzy/command palette — just direct letter keys.

## Decisions

### Make Assertions a real field
Insert `fAssertions` into the enum between `fParams` and `fBody` so the focus order matches the on-screen order (Name → Method → URL → Headers → Params → Assertions → Body):

```go
const ( fName reqField = iota; fMethod; fURL; fHeaders; fParams; fAssertions; fBody; fieldCount )
```

Render the Assertions row with its real index (not `-1`) so it highlights when focused, and add an `fAssertions` case to the Enter handler that pushes the assertions screen.

### One `activateField` helper, shared by Enter and shortcuts
Extract the per-field "activate" behavior into a helper so Enter and the letter shortcuts stay in sync:

```go
func (s *requestEditScreen) activateField(m *Model, f reqField) tea.Cmd {
    s.focus = f
    switch f {
    case fName: s.editing = true; s.name.Focus()
    case fURL:  s.editing = true; s.url.Focus()
    case fBody: s.editing = true; s.body.Focus()
    case fHeaders:    return m.push(newKVEditorScreen("Headers", s.req.Headers, ...))
    case fParams:     return m.push(newKVEditorScreen("Query params", s.req.QueryParams, ...))
    case fAssertions: return m.push(newAssertionsScreen(s.app, s.collectionID, &s.req))
    case fMethod:     // focus only; arrows/Enter cycle the value
    }
    return nil
}
```

- Enter/`i` calls `activateField(s.focus)` (keeping the special case that Enter on Method *cycles* it).
- Each shortcut calls `activateField` for its field:
  `n`→Name, `m`→Method, `u`→URL, `h`→Headers, `p`→Params, `b`→Body, `a`→Assertions.

*Why a helper:* avoids duplicating the open/edit logic in seven shortcut cases and the Enter handler, and guarantees the shortcut and Enter do the same thing.

*Method nuance:* the `m` shortcut only **focuses** Method (it does not cycle), since cycling on mere navigation would be surprising. Cycling stays on the arrows and on Enter.

### Free `h` by dropping vim method-cycling
`h`/`l` currently cycle the method, which collides with `h` = Headers. Drop them so method cycling uses only `left`/`right`:

```go
case "left":  if s.focus == fMethod { ... }
case "right": if s.focus == fMethod { ... }
```

- *Why:* `h` is the only field initial that wasn't free. Arrow keys remain the discoverable way to change the method, and `j`/`k` (which don't collide with any field initial) stay for vertical movement, so vim-style up/down navigation is preserved.
- *Alternative considered:* keep `h`/`l` and pick a different Headers shortcut — rejected because it breaks the clean "first letter of each field" model the change is about.

### Discoverability
The bottom action bar stays concise (navigation + Enter + Run/Test/Save/Back); the per-field letters are a power-user accelerator. The field labels already name each field, making the first-letter mapping easy to guess.

## Risks / Trade-offs

- **Removing `h`/`l` for the method** → mitigated: arrows still cycle it; `j`/`k` still move between fields, so vim users keep vertical navigation.
- **Shortcut firing while typing** → already prevented: letter keys reach `updateEditing` (the input) in edit mode and never `updateNav`.
- **Focus-order vs visual-order drift** → mitigated by inserting `fAssertions` at the position matching the rendered layout.
