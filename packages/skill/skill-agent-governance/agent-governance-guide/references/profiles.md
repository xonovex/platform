# Governance Profile Composition

## Facets

A profile composes independent requirements across:

- lifecycle evidence, when a workflow plane is selected;
- governance policies, applicability, actors, exceptions, and failure behavior;
- permitted and preferred executor classes plus budgets and validation;
- required semantic intents and enforcement guarantees;
- data classification, consent, redaction, retention, residency, and access;
- telemetry semantics, sampling, cost, content capture, and evidence publication;
- module provenance, compatibility, trust, pinning, upgrade, rollback, and retirement.

## Effective composition

1. Resolve applicable profiles and their authority zones.
2. Union strengthening requirements by semantic identity.
3. Detect incompatible actor, data, executor, module, or enforcement requirements.
4. Require an authorized exception for any weakening.
5. Match mandatory controls to supported, non-experimental enforcement points with adequate guarantees and explicit failure behavior.
6. Preserve policy decision, enforcement, and evidence as separate references.
7. Fail visibly on unknown applicability, stale policy, unsupported capability, unresolved conflict, or authority expansion when the profile claims a mandatory guarantee.

Profiles do not prescribe one policy engine, harness, hook schema, provider, or configuration format. A small team may use deterministic rules and provider-native controls; a larger organization may compose a policy service, managed hooks, CI, admission, and protected environments under the same semantics.

## Configuration sources

Organization, project, user, session, and external systems each expose native configuration sources and precedence. Adapters document native behavior. The cross-platform contract specifies authority and non-weakening, not one universal precedence list.
