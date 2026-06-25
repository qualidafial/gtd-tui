## Context

Status changes are currently split across `space` (complete/reopen toggle, relabeled per status), `delete` (drop, valid only from open), and — for projects — `s` (park). The `space` binding is overloaded: it reopens from done, dropped, and park, and its help label must be recomputed on every selection change via `SetHelp`. Two near-identical per-transition overlays already exist (`tui/pages/tasks/taskstatus` and `tui/pages/projects/projectstatus`), each constructed with a fixed transition (Complete/Drop/etc.) plus a confirmation timestamp editor.

Linear's model has no toggle: a single `s` opens a status picker listing every status, and selection drives the transition. Tasks have three statuses (open/done/dropped); projects have four (open/someday/done/dropped). All underlying transitions already exist as service methods.

This document captures the intended approach. Implementation is not yet scheduled.

## Goals / Non-Goals

**Goals:**
- Replace the per-status `space`/`delete`/`s`-park keys with one `s` → status picker for both tasks and projects.
- Eliminate all state-dependent relabeling of status keys.
- Reuse the existing confirmation-overlay + editable-timestamp flow for the actual transition.
- Align the surrounding action keys with Linear muscle memory (`c` create, `shift+c` convert, `f` filter, `a`/`i` assign, `j`/`k` nav) without adding chord/command-palette machinery.

**Non-Goals:**
- `g`-chord navigation, `ctrl+k` command palette, and `x` multi-select (deferred; multi-select notably collides with `shift+arrows` reorder and must be resolved separately).
- Any domain or service change — transitions reuse `CompleteTask`/`ReopenTask`/`DropTask`/`ParkProject`/`ReopenProject`.
- A real search feature; `/` is only *reserved* (left unbound or stubbed) for it.

## Decisions

### Decision: One status-picker overlay parameterized by entity status set
Generalize the two existing transition overlays into a single picker that takes the current status and the ordered set of statuses, renders them as a select list **including the current status preselected**, and on confirm runs the matching service transition through the existing timestamp-confirmation step.

- **Why:** The transition logic, confirmation, and timestamp editor are already shared in spirit across `taskstatus`/`projectstatus`; the only real variation is the status set and which service method each target maps to. A picker is a thin list in front of the same confirmation flow.
- **Why preselect the current status (not hide it):** Showing the current status as the highlighted entry keeps the present state visible and forces the user to deliberately arrow to a *different* status before pressing enter. Hiding it would turn a done/dropped task — whose only real target is open — into a blind single-option confirm, where the selectfield carries no information. Selecting the already-current status is a no-op (treated as cancel).
- **Alternatives considered:** (a) Hide the current status and show only valid targets — rejected per the above; degenerates to a one-item list. (b) Keep separate keys per transition but drop the toggle — rejected: still state-gated, still consumes `space`/`delete`/`s`, doesn't remove relabeling. (c) Direct number keys (`ctrl+alt+1-9` like Linear) — rejected: opaque, no discoverability in a TUI help bar.

### Decision: Picker maps target status → existing service call
The picker does not introduce new transitions. open→done = Complete, done/dropped→open = Reopen, open→dropped = Drop, open→someday (project) = Park, someday→open = ReopenProject. The picker only chooses which existing call to issue; invalid/no-op selections (the current status) are not offered.

- **Why:** Keeps the change purely presentational and avoids touching tested service semantics.

### Decision: `s` is never overloaded by context; in the flat project view it operates on the selected task
Guiding principle (matching Linear, which never shares the `s` key): **`s` always means "status of the currently-focused entity."** Linear keeps bare `s` = issue status everywhere and gives project status a distinct chord (`p` then `s`, consistent with `p`-prefixed project properties such as project priority = `p` `p`). It never reinterprets `s` against context.

The project view embeds a scoped task list whose selected row is a *task*, so the focused entity there is a task. Therefore `s` in the project view SHALL fall through to the embedded task list and change the **selected task's** status — the same as the task list everywhere else. The project view SHALL NOT intercept `s`. This needs no project-view code change: the embedded task list's new `s` binding simply works.

Changing the **project's own** status from within the project view is **deferred** (out of scope for this change). Two future paths, to be chosen in a dedicated proposal:
- **Tabbed project view** (preferred, mirrors Linear's Overview / Activity / Issues tabs): each tab establishes the focused entity, so `s` = project status on Overview/Activity and `s` = task status on the Issues/tasks tab. No overloading — the same "focused entity's status" rule. This is a larger change (tab chrome + focus model; the task view is likewise still "a single bare panel… no tab chrome").
- **Dedicated project-status binding** on the flat view (e.g. a `p`-prefixed chord like Linear, or a single distinct key), leaving `s` for the task list. Smaller, but introduces chord machinery that is otherwise out of scope here.

Until then, project status is changed from the project list (one `esc` away from the project view).

### Decision: Bundle the Linear key remaps with the picker, but keep them mechanically independent
Create `c` (with `insert` alias), convert `shift+c`, filter `f`, assign `a`/`i`, nav `j`/`k`. These are simple `WithKeys` edits in each `keymap.go` plus a new assign handler.

- **Why:** Shipping the status model without aligning the neighbors (`space` freed, `s` repurposed) would leave a confusing half-Linear layout. They are grouped for coherence but carry no dependency on the picker, so they can be split out if the picker slips.
- **`a`/`i` rationale:** `Assignee` already exists on Task and is only settable inside the clarify wizard today; exposing it on the list/view is additive.

## Risks / Trade-offs

- [`space` muscle memory loss for quick-complete] → Quick-complete becomes `s`, arrow to target, enter. The current status is preselected (not the target), so it is deliberately a few keystrokes; accepted as the cost of an explicit, status-visible picker.
- [`f`/`/` swap retrains existing filter habit] → Document in README; `/` left reserved so it doesn't silently do the wrong thing.
- [Picker is more keystrokes than a toggle] → Accepted: the uniformity and removal of relabeling/overloading is the point; the default-highlight keeps the common path short.
- [Scope creep from bundling remaps] → Remaps are isolated `keymap.go` edits; if time-constrained, ship the picker first and remaps second.

## Resolved Questions

- **`delete` stays as a direct drop shortcut** alongside the picker. It remains valid only where drop is valid (open tasks; open/someday projects) and is a fast path to the `dropped` target the picker also offers. The picker is the discoverable, uniform route; `delete` is the power-user shortcut.
- **The task view adopts `s`** to change its task's status. The **project view does not get a project-status key** in this change: `s` there falls through to the embedded task list (selected task's status). Project-status-from-project-view is deferred to a future tabbed-view change (see Decisions).
- **`space` is left unbound**, reserved for a future peek/preview action. It is not wired to anything in this change.
