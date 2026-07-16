# Module Lifecycle Management

## Operations

- **List/inspect** — resolve catalog identity, version, provenance, owner, permissions, capabilities, compatibility, state, evidence, and drift.
- **Enable** — validate dependencies, pinned provenance, trust review or consent, profile fit, conflicts, permissions, data flows, and enforcement guarantees; preview before authorization.
- **Disable** — prove mandatory controls remain covered, preview lost capabilities and evidence, then verify native disablement.
- **Upgrade** — compare contract/capability/permission/provenance changes, pin and verify the target, run concurrency/failure probes, preserve verified rollback, and monitor drift.
- **Remove/retire** — verify dependents and evidence retention, disable safely, remove native state, and record retirement/rollback references.

## Mutation gate

Every changing operation follows discover → assess → preview → authorize → apply → verify → record. Default to preview. Stop when the adapter cannot provide idempotency, verification, rollback, or adequate authority for the requested change.

Module management never edits a universal registry. It invokes the owning package, harness, configuration, or provider adapter and returns opaque native references.
