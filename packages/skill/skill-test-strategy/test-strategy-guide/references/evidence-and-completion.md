# Evidence and Completion

## Entry and Completion

Entry criteria state what must be identifiable and controllable before a test activity
can produce trustworthy evidence: subject revision, environment, data, dependencies,
oracles, authorization, and required instrumentation.

Completion criteria consider:

- coverage of prioritized risks and acceptance examples;
- result provenance against the exact candidate;
- unresolved defect severity and affected scope;
- false-positive, false-negative, and flaky-result investigation;
- environment, data, sample, and observability limitations;
- regression and change-impact evidence;
- explicitly owned residual risk.

Do not use elapsed time, executed test count, pass percentage, or code coverage alone
as a release conclusion.

## Evidence Report

```text
Subject and revision:
Test basis and strategy revision:
Risk or criterion:
Technique, level, and environment:
Result and evidence reference:
Defects and retest state:
Limitations and freshness:
Residual risk:
Recommended next action:
```

Keep independent test evidence separate from implementation claims. Hand candidate
evidence to release readiness; the accountable provider process owns approval.
