## Context

Tasks have a `ProjectID` foreign key and the service layer supports filtering and cascading by project, but the TUI has no surface for assigning tasks to projects or viewing a project's tasks. The project list tab just landed. This change adds the project view screen and the project picker overlay to close the loop.

The task and project TUI packages currently have no imports between them. Both import domain types from the root `gtd` package. The overlay stack (PushMsg/DismissMsg/InitMsg) handles navigation between screens.

## Goals / Non-Goals

**Goals:**
- Allow assigning/unlinking a task's project from the task list via a standalone picker overlay
- Allow viewing a project's tasks by entering the project view from the project list
- Allow creating tasks scoped to a project from the project view
- Reuse the existing tasklist.Model inside the project view without duplication
- Keep task packages free of project package imports

**Non-Goals:**
- Project field in the task edit form (deferred — assignment happens via picker overlay)
- Project chip on task list rows (deferred)
- Inline "create new project" from the picker (deferred)
- Project edit overlay (separate change)
- Project query filter (separate change)

## Decisions

### Project picker is a standalone overlay, not a form field

The picker is its own overlay pushed by `p` from the task list (and future task view). It uses `huh.Select` over open projects plus a "(none)" option. On confirm it calls `TaskService.UpdateTask` to set/clear `ProjectID`, then dismisses.

Alternatives considered:
- **huh.Select field inside taskedit**: requires threading `ProjectService` into task packages, creating conceptual coupling. Also runs into init-on-return issues if we later add inline project create (pushing an overlay on top of an editor would clobber dirty state on dismiss).
- **Autocomplete/filterable custom field**: more work, and huh's built-in `/` filtering on Select already handles large lists.

### Task packages remain project-unaware via factory injection

The tasklist receives a `func(gtd.Task) screen.Screen` factory for the project picker. The caller (app.go or project view) wires this to `projectpicker.New(...)`. The tasklist calls the factory and pushes the result — it never imports project packages.

Alternatives considered:
- **Shared component package**: would work but puts project-specific UI in a generic location. The factory approach keeps ownership clear.
- **Direct import**: creates circular dependency risk since project view imports tasklist.

### Project view embeds tasklist.Model with a scoped TaskService wrapper

A `projectTaskService` wrapper implements `gtd.TaskService`, injecting the project's ID into `ListTasks` (always adds `ProjectID` filter) and `CreateTask` (always stamps `ProjectID`). All other methods delegate unchanged. The wrapper lives in the service layer since it's a service concern, not a presentation concern.

The project view constructs a `tasklist.New(wrappedSvc, "")` with an empty default query (show all project tasks regardless of status).

Alternatives considered:
- **Separate list implementation inside project view**: duplicates rendering, keybindings, query bar logic.
- **Wrapper in the TUI layer**: possible, but scoping service behavior belongs in the service layer.

### Create key changes from `n` to `+`/`insert` everywhere

All "create" actions (new task, new project) use `+`/`insert` instead of `n`. This is a consistency fix applied to task list and project list as part of this change.

### Project picker overlay lives in `tui/pages/projects/projectpicker/`

It has access to `ProjectService` (to list open projects) and `TaskService` (to update the task). Self-contained: loads projects, presents select, saves assignment, dismisses. Matches the pattern of `taskstatus` and `projectstatus`.

## Risks / Trade-offs

### Init-on-return and editor state

The overlay stack calls `Init()` on the parent when a child is dismissed. This is fine for lists (they reload) but would destroy dirty editor state. This change doesn't trigger the issue (picker sits on lists, not editors), but it constrains future work. Noted as a known concern — must be resolved before overlays can stack on editors.

### ProjectService threading to tasklist

The tasklist now needs a `ProjectService` (to build picker options) threaded from app.go. This is one more constructor parameter. Acceptable for now but signals that a services struct or context injection may be warranted if more cross-cutting concerns emerge.

### Project list package location

The project list currently lives in `tui/pages/projects/` (package `projects`) rather than its own `tui/pages/projects/projectlist/` package like the task list does (`tui/pages/tasks/tasklist/`). This inconsistency should be resolved in a future cleanup — move to `tui/pages/projects/projectlist/` to match the task list convention.

### Project view header height reduces task list space

The project detail header (title, status, outcome, due) consumes vertical space above the embedded task list. Non-empty fields only are shown to minimize this. The task list receives the remaining height via WindowSizeMsg.