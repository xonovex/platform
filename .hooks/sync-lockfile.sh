#!/usr/bin/env bash
set -Eeuo pipefail

# Keep package-lock.json in sync with the workspaces at commit time, so CI's
# `npm ci` (which refuses to run against an out-of-date lockfile) never fails
# because someone added/removed a workspace package without updating the lock.
#
# Runs only when a package.json (root or any workspace) is staged.
#   - Online: regenerates the lockfile and stages it into the same commit.
#   - Offline / npm unavailable: falls back to a structural check and blocks the
#     commit with a clear message if a workspace is missing from the lockfile.

root="$(git rev-parse --show-toplevel)"
cd "$root"

# Nothing to do unless a package.json is part of this commit.
if ! git diff --cached --name-only --diff-filter=ACMRD | grep -Eq '(^|/)package\.json$'; then
  exit 0
fi

if command -v npm >/dev/null 2>&1 && \
   npm install --package-lock-only --prefer-offline --no-audit --no-fund >/dev/null 2>&1; then
  if ! git diff --quiet -- package-lock.json; then
    git add package-lock.json
    echo "sync-lockfile: package-lock.json was stale; regenerated and staged into this commit."
  fi
  exit 0
fi

# Fallback (offline or npm failed): verify every workspace package is present in
# the lockfile's "packages" map; block with guidance if not.
missing=0
while IFS= read -r pkg; do
  dir="${pkg%/package.json}"
  grep -q "\"${dir}\":" package-lock.json || { echo "sync-lockfile: ${dir} is missing from package-lock.json"; missing=1; }
done < <(git ls-files 'packages/*/*/package.json' 'packages/*/package.json')

if [ "$missing" -ne 0 ]; then
  echo "sync-lockfile: package-lock.json is out of sync. Run 'npm install' and stage package-lock.json, then commit again." >&2
  exit 1
fi
exit 0
