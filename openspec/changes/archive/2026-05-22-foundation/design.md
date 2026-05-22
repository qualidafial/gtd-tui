## Context

The GTD TUI application has existing design documentation in DESIGN.md and ARCHITECTURE.md that captures product vision, domain model, workflows, and technical decisions. This change formalizes that documentation into OpenSpec specs that future changes can reference and extend.

The existing documentation is authoritative — this change extracts and structures it, not redesigns it.

## Goals / Non-Goals

**Goals:**
- Formalize DESIGN.md content into domain-model and gtd-workflows specs
- Formalize ARCHITECTURE.md content into architecture spec
- Create a baseline that future changes reference
- Enable spec-driven development workflow

**Non-Goals:**
- Changing any existing design decisions
- Writing new code
- Resolving open design questions (those stay in DESIGN.md until addressed)

## Decisions

### Extract vs. Reference
**Decision**: Extract requirements into specs rather than referencing DESIGN.md/ARCHITECTURE.md directly.

**Rationale**: Specs need to be in OpenSpec format (requirements with scenarios) to work with the tooling. The source documents use prose and bullet lists. Extraction also allows the specs to evolve independently as changes are made.

### Three capability specs
**Decision**: Split into domain-model, gtd-workflows, and architecture specs.

**Rationale**: Matches the natural divisions in the source documents. Domain model covers entities and relationships. GTD workflows cover the five core workflows plus clarify outcomes. Architecture covers project layout, service design, and SQLite patterns.

### Scenarios as acceptance criteria
**Decision**: Each requirement has scenarios in WHEN/THEN format.

**Rationale**: Scenarios are testable and unambiguous. They serve as acceptance criteria for implementation and as documentation for expected behavior.

## Risks / Trade-offs

**Spec drift** → Keep source documents (DESIGN.md, ARCHITECTURE.md) as the authoritative narrative; specs are the structured extraction. When making changes, update specs first, then sync prose if needed.

**Over-specification** → Some scenarios may be too granular or test implementation details. This is acceptable for foundational specs; future changes can relax or consolidate as patterns emerge.
