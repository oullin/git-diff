#!/bin/sh
# Launcher script written by the "Install Terminal Helper" menu action.
#
# When invoked as `git-diff <arg> ...`, it forwards each argument to the
# packaged macOS app. Existing filesystem paths are resolved to absolute
# paths first so the app sees the user's actual cwd; commit-SHA tokens and
# flags pass through unchanged.

set -e

if [ "$#" -eq 0 ]; then
  exec open -a "Git Diff Review"
fi

resolved=""

for arg in "$@"; do
  if [ -e "$arg" ]; then
    abs=$(cd "$(dirname "$arg")" 2>/dev/null && pwd)/$(basename "$arg")
    if [ -n "$abs" ]; then
      resolved="$resolved $abs"
      continue
    fi
  fi
  resolved="$resolved $arg"
done

# shellcheck disable=SC2086
exec open -a "Git Diff Review" --args $resolved
