# Architecture and Composition

Release baseline: **6.0.0**. This guide describes semantic contracts, not one harness schema, provider payload, policy language, or storage format.

## Two independent planes

| Plane      | Owns                                                                                                                                                                            | Does not require                                                     |
| ---------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------- |
| Workflow   | Lifecycle capabilities, exact-revision results, profiles, completion, evidence requirements, and provider ports from discovery through retirement                               | Harness hooks, governance modules, one methodology, or YAML sidecars |
| Governance | Applicability, policy decisions, authority, executor bounds, semantic event intents, native enforcement adapters, onboarding, evidence, trust, telemetry, exceptions, and drift | Lifecycle commands or workflow result storage                        |

The planes integrate through semantic references. Governance may require fresh workflow evidence for a selected action; workflow may consume a policy decision or external authorization. Neither plane owns the other's native identity, and neither turns telemetry into a workflow result.

## Adoption modes

| Mode                      | Composition                                                                                                | Expected absence report                                                       |
| ------------------------- | ---------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------- |
| Workflow-only             | Workflow commands and the `workflow-guide`, with selected method and result providers                      | No harness policy or external enforcement guarantee is claimed                |
| Governance-only           | `agent-governance-guide` plus selected domain, harness, policy, evidence, or provider owners               | Lifecycle results are not prerequisites                                       |
| Enablement-only           | Native onboarding adapters used to preview, authorize, apply, verify, roll back, and monitor configuration | No ongoing policy guarantee unless an enforcing module is separately selected |
| External-enforcement-only | Repository, CI, deployment, admission, identity, cloud, or provider controls                               | Agent hooks and lifecycle commands are optional                               |
| Integrated                | Independently selected workflow and governance profiles joined by exact semantic references                | Unselected modules remain explicit gaps, not hidden defaults                  |

Presets are named selections, not indivisible bundles. Users can choose one module or a preset and inspect the resolved methods, executors, providers, enforcement points, permissions, data flows, versions, conflicts, evidence, and limitations before applying changes.

## Effective profile composition

1. Resolve applicable workflow and governance profiles by semantic identity and authority zone.
2. Union strengthening requirements and fail visibly on incompatible actors, data rules, executors, providers, modules, or enforcement guarantees.
3. Select the least-adaptive executor and provider that satisfy the declared contract.
4. Bind every mandatory decision to an adequate native enforcement point and separate evidence reference.
5. Preview the effective composition, requested authority, missing or experimental capabilities, and rollback target.
6. Re-evaluate when a subject revision, profile, policy, module, provider, capability matrix, evidence source, or authority grant changes.

Organization-managed requirements cannot be silently weakened by project, user, or session configuration. A lower-authority composition may strengthen a requirement or present an authorized, scoped, expiring exception.

## Execution selection

| Executor                  | Select when                                                                                        | Required boundary                                                                                                                                               |
| ------------------------- | -------------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Deterministic script/API  | Authoritative inspection, transformation, validation, or privileged mutation is reproducible       | Pin inputs and implementation; declare side effects, idempotency, timeout, exit behavior, and native evidence                                                   |
| Script plus bounded model | A deterministic coordinator needs fixed-scope classification or synthesis                          | Pin context and provider policy; validate a closed schema; cap retries, time, tokens, and cost; label inference evidence                                        |
| Bounded agent             | Adaptive investigation genuinely requires branching tool use                                       | Explicit launch, attenuated tools/files/network/secrets, maximum recursion depth, budgets, cancellation, kill behavior, result validation, and no hidden launch |
| Human                     | Accountable judgment, approval, legal/domain review, or manual action is required                  | Record identity, role, independence, scope, decision, expiry, and authoritative evidence                                                                        |
| External system           | CI, repository, deployment, identity, monitoring, GRC, or another provider owns the action or fact | Preserve native subject, revision, policy/version, outcome, freshness, permissions, and evidence                                                                |

Deterministic or authoritative external evidence takes precedence for facts it can establish. Models and agents may interpret evidence or investigate gaps; they cannot silently replace the authoritative check or create their own authority.

The identity, role, and independence the Human row records are defined in [`actors.md`](../../../skill/skill-agent-governance/agent-governance-guide/references/actors.md), which states which of them code enforces and which a profile or provider must uphold itself. A recorded role is an audit label; no capability selects behavior from it.

## Provider-native results and evidence

Every result preserves a semantic type plus an opaque provider reference and revision/freshness token. Local files, Git, hosted work items, databases, CI runs, artifacts, deployments, monitoring systems, and human records are peer providers selected by profile and context. There is no universal file fallback.

Policy decision, enforcement action, configuration mutation, workflow result, authorization, and evidence are distinct records. Correlation links them without collapsing their ownership.

Detailed contracts live in the [`workflow-guide`](../../../skill/skill-workflow/workflow-guide/SKILL.md) and [`agent-governance-guide`](../../../skill/skill-agent-governance/agent-governance-guide/SKILL.md).
