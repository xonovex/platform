# moon-command-validate

Deterministically validates Xonovex command packages. The generic pass parses every
command's YAML frontmatter, verifies its public title and thin delegation shape, and
checks that hard skill delegations are declared by both package and Claude plugin
manifests. The delegated operation must be registered as an exact reference in the
owner skill. Semantic requirements resolve against the catalog: unavailable required
support fails validation, while unavailable preferred support is a warning.
