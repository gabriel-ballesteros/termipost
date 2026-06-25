## ADDED Requirements

### Requirement: Foldable JSON body view
The system SHALL render the read-only JSON view of a request body and a response body as a foldable view in which every `{ }` object and `[ ]` array that spans more than one line, at any nesting depth, is an independently collapsible section. Folding SHALL apply only when the body content is valid JSON; non-JSON bodies SHALL be displayed as before with no fold gutter. An object or array that occupies a single line (including empty `{}` and `[]`) SHALL NOT be foldable.

#### Scenario: Nested sections are foldable
- **WHEN** the system displays a valid JSON body that contains nested multi-line objects and arrays
- **THEN** each object and array, including nested ones, is an independently foldable section

#### Scenario: Single-line and empty containers are not foldable
- **WHEN** the displayed JSON contains an empty `{}` or `[]`, or an object/array that fits on one line
- **THEN** that container has no fold marker and cannot be collapsed

#### Scenario: Non-JSON body is not foldable
- **WHEN** the body content is empty or is not valid JSON
- **THEN** the system displays it without a fold gutter and without fold behavior

### Requirement: Fold gutter markers
The system SHALL render a left gutter on each line of a foldable body view. A line that opens an expanded section SHALL show `-`, a line whose section is collapsed SHALL show `+`, and any line that does not open a foldable section SHALL show a blank gutter aligned with the markers.

#### Scenario: Expanded section shows minus
- **WHEN** a line opens a section that is currently expanded
- **THEN** that line's gutter shows `-`

#### Scenario: Collapsed section shows plus
- **WHEN** a line opens a section that is currently collapsed
- **THEN** that line's gutter shows `+`

#### Scenario: Non-header line shows blank gutter
- **WHEN** a line does not open a foldable section
- **THEN** that line's gutter is blank but aligned with the `+`/`-` markers

### Requirement: Collapsed section rendering
When a section is collapsed, the system SHALL display only its opening header line followed by a placeholder indicating omitted content and the matching closing bracket, and SHALL hide all lines contained within that section. Lines inside a collapsed section SHALL NOT be shown even if they open their own foldable sections.

#### Scenario: Collapsed object shows placeholder
- **WHEN** an object section is collapsed
- **THEN** the system shows its opening line with a placeholder such as `{ … }` and hides the object's member lines

#### Scenario: Collapsed array shows placeholder
- **WHEN** an array section is collapsed
- **THEN** the system shows its opening line with a placeholder such as `[ … ]` and hides the array's element lines

#### Scenario: Nested content hidden under a collapsed parent
- **WHEN** a section is collapsed and contains its own nested sections
- **THEN** none of the nested lines are shown while the parent stays collapsed

#### Scenario: Collapsed section keeps its trailing punctuation
- **WHEN** a section whose closing line carries a trailing comma is collapsed
- **THEN** the collapsed placeholder line retains that trailing comma so the surrounding JSON stays syntactically consistent on screen

### Requirement: Line cursor navigation
The read-only foldable body view SHALL maintain a line cursor over the currently visible lines and SHALL allow the user to move it up and down. The cursor SHALL only traverse visible lines, skipping lines hidden inside collapsed sections. The line under the cursor SHALL be visually distinguished.

#### Scenario: Move cursor down
- **WHEN** the user presses the down key in a foldable body view
- **THEN** the cursor moves to the next visible line and that line is highlighted

#### Scenario: Move cursor up
- **WHEN** the user presses the up key in a foldable body view
- **THEN** the cursor moves to the previous visible line and that line is highlighted

#### Scenario: Cursor skips hidden lines
- **WHEN** a section is collapsed and the user moves the cursor past it
- **THEN** the cursor lands on the next visible line and never on a line hidden inside the collapsed section

#### Scenario: View follows the cursor in a scrollable body
- **WHEN** the cursor moves beyond the visible window of a body taller than the pane
- **THEN** the view scrolls so the cursor line is fully visible, accounting for lines that soft-wrap across multiple screen rows

### Requirement: Toggle collapse and expand
The system SHALL provide a toggle binding (`space`) that operates on the section associated with the line cursor. When the cursor is on an expanded section header the binding SHALL collapse it; when on a collapsed section header it SHALL expand it. When the cursor is on a line that is not a section header, the binding SHALL toggle the nearest enclosing section.

#### Scenario: Collapse an expanded section
- **WHEN** the cursor is on an expanded section header and the user presses space
- **THEN** that section collapses and its contents are hidden

#### Scenario: Expand a collapsed section
- **WHEN** the cursor is on a collapsed section header and the user presses space
- **THEN** that section expands and its contents become visible

#### Scenario: Toggle from a non-header line
- **WHEN** the cursor is on a line that is not a section header and the user presses space
- **THEN** the nearest enclosing section is toggled

### Requirement: Folding does not modify the body
Folding SHALL be a view-only operation. The system SHALL NOT change the stored request or response body when sections are collapsed or expanded, and folding SHALL NOT be available while the request body is being edited in the textarea.

#### Scenario: Stored body unchanged after folding
- **WHEN** the user collapses or expands sections in a body view
- **THEN** the persisted request body and the captured response body are unchanged, and the request is not marked as having unsaved edits because of folding

#### Scenario: Folding unavailable while editing
- **WHEN** the user is editing the request body in the textarea
- **THEN** the fold gutter, cursor, and toggle binding are not active and the textarea behaves as a normal text editor
