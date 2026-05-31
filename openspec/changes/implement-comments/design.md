## Context

The GTD TUI app currently supports Task entities with CRUD operations via TaskService. The existing patterns include:
- Domain types in root package with value semantics
- SQLite implementation using squirrel for query building
- Embedded SQL migrations with CHECK constraints for validation
- Transaction support via `DB.RunTx()` for atomic operations
- Foreign keys enabled via PRAGMA

Tasks do not currently have a comment mechanism. The `UpdateTask` method takes only a `Task` value with no provision for recording why changes were made. Timeline/audit features planned for the Reflect workflow will require Comment data to display meaningful activity history.

## Goals / Non-Goals

**Goals:**
- Add Comment entity following existing architectural patterns
- Enable edit-with-comment atomicity using existing transaction infrastructure
- Support both task and project comments with single-FK constraint
- Provide standalone comment API for context notes unrelated to metadata changes
- Index comments for efficient timeline queries

**Non-Goals:**
- UI implementation (separate change)
- Timeline/activity feed queries (separate change, will consume Comment data)
- Rich text or attachments in comments
- Comment threading or replies
- Project entity implementation (must exist first; this change depends on it or is sequenced after)

## Decisions

### Decision: Single comments table with dual nullable FKs

**Choice:** One `comments` table with nullable `task_id` and `project_id` columns, CHECK constraint enforcing exactly one is set.

**Alternatives considered:**
1. Separate `task_comments` and `project_comments` tables - simpler constraints but duplicates schema and code
2. Polymorphic association with `parent_type`/`parent_id` - loses FK integrity, harder to query

**Rationale:** Dual-FK with CHECK is the established pattern in this codebase (see the MeetingLink requirement in `openspec/specs/domain-model/spec.md` and the "Schema constraints via CHECK" requirement in `openspec/specs/architecture/spec.md`). SQLite enforces the invariant, the schema is self-documenting, and FK integrity/cascades work naturally.

### Decision: Extend service method signatures for comment parameter

**Choice:** Modify `UpdateTask(ctx, Task, comment string)` signature to accept an optional comment string. Empty string means no comment.

**Alternatives considered:**
1. Separate `UpdateTaskWithComment` method - clutters API, every caller must choose
2. Options struct pattern - overengineered for a single optional parameter
3. Context value - inappropriate for business data

**Rationale:** Extending the signature is minimally invasive. The comment parameter is optional (empty = no comment), so existing call sites just add an empty string. Go's explicit parameters make the contract clear.

### Decision: Comment creation inside existing transaction scope

**Choice:** When `UpdateTask` receives a non-empty comment, the sqlite implementation creates the Comment row within the same `RunTx` transaction that updates the task.

**Alternatives considered:**
1. Service layer wrapping sqlite calls - adds indirection without benefit since sqlite already has transaction support
2. Event sourcing / outbox pattern - overkill for a single-user app

**Rationale:** The existing `DB.RunTx` pattern handles nested operations cleanly. The sqlite Task implementation already uses `RunTx`; extending it to include a Comment insert is straightforward.

### Decision: CommentService for standalone operations

**Choice:** Add `CommentService` interface with CRUD methods mirroring other services. This handles standalone comments (not tied to an update) and comment editing.

**Alternatives considered:**
1. Expose comment operations only through TaskService/ProjectService - awkward API, doesn't match domain separation
2. No standalone comment support - limits usefulness; users want to add context without forcing a metadata change

**Rationale:** Follows existing service-per-entity pattern. Standalone comments are a real use case ("blocked on X" note without changing status).

### Decision: Immutable parent reference on comments

**Choice:** Once created, a Comment's TaskID/ProjectID cannot be changed via UpdateComment. The service ignores attempts to change the parent.

**Alternatives considered:**
1. Allow reparenting - no use case, adds complexity, risk of orphaning audit trail
2. Return error on reparent attempt - too strict; silently preserving is safer

**Rationale:** Comments are event-shaped records attached to an entity's timeline. Moving them between entities would break audit history. Ignoring the field change is defensive without being disruptive.

## Risks / Trade-offs

**[Resolved] Sequenced after implement-projects**
This change is sequenced to land after `implement-projects`, which creates the projects table and ProjectService. The comments migration therefore declares its `project_id` FK against an existing table, and this change re-breaks the project-service signatures (UpdateProject and transitions) to add the optional comment parameter that implement-projects deliberately omitted.

**[Risk] Migration ordering with future project migration**
If projects migration runs after comments migration, the FK will fail.
-> Mitigation: Name comments migration with a number higher than the projects migration (e.g., if projects is 0003, comments is 0004+).

**[Trade-off] Signature change to UpdateTask is breaking**
Existing callers must be updated to pass the comment string.
-> Acceptable: This is an early-stage codebase with few callers. A compiler error is better than silent behavior change.

**[Trade-off] No comment history/versioning**
UpdateComment overwrites the body; previous versions are lost.
-> Acceptable: Comments are short notes, not documents. If version history is needed later, it's an additive change (comment_history table).

## Open Questions

1. **Maximum comment length**: Should we enforce a CHECK constraint on body length?
   - Recommendation: No constraint initially. Let usage patterns inform limits. SQLite handles TEXT efficiently.
