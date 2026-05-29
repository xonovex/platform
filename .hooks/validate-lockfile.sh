#!/usr/bin/env bash
set -Eeuo pipefail

# Validate that package-lock.json is in sync with the workspaces at commit time,
# so CI's `npm ci` (which refuses to run against an out-of-date lockfile) never
# fails because someone added/removed a workspace package without updating the
# lock.
#
# This validates and blocks; it never mutates the commit. Regenerating the lock
# at commit time pulls in unrelated transitive version drift and is
# non-deterministic across npm versions, so we leave that to the developer.
#
# Runs only when a package.json (root or any workspace) is staged. Verifies
# every workspace package is present in the lockfile's "packages" map and blocks
# with guidance if any is missing.

root="$(git rev-parse --show-toplevel)"
cd "$root"

# Nothing to do unless a package.json is part of this commit.
if ! git diff --cached --name-only --diff-filter=ACMRD | grep -Eq '(^|/)package\.json$'; then
  exit 0
fi

missing=0
while IFS= read -r pkg; do
  dir="${pkg%/package.json}"
  grep -q "\"${dir}\":" package-lock.json || { echo "validate-lockfile: ${dir} is missing from package-lock.json"; missing=1; }
done < <(git ls-files 'packages/*/*/package.json' 'packages/*/package.json')

if [ "$missing" -ne 0 ]; then
  echo "validate-lockfile: package-lock.json is out of sync. Run 'npm install' and stage package-lock.json, then commit again." >&2
  exit 1
fi
exit 0
