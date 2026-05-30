## Context

`implement-inbox` shipped the Item entity and four clarify operations (Discard, Incubate, ClarifyAsTask, ClarifyAsProject). It deliberately left out the fifth operation — `FileAsReference` — and the Reference entity it depends on, so the inbox loop could go in without dragging an extra entity along. This change closes that gap.

The Item entity already has a CHECK constraint enforcing mutual exclusion of `clarified_into_task_id` / `clarified_into_project_id` / `discarded`. Adding the references column means rebuilding the items table via the SQLite table-rebuild pattern (drop CHECK, recreate with the extended constraint), since SQLite does not support ALTER TABLE on CHECK constraints in place.

## Goals / Non-Goals

**Goals:**
- Add Reference entity, ReferenceStore, and SQLite implementation
- Add `ClarifiedIntoReferenceID` to the Item entity and the items table
- Add `FileAsReference` clarify operation following the existing transactional pattern
- Surface References in the TUI as a new tab with querybar-driven title filtering

**Non-Goals:**
- Linking References to Projects or Tasks — References stay standalone
- Markdown rendering on the references page (display raw body for v1)
- Search beyond title prefix/substring filter

## Decisions

### Decision: Reference as a standalone entity with no Task/Project links

**Choice**: Reference has only `ID / Title / Body / CreatedAt / UpdatedAt`. No `ProjectID` or `TaskID`.

**Rationale**: References are retrieval content — recipes, snippets, links — not work to be done. Linking them to projects would conflate the GTD reference material role with project resources, which is a separate concern.

**Alternatives considered**:
- Reference with optional `ProjectID`: blurs the reference-vs-project-attachment distinction; can be added later if a real use case emerges

### Decision: Extend the items CHECK constraint via table rebuild

**Choice**: The migration follows the SQLite-recommended pattern: create a `_new` items table with the extended CHECK constraint and the new column, copy rows, drop old, rename. The same pattern is already used elsewhere in the codebase (per past memory notes on SQLite rebuilds).

**Rationale**: SQLite cannot ALTER an existing CHECK constraint in place. Table rebuild is the supported workaround and preserves data integrity.

**Alternatives considered**:
- Skip CHECK enforcement at DB level and rely on service-layer validation: weakens an invariant that's currently enforced by the database
- Add a separate trigger to enforce the new arm: more moving parts than a clean rebuild

### Decision: Extend the InboxService constructor signature

**Choice**: The service-layer `InboxService` constructor adds a `ReferenceStore` parameter. Call sites in app initialization are updated.

**Rationale**: There is no production code path that constructs `InboxService` without intending FileAsReference once this change ships. A wrapper or optional store would be dead complexity.

## Risks / Trade-offs

**Risk**: The items table rebuild is sensitive — wrong column order or missing index would silently break behavior.
**Mitigation**: Mirror the existing rebuild pattern; cover with migration round-trip tests that insert rows under the old shape, run the migration, and assert the new CHECK and column are in place.

**Risk**: A REFERENCES clause on `clarified_into_reference_id` plus existing FK columns may complicate the rebuild ordering.
**Mitigation**: Disable foreign keys for the duration of the rebuild (standard SQLite pattern), re-enable after rename.
