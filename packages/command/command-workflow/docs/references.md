# Provider-Native Resource Bindings

A resource binding says which provider owns one named resource, gives that provider's
opaque reference, and pins its native revision when the resource is mutable.

Provider, reference, revision, and kind belong to the binding. They are not global
workflow settings.

## Simple subject shorthand

A simple command can bind one subject directly:

```text
/xonovex-workflow:review owner/repository#42 \
  --subject-provider github \
  --subject-revision 7f4c2d1 \
  --subject-kind pull-request
```

The command does not parse `owner/repository#42` into a universal Xonovex identifier.
The GitHub provider interprets it. Command normalization wraps the shorthand in a
read-intent locator, supplies the operation's default output kind, creates empty
selection collections with `assisted` resolution, and sets the operation's default
effect mode.

## Named bindings

Use `--request <file>` when a call has multiple resources, providers, revisions, or
slots:

```json
{
  "contractVersion": "xonovex.workflow.request/v1",
  "operation": "review",
  "subject": {
    "kind": "pull-request",
    "intent": "read",
    "locator": {
      "provider": "github",
      "reference": "owner/repository pull request #42",
      "revision": "head-7f31"
    }
  },
  "inputs": {
    "evidence": [
      {
        "kind": "trace-set",
        "intent": "read",
        "locator": {
          "provider": "datadog",
          "reference": "trace query native reference",
          "revision": "immutable-event-set-19"
        }
      },
      {
        "kind": "story",
        "intent": "read",
        "locator": {
          "provider": "jira",
          "reference": "PROJECT-123",
          "revision": "17"
        }
      }
    ]
  },
  "selection": {
    "method": "pull-request-review",
    "perspectives": ["compatibility", "security", "reliability"],
    "criteria": [
      {
        "id": "story-acceptance",
        "statement": "The exact story acceptance criteria are satisfied.",
        "binding": "binding",
        "sourceInput": "evidence"
      }
    ],
    "roleLenses": [],
    "resolutionMode": "assisted",
    "acceptedSuggestions": []
  },
  "effect": {
    "mode": "inspect"
  },
  "output": {
    "kind": "review-report"
  }
}
```

This is the packaged
[multi-provider review fixture](../../../skill/skill-workflow/workflow-guide/assets/examples/multi-provider-review-request.json).
The structure is validated against the request schema shipped with `workflow-guide`.
JSON and YAML representations have the same logical fields.

## Resolution

- Infer a provider only when one provider is unambiguous; record the inference and its
  basis.
- Require an exact native revision before judging or changing a mutable resource.
- Keep source, evidence, policy, work record, and destination as separate named
  bindings.
- Stop rather than interpreting one locator through an arbitrary provider.
- Never replace an unavailable provider with a local file or another provider.

The runtime coordinates policy, authority checks, approvals, idempotency, retry
budgets, and audit. Each selected provider enforces its native authentication,
authorization, locator and revision semantics, concurrency, relationships, and
effects.

## Inline and persisted results

Every operation returns an inline operation-result envelope. A domain result becomes
provider-persisted only through `publish`, which takes a separate destination binding
and returns its observed native locator and revision.

For Publish shorthand, `--destination-revision` normalizes to an update binding with
that exact expected revision. Omitting it normalizes to a collision-safe create; it
never means overwrite-whatever-exists.

A host-owned workflow checkpoint is administrative state, not a domain result
destination. The host may write it only under explicit runtime policy and must record
the checkpoint effect.

## Related guides

- [Command inventory](../README.md)
- [Role lenses](role-lenses.md)
- [Invocation, effects, and execution](invocation.md)
- [Contract migration](migration.md)
