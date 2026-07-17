# Sources

## Azure Pipelines templates

- **URL:** https://learn.microsoft.com/en-us/azure/devops/pipelines/process/templates?view=azure-devops
- **Last reviewed:** 2026-07-16
- **Used for:** `references/pipelines.md`, `references/provider-conformance.md`
- **Aspects extracted:** Include/extends templates, parameters, template approval, repository references, and secure reuse. Pinning, preview, evidence, and failure-policy requirements are Xonovex adapter constraints.

## Azure Pipelines approvals and checks

- **URL:** https://learn.microsoft.com/en-us/azure/devops/pipelines/process/approvals?view=azure-devops
- **Last reviewed:** 2026-07-16
- **Used for:** `references/pipelines.md`, `references/onboarding.md`
- **Aspects extracted:** Protected-resource checks, categories/order, approvals, branch control, required templates, external checks, locks, timeouts, and resource-owner configuration.

## Azure Repos branch policies

- **URL:** https://learn.microsoft.com/en-us/azure/devops/repos/git/branch-policies?view=azure-devops
- **Last reviewed:** 2026-07-16
- **Used for:** `references/repos-and-policies.md`, `references/provider-conformance.md`
- **Aspects extracted:** Reviewer, linked-work-item, comment-resolution, build-validation, and status policies. Exact-revision, spoofing, and bypass probes are Xonovex conformance rules.

## Azure Boards work items

- **URL:** https://learn.microsoft.com/en-us/azure/devops/boards/work-items/about-work-items?view=azure-devops
- **Last reviewed:** 2026-07-16
- **Used for:** `references/boards.md`
- **Aspects extracted:** Work-item types, fields, states, history, links, processes, and native identity.

## Azure Resource Manager service connections

- **URL:** https://learn.microsoft.com/en-us/azure/devops/pipelines/library/connect-to-azure?view=azure-devops
- **Last reviewed:** 2026-07-16
- **Used for:** `references/service-connections.md`, `references/onboarding.md`
- **Aspects extracted:** Workload identity federation, managed identities, service-connection scope, pipeline authorization, and credential alternatives.

## Azure DevOps REST API

- **URL:** https://learn.microsoft.com/en-us/rest/api/azure/devops/?view=azure-devops-rest-7.2
- **Last reviewed:** 2026-07-16
- **Used for:** `references/editions-and-capabilities.md`, `references/boards.md`, `references/repos-and-policies.md`, `references/rest-and-service-hooks.md`
- **Aspects extracted:** API-version selection and native resources for work items, repositories, pull requests, policies, builds, artifacts, and service hooks. Services/Server negotiation remains a tested adapter responsibility.

## Azure DevOps service hooks

- **URL:** https://learn.microsoft.com/en-us/azure/devops/service-hooks/overview?view=azure-devops
- **Last reviewed:** 2026-07-16
- **Used for:** `references/rest-and-service-hooks.md`, `references/provider-conformance.md`
- **Aspects extracted:** Publishers, event subscriptions, consumers, filters, delivery history, and troubleshooting.

## Azure DevOps CLI (az repos / az devops / az login / azure-devops extension)

- **URL:** https://learn.microsoft.com/en-us/cli/azure/repos/pr?view=azure-cli-latest
- **Last reviewed:** 2026-07-17
- **Used for:** `references/first-time-setup.md`, `references/auth.md`, `references/create-pr.md`, `references/update-pr.md`, `references/linking.md`, `references/reviewers-autocomplete.md`, and `SKILL.md` Essentials
- **Aspects extracted:** `az login` sign-in modes and the `azure-devops` extension install/defaults; `az repos pr create` / `update` / `show` / `reviewer add`, `--auto-complete`, `--draft`, `--work-items`, and `!<PRID>` / `#<WorkItemId>` linking; `AZURE_DEVOPS_EXT_PAT` and Basic-auth PAT delivery for raw REST, PAT scoping/rotation at `_usersSettings/tokens`, and pre-call HTTP verification. The Services-only extension limitation, remote-derived coordinates, keychain-first storage ladder, and CI stored-secret-vs-OIDC guidance are the Xonovex CLI operational path.

## Refresh Workflow

1. Re-check the matching Services and supported Server documentation views and REST resource-area versions.
2. Re-run edition/version, work-item relationship, branch-policy, pipeline/check, service-connection, and hook probes.
3. Record live results separately from documentation conformance and update **Last reviewed**.
