## MODIFIED Requirements

### Requirement: Contextual action bar
The system SHALL display a persistent action/help bar at the bottom of every
screen listing the key bindings available in the current context, including the
bindings of the currently focused pane when the composite workspace is shown.

#### Scenario: Show context-specific bindings
- **WHEN** the user is on a given screen
- **THEN** the action bar lists the relevant key bindings for that screen (e.g. navigation, create, edit, delete, send, quit)

#### Scenario: Reflect the focused pane in the workspace
- **WHEN** the composite workspace is shown and a pane is focused
- **THEN** the action bar lists the bindings for that focused pane plus the
  global bindings (pane focus movement, send, quit), and updates when focus moves
  to another pane

#### Scenario: Update on context change
- **WHEN** the user moves to a different screen or mode
- **THEN** the action bar updates to reflect the bindings available in the new context

### Requirement: Screen navigation and back behavior
The system SHALL provide consistent navigation between the workspace and the
secondary screens/overlays it launches, including a way to go back and a global
quit. The core request/response loop is handled by pane focus within the
workspace rather than by pushing a new screen per step.

#### Scenario: Core loop stays in the workspace
- **WHEN** the user picks a request, edits it, sends it, and reads the response
- **THEN** each step happens by focusing the relevant pane in the workspace
  without pushing or popping full screens

#### Scenario: Go back from a secondary overlay
- **WHEN** the user has opened a secondary screen or overlay (e.g. environments,
  secrets, the assertions editor, run-all results, or a prompt) and presses the
  back key (e.g. Esc)
- **THEN** the system returns to the workspace without losing unsaved-prompt context

#### Scenario: Quit the application
- **WHEN** the user presses the soft-quit key (`q`) from the workspace with no
  unsaved editor changes, or presses the hard-quit key (`Ctrl+C`) at any time
- **THEN** the system exits cleanly; `Ctrl+C` is never intercepted by an
  unsaved-changes guard

#### Scenario: Distinguish edit mode from navigation mode
- **WHEN** the user is editing a text field
- **THEN** navigation/quit and pane-focus shortcuts that conflict with text entry are suspended until the user exits the field

### Requirement: Title shows version and navigation breadcrumb
The top chrome SHALL display the application name with its version and a
breadcrumb reflecting the current location, so the user can see which version is
running and the path that Esc will unwind.

#### Scenario: Version in the title
- **WHEN** the application renders with a known build version
- **THEN** the title shows the app name followed by that version (e.g. `termipost v1.1.0`)

#### Scenario: Development build version
- **WHEN** the build version is the development default (`dev` or unset)
- **THEN** the title shows a development indicator (e.g. `termipost (dev)`) rather than a `v`-prefixed placeholder

#### Scenario: Breadcrumb reflects the current location
- **WHEN** the user is in the workspace, optionally with a secondary screen or
  overlay open on top of it
- **THEN** the chrome shows a breadcrumb of the workspace and any open secondary
  screens in order, separated by a consistent separator

#### Scenario: Transient overlays are not shown in the breadcrumb
- **WHEN** a transient overlay (such as a text prompt or a confirmation dialog) is the active screen
- **THEN** the breadcrumb continues to reflect the underlying location and does not add a crumb for the overlay
