# moon-command-validate

Use this validator to reject command packages that violate the Xonovex delegation, manifest, or prose contracts.

The generic pass parses each command's YAML frontmatter, verifies its public title and thin delegation shape, and checks that package and Claude plugin manifests declare each required skill. The owner skill must register the delegated operation as an exact reference. The runtime selects interchangeable supporting skills from installed skill descriptions instead of command metadata.

The validator also rejects an em dash, an ellipsis character, or a typographic quote in every prose file in the package, in literal or Unicode-escaped form. Command titles use a plain hyphen as the separator.
