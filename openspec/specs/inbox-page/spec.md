# inbox-page Specification

## Purpose
Defines the inbox screen, its capture overlay, and the progressive clarify wizard that walks the GTD decision tree on a single screen.

## Requirements

### Requirement: Inbox page for unprocessed items
The system SHALL provide an inbox screen at `tui/pages/inbox/` that displays unclarified, non-discarded inbox items. The screen SHALL be a Screen (per `tui-application`) registered as a tab in the root tabContainer alongside Tasks and Projects.

#### Scenario: Display inbox items
- **WHEN** the inbox screen is active
- **THEN** it displays Items returned by InboxService.List in FIFO order (oldest first)

#### Scenario: Empty inbox state
- **WHEN** InboxService.List returns no items
- **THEN** the inbox screen renders an empty-state message rather than an empty list

#### Scenario: Select item for inspection
- **WHEN** the user moves the cursor onto an item
- **THEN** the item's title and description are visible in the active row

#### Scenario: Multiline captures render on a single row
- **WHEN** an item's title or description contains newlines or other whitespace runs
- **THEN** the list row flattens that whitespace to single spaces and renders on exactly one line
- **AND** the row never grows taller than one line, so content below the list (such as the help bar) stays on screen

### Requirement: Inbox tab registration
The system SHALL register an "Inbox" tab in the root tabContainer so the inbox screen is reachable via tab navigation.

#### Scenario: Inbox tab present
- **WHEN** the application starts
- **THEN** the tabContainer SHALL include an "Inbox" tab in addition to "Tasks" and "Projects"

### Requirement: Item capture overlay
The inbox screen SHALL support capturing new items via a new-item overlay matching the structure of the existing taskedit/projectedit overlays. The overlay accepts Title and Description and is opened via `+`/`insert`. Items are write-once: there is no edit overlay; any refinement of wording happens inside the clarify wizard, which pre-fills task/project title and description from the item and lets the user override.

#### Scenario: Open capture overlay
- **WHEN** the user presses `+` or `insert` on the inbox screen
- **THEN** the system pushes the item capture overlay with empty Title and Description fields

#### Scenario: Save new item from capture overlay
- **WHEN** the user fills the Title field and saves
- **THEN** InboxService.Create is invoked with the supplied Title and Description
- **AND** the overlay dismisses
- **AND** the new item appears at the bottom of the inbox list (FIFO)

#### Scenario: Cancel capture overlay
- **WHEN** the user cancels the overlay
- **THEN** no item is created
- **AND** the overlay dismisses

#### Scenario: No item-edit binding
- **WHEN** the user is on the inbox screen with an item selected
- **THEN** no keybinding opens an item-edit overlay
- **AND** `enter` opens the clarify wizard instead (see "Clarify wizard")

### Requirement: Clarify wizard
The inbox screen SHALL provide a clarify wizard that walks the GTD decision tree progressively on a single screen. The wizard is opened via `enter` on a selected item and routes to the appropriate `InboxService` operation based on the user's answers.

The wizard renders all answered questions as a persistent column on the screen; the next unanswered question appears inline beneath the most recent answer. The user can navigate up to a previous answer to change it; doing so discards all wizard state beneath that point and re-asks subsequent questions.

The wizard does NOT use the `huh` form library; the progressive-disclosure UX is implemented directly so that all questions appear on a single screen as they become relevant.

#### Scenario: Wizard opens with item context
- **WHEN** the user presses `enter` on a selected item
- **THEN** the wizard pushes as an overlay
- **AND** the item's Title and Description are shown at the top of the wizard for reference

#### Scenario: Actionable question is the root
- **WHEN** the wizard opens
- **THEN** the first question is "Is it actionable?" with Yes/No options
- **AND** no further questions appear until this is answered

#### Scenario: Non-actionable splits into Trash or Someday
- **WHEN** the user answers "No" to actionable
- **THEN** the wizard reveals a "Trash or Someday?" question
- **AND** "Reference" is not offered (added by `implement-references`)

#### Scenario: Trash branch requires confirmation
- **WHEN** the user picks Trash
- **THEN** the wizard reveals a "Really discard?" confirmation
- **AND** confirming routes to InboxService.Discard

#### Scenario: Someday branch inline-edits project fields
- **WHEN** the user picks Someday
- **THEN** the wizard reveals Title (prefilled from the item) and Description (prefilled from the item) fields
- **AND** committing routes to InboxService.Incubate with the supplied project fields

#### Scenario: Actionable splits on multi-step
- **WHEN** the user answers "Yes" to actionable
- **THEN** the wizard reveals "Does the desired outcome require more than one step?" with Yes/No options

#### Scenario: Single-task branch runs the per-task block then checkpoints
- **WHEN** the user answers "No" to multi-step
- **THEN** the wizard runs the per-task block (next-action title, <2 min, doer-if-not-2min, project-attach)
- **AND** routes to InboxService.ClarifyAsTask creating the task as open
- **AND** if `<2 min` was answered Yes, the wizard then shows the do-it-now confirmation prompt described below

#### Scenario: Project branch defines outcome then loops per-task block
- **WHEN** the user answers "Yes" to multi-step
- **THEN** the wizard reveals project Title (prefilled), Outcome, and Description (prefilled) fields
- **AND** then runs the per-task block to define the first task
- **AND** routes to InboxService.ClarifyAsProject creating the project and the first task (always open) as a single checkpoint
- **AND** if the first task was marked `<2 min`, the wizard shows the do-it-now confirmation prompt for it
- **AND** the per-task loop continues until a task is defined as NOT `<2 min`; that task becomes the project's open next-action and the wizard exits

#### Scenario: Project branch does not ask project-attach per task
- **WHEN** the per-task block runs inside the project branch
- **THEN** the "Belongs to an existing project?" question is omitted (the task auto-attaches to the new project)

### Requirement: Do-it-now confirmation prompt
For any task that the user marks `<2 min`, the wizard SHALL commit the task as open FIRST (via ClarifyAsTask, ClarifyAsProject, or TaskService.CreateTask depending on context), then display a standalone "do it now, then confirm" prompt. On confirmation the wizard SHALL call TaskService.CompleteTask to mark the task done.

This commit-then-prompt order ensures the user does not lose work when accidentally dismissing the wizard mid-loop: every committed task persists, and any do-it-now task that the user did not get around to confirming remains as an open task they can complete from the task list later.

#### Scenario: Single-task do-it-now commits open then prompts
- **WHEN** the user completes the single-task branch with `<2 min` = Yes
- **THEN** the wizard calls ClarifyAsTask creating the task as open
- **AND** displays the do-it-now confirmation prompt
- **AND** on confirm, calls TaskService.CompleteTask on the new task
- **AND** on Esc, the task remains open and the wizard exits silently

#### Scenario: Project first task do-it-now commits open then prompts
- **WHEN** the user completes the project branch's first task block with `<2 min` = Yes
- **THEN** the wizard calls ClarifyAsProject creating the project and first task (open)
- **AND** displays the do-it-now confirmation prompt for the first task
- **AND** on confirm, calls TaskService.CompleteTask, then loops to the next per-task block
- **AND** on Esc, the first task remains open, the item is clarified, and the wizard exits silently

#### Scenario: Subsequent project tasks created via TaskService
- **WHEN** the wizard loops to define a project's second or later task
- **THEN** the wizard calls TaskService.CreateTask with `ProjectID` set to the existing project (created as open)
- **AND** if `<2 min` = Yes, the do-it-now confirmation prompt runs the same as above

### Requirement: Loop exit condition
The project branch's per-task loop SHALL exit when a task is defined as NOT `<2 min` (the actual next physical action). That task is committed as open and the wizard closes; the wizard does NOT ask "define another?"

#### Scenario: First not-do-it-now task ends the loop
- **GIVEN** the project branch's loop is running
- **WHEN** the user answers `<2 min` = No on a task definition
- **THEN** the wizard commits that task as open (via ClarifyAsProject if it is the first task, otherwise TaskService.CreateTask) and exits

### Requirement: Per-task clarify block ordering
The per-task wizard block SHALL ask questions in the following fixed order. The order encodes the GTD rule that sub-two-minute work is always done by the user, never delegated.

1. Next-action title (prefilled from the item title on the first task only; empty for subsequent tasks in a project loop)
2. `<2 minutes? do it now?` (Yes/No)
3. (only if the previous answer was No) Doer: Me / Someone else
4. (only if doer is Someone else) Assignee
5. (single-task branch only) Belongs to an existing project? (Yes/No → project picker)

#### Scenario: Do-it-now skips delegate question
- **WHEN** the user answers Yes to `<2 minutes?`
- **THEN** the wizard does NOT ask the doer question for that task
- **AND** the task is committed with `Status=done` and no assignee

#### Scenario: Delegate captures assignee
- **WHEN** the user answers "Someone else" to doer
- **THEN** the wizard reveals an Assignee text field
- **AND** the task is committed with the supplied Assignee

#### Scenario: Existing-project attach uses the project picker
- **WHEN** the user answers Yes to "Belongs to an existing project?" in the single-task branch
- **THEN** the wizard reveals the existing project picker (no inline project creation)
- **AND** the resulting task has `ProjectID` set to the chosen project

### Requirement: Wizard back-navigation collapses downstream state
The wizard SHALL allow the user to move the cursor up to a previously-answered question and change its answer. When an answer is changed, all wizard state beneath that question SHALL be discarded.

#### Scenario: Changing multi-step from Yes to No discards project state
- **GIVEN** the user is in the project branch with Outcome filled in and one done task defined
- **WHEN** the user navigates up and changes "multi-step?" from Yes to No
- **THEN** the wizard discards the project Outcome, Description, and all task entries
- **AND** the wizard runs the per-task block fresh in the single-task branch