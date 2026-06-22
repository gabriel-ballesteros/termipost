## Why

Opening a collection shows only the list's pagination footer (e.g. `1/2`) with no request rows — the requests are invisible and unusable. The same affects every screen opened after startup (request list, environment list). Root cause: the navigation stack's `push` gives a newly opened screen its `Init` but never the terminal size, so its `bubbles/list` stays at height 0 and renders nothing but pagination. Only the initial screen, which is present when Bubble Tea sends the startup `WindowSizeMsg`, is sized correctly.

## What Changes

- When a screen is pushed onto the navigation stack, immediately propagate the current terminal size to it (as a `WindowSizeMsg`) so list-based screens are laid out and show their content right away.
- As a result, request lists, environment lists, and any future list screens display their items instead of only the pagination indicator.

## Capabilities

### New Capabilities
<!-- None. -->

### Modified Capabilities
- `tui-navigation`: Opening (pushing) a screen must size it to the current terminal so its content is visible, not just when a later resize happens.

## Impact

- Code: `internal/model/model.go` (the `push` helper on the root model). No data, API, or dependency changes.
- Fixes the request-list and environment-list screens, which are otherwise unusable until the terminal is resized.
