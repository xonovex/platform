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
- **Authenticate over REST only** — the Bitrise API is token-only (no CLI or `auth login`); mint an account-wide personal access token, keychain it, and send it as a raw `Authorization` header verified by `GET /v0.1/me`.
- **Read and re-trigger builds by commit** — map a commit to its build via the Git-host status API, read the log, distinguish emulator/infra flakiness from a real failure, and re-run with the build's real `workflow_id`.

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
- A build-status `key` label is a display name, not a triggerable `workflow_id`; re-triggering an unknown name returns HTTP 200 with `status:error` and starts no build. Confirm the real id from a prior build first.
- A project's concrete app slug, git-host, and workflow-name mappings belong in that project's own instructions (e.g. AGENTS.md), not in this skill.

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
- Read [references/first-time-setup.md](references/first-time-setup.md) - Load when starting from zero on a new machine or account: learning there is no CLI to install, creating a personal access token, storing it, and verifying with `GET /v0.1/me` before resolving an app slug
- Read [references/auth.md](references/auth.md) - Load when authenticating to the Bitrise API, storing/rotating the account-wide personal access token, fixing a 401, or discovering an app slug
- Read [references/builds.md](references/builds.md) - Load when mapping a commit to its build, reading or downloading a build log, triaging emulator flakiness vs a real failure, inspecting a build's real workflow id, or re-triggering a build or PR check
