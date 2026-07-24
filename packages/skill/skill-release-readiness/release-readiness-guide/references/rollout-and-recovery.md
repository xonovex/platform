# Rollout and Recovery

## Progressive Rollout

Define target, order, cohort or traffic selection, stage size, minimum observation
window, concurrency, blast-radius limit, capacity, dependencies, compatibility, and
feature controls. Compare canary and control populations only when their traffic and
conditions support the comparison.

For each stage, name:

```text
Signals and baseline:
Guardrails and thresholds:
Evaluation window:
Proceed, pause, and abort conditions:
Decision owner and operator:
Evidence location:
```

Use both technical and user or business signals. Account for delayed, sparse, batch,
and irreversible outcomes.

## Recovery

Choose rollback, roll-forward, traffic shift, feature disable, or service degradation
per failure mode. Validate compatibility across application versions, schemas,
messages, clients, and data. Define backup or restore point, recovery objectives,
operator permissions, rehearsal evidence, and the point of no return.

Prepare deployment and recovery runbooks, on-call and incident contacts, support
briefing, status and stakeholder communications, user impact language, and change
window constraints. Never assume rollback is safe because the deployment tool exposes
a rollback button.
