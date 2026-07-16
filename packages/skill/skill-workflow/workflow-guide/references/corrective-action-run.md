# corrective-action-run: Implement and Verify Corrective Action

## Core workflow

1. Resolve the exact source Incident, finding, Observation, Assessment, Review, QA, or other
   result plus the affected subjects, current state, evidence, and policy/profile.
2. Record the cause analysis and its confidence separately from the selected correction.
   Define action, accountable owner, scope, priority, due state, dependencies, success
   measures, rollback, and verification method.
3. Execute privileged changes only through the authorized lifecycle capability from
   [operational-contracts.md](operational-contracts.md); ordinary agent tools and advisory
   recommendations grant no authority.
4. Verify the exact changed revision in the required environment against the stated success
   measures. Record independent evidence, evaluator origin, limitations, regressions, and
   whether the action was ineffective, partial, effective, or stale.
5. Publish or revise the Corrective Action result. Closure requires completed action,
   successful verification, effectiveness evidence over the declared window, residual-risk
   disposition, and a linked Learning proposal or an explicit reason none is needed.

Reopen or create additional action when verification fails, the subject changes, the
effect does not persist, or the original cause analysis is disproved.
