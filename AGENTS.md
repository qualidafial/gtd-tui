# Agent Guidelines

## File Edits

Always use the Edit or Write tools to modify files. Do not use `sed`, `awk`, or other shell commands that rewrite files in place — they replace the file inode, which causes VS Code to miss the change and show stale content.
