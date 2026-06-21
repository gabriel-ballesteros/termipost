## ADDED Requirements

### Requirement: Human-readable file storage
The system SHALL persist configuration, collections (including their requests and nested assertions), and environments as human-readable JSON files on disk.

#### Scenario: Persist on change
- **WHEN** the user creates or modifies a collection, request, assertion, or environment
- **THEN** the system writes the change to its JSON file so it survives application restart

#### Scenario: Human-readable formatting
- **WHEN** the system writes a data file
- **THEN** the file is formatted with indentation so a human can read and hand-edit it

### Requirement: Defined storage location and layout
The system SHALL store its data under a predictable user-scoped directory using a defined file layout.

#### Scenario: Use a config/data directory
- **WHEN** the application starts and no data directory exists
- **THEN** the system creates its data directory (e.g. under the user's config home) before writing files

#### Scenario: Defined file layout
- **WHEN** the system writes data
- **THEN** it uses `config.json` for app settings and the active environment, `collections/<id>.json` for each collection with its requests and nested assertions, `environments/<id>.json` for each environment's variables, and `secrets.json` for secret values

#### Scenario: Load existing data on startup
- **WHEN** the application starts and data files already exist
- **THEN** the system loads collections, environments, secrets, and configuration from those files

### Requirement: Keep secrets out of version control
The system SHALL ensure the global secrets file is excluded from version control.

#### Scenario: Generate gitignore
- **WHEN** the system initializes its data directory
- **THEN** the system writes a `.gitignore` entry that excludes `secrets.json` so secret values are never committed

### Requirement: Atomic and safe writes
The system SHALL write data files atomically so that an interrupted write does not corrupt existing data.

#### Scenario: Atomic write
- **WHEN** the system saves a data file
- **THEN** it writes to a temporary file and renames it into place so a crash mid-write leaves the previous file intact

### Requirement: Resilient loading
The system SHALL handle missing or malformed data files without crashing.

#### Scenario: Missing files
- **WHEN** an expected data file does not exist
- **THEN** the system treats it as empty and continues without error

#### Scenario: Malformed file
- **WHEN** a data file contains invalid JSON
- **THEN** the system reports the problem to the user and continues running rather than crashing, without silently overwriting the bad file
