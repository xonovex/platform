---
name: bitbucket-guide
description: "Use when operating or onboarding Bitbucket Cloud or Bitbucket Data Center repositories, pull requests, pipelines, deployments, checks, permissions, OIDC, REST APIs, webhooks, or native build evidence. Triggers on Cloud/Data Center capability differences, shared pipeline configuration, merge/custom checks, branch permissions, build/deployment status, installed apps, provider-native references, rollback, or drift — even when the user doesn't say 'Bitbucket'."
---

# Bitbucket Platform Operations

Operate Bitbucket as an optional source, pull-request, CI/deployment, and native-evidence provider while keeping Cloud and Data Center as separate products.

## Essentials

- **Detect the product first** — resolve Cloud versus Data Center, deployment/version, workspace/project/repository, plan, runner/apps, API surface, and effective actor before selecting a capability.
- **Never infer Cloud parity** — Cloud Pipelines, OIDC, shared configurations, deployments, custom checks, and Cloud scopes do not exist on Data Center merely because both products host repositories.
- **Bind exact revisions** — preserve repository UUID/project key, commit hash, pull-request ID/version, pipeline/build/status, artifact, deployment, check, and webhook-delivery references.
- **Use native gates** — combine branch permissions, merge checks or server-side policy, exact-revision build status, protected deployments, and least-privilege identities according to the detected product.
- **Federate where supported** — constrain Cloud OIDC issuer/audience/subject claims for temporary AWS access and never generate static keys by default.
- **Transact configuration** — discover, preview native mutations and authority, authorize, apply idempotently, re-read and probe, roll back, and monitor drift.

## Workflow

1. Discover the canonical host/product/version and select its pinned capability baseline.
2. Resolve all applicable workspace/project/repository controls, installed apps, runners, credentials, and native evidence sinks.
3. Preview exact API/configuration changes, plan/tier dependencies, permissions, claims/secrets, network/data flows, failure behavior, verification, rollback, and drift.
4. Require authorization for that preview before any write.
5. Apply against observed versions and verify authoritative configuration plus positive, bypass, outage, duplicate, and reference-resolution probes.

## Gotchas

- Bitbucket Cloud and Bitbucket Data Center have different APIs, auth models, feature delivery, and versioning. Route by detected product, never hostname guess alone.
- Bitbucket Data Center capabilities depend on the deployed version and installed apps. Missing Cloud features remain unsupported; do not emulate them with hidden sidecars.
- Cloud merge checks, custom merge checks, branch permissions, pipeline status, and deployment permissions are distinct controls with different plan/tier behavior.
- Webhooks can retry, duplicate, and arrive out of order. Reconcile by delivery/event/native object identity and re-read authoritative state.
- A passing status with the wrong key/name, app/provider, repository, or commit is not equivalent evidence.
- OIDC removes stored cloud keys only when the AWS trust policy constrains the actual Bitbucket claims; wildcard subjects are not least privilege.

## Example

```text
Detected: Bitbucket Cloud · workspace acme · repository mobile · REST 2.0
Preview: import versioned shared pipeline; enable exact-commit merge check;
         protect production deployment; authorize OIDC only for repository/environment claims
Verify: compliant commit passes; missing/spoofed status and wildcard OIDC subject fail
Evidence: opaque pull-request, pipeline, artifact, deployment, check, and webhook references
Rollback: restore captured branch/deployment configuration and remove only owned OIDC trust
```

## Progressive Disclosure

- Read [references/editions-and-capabilities.md](references/editions-and-capabilities.md) - Load when detecting Cloud versus Data Center, account/tier, API version, runners/apps, or supported capabilities
- Read [references/cloud-source-and-controls.md](references/cloud-source-and-controls.md) - Load when operating Cloud repositories, pull requests, branch permissions, merge/custom checks, REST, webhooks, or exact-revision statuses
- Read [references/cloud-pipelines-and-oidc.md](references/cloud-pipelines-and-oidc.md) - Load when configuring Cloud Pipelines, shared configurations, steps, artifacts, deployments, secrets, runners, or AWS OIDC
- Read [references/data-center.md](references/data-center.md) - Load when operating a Data Center deployment, versioned REST, repositories/pull requests, permissions, statuses, webhooks, or installed apps
- Read [references/onboarding.md](references/onboarding.md) - Load when setting up, diagnosing, dry-running, verifying, rolling back, uninstalling, or checking drift
- Read [references/provider-conformance.md](references/provider-conformance.md) - Load when testing edition/version claims, tier restrictions, native references, retries/outages, unsupported features, or fresh-context recovery
