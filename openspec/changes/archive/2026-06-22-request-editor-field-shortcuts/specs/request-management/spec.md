## ADDED Requirements

### Requirement: Navigate and activate request editor fields
In the request editor, every field — including Assertions — SHALL be reachable with the arrow/Tab keys and activatable with Enter, and each field SHALL additionally have a single-letter shortcut that jumps to it and activates it in one keystroke.

#### Scenario: Assertions is reachable by arrow keys
- **WHEN** the user moves the focus through the editor fields with the arrow or Tab keys
- **THEN** the Assertions row is included in the cycle and can be highlighted like any other field

#### Scenario: Open Assertions with Enter
- **WHEN** the Assertions row is focused and the user presses Enter
- **THEN** the system opens the assertions editor for the request

#### Scenario: First-letter shortcut activates a field
- **WHEN** the user presses a field's first-letter shortcut (`n` Name, `m` Method, `u` URL, `h` Headers, `p` Params, `b` Body, `a` Assertions) in navigation mode
- **THEN** the system focuses that field and activates it: text fields (Name, URL, Body) enter edit mode, Headers/Params/Assertions open their editors, and Method receives focus so the arrow keys can change it

#### Scenario: Shortcuts are inert while editing a field
- **WHEN** the user is editing a text field and types a letter that is also a shortcut
- **THEN** the system inserts the character into the field rather than triggering the shortcut

#### Scenario: Method cycles with arrow keys
- **WHEN** the Method field is focused
- **THEN** the left/right arrow keys cycle the HTTP method
