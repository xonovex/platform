# moon-command-validate

Deterministically validates Xonovex command packages. The generic pass parses every
command's YAML frontmatter, verifies its public title and thin delegation shape, and
checks that hard skill delegations are declared by both package and Claude plugin
manifests.

A command package can additionally supply a declarative contract:

```bash
npx moon-command-validate packages/command/command-workflow \
  --contract packages/command/command-workflow/contracts/workflow-commands.v1.json
```

The workflow contract pass compares the command inventory, argument hints,
delegations, skill references, operation families, and effect boundaries against the
machine-owned interface. It performs no provider calls and does not treat prompt
permissions as runtime authorization.
