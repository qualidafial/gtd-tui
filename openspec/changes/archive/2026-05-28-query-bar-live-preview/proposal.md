## Why

The query bar's current commit-on-enter model forces users to guess at results: type a query, press enter, see the filtered list, press `/` again to refine, repeat. There is no way to see how the list narrows as a query is typed without committing first.

A debounced live preview removes that round trip. As the user types, the list refreshes whenever the query parses cleanly, but the bar stays focused so further edits keep refining the preview. Pressing enter still commits (locking the query as the new "applied" state and blurring the bar), and pressing esc reverts to the previously-applied query — visually undoing every previewed refinement in one keystroke.

This also collapses the role of the debounce tick. Today it only updates the inline error highlight; under the new model it does that *and* drives the live reload, so a single timer covers validation feedback and preview.

## What Changes

- **Debounce triggers apply on success.** When the debounce tick fires and the current value parses cleanly, the query bar emits `ApplyMsg{Query: <current value>}` — same message parents already handle for commit — but does NOT blur and does NOT update the stored "applied query". Validation failures still surface a parse error and inline highlight.
- **Esc reverts via re-apply.** Pressing esc restores the bar's value to the last applied query, clears any parse error, blurs, and emits `ApplyMsg{Query: <applied query>}` so the parent reloads with the original filter. The list visibly snaps back from any live-previewed state.
- **Remove `CancelMsg`.** With esc emitting `ApplyMsg` instead, there is no distinct cancel signal for parents to handle. Drop the type and the parent `case querybar.CancelMsg:` branches.
- **Parent screens unchanged structurally.** Both `tasklist` and `projectlist` already handle `ApplyMsg` by parsing and reloading; that handler now runs on every debounce hit as well as on enter and esc. No new state machine.

## Capabilities

### Modified Capabilities
- `query-bar`: Debounce now applies the query on success (stays focused); esc emits `ApplyMsg` with the previously-applied query instead of `CancelMsg`; `CancelMsg` removed.
- `task-list-query-ui`: Live debounced parse now reloads the task list in addition to updating the error display.
- `project-list-ui`: Esc reverts by re-applying the previously-applied query (parent receives `ApplyMsg`, not `CancelMsg`); the visible list snaps back to the last committed filter rather than being left as-is.

## Impact

- Modified: `tui/components/querybar/querybar.go` (debounce branch returns `ApplyMsg` on success; esc branch returns `ApplyMsg{appliedQuery}`; `CancelMsg` type removed)
- Modified: `tui/components/querybar/querybar_test.go` (drop `CancelMsg` assertions; new tests for debounce-apply and esc-revert-via-apply)
- Modified: `tui/pages/tasks/tasklist/model.go` (remove `case querybar.CancelMsg:` branch)
- Modified: `tui/pages/projects/projectlist.go` (remove `case querybar.CancelMsg:` branch)
