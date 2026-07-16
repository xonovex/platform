# Bitbucket Data Center

## Versioned deployment

Detect the exact Data Center version/build, cluster/base URL, project and repository keys, enabled features, installed apps and versions, authentication method, effective permissions, hooks, merge checks, and upgrade state. Select the matching REST reference; do not send Cloud REST 2.0 payloads to Data Center.

## Native operations

Preserve project/repository coordinates, commit ID, pull-request ID and version, participant/reviewer state, comment/task IDs, build/deployment status key, webhook/hook identity, and native URLs. Use the pull-request version or advertised concurrency token for conditional writes and re-read after a conflict.

Repository/project permissions, branch restrictions, merge checks, build status, deployment status, webhooks, and installed-app controls remain separate. Discover global/project/repository layering and every bypass/admin path before claiming mandatory enforcement.

## Installed apps and hooks

Treat each app or server-side hook as executable software. Record publisher/source, exact version, compatibility range, permissions, filesystem/network/secret/data access, side effects, failure behavior, cluster behavior, upgrade, disable, and rollback. A Cloud feature name does not establish an equivalent app contract.

## Upgrade and outage

Pin conformance to the deployed version and app set. Before an upgrade, preview API removals, app compatibility, schema/index changes, clustered rollout, maintenance/outage behavior, rollback support, and evidence continuity. Re-run repository, pull-request, permission, status, webhook, and bypass probes after each node/version transition.

On provider or plugin outage, mandatory merge/deployment policy follows its declared fail-closed behavior or an authorized expiring exception. Advisory integrations fail visibly without inventing a Cloud fallback.
