## Why

On the request editor, "Assertions" is the only row that cannot be reached with the arrow keys — it can only be opened with the `a` shortcut. That's inconsistent: every other row is part of the Tab/arrow focus cycle. At the same time, only Assertions has a quick single-letter shortcut (`a`); the other rows (Name, Method, URL, Headers, Params, Body) require tabbing to them. Since each row's first letter is otherwise unused, we can make the editor both consistent and faster.

## What Changes

- Make **Assertions** a focusable row in the editor: it joins the Tab/↑/↓ focus cycle and opens with Enter, like every other row.
- Add **first-letter shortcuts** that jump to a row and activate it in one keystroke, matching how `a` already opens Assertions:
  - `n` Name, `m` Method, `u` URL, `h` Headers, `p` Params, `b` Body, `a` Assertions.
  - Activating means: text rows (Name/URL/Body) enter edit mode; Headers/Params/Assertions open their editors; Method gets focus so ←/→ can cycle it.
- Free up `h` for the Headers shortcut by dropping the vim-style `h`/`l` method cycling; the ←/→ arrows still cycle the method. (The user's note assumed all first letters were free; `h` was the one exception.)

## Capabilities

### New Capabilities
<!-- None. -->

### Modified Capabilities
- `request-management`: The request editor gains a focusable Assertions row and first-letter shortcuts to jump to and activate each field.

## Impact

- Code: `internal/model/screen_request_edit.go` (the `reqField` enum, `updateNav`, the field rendering, and help bindings). No data, API, or dependency changes.
- Minor keybinding change: `h`/`l` no longer cycle the method (arrows still do).
