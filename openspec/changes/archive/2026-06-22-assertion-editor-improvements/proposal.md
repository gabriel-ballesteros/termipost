## Why

The assertion editor has a navigation bug: it always cycles focus through all four fields (Kind, Operator, Target, Expected), but the Target field is hidden for some assertion kinds (status code, latency, and non-JSON-path body). So when Target is hidden, moving down from Operator lands on the invisible Target field, and it takes an extra ↓ to reach the next visible field (e.g. "Max ms" for latency). Navigation should only stop on fields that are actually shown.

Separately, the status-code assertion only supports an equals check. A common need is the opposite — asserting the response status is *not* a given code (e.g. "must not be 500").

## What Changes

- Fix assertion-editor navigation so ↑/↓/Tab only move between **visible** fields, skipping Target when it is hidden. Focus is also kept on a visible field when changing Kind hides the current field.
- Add a **not-equals** operator to the `status_code` assertion kind: it passes when the response status differs from the expected value. The editor offers it alongside equals, the evaluator checks it, and it is rendered in the assertion summary (e.g. `status code != 500`).
- Make the action bar clearer: display the run and test shortcuts as `shift+r` and `shift+t` instead of `R` and `T` (the binding is unchanged — the uppercase key already requires Shift; the label just makes that explicit). The collection "run all" label is updated to `shift+r` for consistency.

## Capabilities

### New Capabilities
<!-- None. -->

### Modified Capabilities
- `api-testing`: status-code assertions gain a not-equals operator, and the assertion editor only navigates between visible fields.

## Impact

- Code: `internal/model/screen_assertions.go` (field navigation, operator list, summary text) and `internal/runner/runner.go` (status-code evaluation); a new `OpNotEquals` operator in `internal/domain`.
- Code: `internal/model/keys.go` (run/test help labels) and `internal/model/screen_collections.go` (run-all label) for the action-bar wording. This is a label-only change with no requirement impact.
- No data migration: existing assertions keep working; the new operator is opt-in per assertion.
