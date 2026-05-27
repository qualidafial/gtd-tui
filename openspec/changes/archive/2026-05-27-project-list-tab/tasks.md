## 1. Service and data wiring

- [x] 1.1 Uncomment ProjectService in service/project.go and verify it compiles
- [x] 1.2 Add CountTasksByProjects(ctx, []int64) (map[int64]ProjectTaskCounts, error) to sqlite layer — counts pending and non-dropped tasks per project in a single query
- [x] 1.3 Add ProjectTaskCounts type to domain (Pending int, Total int)
- [x] 1.4 Wire ProjectService into cmd/gtd/main.go and tui.New (accept gtd.ProjectService, add Projects tab to tab container)

## 2. Project list model scaffold

- [x] 2.1 Implement projectlist.Model with bubbles/list, ProjectService dependency, Init that loads projects via ListProjects (no filter), and ProjectsLoadedMsg
- [x] 2.2 Load task counts in batch after projects load (TaskCountsLoadedMsg), associate with project rows
- [x] 2.3 Implement WindowSizeMsg handling and View rendering

## 3. Project row rendering

- [x] 3.1 Implement project list delegate with status markers ([ ] open, [?] someday, [x] done, [-] dropped)
- [x] 3.2 Implement title styling per status (default, dimmed, faint, faint+strikethrough) with selected-row bold and cursor prefix
- [x] 3.3 Implement task progress chip ("3/5 tasks", hidden when 0 total) and "needs action" warning chip (teal, open projects with 0 pending tasks)
- [x] 3.4 Implement due/overdue chip with urgency coloring (reuse task chip palette), suppressed for done/dropped
- [x] 3.5 Implement title truncation preserving chips

## 4. Quick-create project

- [x] 4.1 Implement project create overlay (title-only text input, creates open project on enter, dismiss on esc, reject empty title)
- [x] 4.2 Add "n" keybinding to push create overlay

## 5. Status transitions

- [x] 5.1 Implement project status confirmation overlay (reuse taskstatus pattern) for complete and drop with cascade info text
- [x] 5.2 Add space keybinding: complete (with confirmation) for open, reopen (immediate) for someday/done/dropped
- [x] 5.3 Add delete keybinding: drop (with confirmation) for open/someday, disabled for done/dropped
- [x] 5.4 Add "s" keybinding: park (immediate) for open, disabled for other statuses

## 6. Reordering

- [x] 6.1 Implement shift+up/shift+down for open/someday projects (MoveProjectUp/Down), reload with cursor tracking
- [x] 6.2 Disable move bindings at boundaries and for done/dropped

## 7. Keybindings and help

- [x] 7.1 Implement keyMap struct with per-selection enable/disable logic based on project status
- [x] 7.2 Implement KeyMap() help.KeyMap and CapturingInput() for tab container integration

## 8. Tests

- [x] 8.1 Test CountTasksByProjects query (pending/total excluding dropped)
- [x] 8.2 Test project row rendering (status markers, chips, truncation, warning chip)
- [x] 8.3 Smoke-test project list model (load, create, transitions, reorder)