## MODIFIED Requirements

### Requirement: Attach assertions to a request
The system SHALL allow the user to attach assertions directly to a request that specify expected results to evaluate against its response. A request that carries one or more assertions is treated as a test; there is no separate test entity.

#### Scenario: Assert on status code
- **WHEN** the user adds a status-code assertion to a request (e.g. expect 200)
- **THEN** the system stores the expected status code on the request and evaluates it against the response status when the request is run

#### Scenario: Assert status code is not a value
- **WHEN** the user adds a status-code assertion with the not-equals operator (e.g. must not be 500)
- **THEN** the assertion passes only when the response status differs from the expected value, and the assertion is summarized using `!=`

#### Scenario: Assert on response header
- **WHEN** the user adds a header assertion specifying a header name and expected value or pattern
- **THEN** the system evaluates the named response header against the expectation

#### Scenario: Assert on response body
- **WHEN** the user adds a body assertion (contains text, equals, or JSON field equals)
- **THEN** the system evaluates the response body against the expectation

#### Scenario: Assert on latency
- **WHEN** the user adds a latency assertion (e.g. response time under N milliseconds)
- **THEN** the system evaluates the elapsed time against the threshold

#### Scenario: Remove an assertion
- **WHEN** the user removes an assertion from a request
- **THEN** the system deletes that assertion and persists the request

## ADDED Requirements

### Requirement: Assertion editor navigates only visible fields
The assertion editor SHALL only move focus between the fields that are currently visible for the selected assertion kind, skipping any field that is hidden, and SHALL keep focus on a visible field when changing the kind hides the currently focused field.

#### Scenario: Skip a hidden field while navigating
- **WHEN** the editor shows an assertion kind whose Target field is hidden (status code, latency, or non-JSON-path body) and the user moves focus down from the Operator field
- **THEN** focus lands directly on the next visible field (e.g. the expected/max-ms field) without an extra keystroke for the hidden field

#### Scenario: Focus stays visible after changing kind
- **WHEN** the currently focused field becomes hidden because the user changed the assertion kind
- **THEN** the editor moves focus to a visible field rather than leaving it on the hidden one
