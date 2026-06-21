# collection-management Specification

## Purpose
Manage named collections that group related HTTP requests, including creating, browsing, renaming, and deleting collections.

## Requirements

### Requirement: Create collection
The system SHALL allow the user to create a new named collection to group related requests.

#### Scenario: Create a collection with a unique name
- **WHEN** the user triggers the create-collection action and enters a name that does not already exist
- **THEN** the system creates an empty collection with that name and shows it in the collection list

#### Scenario: Reject duplicate collection name
- **WHEN** the user enters a name that matches an existing collection
- **THEN** the system rejects the creation and displays an error message without overwriting the existing collection

#### Scenario: Reject empty collection name
- **WHEN** the user confirms creation with a blank name
- **THEN** the system rejects the creation and prompts for a non-empty name

### Requirement: List and browse collections
The system SHALL display all available collections and allow the user to select one to view its requests.

#### Scenario: View collections on launch
- **WHEN** the application starts
- **THEN** the system loads all persisted collections and displays them in a navigable list

#### Scenario: Open a collection
- **WHEN** the user selects a collection and confirms
- **THEN** the system displays the requests contained in that collection

#### Scenario: Empty state
- **WHEN** there are no collections
- **THEN** the system displays an empty-state message indicating how to create the first collection

### Requirement: Rename collection
The system SHALL allow the user to rename an existing collection.

#### Scenario: Rename to a unique name
- **WHEN** the user renames a collection to a name not already in use
- **THEN** the system updates the collection name and persists the change

#### Scenario: Reject rename to existing name
- **WHEN** the user renames a collection to a name already used by another collection
- **THEN** the system rejects the rename and displays an error

### Requirement: Delete collection
The system SHALL allow the user to delete a collection, including all requests it contains, after confirmation.

#### Scenario: Confirmed deletion
- **WHEN** the user triggers delete on a collection and confirms the prompt
- **THEN** the system removes the collection and its requests from disk and the list

#### Scenario: Cancelled deletion
- **WHEN** the user triggers delete but cancels the confirmation prompt
- **THEN** the system leaves the collection unchanged
