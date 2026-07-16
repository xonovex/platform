# release-run: Execute or Recover a Controlled Release

## Core workflow

1. Resolve the exact integrated revision or immutable artifact digest, release target,
   protected environment, release strategy, observation window, rollback/recovery plan,
   current approvals, evidence, and policy/profile versions.
2. Apply [operational-contracts.md](operational-contracts.md). Verify the selected controlled
   automation and protected environment enforce authorization at the target boundary,
   release least-privilege credentials only after approval, and preserve native evidence.
3. Ask that external automation to execute, verify, pause, resume, or roll back the release.
   An agent may monitor and advise but cannot directly deploy or manufacture an approval.
4. Record rollout stages, artifact and configuration revisions, health gates, user impact,
   failures, cancellation, timeout, and provider-native deployment/audit references.
5. On a missed gate or ambiguous target state, stop progression and invoke the declared
   rollback or recovery path. Never report a failed release as successful because rollback
   succeeded.
6. Re-resolve the target and publish a Release result with exact bindings, approvals,
   automation/enforcement identity, action, outcome, verification, rollback/recovery, and
   linked Observation or Incident results.
