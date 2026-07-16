# Result Provider Port

The semantic workflow plane owns the result-provider port. Provider adapters implement it using native resources, authentication, identifiers, revisions, relationships, and side effects. The port carries domain values and opaque references; it never exposes a provider SDK type or requires a common file, database, or serialized envelope.

## Required capabilities

| Capability     | Semantic guarantee                                                                                                   |
| -------------- | -------------------------------------------------------------------------------------------------------------------- |
| `resolve`      | Reconstruct provider context, result kind, native reference/revision, source relationships, and available operations |
| `read`         | Return the semantic result visible at the requested native reference or revision                                     |
| `publish`      | Create a provider-native result and return its opaque reference plus provider-derived revision                       |
| `revise`       | Apply an authorized change with stale-revision detection where the provider can guarantee it                         |
| `relate`       | Create or resolve provider-native relationships without inventing a workflow-wide identity                           |
| `version`      | Return an immutable revision or an explicitly weaker freshness token and its limitations                             |
| `capabilities` | Report supported operations, guarantees, limits, tested provider version, and review date                            |

An adapter may name operations differently when conformance demonstrates the same semantics. It must report unsupported exact-revision, atomicity, relationship, or rollback behavior rather than simulate a guarantee.

## Conformance rules

- Resolve a native reference after process restart from provider state, not conversation memory, a filename convention, or a runtime trace.
- Preserve the canonical result kind, sources, actor origin, provider revision, and independent publication boundary.
- Bind a revision-sensitive read or mutation to the strongest native version token available and identify whether it is immutable, optimistic, or freshness-only.
- Make publishing and revising safe under retries through a provider idempotency facility or an adapter-owned idempotency key. When neither exists, expose the duplicate risk and require caller reconciliation.
- Return native conflicts, authorization failures, rate limits, unavailable capabilities, and partial success explicitly. Never fall back to another provider for a side effect.
- Keep credentials and provider-specific payloads inside the adapter. Redact sensitive provider context when returning a handle or inspection report.

## Portability fixtures

The required reference fixture is a self-controlled non-file task-system fake. It publishes, revises, relates, snapshots provider state, starts a new provider process, and reconstructs the same opaque reference. Repository-backed, GitHub Issues, and GitLab Issues adapters are optional examples: they must pass the same semantic port while declaring their weaker or provider-specific guarantees.

Reusable assertions and the non-file fake live in `scripts/conformance-helpers.mjs`. The JSON fixture describes test inputs only; it is not a provider manifest or persisted result format.
