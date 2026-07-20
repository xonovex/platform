# Role Lenses

A role lens is an illustrative way to compose the same eight operations. It is not a
permission model, mandatory handoff, stage sequence, or role-specific command set.
Every caller uses the same command semantics.

The noun selected with `--kind` describes the subject or result. The value selected
with `--perspective` describes the evidence or concerns to emphasize. Both can change
between roles without changing what `create`, `review`, `revise`, `decide`, `execute`,
`validate`, `publish`, or `abandon` means.

## Illustrative compositions

```text
PM/PO:              create -> review -> revise -> decide
UX:                 create -> review -> revise -> decide
Developer:          create -> review -> revise -> execute -> validate -> publish
QA:                 create -> review -> execute -> validate -> publish
Developer reviewer: review -> publish
```

These compositions are examples. A call can omit, repeat, or reorder operations when
the subject and selected method call for it.

### PM/PO

A PM/PO might create and refine requirements, review them from a product perspective,
and record a descriptive decision. The kind may be a problem statement, requirement,
or plan. `decide` records an outcome and rationale; it does not grant operational
authority or change a gate.

### UX

A UX practitioner might create and refine a design result, review it from interaction,
content, research, or accessibility perspectives, and record a decision. Those
perspectives select applicable evidence and capabilities rather than a different set
of commands.

### Developer

A developer might create and revise a plan or implementation, execute bounded work,
validate the result, and publish it to an explicit destination. Publishing is separate
from execution and occurs only when the destination and effect are explicit.

### QA

QA creates or reviews test material, executes checks, validates behavior against
criteria, and publishes the resulting evidence. QA owns validation of behavior and
evidence in this example. QA does not review or approve the pull request containing
the implementation.

### Developer reviewer

A separate developer reviewer reviews the exact pull-request revision and records the
review or approval through the selected source-control provider. The illustrative
`review -> publish` composition keeps evaluation separate from the provider action
that publishes the disposition. This reviewer, not QA, owns pull-request review and
approval in these examples.

There are no aliases such as `qa-validate` or `developer-review`. Role names do not
imply permissions, authority, required handoffs, or provider effects.

## Related guides

- [Command inventory](../README.md)
- [Provider-native references](references.md)
- [Invocation and execution](invocation.md)
- [Operation model](../../../diagram/diagram-agent-workflow/operation-model.png)
