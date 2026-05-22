## Why

Bootstrap the OpenSpec system with foundational specifications that define the GTD TUI application. The product vision, domain model, and architecture are documented in DESIGN.md and ARCHITECTURE.md but not yet formalized as specs that future changes can reference and extend.

## What Changes

- Establish the core domain model: Item, Task, Project, Someday, Reference, Meeting, Comment, MeetingLink entities and their relationships
- Define the GTD workflows: Capture, Clarify (with five outcomes), Organize, Engage, Capture Context, Reflect
- Document architectural decisions: Go project layout, SQLite layer, service interfaces, conventions
- Create reference specifications that all future changes build upon

## Capabilities

### New Capabilities

- `domain-model`: Core entities (Item, Task, Project, Someday, Reference, Meeting, Comment, MeetingLink), their attributes, relationships, status enums, and validation rules
- `gtd-workflows`: The GTD workflows — Capture (inbox), Clarify (Discard/Incubate/FileAsReference/ClarifyAsTask/ClarifyAsProject), Organize, Engage, Capture Context (comments + meetings), Reflect (timelines)
- `architecture`: Technical architecture including Go project layout (root/sqlite/service/tui/cmd), service interfaces, SQLite implementation patterns, and conventions

### Modified Capabilities

(none - this is the initial spec baseline)

## Impact

- Creates the foundational specs in `openspec/specs/`
- All future feature changes will reference these specs
- No code changes - this formalizes existing documentation