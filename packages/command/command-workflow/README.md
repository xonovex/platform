# Composable Workflow Commands

Twelve independently invocable commands share one contract: select one operation,
provide a subject, and add only the dimensions needed for that call. The command
semantics do not depend on a lifecycle stage, role, trigger, executor, or agent
maturity label. Each public command preserves its argument contract and delegates
the operation procedure to the required `workflow-guide` skill.

## Guides

- [Role lenses](docs/role-lenses.md) show illustrative compositions for PM/PO, UX,
  developer, QA, and developer-reviewer perspectives.
- [Provider-native references](docs/references.md) explain how subjects, revisions,
  and result destinations are resolved without a universal identifier layer.
- [Invocation and execution](docs/invocation.md) separates how a call starts from
  who or what executes it.
- [Operation model](../../diagram/diagram-agent-workflow/operation-model.png) shows
  the operations, selection axes, invocation inputs, and results without prescribing
  a stage order.

## Operations

The eight core operations are siblings. Each call performs exactly the operation
named by its command; callers compose calls only when their work needs more than one.

| Command                            | Operation                                                                          |
| ---------------------------------- | ---------------------------------------------------------------------------------- |
| [`create`](commands/create.md)     | Produce a new result without changing the subject.                                 |
| [`review`](commands/review.md)     | Evaluate an exact subject against explicit criteria without changing it.           |
| [`revise`](commands/revise.md)     | Produce a traceable new revision from explicit feedback.                           |
| [`decide`](commands/decide.md)     | Record a descriptive outcome and rationale without granting authority.             |
| [`execute`](commands/execute.md)   | Carry out one bounded subject and report the observable result.                    |
| [`validate`](commands/validate.md) | Check an exact subject against explicit criteria and return evidence.              |
| [`publish`](commands/publish.md)   | Publish an exact subject to an explicit destination and return its native locator. |
| [`abandon`](commands/abandon.md)   | Stop work while preserving the reason, partial result, and retry boundary.         |

The four workspace utilities manage explicitly selected workspaces. They sit outside
the operation model: no core operation implicitly creates, merges, abandons, or cleans
a workspace.

| Command                                              | Utility                                                                 |
| ---------------------------------------------------- | ----------------------------------------------------------------------- |
| [`workspace-create`](commands/workspace-create.md)   | Create one workspace from an exact source.                              |
| [`workspace-merge`](commands/workspace-merge.md)     | Validate and merge one workspace into one destination.                  |
| [`workspace-abandon`](commands/workspace-abandon.md) | Preserve the reason and recoverable state when abandoning a workspace.  |
| [`workspace-cleanup`](commands/workspace-cleanup.md) | Preview and remove only explicitly selected stale or merged workspaces. |

## Independent dimensions

Each dimension answers a separate question. Supplying one must not silently select a
value for another.

| Dimension        | Question it answers                                                       | Contract                                                                        |
| ---------------- | ------------------------------------------------------------------------- | ------------------------------------------------------------------------------- |
| Operation        | What should happen?                                                       | One of the eight command verbs.                                                 |
| Kind             | What kind of subject or result is involved?                               | An open domain selection, inferred only when unambiguous.                       |
| Perspective      | From which evidence or stakeholder lens should the subject be considered? | Changes emphasis, not command semantics.                                        |
| Method           | Which procedure should guide the operation?                               | Selects a method capability when one is requested.                              |
| Executor         | Who or what performs the call?                                            | A human, deterministic script, LLM, or agent; it does not change the operation. |
| Agent capability | Which installed skill, tool, or adapter is needed?                        | Loaded only when selected or unambiguous.                                       |
| Trigger          | What initiated the call?                                                  | Manual use, a harness hook, CI/CD, a webhook, or a scheduler.                   |
| Provider         | Which native system resolves or persists the subject or result?           | Owns authentication, locator interpretation, revisions, and effects.            |
| Reference        | Where is the provider-native subject, evidence, or destination?           | Remains opaque to the command.                                                  |

The common arguments also keep criteria, supporting references, source revisions, and
result destinations explicit. Refer to each command file for its required arguments
and side-effect boundary.

## Capability selection

Commands always load the required `workflow-guide` operation, then load only the
additional capabilities needed by the selected kind, perspective, method, or
provider. A selected capability must be installed and available at execution time.
If it is missing, the call stops and identifies the missing capability. It does not
substitute another capability, invent an umbrella domain, method, or provider
capability, or silently fall back to a local file or provider.

## Composition examples

These are examples, not a required workflow:

```text
create -> review -> revise
review -> publish
execute -> validate
create -> decide
```

Any operation may be called alone. Repeated operations and different orderings are
valid when the subject and selected method require them.

## Installation

### Claude Code

```bash
claude plugin marketplace add xonovex/platform
claude plugin install xonovex-workflow@xonovex-marketplace
```

### Codex

```bash
codex plugin marketplace add xonovex/platform
codex plugin add xonovex-workflow@xonovex-marketplace
```
