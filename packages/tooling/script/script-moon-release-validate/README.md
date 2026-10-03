# script-moon-release-validate

Use this validator to enforce repository-wide release contracts before publication. It checks lockstep plugin versions across both marketplaces and the lockfile, root README links, release workflow safety, and the required alignment between `.gitignore` and `hasher.ignorePatterns`.

Run the `release-validate` Moon task from the workspace root. The task is part of this project's `ci-check`. Package-local documentation checks stay with their owning packages.
