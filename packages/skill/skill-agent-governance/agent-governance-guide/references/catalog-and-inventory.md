# Module/Profile Catalog and Effective Inventory

## Catalog independently installable capabilities

Each module/profile entry records:

- identity/version, owner/support/escalation, source/provenance/trust, digest/signature/attestation where selected;
- classification, adoption modes, authority zones, requested/effective permissions, tools, files, network, secrets, models, providers, data flows, and side effects;
- capabilities, semantic intents, enforcement guarantees, evidence origins, failure behavior, timeouts/retries/concurrency/idempotency;
- platform/provider editions, tested versions, compatibility, dependencies, conflicts, presets, limitations, and cost;
- lifecycle status, usage/effective-selection evidence, deprecation/replacement, update/canary/rollback/disable, drift, and retirement.

Search and filter by capability, domain, owner, classification, adoption mode, authority zone, platform, compatibility, trust, permission, lifecycle status, conflict, and profile. The catalog is descriptive; actual effective configuration comes from authoritative native adapters.

## Show adoption mode separately from authority

Classify executable effect as:

- `knowledge-only` — instructions or reference content with no claimed execution/enforcement;
- `advisory` — emits recommendations or non-blocking feedback;
- `evidence-producing` — inspects or evaluates and publishes evidence without applying policy;
- `enforcing` — can allow, deny, require evidence, or gate at a declared point;
- `configuration-changing` — mutates durable native settings;
- `privileged` — can affect protected systems, identities, data, deployment, or organization controls.

Show this classification alongside adoption mode (`workflow-only`, `governance-only`, `enablement-only`, `external-enforcement-only`, `integrated`) and authority zone (`organization-managed`, `project`, `user`, `session-runtime`, `external`). These axes answer different questions and must not be collapsed into a single risk label.

## Resolve a compatible effective composition

1. Discover native versions, editions/tiers, capabilities, configuration, authority, and installed/available modules.
2. Select profile requirements and candidate modules by semantic capability.
3. validate tested compatibility, trust, permissions, dependencies, conflicts, presets, enforcement coverage, data/telemetry rules, and lifecycle status;
4. union strengthening requirements and reject unauthorized weakening or authority expansion;
5. preview exact selected modules and native changes, then authorize/apply/verify through owning adapters;
6. publish the effective selection, gaps, rejected candidates, native references, and drift baseline.

Installed does not mean selected; selected does not mean enabled; enabled does not prove effective execution or enforcement.

## Inventory the effective environment

When an Inventory/AIBOM profile selects agent-environment scope, include the effective harness/configuration, skills, plugins, hooks, extensions, MCP servers, models/routes, tools, CI components, policy bundles, governance modules, providers, versions, provenance, requested/effective permissions, authority zones, data flows, and relationships observable from deterministic sources.

Distinguish available, installed, selected, enabled, observed-effective, and evidence-producing states. Include executable governance modules only when the selected scope requires them; record missing or unobservable categories as gaps. Use **workflow-guide** for the Inventory result and provider publication contract.

## Deprecate and retire explicitly

Deprecated entries name replacement, migration, end-of-support/enablement/execution dates, evidence retention, compatibility implications, and rollback. Retired modules cannot remain effectively selected. Verify replacement coverage and native disable/removal before catalog closure.
