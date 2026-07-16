# Editions and Capabilities

## Detection

Collect these facts read-only before selecting an operation:

1. Canonical host and collection/organization URL.
2. Azure DevOps Services or Azure DevOps Server, including Server release and update level.
3. Advertised REST resource areas and supported API versions.
4. Organization/collection, project, process model, repository, pipeline, agent pool, protected resource, and installed extensions.
5. Actor descriptor, group membership, effective permissions, license/tier, and relevant organization policies.

Pin discovery evidence to its observation time and native version. If product or version cannot be established, return `unknown`; do not assume Services.

## Baselines

| Product                    | Documentation baseline                                  | Candidate capabilities                                                                      | Required probes                                                                                    |
| -------------------------- | ------------------------------------------------------- | ------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------- |
| Azure DevOps Services      | dated documentation snapshot; REST 7.2 where documented | Boards, Repos, Pipelines, artifacts, policies, approvals/checks, service connections, hooks | organization/tier, endpoint version, resource checks, identity mode, native references             |
| Azure DevOps Server 2022.2 | pinned Server release; REST 7.1 baseline                | Boards, Repos, branch/build policies, pipelines/builds, artifacts, hooks                    | installed update/extensions, endpoint version, Server-specific auth, absent Services-only features |

Documentation conformance is not a live-tenant pass. Record `documentation-conformant`, `live-passed`, `live-failed`, or `not-probed` per capability.

## Capability result

Return product, edition, version, API version, scope, account/tier, support state, source snapshot, live-probe reference, limitations, and native reference kinds. States are `supported`, `unsupported`, `tier-restricted`, `extension-dependent`, `stale`, and `unknown`.

Never infer parity from a similar name. A Services-only check, identity mode, or API field is unsupported on Server until the pinned Server probe passes.
