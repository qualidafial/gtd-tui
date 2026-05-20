# Agent Guidelines

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
