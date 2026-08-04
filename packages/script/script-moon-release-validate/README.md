# script-moon-release-validate

Repository-level release validation: lockstep plugin versions across both
marketplaces and the lockfile, root README link resolution, release
workflow safety, and the alignment between `.gitignore` and
`hasher.ignorePatterns` that the glob walk strategy depends on. Runs from
the workspace root via the `release-validate` Moon task, which is part of
this project's `ci-check`.

Package-local documentation checks stay with their owning packages. This
script owns every check that needs repository-wide context.
