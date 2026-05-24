# GTD TUI — Design Document

## What This Is

A personal productivity app for terminal users, built around the GTD (Getting Things Done) methodology. Everything lives
in one place: tasks, projects, and notes — cross-linked and navigable from a single interface.

## Guiding Principles

**Low ceremony.** The biggest failure mode of GTD systems is the overhead of maintaining them. This app should get out
of the way. Capturing a thought, linking it to a project, or finding what to work on next should take seconds, not
minutes.

**One pane of glass.** Tasks, projects, and notes are not separate tools. They are different views of the same
interconnected data. Navigating between them should feel natural, not like switching apps.

**Easy linking.** Relationships between entities are first-class. When you add action items to a meeting note, they
should surface in the relevant project or inbox automatically. Cross-links should be easy to create and easy to follow.

**Timeline as context.** Every entity has a history. You should be able to see what happened on a project, a task, or a
note — in chronological order — without hunting through separate logs.

## Core Use Cases

### Capture

- Quickly add items to the inbox without friction
- Inbox is the default entry point when it contains unprocessed items

### Clarify (Process Inbox)

```
	                       [ RANDOM INPUT ]
	                              │
	                        What is it?
	                              │
	            ┌─────────────────┴─────────────────┐
	            ▼                                   ▼
	      Is it Actionable?                    NOT Actionable?
	            │                                   │
	            ├───────┐                           ├───────┐
	           Yes      No                         No      Yes
	            │       │                           │       │
	            ▼       ▼                           ▼       ▼
	     [ Multi-step? ] │                    [ Trash ]   [ Someday / Maybe ]
	      (Yes: Project) │                    (Discard)   (Incubate for later)
	            │       │                           │
	            ▼       ▼                           ▼
	  [ What's the Next Action? ]             [ Reference Material ]
	            │                             (Keep for easy retrieval)
	            ├───────┬─────────────┐
	          < 2 min   > 2 min       │
	            │       │             ▼
	            ▼       ▼     [ Delegate to someone else ]
	      [ DO IT ]  [ Defer / ]
	      (Do now)   [ Postpone ]
	                     │
	         ┌───────────┴───────────┐
	         ▼                       ▼
	  [ Calendar ]             [ Next Actions ]
	(Specific date/time)     (Do as soon as possible)
```

- Review inbox items one at a time with a detail panel
- Promote an item to a task with full metadata (kind, due date, defer-until, project)
- Assign a task to a project, or create the project inline
- Park ideas as Someday/Maybe, file references, or discard — all from the inbox detail view
- Clarified items are soft-deleted (lineage preserved) so the timeline shows "captured → clarified into X"

### Organize

- View tasks filtered by status, project, or due date
- Navigate to the task list (next actions) when the inbox is empty
- Projects show their linked tasks and notes

### Engage

- See what to work on next from the task list
- Navigate from a task to its linked notes and project context
- View timeline of activity on any task or project

### Capture Context

- Edit a task or project with an optional one-line comment that gets recorded as a timeline event ("changed status to
  done because…")
- Add a standalone comment to a task or project without entering edit mode — for quick contextual updates that don't
  change any attribute ("blocked on infra ticket," "saw this come up again today")
- Stand-alone meeting records capture title, time slot, attendees, and discussion notes; action items captured during a
  meeting flow to the inbox automatically with a link back to the meeting
- Free-form observations ("hallway conversation," "decision") are deferred until a concrete use case justifies them.
  For now, comments-on-edit and meeting notes cover capture-context needs.

### Reflect

- Activity timelines scoped to a project, task, or note
- Global timeline across all entities
- Because capturing is low-friction, timelines become an accurate record of what actually happened and when — useful for
  retroactive reports, status updates, or just remembering why a decision was made

## Entities

**Item** — an unprocessed capture in the inbox. Title, description, timestamps. The staging area for everything that
hasn't been clarified yet. When clarified, an Item is soft-deleted via a `ClarifiedInto` pointer to whatever it became
(Task, Project, Someday, Reference), so the timeline preserves the lineage: "captured here, clarified into that."
Discarded items are likewise marked rather than hard-deleted.

**Task** — a single actionable item.

- **Kind** — `next_action` (do ASAP) or `delegated` (waiting on someone else; carries an `Assignee` string).
- **Status** — `pending`, `done`, or `dropped`. Inbox-ness, waiting-ness, and deferred-ness are not statuses — they
  are separate entities or fields.
- **Due** — a firm deadline, whether voluntarily committed or externally imposed. There is no "calendar" kind;
  calendar scheduling lives in an actual calendar app.
- **DeferUntil** — a soft, self-imposed "don't show me until" date. Tasks with a future `DeferUntil` are filtered
  out of default views; the underlying status stays `pending`.
- **ProjectID** — nullable. A task belongs to zero or one projects (not many). Standalone tasks have a nil project.

**Project** — a multi-step outcome requiring more than one action. Short title, outcome statement (the desired end
state), description, optional due date.

- **Status** — `active`, `someday` (parked, will revisit), `done`, or `dropped`.
- Terminal transitions (`CompleteProject`, `DropProject`) take a flag to either cascade the new status to pending
  tasks or detach them (set `ProjectID = nil`) and leave them standalone. After a terminal transition, no pending
  tasks should remain under the project — this is an enforced invariant; if it ever breaks, the bug is visible
  (pending tasks under closed projects are *not* filtered out).
- Parking (`ParkProject`) is reversible and does not touch task status. Default task views filter out tasks whose
  project is `someday`; unparking restores them automatically.
- Done/dropped tasks always stay attached to the project as historical record.

**Someday** — a parked idea that isn't yet fleshed out enough to be a Project. Title, description, `ReviewedAt`
(defaults to creation time; used to surface stalest items in periodic review). Promotable to Project or Task once it
crystallizes. Distinct from `Project{Status: someday}`, which is for fully-formed projects that are deliberately parked.

**Reference** — standalone markdown content kept for retrieval (a recipe, a config snippet, a link dump). Title +
body. Not linked to projects or tasks; if attachment matters later, it's an additive change.

**Meeting** — a meeting record. Title, body (markdown discussion notes), required start/end times, attendees (JSON
`[]string` for now; promote to a contacts table if querying by attendee becomes useful). Cross-references projects
and tasks via `MeetingLink`. The meeting body is the source of truth for the original phrasing of any action items
captured during the meeting; spawned inbox items are derivatives with their own editable lifecycle.

**Comment** — short, event-shaped text attached to exactly one Task or Project. Spawned implicitly by edits (the
`comment` parameter on `UpdateTask`/`UpdateProject`) and explicitly via the comment API for context that isn't tied
to a metadata change. Editable so the record can be corrected.

**MeetingLink** — join row connecting a Meeting to a Task, Project, or Item. An Item link exists when a meeting
spawned an inbox action item; the link follows the Item through clarification (it is rewritten to point at the
resulting entity) so the meeting's provenance trail is preserved permanently.

## Clarify Flow

The inbox surfaces Items one at a time. Each clarify operation is a single transaction that creates (or doesn't) a
destination entity and stamps the Item's `ClarifiedInto` pointer:

- **Discard** — non-actionable, not wanted. Item marked discarded.
- **Incubate** — non-actionable, revisit later. Spawns a `Someday`; Item points to it.
- **FileAsReference** — non-actionable, keep for retrieval. Spawns a `Reference`; Item points to it.
- **ClarifyAsTask** — actionable, single step. Spawns a `Task` (kind chosen at clarify time, optionally with a
  project); Item points to it. The "do-it-now" case is `ClarifyAsTask` with `Status: done` recorded immediately,
  preserving the timeline entry.
- **ClarifyAsProject** — actionable, multi-step. Spawns a `Project`; Item points to it. The UI then prompts for the
  first next-action task in the new project. The "do-it-now as first step of a new project" case creates the Project,
  a `done` Task within it, and prompts for the actual next action.

Action items captured in a Meeting use a dedicated `AddActionItem` API on `MeetingService` (not parsed from prose).
The service creates the Inbox Item, links it to the Meeting via `MeetingLink`, and appends a uniform line to the
Meeting body — all in one transaction.

## What This Is Not

- A team tool. This is personal, single-user software.
- A pomodoro timer or time tracker.
- A replacement for a calendar. Due dates are supported, but calendar views are out of scope for now.
