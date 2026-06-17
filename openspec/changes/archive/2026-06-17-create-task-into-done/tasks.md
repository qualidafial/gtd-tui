## 1. Form: Enter submits on the last visible field

- [x] 1.1 In `form.Model.Update`, when the keypress is Enter, the focused field does not claim Enter (the existing `keymap.Handles` guard already lets claiming fields fall through to `field.Update`), and no later visible field exists, call `handleSubmit` instead of the no-op `Next()`; keep tab/down on the last field as no-ops
- [x] 1.2 Add form-level tests: Enter on a non-claiming last field submits; Enter on a non-last field still advances (validator-gated); Enter on a claiming last field routes to the field (no submit)

## 2. savefield becomes a valueless focus placeholder

- [x] 2.1 Remove `savefield`'s Enter claim and `form.SubmitRequestMsg` emission so submission while it holds focus comes from the form-level last-field rule; keep `Value()` nil and `Validate()` returning nil
- [x] 2.2 Update/replace `savefield` tests: drop the "Enter on savefield emits SubmitRequestMsg" expectation; assert its `Keys()` no longer claim Enter and that Enter on a `savefield`-terminated form still submits

## 3. taskedit: status radio on create

- [x] 3.1 In `taskedit.New`, for a new task (`ID == 0`) build an inline `radiofield[gtd.TaskStatus]` keyed `status` with options `[Open, Done]` defaulting to `Open` (`radiofield.WithInitialValue(Open)`) and use it as the terminal field in place of the savefield; stop force-setting `task.Status` on create
- [x] 3.2 For an existing task (`ID != 0`), keep the `savefield` "Save" terminal field and the current behavior unchanged (no status field)
- [x] 3.3 In `saveCmd`, on the create branch read `values["status"].(gtd.TaskStatus)` into `task.Status` before `CreateTask`; leave the update branch untouched

## 4. Tests for taskedit create-into-status

- [x] 4.1 New-task editor: terminal field offers exactly `Open` and `Done`, defaults to `Open`, and `Dropped` is absent
- [x] 4.2 Submitting a new task with `Open` selected creates an open task (default path, no extra keypresses)
- [x] 4.3 Submitting a new task with `Done` selected creates a task in done status with no order key and a `StatusChangedAt` set at creation
- [x] 4.4 Existing-task editor presents no status field and update behavior is unchanged

## 5. selectfield: claim Enter only while filtering

- [x] 5.1 Remove `selectfield.WithSubmitOnEnter` and the `submitOnEnter` field, and delete its `form.SubmitRequestMsg` emission in `Update`; in `Keys()` claim the Enter binding only when `list.FilterState() == list.Filtering`
- [x] 5.2 Drop `selectfield.WithSubmitOnEnter[int64]()` from `projectpicker` and `taskpicker` (they submit via the form's last-field rule and already handle `form.SubmittedMsg`)
- [x] 5.3 Add a selectfield test: `Keys()` does not claim Enter when not filtering, and claims it after entering the filtering state; confirm picker tests still pass

## 6. Verification

- [x] 6.1 Run `go test ./...` and confirm the suite passes
- [x] 6.2 Run the TUI, create a task into `Done`, and confirm it lands in the done view (not the active list) with the correct status timestamp
