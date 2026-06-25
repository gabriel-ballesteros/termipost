# payload-formatting Specification

## Purpose
TBD - created by archiving change add-payload-linting-prettify. Update Purpose after archive.
## Requirements
### Requirement: Syntax highlighting of body content
The system SHALL render JSON body content with syntax highlighting, applying distinct colours to object keys, string values, numbers, booleans, null, and structural punctuation. Highlighting SHALL apply both to the request body display and to the response body view.

#### Scenario: Highlight a JSON object
- **WHEN** the system displays a body whose content is valid JSON
- **THEN** keys, string values, numeric values, boolean/null literals, and punctuation are shown in visually distinct colours

#### Scenario: Non-JSON content is shown as plain text
- **WHEN** the body content is not JSON (e.g. plain text, HTML, or form data)
- **THEN** the system displays it as plain, unhighlighted text without error

#### Scenario: Malformed JSON does not break the display
- **WHEN** the body content begins like JSON but cannot be parsed
- **THEN** the system displays the raw content as plain text rather than a corrupted or partial rendering

### Requirement: Prettify request body
The system SHALL provide a prettify action in the request body editor that re-indents the current body as formatted JSON in place.

#### Scenario: Prettify valid JSON
- **WHEN** the user triggers prettify on a request body that is valid JSON
- **THEN** the system replaces the body with a consistently indented version and marks the request as having unsaved edits if the content changed

#### Scenario: Prettify is idempotent
- **WHEN** the user triggers prettify on a body that is already formatted
- **THEN** the system leaves the content effectively unchanged

#### Scenario: Prettify empty body
- **WHEN** the user triggers prettify on an empty or whitespace-only body
- **THEN** the system performs no change and reports no error

### Requirement: Validate JSON on prettify
When prettify is triggered, the system SHALL validate the body as JSON and SHALL NOT modify it if parsing fails.

#### Scenario: Report a syntax error
- **WHEN** the user triggers prettify on a body that is not valid JSON
- **THEN** the system leaves the body unchanged and shows an error message describing the parse failure, including the line and column of the offending input

#### Scenario: No false positives on valid JSON
- **WHEN** the user triggers prettify on syntactically valid JSON
- **THEN** the system reports no error and formats the body

### Requirement: Live JSON validity feedback
While the user is editing a request body whose content looks like JSON, the system SHALL continuously indicate whether the current content is valid JSON and SHALL surface the parse error inline. The feedback SHALL be limited to validity status and error text — the body text itself is not coloured per-token while editing.

#### Scenario: Indicate invalid JSON while typing
- **WHEN** the user is editing a body that starts with `{` or `[` and the current content is not valid JSON
- **THEN** the system shows an invalid indicator and an inline error describing the parse failure with its line and column

#### Scenario: Indicate valid JSON while typing
- **WHEN** the user is editing a body that starts with `{` or `[` and the current content is valid JSON
- **THEN** the system shows a valid indicator and no error

#### Scenario: No indicator for non-JSON or empty bodies
- **WHEN** the body being edited is empty or does not look like JSON (does not start with `{` or `[`)
- **THEN** the system shows no validity indicator and no error

