## ADDED Requirements

### Requirement: Attach assertions to a request
The system SHALL allow the user to attach assertions directly to a request that specify expected results to evaluate against its response. A request that carries one or more assertions is treated as a test; there is no separate test entity.

#### Scenario: Assert on status code
- **WHEN** the user adds a status-code assertion to a request (e.g. expect 200)
- **THEN** the system stores the expected status code on the request and evaluates it against the response status when the request is run

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

### Requirement: Test a single request
The system SHALL provide a test action, distinct from the plain run action, that sends an individual request and evaluates all of its assertions, then reports pass or fail per assertion. The test action SHALL be separate from running a request: running shows only the response, while testing evaluates assertions.

#### Scenario: Test a request with no assertions
- **WHEN** the user triggers the test action on a request that has no assertions
- **THEN** the system declines to run the test and prompts the user to add assertions first

#### Scenario: All assertions pass
- **WHEN** the user tests a request and every assertion is satisfied
- **THEN** the system marks the run as passed and shows each assertion as passing

#### Scenario: One or more assertions fail
- **WHEN** the user tests a request and at least one assertion is not satisfied
- **THEN** the system marks the run as failed and shows which assertions failed with expected vs actual values

#### Scenario: Request error during test
- **WHEN** the underlying request fails to complete
- **THEN** the system marks the run as failed with an error describing the failure

### Requirement: Run all tests in a collection
The system SHALL run every request in a collection that carries assertions and report aggregate results. The collection is the batch-run unit; there is no separate suite entity.

#### Scenario: Run a collection
- **WHEN** the user triggers a run on a collection
- **THEN** the system executes each request that has assertions in order and displays per-request pass/fail results plus a summary count of passed and failed requests

#### Scenario: Requests without assertions are skipped
- **WHEN** a collection contains requests that have no assertions
- **THEN** the system skips those requests during a collection run and reflects them as skipped (not failed) in the summary

#### Scenario: Collection summary status
- **WHEN** a collection run completes
- **THEN** the system reports the collection run as passing only if every executed request passed, otherwise failing
