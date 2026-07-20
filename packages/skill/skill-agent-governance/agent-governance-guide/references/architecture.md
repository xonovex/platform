# Governance Architecture and Authority

## Independent trigger, execution, and oversight dimensions

A hook is one possible trigger, not the organizing abstraction. A workflow invocation keeps three dimensions independent:

| Dimension          | Values                                                                                                   |
| ------------------ | -------------------------------------------------------------------------------------------------------- |
| Trigger origin     | manual action, agent-harness hook, CI/CD hook, provider webhook, schedule, sensor, API, or another agent |
| Execution family   | workflow script; workflow script followed by a bounded LLM; bounded agent executing a workflow skill     |
| Oversight maturity | none for deterministic/model execution; observed `A1`, `A2`, or `A3` controls for an agent launch        |

The trigger supplies origin, actor, idempotency, subject, and exact revision. It does not select authority or adaptivity implicitly. The trusted workflow template selects one of these execution families:

```text
any trigger ─┬─> workflow script
             ├─> workflow script ─> bounded LLM
             └─> observed A1/A2/A3 oversight ─> bounded agent ─> workflow skill
```

The runtime contract in `scripts/workflow-runtime.ts` and the native command implementation in `scripts/workflow-command-runtime.ts` enforce that separation. `scripts/workflow-trigger-adapters.ts` binds minimized, authenticated native-event metadata to a trusted template without converting a trigger into an executor.

## Plane boundaries

Governance owns applicability, policy, authority, actor requirements, exceptions, evidence requirements, and effective composition. Execution implements declared capability contracts. Enforcement applies decisions through native adapters. Enablement changes configuration transactionally. Providers own native state and evidence. Observability correlates executions without becoming a workflow identity. Distribution owns package trust and lifecycle.

Dependencies point inward to semantic contracts and ports:

- policy logic consumes contextual facts and emits decisions without importing a concrete hook, CI system, provider, or policy language;
- enforcement adapters consume decisions and native event context without redefining policy;
- executors receive explicit inputs and authority rather than locating ambient capabilities;
- onboarding invokes native configuration adapters only after preview and authorization;
- governance may evaluate workflow facts, but governance-only operation never requires workflow results.

Use **workflow-guide** for lifecycle capability, result, and profile semantics when that optional plane is installed.

## Trust and authority zones

| Zone                 | Typical authority                                                        | Rule                                                              |
| -------------------- | ------------------------------------------------------------------------ | ----------------------------------------------------------------- |
| Organization-managed | Mandatory policies, managed harness configuration, CI/provider controls  | Lower zones cannot silently weaken it                             |
| Repository/project   | Trusted project instructions, hooks, plugins, CI, provider configuration | Executable content requires repository trust and review           |
| User                 | Personal skills, credentials, models, extensions, preferences            | May strengthen; weakening requires authorized exception           |
| Session/runtime      | Ephemeral context, grants, workspace, temporary credentials, budgets     | Expires with runtime and cannot create durable higher authority   |
| External authority   | Identity, secrets, CI, deployment, monitoring, GRC, provider systems     | Evidence is authoritative only for its declared subject and scope |

Native precedence differs by platform. Preserve one semantic invariant: a lower-authority source cannot silently weaken a mandatory higher-authority control. Conflicts, unsupported enforcement, and authority expansion fail visibly.

## Adoption modes

Governance-only, enablement-only, external-enforcement-only, workflow-only, and integrated compositions are independently valid. A governance profile may reference workflow evidence, but a baseline governance module protects ordinary prompt, tool, model, workspace, or privileged activity without a lifecycle command.
