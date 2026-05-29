## 1. Shared ParseError and query bar component

- [x] 1.1 Create `tui/components/querybar/` package with `ParseError` struct (`Message`, `Start`, `End`; implements `error`), `ApplyMsg`, `CancelMsg` types
- [x] 1.2 Create `querybar.Model` with `textinput.Model`, `ValidateFunc func(string) *ParseError`, configurable debounce interval, focus/blur (trailing space on focus, trim on blur), apply/cancel/debounce logic
- [x] 1.3 Implement single-line `View()` with `ansi.Cut`-based inline error highlighting on the offending range
- [x] 1.4 Add tests for querybar: apply success, apply failure with error highlight, cancel reverts, debounce validation, focus appends space, trim on apply

## 2. Project query parser

- [x] 2.1 Add `Search []string` field to `ProjectFilter` in `project.go`
- [x] 2.2 Implement free-text search in `sqlite/project.go` `ListProjects` using `likeContains` on title, outcome, description
- [x] 2.3 Create `internal/projectquery/` package with `Parse(string) (gtd.ProjectFilter, error)` supporting `status:` (open/someday/done/dropped) and free-text tokens, returning `*querybar.ParseError` on invalid values
- [x] 2.4 Add tests for projectquery: empty query, status filter, free-text search, invalid status error with range, unrecognized key as free text

## 3. Migrate taskquery to shared ParseError

- [x] 3.1 Replace `taskquery.ParseError` with `querybar.ParseError`, update all references
- [x] 3.2 Update taskquery tests to use `querybar.ParseError`

## 4. Centralized app error bar

- [x] 4.1 Add `err error` field to `tui.Model`, intercept `case error:` in `Update`, clear on any keypress (before forwarding key to screen)
- [x] 4.2 In `renderFooter()`, render error in red/bold instead of help bar when `m.err != nil`
- [x] 4.3 Remove `err` field and `case error:` handling from `projectlist.go`, remove error footer rendering
- [x] 4.4 Remove `err` field and `case error:` handling from `projectpicker/model.go`, remove error rendering
- [x] 4.5 Update save-error overlays (taskedit, projectedit, projectstatus, taskstatus): keep internal block state, return error as cmd instead of rendering locally
- [x] 4.6 Add tests for app error bar: error displayed, any key clears, error replacement

## 5. Replace task list inline query bar with shared component

- [x] 5.1 Replace query bar fields in `tasklist/model.go` (`query`, `editing`, `parseErr`, `debounceSeq`, `queryAreaHeight`) with `querybar.Model`
- [x] 5.2 Update `tasklist.Update` to delegate to `querybar.Model`, handle `ApplyMsg`/`CancelMsg`, remove `updateEditing` and `validate` methods
- [x] 5.3 Update `tasklist.View` to use `querybar.View()` (single line), adjust list height calculation (1 line instead of 3)
- [x] 5.4 Update `tasklist/keymap.go`: remove `Apply`/`Cancel` bindings (owned by querybar), keep `FocusQuery`
- [x] 5.5 Update tasklist tests for new query bar integration

## 6. Wire query bar into project list

- [x] 6.1 Add `querybar.Model` to project list `Model`, initialize with `status:open` default query and `projectquery` validate func
- [x] 6.2 Add `/` key binding to `keymap.go` for focusing the query bar
- [x] 6.3 Update `projectlist.Update` to delegate to querybar when focused, handle `ApplyMsg` to parse and reload, handle `CancelMsg`
- [x] 6.4 Update `projectlist.View` to render query bar above the list, adjust list height (1 line for query bar)
- [x] 6.5 Update `projectlist.KeyMap` to show query editing bindings when focused
- [x] 6.6 Add tests for project list query bar: focus, apply, cancel, default filter
