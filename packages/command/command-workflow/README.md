# Composable Workflow Commands

Twelve thin Claude Code commands expose stable workflow operations and delegate their
semantics to `workflow-guide`. Codex installs and invokes the skill directly.

The public model has five layers:

| Layer                  | User question                      | Examples                                     |
| ---------------------- | ---------------------------------- | -------------------------------------------- |
| Operation              | What immediate work should happen? | create, review, execute                      |
| Semantic selection     | What does good work mean here?     | kind, method, perspectives, criteria         |
| Resource binding       | Which exact artifact is involved?  | provider, opaque reference, native revision  |
| Invocation context     | How may this call run?             | effect mode, host-owned trigger and executor |
| Derived implementation | How can the runtime perform it?    | installed skills, tools, provider adapters   |

These layers are related but are not peer command axes. In particular, users request
work and semantic concerns; the runtime derives implementation capabilities and
reports its choices.

Every invocation normalizes one `WorkflowRequest`, asks the host composition runtime
to resolve it, and loads its unique dependency-first guide order before the
operation. Required failures block; preferred failures produce a visible degraded
result. Exact request overrides and requirement-provider bindings resolve ambiguity
without network lookup.

## Result boundary

Create, review, revise, decide, execute, validate, and abandon return a structured
operation-result envelope inline. They never persist their domain result.

Publish is the only core operation that writes a domain result to an external
destination. A host may separately checkpoint workflow administration state when its
policy authorizes that infrastructure effect; a checkpoint is not publication and
must appear in the result's observed effects.

This separation makes authorization, retry, and auditing visible:

```text
review -> inline result -> publish
execute -> inline result -> validate -> inline result -> publish
```

## Commands

The eight core operations are siblings. They can stand alone, repeat, run in
parallel, or compose in any order the work requires.

| Command                            | Operation                                                |
| ---------------------------------- | -------------------------------------------------------- |
| [`create`](commands/create.md)     | Produce a new inline result without changing its source. |
| [`review`](commands/review.md)     | Return evidence-linked findings about an exact subject.  |
| [`revise`](commands/revise.md)     | Return a traceable successor from explicit feedback.     |
| [`decide`](commands/decide.md)     | Record a descriptive outcome without granting authority. |
| [`execute`](commands/execute.md)   | Perform bounded effects under an explicit effect mode.   |
| [`validate`](commands/validate.md) | Return one evidence-backed result per binding criterion. |
| [`publish`](commands/publish.md)   | Persist an exact result to one explicit destination.     |
| [`abandon`](commands/abandon.md)   | Return a reason, partial state, and retry boundary.      |

Workspace operations are explicit primitives outside the core model:

| Command                                              | Exclusive responsibility                    |
| ---------------------------------------------------- | ------------------------------------------- |
| [`workspace-create`](commands/workspace-create.md)   | Create named workspace resources.           |
| [`workspace-merge`](commands/workspace-merge.md)     | Validate and integrate without removal.     |
| [`workspace-abandon`](commands/workspace-abandon.md) | Record abandonment without mutation.        |
| [`workspace-cleanup`](commands/workspace-cleanup.md) | Remove exact previewed workspace resources. |

## Simple and advanced requests

Simple calls use a subject plus subject-specific shorthands:

```text
/xonovex-workflow:review owner/repository#42 \
  --subject-provider github \
  --subject-revision 7f4c2d1 \
  --perspective compatibility \
  --perspective security
```

When subject, evidence, policy, and destination use different providers, use
`--request <file>`. The request binds every named resource independently. A binding
keeps its provider-native reference opaque and carries its own revision and semantic
kind.

Do not flatten a multi-provider request into one global provider or revision. See
[Provider-native resource bindings](docs/references.md).

## Perspectives, roles, and criteria

Perspectives are repeatable semantic lenses. They add questions, evidence needs, and
advisory criteria without changing the operation or granting authority.

Role lenses are optional conveniences that resolve into suggested perspectives,
criteria, evidence, and communication depth. They never identify the executor,
authorize an effect, claim ownership, or prescribe a lifecycle. See
[Role lenses](docs/role-lenses.md).

Criteria carry provenance and status:

- explicit caller and authoritative artifact or policy criteria can be binding;
- method, perspective, and model suggestions are advisory until accepted;
- every derived criterion records its source and reason;
- validation blocks when it cannot resolve any binding criterion.

Constraint resolution defaults to `assisted`. `strict` accepts only explicit or
authoritative constraints. `automatic` may apply high-confidence advisory checks but
can never invent a binding merge, release, or publication gate.

## Effects and authority

| Command           | Effect modes            | Default       |
| ----------------- | ----------------------- | ------------- |
| execute           | inspect, preview, apply | inspect       |
| publish           | preview, apply          | preview       |
| workspace-create  | preview, apply          | preview       |
| workspace-merge   | preview, apply          | preview       |
| workspace-cleanup | preview, apply          | preview       |
| all others        | no external effect      | inline result |

The runtime or provider enforces authentication, authorization, approvals,
idempotency, concurrency, retry limits, and budgets. A skill explains procedure but
cannot grant those powers. `decide`, `review`, and `validate` are always descriptive.

## Durable continuation

Chat history is not workflow storage. Continue cross-session work from an exact
provider-native work record and revision. The work record retains requests, result
envelopes, criteria, selected skills and versions, effects, unresolved questions,
child work, and retry boundaries.

Same-conversation shorthand may refer to the immediately preceding result. A later
session must supply the persisted native reference. The workflow skill owns the
request, result, and work-record schemas.

## Migration

Version 1 of the composition contract replaces ambiguous global flags and overlapping
effects without compatibility aliases. See [Contract migration](docs/migration.md).

## Guides

- [Provider-native resource bindings](docs/references.md)
- [Invocation, effects, and execution](docs/invocation.md)
- [Role lenses](docs/role-lenses.md)
- [Historical design handout](docs/history/workflow-and-skill-composition-handout.md)
- [Operation model](../../diagram/diagram-agent-workflow/operation-model.png)

## Installation

### Claude Code

```bash
claude plugin marketplace add xonovex/platform
claude plugin install xonovex-workflow@xonovex-marketplace
```

### Codex

```bash
codex plugin marketplace add xonovex/platform
codex plugin add xonovex-skill-workflow@xonovex-marketplace
```

Invoke `$workflow-guide` in Codex and name the operation plus its request. The
`/xonovex-workflow:*` namespace is the Claude Code surface.
