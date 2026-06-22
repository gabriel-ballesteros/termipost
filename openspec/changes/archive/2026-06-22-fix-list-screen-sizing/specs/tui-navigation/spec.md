## MODIFIED Requirements

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
