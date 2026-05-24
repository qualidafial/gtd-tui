## ADDED Requirements

### Requirement: Status marker and title styling
Each task row SHALL begin with a status marker followed by the task title. The marker and title styling SHALL reflect the task status:

- `pending` → `[ ]`, default title style.
- `done` → `[x]`, title rendered in a dim "done" color.
- `dropped` → `[-]`, title rendered with strikethrough in a dim "dropped" color.

#### Scenario: Pending task marker
- **WHEN** rendering a pending task titled "Buy milk"
- **THEN** the row reads `[ ] Buy milk` with the default title style

#### Scenario: Done task marker and color
- **WHEN** rendering a done task
- **THEN** the marker is `[x]` and the title is rendered in the dim done color

#### Scenario: Dropped task marker and strikethrough
- **WHEN** rendering a dropped task
- **THEN** the marker is `[-]` and the title is rendered with strikethrough in the dim dropped color

### Requirement: Relative-time WHEN formatting
Date chips SHALL render the date as a relative-time WHEN string computed against the current local time. Day counts SHALL be calendar-day differences in the local timezone, not 24-hour spans. Two ladders apply depending on whether the reference instant is in the future or the past.

The future ladder (used by `due:` and `defer:`):

- today, with a time-of-day component (not midnight), still ahead → clock time, e.g. `3pm`
- today, date-only (midnight) → `today`
- exactly one calendar day ahead → `tomorrow`
- 2–6 calendar days ahead → lowercase weekday name, e.g. `thursday`
- 7–30 days ahead → `Nd`, e.g. `30d`
- more than 30 days ahead → absolute date `YYYY-MM-DD`

The past ladder (used by `overdue:` and `ready:`):

- earlier today, with a time-of-day component → clock time, e.g. `3pm`
- 1–30 days ago → `Nd`, e.g. `3d`
- more than 30 days ago → absolute date `YYYY-MM-DD`
- the past ladder SHALL NOT use `tomorrow` or weekday names

#### Scenario: Future day bands
- **WHEN** a date is 1, 3, 7, and 30 calendar days ahead respectively
- **THEN** the WHEN strings are `tomorrow`, the weekday name, `7d`, and `30d`

#### Scenario: Timed today, still ahead
- **WHEN** a date is today at 15:00 and the current time is earlier than 15:00
- **THEN** the WHEN string is `3pm`

#### Scenario: Beyond 30 days shows absolute date
- **WHEN** a date is 31 calendar days ahead, falling on 2026-06-24
- **THEN** the WHEN string is `2026-06-24`

#### Scenario: Past day band
- **WHEN** a date was 3 calendar days ago
- **THEN** the WHEN string is `3d`

### Requirement: Due and overdue chip
A pending task with a due date SHALL display a due chip. The chip word and reference instant depend on the due timestamp:

- A date-only due date applies at **end of the local day**; a timed due date applies at its exact instant.
- While the reference instant is in the future, the chip is `due:<WHEN>` using the future ladder.
- Once the reference instant has passed, the chip is `overdue:<WHEN>` using the past ladder.

#### Scenario: Date-only due today is not overdue
- **WHEN** a pending task is due today (date-only) and the current time is 17:00
- **THEN** the chip reads `due:today` (date-only due applies at end of day, so it is not yet overdue)

#### Scenario: Date-only due flips to overdue next day
- **WHEN** a pending task is due on the 27th (date-only) and today is the 28th
- **THEN** the chip reads `overdue:1d`

#### Scenario: Timed due earlier today is overdue
- **WHEN** a pending task is due today at 15:00 and the current time is 17:00
- **THEN** the chip reads `overdue:3pm`

### Requirement: Defer and ready chip
A pending task with a defer date SHALL display a defer chip. The chip word and reference instant depend on the defer timestamp:

- A date-only defer date applies at **start of the local day**; a timed defer date applies at its exact instant.
- While the reference instant is in the future, the chip is `defer:<WHEN>` using the future ladder.
- Once the reference instant has passed, the task has resurfaced and the chip is `ready:<WHEN>` using the past ladder.

#### Scenario: Future defer
- **WHEN** a pending task is deferred until the 27th (date-only) and today is the 26th
- **THEN** the chip reads `defer:tomorrow`

#### Scenario: Defer flips to ready at start of day
- **WHEN** a pending task is deferred until the 27th (date-only) and today is the 27th
- **THEN** the chip reads `ready:today` (date-only defer applies at start of day)

#### Scenario: Resurfaced defer counts days since
- **WHEN** a pending task was deferred until the 27th (date-only) and today is the 28th
- **THEN** the chip reads `ready:1d`

### Requirement: Assignee chip
A task with a non-empty assignee SHALL display an `@<assignee>` chip.

#### Scenario: Delegated task shows assignee
- **WHEN** rendering a task assigned to "bob"
- **THEN** the row includes the chip `@bob`

### Requirement: Chip suppression by status
Due and defer chips signal actionability and SHALL be suppressed on done and dropped tasks. The assignee chip SHALL still render on done tasks; dropped tasks SHALL show no chips.

#### Scenario: Done task hides date chips but keeps assignee
- **WHEN** rendering a done task that has a due date and an assignee "alice"
- **THEN** no `due:`/`overdue:`/`defer:`/`ready:` chip is shown, and `@alice` is shown

#### Scenario: Dropped task hides all chips
- **WHEN** rendering a dropped task that has a due date and an assignee
- **THEN** no chips are shown

### Requirement: Chip ordering and alignment
When a task shows multiple chips, they SHALL be left-aligned and placed inline after the title in the order: due/overdue, then defer/ready, then assignee.

#### Scenario: Multiple chips on one row
- **WHEN** a pending task has a due date 28 days out and a defer date 14 days out
- **THEN** the row reads `[ ] <title>  due:28d  defer:14d` with chips left-aligned in that order

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
The list selection highlight SHALL affect only the task title (e.g. a cursor indicator and/or emphasis). Data chips SHALL retain their own urgency colors on the selected row.

#### Scenario: Selected row keeps chip colors
- **WHEN** a row with a due chip is selected
- **THEN** the title reflects the selection highlight and the due chip keeps its urgency color
