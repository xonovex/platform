# Architecture and Adoption

## Ownership boundaries

| Plane                       | Owns                                                                                                           | Must not own                                                                         |
| --------------------------- | -------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------ |
| Semantic workflow           | Capability meaning, result contracts, publication boundaries, profile topology, exit and completion evaluation | Provider representation, harness events, executor implementation, policy enforcement |
| Governance and policy       | Applicability, policy decisions, authority, actor requirements, exceptions, evidence requirements              | Lifecycle topology, provider-native mutations, native hook configuration             |
| Execution                   | How a declared capability runs and returns its result                                                          | Redefining capability meaning or granting its own authority                          |
| Enforcement                 | Applying a policy decision at a native control point                                                           | Embedding the only copy of policy meaning or claiming unsupported guarantees         |
| Enablement                  | Discovery, recommendation, preview, authorized setup, verification, rollback, drift                            | Treating installation as enforcement evidence or mutating before authorization       |
| Provider and evidence       | Native reads, writes, references, revisions, relationships, authentication, and platform procedures            | A universal result schema, synthetic workflow identity, or cross-provider inference  |
| Observability and assurance | Correlation, measurements, audit/evidence publication, freshness                                               | Raw-content capture by default or turning a trace ID into workflow identity          |
| Distribution and trust      | Packaging, provenance, versions, compatibility, permissions, upgrades, retirement                              | Ambient runtime authority merely because a package is installed                      |

Dependencies point toward semantic contracts and ports. Native adapters implement those ports; semantic workflow code never imports a concrete provider, harness, policy engine, telemetry backend, or package format. Governance may evaluate workflow facts and opaque references, but neither plane is a prerequisite for the other.

## Adoption modes

- **Workflow-only** — lifecycle capabilities, method skills, provider skills, result contracts, and profiles without governance hooks.
- **Governance-only** — policy, hooks, modules, evidence, and ordinary agent-activity controls without lifecycle commands.
- **Enablement-only** — discovery, diagnostics, guidance, and previews without enforcement.
- **External-enforcement-only** — CI, repository rules, deployment approvals, admission controls, or provider permissions without harness hooks.
- **Integrated** — any dependency-valid composition of the preceding modes.

Each installed module declares whether it is advisory, evidence-producing, enforcing, configuration-changing, or privileged. Presets are recommendations, never mandatory products.

## Independent variation axes

Resolve every axis independently:

1. explicit command or prompt selection;
2. selected workflow profile;
3. repository or project instructions;
4. unambiguous environment detection;
5. axis-specific default;
6. otherwise fail visibly or request selection.

Apply this order separately to workflow profile, method, artifact provider, work-item provider, code-host provider, workspace provider, quality/security/governance policy, and learning policy. An explicitly selected unavailable provider fails; a side-effecting operation never silently falls back to local storage.

## Trust boundary

Workflow results preserve actor or executor origin and native revision. Policy authority, configuration precedence, and executable-module trust are governed by **agent-governance-guide**. Workflow profiles may require governance evidence through opaque references, but they do not prescribe the engine, hook, configuration format, or storage representation that produces it.
