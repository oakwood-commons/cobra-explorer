#!/usr/bin/env bash
# Re-render the theme screenshots in doc/themes/ from the per-theme VHS tapes
# in doc/themes/tapes/.
#
# Each theme is rendered in its own isolated VHS process because VHS's
# alt-screen renderer occasionally captures a blank frame. The script verifies
# each screenshot is a non-trivial size and retries a few times if a capture
# comes out blank.
#
# Usage (run from anywhere in the repo):
#   doc/themes/render-themes.sh              # render every theme
#   doc/themes/render-themes.sh dark nord    # render only the named themes
#
# Requires the `vhs` CLI (brew install vhs) and the Go toolchain on PATH.
set -euo pipefail

cd "$(git rev-parse --show-toplevel)"

tapes_dir="doc/themes/tapes"
min_size=50000   # bytes; smaller PNGs indicate a blank/failed capture
max_tries=5

if ! command -v vhs >/dev/null 2>&1; then
	echo "error: vhs is not installed (brew install vhs)" >&2
	exit 1
fi

# Determine which themes to render: either the names passed as arguments or,
# by default, every tape in the tapes directory.
themes=()
if [ "$#" -gt 0 ]; then
	themes=("$@")
else
	for tape in "$tapes_dir"/*.tape; do
		themes+=("$(basename "$tape" .tape)")
	done
fi

# png_size prints the byte size of a file, or 0 if it does not exist. Handles
# both macOS (stat -f%z) and GNU/Linux (stat -c%s).
png_size() {
	stat -f%z "$1" 2>/dev/null || stat -c%s "$1" 2>/dev/null || echo 0
}

render_one() {
	local name="$1"
	local tape="$tapes_dir/$name.tape"
	local png="doc/themes/$name.png"

	if [ ! -f "$tape" ]; then
		echo "  $name: no tape found at $tape" >&2
		return 1
	fi

	local try size
	for try in $(seq 1 "$max_tries"); do
		vhs "$tape" >/dev/null 2>&1 || true
		size=$(png_size "$png")
		if [ "$size" -ge "$min_size" ]; then
			echo "  $name: ok ($size bytes, attempt $try)"
			return 0
		fi
		echo "  $name: blank capture ($size bytes), retrying ($try/$max_tries)"
	done
	echo "  $name: FAILED after $max_tries attempts" >&2
	return 1
}

failed=()
for name in "${themes[@]}"; do
	echo ">>> $name"
	render_one "$name" || failed+=("$name")
done

# Remove the throwaway GIFs VHS writes while recording.
rm -f "$tapes_dir"/_scratch*.gif doc/themes/_scratch*.gif

if [ "${#failed[@]}" -gt 0 ]; then
	echo "failed to render: ${failed[*]}" >&2
	exit 1
fi

echo "all screenshots rendered."
