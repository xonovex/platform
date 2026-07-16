# External Enforcement Onboarding

Use provider-native configuration adapters transactionally. Governance-only and external-enforcement-only adoption are valid: neither agent-harness hooks nor lifecycle workflow modules are prerequisites.

## Transaction

1. **Discover** — read the host/product, offering/edition/version, organization/group/repository/project, default and protected targets, existing workflows/includes, rulesets/protections, environments, pipeline/security policies, compliance frameworks, admission resources, provider roles, hooks, secrets, and evidence sinks.
2. **Assess** — map requested semantic intents to tested native capabilities; identify overlapping controls, bypass actors, naming/merge conflicts, unsupported features, missing exact-revision binding, credentials/data flows, and current failure behavior.
3. **Propose** — select reusable native modules and policy controls with immutable versions, least-privilege permissions, unique job/check names, expected evidence, verification probes, rollback targets, and drift ownership.
4. **Preview** — show exact native mutations and before/after state. Include every created/updated/deleted resource, resolved module digest, token/role/secret/network change, target selector, bypass path, and irreversibility.
5. **Authorize** — obtain the required repository, organization, group, environment, cluster, or external-system approval. Preview is never implicit consent.
6. **Apply** — write idempotently against the observed version. Stop and re-preview on drift or a partial apply.
7. **Verify** — re-read authoritative configuration and run positive and negative probes: compliant revision passes, bypass attempt fails, outage follows the declared failure policy, evidence resolves, and rollback is still available.
8. **Record** — retain preview, authorization, apply, verification, policy/module versions, native references, limitations, and rollback target.
9. **Operate** — detect configuration/source/platform drift, review bypass use, rotate credentials, update pins explicitly, re-run conformance, and disable/remove without erasing retained evidence.

## Provider-native proposals

For GitHub, propose a SHA-pinned reusable workflow or composite action, required check/ruleset, protected environment for target-changing jobs, minimal `permissions`, named secrets or OIDC, and artifact/evidence publication. Preserve aggregated ruleset behavior and disclose bypass actors.

For GitLab, propose a SHA- or trusted release-pinned CI/CD component with typed inputs and configurable job names, then use a pipeline execution policy when the project must not be able to omit the control. Preserve project/group policy merge order, job-conflict strategy, protected environment behavior, tier/edition requirements, and compliance-framework scope.

For Kubernetes, propose one namespace `AgentPolicy`, validated `AgentToolchain` resources, LimitRange/ResourceQuota, NetworkPolicy capability, RuntimeClass availability, registry/admission provenance controls where required, and an explicit policy lookup/outage posture.

## Preview record

```text
subject/scope and observed native version
requested semantic intents and selected native enforcement points
existing controls, conflicts, bypass actors, and unsupported requirements
resources and exact before/after mutations
module source, release label, immutable digest, provenance verification
actor/role/token/secret/network/data changes
failure, concurrency, retry, duplicate, naming, and partial-apply behavior
expected decision, enforcement, artifact, deployment, and audit evidence
positive/negative verification probes
rollback/disable/remove steps and retained evidence
drift owner, cadence, and refresh trigger
```

## Governance-only recipes

- **Pull/merge validation only** — reusable validation module plus an unbypassable required check/pipeline policy for the exact revision.
- **Protected release** — validation plus protected target, independent approval, short-lived release credentials, signed/attested artifact evidence, and provider-native release/deployment records.
- **Deployment gate** — immutable artifact digest, protected environment, approval/segregation rules, deployment policy, post-deploy verification, and rollback.
- **Kubernetes agent admission** — namespace AgentPolicy plus toolchain admission, RuntimeClass/network/resource namespace controls, and negative admission tests.
- **Advisory local checks** — tracked hook/pre-commit command for fast feedback, with the same command enforced independently in CI.

Each recipe may be adopted without installing lifecycle commands or harness modules. Record that absent layers are absent; never imply integrated coverage.

## Rollback and drift

Rollback restores the captured native configuration version and module pin, then re-runs both allow and deny probes. If a platform cannot roll a change back atomically, preview the ordered compensating mutations and partial-state risk before authorization.

Drift includes changed provider edition/version, disabled rules, new bypass actors, renamed checks, changed policy merge order, moving/unresolvable module refs, widened token scopes, secret/data-flow changes, stale evidence, admission webhook unavailability, and namespace policy duplication. Mandatory enforcement fails visibly until the effective guarantee is re-verified.
