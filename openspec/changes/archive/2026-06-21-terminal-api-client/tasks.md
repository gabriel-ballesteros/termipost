## 1. Project scaffold

- [x] 1.1 Initialize Go module (`go mod init`) and create the `internal/` package layout (`domain`, `store`, `httpclient`, `runner`, `vars`, `ui`, `model`)
- [x] 1.2 Add dependencies: `bubbletea`, `bubbles`, `lipgloss` (`go get`) and commit `go.mod`/`go.sum`
- [x] 1.3 Create `main.go` that initializes the data directory, loads data via `store`, and starts the Bubble Tea program with the root model
- [x] 1.4 Add a `README.md` with build/run instructions and a `Makefile`/run target

## 2. Domain model

- [x] 2.1 Define `domain` types: `Collection`, `Request` (method, url, headers, query params, body) with nested `Assertion[]`, with stable `id` and display `name`
- [x] 2.2 Define `domain` types: `Assertion` (status_code, header, body, latency) plus run/result types
- [x] 2.3 Define `domain` types: `Environment` (id, name, vars map) and the global `Secrets` map
- [x] 2.4 Add ID generation helper (slug + short random suffix) and HTTP method enum/validation

## 3. Persistence (data-persistence)

- [x] 3.1 Implement storage location resolution (`os.UserConfigDir()/termipost/`, fallback `~/.termipost/`) and directory creation on startup
- [x] 3.2 Implement load: `config.json`, `collections/<id>.json` (with nested assertions), `environments/<id>.json`, `secrets.json`, treating missing files as empty
- [x] 3.3 Implement save with `json.MarshalIndent` and atomic write-to-temp-then-rename
- [x] 3.4 Generate `.gitignore` excluding `secrets.json` on data-dir init
- [x] 3.5 Handle malformed JSON: report the error, continue running, and never overwrite the bad file
- [x] 3.6 Unit-test load/save round-trip, missing-file, malformed-file, and gitignore generation

## 4. TUI foundation (tui-navigation)

- [x] 4.1 Implement root model with shared state (loaded data, active environment), terminal size handling (resize/reflow), and a status/error line
- [x] 4.2 Implement screen enum, sub-model interface, and navigation stack with Esc "back" and global quit (q / Ctrl+C)
- [x] 4.3 Implement `bubbles/key` keymaps per screen and render the contextual action/help bar via `bubbles/help`
- [x] 4.4 Implement edit-vs-navigation mode so global single-key shortcuts are suspended while a text field is focused
- [x] 4.5 Define shared Lip Gloss styles in `ui`

## 5. Collection management (collection-management)

- [x] 5.1 Build the collection list screen (browse, empty state) using `bubbles/list` with arrow + j/k navigation
- [x] 5.2 Implement create collection with unique/non-empty name validation
- [x] 5.3 Implement rename collection with duplicate-name rejection
- [x] 5.4 Implement delete collection with confirmation prompt and cascade removal of its file
- [x] 5.5 Persist all collection mutations via `store`

## 6. Environments & secrets (environment-management)

- [x] 6.1 Build the environment list/editor (create, rename, delete, edit variable key/value pairs) with duplicate-name rejection
- [x] 6.2 Implement active-environment selection persisted in `config.json` and reflected in the UI
- [x] 6.3 Build the global secrets editor with masked display and an explicit reveal action
- [x] 6.4 Implement `vars` resolution: `{{name}}` → active environment then secrets; leave unresolved + warn; track secret-sourced values for masking
- [x] 6.5 Unit-test resolution precedence, unresolved handling, and secret masking

## 7. Request management (request-management)

- [x] 7.1 Build the request list screen within a collection (navigate, create, delete with confirmation)
- [x] 7.2 Build the request editor: name, method selector, URL, headers (key/value), query params, body (`textarea`)
- [x] 7.3 Apply `{{variable}}` substitution across URL, headers, query params, and body before sending using `vars`
- [x] 7.4 Implement non-blocking send in a `tea.Cmd` using stdlib `net/http` with a spinner while in flight
- [x] 7.5 Build the response view: status, headers, elapsed time, scrollable body in a `viewport` with JSON pretty-printing; mask secret-sourced values in any resolved-request preview
- [x] 7.6 Handle network/connection errors with a clear message and no crash; persist request edits

## 8. API testing (api-testing)

- [x] 8.1 Build the assertions panel on a request to add/edit/remove assertions (status_code, header equals/contains/regex, body contains/equals/json-path, latency max)
- [x] 8.2 Implement the assertion evaluator producing per-assertion pass/fail with expected vs actual
- [x] 8.3 Implement `runner` to run a single request as a test (send + evaluate) and report results, including request-error failures
- [x] 8.4 Implement collection run: execute each request with assertions, skip those without, show per-request results + aggregate passed/failed/skipped summary
- [x] 8.5 Unit-test the assertion evaluator and collection-run summary logic

## 9. Polish & verification

- [x] 9.1 Verify a full keyboard-only workflow: create env → set active → create collection → add request → send → add assertions → run request → run collection
- [x] 9.2 Verify resize handling keeps the action bar visible and layout intact across terminal sizes
- [x] 9.3 Verify secrets never appear unmasked in previews and `secrets.json` is gitignored
- [x] 9.4 Run `go vet`/`gofmt` and ensure the build passes; smoke-test against a public test API

## 10. Run/test split & clipboard copy (amendment)

- [x] 10.1 Split the editor actions: `R` runs (send + show response, ignoring assertions) and `T` tests (send + evaluate assertions); update key bindings and the action bar
- [x] 10.2 Make the test action require at least one assertion and prompt to add one otherwise; the run action works regardless of assertions
- [x] 10.3 Add a copy action in the response view that copies the raw body to the clipboard via `atotto/clipboard`, bound to `c`/`y` (not Ctrl+C, reserved for quit)
- [x] 10.4 Update README key reference; unit-test that the response screen stores the raw body, advertises the copy binding, and handles the copy key
