# gtd-tui

A personal, single-user productivity app for terminal users, built around the
[Getting Things Done](https://gettingthingsdone.com/) methodology. Tasks,
projects, and notes live in one cross-linked, navigable interface — no app
switching, no maintenance overhead, no friction between thinking and capturing.

> Status: **early development.** The Tasks and Projects vertical slice is
> usable; Inbox, References, Meetings, Comments, and Timelines are specified
> but not yet implemented. See [Roadmap](#roadmap).

## Why

Most GTD tools fail the same way: the system needs more upkeep than the work it's
tracking. `gtd-tui` is built around the opposite premise — capture should take
seconds, navigation should be one keystroke, and every entity should carry its
own history so you can answer "what happened here, and why?" without hunting
through separate logs.

The product principles live in
[`openspec/specs/product-vision/spec.md`](openspec/specs/product-vision/spec.md):

- **Low ceremony** — capture is a single low-friction interaction.
- **One pane of glass** — tasks, projects, and notes are views of the same data.
- **Easy linking** — relationships are first-class; cross-links are easy to
  create and follow.
- **Timeline as context** — every entity has a chronological history.

Out of scope: team collaboration, time tracking, calendar replacement.

## Install

Requires Go 1.26 or newer. No CGO, no native dependencies.

```sh
go install github.com/qualidafial/gtd-tui/cmd/gtd@latest
```

Or build from source:

```sh
git clone https://github.com/qualidafial/gtd-tui
cd gtd-tui
go build -o gtd ./cmd/gtd
```

## Run

```sh
gtd
```

The database is created on first run at `$XDG_CONFIG_HOME/gtd/gtd.db` (typically
`~/.config/gtd/gtd.db`). Migrations are applied automatically.

## Keybindings

Global:

| Key       | Action          |
| --------- | --------------- |
| `?`       | Toggle help     |
| `tab`     | Next tab        |
| `shift+tab` | Previous tab  |
| `esc`     | Back / cancel   |
| `ctrl+c`  | Quit            |

Task list:

| Key        | Action            |
| ---------- | ----------------- |
| `+` / `insert` | New task      |
| `enter`    | Edit task         |
| `space`    | Toggle complete   |
| `delete`   | Drop task         |
| `p`        | Jump to project   |
| `shift+↑/↓`| Reorder (within current filter) |
| `/`        | Filter            |
| `\`        | Reset filter      |

Project list:

| Key        | Action            |
| ---------- | ----------------- |
| `+` / `insert` | New project   |
| `e`        | Edit project      |
| `enter`    | Open project view |
| `space`    | Toggle complete   |
| `delete`   | Drop project      |
| `s`        | Park (set someday)|
| `shift+↑/↓`| Reorder (within current filter) |
| `/`        | Filter            |
| `\`        | Reset filter      |

## Filter syntax

The `/` query bar accepts a small DSL for narrowing the visible list. Multiple
clauses combine with AND; the last value for a repeated key wins.

Examples:

- `status:open` — only open items (default)
- `status:done` — completed items
- `status:dropped` — dropped items
- `status:someday` — parked projects
- `due:2026-06-01` — items due on a specific date
- `tomato` — free-text title match

Invalid clauses are highlighted in the bar and don't apply until corrected.

## Concepts

The domain model is captured in
[`openspec/specs/domain-model/spec.md`](openspec/specs/domain-model/spec.md).
Briefly:

- **Item** — an unprocessed inbox capture. Clarified into a Task, Project, or
  Reference (lineage preserved via soft-delete).
- **Task** — a single actionable item. `Kind` is `next_action` or `delegated`;
  `Status` is `open`, `done`, or `dropped`. Belongs to zero or one Project.
- **Project** — a multi-step outcome. `Status` is `open`, `someday` (parked),
  `done`, or `dropped`. Terminal transitions cascade or detach open tasks; the
  invariant "no open tasks under a closed project" is enforced.
- **Reference** — standalone markdown content kept for retrieval.
- **Meeting** — title, time slot, attendees, markdown body. Action items
  captured during a meeting flow to the inbox with a link back.
- **Comment** — short, event-shaped text attached to a Task or Project; recorded
  implicitly on edits and explicitly via the comment API.

The clarify workflow (Capture → Clarify → Organize → Engage → Reflect) and the
five clarify outcomes (Discard, Incubate, FileAsReference, ClarifyAsTask,
ClarifyAsProject) are in
[`openspec/specs/gtd-workflows/spec.md`](openspec/specs/gtd-workflows/spec.md),
which also includes a Mermaid diagram of the decision flow.

## Project layout

Follows [Ben Johnson's Go application
structure](https://medium.com/@benbjohnson/structuring-applications-in-go-3b04be4ff091):

```
.                       Root package — domain types and service interfaces only
├── cmd/gtd/            CLI entry point
├── service/            Cross-store orchestration (transactional)
├── sqlite/             SQLite implementation
│   └── migrations/     Embedded SQL migrations, applied in order
├── tui/                Bubbletea v2 UI
│   └── pages/          One package per top-level page (tasks, projects, …)
├── internal/set/       Internal generic Set type
└── openspec/           Specifications and proposed changes
    ├── specs/          Authoritative current behavior
    └── changes/        Proposed and archived changes
```

Architectural rules — value semantics, modernc.org/sqlite driver, squirrel for
queries, CHECK constraints, WAL + foreign keys, service-level transactions —
are in [`openspec/specs/architecture/spec.md`](openspec/specs/architecture/spec.md).

## Development

```sh
go test ./...        # full suite, uses in-memory SQLite
go build ./...
openspec validate --specs  # validate all specs
```

Specs are the source of truth. Behavioral changes start with a proposal under
`openspec/changes/`; see existing changes for the format. The `opsx:propose` /
`opsx:apply` / `opsx:archive` slash commands automate the workflow.

## Roadmap

Implemented:

- Tasks (CRUD, status transitions, ordering, filtering)
- Projects (CRUD, status transitions including someday/park, ordering, filtering,
  project view with linked tasks, project picker overlay)
- Shared query bar with live-preview validation
- TUI view stack with overlay support

Specified, not yet implemented (see [`openspec/changes/`](openspec/changes/)):

- **Inbox** — Item entity + four clarify operations
- **References** — Reference entity + FileAsReference
- **Meetings** — Meeting + MeetingLink + AddActionItem
- **Comments** — edit-with-comment + standalone comments
- **Timelines** — activity history per entity + global Reflect view

## License

[MIT](LICENSE) © 2026 Matthew Hall
