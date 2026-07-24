# Decision and Verification

## Recommendation Contract

```text
Candidate identity:
Target and rollout:
Criteria and evidence summary:
Open defects and evidence gaps:
Exceptions and accountable owners:
Residual risks and expiry:
Recovery readiness:
Recommendation: ready | conditionally ready | not ready | blocked
Conditions and rationale:
Required approver and provider gate:
Post-deployment verification:
```

Ready means the current evidence and controls support presenting this candidate to the
accountable approval process. Conditionally ready names conditions that must be
verified before apply. Not ready identifies failed or missing criteria. Blocked means
the subject, evidence, authority, or recovery path is insufficient to assess.

## Handoff and Verification

Keep the assessment subject revision unchanged. Provider-native approval, scheduling,
environment mutation, and deployment remain separate protected effects. Attach the
readiness evidence to that process without claiming it passed.

After deployment, verify deployed digest and configuration, migration state, health,
guardrails, user-critical journeys, security controls, telemetry, and support signals
for the stated observation window. Record proceed, hold, rollback, or incident handoff
against the deployed revision.
