#!/usr/bin/env bash
# Prints the next patch tag when the Go module changed since the latest
# release, and nothing when it did not. Used by the weekly release workflow.
#
# Usage: scripts/next-release-tag.sh [ref]   (ref defaults to HEAD)
#
# Only tags of the form vX.Y.Z reachable from ref count as releases; the
# latest one is chosen by version, not by date. A change to go.mod, go.sum or
# any *.go file is what changes the binary that `go install` builds.
set -euo pipefail

ref="${1:-HEAD}"

latest=$(git tag --merged "$ref" --list 'v*' |
  grep -E '^v[0-9]+\.[0-9]+\.[0-9]+$' |
  sort -V |
  tail -n 1 || true)
if [ -z "$latest" ]; then
  echo "no vX.Y.Z tag is reachable from $ref" >&2
  exit 1
fi

if git diff --quiet "$latest" "$ref" -- go.mod go.sum '*.go'; then
  echo "no change to go.mod, go.sum or *.go since $latest" >&2
  exit 0
fi

IFS=. read -r major minor patch <<<"${latest#v}"
echo "v${major}.${minor}.$((patch + 1))"
