## Why

Developers who live in the terminal need a fast, keyboard-driven way to build, send, and validate HTTP requests without leaving their shell or reaching for a heavyweight GUI like Postman. Existing terminal HTTP tools (curl, httpie) are great for one-off calls but lack first-class support for organizing requests into collections, persisting them in version-controllable files, switching between environments, and asserting on responses as repeatable tests. `termipost` fills that gap: a Postman-like experience that runs entirely in the terminal, stores everything as human-readable files, and never requires a mouse.

## What Changes

- Introduce a new Go TUI application (`termipost`) built on Bubble Tea, Lip Gloss, and Bubbles.
- Add **collection management**: create, rename, delete, and browse collections that group related requests.
- Add **request management**: create/edit/delete requests with method, URL, headers, query params, and body; **run** them (send and view the formatted response — status, headers, body, timing — regardless of assertions) and copy the response body to the clipboard.
- Add **API testing**: attach assertions (status code, headers, body, latency) directly to a request; a **test** action (distinct from run) sends the request and evaluates its assertions, and a collection can be run to test every request-with-assertions as a batch with an aggregate pass/fail summary. A "test" is simply a request that carries assertions — there is no separate test or suite entity.
- Add **environment management**: define multiple named environments (e.g. `local`, `prod`), each holding a set of variables; one environment is active at a time and its variables resolve `{{name}}` references. Secret values (tokens, keys) live in a single gitignored secrets file and are masked everywhere in the UI.
- Add **keyboard-only navigation**: every action is reachable via the keyboard with no mouse dependency, including vim-style and arrow-key movement.
- Add a persistent **action/help bar** at the bottom of every screen listing the available key bindings for the current context.
- Add **file-based persistence**: configs, collections (with their requests and assertions), and environments are stored as human-readable JSON files on disk; secrets are stored separately and excluded from version control.

## Capabilities

### New Capabilities
- `collection-management`: Creating, organizing, and persisting collections that group requests.
- `request-management`: Authoring, editing, and sending HTTP requests and viewing their responses.
- `api-testing`: Attaching assertions to requests, running a request as a test, and running all tests in a collection.
- `environment-management`: Defining environments and variables, switching the active environment, and handling masked secrets.
- `tui-navigation`: Keyboard-only navigation, screen flow, and the contextual action/help bar.
- `data-persistence`: Reading and writing configs, collections, and environments as human-readable JSON files, with secrets kept out of version control.

### Modified Capabilities
<!-- None — this is a greenfield project with no existing specs. -->

## Impact

- New Go module and project scaffold (`go.mod`, `main.go`, internal packages).
- New dependencies: `github.com/charmbracelet/bubbletea`, `github.com/charmbracelet/lipgloss`, `github.com/charmbracelet/bubbles`.
- New on-disk data layout under a user config/data directory (e.g. `~/.config/termipost/`) holding `config.json`, `collections/<id>.json` (requests with nested assertions), `environments/<id>.json`, a gitignored `secrets.json`, and a generated `.gitignore`.
- No existing code or APIs are affected (greenfield repository).
