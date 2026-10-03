# GitHub Automation and External Enforcement

Use GitHub-native controls as an adapter for semantic governance intent. Workflow YAML, ruleset payloads, environment settings, and REST/GraphQL resources remain GitHub implementation details; do not make them a universal workflow or policy representation.

## Capability declaration

Record the GitHub host/product, plan/features, repository/organization scope, tested date, Actions policy, reusable module revision, runner trust, ruleset IDs, environments, required-check names and source applications, bypass actors, token permissions, evidence resources, and rollback/drift behavior.

| Intent                       | GitHub mechanism                                                                                                | Native evidence                                                                      |
| ---------------------------- | --------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------ |
| Reusable validation          | `workflow_call` reusable workflow; composite action for reusable steps inside a job                             | Workflow run, job/check suite and check run bound to head SHA                        |
| Mandatory merge validation   | Active ruleset or branch protection requiring exact status checks and review/conversation rules                 | Effective ruleset/protection, check run source and conclusion, pull request head SHA |
| Protected release/deployment | Environment protection rules, required reviewer/custom protection, deployment branch/tag restrictions           | Environment/deployment, reviewer/protection result, artifact digest                  |
| Least-privilege execution    | Job/workflow `permissions`, named secrets, environment secrets after approval, OIDC for short-lived credentials | Workflow permissions/config revision, OIDC subject/audience, provider audit record   |
| Supply-chain evidence        | Immutable action/workflow refs, artifact attestations, signed release or deployment evidence                    | Full commit SHA, artifact digest, attestation and producing run                      |

Repository feature and plan availability varies. Discover it before proposing a mandatory control; an unavailable or evaluation-only rule is not enforcing.

## Reusable workflow and action modules

- Put reusable workflows directly under `.github/workflows/` and declare `on: workflow_call`. Call cross-repository workflows at `owner/repository/.github/workflows/file.yml@<full-commit-sha>`; a full commit SHA is the stable security pin. A same-repository `./.github/workflows/file.yml` call uses the caller's commit.
- Use typed `workflow_call.inputs` and explicitly named secrets. Avoid `secrets: inherit` unless every inherited secret is required and reviewed. Environment secrets cannot be passed through `workflow_call`; a called job's environment supplies its own secrets after its protection rules pass.
- Nested workflow permissions can only be maintained or reduced, never elevated. Declare minimal `permissions` at the caller and narrow again per job.
- Use a composite action when reusable steps must run inside a caller job; use a reusable workflow when the module owns jobs, runners, permissions, environments, outputs, or an evidence boundary.
- Pin every third-party `uses:` action to a full-length commit SHA and record its release label separately. Verify the SHA belongs to the intended repository, review source and nested dependencies, and update through a reviewed pin change.
- Treat self-hosted runners as a privileged trust zone. Do not run untrusted pull-request code on a persistent privileged runner or with target-repository secrets.

Minimal caller shape:

```yaml
permissions: {}

jobs:
  governance:
    permissions:
      contents: read
    uses: example/governance/.github/workflows/validate.yml@0123456789abcdef0123456789abcdef01234567
    with:
      subject-sha: ${{ github.event.pull_request.head.sha }}
```

The required check must bind to the exact emitted job/check identity. Preserve stable names across module updates, or update and verify the ruleset in the same authorized transaction.

## Rulesets and required checks

- Prefer an active organization/repository ruleset for shared policy where available. Multiple matching rulesets and legacy branch protections layer; rules aggregate and the most restrictive duplicate rule applies rather than a user-controlled priority order.
- Discover every matching ruleset/protection and its bypass list. A repository admin, team, or GitHub App granted bypass is part of the effective control and evidence.
- Require the exact check name and, where supported, its expected GitHub App/source. Verify a compliant head SHA passes and a branch that omits, renames, skips, or spoofs the check cannot merge.
- Decide how skipped, neutral, cancelled, timed-out, stale, and missing checks behave. Mandatory validation must fail closed at the merge gate.
- Pair code-owner/reviewer/conversation requirements with checks when human independence is required. A review is not a substitute for build evidence, and a check is not an accountable approval.
- Test direct pushes, force pushes, merge queue behavior if enabled, administrators/bypass actors, fork-origin pull requests, and API merges.

## Protected environments and privileged operations

Put release, production deployment, secret rotation, infrastructure mutation, data deletion, and retirement jobs behind a protected environment.

- Restrict deployment branches/tags to protected targets and require an authorized reviewer or custom protection rule. Enable prevention of self-review where segregation is required.
- GitHub allows multiple configured required reviewers but only one approval is required for the job to proceed; do not represent the setting as unanimous approval.
- Store environment secrets only when needed after protection passes. Prefer OIDC-issued short-lived provider credentials over long-lived cloud secrets, and constrain the provider trust policy to repository, workflow/ref/environment, audience, and operation.
- Bind authorization to the exact artifact digest and source revision. Revalidate if the revision, artifact, environment policy, approver eligibility, or prerequisite evidence changes.
- Record deployment and post-deployment verification separately. Preserve a verified rollback artifact and target.

## Tokens, untrusted inputs, and evidence

- Start with `permissions: {}` and add only the required scopes. Validation normally needs `contents: read`; attestation generation additionally needs the documented `id-token: write` and `attestations: write` permissions. Target changes get write permission only in the protected job.
- Never interpolate untrusted context directly into shell source. Pass it as data through an environment variable or a reviewed action input and quote it in the receiving language.
- Keep `pull_request_target` or other base-context privileged workflows from checking out and executing untrusted head code.
- Generate attestations for released binaries/images where required, publish artifacts by digest, and retain the workflow run/job/check/deployment references plus policy and module versions. Verify attestations against the intended repository/workflow identity before promotion.
- Logs and artifacts can contain source, secrets, personal data, or model content. Minimize collection and declare retention/access; do not copy an entire log merely to normalize evidence.

## Onboarding transaction

1. Discover existing workflow files, reusable calls and resolved refs, Actions allow/pinning policy, required checks, matching rulesets/protections, environments, bypass actors, runners, permissions/secrets/OIDC, artifact evidence, and feature availability.
2. Propose a full-SHA-pinned reusable module, stable check identity, active ruleset, protected environment for privileged jobs, least permissions, evidence, negative probes, rollback, and drift owner.
3. Preview exact workflow and provider API mutations, including the before/after ruleset and environment resources, bypass changes, secret/OIDC flow, module pin, and expected checks.
4. Apply only after authorization. Re-read each resource and run positive plus bypass probes; a successful API response is not verification.
5. Roll back to the captured workflow/ruleset/environment versions and pin, then repeat allow/deny probes. Monitor changed pins, permissions, bypass actors, rule status, check names/sources, runner trust, and feature availability.

For governance-only adoption, install only the reusable validation caller plus required ruleset/check, and optionally the protected release environment. Agent-harness hooks and lifecycle modules are not prerequisites; record that those layers are absent.
