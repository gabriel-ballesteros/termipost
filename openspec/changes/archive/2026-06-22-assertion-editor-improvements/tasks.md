## 1. Navigation fix

- [x] 1.1 Add a `visibleFields()` helper and a `step` function to `screen_assertions.go`, and use them for ↑/↓/Tab/Shift+Tab navigation instead of `% aFieldCount`.
- [x] 1.2 After `cycle` changes the Kind, clamp focus to a visible field (e.g. `aExpected`) when the current field became hidden.

## 2. not-equals operator

- [x] 2.1 Add `OpNotEquals` to `internal/domain` and include it in `opsFor(AssertStatusCode)`.
- [x] 2.2 Update `runner.evaluate` for `AssertStatusCode` to honor equals vs not-equals, with matching detail text.
- [x] 2.3 Update `describeAssertion` to render `status code == N` or `status code != N`.

## 3. Action-bar labels

- [x] 3.1 In `internal/model/keys.go`, change the run/test help labels to `shift+r` and `shift+t` (keys unchanged).
- [x] 3.2 In `internal/model/screen_collections.go`, update the `keyRunCol` "run all" help label to `shift+r`.

## 4. Verify

- [x] 4.1 Add a model test: for a latency (target-hidden) assertion, one ↓ from Operator focuses the expected/max-ms field (not a hidden field).
- [x] 4.2 Add a runner test: a `status_code` `not_equals` assertion passes when the status differs and fails when it matches.
- [x] 4.3 Run `go test ./...`, `go vet ./...`, and `gofmt`.
