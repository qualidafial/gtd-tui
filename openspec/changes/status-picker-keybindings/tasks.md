## 1. Status picker overlay

- [ ] 1.1 Generalize `taskstatus`/`projectstatus` into a shared status-picker overlay that takes the current status plus an ordered list of selectable statuses (current included)
- [ ] 1.2 Render the picker as a selectfield with the current status preselected; exclude statuses not reachable from the current status
- [ ] 1.3 Map each selectable target to its service transition (Complete/Reopen/Drop; project Park/ReopenProject) and route through the existing editable-timestamp confirmation overlay
- [ ] 1.4 Treat confirming on the current (unchanged) status as a no-op that dismisses; preserve cascade info in the confirmation for project Complete/Drop
- [ ] 1.5 Unit-test the picker: option sets per current status (task open/done/dropped, project open/someday/done/dropped), no-op on unchanged, transition mapping

## 2. Task status binding (`s`)

- [ ] 2.1 Replace `tasklist.KeyMap.ToggleComplete` with a single `Status` binding (`s`, fixed label `status`); enable when a task is selected. Keep `Drop` (`delete`) as-is; leave `space` unbound
- [ ] 2.2 Wire `s` in the task list to push the status picker seeded from the selected task; remove the `space` toggle handler and the `SetHelp` relabeling (delete→drop stays)
- [ ] 2.3 Replace `taskview.KeyMap` `ToggleComplete` with `Status` (`s`); push the picker from the task view. Keep `delete`→drop; leave `space` unbound
- [ ] 2.4 Update task list and task view tests for `s`-driven status changes

## 3. Project status binding (`s`)

- [ ] 3.1 Replace `projectlist.KeyMap` `ToggleComplete` and `Park` with a single `Status` binding (`s`, fixed label `status`); keep `Drop` (`delete`)
- [ ] 3.2 Wire `s` to push the status picker seeded from the selected project (open→someday/done/dropped, someday→open/dropped, done/dropped→open); remove the `space` toggle and `s`-park handlers and relabeling (delete→drop stays)
- [ ] 3.3 Update `updateKeybindings` so status is enabled on any selection, `delete` stays gated to open/someday, and reorder stays state-gated; drop the space/`s`-park enable logic
- [ ] 3.4 Confirm the project view does NOT intercept `s` (no `projectview.KeyMap` status binding); verify `s` falls through to the embedded task list and changes the selected task's status. Project-status-from-project-view is deferred to a future tabbed project view
- [ ] 3.5 Update project list and project view tests (project view: assert `s` operates on the selected task, not the project)

## 4. Linear-aligned key remaps

- [ ] 4.1 Rebind create from `+`/`insert` to `c` (keep `insert` alias) across task list, project list, and inbox; update help labels
- [ ] 4.2 Rebind Convert from `c` to `shift+c` on task list, task view, project list, and project view
- [ ] 4.3 Rebind filter focus from `/` to `f` on task list and project list; leave `/` unbound (reserved for search)
- [ ] 4.4 Add `a` (assign to person) and `i` (assign to me) bindings on the task list and task view, setting `Task.Assignee` via the service
- [ ] 4.5 Add `j`/`k` as down/up navigation aliases on both lists
- [ ] 4.6 Verify no key collisions remain in any keymap (`s`, `c`, `f`, `a`, `i`, `j`, `k`)

## 5. Docs and verification

- [ ] 5.1 Update README keybindings table and any keymap-resolution references
- [ ] 5.2 Run `go test ./...` and manually verify the picker flow for a task and a project
