# Todos

## Tasks

- Description markdown should be colorized
    - Ditto when you press ctrl+e to open description in an editor
- esc to cancel task editor conflicts with esc to cancel filtering.
- defer until should only display when status is "deferred"
- waiting does not capture who/what is being waited on.

## General

- How do we test UIs? How much is the effort worth?

## Date field

- Date entry: you have to backspace / delete out any current value to enter in something new
- Natural-language parsing is always-on. Add a `.Natural(true)` opt-in toggle so e.g. `"foo"` doesn't silently resolve to today.
- goja JS engine pulled in by naturaltime adds a few MB to the binary — revisit if size matters.

## Vertical slice scaffolding

All currently commented out in `tui/app.go` and elsewhere:

- Projects screen
- Project tasks
- Notes screen
- Timeline screen
