# Sources

## Bitrise Workflows and Pipelines

- **URL:** https://docs.bitrise.io/en/bitrise-ci/workflows-and-pipelines.html
- **Last reviewed:** 2026-07-16
- **Used for:** `references/workflows-pipelines-and-steps.md`, `references/onboarding.md`
- **Aspects extracted:** Workflows, Pipelines, Steps, configuration, and mobile CI composition.

## Bitrise Secrets

- **URL:** https://docs.bitrise.io/en/bitrise-ci/configure-builds/secrets
- **Last reviewed:** 2026-07-16
- **Used for:** `references/secrets-artifacts-and-status.md`, `references/provider-conformance.md`
- **Aspects extracted:** Secret storage, protection, pull-request exposure controls, and external secret considerations.

## Bitrise artifacts and build status

- **URL:** https://docs.bitrise.io/en/bitrise-ci/run-and-analyze-builds/managing-build-files/build-artifacts-online
- **Additional URLs:** https://docs.bitrise.io/en/bitrise-ci/configure-builds/configuring-build-settings/reporting-the-build-status-to-your-git-hosting-provider
- **Last reviewed:** 2026-07-16
- **Used for:** `references/secrets-artifacts-and-status.md`, `references/provider-conformance.md`
- **Aspects extracted:** Native artifact handling/access and Git-host build-status publication.

## Bitrise build triggers

- **URL:** https://docs.bitrise.io/en/bitrise-ci/run-and-analyze-builds/build-triggers/configuring-build-triggers
- **Last reviewed:** 2026-07-16
- **Used for:** `references/triggers-api-and-runners.md`, `references/provider-conformance.md`
- **Aspects extracted:** Push, tag, and pull-request trigger selection and target/source matching.

## Bitrise OIDC for AWS

- **URL:** https://docs.bitrise.io/en/bitrise-platform/integrations/oidc-authentication/oidc-for-aws
- **Last reviewed:** 2026-07-16
- **Used for:** `references/aws-oidc.md`, `references/provider-conformance.md`
- **Aspects extracted:** Bitrise OIDC identity and temporary AWS role access. Claim minimization and no-static-key defaults are Xonovex security constraints.

## Bitrise Verified Steps

- **URL:** https://docs.bitrise.io/en/bitrise-ci/workflows-and-pipelines/developing-your-own-bitrise-step/verified-steps
- **Last reviewed:** 2026-07-16
- **Used for:** `references/workflows-pipelines-and-steps.md`
- **Aspects extracted:** Verified Step distribution expectations; verification remains distinct from version pinning, permissions, and runtime trust.

## Bitrise API (v0.1)

- **URL:** https://api-docs.bitrise.io/ (Bitrise REST API reference)
- **Last reviewed:** 2026-07-17
- **Used for:** `references/auth.md`, `references/builds.md`, `references/first-time-setup.md`
- **Aspects extracted:** the `/v0.1/me`, `/v0.1/apps`, `/v0.1/apps/{app}/builds`, `/v0.1/apps/{app}/builds/{build}`, and `/v0.1/apps/{app}/builds/{build}/log` endpoints; the raw `Authorization: <token>` account-wide personal-access-token header; the `log` response shape (`is_archived`, `log_chunks`, `expiring_raw_log_url`); and the `POST /builds` `hook_info` + `build_params` trigger payload with its HTTP-200 `status:error` no-op on an unknown `workflow_id`.

## Bitrise CLI

- **URL:** https://github.com/bitrise-io/bitrise
- **Last reviewed:** 2026-07-17
- **Used for:** `references/first-time-setup.md`
- **Aspects extracted:** the `bitrise` CLI is the local/build-machine `bitrise.yml` workflow runner (`bitrise run`, `bitrise validate`) and exposes no REST API commands — no personal-access-token auth, app listing, build reads, or log fetching — confirming the API path is curl/token-only.

## Refresh Workflow

1. Re-check workspace plan, stack/runner, Workflow/Pipeline/Step, trigger, secret, artifact, status, API/webhook, and OIDC behavior.
2. Re-run hosted/self-hosted, fork-secret, exact-status, artifact-retention, duplicate/outage, OIDC, rollback, and drift probes.
3. Separate documentation conformance from live results and update **Last reviewed**.
