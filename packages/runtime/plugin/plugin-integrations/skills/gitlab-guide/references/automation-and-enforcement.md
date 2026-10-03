# GitLab Automation and External Enforcement

Use GitLab-native controls as an adapter for semantic governance intent. Component YAML, pipeline-policy YAML, framework labels, protected-environment settings, and API resources remain provider details rather than a universal workflow or policy representation.

## Capability declaration

Record offering (GitLab.com, Self-Managed, or Dedicated), edition/tier/version, group/project scope, tested date, component project/ref, policy project and linked scopes, pipeline strategy and conflict behavior, compliance frameworks, protected environments, runner trust, roles/tokens/secrets, evidence resources, and rollback/drift behavior.

| Intent                                   | GitLab mechanism                                                               | Native evidence                                                        |
| ---------------------------------------- | ------------------------------------------------------------------------------ | ---------------------------------------------------------------------- |
| Reusable validation                      | Versioned CI/CD component with `spec:inputs`                                   | Pipeline configuration, component SHA/version, job and artifact/report |
| Mandatory project-independent validation | Group/project pipeline execution policy                                        | Effective policy/project revision, pipeline and policy job             |
| Scope/report governance                  | Compliance framework and framework controls/reporting                          | Framework assignment/control status; not legal certification           |
| Protected target change                  | Protected branch/tag/environment, deployment approvals, provider roles         | Exact commit/artifact, actor/approval, job/deployment and audit event  |
| Supply-chain evidence                    | Pinned component/image/dependency, job artifacts/reports, releases/deployments | Full commit SHA, artifact digest, pipeline/job/deployment references   |

Feature availability and behavior vary by version, offering, tier, and feature flag. Detect them before selecting a mandatory control; a configured but unsupported feature is not enforcement.

## CI/CD component modules

- Keep each component in `templates/<component>.yml` and define its interface through `spec:inputs`. Prefer inputs over ambient custom variables because missing required inputs fail pipeline creation and the interface is explicit.
- Make stage and job name/prefix configurable inputs. Hard-coded generic job names collide when a component is included twice or meets project/policy jobs; a unique stable name is also necessary for reliable evidence and dependencies such as `needs`.
- Include a component at `$CI_SERVER_FQDN/group/project/component@<full-commit-sha>` for the strongest immutable pin. A trusted release tag can carry a human-facing semantic version, but branches, partial versions, and `~latest` move; record the resolved SHA.
- Test the component in its project pipeline before release: include the component from the current commit, assert generated jobs/inputs/outputs/artifacts, exercise invalid/missing inputs and duplicate inclusion, and release only after the component test passes.
- Audit/pin nested components, container images, scripts, package dependencies, runner tags, artifacts/caches, and token access. Do not pass credentials or artifacts into a component unless its declared contract needs them.
- Protect component release branches/tags and review every pin update. A catalog or verified-creator badge is a trust signal, not a substitute for source and permission review.

Minimal component and caller shape:

```yaml
spec:
  inputs:
    job-prefix:
      description: Unique governance job prefix
    stage:
      default: test
---
"$[[ inputs.job-prefix ]]:validate":
  stage: $[[ inputs.stage ]]
  script: ./governance validate
```

```yaml
include:
  - component: $CI_SERVER_FQDN/example/governance/validate@e3262fdd0914fa823210cdb79a8c421e2cef79d8
    inputs:
      job-prefix: organization-governance
```

## Pipeline execution policies and precedence

- Use a pipeline execution policy when governed projects must not be able to omit the job; applicable policies remain effective even when a project has no `.gitlab-ci.yml`.
- Protect the security policy project and its default/release branches, restrict policy edits and linking permissions, and bind evidence to its exact commit.
- Prefer unique names such as `organization-governance:validate`. With `suffix: on_conflict` (default), GitLab adds a provider-defined policy suffix on collision; with `suffix: never`, a collision fails the pipeline. Select and test one deliberately. Never depend on an ambiguous duplicated job name because `needs` can select an instance unpredictably.
- Preserve merge order: project pipeline jobs, project-policy jobs, then group-policy jobs by hierarchy with higher-level group policy applied later. Do not simulate this by concatenating files.
- Policy jobs may use the reserved `.pipeline-policy-pre` and `.pipeline-policy-post` stages, but a pipeline containing only `.pre`/`.post` jobs can evaluate as empty. Exercise the real merged pipeline.
- Declare `inject_policy` versus `override_project_ci` explicitly. The latter replaces the project configuration and has different variable-precedence implications; preview the rendered/merged result and test project variables that attempt to weaken the policy.
- Treat invalid policy config, cyclic stages, name conflicts, missing includes, policy-project outage, cancelled/skipped policy jobs, and unsupported versions with an explicit fail-closed or fail-visible result. Mandatory controls never become success silently.

## Compliance frameworks and protected operations

- Use a compliance framework to classify project scope and attach supported framework controls/reporting. Record its native ID and applied projects. The framework label or report is not proof of certification, legal applicability, or control effectiveness.
- Put release, production deployment, secret rotation, infrastructure mutation, deletion, and retirement behind protected branches/tags/environments plus authorized provider roles and deployment approvals where available.
- Prevent the change author or deployment job from self-authorizing when independence is required. Check tier/version support and preserve the actual approver/actor audit reference.
- Release credentials only to the protected job. Prefer short-lived identity federation where supported; otherwise use a narrowly scoped masked/protected variable or external secret and rotate it.
- Bind approval and deployment to the exact commit and artifact digest. Record pipeline/job/artifact/report/release/deployment references separately, verify the target after mutation, and retain a rollback artifact.

## Permissions, untrusted changes, and evidence

- Use the narrowest project/group access token or job-token allowlist/permissions supported for the operation. Separate read-only validation from target-changing jobs; `api` grants broad authority and should not be ambient in untrusted jobs.
- Never run merge-request-controlled scripts/images/includes with protected variables or a privileged persistent runner. Verify pipeline source, protected-ref state, and component/policy revision before secrets are available.
- Use unique immutable job and artifact identities. Retain the subject commit, component/policy commit, pipeline and job IDs, artifact/report digest, environment/deployment, actor, outcome, and limitations.
- Minimize logs, artifacts, and reports that may contain source, secrets, personal data, prompts, or model content. Declare access and retention at the provider boundary.

## Onboarding transaction

1. Discover includes/components and resolved refs, rendered pipeline behavior, policy projects/links and merge strategies, job names/dependencies, compliance frameworks, protected branches/tags/environments, runners, roles/tokens/variables, bypass paths, evidence, and exact edition/version/tier.
2. Propose a pinned typed/tested component with configurable stable names, a pipeline execution policy for mandatory injection, framework scope where useful, protected privileged operations, minimal credentials, evidence, negative probes, rollback, and drift owner.
3. Preview exact repository and provider mutations, resolved component/policy SHAs, rendered/merged job graph, precedence/conflict behavior, framework assignments, environment/role/variable/data-flow changes, and rollback order.
4. Apply only after authorization. Re-read authoritative state and run compliant, omission, duplicate-name, variable-override, bypass-role, protected-environment, and policy-outage probes.
5. Roll back policy links/config, framework assignment, protected settings, credentials, and component pin to captured versions, then repeat allow/deny probes. Monitor version/tier, pins, policy merge behavior, job names, roles, tokens, runners, and evidence freshness.

For governance-only adoption, include the component and enforce it with a pipeline execution policy; add protected release/deployment controls only when those operations are in scope. Agent-harness hooks and lifecycle modules are not prerequisites, and absent layers must be reported as absent.
