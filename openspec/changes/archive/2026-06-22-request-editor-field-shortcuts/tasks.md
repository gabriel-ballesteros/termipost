## 1. Implementation

- [x] 1.1 Add `fAssertions` to the `reqField` enum between `fParams` and `fBody`.
- [x] 1.2 Extract an `activateField(m, f)` helper that sets focus and performs each field's action (text → edit mode; Headers/Params/Assertions → open editor; Method → focus only).
- [x] 1.3 Route Enter/`i` through `activateField` (keep Enter-on-Method cycling), and add the `fAssertions` case.
- [x] 1.4 Render the Assertions row with its real `fAssertions` index so it is focusable/highlightable.
- [x] 1.5 Add navigation-mode shortcuts `n`/`m`/`u`/`h`/`p`/`b`/`a` that call `activateField` for their fields.
- [x] 1.6 Drop `h`/`l` from method cycling (keep `left`/`right` arrows).

## 2. Verify

- [x] 2.1 Add a model test: in the editor, Tab through fields and confirm Assertions is in the cycle and Enter on it opens the assertions screen.
- [x] 2.2 Add a model test: pressing `a` and `p`/`h` opens the assertions and KV editor screens, and `n`/`u`/`b` enter edit mode (shortcuts inert while editing).
- [x] 2.3 Manually confirm arrow keys still cycle the method and `h` now jumps to Headers.
- [x] 2.4 Run `go test ./...`, `go vet ./...`, and `gofmt`.
