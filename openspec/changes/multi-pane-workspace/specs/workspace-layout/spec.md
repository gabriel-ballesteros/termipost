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
The system SHALL let the user move focus between panes using both a cycle
(forward and backward) and direct directional jumps, all from the keyboard.

#### Scenario: Cycle focus forward and backward
- **WHEN** the user presses the cycle-forward key (Tab) or the cycle-backward
  key (Shift+Tab)
- **THEN** focus moves to the next or previous pane in a consistent order and
  wraps around at the ends

#### Scenario: Jump focus directionally
- **WHEN** the user presses a directional focus key toward an adjacent pane
- **THEN** focus moves to the pane in that direction, and is a no-op when there
  is no pane in that direction

#### Scenario: Focus keys do not clash with field shortcuts
- **WHEN** the request editor pane is focused and exposes single-letter field
  jump shortcuts
- **THEN** the pane-focus keys remain distinct from those shortcuts so neither
  intercepts the other

### Requirement: Per-pane tabbed sections
The system SHALL let the request editor and response panes group their content
into tabs, with one active tab at a time switchable from the keyboard when the
pane is focused.

#### Scenario: Switch the active tab in a focused pane
- **WHEN** the request editor or response pane is focused and the user presses
  the tab-switch key
- **THEN** the pane shows the next section (e.g. Headers → Query → Body for the
  editor; Body → Headers for the response) and the previously hidden content is
  reachable again by switching back

#### Scenario: Tabs reflect available content
- **WHEN** a pane renders its tab strip
- **THEN** it lists the sections available for the current request or response
  and indicates which tab is active

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
