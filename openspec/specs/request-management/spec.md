# request-management Specification

## Purpose
Create, edit, delete, run, and substitute variables in HTTP requests within collections, and view their responses.

## Requirements

### Requirement: Create and edit requests
The system SHALL allow the user to create and edit an HTTP request within a collection, specifying the HTTP method, URL, headers, query parameters, and body.

#### Scenario: Create a new request
- **WHEN** the user creates a request inside a collection and provides a name, method, and URL
- **THEN** the system adds the request to the collection and persists it

#### Scenario: Edit request fields
- **WHEN** the user edits a request's method, URL, headers, query parameters, or body
- **THEN** the system saves the updated values to the request

#### Scenario: Select HTTP method
- **WHEN** the user changes the method field
- **THEN** the system offers the standard HTTP methods (GET, POST, PUT, PATCH, DELETE, HEAD, OPTIONS) for selection

#### Scenario: Edit headers as key/value pairs
- **WHEN** the user adds, edits, or removes a header entry
- **THEN** the system stores headers as key/value pairs on the request

### Requirement: Delete request
The system SHALL allow the user to delete a request from a collection after confirmation.

#### Scenario: Confirmed request deletion
- **WHEN** the user deletes a request and confirms
- **THEN** the system removes the request from the collection and from disk

### Requirement: Run request and view response
The system SHALL provide a run action that sends the configured HTTP request and displays the response status code, headers, body, and elapsed time. Running a request SHALL be independent of assertions — it does not require any to be defined and does not evaluate them (see the api-testing capability for the separate test action).

#### Scenario: Successful run
- **WHEN** the user triggers the run action on a valid request
- **THEN** the system performs the HTTP call and displays the status code, response headers, response body, and total elapsed time

#### Scenario: Run a request that has no assertions
- **WHEN** the user runs a request that has no assertions defined
- **THEN** the system sends it and shows the response without reporting any test outcome or error about missing assertions

#### Scenario: Format JSON response body
- **WHEN** the response body is valid JSON
- **THEN** the system displays the body pretty-printed for readability

#### Scenario: Network or connection error
- **WHEN** the request fails to reach the server (DNS failure, timeout, refused connection)
- **THEN** the system displays a clear error message and does not crash

#### Scenario: Non-blocking run
- **WHEN** a request is in flight
- **THEN** the system shows a loading indicator and the interface remains responsive to cancel/quit

#### Scenario: Copy response body to clipboard
- **WHEN** the user triggers the copy action while viewing a response
- **THEN** the system copies the raw (unformatted) response body to the system clipboard and confirms the copy
- **AND** the copy action does not use Ctrl+C, which is reserved for quitting

### Requirement: Variable substitution
The system SHALL substitute variables referenced in request URL, headers, query parameters, and body before sending, resolving each `{{name}}` reference against the active environment and then the global secrets store (see the environment-management capability for resolution rules).

#### Scenario: Substitute a defined variable
- **WHEN** a request field references a variable using the `{{name}}` syntax and that name resolves from the active environment or secrets
- **THEN** the system replaces the reference with the resolved value before sending

#### Scenario: Undefined variable
- **WHEN** a request field references a variable that resolves from neither the active environment nor secrets
- **THEN** the system leaves the reference unresolved and warns the user before or after sending
