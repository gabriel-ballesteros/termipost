# tui-navigation Specification

## Purpose
Provide keyboard-only navigation, contextual help, consistent screen transitions, and a responsive layout for the terminal UI.
## Requirements
### Requirement: Keyboard-only operation
The system SHALL make every action reachable and executable using only the keyboard, with no dependency on a mouse.

#### Scenario: Navigate lists with keyboard
- **WHEN** a list of items (collections, requests, tests) is displayed
- **THEN** the user can move the selection with both arrow keys and vim-style keys (j/k) and confirm with Enter

#### Scenario: Complete a full workflow without a mouse
- **WHEN** the user creates a collection, adds a request, sends it, and runs a test
- **THEN** every step is achievable through keyboard input alone

### Requirement: Contextual action bar
The system SHALL display a persistent action/help bar at the bottom of every screen listing the key bindings available in the current context.

#### Scenario: Show context-specific bindings
- **WHEN** the user is on a given screen
- **THEN** the action bar lists the relevant key bindings for that screen (e.g. navigation, create, edit, delete, send, quit)

#### Scenario: Update on context change
- **WHEN** the user moves to a different screen or mode
- **THEN** the action bar updates to reflect the bindings available in the new context

### Requirement: Screen navigation and back behavior
The system SHALL provide consistent navigation between screens, including a way to go back and a global quit.

#### Scenario: Go back to the previous screen
- **WHEN** the user presses the back key (e.g. Esc) on a nested screen
- **THEN** the system returns to the previous screen without losing unsaved-prompt context

#### Scenario: Quit the application
- **WHEN** the user presses the quit key (e.g. q or Ctrl+C) at a top-level screen
- **THEN** the system exits cleanly

#### Scenario: Distinguish edit mode from navigation mode
- **WHEN** the user is editing a text field
- **THEN** navigation/quit shortcuts that conflict with text entry are suspended until the user exits the field

### Requirement: Responsive layout
The system SHALL render correctly across a range of terminal sizes, adapt to terminal resize events, and size each screen to the current terminal as soon as it is opened so its content is visible immediately.

#### Scenario: Adapt to terminal resize
- **WHEN** the terminal window is resized
- **THEN** the system reflows its layout to fit the new dimensions and keeps the action bar visible

#### Scenario: Newly opened screen is sized immediately
- **WHEN** the user opens (navigates into) a screen after startup
- **THEN** the system sizes that screen to the current terminal dimensions before it is first rendered

#### Scenario: List screens show their items on open
- **WHEN** the user opens a screen containing a scrollable list (such as a collection's requests or the environments list)
- **THEN** the list renders its item rows immediately, not only a pagination indicator

