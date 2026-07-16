---
name: bitrise-guide
description: "Use when operating or onboarding Bitrise Workflows, Pipelines, Steps, triggers, secrets, artifacts, build statuses, APIs/webhooks, hosted or self-hosted runners, or AWS OIDC. Triggers on bitrise.yml composition, Step version/provenance, pull-request secret exposure, mobile CI artifacts, Git-provider status linkage, workspace plan capabilities, temporary AWS credentials, rollback, drift, or native build evidence — even when the user doesn't say 'Bitrise'."
---

# Bitrise Platform Operations

Operate Bitrise as an optional mobile CI, artifact, deployment-status, and native-evidence provider. Keep source-host and AWS authority outside Bitrise while preserving their opaque references.

## Essentials

- **Discover the execution context** — resolve workspace/app, plan, stack, hosted/self-hosted runner, Git integration, trigger map, secrets policy, and API capability before changing a workflow.
- **Compose native units** — Workflows group Steps; Pipelines orchestrate Workflows; triggers select them. Keep each identity/version and evidence reference explicit.
- **Pin executable Steps** — prefer verified sources and exact reviewed versions; verification is provenance evidence, not permission or compatibility proof.
- **Protect secrets and artifacts** — disclose pull-request exposure, runner trust, variable scope, artifact access/retention, network/data flow, and build-status publication.
- **Federate AWS access** — constrain Bitrise OIDC claims to the intended workspace/app/repository/workflow/environment.
- **Transact setup** — discover, preview exact native changes, authorize, apply idempotently, re-read and probe, roll back, and detect drift.

## Workflow

1. Discover workspace/app capabilities, current configuration, runner boundary, integrations, identities, secrets, artifacts, and status targets.
2. Resolve the exact source commit and version every Step, stack, image, reusable workflow/pipeline input, and external integration.
3. Preview configuration/API changes, permissions, secrets, OIDC claims, network/data/cost effects, expected evidence, failures, verification, rollback, and drift.
4. Require authorization before mutation, apply against the observed config version, then re-read and run positive/negative probes.
5. Return opaque app, workflow/pipeline, build, Step, artifact, deployment/status, webhook, and external evidence references.

## Gotchas

- Workflow, Pipeline, trigger, and Git-provider build status are distinct. A build can run without proving the correct trigger or status protected the exact commit.
- Pull-request secret exposure depends on repository trust and workspace/app settings; untrusted fork code must never receive protected secrets.
- Self-hosted runners expand the workspace trust boundary to their host, network, caches, and credentials. Hosted behavior cannot be assumed.
- A Verified Step still needs exact version selection, source review, permissions/data-flow inspection, and compatibility testing.
- OIDC is safe only when issuer, audience, subject/repository/app/workspace, and deployment context are constrained in AWS.
- Artifact availability is not permanent evidence: record retention/access policy and an immutable content digest where the build provides one.

## Example

```text
Detected: Bitrise workspace acme · app mobile · hosted stack · OIDC available
Preview: pin shared Workflow and Steps; protect production trigger; expose no PR secrets;
         trust OIDC only for this app/workflow/environment; role permits one deploy action
Verify: expected commit/build/status/artifact link resolves; fork PR gets no secret;
        cross-app role assumption and expired token fail
Rollback: restore captured configuration and remove only owned AWS trust/grants
```

## Progressive Disclosure

- Read [references/workflows-pipelines-and-steps.md](references/workflows-pipelines-and-steps.md) - Load when composing Workflows/Pipelines, choosing stacks/runners, pinning Steps, or reviewing execution provenance
- Read [references/secrets-artifacts-and-status.md](references/secrets-artifacts-and-status.md) - Load when handling secrets, pull requests, artifacts, deployments, Git-provider statuses, retention, or access
- Read [references/triggers-api-and-runners.md](references/triggers-api-and-runners.md) - Load when configuring triggers, API/webhooks, duplicate builds, hosted/self-hosted execution, retries, or native references
- Read [references/aws-oidc.md](references/aws-oidc.md) - Load when onboarding Bitrise-to-AWS temporary credentials, authoring trust claims, verifying expiry, or removing access
- Read [references/onboarding.md](references/onboarding.md) - Load when setting up, diagnosing, dry-running, verifying, rolling back, uninstalling, or checking drift
- Read [references/provider-conformance.md](references/provider-conformance.md) - Load when testing plan/runner/Step claims, secret exposure, artifacts/statuses, OIDC, outages, or fresh-context recovery
