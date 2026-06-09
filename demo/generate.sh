#!/usr/bin/env bash
# Regenerate the README demo GIFs from the .tape scripts in this directory.
#
# Each tape runs against a freshly seeded, throwaway database (GTD_DB) so the
# recordings are deterministic and never touch your real ~/.config/gtd/gtd.db.
#
# Requirements: Go, plus vhs and its runtime deps (ttyd, ffmpeg) on PATH.
#   go install github.com/charmbracelet/vhs@latest
#   # Debian/Ubuntu: sudo apt-get install ttyd ffmpeg
#   # macOS:         brew install vhs ttyd ffmpeg
set -euo pipefail

cd "$(dirname "$0")/.."

for bin in vhs ttyd ffmpeg; do
	if ! command -v "$bin" >/dev/null 2>&1; then
		echo "error: '$bin' not found on PATH; see demo/generate.sh header for install steps" >&2
		exit 1
	fi
done

workdir="$(mktemp -d)"
trap 'rm -rf "$workdir"' EXIT

# Build gtd into a temp dir and put it first on PATH so the tapes' `gtd`
# (and `Require gtd`) resolve to this build, not an installed one.
echo "building gtd..."
go build -o "$workdir/gtd" ./cmd/gtd
export PATH="$workdir:$PATH"
export GTD_DB="$workdir/demo.db"

for tape in demo/inbox.tape demo/tasks.tape demo/projects.tape; do
	echo "seeding + recording $tape..."
	go run ./cmd/gtd-seed "$GTD_DB" >/dev/null
	vhs "$tape"
done

echo "done. GIFs written to demo/*.gif"
