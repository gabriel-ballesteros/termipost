## ADDED Requirements

### Requirement: Title shows version and navigation breadcrumb
The top chrome SHALL display the application name with its version and a breadcrumb reflecting the current navigation stack, so the user can see which version is running and the path that Esc will unwind.

#### Scenario: Version in the title
- **WHEN** the application renders with a known build version
- **THEN** the title shows the app name followed by that version (e.g. `termipost v1.1.0`)

#### Scenario: Development build version
- **WHEN** the build version is the development default (`dev` or unset)
- **THEN** the title shows a development indicator (e.g. `termipost (dev)`) rather than a `v`-prefixed placeholder

#### Scenario: Breadcrumb reflects the open screens
- **WHEN** the user has navigated into nested screens (for example a collection, then a request, then its assertions)
- **THEN** the chrome shows a breadcrumb of those screens in order, separated by a consistent separator

#### Scenario: Transient overlays are not shown in the breadcrumb
- **WHEN** a transient overlay (such as a text prompt or a confirmation dialog) is the active screen
- **THEN** the breadcrumb continues to reflect the underlying navigation path and does not add a crumb for the overlay
