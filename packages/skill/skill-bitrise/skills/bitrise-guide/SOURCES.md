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

## Refresh Workflow

1. Re-check workspace plan, stack/runner, Workflow/Pipeline/Step, trigger, secret, artifact, status, API/webhook, and OIDC behavior.
2. Re-run hosted/self-hosted, fork-secret, exact-status, artifact-retention, duplicate/outage, OIDC, rollback, and drift probes.
3. Separate documentation conformance from live results and update **Last reviewed**.
