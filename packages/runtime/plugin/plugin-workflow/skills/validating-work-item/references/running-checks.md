# Running the checks

## Contents

- Running the verification
- Criteria with a visible result

## Running the verification

1. Run each command the `verification` names, from the worktree, exactly as written. Record the command as run.
2. Read what the check asserts. A passing test proves only its assertions, so compare them with the criterion's text. A suite that is green and never touches the changed behavior passes nothing.
3. Compare a failure with the state before the change when you can, so an old failure is not blamed on the work, and a new failure is not excused as old. If you cannot establish the baseline, say so in the limits.
4. A check that depends on a service, a seed or the clock gets its conditions written down.
5. Treat editor diagnostics and stale indexes as hints. A real build or test run settles a dispute.
6. Run each check that matters twice when its result looks unstable. Report an unstable check as such and treat the criterion as `blocked` unless one stable reading decides it.

## Criteria with a visible result

For a criterion about something a person sees, run the artifact and open the capture: the screenshot, the rendered page, the output file. An assertion can pass while the capture shows an empty panel. If you cannot render or open it, the criterion is `blocked`, with the assertion's result in the limits.
