# Agent Guidelines

## Go Language Features

This project uses Go 1.26. Use `new(value)` to create a pointer to the value of an expression
(e.g. `new("alice")` returns `*string` pointing to `"alice"`). Do not create `ptr()` helper functions.

## Huh Form Value Pointers

Huh forms bind to values via `Value(ptr)`. Because bubbletea uses value-receiver `Update` methods, the
model is copied on every update cycle. Any pointer huh holds must survive those copies — it must point
through a stable indirection, not directly into a model field.

- For fields on a `*struct` stored in the model (e.g. `m.task` is `*gtd.Task`), `Value(&m.task.Field)` is
  safe because the pointer dereference is stable across copies.
- For standalone fields on the model struct itself, use a double-pointer (`**T`): store `**T` in the model,
  initialize with `new(initialValue)`, and pass the `*T` (i.e. `m.field`, not `&m.field`) to `Value()`.
  This way huh writes through the stable `*T`, not into a field that moves with each model copy.
- For `huh.Select[*T]` where options carry pointer values: huh matches the current value against options
  using pointer equality (`==`), not deep equality. The initial `*selected` must point to the exact same
  allocation as the matching option's value — not a separate copy with the same contents.

## Tests

All code modifications must be accompanied by appropriate test changes. When adding behavior, add tests covering it.
When changing behavior, update or extend the affected tests. When removing behavior, remove the now-obsolete tests. Run
the relevant tests before marking the task complete.

## TODO.md Hygiene

`TODO.md` tracks the running backlog. When a task in the file is completed by the current change, remove its entry (or
check it off if the surrounding section uses checkboxes) as part of the same commit. Don't leave finished work listed —
it makes the file lie about what's still pending.

## Avoid Refuctoring

"Refuctoring" is letting several logically independent changes — a refactor, a feature, a rename, a scope change — pile
up in the working tree until they can no longer be untangled into clean commits. Once a single file mixes two unrelated
concerns, history-splitting requires authoring intermediate states by hand, and reviewers lose the ability to bisect or
revert one concern without the others.

While working, watch for these signals and warn me before continuing:

- A pending change starts touching files that are already dirty for an unrelated reason (e.g. a feature edit lands in a
  file mid-refactor).
- The working tree spans more than one coherent commit subject — if the commit message would need "and" or a bulleted
  list of unrelated topics, it's already too entangled.
- A new concern modifies the same hunks as an in-progress concern, so the diffs can no longer be staged independently
  with `git add -p`.
- Cross-cutting renames, signature changes, or scope pivots are being layered on top of unfinished feature work.

When this happens, stop and surface the entanglement. Offer to commit (or stash) the in-flight work first so the next
change starts from a clean baseline. Don't silently absorb the new concern into the existing diff.

## Git Commit Messages

Follow standard Git commit message conventions:

- Subject line in the imperative mood, capitalized, no trailing period, 50 characters or less.
- Blank line between the subject and the body.
- Wrap body lines at 72 characters.

Keep messages feature-oriented: describe what the change accomplishes for the user and why, not the technical mechanics
of how it was implemented. Save implementation details for the code, comments, or PR discussion.

Do not add `Co-Authored-By: Claude ...` trailers or any other AI-authorship attribution to commit messages.
