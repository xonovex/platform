# Agent-Harness Adapter Contract

## Boundary

The governance owner defines semantic intents, guarantees, evidence, trust, and lifecycle requirements. A harness owner translates those requirements to native events, handlers, configuration scopes, diagnostics, and failure behavior. The adapter never creates a universal hook filename or treats similar event names as equivalent guarantees.

Keep platform URLs and provenance in the harness owner's `SOURCES.md`.

## Adapter declaration

Declare one versioned matrix per product surface:

```text
platform and surface
matrix version and owner
documentation snapshot date
runtime version probe command and observed result
configuration schema version, scopes, precedence, and native paths
trust boundary and managed-configuration behavior
capabilities[]
limitations, probes, and next review date
```

For each capability record:

```text
semantic intent
native event
support: supported | unsupported | experimental | unknown
handler type and whether it executes
coverage: complete | partial | none
blocking and advisory behavior
context input/output behavior
ordering, matching, concurrency, timeout, retry, and reentrancy
configuration precedence and disable behavior
permissions, secrets, filesystem, network, and data exposure
limitation and deterministic conformance probe
```

`runtime version: not installed` is valid evidence that runtime support was not tested. It is never rewritten as a passing runtime result. The documentation snapshot still versions the matrix and starts a refresh clock.

## Translation rules

1. Select the semantic intent before looking at a native event name.
2. Resolve the platform surface and observed version.
3. Reject a stale matrix, unsupported handler, non-executing handler, or partial coverage when the profile requires a complete mandatory guarantee.
4. Preserve native matching, ordering, concurrency, timeout, and output semantics.
5. Keep decision, enforcement, and evidence references separate.
6. Use independent external enforcement when the native surface cannot cover the operation.

Do not infer:

- that `PreToolUse` on two products intercepts the same tools;
- that a non-zero exit blocks without checking the event-specific rule;
- that a parsed handler type executes;
- that managed settings distribute executable files;
- that install, discovery, or a valid manifest establishes trust;
- that one denial cancels already-started sibling handlers.

## Capability selection

A mandatory intent is satisfiable only when all required dimensions pass:

```text
support == supported
handler executes
coverage satisfies the requested operation set
blocking == true when denial is required
failure behavior matches the profile
configuration authority cannot be weakened by a lower scope
runtime or accepted documentation-only evidence is fresh enough for the profile
```

Experimental support may be selected for an advisory or evidence-only profile when limitations are disclosed. Unknown support is a gap, not a best-effort match.

## Conformance assets

`assets/harness-conformance-fixtures.json` contains documentation-versioned platform matrices and semantic cases. `assets/harness-module-templates.json` contains semantic templates that must be translated by a harness owner. Run `node scripts/validate-harness-fixtures.mjs`; the validator rejects silent degradation, incomplete onboarding, unsafe trust, assumed serial execution, and unbounded evaluator or agent templates.

## Runtime implementation status

The bundled enforcing harness adapter is deliberately narrow: Claude Code `PreToolUse` for `Edit` and `Write`, implemented by `governance-pre-tool-use.sh` and backed by a pinned live capability probe. It validates the full decision contract and records a separate enforcement receipt. The other harness matrices and templates are documentation/conformance assets until their own native adapter and runtime probe exist; they must not be reported as installed or runtime-verified.

CI/CD hooks, provider webhooks, schedules, sensors, APIs, manual calls, and agent-originated events enter through the source-neutral workflow trigger contract. That contract proves origin preservation and execution independence; it does not claim that a particular hosted provider has been configured.
