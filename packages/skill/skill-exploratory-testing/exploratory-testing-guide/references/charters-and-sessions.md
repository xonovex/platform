# Charters and Sessions

Create one charter:

```markdown
## Mission

Explore <target> with <resources and techniques> to discover information about <risk>.

Subject and revision:
Environment:
In scope:
Out of scope:
Starting state and data:
Time box:
Oracles:
Evidence to retain:
Stop conditions:
```

During the session:

1. Confirm the exact build, environment, dependencies, flags, identities, and data.
2. Establish a known starting state and verify observability before testing.
3. Follow the charter while adapting tests to new observations.
4. Record concise time-ordered notes: setup, action, observation, question, defect,
   tool or data issue, and coverage.
5. Preserve identifiers and artifacts needed to reproduce a finding without collecting
   unrelated sensitive data.
6. Pause on unsafe effects, corrupted shared state, missing authority, or an
   environment that no longer represents the target.
7. Stop at the time box, summarize unfinished threads, and create a focused follow-up
   charter rather than silently extending scope.

Separate test execution time from setup, investigation, defect reporting, and blocked
time so the debrief can explain where the session went.
