## Why

Six commits shipped behavior that the living specs in `openspec/specs/` no
longer describe. The specs are advertised (in the README and the development
workflow) as the source of truth, so the drift undermines that contract.
This change reconciles three specs with the code as it already ships — no code
changes, specs only.

## What Changes

- **task-list-presentation**: replace the trailing `+<title>` project **chip**
  with a **leading project label** rendered ahead of the task title, with no
  `+` prefix; change the project color from the green/cyan family to indigo
  (huh's logo color) so it reads distinctly from the green `ready:` chip; and
  allow the selection highlight to brighten the project label alongside the
  title (it is no longer "chip-only" data that keeps a fixed color).
- **project-view-screen**: document that `enter` drills into the selected
  task's view screen (with `e` still opening the editor), and that go-to-project
  is disabled on the in-project task view since the project is already the
  parent screen.
- **form-field-toolkit**: document the dynamic-options capability that lets a
  select be populated after construction — `selectfield.SetOptions`, the
  `WithHideWhenEmpty` option, and `form.UpdateField` for editing a field in
  place by key.
- **domain-model**: drop the obsolete Task `Kind` field from the Task entity
  description; delegation is inferred from a non-nil `Assignee` (the `Kind`
  field and `TaskKind` type were already removed from the code).

No requirement listed here is a behavior change to the app; each amends a spec
to match shipped code.

## Capabilities

### New Capabilities
<!-- None. This change only reconciles existing specs with shipped code. -->

### Modified Capabilities
- `task-list-presentation`: project association moves from a trailing chip to a
  leading label (no `+`), gains an indigo color distinct from `ready:`, and is
  included in the selection-highlight scope.
- `project-view-screen`: adds an `enter`-opens-task-view requirement and the
  in-project go-to-project suppression.
- `form-field-toolkit`: adds requirements for post-construction select option
  population (`SetOptions`, `WithHideWhenEmpty`, `UpdateField`).
- `domain-model`: Task entity loses its `Kind` field; delegation is expressed
  via a non-nil `Assignee`.

## Impact

- Affected specs: `openspec/specs/task-list-presentation/spec.md`,
  `openspec/specs/project-view-screen/spec.md`,
  `openspec/specs/form-field-toolkit/spec.md`,
  `openspec/specs/domain-model/spec.md`.
- No production code changes — code already implements the reconciled behavior
  (commits `9d6ba0d`..`fc43e01`; the Task `Kind` removal predates that window).
- No dependency or storage impact.
