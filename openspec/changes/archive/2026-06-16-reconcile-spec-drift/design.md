## Context

Commits `9d6ba0d`..`fc43e01` shipped user-visible behavior without amending the
living specs in `openspec/specs/`. A drift scan found three affected specs:
`task-list-presentation` (project moved from a trailing chip to a leading indigo
label; selection now brightens it), `project-view-screen` (`enter` drills into
the task view; in-project go-to-project disabled), and `form-field-toolkit`
(post-construction select population). The specs are the documented source of
truth, so they must catch up to the code. No production code changes are in
scope — this change only edits spec files.

## Goals / Non-Goals

**Goals:**
- Bring the three specs in line with shipped behavior so they remain
  authoritative.
- Keep the spec deltas archivable cleanly (correct ADDED/MODIFIED/REMOVED
  operations).

**Non-Goals:**
- Any change to application code, tests, dependencies, or storage.
- Re-litigating the shipped design decisions (chip→label, color, enter-opens-view).
- Reconciling drift outside the last six commits (e.g. the README's `Task.Kind`
  description, which predates this window).

## Decisions

- **Project chip → label uses REMOVED + ADDED, not MODIFIED.** The requirement
  is both renamed ("Project chip" → "Project label") and rewritten (position,
  prefix, color). MODIFIED matches by header text and a rename mid-rewrite is
  ambiguous at archive time, so the old requirement is REMOVED (with Reason +
  Migration pointing at the new one) and a fresh "Project label" requirement is
  ADDED. This mirrors the REMOVED convention already used in the archived
  `rename-pending-to-open` change.
- **Project color and status-suppression move into the new "Project label"
  requirement.** The old `+project` bullet in "Urgency colors" and the project
  sentence in "Chip suppression by status" are removed via MODIFIED, since the
  project is no longer a chip; its color and done/dropped rules live in the
  ADDED requirement instead.
- **`enter`-opens-view and dynamic select options use ADDED, not MODIFIED.**
  Neither concern was described in the existing specs, so they are additive
  rather than modifications of existing requirements.

## Risks / Trade-offs

- [REMOVED Project chip could orphan references in other specs] → grep confirmed
  the `+project`/"Project chip" wording is local to `task-list-presentation`;
  no cross-spec references break.
- [Spec says one thing, code another if a scenario is mis-stated] → each
  scenario was written against the shipped code (`render.go`, `app.go`,
  `selectfield.go`, `form.go`); `openspec validate --strict` passes.
