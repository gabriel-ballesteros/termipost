# environment-management Specification

## Purpose
Manage named environments and a global secrets store, control the single active environment, and resolve `{{name}}` variable references for requests.

## Requirements

### Requirement: Manage environments
The system SHALL allow the user to create, rename, and delete named environments, each holding a set of variables as key/value pairs.

#### Scenario: Create an environment
- **WHEN** the user creates an environment with a name not already in use
- **THEN** the system creates an empty environment with that name and persists it

#### Scenario: Reject duplicate environment name
- **WHEN** the user creates or renames an environment to a name already in use
- **THEN** the system rejects the operation and displays an error

#### Scenario: Edit environment variables
- **WHEN** the user adds, edits, or removes a variable in an environment
- **THEN** the system stores the change and persists the environment

#### Scenario: Delete an environment
- **WHEN** the user deletes an environment and confirms
- **THEN** the system removes the environment and, if it was active, clears the active selection

### Requirement: Single active environment
The system SHALL keep at most one environment active at a time, and SHALL use only the active environment's variables when resolving references (single-layer resolution, no collection-level or global variable layering).

#### Scenario: Switch the active environment
- **WHEN** the user selects a different environment as active
- **THEN** the system records it as active in configuration and subsequent variable resolution uses that environment's variables

#### Scenario: No active environment
- **WHEN** no environment is active
- **THEN** non-secret variable references resolve only against secrets, and any other reference is treated as undefined

### Requirement: Variable reference resolution
The system SHALL resolve `{{name}}` references first against the active environment's variables and then against the global secrets store.

#### Scenario: Resolve from active environment
- **WHEN** a `{{name}}` reference matches a variable in the active environment
- **THEN** the system substitutes the environment variable's value

#### Scenario: Resolve from secrets
- **WHEN** a `{{name}}` reference does not match an environment variable but matches a key in the global secrets store
- **THEN** the system substitutes the secret value

### Requirement: Global secrets store
The system SHALL store secret values (such as tokens and API keys) in a single global secrets file that is excluded from version control, separate from environment files.

#### Scenario: Secrets file is gitignored
- **WHEN** the system initializes its data directory
- **THEN** the system writes a `.gitignore` that excludes the secrets file so secret values are never committed

#### Scenario: Secrets are shared across environments
- **WHEN** the user defines a secret
- **THEN** the secret is available to variable resolution regardless of which environment is active

### Requirement: Mask secret values in the UI
The system SHALL mask secret values everywhere they would otherwise be displayed, while still sending the real value over the wire when a request is run.

#### Scenario: Mask in the secrets editor
- **WHEN** the user views the secrets store
- **THEN** the system displays secret values masked, with an explicit reveal action required to show a value

#### Scenario: Mask in a resolved-request preview
- **WHEN** a screen shows a request with its variables already resolved
- **THEN** the system masks any value that came from the secrets store so it does not leak into the display

#### Scenario: Real value is sent
- **WHEN** a request that references a secret is sent
- **THEN** the system uses the real secret value in the outgoing HTTP request despite masking it in the UI
