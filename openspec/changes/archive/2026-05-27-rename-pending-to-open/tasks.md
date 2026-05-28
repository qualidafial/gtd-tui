## 1. Domain layer

- [x] 1.1 Rename `TaskStatusPending` to `TaskStatusOpen` with value `"open"` in `task.go`
- [x] 1.2 Change `Task.Assignee` from `string` to `*string` in `task.go`
- [x] 1.3 Update `StatusChangedAt` comment from "pending" to "open"
- [x] 1.4 Remove `TaskKind` type, constants, `Task.Kind` field, `TaskFilter.Kind`, and `WithKind()` from `task.go`

## 2. SQLite migration

- [x] 2.1 Add migration to recreate tasks table: `status DEFAULT 'open'`, `CHECK(status IN ('open','done','dropped'))`, `assignee TEXT` (nullable), no `kind` column
- [x] 2.2 Migration copies data converting `status='pending'` → `'open'`, `assignee=''` → `NULL`, and dropping `kind`

## 3. SQLite layer

- [x] 3.1 Update `scanTask` to handle nullable assignee (scan into `*string`) and remove kind from column list/scan
- [x] 3.2 Update `CreateTask` to write `*string` assignee and remove kind
- [x] 3.3 Update `UpdateTask` to write `*string` assignee and remove kind
- [x] 3.4 Update `ListTasks` — remove kind filter, handle assignee NULL in filter and search
- [x] 3.5 Update `pendingOrder` and `nextOrderKey` for `TaskStatusOpen`
- [x] 3.6 Update SQLite tests for `TaskStatusOpen`, `*string` assignee, and no kind

## 4. Query parser

- [x] 4.1 Change `taskquery.go`: map `"open"` for status, remove `kind` key and `parseKind`
- [x] 4.2 Update query parser tests

## 5. Service layer

- [x] 5.1 Update `service/project_task_test.go` for `TaskStatusOpen`, `*string` assignee, and no kind

## 6. TUI layer

- [x] 6.1 Update default query in `tui/app.go` from `status:pending ready:now` to `status:open ready:now`
- [x] 6.2 Update `tasklist/render.go` assignee chip check from `!= ""` to `!= nil` and dereference
- [x] 6.3 Update `tasklist/render_test.go` for `TaskStatusOpen`, `*string` assignee, no kind
- [x] 6.4 Update `tasklist/model.go` for `TaskStatusOpen` and remove kind from new-task default
- [x] 6.5 Update `tasklist/model_test.go` for `TaskStatusOpen`, `*string` assignee, no kind
- [x] 6.6 Update `taskedit/model.go`: remove Kind form field, convert `*string` assignee for form binding, default status to open
- [x] 6.7 Update `taskedit/model_test.go` for `*string` assignee and no kind
- [x] 6.8 Update `taskstatus/transition.go` for `TaskStatusOpen`
- [x] 6.9 Update `taskstatus/model_test.go` for `TaskStatusOpen`
- [x] 6.10 Update `projects/projectstatus/model.go` if it references `TaskStatusPending`
- [x] 6.11 Update `projects/render.go` and `render_test.go` for `TaskStatusOpen`
- [x] 6.12 Update `projects/projectpicker/model_test.go` for `TaskStatusOpen` and no kind
- [x] 6.13 Update `projects/projectview/model_test.go` for `TaskStatusOpen` and no kind

## 7. Verify

- [x] 7.1 Run full test suite, confirm no remaining references to `TaskStatusPending`, `"pending"` status string, `TaskKind`, or `kind` column