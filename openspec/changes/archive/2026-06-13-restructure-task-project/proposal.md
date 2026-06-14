## Why

Task-vs-project is decided once, at clarify time, and is irreversible today. When the user realizes after the fact that a "single action" is actually a multi-step outcome — or that an open project is really just one action — there is no way to fix the structure. They are stuck deleting and re-capturing. These restructuring actions are post-clarify "I clarified wrong (or the work grew)" corrections.

## What Changes

- **Convert to Project** (from the task list): promote a standalone task into a new open project. The task's title/description seed the project; the user supplies the project's Outcome and re-scopes the original task down to the reduced "what I can actually do next" first action. Atomic, form-first (the source task is never at risk, so no early checkpoint is needed).
- **Convert to Task** (from project view and project list): collapse an empty open project into a standalone task. Guarded to `open` status with **zero tasks of any status** (the only lossless, reversible case). The project's Outcome is folded into the task Description so nothing is lost.
- **Link Task** (from project view): re-parent an existing standalone task into the current project via a new task-picker overlay. Candidates are restricted to standalone (`ProjectID == nil`) open tasks; the linked task appends to the bottom of the project's task list. Linking into a someday project silently removes the task from default views per the existing `IncludeSomedayProjects` rule.
- Three new transactional service methods spanning the task and project stores: `ConvertTaskToProject`, `ConvertProjectToTask`, `LinkTaskToProject`.
- A new `taskpicker` TUI overlay mirroring the existing `projectpicker`.

## Capabilities

### New Capabilities
- `task-picker-overlay`: a standalone-task selection overlay (mirror of `project-picker-overlay`) used by the Link Task flow, filtered to standalone open tasks.
- `task-project-restructure-ui`: the TUI surfaces and flows for the three restructuring actions — the convert-to-project wizard launched from the task list, the convert-to-task confirm launched from project view and project list, and the link action launched from project view.

### Modified Capabilities
- `project-task-relationship`: add the three service-level restructuring operations (`ConvertTaskToProject`, `ConvertProjectToTask`, `LinkTaskToProject`) and their invariants (standalone-only guards, empty-project guard, open-status guard, field-flow rules, atomicity).

## Impact

- **Service layer**: new methods on `ProjectService`/`TaskService` (or a small set of methods backed by `*sqlite.DB` + `RunTx`); no new orchestrator. Interface additions in `task.go` / `project.go`.
- **TUI**: new `tui/pages/projects/taskpicker/` overlay; new convert-to-project wizard (likely under `tui/pages/tasks/`); keymap + action wiring in the task list, project view, and project list pages.
- **No schema migration**: all three operations are expressible with existing columns (`tasks.project_id`, project/task status, ordering rank). Convert-to-Task deletes a project row and inserts a task row; Convert-to-Project inserts a project row and re-parents a task; Link updates `project_id`.
- **No breaking changes** to existing clarify or query behavior.
