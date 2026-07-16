# Governance Provider Ports

Governance owns three driven ports in its own vocabulary. A deterministic rule set, policy engine, configuration API, telemetry backend, evidence database, or hosted platform is an adapter. The adapter translates native state into semantic decisions and opaque references; the governance core never imports its payload schema, SDK type, storage format, or precedence model.

## Policy provider

Required semantic operations are `evaluate`, `request-evidence`, `explain`, `resolve-exception`, `version`, and `replay`.

- Evaluate explicit subject, action, actor, authority, contextual facts, applicability, policy/profile versions, and evidence freshness deterministically for the same inputs and versions.
- Return outcome, reasons, required evidence or remediation, explanation, expiry, limitations, and an opaque decision reference. Keep the decision distinct from enforcement and enforcement evidence.
- Resolve exceptions and break-glass references only within their declared scope, authority, time, compensating controls, and review requirements.
- Retain or reconstruct the policy version and input/evidence references required to replay a historical decision. Replay reports missing or changed inputs rather than silently evaluating current state as history.
- Treat a native policy language or engine as optional. A deterministic rules adapter and an OPA adapter conform when they produce the same semantic decision for the same versioned facts.

## Configuration provider

Required semantic operations are `inspect`, `diff`, `preview`, `apply`, `verify`, `rollback`, `export`, `import`, and `detect-drift`.

- Inspect native effective state and authority sources read-only before proposing a change.
- Diff intended and observed state; preview exact native mutations, permissions, secrets, data flows, side effects, target version, verification, and rollback before authorization.
- Apply idempotently against an expected native version. Return separate references for authorization, apply outcome, verification, and any partial application.
- Verify by re-reading authoritative native state. A successful API response alone is not verification.
- Roll back or explicitly declare irreversibility and fail the mutation gate before apply when the selected profile requires rollback.
- Export/import are adapter-native interchange capabilities, not a universal governance file. Detect drift against the intended authority source and version.

## Evidence and telemetry provider

Required semantic operations are `publish`, `resolve`, `correlate`, `redact`, `retention`, `authorize-access`, and `version`.

- Publish provider-native evidence with subject/revision, origin, time window, policy/control/module version, freshness, limitations, and an opaque reference.
- Correlate executions and evidence without promoting trace/session identifiers into workflow identity.
- Minimize fields and content before collection. Raw prompts, tool outputs, source, secrets, personal data, and model inputs/outputs require explicit purpose and authorization.
- Declare redaction, sampling, retention, residency, deletion, and access rules; test them at the provider boundary.
- Preserve authoritative external evidence by reference. Do not copy sensitive payloads merely to normalize them.

## Conformance

For every port, test an adapter that does not use files or a shared serialized manifest. Exercise provider outage, stale version, unauthorized access, duplicate request, partial change, rollback, drift, redaction, retention, and reconstruction after process restart. Reusable assertions live in `scripts/conformance-helpers.mjs`.
