# Todos

## Design — open questions

- **Completing a project with pending tasks.** `DropProject` already takes a `dropOpenTasks` flag (cascade-drop vs
  detach). Decide whether marking a project `done` needs the same treatment: cascade-complete pending tasks, detach
  them, or refuse and force the user to resolve them first.
- **ClarifyAsProject with a completed first action.** Mirror the "do-it-now" path from `ClarifyAsTask`: allow
  creating a project where the spawned task starts in `done` status, so the UI can immediately prompt for the actual
  next action. Open: is this a parameter on `ClarifyAsProject`, or a follow-on call?

## Design — implementation queue

Order to land the design decisions captured in `DESIGN.md`:

1. **Task** — add `Kind` (next_action/delegated), `Assignee`; collapse `Status` to pending/done/dropped; add
   `ProjectID *int64`; add `comment string` parameter to `UpdateTask` (non-status edits only). Split status
   transitions out of `Update` into per-transition methods: `CompleteTask`, `DropTask` (exists), `ReopenTask`. Each
   carries its own `comment string` for the timeline entry. UI moves status selection out of the edit form into
   list-view actions.
2. **Project** — uncomment `project.go`; add `someday` status; `comment string` parameter on `UpdateProject`
   (non-status edits only). Per-transition methods: `CompleteProject(dropOpenTasks bool, comment string)`,
   `DropProject(dropOpenTasks bool, comment string)`, `ParkProject(comment)` (→someday), `ReopenProject(comment)`
   (→active). Cascade logic for done/dropped lives only in the terminal methods; parking leaves task status
   untouched. Default `TaskFilter` excludes tasks under `someday` projects; tasks under done/dropped projects are
   *not* filtered (invariant: there should be none, and surfacing any is a visible bug).
3. **Item** — add `ClarifiedInto` lineage (nullable FKs per destination type + CHECK).
4. **InboxService** — full clarify surface: `Discard`, `Incubate`, `FileAsReference`, `ClarifyAsTask`,
   `ClarifyAsProject`. Each stamps the Item's `ClarifiedInto` in the same tx as creating the destination.
5. **Comment** — entity + service (dual nullable FKs to Task|Project, CHECK).
6. **Someday** — entity + service (title/description/`ReviewedAt`/timestamps).
7. **Reference** — entity + service (title/body/timestamps).
8. **Meeting** — rename existing `Note` to `Meeting`; struct with title/body/start/end/attendees; `MeetingService`
   with CRUD + `AddActionItem` + link management; `MeetingLink` table (dual nullable FKs to Task|Project|Item).
9. **Delete `project_task.go`** — the m:n join is replaced by `Task.ProjectID`.

## Tasks

- Description markdown should be colorized
    - Ditto when you press ctrl+e to open description in an editor
- esc to cancel task editor conflicts with esc to cancel filtering.
- defer until should only display when status is "deferred"
- waiting does not capture who/what is being waited on.

## General

- How do we test UIs? How much is the effort worth?

## Date field

- Date entry: you have to backspace / delete out any current value to enter in something new
- Natural-language parsing is always-on. Add a `.Natural(true)` opt-in toggle so e.g. `"foo"` doesn't silently resolve to today.
- goja JS engine pulled in by naturaltime adds a few MB to the binary — revisit if size matters.

## Vertical slice scaffolding

All currently commented out in `tui/app.go` and elsewhere:

- Projects screen
- Project tasks
- Notes screen
- Timeline screen
