# Tasks: browse the list, complete a task, filter with the query bar.
# Regenerate with: ./demo/generate.sh  (or `vhs demo/tasks.tape` with GTD_DB set).
Output demo/tasks.gif
Require gtd

Set Shell bash
Set FontSize 18
Set Width 1200
Set Height 720
Set Padding 24
Set Theme "Catppuccin Mocha"

Hide
Type "gtd" Enter
Sleep 1.5s
Show

Sleep 800ms

# Switch to the Tasks tab.
Tab
Sleep 1.2s

# Move down the list.
Down 2
Sleep 500ms

# Complete the selected task.
Space
Sleep 1.2s
Enter 2

# Filter to completed tasks via the query bar, then apply with enter.
Type "/"
Sleep 500ms
Backspace 22
Type "status:done"
Sleep 1.2s
Enter
Sleep 1.5s

# Reset the filter back to its default.
Type "\"
Sleep 1.5s

Ctrl+C
