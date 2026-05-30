## ADDED Requirements

### Requirement: Project chip
A task with a non-nil `ProjectID` SHALL display a `+<project title>` chip when the list is configured to show project chips and a `ProjectNameFunc` is provided. Lists rendered inside a single-project view (the in-project task list) SHALL NOT show the chip on any row. The chip SHALL be visually distinct from `due:`, `defer:`, `ready:`, `overdue:`, and `@assignee` chips.

#### Scenario: Task in a project shows project chip in the global task list
- **WHEN** the global task list renders a task whose ProjectID resolves to title "Kitchen remodel"
- **THEN** the row includes the chip `+Kitchen remodel`

#### Scenario: Standalone task shows no project chip
- **WHEN** the task list renders a task with nil ProjectID
- **THEN** no `+...` chip is shown

#### Scenario: In-project task list suppresses project chip on every row
- **WHEN** the task list inside a project view renders any task
- **THEN** no `+...` chip is shown, regardless of the task's ProjectID

#### Scenario: Project resolution failure renders no chip
- **WHEN** the configured ProjectNameFunc returns an empty string for the task's ProjectID
- **THEN** no `+...` chip is shown for that row

## MODIFIED Requirements

### Requirement: Chip suppression by status
Due and defer chips signal actionability and SHALL be suppressed on done and dropped tasks. The assignee chip SHALL still render on done tasks; dropped tasks SHALL show no chips. The project chip follows the assignee rule: rendered on open and done tasks, suppressed on dropped tasks.

#### Scenario: Done task hides date chips but keeps assignee
- **WHEN** rendering a done task that has a due date and an assignee "alice"
- **THEN** no `due:`/`overdue:`/`defer:`/`ready:` chip is shown, and `@alice` is shown

#### Scenario: Done task keeps project chip
- **WHEN** rendering a done task with a non-nil ProjectID
- **THEN** the `+<project>` chip is shown

#### Scenario: Dropped task hides all chips
- **WHEN** rendering a dropped task that has a due date, an assignee, and a project
- **THEN** no chips are shown

### Requirement: Chip ordering and alignment
When a task shows multiple chips, they SHALL be left-aligned and placed inline after the title in the order: due/overdue, then defer/ready, then assignee, then project.

#### Scenario: Multiple chips on one row
- **WHEN** a pending task has a due date 28 days out and a defer date 14 days out
- **THEN** the row reads `[ ] <title>  due:28d  defer:14d` with chips left-aligned in that order

#### Scenario: Full chip set with project
- **WHEN** a pending task in project "Kitchen remodel" is due in 3 days, assigned to alice, with no defer
- **THEN** the row reads `[ ] <title>  due:3d  @alice  +Kitchen remodel`

### Requirement: Urgency colors
Chips SHALL be colorized to convey urgency, using huh theme colors where available and static colors otherwise:

- `overdue:` → red (highest urgency)
- `due:` today → orange
- `due:` 2–6 days → yellow
- `due:` 7+ days → dim/neutral
- `defer:` → dim blue (low urgency)
- `ready:` → teal/cyan (attention without alarm; visually distinct from overdue red)
- `@assignee` → magenta (visually distinct from all date chips)
- `+project` → a color distinct from all of the above (green/cyan family)

#### Scenario: Project chip uses a color distinct from other chips
- **WHEN** a row renders a project chip alongside any other chip
- **THEN** the project chip's foreground color is visually distinguishable from due/defer/ready/overdue/assignee chips