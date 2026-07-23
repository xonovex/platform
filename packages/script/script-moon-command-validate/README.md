# moon-command-validate

Deterministically validates Xonovex command packages. The generic pass parses every
command's YAML frontmatter, verifies its public title and thin delegation shape, and
checks that hard skill delegations are declared by both package and Claude plugin
manifests. The delegated operation must be registered as an exact reference in the
owner skill. Semantic requirements resolve against the catalog: unavailable required
support fails validation, while unavailable preferred support is a warning.

A command package can additionally supply a declarative contract:

```bash
npx moon-command-validate packages/command/command-workflow \
  --contract packages/command/command-workflow/contracts/workflow-commands.v1.json
```

The workflow contract pass compares the command inventory, argument hints,
delegations, skill references, operation families, and effect boundaries against the
machine-owned interface. It performs no provider calls and does not treat prompt
permissions as runtime authorization.

Workflow request composition definitions are generated from the shared Zod contract
and checked for drift:

```bash
npx moon-command-workflow-schema-sync \
  packages/skill/skill-workflow/workflow-guide/assets/workflow-request.schema.json
```
