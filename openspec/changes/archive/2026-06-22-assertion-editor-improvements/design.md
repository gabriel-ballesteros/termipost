## Context

The assertion editor (`assertionEditScreen` in `internal/model/screen_assertions.go`) has four fields:

```go
const ( aKind assertionField = iota; aOp; aTarget; aExpected; aFieldCount )
```

`updateNav` moves focus with `(s.focus ± 1) % aFieldCount`, always cycling through all four. But `usesTarget(a)` hides the Target field for status-code, latency, and non-JSON-path body assertions, and `View` only renders Target when `usesTarget` is true. The mismatch is the bug: focus stops on the hidden `aTarget` index, costing an extra ↓.

Operators come from `opsFor(kind)`; status code currently returns only `OpEquals`. Evaluation lives in `runner.evaluate`, where `AssertStatusCode` does a hard `resp.StatusCode == want`.

## Goals / Non-Goals

**Goals:**
- Navigation stops only on visible fields, for every assertion kind.
- Changing kind never leaves focus on a now-hidden field.
- A `not_equals` operator for status-code assertions, surfaced in the editor, evaluator, and summary.

**Non-Goals:**
- No change to other assertion kinds' operators or to the field set.
- No data migration; `OpEquals` remains the default operator.

## Decisions

### Navigate a dynamic list of visible fields
Add a helper that returns the visible fields in order:

```go
func (s *assertionEditScreen) visibleFields() []assertionField {
    if usesTarget(s.a) {
        return []assertionField{aKind, aOp, aTarget, aExpected}
    }
    return []assertionField{aKind, aOp, aExpected}
}
```

Navigation finds the current field's position in this slice and steps within it (wrapping), instead of using `% aFieldCount`:

```go
case "tab", "down", "j":   s.focus = step(s.visibleFields(), s.focus, +1)
case "shift+tab", "up","k": s.focus = step(s.visibleFields(), s.focus, -1)
```

`step` locates `focus` in the slice and returns the neighbor (wrapping); if `focus` isn't in the slice (it just became hidden), it returns the first visible field.

- *Why a dynamic list:* it is the single source of truth for "what's on screen," so navigation and rendering can't drift. Skipping by index math (`if next == aTarget && !usesTarget { next++ }`) would special-case one field and break if more conditional fields are added.

### Keep focus visible after a kind change
`cycle` changes Kind (and resets Op), which can hide Target. After cycling, clamp focus to a visible field:

```go
if !contains(s.visibleFields(), s.focus) {
    s.focus = aExpected // always visible
}
```

Expected is always present, so it is a safe landing spot. (In practice the only hideable field is Target, so this only triggers when leaving a header/json-path assertion while focused on Target.)

### Add `not_equals` for status code
- New operator constant `OpNotEquals MatchOp = "not_equals"` in `internal/domain`.
- `opsFor(AssertStatusCode)` returns `[]MatchOp{OpEquals, OpNotEquals}` so the editor can cycle to it.
- `runner.evaluate` for `AssertStatusCode` compares with the operator:
  ```go
  pass := resp.StatusCode == want
  if a.Op == domain.OpNotEquals { pass = resp.StatusCode != want }
  ```
  with detail text reflecting `==` / `!=`.
- `describeAssertion` renders `status code == N` or `status code != N` based on the operator.

- *Why scoped to status code:* the request only asks for status-code not-equals, and equality semantics for headers/body already have `contains`/`regex` alternatives. Keeping it status-code-only avoids defining not-equals semantics for every kind.

## Risks / Trade-offs

- **Existing status-code assertions** have an empty or `equals` operator → both already mean "equals"; the evaluator's default branch keeps treating them as equals, so nothing breaks.
- **`step` receiving a focus value not in the visible list** → handled by falling back to the first visible field, which also covers the post-kind-change case.
