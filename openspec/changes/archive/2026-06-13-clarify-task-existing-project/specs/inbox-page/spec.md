## MODIFIED Requirements

### Requirement: Per-task clarify block ordering
The per-task wizard block SHALL ask questions in the following fixed order. The order encodes the GTD rule that sub-two-minute work is always done by the user, never delegated.

1. Next-action title (prefilled from the item title on the first task only; empty for subsequent tasks in a project loop)
2. `<2 minutes? do it now?` (Yes/No)
3. (only if the previous answer was No) Doer: Me / Someone else
4. (only if doer is Someone else) Assignee
5. (single-task branch only, and only when at least one open project exists) Project: an optional select of open projects with a `(none)` option, defaulting to `(none)` (standalone task). When shown it appears for every single task, including the sub-two-minute do-it-now path, and is independent of the `<2 min`/doer answers. When there are no open projects the field is omitted entirely.

#### Scenario: Do-it-now skips delegate question
- **WHEN** the user answers Yes to `<2 minutes?`
- **THEN** the wizard does NOT ask the doer question for that task
- **AND** the task is committed with `Status=done` and no assignee

#### Scenario: Delegate captures assignee
- **WHEN** the user answers "Someone else" to doer
- **THEN** the wizard reveals an Assignee text field
- **AND** the task is committed with the supplied Assignee

#### Scenario: Existing-project attach uses an optional select
- **GIVEN** at least one open project exists
- **WHEN** the single-task branch runs
- **THEN** the wizard shows a Project select listing the open projects plus a `(none)` option
- **AND** the select defaults to `(none)`

#### Scenario: Selecting a project sets ProjectID
- **WHEN** the user picks an open project in the single-task branch
- **THEN** the resulting task has `ProjectID` set to the chosen project
- **AND** the wizard routes to `InboxService.ClarifyAsTask` unchanged

#### Scenario: Leaving the select on none yields a standalone task
- **WHEN** the user leaves the Project select on `(none)`
- **THEN** the resulting task has a nil `ProjectID`

#### Scenario: Do-it-now task can still belong to a project
- **WHEN** the user marks a single task `<2 min` = Yes and also picks a project
- **THEN** the task is committed open with the chosen `ProjectID`, the do-it-now prompt runs, and on confirm the task is completed via `TaskService.CompleteTask`

#### Scenario: Attaching to an existing project does not start the project loop
- **WHEN** the user attaches a single task to an existing project
- **THEN** the wizard commits the one task and dismisses (or runs the do-it-now prompt) without entering the per-task project loop

## ADDED Requirements

### Requirement: Clarify wizard loads open projects
The clarify wizard SHALL load the list of open projects from ProjectService before presenting the single-task block, so the single-task branch's Project select can be populated. Until the projects have loaded the wizard SHALL show a loading indicator and SHALL NOT capture input.

#### Scenario: Wizard loads open projects before the single-task block
- **WHEN** the wizard opens
- **THEN** it loads the open projects from ProjectService
- **AND** shows a loading indicator until they are available
- **AND** populates the single-task branch's Project select from the loaded list

#### Scenario: No open projects hides the Project select
- **WHEN** the wizard opens and there are no open projects
- **THEN** the single-task branch does not show the Project select
- **AND** a single task created from this branch is standalone
