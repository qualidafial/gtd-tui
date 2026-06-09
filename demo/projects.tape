# Projects: open a project view with its linked tasks, filter by status.
# Regenerate with: ./demo/generate.sh  (or `vhs demo/projects.tape` with GTD_DB set).
Output demo/projects.gif
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

# Switch to the Projects tab (Inbox -> Tasks -> Projects).
Tab
Sleep 400ms
Tab
Sleep 1.2s

# Open the selected project to see its outcome and linked next actions.
Enter
Sleep 2s

# Back to the project list.
Escape
Sleep 1s

# Show parked (someday) projects via the query bar.
Type "/"
Sleep 500ms
Backspace 12
Sleep 1.2s
Type "status:someday"
Sleep 1.2s
Enter
Sleep 1.5s

# Reset the filter.
Type "\"
Sleep 1.5s

Ctrl+C
