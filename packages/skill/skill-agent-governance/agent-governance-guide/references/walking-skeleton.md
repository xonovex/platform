# Governance Onboarding Walking Skeleton

Use the walking skeleton to prove that governance, workflow results, harness adaptation, external enforcement, and provider evidence compose without collapsing into one platform or persisted envelope. The skeleton ships as three artifacts with distinct evidence value:

- **Executable local run** (`assets/walking-skeleton/run-skeleton.sh`) — a self-contained script that executes the onboarding lifecycle against a temporary workspace and exits nonzero on any failed check. What it demonstrates was actually run.
- **Live harness run** (`assets/walking-skeleton/run-live-lifecycle.sh`) — a maintainer-run probe that invokes the real Claude Code command harness, the installed workflow command definitions, and the decision service. It records separate decision and enforcement receipts for each gate, proves self-approval cannot advance, and makes a real agent `Write` attempt whose protected mutation must be blocked by the native hook.
- **Recorded conformance scenario** (the fixture replay validated by `scripts/validate-walking-skeleton-fixtures.mjs`) — a cross-checked composition over the governance, harness, external-enforcement, enterprise-platform, and module-template fixtures. It proves the contracts compose; it is not a claim that a live tenant or hosted platform passed.

## Executable local run

`run-skeleton.sh` mutates only a `mktemp -d` workspace and removes it on exit; the repository is never touched. Without arguments it discovers the environment, previews the exact change (module identity, checksum, permissions, failure mode, rollback), and stops — consent is a real gate, not documentation. With `--yes` it continues through apply, verification, and rollback, executing 17 checks:

- idempotent apply (re-application converges to the same reference);
- a permitted operation allowed and a protected operation denied with a reason naming the policy and remediation (`guard.sh`, exit 0/2 semantics);
- an independent CI-shaped second enforcement point that still denies with the hook disabled;
- weakening drift detected against the applied reference and remediated;
- a tampered module refused by checksum, an unsupported hook intent rejected, mandatory fail-closed and advisory observe under policy outage, concurrent duplicate evidence deduplicated, a recursive agent launch refused at the depth limit, and an expired exception denying while citing its exception id;
- rollback that removes the applied configuration and leaves drift clean.

That second enforcement point is a locally re-invoked gate shaped like a required CI check; hosted-platform enforcement remains fixture-recorded (below). `guard.sh` is the deterministic policy decision point: JSON event on stdin, one JSON decision on stdout, exit 0 allow / 2 deny.

## Live harness run

The live counterpart is intentionally excluded from CI because it uses the installed provider runtime and credentials. It copies the repository's discovery, acceptance-validation, and integration-validation command definitions into a temporary Claude Code project, reads the recorded Phase 2 native-capability probe, and starts the Phase 1 decision service. The harness outputs are observations, not fabricated human acceptance or external mutations. Both high-impact gates carry the exact Git revision in their decision and workflow-script enforcement evidence. The negative path uses the same author and decider and must produce `acceptance-independence-failed`; it then invokes the real agent against a protected path and requires both file absence and a content-addressed `PreToolUse` enforcement record.

Run it from the repository root with an accountable local identity and optionally retain sanitized probe output:

```sh
LIVE_PROBE_DECIDER='human:maintainer-id' \
  LIVE_PROBE_OUTPUT_DIR=/tmp/xonovex-live-lifecycle \
  npx moon run skill-agent-governance:test-live
```

## Selected coordinates

| Concern                | Fixture selection                                              | Purpose                                                                                                   |
| ---------------------- | -------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------- |
| Harness                | Claude Code versioned capability matrix                        | Mature deterministic command hooks, bounded prompt evaluator, native diagnostics, and compaction recovery |
| Hosted CI              | GitHub required-check and ruleset fixture                      | Independent exact-revision enforcement when the harness hook is disabled                                  |
| Runtime control        | Kubernetes `AgentPolicy` fixture                               | Existing fail-closed admission control discovered without making it a harness prerequisite                |
| Enterprise composition | Azure DevOps, Bitbucket, Bitrise, AWS, and Datadog mixed stack | Provider-native results, Bitrise-to-AWS federation, and minimized telemetry                               |
| Evidence               | Self-controlled database/API provider                          | Opaque references and fresh-process reconstruction without YAML sidecars                                  |

The recorded runtime version is fixture evidence. Re-run native version and diagnostic probes before applying the composition to a real environment.

## Replay sequence

1. Resolve every selection against its versioned source fixture.
2. Discover the harness surface and version, configuration scopes, supported native capabilities, installed modules, GitHub controls, operator policy, provider capabilities, and selected profile facets.
3. Exercise governance-only around an ordinary agent task, workflow-only through enterprise provider references, and their integrated composition.
4. Recommend the smallest composition: context injection, protected-path and tool policy, post-change validation, minimized audit evidence, and a required GitHub check.
5. Preview exact native subjects and digests, permissions, data destinations, bounded model use, failure behavior, verification, and rollback.
6. Bind authorization to the preview digest, apply idempotently, re-read native state, and run the harness diagnostics.
7. Record one permitted tool operation and one explained denial with separate policy-decision and hook-enforcement references.
8. Run the bounded evaluator as advisory inference with a closed result schema and explicit time, token, cost, retry, tool, filesystem, network, and secret scopes.
9. Publish evidence to the non-file provider and reconstruct the subject and revision in a fresh process from opaque references only.
10. Disable the harness hook and prove that the GitHub required check still denies the target mutation.
11. Detect weakening drift, authorize remediation, restore the intended digest, roll back every installed module, and retain evidence.

## Failure probes

The replay derives the safe result for six adversarial cases rather than accepting a declared expected flag:

- an executable project module without repository trust or review is denied;
- an experimental handler selected for a mandatory intent is rejected as unsupported;
- a mandatory policy-service outage without an authenticated cache denies;
- a concurrent duplicate event reconciles to one side effect;
- an agent launch beyond depth one is denied; and
- an expired exception denies.

The harness hook is early feedback, not the mandatory boundary. The required CI check remains the independent authority for the selected repository mutation.

## Validation

Run `assets/walking-skeleton/run-skeleton.sh --yes` for the executable proof and `node scripts/validate-walking-skeleton-fixtures.mjs` from this guide directory for the recorded scenario, or `npx moon run skill-agent-governance:test` from the repository root to run both. The validator cross-checks the walking skeleton against the governance, harness, external-enforcement, enterprise-platform, and module-template fixtures before replaying its recorded outcomes.
