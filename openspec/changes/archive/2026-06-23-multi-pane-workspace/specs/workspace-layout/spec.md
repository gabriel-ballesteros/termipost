## ADDED Requirements

### Requirement: Composite single-screen workspace
The system SHALL provide a primary workspace that presents the request/response
loop on a single screen composed of multiple panes: a top method+URL+Send bar, a
collection/request tree pane, a request editor pane, and a response pane — all
visible at once.

#### Scenario: All core panes visible together
- **WHEN** the workspace is open
- **THEN** the collection/request tree, the request editor, and the response are
  all rendered on the same screen along with the method+URL+Send bar, without
  navigating between separate full screens

#### Scenario: Selecting a request populates the editor and response panes
- **WHEN** the user selects a request in the tree pane
- **THEN** the request editor pane shows that request's fields and the response
  pane shows that request's last response (or an empty state if none), without a
  screen transition

### Requirement: Collapsible collection tree
The tree pane SHALL present collections as expandable nodes containing their
requests, with each collection independently collapsible from the keyboard.

#### Scenario: Toggle a collection open or closed
- **WHEN** the tree pane is focused with a collection node selected and the user
  presses the toggle key (Enter or Space)
- **THEN** that collection expands to reveal its requests or collapses to hide
  them, with a glyph indicating its state (e.g. `▸` collapsed, `▾` expanded)

#### Scenario: Enter on a request loads it
- **WHEN** the selected tree node is a request and the user presses Enter
- **THEN** that request loads into the editor and response panes (subject to the
  unsaved-edit guard), rather than toggling anything

#### Scenario: Navigation skips collapsed requests
- **WHEN** a collection is collapsed
- **THEN** its requests are hidden and skipped by up/down selection movement, and
  become selectable again once the collection is expanded

### Requirement: Single-pane focus
The system SHALL keep exactly one pane focused at a time, route content key
input to the focused pane only, and visually indicate which pane is focused.

#### Scenario: Only the focused pane consumes content keys
- **WHEN** a pane is focused and the user presses a content key (e.g. a list
  movement key or a text edit key)
- **THEN** only the focused pane reacts and the other panes are unaffected

#### Scenario: Focused pane is visually distinct
- **WHEN** a pane has focus
- **THEN** it is rendered with a distinct indicator (such as a highlighted
  border or title) that the unfocused panes do not have

### Requirement: Pane focus movement
The system SHALL let the user move focus between panes with a vim-style window
chord — `ctrl+w` followed by a direction (`h`/`j`/`k`/`l`) — while reserving
`Tab`/`Shift+Tab` and the arrow/`j`/`k` keys for navigation *within* the focused
pane.

#### Scenario: Jump focus directionally with the window chord
- **WHEN** the user presses `ctrl+w` then a direction key (`h`/`j`/`k`/`l`)
- **THEN** focus moves to the adjacent pane in that direction, and is a no-op
  when there is no pane in that direction

#### Scenario: Tab navigates within the focused pane, not between panes
- **WHEN** a pane is focused and the user presses `Tab` or `Shift+Tab`
- **THEN** the selection moves between fields/rows inside that pane and focus
  does not leave the pane

#### Scenario: Window chord does not clash with in-pane keys
- **WHEN** the request editor pane is focused and exposes single-letter field
  jump shortcuts and `Tab`/`j`/`k` in-pane navigation
- **THEN** the `ctrl+w` window chord remains distinct from all of them so neither
  intercepts the other

### Requirement: Per-pane tabbed sections
The system SHALL let the request editor and response panes group their content
into tabs, with one active tab at a time switchable from the keyboard when the
pane is focused.

#### Scenario: Switch the active tab in a focused pane
- **WHEN** the request editor or response pane is focused and the user presses
  the tab-switch keys (`]` for next, `[` for previous)
- **THEN** the pane shows the next/previous section (e.g. Headers → Query → Body
  for the editor; Body → Headers for the response) and the previously hidden
  content is reachable again by switching back

#### Scenario: Tabs reflect available content
- **WHEN** a pane renders its tab strip
- **THEN** it lists the sections available for the current request or response
  and indicates which tab is active

### Requirement: Inline key/value editing in the request pane
The system SHALL let the user edit request headers and query parameters as rows
directly within the request editor pane, without opening a separate screen.

#### Scenario: Edit header rows in place
- **WHEN** the Headers (or Query) tab is active in the focused request editor
  pane
- **THEN** the existing key/value rows are shown in the pane and the user can
  add, edit, and remove rows from the keyboard without a screen transition

#### Scenario: Inline edits persist with the request
- **WHEN** the user changes a header or query-param row inline and saves the
  request
- **THEN** the change is stored with that request the same as any other field

### Requirement: Confirm before discarding unsaved edits
The system SHALL warn the user before switching the selected request away from an
editor that has unsaved changes, and SHALL only switch if the user confirms.

#### Scenario: Prompt when leaving a dirty editor
- **WHEN** the request editor pane has unsaved changes and the user selects a
  different request in the tree pane
- **THEN** the system prompts the user to confirm discarding the changes before
  loading the other request

#### Scenario: Confirm or cancel the switch
- **WHEN** the discard-changes prompt is shown
- **THEN** confirming loads the newly selected request and discards the edits,
  and cancelling keeps the current request and its unsaved edits intact

#### Scenario: No prompt when the editor is clean
- **WHEN** the request editor has no unsaved changes and the user selects a
  different request
- **THEN** the new request loads immediately without a prompt

#### Scenario: Confirm on soft-quit with unsaved edits
- **WHEN** the request editor has unsaved changes and the user presses the
  soft-quit key (`q`)
- **THEN** the system prompts the user to confirm discarding the changes before
  exiting, while the hard-quit key (`Ctrl+C`) still exits immediately without a
  prompt

### Requirement: Responsive multi-pane reflow
The system SHALL size and arrange the panes to the current terminal dimensions,
reflow on resize, and keep the method+URL+Send bar and the action bar visible.

#### Scenario: Reflow on terminal resize
- **WHEN** the terminal is resized while the workspace is open
- **THEN** the panes resize to fit the new dimensions and the top bar and bottom
  action bar remain visible

#### Scenario: Panes sized on first render
- **WHEN** the workspace is first shown after the terminal size is known
- **THEN** each pane renders its content immediately at the correct size rather
  than an empty or placeholder state

#### Scenario: Degrade to a single pane on small terminals
- **WHEN** the terminal is too small to fit the multi-pane layout at its minimum
  pane sizes
- **THEN** the system shows only the focused pane filling the available area, and
  the pane-focus chord switches which single pane is shown

#### Scenario: Return to the multi-pane layout when space allows
- **WHEN** the terminal is resized back to at least the minimum multi-pane size
  while in the single-pane fallback
- **THEN** the system restores the full multi-pane layout, keeping the same pane
  focused
