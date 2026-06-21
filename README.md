# termipost

A Postman-like HTTP client for the terminal. Build, organize, send, and validate
HTTP requests entirely from the keyboard — no mouse required. Built with
[Bubble Tea](https://github.com/charmbracelet/bubbletea),
[Bubbles](https://github.com/charmbracelet/bubbles), and
[Lip Gloss](https://github.com/charmbracelet/lipgloss).

## Features

- **Collections** — group related requests; create, rename, delete.
- **Requests** — set method, URL, headers, query params, and body; send and view
  the formatted response (status, headers, pretty-printed JSON body, timing).
- **Tests** — attach assertions (status code, header, body, latency) to a
  request and run it as a test. Run a whole collection to get an aggregate
  pass/fail/skip summary. A "test" is just a request with assertions.
- **Environments** — define multiple environments (e.g. `local`, `prod`), each
  with its own variables; switch the active one. References use `{{name}}`.
- **Secrets** — keep tokens and keys in a single gitignored secrets file; values
  are masked everywhere in the UI but sent for real over the wire.
- **Keyboard-only** — every action has a key binding, and the bar at the bottom
  of each screen always shows what is available in the current context.

## Install / Run

Requires Go 1.21+.

```sh
make run        # build and run
# or
go run .
```

Build a binary:

```sh
make build      # produces ./termipost
./termipost
```

## Data location

All data is stored as human-readable JSON under your OS config dir
(`~/.config/termipost` on Linux/macOS, falling back to `~/.termipost`):

```
config.json              app settings + active environment
collections/<id>.json    a collection with its requests and assertions
environments/<id>.json   an environment's variables
secrets.json             global secrets (gitignored automatically)
.gitignore               excludes secrets.json
```

Files are safe to read and hand-edit. Malformed files are reported on startup and
skipped rather than overwritten.

## Keys (high level)

- `↑/↓` or `j/k` to move, `enter` to open, `esc` to go back, `q` to quit.
- Lists: `n` new, `r` rename, `d` delete.
- Collections: `R` run all tests, `e` environments.
- Request editor: `tab` to move between fields, `enter` to edit a field,
  `R` run (send and show the response, ignoring assertions),
  `T` test (send and check assertions), `a` assertions, `ctrl+s` save.
- Response view: `↑/↓` scroll, `c` (or `y`) copy the body to the clipboard
  (`ctrl+c` is reserved for quit), `esc` back.
- Environments: `a` set active, `enter` edit variables, `s` secrets.

The bottom action bar always lists the keys for the current screen.

## Development

```sh
make test       # run unit + TUI smoke tests
make fmt        # gofmt
make vet        # go vet
```
