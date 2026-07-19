# External Enforcement Adapter Contract

## Contents

[Adapter declaration](#adapter-declaration) · [Enforcement-point mapping](#enforcement-point-mapping) · [Native precedence and bypass](#native-precedence-and-bypass) · [Reusable CI and supply-chain rules](#reusable-ci-and-supply-chain-rules) · [Policy decision services](#policy-decision-services) · [Privileged operations](#privileged-operations) · [AgentPolicy mapping](#xonovex-agentpolicy-and-admission-mapping) · [Local hooks](#local-hooks)

External enforcement applies a semantic policy decision through a native control outside the agent harness. It is an independent enforcement type, not a portable configuration format and not proof that another enforcement point ran.

## Adapter declaration

Declare each adapter with:

```text
adapter identity, version, owner, support state, and tested platform/edition/date
semantic intents and native enforcement point for each intent
subject and immutable revision binding
native configuration, precedence, merge, bypass, and protected-target semantics
required actor, provider role, token scopes, secrets, network, and data access
decision input and policy/profile version
allow, deny, ask/approval, evidence-only, and unsupported mappings
fail-closed, fail-visible, or advisory behavior for timeout, outage, cancellation, retry, and partial execution
idempotency, concurrency, naming-conflict, duplicate, and reentrancy behavior
enforcement and evidence references, freshness, retention, redaction, and access
preview, authorization, apply, verify, rollback, drift, upgrade, and removal capabilities
limitations and controls that require another independent enforcement point
```

The semantic adapter consumes policy intent and explicit native context. It never imports a provider YAML/JSON schema into governance policy, treats a successful configuration write as verification, or turns a CI run identifier into workflow-result identity.

## Enforcement-point mapping

| Semantic intent               | Suitable native points                                             | Required evidence                                                                         |
| ----------------------------- | ------------------------------------------------------------------ | ----------------------------------------------------------------------------------------- |
| Validate an exact change      | Required CI job/check, merge gate, pipeline policy                 | Commit/revision, check/job/pipeline, policy and module version, conclusion                |
| Protect a repository target   | Ruleset, branch/tag protection, approval rule, provider permission | Target/ref, effective layered rules, bypass actors, authorization and mutation result     |
| Authorize a deployment        | Protected environment, deployment approval, deployment policy      | Artifact digest, environment, approver/actor, policy version, deployment result           |
| Admit a runtime workload      | Admission webhook, validating policy, namespace policy             | Object UID/generation, admitted image/toolchain/runtime/network/resource facts, verdict   |
| Restrict a provider operation | Provider role/token scope, protected API, secret policy            | Actor/token identity class, granted scope, target, operation and provider audit reference |
| Obtain a policy decision      | Deterministic rules provider or policy decision service            | Decision reference, input digest, policy/data version, expiry and explanation             |
| Assist locally                | Tracked Git hook or pre-commit framework                           | Advisory result only; never mandatory-control evidence                                    |

Mandatory intent is conformant only when at least one selected point cannot be bypassed by the governed actor, binds to the exact subject revision, fails with the declared behavior, and produces independently resolvable evidence. Harness hooks may add an earlier layer but cannot substitute for that external guarantee.

## Native precedence and bypass

- Discover every applicable organization, group, repository/project, environment, namespace, and external-authority source before claiming an effective result.
- Preserve native layering. GitHub rulesets aggregate and the most restrictive duplicate rule applies; GitLab project and group pipeline-policy jobs merge in provider-defined order and job-name handling changes the result. Provider skills own the detailed mechanics.
- Record every bypass actor, application, role, exception, and emergency-exception path. A control with an undisclosed bypass path is unsupported for mandatory use.
- Use unique stable job/check names. Verify that the merge rule requires the exact emitted name and expected application/provider; a similarly named user job is not equivalent evidence.
- Treat skipped, neutral, cancelled, timed-out, missing, stale, or duplicate checks according to an explicit policy. Never coerce them to success.
- Keep policy decision, native enforcement action, target mutation, and evidence publication as separate records.

## Reusable CI and supply-chain rules

- Prefer provider-native reusable workflows, actions, or components to copied pipeline bodies.
- Pin executable third-party modules to an immutable commit digest where supported; record the human-readable release separately. Moving branches, floating tags, partial versions, and `latest` are not immutable pins.
- Verify source, publisher, requested permissions, nested dependencies, runner/image provenance, network/data flows, and the resolved digest before activation and upgrade.
- Start tokens at no permissions and add only the scopes used by each job. Do not pass all secrets to nested modules when named secrets suffice.
- Separate untrusted-change validation from secret-bearing or target-changing jobs. Never execute pull/merge-request-controlled code with privileged credentials.
- Publish evidence by native reference. Bind attestations, artifacts, deployments, and reports to their digest and producing revision; minimize copied logs and sensitive content.

## Policy decision services

A remote service is an optional policy-provider adapter; OPA is one implementation, not a required architecture.

- Send canonical, minimized facts with subject revision, actor, action, authority, profile/policy version, and evidence freshness. Do not send raw prompts, source, secrets, or personal data without a declared need.
- Record the returned decision reference, policy/data bundle version, input digest, outcome, explanation, expiry, and service identity. Preserve enough historical material to replay the original version.
- Cache only signed or otherwise authenticated decisions with an explicit subject, policy version, expiry, and cache key. A cache miss never becomes allow.
- For mandatory privileged operations, outage, invalid response, expired cache, version mismatch, or replay failure is fail-closed unless an authorized, scoped, expiring exception path, including an emergency one, says otherwise.
- Advisory checks may fail-visible and continue, but must state that no enforcement occurred.
- Prevent thundering-herd retries with bounded backoff and cancellation. Reconcile duplicate decisions and evidence idempotently.

### Xonovex decision-service deployment topology

The operator distribution runs the decision service as a sidecar so admission-to-decision traffic stays on loopback and shares the Pod's lifecycle. A separate Deployment remains supported by the same HTTP contract when independent scaling or ownership is required. In that topology, expose only the decision port through a ClusterIP Service, apply default-deny ingress, and allow port `8787` solely from operator-manager Pods. The example in `config/manager/decision-service-network-policy.example.yaml` is intentionally not part of the sidecar kustomization; copy and specialize its namespace and labels only when selecting the separate-Deployment topology.

## Privileged operations

Integration, Release, production deployment, secret access/rotation, infrastructure mutation, data deletion, and Retirement require a protected external point with:

- exact source and target revisions or resource identities;
- authorized actor and segregation-of-duties requirements;
- fresh prerequisite assessment/acceptance/evidence references;
- least-privilege short-lived credentials released only after approval;
- protected environment, provider permission, admission, or equivalent target-side gate;
- immutable artifact/digest and provenance verification where an artifact is executed or deployed;
- explicit rollback or irreversibility handling; and
- retained decision, enforcement, mutation, and post-operation verification evidence.

Client-side hooks, labels, issue comments, chat approval, an installed skill, or a passing harness hook alone do not authorize a privileged target change.

## Xonovex AgentPolicy and admission mapping

`AgentPolicy` is a namespace-scoped Kubernetes admission adapter for `AgentRun`; `AgentToolchain` admission is a separate enforcing resource.

| Semantic intent               | Native field/control                               | Scope and limitation                                                                                                                         |
| ----------------------------- | -------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------- |
| Kernel-isolated runtime       | `runtimeClassName`, `allowedRuntimeClassNames`     | Admission verifies the requested runtime class; cluster RuntimeClass/runtime implementation remains an external dependency                   |
| Harden container execution    | `requireSecurityContext`                           | Rejects explicit privilege escalation or root weakening; pod builder supplies hardened defaults                                              |
| Restrict egress               | `requireNetworkPolicy`, `network`, `networkPolicy` | Requires the operator NetworkPolicy path; cluster network plugin behavior must be verified independently                                     |
| Bound duration                | `maxTimeout`                                       | Rejects a missing or excessive effective admission value                                                                                     |
| Bound CPU/memory              | `maxResources`                                     | Requires and caps configured request/limit quantities for named resources; namespace LimitRange/ResourceQuota provide an independent control |
| Restrict images               | `allowedImages`                                    | Requires an allowed explicit/default image prefix; digest/provenance policy needs an additional admission or registry control                |
| Pin toolchains                | `AgentToolchain`/inline Nix validation             | Requires a pinned nixpkgs revision, one source, and a digest-pinned pre-built image                                                          |
| Establish namespace authority | One `AgentPolicy` per namespace                    | Missing policy means baseline operator hardening only; lookup failure or multiple policies fail closed                                       |

Admission evidence identifies the Kubernetes object UID/generation and policy resource version. It does not prove the runtime plugin, network plugin, registry, or later deployment behaved correctly; retain their native evidence as independent controls.

## Local hooks

Share local checks through a tracked hooks directory selected by `core.hooksPath`, or through a pinned pre-commit framework configuration. Hooks should be deterministic, fast, idempotent, offline where practical, and expose a direct command that CI runs independently.

Treat every client hook as advisory: a user can skip it, replace `core.hooksPath`, use another Git client, or push through an API. Never place the only mandatory check, secret-bearing operation, approval, or evidence publication in a client-side hook. The corresponding required CI/provider control is the authority.
