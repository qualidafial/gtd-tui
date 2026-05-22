## Context

The GTD TUI app needs to transform inbox items into actionable outcomes. The foundation specs (gtd-workflows, domain-model, architecture) define the requirements; this design covers implementation.

Current state: Task and Project entities exist with SQLite stores. Item entity is being added by implement-inbox. This change adds Someday, Reference entities and the InboxService that orchestrates clarify operations.

Constraints:
- Follows Ben Johnson's Go structure (domain in root, SQLite in sqlite/, orchestration in service/)
- Pure Go SQLite driver (modernc.org/sqlite), Squirrel for queries
- Value semantics throughout, int64 IDs, nullable FKs as pointers
- Each clarify operation must be atomic (single transaction)

## Goals / Non-Goals

**Goals:**
- Implement Someday and Reference entities with SQLite stores
- Implement InboxService with 5 transactional clarify operations
- Each operation atomically creates destination entity and stamps Item.ClarifiedInto
- Follow existing patterns from Task/Project implementations

**Non-Goals:**
- TUI views for clarify flow (separate change)
- MeetingLink rewriting on clarification (meeting entity not yet implemented)
- Someday promotion to Project/Task (separate change after base entities exist)
- Periodic review surfacing of stale Someday items (future feature)

## Decisions

### Decision 1: Someday and Reference as separate entities

**Choice**: Create distinct Someday and Reference entities rather than a unified "filed item" with a type discriminator.

**Rationale**: They have different fields (Someday has ReviewedAt for periodic review surfacing; Reference is pure content) and different future behaviors (Someday promotes to Task/Project; Reference stays as-is). Separate entities make the code clearer and avoid nullable fields.

**Alternative considered**: Single "FiledItem" with Type enum. Rejected because it conflates different use cases and requires nullable fields for type-specific data.

### Decision 2: InboxService in service/ package

**Choice**: Place InboxService in service/ package, not sqlite/ package.

**Rationale**: Clarify operations span multiple stores (Item, Task, Project, Someday, Reference). The service/ package exists for cross-service orchestration per ARCHITECTURE.md. InboxService takes store interfaces and coordinates transactions.

**Alternative considered**: Put clarify methods directly on sqlite.ItemStore. Rejected because clarify creates entities in other stores, violating single-responsibility. A store should only manage its own table.

### Decision 3: Store interfaces for Someday and Reference

**Choice**: Define SomedayStore and ReferenceStore interfaces in root package alongside the domain types, following TaskStore/ProjectStore pattern.

**Rationale**: Consistent with existing architecture. Enables testing InboxService with mock stores.

**Interfaces**:
```go
type SomedayStore interface {
    Create(ctx context.Context, s Someday) (Someday, error)
    Get(ctx context.Context, id int64) (Someday, error)
    Update(ctx context.Context, s Someday) (Someday, error)
    Delete(ctx context.Context, id int64) error
    List(ctx context.Context, opts SomedayListOptions) ([]Someday, error)
}

type ReferenceStore interface {
    Create(ctx context.Context, r Reference) (Reference, error)
    Get(ctx context.Context, id int64) (Reference, error)
    Update(ctx context.Context, r Reference) (Reference, error)
    Delete(ctx context.Context, id int64) error
    List(ctx context.Context, opts ReferenceListOptions) ([]Reference, error)
}
```

### Decision 4: Transaction coordination via callback

**Choice**: InboxService takes a transaction-provider function from the SQLite layer for atomic operations.

**Rationale**: Following the existing service-level transaction pattern from ARCHITECTURE.md. The sqlite package provides a `WithTx` function; InboxService calls it to wrap multi-store operations.

**Pattern**:
```go
func (s *InboxService) ClarifyAsTask(ctx context.Context, itemID int64, task Task) (Task, error) {
    var result Task
    err := s.withTx(ctx, func(tx *sql.Tx) error {
        // 1. Get item
        // 2. Create task
        // 3. Update item.ClarifiedIntoTaskID
        return nil
    })
    return result, err
}
```

### Decision 5: Item.ClarifiedInto as nullable FKs with CHECK constraint

**Choice**: Item has four nullable FK fields: ClarifiedIntoTaskID, ClarifiedIntoProjectID, ClarifiedIntoSomedayID, ClarifiedIntoReferenceID. A CHECK constraint ensures at most one is set. A separate Discarded bool field marks discarded items.

**Rationale**: Consistent with existing pattern for dual-FK tables (Comments, MeetingLinks). Allows direct FK relationships with referential integrity. The Discarded field is separate because discarding doesn't create a destination entity.

**Alternative considered**: Single ClarifiedIntoType enum + ClarifiedIntoID. Rejected because it loses FK constraints and type safety.

### Decision 6: Someday.ReviewedAt defaults to CreatedAt

**Choice**: ReviewedAt field defaults to CreatedAt on creation. Updated explicitly when user reviews the item.

**Rationale**: Per DESIGN.md, ReviewedAt is used to surface stalest items in periodic review. Defaulting to creation time means new items appear at the end of the stale list initially.

## Risks / Trade-offs

**[Risk: implement-inbox dependency not complete]** → This change depends on Item entity existing. If implement-inbox is delayed, stub the Item store interface and proceed with Someday/Reference implementation. The InboxService tests can use mocks.

**[Risk: Transaction spanning stores adds complexity]** → Each clarify operation touches multiple stores within one transaction. Mitigation: Keep transaction logic in InboxService only; stores operate on provided tx. Test the atomic behavior explicitly.

**[Trade-off: Someday promotion not included]** → Promoting Someday to Task/Project requires additional service methods. This is deferred to keep scope focused on the 5 clarify outcomes. Future change can add PromoteSomeday methods.

**[Trade-off: No MeetingLink rewriting]** → Meeting entity isn't implemented yet. When it is, MeetingLink.ItemID should be rewritten to point at the clarified entity. Documented as future work.
