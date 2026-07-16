# Sources

## Bitbucket Cloud Pipelines and shared configuration

- **URL:** https://support.atlassian.com/bitbucket-cloud/docs/bitbucket-pipelines-configuration-reference/
- **Additional URLs:** https://support.atlassian.com/bitbucket-cloud/docs/share-pipelines-configurations/
- **Last reviewed:** 2026-07-16
- **Used for:** `references/editions-and-capabilities.md`, `references/cloud-pipelines-and-oidc.md`
- **Aspects extracted:** Pipeline configuration, steps, runners, artifacts, deployments, secured variables, and shared configurations. Immutable selection and conformance rules are Xonovex constraints.

## Bitbucket Cloud OIDC and deployments

- **URL:** https://support.atlassian.com/bitbucket-cloud/docs/integrate-pipelines-with-resource-servers-using-oidc/
- **Additional URLs:** https://support.atlassian.com/bitbucket-cloud/docs/set-up-and-monitor-deployments/
- **Last reviewed:** 2026-07-16
- **Used for:** `references/cloud-pipelines-and-oidc.md`, `references/onboarding.md`
- **Aspects extracted:** Pipeline OIDC identity/token claims and deployment environment/status behavior. Trust-policy minimization and no-static-key defaults are Xonovex security constraints.

## Bitbucket Cloud merge and branch controls

- **URL:** https://support.atlassian.com/bitbucket-cloud/docs/suggest-or-require-checks-before-a-merge/
- **Additional URLs:** https://support.atlassian.com/bitbucket-cloud/docs/use-branch-permissions/ · https://support.atlassian.com/bitbucket-cloud/docs/set-up-and-use-custom-merge-checks/
- **Last reviewed:** 2026-07-16
- **Used for:** `references/cloud-source-and-controls.md`, `references/provider-conformance.md`
- **Aspects extracted:** Merge checks, branch permissions/restrictions, custom checks, and plan-dependent enforcement.

## Bitbucket Cloud REST and webhooks

- **URL:** https://developer.atlassian.com/cloud/bitbucket/rest/intro/
- **Additional URLs:** https://support.atlassian.com/bitbucket-cloud/docs/manage-webhooks/
- **Last reviewed:** 2026-07-16
- **Used for:** `references/cloud-source-and-controls.md`, `references/provider-conformance.md`
- **Aspects extracted:** Cloud REST 2.0 resources, pagination/auth/rate behavior, webhook subscriptions and deliveries.

## Bitbucket Data Center

- **URL:** https://confluence.atlassian.com/bitbucketserver
- **Additional URLs:** https://developer.atlassian.com/server/bitbucket/rest/v1003/
- **Last reviewed:** 2026-07-16
- **Used for:** `references/editions-and-capabilities.md`, `references/data-center.md`, `references/provider-conformance.md`
- **Aspects extracted:** Data Center 10.3 documentation and REST v1003 resources for projects, repositories, pull requests, permissions, build/deployment status, hooks, administration, and installed-app boundaries.

## Refresh Workflow

1. Re-check Cloud plan/runner/check/Pipelines behavior and the supported Data Center release/API/app matrix independently.
2. Re-run Cloud and Data Center native-reference, exact-status, bypass, webhook, outage, OIDC, upgrade, rollback, and drift probes.
3. Keep documentation conformance distinct from live results and update **Last reviewed**.
