# script-moon-release-validate

Repository-level release validation: lockstep plugin versions across both
marketplaces and the lockfile, root README link resolution, and release
workflow safety. Runs from the workspace root via the `release-validate`
Moon task, which is part of this project's `ci-check`.

Package-local documentation checks stay with their owning packages
(for example `command-workflow`'s `validate-documentation.mjs`); this
script owns every check that needs repository-wide context.
