## Context

The root model (`internal/model/model.go`) keeps a stack of `Screen`s. Bubble Tea delivers `tea.WindowSizeMsg` only once at startup and again on each terminal resize. The root `Update` forwards that message to the *currently active* screen, and list-based screens use it to call `list.SetSize(...)`.

`push` adds a screen and calls only `Init`:

```go
func (m *Model) push(s Screen) tea.Cmd {
    m.stack = append(m.stack, s)
    m.clearStatus()
    return s.Init(m)
}
```

Because the startup `WindowSizeMsg` already fired before later screens exist, any screen opened after startup never receives a size. Its `bubbles/list` defaults to 0×0, so `View` shows only the pagination footer and no rows. The initial collection-list screen is unaffected because it is on the stack when the startup message arrives.

## Goals / Non-Goals

**Goals:**
- A screen is sized to the current terminal the moment it is opened, so its content renders immediately.
- Fix applies uniformly to all current and future pushed screens (request list, environment list, etc.).

**Non-Goals:**
- No change to resize handling, the `Screen` interface, or individual screens.
- No redesign of how screens compute their own sizes.

## Decisions

### Propagate the current size on push
After `Init`, have `push` send the newly added screen a synthetic `tea.WindowSizeMsg` carrying the model's current `width`/`height`, and batch any resulting command with the `Init` command:

```go
func (m *Model) push(s Screen) tea.Cmd {
    m.stack = append(m.stack, s)
    m.clearStatus()
    cmd := s.Init(m)
    if m.width > 0 { // size known (startup WindowSizeMsg already received)
        sizeCmd := s.Update(m, tea.WindowSizeMsg{Width: m.width, Height: m.height})
        cmd = tea.Batch(cmd, sizeCmd)
    }
    return cmd
}
```

- *Why here:* `push` is the single choke point through which every screen is opened, so one change fixes every screen, including future ones. Screens already handle `WindowSizeMsg` via `list.SetSize` / `viewport` resizing, so reusing that path keeps each screen's sizing logic in one place.
- *Why guard on `m.width > 0`:* before the startup `WindowSizeMsg`, dimensions are unknown (0). Sending a 0×0 size would be pointless; the real startup message will size the active screen normally. The initial screen is pushed before any size is known, so the guard avoids a meaningless 0×0 pass.
- *Alternatives considered:*
  - *Re-emit `WindowSizeMsg` from each screen's `Init`* — requires touching every screen and is easy to forget for new screens. Rejected.
  - *Size lists inside `View` each frame* — `SetSize` in render is a side effect in a read path and must be repeated in every screen. Rejected.

## Risks / Trade-offs

- **Double sizing of the active screen on a real resize** → harmless: `SetSize`/viewport resizing is idempotent, and `push` only runs on open, not on resize.
- **A screen whose `Update` returns a command for `WindowSizeMsg`** → handled by batching that command with `Init`'s, so no command is dropped.
