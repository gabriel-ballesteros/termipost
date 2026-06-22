## 1. Fix

- [x] 1.1 In `internal/model/model.go`, update `push` to send the newly pushed screen a `tea.WindowSizeMsg` with the model's current width/height (guarded on `m.width > 0`), batching the result with the `Init` command.

## 2. Verify

- [x] 2.1 Add a model test: after sizing the root, push the request-list screen for a collection with ≥2 requests and assert the rendered view contains the request names (not only a pagination indicator).
- [x] 2.2 Manually confirm opening a collection shows its request rows, and opening Environments shows environment rows, without needing to resize the terminal.
- [x] 2.3 Run `go test ./...`, `go vet ./...`, and `gofmt`.
