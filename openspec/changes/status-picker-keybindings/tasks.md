## 1. Status picker overlay

- [x] 1.1 Generalize `taskstatus`/`projectstatus` into a shared status-picker overlay that takes the current status plus an ordered list of selectable statuses (current included)
- [x] 1.2 Render the picker as a selectfield with the current status preselected; exclude statuses not reachable from the current status
- [x] 1.3 Map each selectable target to its service transition (Complete/Reopen/Drop; project Park/ReopenProject) and apply with the inline editable-timestamp `When` field (folded into the picker, shown once the selection differs from the current status)
- [x] 1.4 Treat confirming on the current (unchanged) status as a no-op that dismisses; preserve cascade info in the confirmation for project Complete/Drop
- [x] 1.5 Unit-test the picker: option sets per current status (task open/done/dropped, project open/someday/done/dropped), no-op on unchanged, transition mapping

## 2. Task status binding (`s`)

- [x] 2.1 Replace `tasklist.KeyMap.ToggleComplete` with a single `Status` binding (`s`, fixed label `status`); enable when a task is selected. Keep `Drop` (`delete`) as-is; leave `space` unbound
- [x] 2.2 Wire `s` in the task list to push the status picker seeded from the selected task; remove the `space` toggle handler and the `SetHelp` relabeling (delete→drop stays)
- [x] 2.3 Replace `taskview.KeyMap` `ToggleComplete` with `Status` (`s`); push the picker from the task view. Keep `delete`→drop; leave `space` unbound
- [x] 2.4 Update task list and task view tests for `s`-driven status changes

## 3. Project status binding (`s`)

- [x] 3.1 Replace `projectlist.KeyMap` `ToggleComplete` and `Park` with a single `Status` binding (`s`, fixed label `status`); keep `Drop` (`delete`)
- [x] 3.2 Wire `s` to push the status picker seeded from the selected project (open→someday/done/dropped, someday→open/dropped, done/dropped→open); remove the `space` toggle and `s`-park handlers and relabeling (delete→drop stays)
- [x] 3.3 Update `updateKeybindings` so status is enabled on any selection, `delete` stays gated to open/someday, and reorder stays state-gated; drop the space/`s`-park enable logic
- [x] 3.4 Confirm the project view does NOT intercept `s` (no `projectview.KeyMap` status binding); verify `s` falls through to the embedded task list and changes the selected task's status. Project-status-from-project-view is deferred to a future tabbed project view
- [x] 3.5 Update project list and project view tests (project view: assert `s` operates on the selected task, not the project)

## 4. Linear-aligned key remaps

- [x] 4.1 Rebind create from `+`/`insert` to `c` (keep `insert` alias) across task list, project list, and inbox; update help labels
- [x] 4.2 Rebind Convert from `c` to `shift+c` on task list, task view, project list, and project view
- [x] 4.3 Dropped — keep `/` as the filter-focus key (decision: do not rebind to `f`); specs amended accordingly
- [ ] 4.4 Deferred — Add `a` (assign to person) and `i` (assign to me) bindings on the task list and task view, setting `Task.Assignee` via the service (no person-picker overlay or current-user concept exists yet; revisit in a follow-up change)
- [x] 4.5 Add `j`/`k` as down/up navigation aliases on both lists (already provided by the `bubbles/v2/list` default keymap — `CursorUp` binds `up`/`k`, `CursorDown` binds `down`/`j`)
- [x] 4.6 Verify no key collisions remain in any keymap (`s`, `c`, `f`, `a`, `i`, `j`, `k`)

## 5. Docs and verification

- [x] 5.1 Update README keybindings table and any keymap-resolution references
- [x] 5.2 Run `go test ./...` (all green) and exercise the picker flow for a task and a project via screen-level tests (`s` → picker → select → confirmation; no-op on unchanged). Interactive run pending a TTY (binary builds cleanly)
