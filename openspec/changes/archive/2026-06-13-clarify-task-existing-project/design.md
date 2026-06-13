## Context

`InboxService.ClarifyAsTask` already accepts `task.ProjectID` and validates the project exists inside its transaction. The only gap is the wizard UI: the clarify single-task branch never exposes a project selector, even though the `inbox-page` spec calls for one. This design covers how to surface that selector inside the existing form-based wizard.

## Decisions

### 1. Reuse `selectfield`, embedded inline (not the `projectpicker` overlay)

The `projectpicker` overlay (`tui/pages/projects/projectpicker`) already attaches a project to a task via the `selectfield` component, but it is a self-contained screen that calls `TaskService.UpdateTask` and dismisses. Reusing it mid-clarify would mean a second write (create open task, then update its project) and an awkward push/pop.

Instead, embed the same `selectfield` directly in the clarify initial form's single-task block:

```go
selectfield.New("project", "Project", openProjectOpts,
    selectfield.WithNone[int64]("(none)"))
```

The chosen ID flows through `taskFromVals` into `task.ProjectID`, and the single existing `ClarifyAsTask` call does the rest. One write, no extra screen.

### 2. Loading phase before the form (chosen over build-empty-then-inject)

The clarify form is built synchronously in `clarify.New`, but the open-project list requires an async `ProjectService.ListProjects`. Two options were considered:

- **Loading phase (chosen):** `clarify.New` no longer builds the form eagerly. `Init` issues a `loadProjectsCmd`; until it returns the view shows `Loading…`. On the loaded message the wizard builds the initial form with the project options in hand and sets `ready = true`. This mirrors the proven `projectpicker` `ready` pattern and keeps a single, fully-formed form.
- **Build empty + inject:** build the form immediately with no options and patch/rebuild the `selectfield` when projects arrive. Rejected — rebuilding mid-edit risks losing field state, and patching a field's options after construction is not part of the form component's contract.

Cost: the whole wizard (including the trash/someday/new-project branches that don't need projects) waits on one local SQLite query. That query is sub-millisecond, so blocking is acceptable for the simpler, correct single-form model.

### 3. Single optional select, default `(none)` — refine the spec's Yes/No gate

The current `inbox-page` spec describes step 5 as "Belongs to an existing project? (Yes/No → project picker)" — a boolean gate that then reveals a picker. A single `selectfield` whose default option is `(none)` collapses that into one field: leaving it on `(none)` yields a standalone task, picking a project attaches it. This is fewer keystrokes and matches the `projectpicker` overlay exactly. The spec is updated to match.

Field order follows the spec's existing per-task numbering: the project select is the last field in the single-task block (after the optional assignee). It is independent of the `<2 min`/doer answers, so it renders on the do-it-now path too.

### 4. Single task stays single — no loop after attach

Attaching to an existing project does NOT enter `phaseProjectLoop`. That loop exists to capture the *first and subsequent* next-actions of a freshly-created project as a single checkpoint. An existing project already has its own task-add affordances in the project view; clarify's job here is to file this one action. After `ClarifyAsTask` returns, the wizard dismisses (or runs the do-it-now prompt for a sub-2-minute task) exactly as it does for a standalone task today.

### 5. `taskFromVals` must tolerate the absent project key

`taskFromVals` is shared by the initial form and the loop form. The loop form has no `"project"` field, so the lookup must default to "no project". A missing map key type-asserts to the zero `int64` (0), which the helper treats as standalone — so the behavior is already safe, but the helper reads the key explicitly and comments the default so a future loop-form change cannot silently attach loop tasks to a stray value.

## Risks / Trade-offs

- **Loading flash:** very large project lists could make the `Loading…` phase briefly visible. Acceptable; the query is local and the someday/trash/project branches tolerate the same wait.
- **Spec divergence resolved, not hidden:** the Yes/No → select change is a deliberate UX refinement recorded in the `inbox-page` delta, not an undocumented deviation.
