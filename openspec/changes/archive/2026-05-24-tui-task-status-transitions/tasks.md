## 1. Generalize the confirmation overlay

- [x] \1 Introduce a `Transition` type (Complete / Drop / Reopen) with a table mapping each to its confirm title, description, affirmative label, and service method
- [x] 1.2 Change the overlay constructor to `New(task, svc, transition)`, building the `huh.NewConfirm` fields from the transition's table entry
- [x] 1.3 Replace the hardcoded `DropTask` call with a dispatch on the transition (CompleteTask / DropTask / ReopenTask); keep the `TasksChanged()` emit on success
- [x] 1.4 Rename the package/files if appropriate (e.g. `taskdelete` → `taskstatus`/`tasktransition`) and update imports; remove the drop-only constructor

## 2. Wire keybindings in the task list

- [x] 2.1 Add a `space` binding to `tasklist/keymap.go` (drop-only `KeyDelete` stays)
- [x] 2.2 In the model keypress switch, handle `space`: resolve transition from selected task status (pending→Complete, done/dropped→Reopen) and open the confirm overlay
- [x] 2.3 Update the `delete` handler to open the confirm overlay with the Drop transition for pending/done, and no-op for dropped
- [x] 2.4 Ensure `space` is handled before the fall-through to `m.list.Update` so the bubbles list does not consume it

## 3. Contextual help label

- [x] 3.1 Thread the selected task's status into `KeyMap()` so the `space` binding's help text is `complete` (pending) or `reopen` (done/dropped)
- [x] 3.2 Handle the empty-selection case (neutral label or omit `space`); omit `delete` from help when the selected task is dropped

## 4. Verify

- [x] 4.1 Run the app: complete a pending task, reopen a done and a dropped task, drop a pending and a done task, confirm delete is inert on dropped
- [x] 4.2 Confirm the help label switches between `complete`/`reopen` as selection moves across statuses
- [x] 4.3 Confirm a transitioned task disappears from a non-matching filter after refresh
- [x] 4.4 Update or add tasklist model tests covering transition resolution and the no-op case; run `go test ./...`
