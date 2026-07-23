---
name: azure-devops-guide
description: "Use when operating or onboarding Azure Boards, Repos, Pipelines, service connections, REST APIs, or service hooks on Azure DevOps Services or Server. Triggers on work-item relationships, pull requests and branch policies, reusable pipeline templates, approvals/checks, artifacts, workload identity federation, API-version selection, native evidence, rollback, or edition-aware capability checks — even when the user doesn't say 'Azure DevOps'."
---

# Azure DevOps Platform Operations

Operate Azure DevOps as an optional provider for work items, source and pull requests, CI/CD, protected resources, integrations, and native evidence. Keep Azure DevOps Services and Server capabilities explicit and independently tested.

## Essentials

- **Detect before selecting** — resolve host, Services versus Server, server/API version, organization, project, process, agent pool, installed extensions, and actor before claiming support.
- **Keep native owners separate** — Boards owns work items and links; Repos owns commits, pull requests, policies, and statuses; Pipelines owns runs, stages, checks, artifacts, and protected-resource use.
- **Preserve opaque references** — return provider-native work-item, pull-request, commit, policy, build, artifact, approval, service-connection, and hook identifiers without turning them into a universal result file.
- **Prefer reusable enforcement** — use version-pinned templates, build-validation policies, approvals/checks on protected resources, scoped permissions, and independently resolvable evidence.
- **Federate cloud access** — prefer workload identity federation or another temporary credential path; never create long-lived cloud secrets by default.
- **Transact every setup** — discover, preview exact native changes and authority, authorize, apply idempotently, re-read to verify, retain rollback, and detect drift.
- **Drive the Services CLI directly** — install and connect `az` plus the `azure-devops` extension, then `az repos pr create` / `update` / link / review from the shell, deriving `--org` and `--project` from the git remote (`v3/<org>/<project>/<repo>`), not the configured default.
- **Handle CLI credentials safely** — prefer a short-lived `az login` or workload-identity federation; fall back only to a scoped, keychained PAT read into `AZURE_DEVOPS_EXT_PAT` at call time and never inlined, `echo`ed, or committed, see [references/auth.md](references/auth.md).

## Workflow

1. Run read-only discovery and choose the exact Services or Server capability baseline.
2. Resolve the native subject and immutable revision where the operation supports one.
3. Preview REST/CLI/configuration mutations, permissions, identities, secrets, network/data flows, evidence, failure behavior, verification, rollback, and drift.
4. Require authorization bound to that preview before a write.
5. Apply against the observed version, verify by re-reading native state and probing positive/negative behavior, then return opaque native references.

## Gotchas

- Azure DevOps Services documentation does not prove Azure DevOps Server behavior. Select the matching versioned view and return unsupported for untested features.
- Approvals and checks are owned by protected resources, not merely by pipeline YAML; changing YAML must not silently bypass a mandatory resource check.
- A successful REST response is an apply result, not verification. Re-read the effective policy, resource check, service connection, build, or work-item relationship.
- Branch build validation, reviewer policies, status policies, pipeline checks, and environment/resource approvals are distinct native controls with distinct evidence.
- PATs are not the default for durable automation. Prefer Microsoft Entra identities and workload federation where supported; disclose every fallback credential scope and expiry.
- API versions differ across Services and Server. Never silently downgrade an unsupported field or endpoint to a weaker operation.

## Example

```text
Detected: Azure DevOps Services · organization acme · project mobile · REST 7.2
Requested: protect refs/heads/main and publish build evidence
Preview: add build-validation policy -> pipeline 42; require exact source revision;
         service connection uses workload identity; no client secret created
Verify: policy re-read + compliant PR passes + omitted/spoofed status fails
Evidence: opaque policy, pull-request, build, timeline, artifact, and approval references
Rollback: restore captured policy revision and verify the deny probe still holds
```

## Progressive Disclosure

- Read [references/editions-and-capabilities.md](references/editions-and-capabilities.md) - Load when detecting Services versus Server, selecting an API version, or deciding whether a capability is supported
- Read [references/boards.md](references/boards.md) - Load when reading or changing work items, fields, states, history, queries, or native relationships
- Read [references/repos-and-policies.md](references/repos-and-policies.md) - Load when operating repositories, pull requests, branch policies, build validation, approvals, or native status evidence
- Read [references/pipelines.md](references/pipelines.md) - Load when authoring reusable templates, approvals/checks, protected resources, artifacts, runs, or pipeline evidence
- Read [references/service-connections.md](references/service-connections.md) - Load when onboarding Azure or external cloud access, choosing workload identity, or reviewing service-connection authority
- Read [references/rest-and-service-hooks.md](references/rest-and-service-hooks.md) - Load when calling REST APIs, selecting versions, handling pagination/retries, or configuring service-hook delivery
- Read [references/onboarding.md](references/onboarding.md) - Load when setting up, diagnosing, dry-running, verifying, rolling back, uninstalling, or checking drift
- Read [references/provider-conformance.md](references/provider-conformance.md) - Load when testing edition/version claims, native references, outage behavior, retries, unsupported features, or fresh-context recovery
- Read [references/first-time-setup.md](references/first-time-setup.md) - Load when installing `az`, signing in (`az login`, device code, managed identity, tenant), adding the `azure-devops` extension, setting org/project defaults, or verifying CLI access from a fresh machine
- Read [references/auth.md](references/auth.md) - Load when choosing between `az login` and a PAT, scoping/storing/rotating a PAT, sending it via `AZURE_DEVOPS_EXT_PAT` or Basic auth for raw REST, or verifying a credential before a call
- Read [references/create-pr.md](references/create-pr.md) - Load when opening a PR with `az repos pr create` (push first, derive coordinates from the remote, capture the id/URL, confirm merge status)
- Read [references/update-pr.md](references/update-pr.md) - Load when changing a PR's title, description, or draft state, or refreshing it after new commits
- Read [references/linking.md](references/linking.md) - Load when cross-linking PRs with `!<PRID>` or linking work items with `#<WorkItemId>` and `--work-items`
- Read [references/reviewers-autocomplete.md](references/reviewers-autocomplete.md) - Load when adding reviewers or setting auto-complete and draft on a PR
