# Agent Guidelines

## Tests

All code modifications must be accompanied by appropriate test changes. When adding behavior, add tests covering it.
When changing behavior, update or extend the affected tests. When removing behavior, remove the now-obsolete tests. Run
the relevant tests before marking the task complete.

## Git Commit Messages

Follow standard Git commit message conventions:

- Subject line in the imperative mood, capitalized, no trailing period, 50 characters or less.
- Blank line between the subject and the body.
- Wrap body lines at 72 characters.

Keep messages feature-oriented: describe what the change accomplishes for the user and why, not the technical mechanics
of how it was implemented. Save implementation details for the code, comments, or PR discussion.

Do not add `Co-Authored-By: Claude ...` trailers or any other AI-authorship attribution to commit messages.
