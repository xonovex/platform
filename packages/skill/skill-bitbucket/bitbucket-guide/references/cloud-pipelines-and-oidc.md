# Bitbucket Cloud Pipelines and OIDC

## Pipeline configuration

Detect workspace plan, repository Pipelines state, hosted/self-hosted runners, step sizes/images, deployment environments, secured variables, imported/shared configuration support, and existing status/merge integrations.

Prefer a versioned shared pipeline configuration or pipe with reviewed provenance over copied YAML. Pin imported configuration and third-party pipes to immutable revisions or the strongest supported protected version. Keep untrusted pull-request code separate from secret-bearing or target-changing steps.

Preserve pipeline UUID/build number, repository/commit, step identity, artifact metadata/digest, deployment environment/result, and commit-status/check references. Review artifact access and retention; do not copy logs or secret-bearing artifacts into a central evidence store.

## Deployments and secrets

Preview deployment environment restrictions, variables, permissions, concurrency, manual triggers, target identity, network/data flows, expected evidence, rollback, and partial-state behavior. Secured variable identifiers may appear in a preview; values may not.

## AWS OIDC

1. Read the repository's Bitbucket OIDC issuer, audience, and documented claims.
2. Preview the AWS identity provider and role trust policy with exact issuer, audience, workspace, repository, environment, branch/tag or other supported subject constraints.
3. Preview the role permissions and optional session policy at least privilege.
4. Verify the build receives temporary credentials, record expiry and role-session evidence, prove the intended AWS call succeeds, and prove a different repository/environment/action fails.
5. Roll back the owned trust and role grants, then verify assumption fails while retained pipeline/audit evidence still resolves.

If the required claims cannot be constrained or the tier/runner lacks OIDC, return unsupported.
