# Security and Policy Enforcement

## Executable module trust

Hooks, scripts, plugins, extensions, packages, MCP servers, skill scripts, CI modules, model evaluators, and agent launchers execute with user, repository, organization, CI, cloud, or provider authority. Discovery or installation does not grant trust.

Before activation, pin and review source, version, digest, publisher or signer, dependencies, install scripts, compatibility, permissions, tools, filesystem roots, network destinations, secrets, model/provider use, data flows, side effects, concurrency, failure behavior, update path, disable/removal, and rollback. Re-review on any authority or provenance change.

Use package locks, immutable revisions, checksums, signatures, transparency, attestations, or protected provider releases as appropriate to the threat model. No one provenance mechanism is mandatory, but a moving or mismatched source cannot satisfy a pinned mandatory module.

## Least privilege and data flow

- Grant only declared tools, paths, commands, network destinations, secrets, providers, and data classes. Session authority expires and cannot create durable higher authority.
- Keep secret values out of prompts, process arguments, logs, artifacts, telemetry, generated configuration, and source. Prefer native secret stores and short-lived workload identity.
- Treat prompt, context, transcript, source, tool output, model input/output, MCP payload, CI log, provider event, and telemetry content as separate data categories with purpose, classification, minimization, redaction, residency, retention, access, deletion/export, and cost rules.
- A harness permission model, project trust prompt, or `allowed-tools` declaration reduces blast radius but is not a complete sandbox. Use an external isolation boundary when the risk requires it.
- MCP capability negotiation and consent do not replace host policy, provider permissions, output validation, network controls, or independent enforcement.

## Policy decision and enforcement

A decision point evaluates versioned policy against subject, action, actor/executor, authority zone, facts, applicability, and evidence. It returns `allow`, `deny`, `ask`, `advise`, `observe`, `require-evidence`, `exception`, or `break-glass`, plus reasons, expiry, and evidence requests.

An enforcement point separately applies that decision at a native harness event, CI gate, repository rule, protected environment, deployment/admission control, identity/provider operation, or accountable human control. Record the decision, enforcement action, native subject/revision, result, and evidence separately.

Start with deterministic rules for explicit facts and provider-native policy when that provider owns the authoritative state or action. OPA/Rego is an optional interchangeable policy-decision implementation when centralized declarative policy adds value; profiles do not require it merely because fixtures demonstrate it.

## Defense in depth and failure

A hook can cover one event or product surface but cannot establish an organization-wide guarantee by itself. Match mandatory controls to enforcement with sufficient coverage, blocking, ordering, integrity, availability, bypass resistance, and evidence. Add repository, CI, deployment, admission, identity, cloud, or provider controls when the harness boundary is inadequate.

Every control and module declares `fail-closed`, `fail-visible`, or `advisory` behavior. Test allow/deny, stale policy, unsupported capability, invalid output, timeout, dependency outage, bypass, concurrent siblings, duplicate/retry, partial application, evidence loss, emergency disable, and rollback.

## Exceptions and break glass

An exception records control, subject/scope, owner, authorized approver, rationale, start/expiry, compensating controls, evidence, affected operations, and review. It cannot exceed the approver's authority or become a silent default.

Break glass additionally requires an emergency reason, explicit invocation, strong authentication, shortest viable duration, authoritative access-system evidence, immediate notification where applicable, containment, automatic expiry/revocation, and post-event review. It does not erase policy, audit, or retained evidence.

Detailed module and policy contracts live in [`module-trust.md`](../../../skill/skill-agent-governance/agent-governance-guide/references/module-trust.md) and [`policy-bundles.md`](../../../skill/skill-agent-governance/agent-governance-guide/references/policy-bundles.md).
