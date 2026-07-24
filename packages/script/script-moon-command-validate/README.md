# moon-command-validate

Deterministically validates Xonovex command packages. The generic pass parses every
command's YAML frontmatter, verifies its public title and thin delegation shape, and
checks that hard skill delegations are declared by both package and Claude plugin
manifests. The delegated operation must be registered as an exact reference in the
owner skill. Interchangeable supporting skills are selected at runtime from installed
skill descriptions rather than declared in command metadata.
