## 1. Query bar component

- [x] 1.1 Update `debounceMsg` handler in `querybar.Update` to validate the current value and, on success, return `ApplyMsg{Query: <trimmed value>}` — without blurring or updating `appliedQuery`
- [x] 1.2 Update esc branch in `querybar.Update` to emit `ApplyMsg{Query: m.appliedQuery}` after reverting the input, clearing the parse error, and blurring
- [x] 1.3 Remove the `CancelMsg` type and its package-level comment

## 2. Tests

- [x] 2.1 Replace `TestCancel_RevertsAndBlurs` with a test that verifies esc reverts the value, blurs, and emits `ApplyMsg` carrying the previously-applied query
- [x] 2.2 Replace `TestDebounce_ValidatesOnCurrentSeq` (or add a peer test) covering the success path: a valid query at the current debounce seq emits `ApplyMsg` with the trimmed value, the bar remains focused, and `appliedQuery` is unchanged
- [x] 2.3 Keep the existing failure-path debounce test, adjusted if needed so the validator returns `*ParseError` and the cmd yields the error (not an `ApplyMsg`)

## 3. Parent screens

- [x] 3.1 Remove the `case querybar.CancelMsg:` branch from `tui/pages/tasks/tasklist/model.go`
- [x] 3.2 Remove the `case querybar.CancelMsg:` branch from `tui/pages/projects/projectlist.go`
- [x] 3.3 Run `go test ./...` and verify no remaining references to `querybar.CancelMsg`
