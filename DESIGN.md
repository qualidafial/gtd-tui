# GTD TUI — Design Document

## What This Is

A personal productivity app for terminal users, built around the GTD (Getting Things Done) methodology. Everything lives in one place: tasks, projects, and notes — cross-linked and navigable from a single interface.

## Guiding Principles

**Low ceremony.** The biggest failure mode of GTD systems is the overhead of maintaining them. This app should get out of the way. Capturing a thought, linking it to a project, or finding what to work on next should take seconds, not minutes.

**One pane of glass.** Tasks, projects, and notes are not separate tools. They are different views of the same interconnected data. Navigating between them should feel natural, not like switching apps.

**Easy linking.** Relationships between entities are first-class. When you add action items to a meeting note, they should surface in the relevant project or inbox automatically. Cross-links should be easy to create and easy to follow.

**Timeline as context.** Every entity has a history. You should be able to see what happened on a project, a task, or a note — in chronological order — without hunting through separate logs.

## Core Use Cases

### Capture
- Quickly add items to the inbox without friction
- Inbox is the default entry point when it contains unprocessed items

### Clarify (Process Inbox)
- Review inbox items one at a time with a detail panel
- Promote an item to a task with full metadata (status, due date, defer-until)
- Link tasks to one or more projects
- Create a new project inline while processing a task

### Organize
- View tasks filtered by status, project, or due date
- Navigate to the task list (next actions) when the inbox is empty
- Projects show their linked tasks and notes

### Engage
- See what to work on next from the task list
- Navigate from a task to its linked notes and project context
- View timeline of activity on any task or project

### Capture Context
- Quickly attach a note to a project or task to record what just happened: a hallway conversation, a decision, an observation
- Notes are timestamped and appear in the project/task timeline
- The bar for writing a note should be as low as possible — a single line is enough

### Reflect
- Activity timelines scoped to a project, task, or note
- Global timeline across all entities
- Because capturing is low-friction, timelines become an accurate record of what actually happened and when — useful for retroactive reports, status updates, or just remembering why a decision was made

## Entities

**Task** — a single actionable item. Has a title, description, status, optional due date, and optional defer-until date. May belong to multiple projects and have multiple linked notes.

**Project** — a multi-step outcome requiring more than one action. Has a short title (for lists), an outcome statement (the desired end state), status, optional due date, and description. May have many tasks and notes.

**Note** — a markdown document attached to a project or task. May be a meeting note, a quick observation, a decision record, or a reference document. Notes are timestamped and contribute to the timeline of anything they're linked to. Notes can contain action items that automatically link to the associated project or appear in the inbox.

## What This Is Not

- A team tool. This is personal, single-user software.
- A pomodoro timer or time tracker.
- A replacement for a calendar. Due dates are supported, but calendar views are out of scope for now.
