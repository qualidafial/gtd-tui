## 1. form package skeleton

- [x] 1.1 Create `tui/components/form/` exposing `form.Model` (value type), `form.New(fields ...Field) Model` (returns value, panics on duplicate `Key()`), the `form.Field` interface (with `Key() string`, `Visible(Values) bool`, value-returning `Focus`/`Blur`/`Update`), `form.Values` (`Get(key) any`), `form.KeyMap`, and `form.SubmittedMsg`.
- [x] 1.2 Implement `Init`/`Update`/`View`/`Next`/`Prev`/`Submit`/`Focused` with value-returning signatures (no pointer receivers on the form itself). Build a `Values` snapshot at the start of every render/navigation/submit pass and pass it into `field.Visible(values)`. `Submit` is a synchronous for-loop over visible fields' `Validate()`; on first failure it focuses the offending field and returns `ok=false`. On success `Update` emits `SubmittedMsg`.
- [x] 1.3 Honor `Visible(Values)` everywhere: skip hidden fields in `View` rendering, `Next`/`Prev` navigation, `Submit` validation, and `ShortHelp`/`FullHelp`. On each `Update`, if the focused field has flipped to `Visible(values) == false`, move focus to the next visible field.
- [x] 1.4 Implement `ShortHelp()` / `FullHelp()` on `*Model` that compose the form's KeyMap with the focused (visible) field's bindings. Screens delegate help with `return f.ShortHelp()`.
- [x] 1.5 Unit tests covering: initial focus, tab/shift-tab navigation, `Submit` happy path, `Submit` halts on first invalid field, ctrl+s through `Update`, `ShortHelp` updates as focus moves, **hidden fields are skipped in navigation/render/submit**, **focus advances when the focused field becomes hidden**, **`Values.Get` returns the value of a hidden field unchanged**, **`Values.Get` on an unknown key returns nil**, **`form.New` panics on duplicate keys**.

## 2. inputfield subpackage

- [x] 2.1 Create `tui/components/form/inputfield/` with `inputfield.New(key, label string, opts ...Option) Model` (returns value; key is required and non-empty). Backed by `bubbles/v2/textinput`. `Option` variadic supports `WithValidator(func(string) error)`, `WithPlaceholder(string)`, `WithValue(string)`, `WithVisible(func(form.Values) bool)`.
- [x] 2.2 Implement the `form.Field` interface, including `ShortHelp`/`FullHelp` (likely empty or just cursor bindings).
- [x] 2.3 `View()` renders label + textinput view + cached error.
- [x] 2.4 Unit tests: validator pass/fail, value getter/setter, focus toggles cursor.

## 3. textfield subpackage

- [x] 3.1 Create `tui/components/form/textfield/` with `textfield.New(key, label string, opts ...Option) Model`. Backed by `bubbles/v2/textarea`. Same `Option` shape as `inputfield` (key required, value returned by value).
- [x] 3.2 Wire `alt+enter`/`ctrl+j` newline; expose those bindings in `ShortHelp()` so the form's help footer surfaces them when this field is focused.
- [x] 3.3 Plain `Enter` does NOT insert a newline and does NOT advance the form (form-level navigation is `tab`/`shift+tab`/`ctrl+s`); confirm this by test.
- [x] 3.4 Unit tests: alt+enter inserts newline; tab is not consumed by the textarea (it bubbles up to the form).

## 4. selectfield subpackage

- [x] 4.1 Create `tui/components/form/selectfield/` with `selectfield.New[T](key, label string, options []Option[T], opts ...Option) Model[T]`. Backed by `bubbles/v2/list.Model`. `Option[T]` carries display text and a value of type `T`.
- [x] 4.2 `Value()` returns the currently selected `T`; supports a configurable "none" option for pointer-typed `T`.
- [x] 4.3 Filtering (`/` key, etc.) is inherited from `list.Model` and surfaces in `ShortHelp()`.
- [x] 4.4 Unit tests: navigation, filter, value reporting.

## 4a. radiofield subpackage

- [x] 4a.1 Create `tui/components/form/radiofield/` with `radiofield.New[T](key, label string, options []Option[T], opts ...Option) Model[T]`. Inline rendering: `( ) Option1  (•) Option2  ( ) Option3`. No filtering, no scroll — intended for binary or small-N choices.
- [x] 4a.2 `left`/`right` arrow keys move between options; the selected option highlights in an accent color when the field is focused.
- [x] 4a.3 `Value()` returns the currently selected `T`. `Validate()` runs the caller-supplied validator (rare, but supported via `WithValidator`).
- [x] 4a.4 `ShortHelp()` advertises `←/→: choose`.
- [x] 4a.5 Unit tests: arrow navigation, focus highlighting, value reporting, `WithVisible` predicate.

## 5. datefield subpackage (replaces tui/components/date)

- [x] 5.1 Move/rewrite `tui/components/date/date.go` to `tui/components/form/datefield/datefield.go`. Implement `form.Field` instead of `huh.Field`. Drop `huh.Accessor`, `huh.Theme`, `huh.InputKeyMap`, `huh.PrevField`/`huh.NextField` references.
- [x] 5.2 Keep the `naturaltime` parsing and `bubbles/textinput` backing.
- [x] 5.3 On `Blur()`, if the current text parses to a non-nil time, replace the displayed text with the absolute representation (e.g. `2026-06-02 15:00`) so the user sees what's about to be saved. On `Focus()`, leave the absolute text as-is for further editing.
- [x] 5.4 Update or migrate any tests that referenced the huh interface or the old import path. *(No tests existed for the old `date` package; new `datefield` has its own.)*
- [x] 5.5 Update every caller of `tui/components/date` to import `tui/components/form/datefield` instead. *(All four callers — `projectedit`, `taskedit`, `projectstatus`, `taskstatus` — flipped during their respective migrations in sections 8/9/11/12; the old `date` package is gone.)*

## 6. savefield subpackage

- [x] 6.1 Create `tui/components/form/savefield/` with `savefield.New(key string, opts ...Option) Model` (default label "Save"; `WithLabel(string)` to override).
- [x] 6.2 `View()` renders a focusable button (`[ Save ]` styled when focused, dimmed otherwise). `Validate()` always returns nil.
- [x] 6.3 On `Enter` while focused, emit the form's submit signal — same path as `ctrl+s`, routed through `form.Model.Submit`.
- [x] 6.4 `ShortHelp()` advertises `enter: save`.
- [x] 6.5 Unit tests: enter triggers SubmittedMsg via the parent form; tab leaves and re-enters without triggering submit.

## 7. Migrate itemcapture

- [x] 7.1 Rewrite `tui/pages/inbox/itemcapture/model.go` using `form.Model` with `inputfield` (Title), `textfield` (Description), and a trailing `savefield`. Remove `formx.Save` intercept; rely on `SubmittedMsg`.
- [x] 7.2 Remove all `m.form.State == huh.StateNormal` branches and the `(form, ok := form.(*huh.Form))` cast pattern.
- [x] 7.3 Replace `ShortHelp`/`FullHelp` with `return f.ShortHelp()` (and `[][]key.Binding{f.ShortHelp()}` for full).
- [x] 7.4 Re-run `tui/pages/inbox/itemcapture/model_test.go`; adjust assertions only if they referenced `huh` types directly.

## 8. Migrate taskstatus

- [x] 8.1 Rewrite `tui/pages/tasks/taskstatus/model.go` using `form.Model` with a `datefield` (When) and a trailing `savefield`. The previous `huh.Confirm` is replaced by save-button confirmation.
- [x] 8.2 Re-run associated tests.

## 9. Migrate projectstatus

- [x] 9.1 Same shape as taskstatus: `datefield` + `savefield`.
- [x] 9.2 Re-run associated tests. *(No test file in this package.)*

## 10. Migrate projectpicker

- [x] 10.1 Rewrite `tui/pages/projects/projectpicker/model.go` using `form.Model` with a single `selectfield[*int64]`. No trailing savefield needed — `Enter` on a selection commits and propagates through `SubmittedMsg`.
- [x] 10.2 Verify `/`-to-filter behavior matches the current overlay. *(Filtering is inherited from `bubbles/list`; Enter is intercepted by `WithSubmitOnEnter` only when `FilterState() != Filtering`, so `/`-then-Enter still applies the filter rather than submitting.)*
- [x] 10.3 Re-run `tui/pages/projects/projectpicker/model_test.go`.

## 11. Migrate projectedit

- [x] 11.1 Rewrite `tui/pages/projects/projectedit/model.go` (Title, Outcome, Description, Due, `savefield`) on `form.Model`. Remove `formx.Save` intercept and ShortHelp append.
- [x] 11.2 Verify create-on-submit still triggers the existing dismiss-then-push sequence via `viewFactory` (now wired off `SubmittedMsg`).
- [x] 11.3 Re-run `tui/pages/projects/projectedit/model_test.go`.

## 12. Migrate taskedit

- [x] 12.1 Rewrite `tui/pages/tasks/taskedit/model.go` (Title, Description, Assignee, Due, Defer Until, `savefield`) on `form.Model`. Remove `formx.Save` intercept and ShortHelp append.
- [x] 12.2 Verify the read-only header (Task ID/Created/Updated/Status/Project) still renders — it lives outside the form.
- [x] 12.3 Re-run `tui/pages/tasks/taskedit/model_test.go`.
- [x] 12.4 Measure ctrl+s save latency on a 5-field edit; confirm it's imperceptible (sub-frame). *(Deferred to section 14 validation.)*

## 12a. Clarify wizard rewrite

- [x] 12a.1 Audit `tui/pages/inbox/clarify/**` to enumerate every possible field across all branches (task/project, self/delegated, do-it-now, defer, someday, trash) and their visibility predicates.
- [x] 12a.2 Replace the hand-rolled step machine with a single `form.Model` holding every possible field as a member. Each conditional field is constructed with `WithVisible(func(v form.Values) bool { return v.Get("kind") == ... })` referencing upstream field keys. *(Single mainForm for the linear walk + per-task loopForm rebuilt after each project-loop commit, per user direction.)*
- [x] 12a.3 Verify the user experience matches the existing wizard: questions reveal as previous answers commit, ESC backs out, ctrl+s and the trailing `savefield` both commit. *(Reveal-on-answer is implicit in visibility predicates; ESC dismisses; ctrl+s and the branch-specific savefield buttons both commit.)*
- [x] 12a.4 Delete the custom step/branch code that becomes dead. *(Replaced `model.go` whole; deleted `keymap.go`; doitnow side-overlay retained.)*
- [x] 12a.5 Re-run clarify tests; adjust assertions that referenced the step machine. *(Rewritten for form-style UX: tab navigation + arrow keys on radios + ctrl+s submit.)*

## 13. Cleanup

- [x] 13.1 Delete `tui/components/formx/` entirely.
- [x] 13.2 Delete `tui/components/date/` (replaced by `tui/components/form/datefield/`); confirm no stragglers.
- [x] 13.3 Grep for any remaining `huh.` or `charm.land/huh` references across `tui/`; remove or migrate. *(Two `render.go` files used `huh.ThemeFunc` for a red color; replaced with static `lipgloss.Color("9")`.)*
- [x] 13.4 Remove `charm.land/huh/v2` from `go.mod` and run `go mod tidy`. *(`go mod graph | grep huh` returns nothing.)*
- [x] 13.5 Run `go test ./...`; confirm everything passes.
- [x] 13.6 Delete `openspec/changes/submit-form-shortcut/` (superseded; never archived into main specs).

## 14. Validation

- [x] 14.1 Run the app: confirm itemcapture, taskedit, projectedit, taskstatus, projectstatus, and projectpicker all behave correctly (functionality + traversal — rendering may legitimately differ from huh's output).
- [x] 14.2 Confirm `ctrl+s` saves instantly on the 5-field task editor.
- [x] 14.3 Confirm `Enter` on `[ Save ]` saves; `Esc` cancels everywhere.
- [x] 14.4 Confirm date fields visibly snap to their parsed absolute form when blurred.
- [x] 14.5 Confirm help footers in each overlay read sensibly and update as focus moves.
- [x] 14.6 Run `openspec validate remove-huh-forms --strict`. *(Green.)*
