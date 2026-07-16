# Bitrise-to-AWS OIDC

## Trust preview

Read the actual Bitrise OIDC configuration and sample claims for the selected workspace/app/workflow/environment. Preview:

- exact issuer URL and thumbprint/identity-provider behavior where applicable;
- audience;
- subject and supported workspace/app/repository/workflow/branch/tag/environment claims;
- AWS account, identity provider, role, trust conditions, permissions boundary, role policy, maximum session duration, and session tags;
- Bitrise Workflow/Step changes, network path, data accessed, expected CloudTrail/build evidence, verification, rollback, and drift.

Reject wildcard or unverified claims that let another workspace, app, repository, workflow, or environment assume the role.

## Verification

1. Authorize the preview and apply the identity provider/trust and Bitrise configuration idempotently.
2. Run the intended build and confirm temporary role credentials and expiry without logging the token.
3. Prove the intended AWS action succeeds and a different app/environment/action fails.
4. Resolve Bitrise build/Step and AWS STS/CloudTrail references and bind them to the source commit.
5. Expire/revoke the trust and confirm assumption fails.
6. Roll back only owned role/provider/configuration entries and preserve retained evidence.

Drift includes issuer/audience/subject changes, broadened trust, role-policy/permissions-boundary changes, session-duration changes, Bitrise Workflow/Step changes, and new secret-based credential paths.
