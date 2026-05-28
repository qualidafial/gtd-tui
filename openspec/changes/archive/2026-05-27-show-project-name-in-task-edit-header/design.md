## Context

The task edit overlay (`tui/pages/tasks/taskedit`) shows a read-only header with Task ID, Created, Updated, and Status. Tasks may link to a project via `ProjectID *int64`, but the editor doesn't display this relationship.

The taskedit model currently takes only a `gtd.Task` and `gtd.TaskService`. It has no access to project data.

## Goals / Non-Goals

**Goals:**
- Show the linked project's title in the task edit header when the task belongs to a project.
- Keep the taskedit package decoupled from ProjectService.

**Non-Goals:**
- Making the project link editable from the task editor (that's the project picker overlay's job).
- Showing project info for new/unsaved tasks.

## Decisions

**Pass project name as a string to `taskedit.New`**

The caller already has access to ProjectService and knows the task's ProjectID. Resolving the name at construction time keeps the taskedit model simple (no async lookup, no service dependency).

Alternative considered: passing `gtd.ProjectService` into the model and looking up the name in `Init()`. Rejected because it adds coupling and async complexity for a single display string.

**Render "Project" line between "Status" and the blank separator**

This places it logically after the task's own metadata (ID, timestamps, status) and visually groups it with the header block. It only appears when the task has a ProjectID.

## Risks / Trade-offs

- [Stale name] If a project is renamed while the editor is open, the displayed name is stale. → Acceptable: the editor is short-lived and re-opened frequently.
- [Caller responsibility] Every call site that opens the editor must resolve the project name. → Currently there are few call sites; the cost is low.