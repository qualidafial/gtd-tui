## MODIFIED Requirements

### Requirement: Status marker and title styling
Each task row SHALL begin with a status marker, optionally followed by a leading project label (see Requirement: Project label), and then the task title. The marker and title styling SHALL reflect the task status:

- `open` → `[ ]`, default title style.
- `done` → `[x]`, title rendered in a dim "done" color.
- `dropped` → `[-]`, title rendered with strikethrough in a dim "dropped" color.

#### Scenario: Open task marker
- **WHEN** rendering an open task titled "Buy milk" with no project
- **THEN** the row reads `[ ] Buy milk` with the default title style

#### Scenario: Done task marker and color
- **WHEN** rendering a done task
- **THEN** the marker is `[x]` and the title is rendered in the dim done color

#### Scenario: Dropped task marker and strikethrough
- **WHEN** rendering a dropped task
- **THEN** the marker is `[-]` and the title is rendered with strikethrough in the dim dropped color

### Requirement: Chip suppression by status
Due and defer chips signal actionability and SHALL be suppressed on done and dropped tasks. The assignee chip SHALL still render on done tasks; dropped tasks SHALL show no chips.

#### Scenario: Done task hides date chips but keeps assignee
- **WHEN** rendering a done task that has a due date and an assignee "alice"
- **THEN** no `due:`/`overdue:`/`defer:`/`ready:` chip is shown, and `@alice` is shown

#### Scenario: Dropped task hides all chips
- **WHEN** rendering a dropped task that has a due date and an assignee
- **THEN** no chips are shown

### Requirement: Chip ordering and alignment
When a task shows multiple chips, they SHALL be left-aligned and placed inline after the title in the order: due/overdue, then defer/ready, then assignee. The project association is not a chip; it renders as a leading label ahead of the title (see Requirement: Project label).

#### Scenario: Multiple chips on one row
- **WHEN** a pending task has a due date 28 days out and a defer date 14 days out
- **THEN** the row reads `[ ] <title>  due:28d  defer:14d` with chips left-aligned in that order

#### Scenario: Full chip set with project
- **WHEN** a pending task in project "Kitchen remodel" is due in 3 days, assigned to alice, with no defer
- **THEN** the row reads `[ ] Kitchen remodel <title>  due:3d  @alice` — the project label leads ahead of the title and the chips trail after it

### Requirement: Urgency colors
Chips SHALL be colorized to convey urgency, using huh theme colors where available and static colors otherwise:

- `overdue:` → red (highest urgency)
- `due:` today → orange
- `due:` 2–6 days → yellow
- `due:` 7+ days → dim/neutral
- `defer:` → dim blue (low urgency)
- `ready:` → teal/cyan (attention without alarm; visually distinct from overdue red)
- `@assignee` → magenta (visually distinct from all date chips)

#### Scenario: Overdue is the most urgent color
- **WHEN** an overdue chip and a far-future due chip are rendered
- **THEN** the overdue chip uses the red urgency color and the far-future due chip uses the dim/neutral color

#### Scenario: Ready is distinct from overdue
- **WHEN** a ready chip is rendered
- **THEN** it uses the teal/cyan color, not the overdue red

### Requirement: Selection highlight scope
The list selection highlight SHALL affect the task title and, when present, the leading project label (e.g. a cursor indicator and/or emphasis); both SHALL brighten together on the selected row. Data chips SHALL retain their own urgency colors on the selected row.

#### Scenario: Selected row keeps chip colors
- **WHEN** a row with a due chip is selected
- **THEN** the title reflects the selection highlight and the due chip keeps its urgency color

#### Scenario: Selected row brightens the project label
- **WHEN** a row with a project label is selected
- **THEN** the project label brightens (emphasizes) alongside the title

## ADDED Requirements

### Requirement: Project label
A task with a non-nil `ProjectID` SHALL display its project title as a leading label rendered ahead of the task title (after the status marker), with no `+` prefix, when the list is configured to show the project label and a `ProjectNameFunc` is provided. Lists rendered inside a single-project view (the in-project task list) SHALL NOT show the label on any row. The label SHALL be suppressed when the resolved title is empty. The label SHALL render on open and done tasks and SHALL be suppressed on dropped tasks. The label's foreground color SHALL be indigo (huh's logo color), visually distinct from the green `ready:` chip and from every other chip.

#### Scenario: Task in a project shows a leading label in the global task list
- **WHEN** the global task list renders a task whose ProjectID resolves to title "Kitchen remodel"
- **THEN** the row shows "Kitchen remodel" as a leading label ahead of the title, with no `+` prefix

#### Scenario: Standalone task shows no project label
- **WHEN** the task list renders a task with nil ProjectID
- **THEN** no project label is shown

#### Scenario: In-project task list suppresses the project label on every row
- **WHEN** the task list inside a project view renders any task
- **THEN** no project label is shown, regardless of the task's ProjectID

#### Scenario: Project resolution failure renders no label
- **WHEN** the configured ProjectNameFunc returns an empty string for the task's ProjectID
- **THEN** no project label is shown for that row

#### Scenario: Done task keeps the project label
- **WHEN** rendering a done task with a non-nil ProjectID
- **THEN** the project label is shown

#### Scenario: Dropped task hides the project label
- **WHEN** rendering a dropped task with a non-nil ProjectID
- **THEN** no project label is shown

#### Scenario: Project label color is distinct from ready
- **WHEN** a row renders a project label alongside a `ready:` chip
- **THEN** the label's indigo foreground is visually distinguishable from the ready chip's green

## REMOVED Requirements

### Requirement: Project chip
**Reason**: The project association moved from a trailing `+<title>` chip to a leading indigo label rendered ahead of the title. Its behavior is now covered by the ADDED Requirement: Project label.
**Migration**: No code migration needed (the leading-label rendering already ships). Refer to Requirement: Project label for the project-association presentation contract.
