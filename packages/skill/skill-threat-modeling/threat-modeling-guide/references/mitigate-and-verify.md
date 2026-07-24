# Mitigate and Verify

For every material threat, choose one or more explicit treatments:

- eliminate the risky flow, privilege, data, or feature;
- prevent the precondition or adverse transition;
- reduce blast radius with isolation, least privilege, quotas, or bounded effects;
- detect and investigate with protected, useful evidence;
- recover or compensate safely after partial effects;
- record residual risk for an accountable external decision.

Define each mitigation:

```markdown
Threat:
Affected component and revision:
Owner:
Design or control:
Security objective:
Implementation boundary:
Verification:
Operational signal:
Failure and fallback:
Residual risk:
```

Verification should attempt the threat path and control failure, not merely inspect
that a setting exists. Include negative, bypass, replay, ordering, concurrency,
privilege, partial-failure, recovery, and observability cases when relevant.

Update the model after new data flows, identities, integrations, privileges,
deployment boundaries, incidents, or mitigation changes. Mark prior evidence stale
when it assessed a different revision or assumption set.
