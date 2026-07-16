# script-moon-release-validate

Repository-level release and traceability validation: lockstep plugin
versions across both marketplaces and the lockfile, root README link
resolution, plan-set traceability (source/decision/control ID resolution,
subplan task counts), harness and enterprise eval presence, conformance
fixture coverage terms, and diagram content. Runs from the workspace root
via the `release-validate` moon task, which is part of this project's
`ci-check`.

Package-local documentation checks stay with their owning packages
(for example `command-workflow`'s `validate-documentation.mjs`); this
script owns every check that needs repository-wide context.
