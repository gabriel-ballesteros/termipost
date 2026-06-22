## 1. Version in title

- [x] 1.1 Add a `version` field to `Model` and a `version string` parameter to `New`; update `main.go` to pass its build `version`.
- [x] 1.2 Render the app name with version in `View` (`termipost vX.Y.Z`), using a `(dev)` form when the version is `dev`/empty and trimming any leading `v`.

## 2. Breadcrumb

- [x] 2.1 Add an optional `crumber` interface and a `breadcrumb()` helper that joins each stack screen's `Crumb()`/`Title()` (omitting empty crumbs) with a chevron separator.
- [x] 2.2 Have `promptScreen` and `confirmScreen` return `""` from `Crumb()`, and give the request editor a stable `Crumb()` ("Edit request").
- [x] 2.3 Render the breadcrumb in the title row next to the name; elide from the front to fit the terminal width when it overflows.

## 3. Verify

- [x] 3.1 Add a model test: with a known version, the view shows `termipost v<version>`; with `dev`, it shows `(dev)`.
- [x] 3.2 Add a model test: after navigating collection → request → editor, the view contains the breadcrumb crumbs joined by the separator, and opening a prompt does not add a crumb.
- [x] 3.3 Run `go test ./...`, `go vet ./...`, and `gofmt`.
