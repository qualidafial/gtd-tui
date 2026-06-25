## Why

Status changes today are spread across three keys with state-dependent meaning: `space` is a complete/reopen toggle whose label flips by status, `delete` drops (only valid from open), and `s` parks projects. The `space` binding is overloaded — it reopens from done, dropped, *and* (for projects) park — and its label must be relabeled per selection. Adopting Linear's model, where a single `s` opens a **status picker** listing every valid target status, collapses complete/reopen/drop/park/unpark into one affordance and removes all state-dependent relabeling. This is a capture-only proposal; not scheduled for implementation yet.

## What Changes

- **Status picker replaces the toggle.** A single `s` opens an overlay listing all valid target statuses for the selection (task: open/done/dropped; project: open/someday/done/dropped). Selecting one applies that transition through the existing confirmation overlay with editable timestamp.
- **Retire the `space` toggle and `s` park** as status mechanisms. `space` is left unbound (reserved for a future peek/preview); project park becomes the `someday` entry in the picker. **`delete` is kept** as a fast drop shortcut alongside the picker (valid where drop is valid: open tasks; open/someday projects).
- **The task view also adopts `s`** to change its task's status. `s` is never overloaded by context — it always means "status of the focused entity." In the project view (project header + embedded task list) the focused entity is the selected task, so `s` falls through to the embedded list and changes the *task's* status. Changing the **project's own** status from the project view is deferred to a future tabbed project view (Overview/Activity/Issues), where each tab establishes the focused entity; until then project status is changed from the project list.
- **Linear-aligned key remaps** (secondary, bundled for muscle-memory consistency):
  - Create: `+`/`insert` → `c` (insert kept as alias).
  - Convert task↔project: `c` → `shift+c` (freed by create move).
  - Add `a` (assign to person) and `i` (assign to me), exposing the existing `Assignee` field outside the clarify wizard.
  - Add `j`/`k` as down/up navigation aliases.
- **Out of scope** (deferred to later changes): a tabbed project view (Overview/Activity/Issues) and changing project status from the project view; `g`-chord navigation; `ctrl+k` command palette; `x` multi-select.

## Capabilities

### New Capabilities
- `status-picker-overlay`: a status-selection overlay that lists the valid target statuses for a task or project and applies the chosen transition with an inline editable timestamp (the `When` field appears once the selection differs from the current status). Implemented as a picker mode on the existing `taskstatus`/`projectstatus` overlays rather than a separate component.

### Modified Capabilities
- `task-status-ui`: replace the `space` toggle with `s` → status picker; `delete` drop is retained unchanged.
- `task-view-screen`: replace the `space` status key with `s` → status picker; `delete` drop is retained unchanged.
- `project-list-ui`: replace the `space` toggle and `s` park with `s` → status picker; `delete` drop is retained; rebind quick-create from `+`/`insert` to `c` (insert alias kept).

(`project-view-screen` is intentionally unchanged: `s` falls through to the embedded task list with no project-view interception, so there is no spec-level requirement change there.)

The remaining Linear-aligned remaps — convert `c`→`shift+c`, assign `a`/`i`, nav `j`/`k` — are not encoded in any current spec requirement (they are binding-level implementation details); they are described in design.md and applied directly in the relevant `keymap.go` files.

## Impact

- Code: `tui/pages/tasks/taskstatus`, `tui/pages/projects/projectstatus` (add a status-picker mode with an inline editable timestamp), `tui/pages/tasks/tasklist`, `tui/pages/tasks/taskview`, `tui/pages/projects/projectlist`, `tui/pages/projects/projectview`, and the corresponding `keymap.go` files; `tui/pages/inbox` create-key alias.
- Behavior: no domain/service changes — all transitions reuse existing `CompleteTask`/`ReopenTask`/`DropTask`/`ParkProject`/`ReopenProject`. Purely TUI keybinding and overlay restructuring.
- Docs: README keybindings table and any keymap-resolution references.
